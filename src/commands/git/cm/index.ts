import type { Command } from 'commander'
import type { CmOptions } from './types'

export function register(parent: Command): void {
  parent
    .command('cm')
    .description('Generate an Angular-style commit message from uncommitted changes')
    .option('--profile <name>', 'CM provider profile to use')
    .option('--lang <lang>', 'Commit message language: en or zh', 'en')
    .option('--history', 'Use recent commit subjects as style reference')
    .option('--staged', 'Only use staged changes')
    .option('--commit', 'Commit staged changes after confirmation')
    .option('--stage-all', 'With --commit, stage all changes before generating and committing')
    .option('--dry-run', 'Generate and print only')
    .option('--body', 'Allow a short commit body')
    .action(async (options: CmOptions) => {
      const { runGitCm } = await import('./run')
      await runGitCm(options)
    })
}
