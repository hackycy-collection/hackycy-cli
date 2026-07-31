import type { Command } from 'commander'
import type { ServeOptions } from './types'
import process from 'node:process'
import { parseIntArg } from '../../shared/utils'

export function register(program: Command): void {
  program
    .command('serve [directory]')
    .description('Serve static files from a directory (defaults to current directory)')
    .option('-p, --port <number>', 'Port to serve on', parseIntArg, 1204)
    .option('-a, --address <string>', 'Address to bind to', '0.0.0.0')
    .option('-m, --manage', 'Enable uploads, remote downloads, and filesystem management', false)
    .action(async (directory: string | undefined, options: Omit<ServeOptions, 'directory'>) => {
      const { runServeCommand } = await import('./run')
      await runServeCommand({
        directory: directory ?? process.cwd(),
        ...options,
      })
    })
}
