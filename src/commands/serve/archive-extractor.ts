import type { StatsFs } from 'node:fs'
import fs from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'
import { ensureSevenZipRuntime } from './archive-runtime'
import { isLayeredTarArchiveName } from './archive-support'
import { ServeWorkspaceError } from './types'

export interface ArchiveInspection {
  uncompressedBytes: number
  entryCount: number
}

export interface ArchiveExtractorOptions {
  signal?: AbortSignal
  onProgress?: (progress: number) => void
  onInspect?: (inspection: ArchiveInspection) => void
}

export interface ArchiveExtractor {
  extract: (source: string, destination: string, options?: ArchiveExtractorOptions) => Promise<ArchiveInspection>
}

export interface SevenZipArchiveExtractorOptions {
  executable?: () => Promise<string>
  statfs?: (target: string) => Promise<StatsFs>
}

function archiveFailure(output: string): ServeWorkspaceError {
  const normalized = output.toLowerCase()
  if (normalized.includes('wrong password') || normalized.includes('enter password') || normalized.includes('encrypted'))
    return new ServeWorkspaceError('ENCRYPTED_ARCHIVE', 'Encrypted archives are not supported')
  return new ServeWorkspaceError('INVALID_ARCHIVE', 'Archive is invalid, damaged, or unsupported')
}

function validEntryPath(entryPath: string): boolean {
  const containsControlCharacter = [...entryPath].some((character) => {
    const code = character.charCodeAt(0)
    return code <= 0x1F || code === 0x7F
  })
  if (!entryPath || containsControlCharacter || entryPath.includes('\\') || entryPath.startsWith('/') || /^[a-z]:/i.test(entryPath))
    return false
  return !entryPath.split('/').some(segment => segment === '.' || segment === '..')
}

async function readError(stream: ReadableStream<Uint8Array>): Promise<string> {
  const text = await new Response(stream).text()
  return text.slice(-64 * 1024)
}

