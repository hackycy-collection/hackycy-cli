import { realpath } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'
import { startDiffHttpServer } from './server'
import { createComparisonWorkspace } from './workspace'

export interface DiffCommandOptions {
  baselineDirectory: string
  targetDirectory: string
  address: string
  port: number
  exclude: string[]
  gitignore: boolean
}

export async function runDiffCommand(options: DiffCommandOptions): Promise<void> {
  const [baselineDirectory, targetDirectory] = await Promise.all([
    realpath(path.resolve(options.baselineDirectory)),
    realpath(path.resolve(options.targetDirectory)),
  ])
  const workspace = await createComparisonWorkspace({
    baselineDirectory,
    targetDirectory,
    gitignore: options.gitignore,
    exclusions: options.exclude,
  })
  const server = startDiffHttpServer({
    workspace,
    address: options.address,
    port: options.port,
    initialRefresh: true,
  })

  console.log(`Directory diff: ${server.url}`)
  console.log(`Baseline: ${baselineDirectory}`)
  console.log(`Target:   ${targetDirectory}`)
  if (!['127.0.0.1', '::1', 'localhost'].includes(options.address)) {
    console.warn('Warning: this unauthenticated server exposes source files to every client that can reach this address.')
  }

  const stop = async (): Promise<void> => {
    await server.stop()
    process.exit(0)
  }
  process.once('SIGINT', stop)
  process.once('SIGTERM', stop)

  await server.finished
}
