import process from 'node:process'
import { decrypt, deriveKey, encrypt } from './crypto'
import { readConfig, writeConfig } from './store'

export interface RememberedTunnelConnection {
  server: string
  token: string
}

export async function readRememberedTunnelConnection(env: NodeJS.ProcessEnv = process.env): Promise<RememberedTunnelConnection | undefined> {
  const config = await readConfig(env)
  if (!config.tunnel)
    return undefined
  const key = await deriveKey(config.salt)
  try {
    return {
      server: config.tunnel.server,
      token: decrypt(config.tunnel.token, key),
    }
  }
  catch {
    return undefined
  }
}

export async function rememberTunnelConnection(connection: { server: URL, token: string }, env: NodeJS.ProcessEnv = process.env): Promise<void> {
  const config = await readConfig(env)
  const key = await deriveKey(config.salt)
  config.tunnel = {
    server: connection.server.origin,
    token: encrypt(connection.token, key),
  }
  await writeConfig(config, env)
}
