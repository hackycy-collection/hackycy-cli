import type { FrpcDesiredConfiguration, ServerTunnelConfig } from '../types'

export interface FrpConfigurationVerificationOptions {
  signal?: AbortSignal
  timeoutMs?: number
}

export async function verifyFrpConfiguration(binaryPath: string, configurationPath: string, options: FrpConfigurationVerificationOptions = {}): Promise<void> {
  options.signal?.throwIfAborted()
  const child = Bun.spawn([binaryPath, 'verify', '-c', configurationPath], { stdin: 'ignore', stdout: 'pipe', stderr: 'pipe' })
  let timedOut = false
  const timeout = setTimeout(() => {
    timedOut = true
    child.kill('SIGKILL')
  }, options.timeoutMs ?? 10_000)
  const abort = (): void => child.kill('SIGTERM')
  options.signal?.addEventListener('abort', abort, { once: true })
  try {
    const [code, stdout, stderr] = await Promise.all([child.exited, new Response(child.stdout).text(), new Response(child.stderr).text()])
    options.signal?.throwIfAborted()
    if (timedOut)
      throw new Error(`FRP configuration verification timed out after ${options.timeoutMs ?? 10_000}ms`)
    if (code !== 0)
      throw new Error((stderr || stdout || `FRP configuration verification exited with code ${code}`).trim())
  }
  finally {
    clearTimeout(timeout)
    options.signal?.removeEventListener('abort', abort)
  }
}

function tomlString(value: string): string {
  return JSON.stringify(value)
}

function renderProxyOptions(lines: string[], tunnel: FrpcDesiredConfiguration['snapshot']['tunnels'][number]): void {
  const { transport, healthCheck } = tunnel.options
  if (transport.bandwidthLimit) {
    lines.push(`transport.bandwidthLimit = ${tomlString(`${transport.bandwidthLimit.value}${transport.bandwidthLimit.unit}`)}`)
    lines.push(`transport.bandwidthLimitMode = ${tomlString(transport.bandwidthLimit.mode)}`)
  }
  if (transport.useEncryption)
    lines.push('transport.useEncryption = true')
  if (transport.useCompression)
    lines.push('transport.useCompression = true')
  if (transport.proxyProtocolVersion)
    lines.push(`transport.proxyProtocolVersion = ${tomlString(transport.proxyProtocolVersion)}`)

  if (healthCheck) {
    lines.push(`healthCheck.type = ${tomlString(healthCheck.type)}`)
    lines.push(`healthCheck.timeoutSeconds = ${healthCheck.timeoutSeconds}`)
    lines.push(`healthCheck.maxFailed = ${healthCheck.maxFailed}`)
    lines.push(`healthCheck.intervalSeconds = ${healthCheck.intervalSeconds}`)
    if (healthCheck.type === 'http') {
      lines.push(`healthCheck.path = ${tomlString(healthCheck.path)}`)
      if (healthCheck.headers.length) {
        const headers = healthCheck.headers.map(header => `{ name = ${tomlString(header.name)}, value = ${tomlString(header.value)} }`).join(', ')
        lines.push(`healthCheck.httpHeaders = [${headers}]`)
      }
    }
  }
}

export function renderFrpsConfig(config: ServerTunnelConfig, internalFrpToken: string): string {
  return [
    `bindAddr = ${tomlString(config.address)}`,
    `bindPort = ${config.frpPort}`,
    `vhostHTTPPort = ${config.httpPort}`,
    '',
    'auth.method = "token"',
    `auth.token = ${tomlString(internalFrpToken)}`,
    '',
    `allowPorts = [{ start = ${config.portRange.start}, end = ${config.portRange.end} }]`,
    '',
    'log.to = "console"',
    'log.level = "warn"',
    '',
  ].join('\n')
}

export function renderFrpcConfig(input: FrpcDesiredConfiguration): string {
  const lines = [
    `serverAddr = ${tomlString(input.advertisedFrpHost)}`,
    `serverPort = ${input.advertisedFrpPort}`,
    `user = ${tomlString(`ycy_${input.snapshot.clientKey.replace(/[^\w-]/g, '_')}`)}`,
    'loginFailExit = false',
    '',
    'auth.method = "token"',
    `auth.token = ${tomlString(input.internalFrpToken)}`,
    '',
    'log.to = "console"',
    'log.level = "warn"',
  ]

  for (const tunnel of input.snapshot.tunnels.filter(candidate => candidate.enabled)) {
    lines.push(
      '',
      '[[proxies]]',
      `name = ${tomlString(`t_${tunnel.id.replace(/[^\w-]/g, '_')}`)}`,
      `type = ${tomlString(tunnel.protocol)}`,
      `localIP = ${tomlString(tunnel.localHost)}`,
      `localPort = ${tunnel.localPort}`,
    )
    renderProxyOptions(lines, tunnel)
    if (tunnel.protocol === 'http') {
      lines.push(`customDomains = [${tunnel.customDomains.map(tomlString).join(', ')}]`)
      if (tunnel.location !== null)
        lines.push(`locations = [${tomlString(tunnel.location)}]`)
      const http = tunnel.options.http!
      if (http.basicAuth) {
        lines.push(`httpUser = ${tomlString(http.basicAuth.username)}`)
        lines.push(`httpPassword = ${tomlString(http.basicAuth.password)}`)
      }
      if (http.hostHeaderRewrite)
        lines.push(`hostHeaderRewrite = ${tomlString(http.hostHeaderRewrite)}`)
      for (const header of http.requestHeaders)
        lines.push(`requestHeaders.set.${tomlString(header.name)} = ${tomlString(header.value)}`)
      for (const header of http.responseHeaders)
        lines.push(`responseHeaders.set.${tomlString(header.name)} = ${tomlString(header.value)}`)
    }
    else {
      lines.push(`remotePort = ${tunnel.serverPort}`)
    }
  }
  lines.push('')
  return lines.join('\n')
}
