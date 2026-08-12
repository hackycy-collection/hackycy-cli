import type { FrpSupervisor, FrpSupervisorOptions } from '../frp/supervisor'
import type { ServerTunnelConfig } from '../types'
import process from 'node:process'
import { acquireStateDirectoryLock } from '../lock'
import { AgentGateway } from './agent-gateway'
import { TunnelControlPlane } from './control-plane'
import { openTunnelDatabase } from './database'
import { startTunnelHttpServer } from './http'
import { ManagedFrpsController } from './managed-frps'
import { TunnelManagement } from './tunnel-management'

export interface TunnelServerRunOptions {
  signal?: AbortSignal
  ensureFrpsBinary?: (signal: AbortSignal) => Promise<string>
  verifyFrpsConfiguration?: (binaryPath: string, configurationPath: string, signal: AbortSignal) => Promise<void>
  createFrpsSupervisor?: (options: FrpSupervisorOptions) => FrpSupervisor
}

export async function runTunnelServer(config: ServerTunnelConfig, options: TunnelServerRunOptions = {}): Promise<void> {
  const lock = await acquireStateDirectoryLock(config.dataDir)
  const shutdown = new AbortController()
  let database: ReturnType<typeof openTunnelDatabase> | undefined
  let gateway: AgentGateway | undefined
  let frps: ManagedFrpsController | undefined
  let server: ReturnType<typeof startTunnelHttpServer> | undefined
  let management: TunnelManagement | undefined
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
    const frpToken = config.frpToken ?? controlPlane.internalFrpToken()
    frps = new ManagedFrpsController({
      config,
      frpToken,
      signal: shutdown.signal,
      ensureBinary: options.ensureFrpsBinary,
      verifyConfiguration: options.verifyFrpsConfiguration,
      createSupervisor: options.createFrpsSupervisor,
    })
    gateway = new AgentGateway(controlPlane, config.frpPort, frpToken, config.advertiseFrpAddress, frps)
    management = await TunnelManagement.create({ database, controlPlane, gateway, frps, serverConfig: config })
    if (shutdown.signal.aborted)
      return
    server = startTunnelHttpServer({ management, gateway, address: config.address, controlPort: config.controlPort })
    console.log(`Tunnel control plane: ${server.url}`)
    console.log(`FRP bind: ${config.address}:${config.frpPort}`)
    console.log(`FRP HTTP vhost: ${config.address}:${config.httpPort}`)
    console.log(`Server Port Pool: ${config.portRange.start}-${config.portRange.end} TCP/UDP`)
    console.log(`State directory: ${config.dataDir}`)
    void frps.start().catch(() => {})
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
    management?.stop()
    gateway?.stop()
    await server?.stop()
    await frps?.stop()
    database?.close()
    await lock.release()
  }
}
