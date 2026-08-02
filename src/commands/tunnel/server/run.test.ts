import type { ServerTunnelConfig } from '../types'
import { mkdtemp, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { describe, expect, test } from 'bun:test'
import { runTunnelServer } from './run'

describe('Tunnel server lifecycle', () => {
  test('cancels bootstrap and releases its state directory when shutdown is requested', async () => {
    const dataDir = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-server-run-'))
    const shutdown = new AbortController()
    let bootstrapStarted!: () => void
    const started = new Promise<void>(resolve => bootstrapStarted = resolve)
    let receivedSignal: AbortSignal | undefined
    const config: ServerTunnelConfig = {
      address: '127.0.0.1',
      controlPort: 17500,
      frpPort: 17501,
      httpPort: 17502,
      portRange: { start: 20000, end: 20002 },
      dataDir,
      adminUser: 'admin',
      adminPassword: 'admin-secret',
    }
    try {
      const running = runTunnelServer(config, {
        signal: shutdown.signal,
        ensureFrpsBinary: async (signal) => {
          receivedSignal = signal
          bootstrapStarted()
          await new Promise<void>((_resolve, reject) => signal.addEventListener('abort', () => reject(new Error('aborted')), { once: true }))
          return '/frps'
        },
      })
      await started
      shutdown.abort()
      await running
      expect(receivedSignal?.aborted).toBe(true)
      const reacquired = await Bun.file(path.join(dataDir, '.lock', 'owner.json')).exists()
      expect(reacquired).toBe(false)
    }
    finally {
      await rm(dataDir, { recursive: true, force: true })
    }
  })

  test('rejects an invalid frps configuration without entering recovery', async () => {
    const dataDir = await mkdtemp(path.join(tmpdir(), 'ycy-tunnel-server-verify-'))
    let verifications = 0
    const config: ServerTunnelConfig = {
      address: '127.0.0.1',
      controlPort: 17600,
      frpPort: 17601,
      httpPort: 17602,
      portRange: { start: 20100, end: 20102 },
      dataDir,
      adminUser: 'admin',
      adminPassword: 'admin-secret',
    }
    try {
      const running = runTunnelServer(config, {
        ensureFrpsBinary: async () => '/frps',
        verifyFrpsConfiguration: async () => {
          verifications++
          throw new Error('invalid server configuration')
        },
      })
      await expect(running).rejects.toThrow('invalid server configuration')
      expect(verifications).toBe(1)
      expect(await Bun.file(path.join(dataDir, '.lock', 'owner.json')).exists()).toBe(false)
    }
    finally {
      await rm(dataDir, { recursive: true, force: true })
    }
  })
})
