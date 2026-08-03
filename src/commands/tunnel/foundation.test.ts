import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, test } from 'bun:test'
import { readConfig, writeConfig } from '../../config/store'
import { readRememberedTunnelConnection, rememberTunnelConnection } from '../../config/tunnel'
import { DEFAULT_TUNNEL_SERVER, normalizeControlPlaneUrl, parseHostPort, parsePortRange, resolveClientConfig, resolveServerConfig } from './config'
import { ensureFrpBinary } from './frp/binary'
import { renderFrpcConfig, renderFrpsConfig } from './frp/config'
import { FRP_ARTIFACTS, resolveFrpArtifact } from './frp/manifest'
import { acquireStateDirectoryLock } from './lock'
import { clientStateDirectory, defaultServerDataDirectory, managedFrpBinaryPath } from './paths'

const temporaryDirectories: string[] = []

afterEach(async () => {
  await Promise.all(temporaryDirectories.splice(0).map(directory => rm(directory, { recursive: true, force: true })))
})

describe('tunnel configuration', () => {
  test('uses CLI, environment, then defaults and parses deployment values', () => {
    const config = resolveServerConfig({ controlPort: 7600 }, {
      HOME: '/home/test',
      YCY_TUNNEL_CONTROL_PORT: '7700',
      YCY_TUNNEL_FRP_PORT: '7100',
      YCY_TUNNEL_PORT_RANGE: '30000-30010',
      YCY_TUNNEL_ADVERTISE_FRP_ADDR: '[2001:db8::1]:7001',
      YCY_TUNNEL_ADMIN_PASSWORD: 'environment-secret',
    })
    expect(config.controlPort).toBe(7600)
    expect(config.frpPort).toBe(7100)
    expect(config.httpPort).toBe(8080)
    expect(config.portRange).toEqual({ start: 30000, end: 30010 })
    expect(config.advertiseFrpAddress).toEqual({ host: '2001:db8::1', port: 7001 })
    expect(config.adminUser).toBe('admin')
    expect(resolveServerConfig({}, { YCY_TUNNEL_ADMIN_USER: '-operator', YCY_TUNNEL_ADMIN_PASSWORD: 'environment-secret' }).adminUser).toBe('-operator')
  })

  test('validates port ranges, host-port values, and control origins', () => {
    expect(parsePortRange('20000-20100')).toEqual({ start: 20000, end: 20100 })
    expect(() => parsePortRange('20100-20000')).toThrow('must not exceed')
    expect(parseHostPort('tunnel.example.com:7000')).toEqual({ host: 'tunnel.example.com', port: 7000 })
    expect(normalizeControlPlaneUrl('tunnel.example.com').href).toBe('https://tunnel.example.com/')
    expect(normalizeControlPlaneUrl('http://localhost:7500').href).toBe('http://localhost:7500/')
    expect(() => normalizeControlPlaneUrl('ftp://example.com')).toThrow('HTTP or HTTPS')
    const environment = { YCY_TUNNEL_ADMIN_PASSWORD: 'environment-secret' }
    expect(() => resolveServerConfig({}, {})).toThrow('YCY_TUNNEL_ADMIN_PASSWORD')
    expect(() => resolveServerConfig({}, { YCY_TUNNEL_ADMIN_PASSWORD: 'tiny' })).toThrow('5-256')
    expect(resolveServerConfig({}, { YCY_TUNNEL_ADMIN_PASSWORD: '12345' }).adminPassword).toBe('12345')
    expect(() => resolveServerConfig({}, { YCY_TUNNEL_ADMIN_USER: ' admin ', YCY_TUNNEL_ADMIN_PASSWORD: 'environment-secret' })).toThrow('1-64')
    expect(() => resolveServerConfig({ controlPort: 7000, frpPort: 7000 }, environment)).toThrow('must be distinct')
    expect(() => resolveServerConfig({ controlPort: 20000 }, environment)).toThrow('must not include')
  })

  test('resolves Client Token precedence including a secret file', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-config-'))
    temporaryDirectories.push(root)
    const secret = path.join(root, 'token')
    await writeFile(secret, ' file-token\n')
    const fromFile = await resolveClientConfig({ server: 'localhost' }, { HOME: root, YCY_TUNNEL_TOKEN_FILE: secret })
    const fromEnvironment = await resolveClientConfig({ server: 'localhost' }, { HOME: root, YCY_TUNNEL_TOKEN: 'env-token', YCY_TUNNEL_TOKEN_FILE: secret })
    const fromCli = await resolveClientConfig({ server: 'localhost', token: 'cli-token' }, { HOME: root, YCY_TUNNEL_TOKEN: 'env-token' })
    expect(fromFile.config.token).toBe('file-token')
    expect(fromEnvironment.config.token).toBe('env-token')
    expect(fromCli.config.token).toBe('cli-token')
    expect(fromFile.rememberOnAuthentication).toBe(false)
    expect(fromEnvironment.rememberOnAuthentication).toBe(false)
    expect(fromCli.rememberOnAuthentication).toBe(true)
  })

  test('encrypts and replaces one remembered server-token pair without losing other configuration', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-memory-'))
    temporaryDirectories.push(root)
    const env = { HOME: root }
    const config = await readConfig(env)
    config.cm = { defaultProfile: 'work', profiles: {} }
    config.fork.instances.github = { host: 'github.com', type: 'github', token: 'existing-encrypted-token' }
    await writeConfig(config, env)

    await rememberTunnelConnection({ server: new URL('https://first.example.com/'), token: 'first-token' }, env)
    const raw = JSON.parse(await readFile(path.join(root, '.ycy-cli', 'config.json'), 'utf8'))
    expect(raw.tunnel.server).toBe('https://first.example.com')
    expect(raw.tunnel.token).not.toContain('first-token')
    expect(raw.cm.defaultProfile).toBe('work')
    expect(raw.fork.instances.github.host).toBe('github.com')
    expect(await readRememberedTunnelConnection(env)).toEqual({ server: 'https://first.example.com', token: 'first-token' })

    await rememberTunnelConnection({ server: new URL('http://second.example.com:7500/'), token: 'second-token' }, env)
    expect(await readRememberedTunnelConnection(env)).toEqual({ server: 'http://second.example.com:7500', token: 'second-token' })

    const corrupted = await readConfig(env)
    corrupted.tunnel!.token = 'invalid-ciphertext'
    await writeConfig(corrupted, env)
    expect(await readRememberedTunnelConnection(env)).toBeUndefined()
    const replacement = await resolveClientConfig({ server: 'replacement.example.com', token: 'replacement-token' }, env)
    expect(replacement.rememberOnAuthentication).toBe(true)
  })

  test('uses a remembered pair only for its normalized server', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-resolution-'))
    temporaryDirectories.push(root)
    const env = { HOME: root }
    await rememberTunnelConnection({ server: new URL('https://remembered.example.com/'), token: 'remembered-token' }, env)

    const remembered = await resolveClientConfig({}, env)
    const sameServer = await resolveClientConfig({ server: 'remembered.example.com' }, env)
    expect(remembered.config.server.origin).toBe('https://remembered.example.com')
    expect(remembered.config.token).toBe('remembered-token')
    expect(sameServer.config.token).toBe('remembered-token')
    await expect(resolveClientConfig({ server: 'other.example.com' }, env)).rejects.toThrow('matching remembered connection')
  })

  test('uses the code default after environment and memory, and remembers only a CLI token', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-default-'))
    temporaryDirectories.push(root)
    const env = { HOME: root }
    expect(DEFAULT_TUNNEL_SERVER).toBe('')
    await expect(resolveClientConfig({}, env)).rejects.toThrow('DEFAULT_TUNNEL_SERVER')

    const fromDefault = await resolveClientConfig({ token: 'cli-token' }, env, 'default.example.com')
    expect(fromDefault.config.server.origin).toBe('https://default.example.com')
    expect(fromDefault.rememberOnAuthentication).toBe(true)

    const environmentServer = await resolveClientConfig(
      { token: 'cli-token' },
      { HOME: root, YCY_TUNNEL_SERVER: 'environment.example.com' },
      'default.example.com',
    )
    const cliServer = await resolveClientConfig(
      { server: 'cli.example.com', token: 'cli-token' },
      { HOME: root, YCY_TUNNEL_SERVER: 'environment.example.com' },
      'default.example.com',
    )
    expect(environmentServer.config.server.origin).toBe('https://environment.example.com')
    expect(cliServer.config.server.origin).toBe('https://cli.example.com')

    const fromEnvironment = await resolveClientConfig(
      { server: 'cli.example.com' },
      { HOME: root, YCY_TUNNEL_TOKEN: 'env-token' },
      'default.example.com',
    )
    expect(fromEnvironment.config.server.origin).toBe('https://cli.example.com')
    expect(fromEnvironment.rememberOnAuthentication).toBe(false)

    await rememberTunnelConnection({ server: new URL('https://remembered.example.com/'), token: 'old-token' }, env)
    const rotated = await resolveClientConfig({ token: 'new-token' }, env, 'default.example.com')
    expect(rotated.config.server.origin).toBe('https://remembered.example.com')
    expect(rotated.rememberOnAuthentication).toBe(true)
  })

  test('uses fixed platform-specific server, client, and managed FRP paths', () => {
    expect(defaultServerDataDirectory({ HOME: '/Users/test' }, 'darwin')).toBe('/Users/test/Library/Application Support/ycy/tunnel/server')
    expect(clientStateDirectory({ HOME: '/home/test', XDG_STATE_HOME: '/state' }, 'linux')).toBe('/state/ycy/tunnel/client')
    expect(managedFrpBinaryPath('frpc', { LOCALAPPDATA: 'C:\\State' }, 'win32')).toBe('C:\\State/ycy/frp/0.70.1/frpc.exe')
    expect(managedFrpBinaryPath('frps', { YCY_TUNNEL_DOCKER: '1' }, 'linux')).toBe('/opt/ycy/frp/0.70.1/frps')
  })
})

