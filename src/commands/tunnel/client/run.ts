import type { ClientTunnelConfig } from '../types'
import process from 'node:process'
import { version } from '../../../../package.json'
import { ensureFrpBinary } from '../frp/binary'
import { FRP_ACTIVATION_GRACE_MS, FrpSupervisor } from '../frp/supervisor'
import { acquireStateDirectoryLock } from '../lock'
import { TunnelClientAgent } from './agent'
import { ClientReconciler, SupervisorClientRuntime } from './reconciler'
import { readAppliedClientState } from './state'

export interface TunnelClientRunOptions {
  signal?: AbortSignal
  ensureFrpcBinary?: (signal: AbortSignal) => Promise<string>
}

export async function runTunnelClient(config: ClientTunnelConfig, options: TunnelClientRunOptions = {}): Promise<void> {
  const lock = await acquireStateDirectoryLock(config.stateDir)
  const shutdown = new AbortController()
  let supervisor: FrpSupervisor | undefined
  let reconciler: ClientReconciler | undefined
  let agent: TunnelClientAgent | undefined

  console.log(`Tunnel client: ${config.server.origin}`)
  console.log(`State directory: ${config.stateDir}`)

  const stop = (): void => {
    shutdown.abort()
    void agent?.stop().catch(() => {})
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
    const applied = await readAppliedClientState(config.stateDir)
    if (shutdown.signal.aborted)
      return
    agent = new TunnelClientAgent({
      server: config.server,
      token: config.token,
      ycyVersion: version,
      lastAppliedRevision: applied?.revision ?? 0,
      async createReconciler() {
        if (reconciler)
          return reconciler
        const binaryPath = await (options.ensureFrpcBinary ?? (signal => ensureFrpBinary('frpc', { signal })))(shutdown.signal)
        if (shutdown.signal.aborted)
          throw shutdown.signal.reason
        supervisor = new FrpSupervisor({ binaryPath, role: 'frpc', activationGraceMs: FRP_ACTIVATION_GRACE_MS })
        const runtime = new SupervisorClientRuntime(binaryPath, supervisor)
        reconciler = new ClientReconciler(config.stateDir, runtime)
        supervisor.observe(state => agent?.reportProcessState(state.state, state.error))
        return reconciler
      },
      onMessage: message => console.log(message),
    })
    await agent.run()
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
    shutdown.abort()
    await agent?.stop()
    await lock.release()
  }
}
