import type { Command } from 'commander'

export function register(program: Command): void {
  program
    .command('task [name]')
    .description('Run a configured command group')
    .option('-d, --delay <seconds>', 'Delay between commands in seconds', '2')
    .action(async (name?: string, options?: { delay?: string }) => {
      const { runTask } = await import('./task')
      await runTask(name, options)
    })
}
