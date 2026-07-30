import type { Stats } from 'node:fs'
import type {
  ServeDirectoryEntry,
  ServeDirectoryListing,
  ServeFile,
  ServePreviewKind,
  ServeTextPreview,
  ServeUploadResult,
  ServeWorkspace,
} from './types'
import fs from 'node:fs/promises'
import path from 'node:path'
import { ServeWorkspaceError } from './types'

const IMAGE_MIME_TYPES = new Map([
  ['.avif', 'image/avif'],
  ['.gif', 'image/gif'],
  ['.jpeg', 'image/jpeg'],
  ['.jpg', 'image/jpeg'],
  ['.png', 'image/png'],
  ['.svg', 'image/svg+xml'],
  ['.webp', 'image/webp'],
])

export const MAX_TEXT_PREVIEW_BYTES = 2 * 1024 * 1024
export const MAX_UPLOAD_BYTES = 1024 * 1024 * 1024

function normalizedRelativePath(value: string): string {
  if (value.includes('\0') || value.includes('\\') || value.startsWith('/') || /^[a-z]:/i.test(value))
    throw new ServeWorkspaceError('INVALID_PATH', 'Path must be relative to the served directory')

  const segments = value.split('/')
  if (segments.some(segment => segment === '.' || segment === '..'))
    throw new ServeWorkspaceError('PATH_FORBIDDEN', 'Path escapes the served directory')
  return segments.filter(Boolean).join('/')
}

function isWithinRoot(root: string, candidate: string): boolean {
  const rootWithSeparator = root.endsWith(path.sep) ? root : `${root}${path.sep}`
  return candidate === root || candidate.startsWith(rootWithSeparator)
}

function absolutePath(root: string, relativePath: string): string {
  return path.join(root, ...relativePath.split('/').filter(Boolean))
}

function encodedPath(relativePath: string): string {
  return relativePath.split('/').map(encodeURIComponent).join('/')
}

function uploadFilename(value: string): string {
  const filename = value.trim()
  if (!filename || filename === '.' || filename === '..' || filename.includes('/') || filename.includes('\\') || filename.includes('\0'))
    throw new ServeWorkspaceError('INVALID_UPLOAD', 'Upload filename is invalid')
  return filename
}

function collisionFilename(filename: string, index: number): string {
  if (index === 0)
    return filename
  const lastDot = filename.lastIndexOf('.')
  const base = lastDot > 0 ? filename.slice(0, lastDot) : filename
  const extension = lastDot > 0 ? filename.slice(lastDot) : ''
  return `${base} (${index})${extension}`
}

function previewKind(mimeType: string): ServePreviewKind {
  const baseMimeType = mimeType.split(';')[0]!.trim().toLowerCase()
  if (baseMimeType.startsWith('image/'))
    return 'image'
  if (baseMimeType.startsWith('video/'))
    return 'video'
  if (baseMimeType.startsWith('audio/'))
    return 'audio'
  if (baseMimeType === 'application/pdf')
    return 'pdf'
  if (baseMimeType.startsWith('text/') || [
    'application/json',
    'application/javascript',
    'application/ld+json',
    'application/xml',
    'application/xhtml+xml',
  ].includes(baseMimeType)) {
    return 'text'
  }
  return 'none'
}

function decodeText(bytes: Uint8Array): { text: string, encoding: 'utf-8' | 'utf-16le' | 'utf-16be' } | undefined {
  try {
    if (bytes[0] === 0xFF && bytes[1] === 0xFE)
      return { text: new TextDecoder('utf-16le', { fatal: true }).decode(bytes), encoding: 'utf-16le' }
    if (bytes[0] === 0xFE && bytes[1] === 0xFF)
      return { text: new TextDecoder('utf-16be', { fatal: true }).decode(bytes), encoding: 'utf-16be' }
    if (bytes.includes(0))
      return undefined
    return { text: new TextDecoder('utf-8', { fatal: true }).decode(bytes), encoding: 'utf-8' }
  }
  catch {
    return undefined
  }
}

