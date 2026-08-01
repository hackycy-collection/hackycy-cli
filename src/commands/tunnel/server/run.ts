import type { ServerTunnelConfig } from '../types'
import path from 'node:path'
import process from 'node:process'
import { atomicWrite } from '../atomic-file'
import { ensureFrpBinary } from '../frp/binary'
import { renderFrpsConfig, verifyFrpConfiguration } from '../frp/config'
import { FrpSupervisor } from '../frp/supervisor'
import { acquireStateDirectoryLock } from '../lock'
import { TunnelError } from '../types'
import { AgentGateway } from './agent-gateway'
import { TunnelControlPlane } from './control-plane'
import { openTunnelDatabase } from './database'
import { startTunnelHttpServer } from './http'

export interface TunnelServerRunOptions {
  signal?: AbortSignal
  ensureFrpsBinary?: (signal: AbortSignal) => Promise<string>
  verifyFrpsConfiguration?: (binaryPath: string, configurationPath: string, signal: AbortSignal) => Promise<void>
}

export async function runTunnelServer(config: ServerTunnelConfig, options: TunnelServerRunOptions = {}): Promise<void> {
  const lock = await acquireStateDirectoryLock(config.dataDir)
  const shutdown = new AbortController()
  let database: ReturnType<typeof openTunnelDatabase> | undefined
  let gateway: AgentGateway | undefined
  let frps: FrpSupervisor | undefined
  let server: ReturnType<typeof startTunnelHttpServer> | undefined
  const stop = (): void => {
    shutdown.abort()
    gateway?.stop()
    void server?.stop()
  }
  const usesProcessSignals = !options.signal
  if (options.signal?.aborted)
    stop()
  else
    options.signal?.addEventListener('abort', stop, { once: true })
  if (usesProcessSignals) {
    process.once('SIGINT', stop)
    process.once('SIGTERM', stop)
  }

  try {
    if (shutdown.signal.aborted)
      return
    database = openTunnelDatabase(config.dataDir)
    const controlPlane = new TunnelControlPlane(database, config.portRange)
    const binaryPath = await (options.ensureFrpsBinary ?? (signal => ensureFrpBinary('frps', { signal })))(shutdown.signal)
    if (shutdown.signal.aborted)
      return
    const frpsConfigPath = path.join(config.dataDir, 'frps.toml')
    await atomicWrite(frpsConfigPath, renderFrpsConfig(config, controlPlane.internalFrpToken()))
    if (shutdown.signal.aborted)
      return
    try {
      await (options.verifyFrpsConfiguration ?? ((binary, configuration, signal) => verifyFrpConfiguration(binary, configuration, { signal })))(binaryPath, frpsConfigPath, shutdown.signal)
    }
    catch (cause) {
      if (shutdown.signal.aborted)
        return
      throw new TunnelError('CONFIGURATION_FAILED', cause instanceof Error ? cause.message : String(cause))
    }
    if (shutdown.signal.aborted)
      return
    frps = new FrpSupervisor({ binaryPath, role: 'frps' })
    gateway = new AgentGateway(controlPlane, config.frpPort, config.advertiseFrpAddress)
    server = startTunnelHttpServer({ config, controlPlane, gateway, frps, frpsConfigPath })
    await frps.start(frpsConfigPath)
    if (shutdown.signal.aborted)
      return
    console.log(`Tunnel control plane: ${server.url}`)
    console.log(`FRP bind: ${config.address}:${config.frpPort}`)
    console.log(`FRP HTTP vhost: ${config.address}:${config.httpPort}`)
    console.log(`Server Port Pool: ${config.portRange.start}-${config.portRange.end} TCP/UDP`)
    console.log(`State directory: ${config.dataDir}`)
    await server.finished
  }
  catch (cause) {
    if (!shutdown.signal.aborted)
      throw cause
  }
  finally {
    options.signal?.removeEventListener('abort', stop)
    if (usesProcessSignals) {
      process.removeListener('SIGINT', stop)
      process.removeListener('SIGTERM', stop)
    }
    gateway?.stop()
    await server?.stop()
    await frps?.stop()
    database?.close()
    await lock.release()
  }
}