describe('FRP foundation', () => {
  test('pins complete official artifact metadata for the native matrix', () => {
    expect(FRP_ARTIFACTS).toHaveLength(6)
    expect(FRP_ARTIFACTS.map(artifact => `${artifact.platform}/${artifact.architecture}`).sort()).toEqual([
      'darwin/arm64',
      'darwin/x64',
      'linux/arm64',
      'linux/x64',
      'win32/arm64',
      'win32/x64',
    ])
    expect(resolveFrpArtifact('linux', 'arm64')).toMatchObject({ version: '0.70.1', archive: 'frp_0.70.1_linux_arm64.tar.gz' })
    for (const artifact of FRP_ARTIFACTS) {
      expect(artifact.url).toBe(`https://github.com/fatedier/frp/releases/download/v${artifact.version}/${artifact.archive}`)
      expect(artifact.sha256).toHaveLength(64)
      expect(artifact.frpcSha256).toHaveLength(64)
      expect(artifact.frpsSha256).toHaveLength(64)
    }
  })

  test('renders exact server and enabled client proxy TOML', () => {
    const server = resolveServerConfig(
      { address: '127.0.0.1', frpPort: 7001, httpPort: 8081, portRange: '21000-21005', dataDir: '/tmp/tunnel' },
      { YCY_TUNNEL_ADMIN_PASSWORD: 'environment-secret' },
    )
    expect(renderFrpsConfig(server, 'internal')).toContain('allowPorts = [{ start = 21000, end = 21005 }]')
    const client = renderFrpcConfig({
      advertisedFrpHost: 'frp.example.com',
      advertisedFrpPort: 7001,
      internalFrpToken: 'internal',
      snapshot: {
        clientKey: 'client-key',
        revision: 2,
        tunnels: [
          {
            id: 'http-id',
            label: 'Ticket H5',
            protocol: 'http',
            customDomains: ['app.example.com', 'app-alt.example.com'],
            location: '/service-a',
            serverPort: null,
            localHost: '127.0.0.1',
            localPort: 3000,
            enabled: true,
            options: {
              transport: { useEncryption: true, useCompression: true, bandwidthLimit: { value: 2, unit: 'MB', mode: 'server' }, proxyProtocolVersion: 'v2' },
              healthCheck: { type: 'http', path: '/health', intervalSeconds: 10, timeoutSeconds: 3, maxFailed: 2, headers: [{ name: 'X-Probe', value: 'ycy' }] },
              http: {
                basicAuth: { username: 'operator', password: 'secret-value' },
                hostHeaderRewrite: 'internal.example.com',
                requestHeaders: [{ name: 'X-Forwarded-By', value: 'ycy' }],
                responseHeaders: [{ name: 'X-Tunnel', value: 'ticket' }],
              },
            },
            createdAt: '',
            updatedAt: '',
          },
          { id: 'tcp-id', label: '', protocol: 'tcp', serverPort: 21000, localHost: 'db', localPort: 5432, enabled: true, options: { transport: { useEncryption: false, useCompression: false, bandwidthLimit: null, proxyProtocolVersion: null }, healthCheck: null, http: null }, createdAt: '', updatedAt: '' },
          { id: 'off-id', label: '', protocol: 'udp', serverPort: 21001, localHost: 'dns', localPort: 53, enabled: false, options: { transport: { useEncryption: false, useCompression: false, bandwidthLimit: null, proxyProtocolVersion: null }, healthCheck: null, http: null }, createdAt: '', updatedAt: '' },
        ],
      },
    })
    expect(client).toContain('customDomains = ["app.example.com", "app-alt.example.com"]')
    expect(client).toContain('locations = ["/service-a"]')
    expect(client).toContain('transport.bandwidthLimit = "2MB"')
    expect(client).toContain('transport.bandwidthLimitMode = "server"')
    expect(client).toContain('transport.useEncryption = true')
    expect(client).toContain('transport.useCompression = true')
    expect(client).toContain('transport.proxyProtocolVersion = "v2"')
    expect(client).toContain('httpUser = "operator"')
    expect(client).toContain('httpPassword = "secret-value"')
    expect(client).toContain('hostHeaderRewrite = "internal.example.com"')
    expect(client).toContain('requestHeaders.set."X-Forwarded-By" = "ycy"')
    expect(client).toContain('responseHeaders.set."X-Tunnel" = "ticket"')
    expect(client).toContain('healthCheck.httpHeaders = [{ name = "X-Probe", value = "ycy" }]')
    expect(client).toContain('remotePort = 21000')
    expect(client).not.toContain('off-id')
    expect(client).not.toContain('21001')
  })

  test('rejects an archive with the wrong SHA and prints manual placement details', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-frp-'))
    temporaryDirectories.push(root)
    const installation = ensureFrpBinary('frpc', {
      env: { HOME: root },
      fetch: (async () => new Response('not an official archive')) as unknown as typeof globalThis.fetch,
      verifyVersion: async () => {},
    })
    await expect(installation).rejects.toThrow('failed SHA-256 verification')
    await expect(installation).rejects.toThrow('Official archive:')
  })

  test('passes supervisor cancellation to the FRP download', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-frp-abort-'))
    temporaryDirectories.push(root)
    const cancellation = new AbortController()
    cancellation.abort()
    let receivedSignal: AbortSignal | null | undefined
    const installation = ensureFrpBinary('frps', {
      env: { HOME: root },
      signal: cancellation.signal,
      fetch: (async (_input: string | URL | Request, init?: RequestInit) => {
        receivedSignal = init?.signal
        throw new Error('download aborted')
      }) as unknown as typeof globalThis.fetch,
      verifyVersion: async () => {},
    })
    await expect(installation).rejects.toThrow('download aborted')
    expect(receivedSignal).toBe(cancellation.signal)
  })
})

describe('single-instance lock', () => {
  test('rejects a second owner and releases the state directory', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-lock-'))
    temporaryDirectories.push(root)
    const first = await acquireStateDirectoryLock(root)
    await expect(acquireStateDirectoryLock(root)).rejects.toThrow(`process ${process.pid}`)
    await first.release()
    const second = await acquireStateDirectoryLock(root)
    await second.release()
  })

  test('removes a stale owner before acquiring the lock', async () => {
    const root = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-stale-lock-'))
    temporaryDirectories.push(root)
    await mkdir(path.join(root, '.lock'))
    await writeFile(path.join(root, '.lock', 'owner.json'), JSON.stringify({ id: 'stale', pid: 2147483647, startedAt: '', stateDirectory: root }))
    const lock = await acquireStateDirectoryLock(root)
    expect(lock.owner.pid).toBe(process.pid)
    await lock.release()
  })
})
