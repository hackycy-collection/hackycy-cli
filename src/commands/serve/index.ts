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
    .option('-u, --upload', 'Enable file uploads', false)
    .action(async (directory: string | undefined, options: Omit<ServeOptions, 'directory'>) => {
      const { serve } = await import('./serve')
      await serve({
        directory: directory ?? process.cwd(),
        ...options,
      })
    })
}
