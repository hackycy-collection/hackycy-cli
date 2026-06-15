import type { Command } from 'commander'
import { register as registerAi } from './ai'
import { register as registerConfig } from './config'
import { register as registerFork } from './fork'
import { register as registerPulse } from './pulse'

export function register(program: Command): void {
  const git = program.command('git').description('Git utilities')
  registerAi(git)
  registerPulse(git)
  registerFork(git)
  registerConfig(git)
}
