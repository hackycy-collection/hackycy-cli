import type { ServeErrorCode, ServeWorkspace } from './types'
import { isIP } from 'node:net'
import { z } from 'zod'
import { ServeWorkspaceError } from './types'
import serveWebApp from './web/index.html'
import { MAX_UPLOAD_BYTES } from './workspace'

export interface RunningServeServer {
  readonly url: URL
  readonly finished: Promise<void>
  stop: () => Promise<void>
}

const API_HEADERS = {
  'Cache-Control': 'no-store',
  'Content-Security-Policy': 'default-src \'none\'; frame-ancestors \'none\'',
  'Referrer-Policy': 'no-referrer',
  'X-Content-Type-Options': 'nosniff',
}

const APP_CONTENT_SECURITY_POLICY = 'default-src \'self\'; script-src \'self\'; style-src \'self\' \'unsafe-inline\'; img-src \'self\' blob: data:; media-src \'self\'; frame-src \'self\'; connect-src \'self\'; object-src \'none\'; base-uri \'none\'; frame-ancestors \'none\''
const ACTIVE_FILE_CONTENT_SECURITY_POLICY = 'sandbox; default-src \'none\'; style-src \'unsafe-inline\'; img-src data:; object-src \'none\'; base-uri \'none\'; frame-ancestors \'none\''

const operationPathSchema = z.string().max(4096)
const operationPathsSchema = z.array(operationPathSchema).min(1).max(1000)
const operationSchema = z.discriminatedUnion('action', [
  z.object({ action: z.literal('create-directory'), parentPath: operationPathSchema, name: z.string().max(4096) }).strict(),
  z.object({ action: z.literal('rename'), path: operationPathSchema, newName: z.string().max(4096) }).strict(),
  z.object({ action: z.literal('copy'), paths: operationPathsSchema, destinationPath: operationPathSchema }).strict(),
  z.object({ action: z.literal('move'), paths: operationPathsSchema, destinationPath: operationPathSchema }).strict(),
  z.object({ action: z.literal('delete'), paths: operationPathsSchema }).strict(),
])

function configureWebBundleHeaders(bundle: Bun.HTMLBundle): void {
  for (const file of bundle.files ?? []) {
    file.headers['referrer-policy'] = 'no-referrer'
    file.headers['x-content-type-options'] = 'nosniff'
    if (file.loader === 'html') {
      file.headers['cache-control'] = 'no-store'
      file.headers['content-security-policy'] = APP_CONTENT_SECURITY_POLICY
    }
    else {
      file.headers['cache-control'] = 'public, max-age=31536000, immutable'
    }
  }
}

const ERROR_STATUS: Record<ServeErrorCode, number> = {
  INVALID_PATH: 400,
  INVALID_UPLOAD: 400,
  INVALID_NAME: 400,
  INVALID_OPERATION: 409,
  PATH_FORBIDDEN: 403,
  NOT_FOUND: 404,
  NOT_DIRECTORY: 409,
  NOT_FILE: 409,
  TOO_LARGE: 413,
  NAME_EXHAUSTED: 409,
  ALREADY_EXISTS: 409,
  ROOT_IMMUTABLE: 409,
  UNAVAILABLE: 500,
}

function json(data: unknown, status = 200): Response {
  return Response.json(data, { status, headers: API_HEADERS })
}

function error(code: string, message: string, status: number): Response {
  return json({ version: 1, error: { code, message } }, status)
}

function workspaceError(cause: unknown): Response {
  if (cause instanceof ServeWorkspaceError)
    return error(cause.code, cause.message, ERROR_STATUS[cause.code])
  return error('INTERNAL_ERROR', cause instanceof Error ? cause.message : String(cause), 500)
}

function requireMethod(request: Request, method: string): Response | undefined {
  return request.method === method ? undefined : error('METHOD_NOT_ALLOWED', `Use ${method}`, 405)
}

