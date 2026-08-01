import type { FrpChild } from './supervisor'
import { describe, expect, test } from 'bun:test'
import { FrpSupervisor } from './supervisor'

class FakeChild implements FrpChild {
  readonly pid: number
  readonly exited: Promise<number>
  private exit!: (code: number) => void

  constructor(pid: number) {
    this.pid = pid
    this.exited = new Promise(resolve => this.exit = resolve)
  }

  kill(): void {
    this.exit(0)
  }

  crash(code = 1): void {
    this.exit(code)
  }
}

describe('FrpSupervisor', () => {
  test('owns at most one child and suppresses recovery after manual stop', async () => {
    const children: FakeChild[] = []
    const supervisor = new FrpSupervisor({
      binaryPath: '/frpc',
      role: 'frpc',
      backoffMs: [1],
      stableAfterMs: 1000,
      spawn: () => {
        const child = new FakeChild(children.length + 1)
        children.push(child)
        return child
      },
    })
    await supervisor.start('/config')
    await supervisor.start('/config')
    expect(children).toHaveLength(1)
    children[0]!.crash()
    await Bun.sleep(10)
    expect(children).toHaveLength(2)
    await supervisor.stop()
    await Bun.sleep(10)
    expect(children).toHaveLength(2)
    expect(supervisor.state().state).toBe('stopped')
  })

  test('does not restart deterministic configuration failures', async () => {
    const children: FakeChild[] = []
    const supervisor = new FrpSupervisor({ binaryPath: '/frpc', role: 'frpc', backoffMs: [1], spawn: () => {
      const child = new FakeChild(children.length + 1)
      children.push(child)
      return child
    } })
    await supervisor.start('/config')
    await supervisor.configurationFailed({ code: 'INVALID', message: 'bad config' })
    await Bun.sleep(5)
    expect(children).toHaveLength(1)
    expect(supervisor.state()).toMatchObject({ state: 'configuration_failed', error: { code: 'INVALID' } })
  })

  test('rejects a child that exits during the activation grace period', async () => {
    const children: FakeChild[] = []
    const supervisor = new FrpSupervisor({
      binaryPath: '/frpc',
      role: 'frpc',
      backoffMs: [1],
      activationGraceMs: 10,
      spawn: () => {
        const child = new FakeChild(children.length + 1)
        children.push(child)
        queueMicrotask(() => child.crash(1))
        return child
      },
    })

    await expect(supervisor.start('/candidate')).rejects.toThrow('exited with code 1 during startup')
    await Bun.sleep(15)
    expect(children).toHaveLength(1)
    expect(supervisor.state().state).toBe('stopped')
  })
})
