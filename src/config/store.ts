import type { AppConfig } from './types'
import { mkdir } from 'node:fs/promises'
import path from 'node:path'
import { generateSalt, getConfigDir } from './crypto'

const CONFIG_FILE = 'config.json'

function getConfigPath(): string {
  return path.join(getConfigDir(), CONFIG_FILE)
}

function emptyConfig(): AppConfig {
  return {
    salt: generateSalt(),
    fork: {
      instances: {},
    },
  }
}

export async function readConfig(): Promise<AppConfig> {
  const file = Bun.file(getConfigPath())
  if (!(await file.exists()))
    return emptyConfig()

  const config: AppConfig = await file.json()
  let dirty = false

  for (const [, instance] of Object.entries(config.fork.instances)) {
    if (instance.host.includes('://')) {
      const url = new URL(instance.host)
      instance.scheme = url.protocol.slice(0, -1) as 'http' | 'https'
      instance.host = url.host
      dirty = true
    }
  }

  if (dirty)
    await writeConfig(config)

  return config
}

export async function writeConfig(config: AppConfig): Promise<void> {
  const dir = getConfigDir()
  await mkdir(dir, { recursive: true })
  await Bun.write(getConfigPath(), JSON.stringify(config, null, 2))
}