function bindingAllowsHostname(bindingAddress: string, hostname: string): boolean {
  if (bindingAddress === '0.0.0.0')
    return hostname === 'localhost' || isIP(hostname) === 4
  if (bindingAddress === '::')
    return hostname === 'localhost' || isIP(hostname) !== 0
  if (bindingAddress === '127.0.0.1' || bindingAddress === '::1')
    return hostname === bindingAddress || hostname === 'localhost'
  return hostname === bindingAddress
}

function validateMutationOrigin(request: Request, bindingAddress: string): Response | undefined {
  const requestUrl = new URL(request.url)
  const origin = request.headers.get('Origin')
  if (!origin || origin !== requestUrl.origin || !bindingAllowsHostname(bindingAddress, requestUrl.hostname))
    return error('ORIGIN_FORBIDDEN', 'Management requests must come from the bound same origin', 403)
  return undefined
}

function encodedPath(relativePath: string): string {
  return relativePath.split('/').filter(Boolean).map(encodeURIComponent).join('/')
}

function requestFilePath(pathname: string): string {
  const encoded = pathname === '/files' ? '' : pathname.slice('/files/'.length)
  try {
    return encoded.split('/').filter(Boolean).map(decodeURIComponent).join('/')
  }
  catch {
    throw new ServeWorkspaceError('INVALID_PATH', 'File path is not valid URL encoding')
  }
}

function inlineMimeType(mimeType: string): boolean {
  const base = mimeType.split(';')[0]!.trim().toLowerCase()
  const [type, subtype] = base.split('/')
  if (!type || !subtype)
    return false
  if (['text', 'image', 'video', 'audio'].includes(type))
    return true
  return type === 'application' && ['pdf', 'json', 'xml', 'javascript', 'xhtml+xml', 'ld+json'].includes(subtype)
}

function requiresDocumentSandbox(mimeType: string): boolean {
  return [
    'application/xhtml+xml',
    'application/xml',
    'image/svg+xml',
    'text/html',
    'text/xml',
  ].includes(mimeType.split(';')[0]!.trim().toLowerCase())
}

