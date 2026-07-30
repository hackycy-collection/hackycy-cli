import type { RunningServeServer } from './server'
import { mkdtemp, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, test } from 'bun:test'
import { startServeHttpServer } from './server'
import { createServeWorkspace, MAX_UPLOAD_BYTES } from './workspace'

const temporaryDirectories: string[] = []
const servers: RunningServeServer[] = []

afterEach(async () => {
  await Promise.all(servers.splice(0).map(server => server.stop()))
  await Promise.all(temporaryDirectories.splice(0).map(directory => rm(directory, { recursive: true, force: true })))
})

async function startFixtureServer(uploadEnabled = false): Promise<{ server: RunningServeServer, root: string }> {
  const root = await mkdtemp(path.join(tmpdir(), 'ycy-serve-http-'))
  temporaryDirectories.push(root)
  await writeFile(path.join(root, 'hello.txt'), 'hello world')
  const workspace = await createServeWorkspace(root)
  const server = startServeHttpServer({ workspace, address: '127.0.0.1', port: 0, uploadEnabled })
  servers.push(server)
  return { server, root }
}

describe('ServeHttpServer', () => {
  test('serves a versioned directory listing with API security headers', async () => {
    const { server, root } = await startFixtureServer()

    const response = await fetch(new URL('/api/directory?path=', server.url))

    expect(response.status).toBe(200)
    expect(await response.json()).toEqual({
      version: 1,
      rootName: path.basename(root),
      path: '',
      uploadEnabled: false,
      maxUploadBytes: MAX_UPLOAD_BYTES,
      entries: [expect.objectContaining({ name: 'hello.txt', path: 'hello.txt' })],
    })
    expect(response.headers.get('cache-control')).toBe('no-store')
    expect(response.headers.get('x-content-type-options')).toBe('nosniff')
    expect(response.headers.get('referrer-policy')).toBe('no-referrer')
    expect(response.headers.get('access-control-allow-origin')).toBeNull()
  })

  test('serves original files under /files with HEAD, download, cache, and CORS semantics', async () => {
    const { server, root } = await startFixtureServer()
    await Promise.all([
      writeFile(path.join(root, 'page.html'), '<script>alert(1)</script>'),
      writeFile(path.join(root, 'vector.svg'), '<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>'),
    ])

    const file = await fetch(new URL('/files/hello.txt', server.url))
    const head = await fetch(new URL('/files/hello.txt', server.url), { method: 'HEAD' })
    const download = await fetch(new URL('/files/hello.txt?download=1', server.url))
    const options = await fetch(new URL('/files/hello.txt', server.url), { method: 'OPTIONS' })
    const directory = await fetch(new URL('/files/', server.url), { redirect: 'manual' })
    const html = await fetch(new URL('/files/page.html', server.url))
    const svg = await fetch(new URL('/files/vector.svg', server.url))

    expect(file.status).toBe(200)
    expect(await file.text()).toBe('hello world')
    expect(file.headers.get('content-type')).toBe('text/plain;charset=utf-8')
    expect(file.headers.get('content-disposition')).toStartWith('inline;')
    expect(file.headers.get('content-length')).toBe('11')
    expect(file.headers.get('last-modified')).not.toBeNull()
    expect(file.headers.get('etag')).not.toBeNull()
    expect(file.headers.get('access-control-allow-origin')).toBe('*')
    expect(file.headers.get('accept-ranges')).toBe('bytes')
    expect(head.status).toBe(200)
    expect(head.headers.get('content-length')).toBe('11')
    expect(await head.text()).toBe('')
    expect(download.headers.get('content-disposition')).toStartWith('attachment;')
    expect(options.status).toBe(204)
    expect(options.headers.get('access-control-allow-origin')).toBe('*')
    expect(options.headers.get('access-control-allow-methods')).toBe('GET, HEAD, OPTIONS')
    expect(directory.status).toBe(302)
    expect(directory.headers.get('location')).toBe('/')
    expect(html.headers.get('content-security-policy')).toContain('sandbox')
    expect(svg.headers.get('content-security-policy')).toContain('sandbox')
  })

  test('supports conditional requests and one byte range', async () => {
    const { server } = await startFixtureServer()
    const initial = await fetch(new URL('/files/hello.txt', server.url))
    const etag = initial.headers.get('etag')!

    const cached = await fetch(new URL('/files/hello.txt', server.url), {
      headers: { 'If-None-Match': etag },
    })
    const range = await fetch(new URL('/files/hello.txt', server.url), {
      headers: { Range: 'bytes=0-4' },
    })
    const suffix = await fetch(new URL('/files/hello.txt', server.url), {
      headers: { Range: 'bytes=-5' },
    })
    const invalid = await fetch(new URL('/files/hello.txt', server.url), {
      headers: { Range: 'bytes=99-120' },
    })
    const multiple = await fetch(new URL('/files/hello.txt', server.url), {
      headers: { Range: 'bytes=0-1,4-5' },
    })
    const weakIfRange = await fetch(new URL('/files/hello.txt', server.url), {
      headers: { 'Range': 'bytes=0-4', 'If-Range': etag },
    })

    expect(cached.status).toBe(304)
    expect(await cached.text()).toBe('')
    expect(range.status).toBe(206)
    expect(range.headers.get('content-range')).toBe('bytes 0-4/11')
    expect(range.headers.get('content-length')).toBe('5')
    expect(await range.text()).toBe('hello')
    expect(suffix.status).toBe(206)
    expect(await suffix.text()).toBe('world')
    expect(invalid.status).toBe(416)
    expect(invalid.headers.get('content-range')).toBe('bytes */11')
    expect(multiple.status).toBe(416)
    expect(weakIfRange.status).toBe(200)
    expect(await weakIfRange.text()).toBe('hello world')
  })

  test('serves bounded text preview results as data', async () => {
    const { server, root } = await startFixtureServer()
    await Promise.all([
      writeFile(path.join(root, 'page.html'), '<script>alert(1)</script>'),
      writeFile(path.join(root, 'bad.txt'), Uint8Array.from([0xC3, 0x28])),
    ])

    const text = await fetch(new URL('/api/text?path=page.html', server.url))
    const binary = await fetch(new URL('/api/text?path=bad.txt', server.url))
    const invalidMethod = await fetch(new URL('/api/text?path=page.html', server.url), { method: 'POST' })

    expect(await text.json()).toEqual({
      version: 1,
      status: 'ready',
      text: '<script>alert(1)</script>',
      encoding: 'utf-8',
      size: 25,
    })
    expect(text.headers.get('content-type')).toContain('application/json')
    expect(await binary.json()).toEqual({ version: 1, status: 'binary', size: 2 })
    expect(invalidMethod.status).toBe(405)
  })

  test('uploads only when enabled and requested by the bound same origin', async () => {
    const disabledFixture = await startFixtureServer(false)
    const enabledFixture = await startFixtureServer(true)
    const form = (): FormData => {
      const body = new FormData()
      body.set('file', new File(['uploaded'], 'notes.txt'))
      return body
    }

    const disabled = await fetch(new URL('/api/upload?path=', disabledFixture.server.url), {
      method: 'POST',
      headers: { Origin: disabledFixture.server.url.origin },
      body: form(),
    })
    const missingOrigin = await fetch(new URL('/api/upload?path=', enabledFixture.server.url), {
      method: 'POST',
      body: form(),
    })
    const crossOrigin = await fetch(new URL('/api/upload?path=', enabledFixture.server.url), {
      method: 'POST',
      headers: { Origin: 'https://attacker.example' },
      body: form(),
    })
    const unsupportedMediaType = await fetch(new URL('/api/upload?path=', enabledFixture.server.url), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Origin': enabledFixture.server.url.origin,
      },
      body: '{}',
    })
    const accepted = await fetch(new URL('/api/upload?path=', enabledFixture.server.url), {
      method: 'POST',
      headers: { Origin: enabledFixture.server.url.origin },
      body: form(),
    })

    expect(disabled.status).toBe(403)
    expect(await disabled.json()).toEqual(expect.objectContaining({ error: expect.objectContaining({ code: 'UPLOAD_DISABLED' }) }))
    expect(missingOrigin.status).toBe(403)
    expect(crossOrigin.status).toBe(403)
    expect(unsupportedMediaType.status).toBe(415)
    expect(await unsupportedMediaType.json()).toEqual(expect.objectContaining({
      error: expect.objectContaining({ code: 'UNSUPPORTED_MEDIA_TYPE' }),
    }))
    expect(await accepted.json()).toEqual({
      version: 1,
      filename: 'notes.txt',
      path: 'notes.txt',
      size: 8,
    })
    const listing = await fetch(new URL('/api/directory?path=', enabledFixture.server.url)).then(response => response.json())
    expect(listing.entries).toEqual(expect.arrayContaining([expect.objectContaining({ name: 'notes.txt' })]))
  })

  test('serves the embedded React shell for root and browser routes only', async () => {
    const { server } = await startFixtureServer()

    const root = await fetch(server.url)
    const emptyBrowser = await fetch(new URL('/browse', server.url))
    const nested = await fetch(new URL('/browse/docs/examples', server.url))
    const invalidMethod = await fetch(server.url, { method: 'POST' })
    const missingApi = await fetch(new URL('/api/unknown', server.url))

    expect(root.status).toBe(200)
    const html = await root.text()
    expect(html).toContain('<div id="root"></div>')
    expect(html).toContain('http-equiv="Content-Security-Policy"')
    expect(html).toContain('default-src \'self\'')
    expect(emptyBrowser.status).toBe(200)
    expect(nested.status).toBe(200)
    expect(nested.headers.get('content-type')).toContain('text/html')
    expect(invalidMethod.status).toBe(405)
    expect(await invalidMethod.json()).toEqual({
      version: 1,
      error: { code: 'METHOD_NOT_ALLOWED', message: 'Use GET' },
    })
    expect(missingApi.status).toBe(404)
    expect(await missingApi.json()).toEqual({
      version: 1,
      error: { code: 'NOT_FOUND', message: 'Route not found' },
    })
  })
})
