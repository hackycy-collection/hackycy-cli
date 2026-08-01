import os from 'node:os'
import path from 'node:path'
import process from 'node:process'
import { FRP_VERSION } from './frp/manifest'

function home(env: NodeJS.ProcessEnv): string {
  return env.USERPROFILE || env.HOME || os.homedir()
}

function applicationStateRoot(env: NodeJS.ProcessEnv, platform: NodeJS.Platform = process.platform): string {
  if (platform === 'win32')
    return env.LOCALAPPDATA || path.join(home(env), 'AppData', 'Local')
  if (platform === 'darwin')
    return path.join(home(env), 'Library', 'Application Support')
  return env.XDG_STATE_HOME || path.join(home(env), '.local', 'state')
}

export function defaultServerDataDirectory(env: NodeJS.ProcessEnv = process.env, platform: NodeJS.Platform = process.platform): string {
  return path.join(applicationStateRoot(env, platform), 'ycy', 'tunnel', 'server')
}

export function clientStateDirectory(env: NodeJS.ProcessEnv = process.env, platform: NodeJS.Platform = process.platform): string {
  return path.join(applicationStateRoot(env, platform), 'ycy', 'tunnel', 'client')
}

export function managedFrpDirectory(env: NodeJS.ProcessEnv = process.env, platform: NodeJS.Platform = process.platform): string {
  if (env.YCY_TUNNEL_DOCKER === '1')
    return path.join('/opt', 'ycy', 'frp', FRP_VERSION)
  return path.join(applicationStateRoot(env, platform), 'ycy', 'frp', FRP_VERSION)
}

export function managedFrpBinaryPath(role: 'frpc' | 'frps', env: NodeJS.ProcessEnv = process.env, platform: NodeJS.Platform = process.platform): string {
  return path.join(managedFrpDirectory(env, platform), platform === 'win32' ? `${role}.exe` : role)
}