function contentDisposition(filename: string, inline: boolean): string {
  const encoded = encodeURIComponent(filename).replace(/'/g, '%27')
  return `${inline ? 'inline' : 'attachment'}; filename="${encoded}"; filename*=UTF-8''${encoded}`
}

function isNotModified(request: Request, etag: string, modifiedAt: Date): boolean {
  const ifNoneMatch = request.headers.get('If-None-Match')
  if (ifNoneMatch)
    return ifNoneMatch.split(',').map(value => value.trim()).some(value => value === '*' || value === etag)
  const ifModifiedSince = request.headers.get('If-Modified-Since')
  if (!ifModifiedSince)
    return false
  const timestamp = Date.parse(ifModifiedSince)
  return Number.isFinite(timestamp) && modifiedAt.getTime() < timestamp + 1000
}

function parseByteRange(value: string, size: number): { start: number, end: number } | undefined {
  if (value.includes(','))
    return undefined
  const match = /^bytes=(\d*)-(\d*)$/.exec(value.trim())
  if (!match || (!match[1] && !match[2]) || size === 0)
    return undefined

  if (!match[1]) {
    const suffixLength = Number(match[2])
    if (!Number.isSafeInteger(suffixLength) || suffixLength <= 0)
      return undefined
    return { start: Math.max(size - suffixLength, 0), end: size - 1 }
  }

  const start = Number(match[1])
  const requestedEnd = match[2] ? Number(match[2]) : size - 1
  if (!Number.isSafeInteger(start) || !Number.isSafeInteger(requestedEnd) || start < 0 || start >= size || requestedEnd < start)
    return undefined
  return { start, end: Math.min(requestedEnd, size - 1) }
}

function rangeAllowed(request: Request, etag: string, modifiedAt: Date): boolean {
  const ifRange = request.headers.get('If-Range')
  if (!ifRange)
    return true
  if (ifRange.startsWith('W/'))
    return false
  if (ifRange.startsWith('"'))
    return ifRange === etag
  const timestamp = Date.parse(ifRange)
  return Number.isFinite(timestamp) && modifiedAt.getTime() < timestamp + 1000
}

async function serveOriginalFile(request: Request, workspace: ServeWorkspace): Promise<Response> {
  if (request.method === 'OPTIONS') {
    return new Response(null, {
      status: 204,
      headers: {
        'Access-Control-Allow-Headers': 'Range, If-None-Match, If-Modified-Since, If-Range',
        'Access-Control-Allow-Methods': 'GET, HEAD, OPTIONS',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Max-Age': '86400',
      },
    })
  }
  if (!['GET', 'HEAD'].includes(request.method))
    return error('METHOD_NOT_ALLOWED', 'Use GET or HEAD', 405)
  const url = new URL(request.url)
  let relativePath: string
  try {
    relativePath = requestFilePath(url.pathname)
  }
  catch (cause) {
    return workspaceError(cause)
  }

  try {
    const file = await workspace.openFile(relativePath)
    const forceDownload = url.searchParams.get('download') === '1'
    const etag = `W/"${file.size}-${file.modifiedAt.getTime()}"`
    const headers = new Headers({
      'Accept-Ranges': 'bytes',
      'Access-Control-Allow-Origin': '*',
      'Cache-Control': 'no-cache',
      'Content-Disposition': contentDisposition(file.name, !forceDownload && inlineMimeType(file.mimeType)),
      'Content-Length': String(file.size),
      'Content-Type': file.mimeType,
      'ETag': etag,
      'Last-Modified': file.modifiedAt.toUTCString(),
      'Referrer-Policy': 'no-referrer',
      'X-Content-Type-Options': 'nosniff',
    })
    if (requiresDocumentSandbox(file.mimeType))
      headers.set('Content-Security-Policy', ACTIVE_FILE_CONTENT_SECURITY_POLICY)

    if (isNotModified(request, etag, file.modifiedAt)) {
      headers.delete('Content-Length')
      return new Response(null, { status: 304, headers })
    }

    const rangeHeader = request.headers.get('Range')
    const useRange = rangeHeader && rangeAllowed(request, etag, file.modifiedAt)
    if (useRange) {
      const range = parseByteRange(rangeHeader, file.size)
      if (!range) {
        headers.set('Content-Range', `bytes */${file.size}`)
        headers.delete('Content-Length')
        return new Response(null, { status: 416, headers })
      }
      headers.set('Content-Range', `bytes ${range.start}-${range.end}/${file.size}`)
      headers.set('Content-Length', String(range.end - range.start + 1))
      return new Response(request.method === 'HEAD' ? null : file.body.slice(range.start, range.end + 1), {
        status: 206,
        headers,
      })
    }
    const body = request.method === 'HEAD'
      ? null
      : rangeHeader ? file.body.stream() : file.body
    return new Response(body, { headers })
  }
  catch (cause) {
    if (cause instanceof ServeWorkspaceError && cause.code === 'NOT_FILE') {
      try {
        const listing = await workspace.listDirectory(relativePath)
        const location = listing.path ? `/browse/${encodedPath(listing.path)}` : '/'
        return new Response(null, { status: 302, headers: { Location: location } })
      }
      catch (directoryCause) {
        return workspaceError(directoryCause)
      }
    }
    return workspaceError(cause)
  }
}

export function startServeHttpServer(options: {
  workspace: ServeWorkspace
  address: string
  port: number
  managementEnabled: boolean
}): RunningServeServer {
  configureWebBundleHeaders(serveWebApp)
  // Bun accepts an HTMLBundle for a method route, though its ambient type only lists it for whole-route values.
  const appRoute = {
    GET: serveWebApp as unknown as Response,
    POST: () => error('METHOD_NOT_ALLOWED', 'Use GET', 405),
    PUT: () => error('METHOD_NOT_ALLOWED', 'Use GET', 405),
    DELETE: () => error('METHOD_NOT_ALLOWED', 'Use GET', 405),
    PATCH: () => error('METHOD_NOT_ALLOWED', 'Use GET', 405),
    HEAD: () => error('METHOD_NOT_ALLOWED', 'Use GET', 405),
    OPTIONS: () => error('METHOD_NOT_ALLOWED', 'Use GET', 405),
  }
  let finish: (() => void) | undefined
  const finished = new Promise<void>((resolve) => {
    finish = resolve
  })
  const server = Bun.serve({
    hostname: options.address,
    port: options.port,
    maxRequestBodySize: MAX_UPLOAD_BYTES + 1024 * 1024,
    routes: {
      '/': appRoute,
      '/browse': appRoute,
      '/browse/*': appRoute,
    },
    async fetch(request) {
      const url = new URL(request.url)
      if (url.pathname === '/api/directory') {
        const invalidMethod = requireMethod(request, 'GET')
        if (invalidMethod)
          return invalidMethod
        try {
          const listing = await options.workspace.listDirectory(url.searchParams.get('path') ?? '')
          return json({
            version: 1,
            ...listing,
            managementEnabled: options.managementEnabled,
            maxUploadBytes: MAX_UPLOAD_BYTES,
          })
        }
        catch (cause) {
          return workspaceError(cause)
        }
      }
      if (url.pathname === '/api/text') {
        const invalidMethod = requireMethod(request, 'GET')
        if (invalidMethod)
          return invalidMethod
        try {
          return json({
            version: 1,
            ...await options.workspace.readTextPreview(url.searchParams.get('path') ?? ''),
          })
        }
        catch (cause) {
          return workspaceError(cause)
        }
      }
      if (url.pathname === '/api/upload') {
        const invalidMethod = requireMethod(request, 'POST')
        if (invalidMethod)
          return invalidMethod
        if (!options.managementEnabled)
          return error('MANAGEMENT_DISABLED', 'Start serve with --manage to enable filesystem management', 403)
        const invalidOrigin = validateMutationOrigin(request, options.address)
        if (invalidOrigin)
          return invalidOrigin
        if (request.headers.get('Content-Type')?.split(';')[0]?.trim().toLowerCase() !== 'multipart/form-data')
          return error('UNSUPPORTED_MEDIA_TYPE', 'Upload requests must use multipart form data', 415)
        let file: FormDataEntryValue | null
        try {
          file = (await request.formData()).get('file')
        }
        catch {
          return error('INVALID_UPLOAD', 'Request body must be multipart form data', 400)
        }
        if (!(file instanceof File))
          return error('INVALID_UPLOAD', 'A file field is required', 400)
        try {
          return json({
            version: 1,
            ...await options.workspace.uploadFile(url.searchParams.get('path') ?? '', file),
          })
        }
        catch (cause) {
          return workspaceError(cause)
        }
      }
      if (url.pathname === '/api/operations') {
        const invalidMethod = requireMethod(request, 'POST')
        if (invalidMethod)
          return invalidMethod
        if (!options.managementEnabled)
          return error('MANAGEMENT_DISABLED', 'Start serve with --manage to enable filesystem management', 403)
        const invalidOrigin = validateMutationOrigin(request, options.address)
        if (invalidOrigin)
          return invalidOrigin
        if (request.headers.get('Content-Type')?.split(';')[0]?.trim().toLowerCase() !== 'application/json')
          return error('UNSUPPORTED_MEDIA_TYPE', 'Filesystem operations must use JSON', 415)
        let body: unknown
        try {
          body = await request.json()
        }
        catch {
          return error('INVALID_OPERATION', 'Request body must be valid JSON', 400)
        }
        const operation = operationSchema.safeParse(body)
        if (!operation.success)
          return error('INVALID_OPERATION', 'Filesystem operation is invalid', 400)
        return json({
          version: 1,
          ...await options.workspace.applyOperation(operation.data),
        })
      }
      if (url.pathname === '/files' || url.pathname.startsWith('/files/'))
        return serveOriginalFile(request, options.workspace)
      return error('NOT_FOUND', 'Route not found', 404)
    },
  })

  return {
    url: new URL(server.url),
    finished,
    async stop() {
      await server.stop(true)
      finish?.()
    },
  }
}
