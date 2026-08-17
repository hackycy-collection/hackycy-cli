import { createHash } from 'node:crypto'
import path from 'node:path'
import process from 'node:process'
import { applicationStateRoot } from '../../shared/application-state'

export const DEFAULT_SESSION_IDLE_DAYS = 7

export interface FsSessionOptions {
  sessionDir?: string
  sessionIdleDays?: number
}

function positiveDays(value: string | number, label: string): number {
  const days = typeof value === 'number' ? value : Number(value)
  if (!Number.isSafeInteger(days) || days < 1)
    throw new Error(`${label} must be a positive integer number of days`)
  return days
}

export function defaultFsSessionDirectory(directory: string, env: NodeJS.ProcessEnv = process.env, platform: NodeJS.Platform = process.platform): string {
  const workspaceId = createHash('sha256').update(path.resolve(directory)).digest('hex')
  return path.join(applicationStateRoot(env, platform), 'ycy', 'fs', 'sessions', workspaceId)
}

export function resolveFsSessionOptions(input: FsSessionOptions, directory: string, env: NodeJS.ProcessEnv = process.env): { directory: string, idleLifetimeMs: number } {
  const sessionDirectory = input.sessionDir ?? env.YCY_FS_SESSION_DIR ?? defaultFsSessionDirectory(directory, env)
  const rawIdleDays = input.sessionIdleDays ?? env.YCY_FS_SESSION_IDLE_DAYS ?? DEFAULT_SESSION_IDLE_DAYS
  return {
    directory: path.resolve(sessionDirectory),
    idleLifetimeMs: positiveDays(rawIdleDays, 'File session idle lifetime') * 24 * 60 * 60 * 1000,
  }
}
