import type { Dirent } from 'node:fs'
import type { ServeOptions } from './types'
import fs from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import process from 'node:process'
import { cancel, intro, note, outro } from '@clack/prompts'
import ansis from 'ansis'
import { printTitle } from '../../shared/utils'
import indexHtmlAsset from './views/index.html' with { type: 'text' }

// Bun's ambient .html type is HTMLBundle even when the text loader is selected.
const indexHtml = indexHtmlAsset as unknown as string

interface DirectoryEntry {
  name: string
  isDirectory: boolean
  isPreviewableImage: boolean
  size: number
  mtime: Date
  href: string
}

// ─── Utility Functions ────────────────────────────────────────────────────────

function formatFileSize(bytes: number): string {
  if (bytes === 0)
    return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]!}`
}

function formatDate(date: Date): string {
  const y = date.getFullYear()
  const mo = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const h = String(date.getHours()).padStart(2, '0')
  const mi = String(date.getMinutes()).padStart(2, '0')
  return `${y}-${mo}-${d} ${h}:${mi}`
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

const PREVIEWABLE_IMAGE_MIME_TYPES = new Map([
  ['.avif', 'image/avif'],
  ['.gif', 'image/gif'],
  ['.jpeg', 'image/jpeg'],
  ['.jpg', 'image/jpeg'],
  ['.png', 'image/png'],
  ['.svg', 'image/svg+xml'],
  ['.webp', 'image/webp'],
])

function getPreviewableImageMimeType(name: string): string | undefined {
  return PREVIEWABLE_IMAGE_MIME_TYPES.get(path.extname(name).toLowerCase())
}

// ─── Security ─────────────────────────────────────────────────────────────────

async function resolveSafePath(root: string, urlPath: string): Promise<string | null> {
  let decoded: string
  try {
    decoded = decodeURIComponent(urlPath)
  }
  catch {
    return null
  }

  const candidate = path.resolve(root, decoded.replace(/^\/+/, ''))
  const rootWithSep = root.endsWith(path.sep) ? root : root + path.sep
  const isWithinRoot = candidate === root || candidate.startsWith(rootWithSep)

  if (!isWithinRoot)
    return null

  try {
    const realCandidate = await fs.realpath(candidate)
    const realRoot = await fs.realpath(root)
    const realRootWithSep = realRoot.endsWith(path.sep) ? realRoot : realRoot + path.sep
    const realIsWithin = realCandidate === realRoot || realCandidate.startsWith(realRootWithSep)
    if (!realIsWithin)
      return null
  }
  catch {
    // Path doesn't exist yet — caller's fs.stat will produce the 404
  }

  return candidate
}

// ─── File Serving ─────────────────────────────────────────────────────────────

function isInlineMimeType(mimeType: string): boolean {
  const base = mimeType.split(';')[0]!.trim().toLowerCase()
  const [type, subtype] = base.split('/')
  if (!type || !subtype)
    return false
  if (['text', 'image', 'video', 'audio'].includes(type))
    return true
  if (type === 'application')
    return ['pdf', 'json', 'xml', 'javascript', 'xhtml+xml', 'atom+xml', 'rss+xml', 'ld+json'].includes(subtype)
  return false
}

async function serveFile(filePath: string, stat: Awaited<ReturnType<typeof fs.stat>>): Promise<Response> {
  const file = Bun.file(filePath)
  const mimeType = getPreviewableImageMimeType(filePath) || file.type || 'application/octet-stream'
  const encoded = encodeURIComponent(path.basename(filePath)).replace(/'/g, '%27')
  const disposition = isInlineMimeType(mimeType)
    ? `inline; filename="${encoded}"; filename*=UTF-8''${encoded}`
    : `attachment; filename="${encoded}"; filename*=UTF-8''${encoded}`
  return new Response(file, {
    headers: {
      'Content-Type': mimeType,
      'Content-Disposition': disposition,
      'Content-Length': String(stat.size),
      'Last-Modified': stat.mtime.toUTCString(),
    },
  })
}

// ─── HTML Template ────────────────────────────────────────────────────────────

function buildBreadcrumb(urlPath: string): string {
  const parts = urlPath.split('/').filter(Boolean)
  let html = `<a href="/">/</a>`
  let accumulated = ''
  for (const part of parts) {
    accumulated += `/${encodeURIComponent(part)}`
    html += `&nbsp;/&nbsp;<a href="${accumulated}/">${escapeHtml(decodeURIComponent(part))}</a>`
  }
  return html
}

type TemplateKey
  = | 'TITLE'
    | 'BREADCRUMB'
    | 'PARENT_ROW'
    | 'ENTRY_ROWS'
    | 'ITEM_COUNT'
    | 'ITEM_SUFFIX'
    | 'UPLOAD_ENABLED'

const TEMPLATE_SLOT_PATTERN = /\{\{([A-Z_]+)\}\}/g

function renderIndexTemplate(values: Record<TemplateKey, string>): string {
  return indexHtml.replace(TEMPLATE_SLOT_PATTERN, (placeholder, key: string) => {
    if (!(key in values))
      throw new Error(`Unknown HTML template placeholder: ${placeholder}`)

    return values[key as TemplateKey]
  })
}

function buildDirectoryHtml(urlPath: string, entries: DirectoryEntry[], uploadEnabled: boolean): string {
  const isRoot = urlPath === '/'
  const title = `Index of ${escapeHtml(urlPath)}`
  const breadcrumb = buildBreadcrumb(urlPath)

  const parentHref = urlPath.replace(/[^/]+\/$/, '') || '/'
  const parentRow = isRoot
    ? ''
    : `<tr class="parent-row">
        <td class="name-cell parent-link" colspan="3">
          <span class="icon">&#x2B06;</span>
          <a href="${parentHref}">Parent directory</a>
        </td>
      </tr>`

  const entryRows = entries.length === 0
    ? '<tr><td colspan="3" class="empty-state">Empty directory</td></tr>'
    : entries.map((e) => {
        const icon = e.isDirectory ? '&#x1F4C1;' : '&#x1F4C4;'
        const entryIcon = e.isPreviewableImage
          ? `<span class="thumbnail" aria-hidden="true">
              <img src="${escapeHtml(e.href)}" alt="" loading="lazy" decoding="async" onerror="this.parentElement.classList.add('failed')">
              <span class="thumbnail-fallback">&#x1F4C4;</span>
            </span>`
          : `<span class="icon" aria-hidden="true">${icon}</span>`
        const sizeStr = e.isDirectory ? '-' : formatFileSize(e.size)
        const dateStr = formatDate(e.mtime)
        const nameClass = e.isDirectory ? 'dir-link' : 'file-link'
        return `<tr>
      <td class="name-cell ${nameClass}">
        <a href="${escapeHtml(e.href)}">${entryIcon}<span>${escapeHtml(e.name)}${e.isDirectory ? '/' : ''}</span></a>
      </td>
      <td class="size-col">${sizeStr}</td>
      <td class="date-col">${dateStr}</td>
    </tr>`
      }).join('\n')

  return renderIndexTemplate({
    TITLE: title,
    BREADCRUMB: breadcrumb,
    PARENT_ROW: parentRow,
    ENTRY_ROWS: entryRows,
    ITEM_COUNT: String(entries.length),
    ITEM_SUFFIX: entries.length !== 1 ? 's' : '',
    UPLOAD_ENABLED: String(uploadEnabled),
  })
}

// ─── Directory Listing ────────────────────────────────────────────────────────

async function serveDirectory(dirPath: string, urlPath: string, uploadEnabled: boolean): Promise<Response> {
  if (!urlPath.endsWith('/')) {
    return Response.redirect(`${urlPath}/`, 301)
  }

  let rawEntries: Dirent[]
  try {
    rawEntries = await fs.readdir(dirPath, { withFileTypes: true })
  }
  catch {
    return new Response('403 Forbidden', { status: 403, headers: { 'Content-Type': 'text/plain' } })
  }

  const entries: DirectoryEntry[] = []
  for (const dirent of rawEntries) {
    const fullPath = path.join(dirPath, dirent.name)
    let entryStat: Awaited<ReturnType<typeof fs.stat>>
    try {
      entryStat = await fs.stat(fullPath)
    }
    catch {
      continue
    }
    const isDir = entryStat.isDirectory()
    const encodedName = encodeURIComponent(dirent.name)
    const href = isDir ? `${urlPath}${encodedName}/` : `${urlPath}${encodedName}`
    entries.push({
      name: dirent.name,
      isDirectory: isDir,
      isPreviewableImage: !isDir && getPreviewableImageMimeType(dirent.name) !== undefined,
      size: isDir ? 0 : entryStat.size,
      mtime: entryStat.mtime,
      href,
    })
  }

  entries.sort((a, b) => {
    if (a.isDirectory !== b.isDirectory)
      return a.isDirectory ? -1 : 1
    return a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })
  })

  const html = buildDirectoryHtml(urlPath, entries, uploadEnabled)
  return new Response(html, { headers: { 'Content-Type': 'text/html; charset=utf-8' } })
}

// ─── CORS ─────────────────────────────────────────────────────────────────────

const CORS_HEADERS: Record<string, string> = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, HEAD, OPTIONS',
  'Access-Control-Allow-Headers': '*',
  'Access-Control-Max-Age': '86400',
}

function withCors(res: Response): Response {
  for (const [key, value] of Object.entries(CORS_HEADERS)) {
    res.headers.set(key, value)
  }
  return res
}

// ─── Upload ───────────────────────────────────────────────────────────────────

const MAX_UPLOAD_SIZE = 1024 * 1024 * 1024 // 1 GB

function sanitizeFilename(name: string): string | null {
  const cleaned = name.replace(/[/\\]/g, '').replace(/\0/g, '').trim()
  if (!cleaned || cleaned === '.' || cleaned === '..')
    return null
  return cleaned
}

async function resolveUploadFilename(dir: string, name: string): Promise<string> {
  const lastDot = name.lastIndexOf('.')
  const base = lastDot > 0 ? name.slice(0, lastDot) : name
  const ext = lastDot > 0 ? name.slice(lastDot) : ''

  try {
    await fs.access(path.join(dir, name))
  }
  catch {
    return name
  }

  for (let i = 1; i <= 9999; i++) {
    const candidate = `${base} (${i})${ext}`
    try {
      await fs.access(path.join(dir, candidate))
    }
    catch {
      return candidate
    }
  }
  throw new Error('Too many files with the same name')
}

async function handleUpload(req: Request, root: string): Promise<Response> {
  const contentLength = req.headers.get('Content-Length')
  if (contentLength && Number(contentLength) > MAX_UPLOAD_SIZE) {
    return new Response('413 Payload Too Large', { status: 413, headers: { 'Content-Type': 'text/plain' } })
  }

  const url = new URL(req.url)
  const dirParam = url.searchParams.get('dir') ?? '/'

  const targetDir = await resolveSafePath(root, dirParam)
  if (targetDir === null) {
    return new Response('403 Forbidden', { status: 403, headers: { 'Content-Type': 'text/plain' } })
  }

  let dirStat: Awaited<ReturnType<typeof fs.stat>>
  try {
    dirStat = await fs.stat(targetDir)
    if (!dirStat.isDirectory()) {
      return new Response('400 Bad Request: Not a directory', { status: 400, headers: { 'Content-Type': 'text/plain' } })
    }
  }
  catch {
    return new Response('404 Not Found', { status: 404, headers: { 'Content-Type': 'text/plain' } })
  }

  let file: File | string | null
  try {
    const formData = await req.formData()
    file = formData.get('file') as File | string | null
  }
  catch {
    return new Response('400 Bad Request: Invalid form data', { status: 400, headers: { 'Content-Type': 'text/plain' } })
  }

  if (!(file instanceof File)) {
    return new Response('400 Bad Request: No file provided', { status: 400, headers: { 'Content-Type': 'text/plain' } })
  }

  if (file.size > MAX_UPLOAD_SIZE) {
    return new Response('413 Payload Too Large', { status: 413, headers: { 'Content-Type': 'text/plain' } })
  }

  const rawName = sanitizeFilename(file.name)
  if (!rawName) {
    return new Response('400 Bad Request: Invalid filename', { status: 400, headers: { 'Content-Type': 'text/plain' } })
  }

  let finalName: string
  try {
    finalName = await resolveUploadFilename(targetDir, rawName)
  }
  catch (err) {
    return new Response(`409 Conflict: ${err instanceof Error ? err.message : String(err)}`, { status: 409, headers: { 'Content-Type': 'text/plain' } })
  }

  const tmpPath = path.join(targetDir, `.upload-${crypto.randomUUID()}.tmp`)
  const finalPath = path.join(targetDir, finalName)

  try {
    await Bun.write(tmpPath, file)
    await fs.rename(tmpPath, finalPath)
  }
  catch (err) {
    try {
      await fs.unlink(tmpPath)
    }
    catch {}
    return new Response(`500 Internal Server Error: ${err instanceof Error ? err.message : String(err)}`, { status: 500, headers: { 'Content-Type': 'text/plain' } })
  }

  return new Response(JSON.stringify({ ok: true, filename: finalName }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  })
}

// ─── Request Router ───────────────────────────────────────────────────────────

async function handleRequest(req: Request, root: string, uploadEnabled: boolean): Promise<Response> {
  if (req.method === 'OPTIONS') {
    return new Response(null, { status: 204, headers: CORS_HEADERS })
  }

  const url = new URL(req.url)

  if (uploadEnabled && req.method === 'POST' && url.pathname === '/__upload') {
    return withCors(await handleUpload(req, root))
  }

  const safePath = await resolveSafePath(root, url.pathname)

  if (safePath === null) {
    return withCors(new Response('403 Forbidden', { status: 403, headers: { 'Content-Type': 'text/plain' } }))
  }

  let stat: Awaited<ReturnType<typeof fs.stat>>
  try {
    stat = await fs.stat(safePath)
  }
  catch {
    return withCors(new Response('404 Not Found', { status: 404, headers: { 'Content-Type': 'text/plain' } }))
  }

  if (stat.isDirectory()) {
    return withCors(await serveDirectory(safePath, url.pathname, uploadEnabled))
  }

  return withCors(await serveFile(safePath, stat))
}

// ─── Main Export ──────────────────────────────────────────────────────────────

function getLanAddresses(): string[] {
  const interfaces = os.networkInterfaces()
  const addresses: string[] = []
  for (const nets of Object.values(interfaces)) {
    if (!nets)
      continue
    for (const net of nets) {
      if (net.family === 'IPv4' && !net.internal)
        addresses.push(net.address)
    }
  }
  return addresses
}

export async function serve(opt: ServeOptions): Promise<void> {
  printTitle()
  intro(ansis.bold('Static File Server'))

  const root = path.resolve(opt.directory)

  let rootStat: Awaited<ReturnType<typeof fs.stat>>
  try {
    rootStat = await fs.stat(root)
  }
  catch {
    cancel(`Directory not found: ${ansis.dim(root)}`)
    return
  }

  if (!rootStat.isDirectory()) {
    cancel(`Path is not a directory: ${ansis.dim(root)}`)
    return
  }

  let server: ReturnType<typeof Bun.serve>
  try {
    server = Bun.serve({
      port: opt.port,
      hostname: opt.address,
      fetch(req) {
        return handleRequest(req, root, opt.upload)
      },
    })
  }
  catch (err) {
    cancel(`Failed to start server: ${err instanceof Error ? err.message : String(err)}`)
    return
  }

  const displayAddress = opt.address === '0.0.0.0' ? 'localhost' : opt.address
  const url = `http://${displayAddress}:${server.port}`

  const msgs: string[] = []
  msgs.push(`  ${ansis.dim('Local')}     ${ansis.cyan(url)}`)

  if (opt.address === '0.0.0.0') {
    const lanAddrs = getLanAddresses()
    for (const addr of lanAddrs) {
      msgs.push(`  ${ansis.dim('Network')}   ${ansis.cyan(`http://${addr}:${server.port}`)}`)
    }
  }

  msgs.push(`  ${ansis.dim('Directory')} ${ansis.dim(root)}`)
  msgs.push(`  ${ansis.dim('Bind')}      ${ansis.dim(`${opt.address}:${server.port}`)}`)
  msgs.push(`  ${ansis.dim('Upload')}    ${opt.upload ? ansis.green('enabled') : ansis.dim('disabled')}`)
  note(msgs.join('\n'), `Server running`)

  await new Promise<void>((resolve) => {
    const shutdown = (): void => {
      server.stop(true)
      outro('Server stopped.')
      resolve()
    }
    process.once('SIGINT', shutdown)
    process.once('SIGTERM', shutdown)
  })
}
