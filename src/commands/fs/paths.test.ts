import { describe, expect, test } from 'bun:test'
import { defaultFsSessionDirectory, resolveFsSessionOptions } from './paths'

describe('fs session paths', () => {
  test('uses CLI values before environment values and isolates default state by root', () => {
    const env = { HOME: '/Users/test', YCY_FS_SESSION_DIR: '/env-sessions', YCY_FS_SESSION_IDLE_DAYS: '3' }

    expect(resolveFsSessionOptions({ sessionDir: '/cli-sessions', sessionIdleDays: 2 }, '/workspace', env)).toEqual({
      directory: '/cli-sessions',
      idleLifetimeMs: 2 * 24 * 60 * 60 * 1000,
    })
    expect(resolveFsSessionOptions({}, '/workspace', env)).toEqual({
      directory: '/env-sessions',
      idleLifetimeMs: 3 * 24 * 60 * 60 * 1000,
    })
    expect(defaultFsSessionDirectory('/workspace', { HOME: '/Users/test' }, 'darwin')).not.toBe(defaultFsSessionDirectory('/other-workspace', { HOME: '/Users/test' }, 'darwin'))
  })

  test('rejects an invalid idle lifetime', () => {
    expect(() => resolveFsSessionOptions({ sessionIdleDays: 0 }, '/workspace', {})).toThrow('positive integer')
  })
})
