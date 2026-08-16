import type { Command } from 'commander'
import type { FsOptions } from './types'
import process from 'node:process'
import { parseIntArg } from '../../shared/utils'

export function register(program: Command): void {
  program
    .command('fs [directory]')
    .description('Browse files in a directory (defaults to current directory)')
    .option('-p, --port <number>', 'Port for the file browser', parseIntArg, 1204)
    .option('-a, --address <string>', 'Address to bind to', '0.0.0.0')
    .option('-m, --manage', 'Enable uploads, downloads, extraction, and filesystem management', false)
    .option('--safe-html', 'Disable HTML and XHTML execution and force downloads', false)
    .option('--account <username:password>', 'Require login with an account (repeatable)', (value, values: string[]) => [...values, value], [])
    .action(async (directory: string | undefined, options: Omit<FsOptions, 'directory' | 'accounts'> & { account: string[] }) => {
      const { runFsCommand } = await import('./run')
      const { account, ...fsOptions } = options
      await runFsCommand({
        directory: directory ?? process.cwd(),
        ...fsOptions,
        accounts: account,
      })
    })
}
