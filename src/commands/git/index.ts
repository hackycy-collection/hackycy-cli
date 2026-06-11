import type { Command } from 'commander'
import { register as registerConfig } from './config'
import { register as registerFork } from './fork'
import { register as registerPulse } from './pulse'

export function register(program: Command): void {
  const git = program.command('git').description('Git utilities')
  registerPulse(git)
  registerFork(git)
  registerConfig(git)
}
