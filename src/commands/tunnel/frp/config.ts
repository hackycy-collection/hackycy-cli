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
    if (tunnel.protocol === 'http')
      lines.push(`customDomains = [${tomlString(tunnel.hostname!)}]`)
    else
      lines.push(`remotePort = ${tunnel.serverPort}`)
  }
  lines.push('')
  return lines.join('\n')
}
