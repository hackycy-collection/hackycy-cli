import { randomUUID } from 'node:crypto'
import { mkdir, readFile, rename, rm, writeFile } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'
import { TunnelError } from './types'

interface LockOwner {
  id: string
  pid: number
  startedAt: string
  stateDirectory: string
}

export interface StateDirectoryLock {
  owner: LockOwner
  release: () => Promise<void>
}

function processExists(pid: number): boolean {
  if (!Number.isSafeInteger(pid) || pid <= 0)
    return false
  try {
    process.kill(pid, 0)
    return true
  }
  catch (cause) {
    return (cause as NodeJS.ErrnoException).code === 'EPERM'
  }
}

async function readOwner(lockDirectory: string): Promise<LockOwner | undefined> {
  try {
    return JSON.parse(await readFile(path.join(lockDirectory, 'owner.json'), 'utf8')) as LockOwner
  }
  catch {
    return undefined
  }
}

export async function acquireStateDirectoryLock(stateDirectory: string): Promise<StateDirectoryLock> {
  await mkdir(stateDirectory, { recursive: true })
  const lockDirectory = path.join(stateDirectory, '.lock')
  const owner: LockOwner = {
    id: randomUUID(),
    pid: process.pid,
    startedAt: new Date().toISOString(),
    stateDirectory,
  }

  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      await mkdir(lockDirectory)
      await writeFile(path.join(lockDirectory, 'owner.json'), `${JSON.stringify(owner, null, 2)}\n`, { flag: 'wx' })
      return {
        owner,
        async release() {
          const current = await readOwner(lockDirectory)
          if (current?.id === owner.id)
            await rm(lockDirectory, { recursive: true, force: true })
        },
      }
    }
    catch (cause) {
      if ((cause as NodeJS.ErrnoException).code !== 'EEXIST')
        throw cause
      const active = await readOwner(lockDirectory)
      if (active && processExists(active.pid)) {
        throw new TunnelError(
          'INSTANCE_ACTIVE',
          `Tunnel supervisor process ${active.pid} already owns state directory ${stateDirectory}`,
        )
      }
      const staleDirectory = `${lockDirectory}.stale-${randomUUID()}`
      try {
        await rename(lockDirectory, staleDirectory)
        await rm(staleDirectory, { recursive: true, force: true })
      }
      catch (staleCause) {
        const code = (staleCause as NodeJS.ErrnoException).code
        if (code !== 'ENOENT')
          throw staleCause
      }
    }
  }
  throw new TunnelError('LOCK_UNAVAILABLE', `Could not acquire tunnel state directory ${stateDirectory}`)
}
