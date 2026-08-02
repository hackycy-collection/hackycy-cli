import type { AppConfig } from './types'
import { mkdir } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'
import { generateSalt, getConfigDir } from './crypto'

const CONFIG_FILE = 'config.json'

function getConfigPath(env: NodeJS.ProcessEnv): string {
  return path.join(getConfigDir(env), CONFIG_FILE)
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
  const tunnel = isRecord(raw.tunnel)
    && typeof raw.tunnel.server === 'string'
    && raw.tunnel.server
    && typeof raw.tunnel.token === 'string'
    && raw.tunnel.token
    ? { server: raw.tunnel.server, token: raw.tunnel.token }
    : undefined
  return {
    salt,
    fork: {
      instances: instances as AppConfig['fork']['instances'],
    },
    cm: cm as AppConfig['cm'],
    tunnel,
  }
}

export async function readConfig(env: NodeJS.ProcessEnv = process.env): Promise<AppConfig> {
  const file = Bun.file(getConfigPath(env))
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
    await writeConfig(config, env)

  return config
}

export async function writeConfig(config: AppConfig, env: NodeJS.ProcessEnv = process.env): Promise<void> {
  const dir = getConfigDir(env)
  await mkdir(dir, { recursive: true })
  await Bun.write(getConfigPath(env), JSON.stringify(config, null, 2))
}