async function resolvedInsideRoot(root: string, candidate: string): Promise<string> {
  let resolved: string
  try {
    resolved = await fs.realpath(candidate)
  }
  catch {
    throw new ServeWorkspaceError('NOT_FOUND', 'Path does not exist')
  }
  if (!isWithinRoot(root, resolved))
    throw new ServeWorkspaceError('PATH_FORBIDDEN', 'Path escapes the served directory')
  return resolved
}

async function browserEntry(root: string, directoryPath: string, name: string): Promise<ServeDirectoryEntry> {
  const relativePath = [directoryPath, name].filter(Boolean).join('/')
  const candidate = absolutePath(root, relativePath)
  let linkStat: Stats
  try {
    linkStat = await fs.lstat(candidate)
  }
  catch {
    return { name, path: relativePath, kind: 'unavailable', isSymlink: false, previewKind: 'none' }
  }

  const isSymlink = linkStat.isSymbolicLink()
  let resolved: string
  let stat: Stats
  try {
    resolved = await resolvedInsideRoot(root, candidate)
    stat = await fs.stat(resolved)
  }
  catch {
    return {
      name,
      path: relativePath,
      kind: 'unavailable',
      isSymlink,
      modifiedAt: linkStat.mtime.toISOString(),
      previewKind: 'none',
    }
  }

  if (stat.isDirectory()) {
    return {
      name,
      path: relativePath,
      kind: 'directory',
      isSymlink,
      modifiedAt: stat.mtime.toISOString(),
      previewKind: 'none',
      browseUrl: `/browse/${encodedPath(relativePath)}`,
    }
  }

  if (!stat.isFile()) {
    return {
      name,
      path: relativePath,
      kind: 'unavailable',
      isSymlink,
      modifiedAt: stat.mtime.toISOString(),
      previewKind: 'none',
    }
  }

  const extensionMimeType = IMAGE_MIME_TYPES.get(path.extname(name).toLowerCase())
  const mimeType = extensionMimeType ?? Bun.file(resolved).type ?? 'application/octet-stream'
  const urlPath = `/files/${encodedPath(relativePath)}`
  return {
    name,
    path: relativePath,
    kind: 'file',
    isSymlink,
    size: stat.size,
    modifiedAt: stat.mtime.toISOString(),
    mimeType,
    previewKind: previewKind(mimeType),
    fileUrl: urlPath,
    downloadUrl: `${urlPath}?download=1`,
  }
}

