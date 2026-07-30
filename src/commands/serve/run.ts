import type { NetworkInterfaceInfo } from 'node:os'
import type { ServeOptions } from './types'
import os from 'node:os'
import path from 'node:path'
import process from 'node:process'
import { cancel, intro, note, outro } from '@clack/prompts'
import ansis from 'ansis'
import { printTitle } from '../../shared/utils'
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

  let server: ReturnType<typeof startServeHttpServer>
  try {
    server = startServeHttpServer({
      workspace,
      address: options.address,
      port: options.port,
      uploadEnabled: options.upload,
    })
  }
  catch (cause) {
    cancel(`Failed to start server: ${cause instanceof Error ? cause.message : String(cause)}`)
    return
  }

  const urls = formatServeUrls(options.address, server.url.port)
  const messages = urls.map(({ label, url }) => `  ${ansis.dim(label.padEnd(9))} ${ansis.cyan(url)}`)
  messages.push(`  ${ansis.dim('Directory'.padEnd(9))} ${ansis.dim(path.resolve(options.directory))}`)
  messages.push(`  ${ansis.dim('Bind'.padEnd(9))} ${ansis.dim(`${options.address}:${server.url.port}`)}`)
  messages.push(`  ${ansis.dim('Upload'.padEnd(9))} ${options.upload ? ansis.green('enabled') : ansis.dim('disabled')}`)
  note(messages.join('\n'), 'Server running')

  if (options.upload && options.address === '0.0.0.0') {
    note(
      ansis.yellow('Anyone who can reach this server can upload files into the served directory.'),
      'Trusted networks only',
    )
  }

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