async function inspectArchive(executable: string, source: string, signal?: AbortSignal): Promise<ArchiveInspection> {
  signal?.throwIfAborted()
  const child = Bun.spawn([executable, 'l', '-slt', '-sccUTF-8', '--', source], {
    env: { ...process.env, LANG: 'C', LC_ALL: 'C' },
    stdin: 'ignore',
    stdout: 'pipe',
    stderr: 'pipe',
  })
  const abort = (): void => child.kill()
  signal?.addEventListener('abort', abort, { once: true })
  let buffer = ''
  let current: Record<string, string> = {}
  let entriesStarted = false
  let uncompressedBytes = 0
  let entryCount = 0
  const finishEntry = (): void => {
    if (!current.Path) {
      current = {}
      return
    }
    if (!validEntryPath(current.Path)) {
      child.kill()
      throw new ServeWorkspaceError('INVALID_ARCHIVE', 'Archive contains an unsafe entry path')
    }
    if (current.Encrypted === '+') {
      child.kill()
      throw new ServeWorkspaceError('ENCRYPTED_ARCHIVE', 'Encrypted archives are not supported')
    }
    const size = Number(current.Size ?? 0)
    if (!Number.isSafeInteger(size) || size < 0 || uncompressedBytes + size > Number.MAX_SAFE_INTEGER) {
      child.kill()
      throw new ServeWorkspaceError('INVALID_ARCHIVE', 'Archive reports an invalid unpacked size')
    }
    uncompressedBytes += size
    entryCount++
    current = {}
  }
  const parseLine = (line: string): void => {
    if (/^-{10,}\r?$/.test(line)) {
      finishEntry()
      entriesStarted = true
      return
    }
    const separator = line.indexOf(' = ')
    if (!entriesStarted) {
      if (separator === -1)
        return
      const key = line.slice(0, separator)
      const value = line.slice(separator + 3).replace(/\r$/, '')
      if ((key === 'Type' && value === 'Split') || (key === 'Volumes' && Number(value) > 1)) {
        child.kill()
        throw new ServeWorkspaceError('INVALID_ARCHIVE', 'Multi-volume archives are not supported')
      }
      return
    }
    if (!line.trim()) {
      finishEntry()
      return
    }
    if (separator !== -1)
      current[line.slice(0, separator)] = line.slice(separator + 3).replace(/\r$/, '')
  }
  const stdout = (async (): Promise<void> => {
    const reader = child.stdout.getReader()
    const decoder = new TextDecoder()
    try {
      while (true) {
        const chunk = await reader.read()
        if (chunk.done)
          break
        buffer += decoder.decode(chunk.value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''
        for (const line of lines)
          parseLine(line)
      }
      buffer += decoder.decode()
      if (buffer)
        parseLine(buffer)
      finishEntry()
    }
    finally {
      reader.releaseLock()
    }
  })()
  const stderr = readError(child.stderr)
  try {
    const [exitCode, errorOutput] = await Promise.all([child.exited, stderr, stdout]).then(values => [values[0], values[1]] as const)
    signal?.throwIfAborted()
    if (exitCode !== 0)
      throw archiveFailure(errorOutput)
    return { uncompressedBytes, entryCount }
  }
  finally {
    signal?.removeEventListener('abort', abort)
  }
}

async function requireCapacity(target: string, inspection: ArchiveInspection, statfs: (target: string) => Promise<StatsFs>): Promise<void> {
  const filesystem = await statfs(target)
  const availableBytes = filesystem.bavail * filesystem.bsize
  const reservedBytes = Math.min(1024 ** 3, Math.floor(availableBytes * 0.1))
  if (inspection.uncompressedBytes > availableBytes - reservedBytes)
    throw new ServeWorkspaceError('INSUFFICIENT_SPACE', 'Archive does not fit in the available disk space')
  if (filesystem.ffree > 0) {
    const reservedEntries = Math.min(1024, Math.floor(filesystem.ffree * 0.1))
    if (inspection.entryCount > filesystem.ffree - reservedEntries)
      throw new ServeWorkspaceError('INSUFFICIENT_SPACE', 'Archive does not fit in the available filesystem entries')
  }
}

async function extractWithProgress(
  executable: string,
  source: string,
  destination: string,
  signal: AbortSignal | undefined,
  onProgress: ((progress: number) => void) | undefined,
  start: number,
  span: number,
): Promise<void> {
  signal?.throwIfAborted()
  const child = Bun.spawn([
    executable,
    'x',
    '-y',
    '-snld0',
    '-sccUTF-8',
    '-bso0',
    '-bse2',
    '-bsp1',
    `-o${destination}`,
    '--',
    source,
  ], {
    env: { ...process.env, LANG: 'C', LC_ALL: 'C' },
    stdin: 'ignore',
    stdout: 'pipe',
    stderr: 'pipe',
  })
  const abort = (): void => child.kill()
  signal?.addEventListener('abort', abort, { once: true })
  const progress = (async (): Promise<void> => {
    const reader = child.stdout.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    try {
      while (true) {
        const chunk = await reader.read()
        if (chunk.done)
          break
        buffer += decoder.decode(chunk.value, { stream: true })
        const matches = [...buffer.matchAll(/(?:^|\r|\n)\s*(\d{1,3})%/g)]
        const latest = matches.at(-1)?.[1]
        if (latest)
          onProgress?.(Math.min(100, Math.round(start + Number(latest) / 100 * span)))
        buffer = buffer.slice(-128)
      }
    }
    finally {
      reader.releaseLock()
    }
  })()
  const stderr = readError(child.stderr)
  try {
    const [exitCode, errorOutput] = await Promise.all([child.exited, stderr, progress]).then(values => [values[0], values[1]] as const)
    signal?.throwIfAborted()
    if (exitCode !== 0)
      throw archiveFailure(errorOutput)
    onProgress?.(Math.round(start + span))
  }
  finally {
    signal?.removeEventListener('abort', abort)
  }
}

async function onlyExtractedTar(directory: string): Promise<string> {
  const entries = await fs.readdir(directory, { recursive: true, withFileTypes: true })
  const files = entries.filter(entry => entry.isFile())
  if (files.length !== 1 || path.extname(files[0]!.name).toLowerCase() !== '.tar')
    throw new ServeWorkspaceError('INVALID_ARCHIVE', 'Compressed TAR archive did not contain one TAR stream')
  return path.join(files[0]!.parentPath, files[0]!.name)
}

export function createSevenZipArchiveExtractor(options: SevenZipArchiveExtractorOptions = {}): ArchiveExtractor {
  const executable = options.executable ?? ensureSevenZipRuntime
  const statfs = options.statfs ?? (target => fs.statfs(target))
  return {
    async extract(source, destination, extractOptions = {}) {
      const binary = await executable()
      const outerInspection = await inspectArchive(binary, source, extractOptions.signal)
      await requireCapacity(path.dirname(destination), outerInspection, statfs)
      if (!isLayeredTarArchiveName(path.basename(source))) {
        extractOptions.onInspect?.(outerInspection)
        await extractWithProgress(binary, source, destination, extractOptions.signal, extractOptions.onProgress, 0, 100)
        return outerInspection
      }

      const outerDirectory = `${destination}.outer`
      await fs.mkdir(outerDirectory)
      try {
        await extractWithProgress(binary, source, outerDirectory, extractOptions.signal, extractOptions.onProgress, 0, 35)
        const tarPath = await onlyExtractedTar(outerDirectory)
        const innerInspection = await inspectArchive(binary, tarPath, extractOptions.signal)
        await requireCapacity(path.dirname(destination), innerInspection, statfs)
        extractOptions.onInspect?.(innerInspection)
        await extractWithProgress(binary, tarPath, destination, extractOptions.signal, extractOptions.onProgress, 35, 65)
        return innerInspection
      }
      finally {
        await fs.rm(outerDirectory, { recursive: true, force: true })
      }
    },
  }
}