export async function createServeWorkspace(directory: string): Promise<ServeWorkspace> {
  let root: string
  try {
    root = await fs.realpath(path.resolve(directory))
  }
  catch {
    throw new ServeWorkspaceError('NOT_FOUND', `Directory not found: ${path.resolve(directory)}`)
  }
  const rootStat = await fs.stat(root)
  if (!rootStat.isDirectory())
    throw new ServeWorkspaceError('NOT_DIRECTORY', `Path is not a directory: ${root}`)

  return {
    async listDirectory(requestedPath): Promise<ServeDirectoryListing> {
      const relativePath = normalizedRelativePath(requestedPath)
      const resolved = await resolvedInsideRoot(root, absolutePath(root, relativePath))
      const stat = await fs.stat(resolved)
      if (!stat.isDirectory())
        throw new ServeWorkspaceError('NOT_DIRECTORY', 'Path is not a directory')

      let names: string[]
      try {
        names = await fs.readdir(resolved)
      }
      catch {
        throw new ServeWorkspaceError('UNAVAILABLE', 'Directory cannot be read')
      }
      const entries = await Promise.all(names.map(name => browserEntry(root, relativePath, name)))
      entries.sort((left, right) => {
        const leftRank = left.kind === 'directory' ? 0 : left.kind === 'file' ? 1 : 2
        const rightRank = right.kind === 'directory' ? 0 : right.kind === 'file' ? 1 : 2
        return leftRank - rightRank || left.name.localeCompare(right.name, undefined, { sensitivity: 'base' }) || left.name.localeCompare(right.name)
      })

      const parentSegments = relativePath.split('/').filter(Boolean)
      parentSegments.pop()
      return {
        rootName: path.basename(root) || root,
        path: relativePath,
        parentPath: relativePath ? parentSegments.join('/') : undefined,
        entries,
      }
    },
    async openFile(requestedPath): Promise<ServeFile> {
      const relativePath = normalizedRelativePath(requestedPath)
      const resolved = await resolvedInsideRoot(root, absolutePath(root, relativePath))
      const stat = await fs.stat(resolved)
      if (!stat.isFile())
        throw new ServeWorkspaceError('NOT_FILE', 'Path is not a file')
      const file = Bun.file(resolved)
      return {
        name: path.basename(relativePath),
        size: stat.size,
        modifiedAt: stat.mtime,
        mimeType: IMAGE_MIME_TYPES.get(path.extname(relativePath).toLowerCase()) ?? file.type ?? 'application/octet-stream',
        body: file,
      }
    },
    async readTextPreview(requestedPath): Promise<ServeTextPreview> {
      const relativePath = normalizedRelativePath(requestedPath)
      const resolved = await resolvedInsideRoot(root, absolutePath(root, relativePath))
      const stat = await fs.stat(resolved)
      if (!stat.isFile())
        throw new ServeWorkspaceError('NOT_FILE', 'Path is not a file')
      if (stat.size > MAX_TEXT_PREVIEW_BYTES) {
        return {
          status: 'too_large',
          size: stat.size,
          maxBytes: MAX_TEXT_PREVIEW_BYTES,
        }
      }
      const handle = await fs.open(resolved, 'r')
      let bytes: Uint8Array
      let size: number
      try {
        const buffer = new Uint8Array(MAX_TEXT_PREVIEW_BYTES)
        let bytesRead = 0
        while (bytesRead < buffer.length) {
          const result = await handle.read(buffer, bytesRead, buffer.length - bytesRead, bytesRead)
          if (result.bytesRead === 0)
            break
          bytesRead += result.bytesRead
        }
        size = (await handle.stat()).size
        if (size > MAX_TEXT_PREVIEW_BYTES) {
          return {
            status: 'too_large',
            size,
            maxBytes: MAX_TEXT_PREVIEW_BYTES,
          }
        }
        bytes = buffer.subarray(0, bytesRead)
      }
      finally {
        await handle.close()
      }
      const decoded = decodeText(bytes)
      return decoded
        ? { status: 'ready', ...decoded, size }
        : { status: 'binary', size }
    },
    async uploadFile(requestedDirectory, file): Promise<ServeUploadResult> {
      if (file.size > MAX_UPLOAD_BYTES)
        throw new ServeWorkspaceError('TOO_LARGE', 'Upload exceeds the 1 GiB file limit')
      const filename = uploadFilename(file.name)
      const directoryPath = normalizedRelativePath(requestedDirectory)
      const directory = await resolvedInsideRoot(root, absolutePath(root, directoryPath))
      const directoryStat = await fs.stat(directory)
      if (!directoryStat.isDirectory())
        throw new ServeWorkspaceError('NOT_DIRECTORY', 'Upload target is not a directory')

      const temporaryPath = path.join(directory, `.upload-${crypto.randomUUID()}.tmp`)
      try {
        await Bun.write(temporaryPath, file)
        for (let index = 0; index <= 9999; index++) {
          const finalFilename = collisionFilename(filename, index)
          const finalPath = path.join(directory, finalFilename)
          try {
            await fs.link(temporaryPath, finalPath)
            await fs.unlink(temporaryPath)
            return {
              filename: finalFilename,
              path: [directoryPath, finalFilename].filter(Boolean).join('/'),
              size: file.size,
            }
          }
          catch (error) {
            if ((error as NodeJS.ErrnoException).code !== 'EEXIST')
              throw error
          }
        }
        throw new ServeWorkspaceError('NAME_EXHAUSTED', 'Too many files have the same name')
      }
      catch (error) {
        await fs.unlink(temporaryPath).catch(() => {})
        if (error instanceof ServeWorkspaceError)
          throw error
        throw new ServeWorkspaceError('UNAVAILABLE', error instanceof Error ? error.message : String(error))
      }
    },
  }
}
