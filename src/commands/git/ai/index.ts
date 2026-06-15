import type { Command } from 'commander'
import type { CmOptions } from './types'

export function register(parent: Command): void {
  const ai = parent
    .command('ai')
    .description('AI-powered git utilities')

  ai
    .command('cm')
    .description('Generate an Angular-style commit message from uncommitted changes')
    .option('--profile <name>', 'AI profile to use')
    .option('--lang <lang>', 'Commit message language: en or zh', 'en')
    .option('--history', 'Use recent commit subjects as style reference')
    .option('--staged', 'Only use staged changes')
    .option('--commit', 'Commit staged changes after confirmation')
    .option('--stage-all', 'With --commit, stage all changes before generating and committing')
    .option('--dry-run', 'Generate and print only')
    .option('--body', 'Allow a short commit body')
    .action(async (options: CmOptions) => {
      const { runGitAiCm } = await import('./cm')
      await runGitAiCm(options)
    })

  const config = ai
    .command('config')
    .description('Manage AI provider profiles')

  config
    .command('add')
    .description('Add an OpenAI-compatible provider profile')
    .action(async () => {
      const { runAiConfigAdd } = await import('./config-command')
      await runAiConfigAdd()
    })

  config
    .command('list')
    .description('List AI provider profiles')
    .action(async () => {
      const { runAiConfigList } = await import('./config-command')
      await runAiConfigList()
    })

  config
    .command('use <profile>')
    .description('Set the default AI provider profile')
    .action(async (profile: string) => {
      const { runAiConfigUse } = await import('./config-command')
      await runAiConfigUse(profile)
    })

  config
    .command('remove <profile>')
    .description('Remove an AI provider profile')
    .action(async (profile: string) => {
      const { runAiConfigRemove } = await import('./config-command')
      await runAiConfigRemove(profile)
    })

  config
    .command('set <profile> <key> <value>')
    .description('Set an optional AI profile value')
    .action(async (profile: string, key: string, value: string) => {
      const { runAiConfigSet } = await import('./config-command')
      await runAiConfigSet(profile, key, value)
    })

  config
    .command('test [profile]')
    .description('Test an AI provider profile')
    .action(async (profile?: string) => {
      const { runAiConfigTest } = await import('./config-command')
      await runAiConfigTest(profile)
    })
}
