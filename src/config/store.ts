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

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}

function normalizeConfig(raw: unknown): AppConfig {
  const fallback = emptyConfig()
  if (!isRecord(raw))
    return fallback

  const salt = typeof raw.salt === 'string' && raw.salt ? raw.salt : fallback.salt

  const legacyInstances = isRecord(raw.instances) ? raw.instances : undefined
  const fork = isRecord(raw.fork) ? raw.fork : undefined
  const forkInstances = isRecord(fork?.instances) ? fork.instances : undefined
  const instances = forkInstances ?? legacyInstances ?? {}

  const cm = isRecord(raw.cm)
    ? raw.cm
    : isRecord(raw.ai)
      ? raw.ai
      : undefined
  return {
    salt,
    fork: {
      instances: instances as AppConfig['fork']['instances'],
    },
    cm: cm as AppConfig['cm'],
  }
}

export async function readConfig(): Promise<AppConfig> {
  const file = Bun.file(getConfigPath())
  if (!(await file.exists()))
    return emptyConfig()

  const config = normalizeConfig(await file.json())
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
