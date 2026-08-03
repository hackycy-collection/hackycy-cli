import type { NetworkInterfaceInfo } from 'node:os'
import type { ServeOptions } from './types'
import os from 'node:os'
import path from 'node:path'
import process from 'node:process'
import { cancel, intro, note, outro } from '@clack/prompts'
import ansis from 'ansis'
import { printTitle } from '../../shared/utils'
import { createServeAuthentication } from './authentication'
import { startServeHttpServer } from './server'
import { createServeWorkspace } from './workspace'

export function formatServeUrls(
  address: string,
  port: string,
  interfaces: NodeJS.Dict<NetworkInterfaceInfo[]> = os.networkInterfaces(),
): Array<{ label: 'Local' | 'Network', url: string }> {
  if (address !== '0.0.0.0')
    return [{ label: 'Local', url: `http://${address}:${port}` }]

  const urls: Array<{ label: 'Local' | 'Network', url: string }> = [
    { label: 'Local', url: `http://localhost:${port}` },
  ]
  for (const networks of Object.values(interfaces)) {
    if (!networks)
      continue
    for (const network of networks) {
      if (network.family === 'IPv4' && !network.internal)
        urls.push({ label: 'Network', url: `http://${network.address}:${port}` })
    }
  }
  return urls
}

export async function runServeCommand(options: ServeOptions): Promise<void> {
  printTitle()
  intro(ansis.bold('Static File Server'))

  let workspace: Awaited<ReturnType<typeof createServeWorkspace>>
  try {
    workspace = await createServeWorkspace(options.directory)
  }
  catch (cause) {
    cancel(cause instanceof Error ? cause.message : String(cause))
    return
  }

  let authentication: Awaited<ReturnType<typeof createServeAuthentication>>
  try {
    authentication = await createServeAuthentication(options.accounts)
  }
  catch (cause) {
    cancel(`Invalid account configuration: ${cause instanceof Error ? cause.message : String(cause)}`)
    return
  }

  let server: ReturnType<typeof startServeHttpServer>
  try {
    server = startServeHttpServer({
      workspace,
      address: options.address,
      port: options.port,
      managementEnabled: options.manage,
      safeHtml: options.safeHtml,
      authentication,
    })
  }
  catch (cause) {
    authentication?.close()
    cancel(`Failed to start server: ${cause instanceof Error ? cause.message : String(cause)}`)
    return
  }

  const urls = formatServeUrls(options.address, server.url.port)
  const messages = urls.map(({ label, url }) => `  ${ansis.dim(label.padEnd(9))} ${ansis.cyan(url)}`)
  messages.push(`  ${ansis.dim('Directory'.padEnd(9))} ${ansis.dim(path.resolve(options.directory))}`)
  messages.push(`  ${ansis.dim('Bind'.padEnd(9))} ${ansis.dim(`${options.address}:${server.url.port}`)}`)
  messages.push(`  ${ansis.dim('Management'.padEnd(11))} ${options.manage ? ansis.green('enabled') : ansis.dim('disabled')}`)
  messages.push(`  ${ansis.dim('HTML execution'.padEnd(15))} ${options.safeHtml ? ansis.dim('disabled (download)') : ansis.green('enabled')}`)
  messages.push(`  ${ansis.dim('Authentication'.padEnd(15))} ${authentication ? ansis.green(`enabled (${authentication.accountCount} ${authentication.accountCount === 1 ? 'account' : 'accounts'})`) : ansis.dim('disabled')}`)
  note(messages.join('\n'), 'Server running')

  let stopping = false
  const stop = async (): Promise<void> => {
    if (stopping)
      return
    stopping = true
    await server.stop()
    outro('Server stopped.')
  }
  process.once('SIGINT', stop)
  process.once('SIGTERM', stop)

  try {
    await server.finished
  }
  finally {
    process.removeListener('SIGINT', stop)
    process.removeListener('SIGTERM', stop)
  }
}
