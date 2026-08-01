import { describe, expect, test } from 'bun:test'
import { TunnelError } from '../types'
import { normalizeExactHostname, TunnelControlPlane } from './control-plane'
import { TunnelDatabase } from './database'

function fixture(range = { start: 20000, end: 20002 }): { database: TunnelDatabase, controlPlane: TunnelControlPlane } {
  const database = new TunnelDatabase(':memory:')
  return { database, controlPlane: new TunnelControlPlane(database, range) }
}

describe('TunnelControlPlane', () => {
  test('normalizes exact internationalized hostnames and rejects non-hostnames', () => {
    expect(normalizeExactHostname('Example.COM.')).toBe('example.com')
    expect(normalizeExactHostname('例子.测试')).toBe('xn--fsqu00a.xn--0zwm56d')
    for (const invalid of ['https://example.com', '*.example.com', 'example.com/path', '127.0.0.1', 'localhost', 'example.com:80'])
      expect(() => normalizeExactHostname(invalid)).toThrow(TunnelError)
  })

  test('keeps Client Tokens recoverable and preserves tunnels across rotation', () => {
    const { database, controlPlane } = fixture()
    const client = controlPlane.createClient('  Office Mac  ')
    expect(client.remark).toBe('Office Mac')
    expect(controlPlane.createClient().remark).toBe('')
    expect(controlPlane.createClient('   ').remark).toBe('')
    expect(controlPlane.createClient('  line one\nline two  ').remark).toBe('line one\nline two')
    expect(() => controlPlane.createClient('x'.repeat(101))).toThrow('Client Remark')
    expect(controlPlane.updateClientRemark(client.id, 'Office\ngateway').remark).toBe('Office\ngateway')
    expect(controlPlane.updateClientRemark(client.id, '').remark).toBe('')
    expect(controlPlane.getClient(client.id).token).toBe(client.token)
    const tunnel = controlPlane.createTunnel(client.id, { protocol: 'http', hostname: 'app.example.com', localPort: 3000 })
    const rotated = controlPlane.rotateClientToken(client.id)
    expect(rotated.token).not.toBe(client.token)
    expect(rotated.revocationPending).toBe(true)
    expect(controlPlane.findClientByToken(client.token)).toBeUndefined()
    expect(controlPlane.listTunnels(client.id)).toEqual([tunnel])
    controlPlane.acknowledgeReplacementToken(client.id)
    expect(controlPlane.getClient(client.id).revocationPending).toBe(false)
    database.close()
  })

  test('reserves hostnames globally even when disabled and releases them on deletion', () => {
    const { database, controlPlane } = fixture()
    const first = controlPlane.createClient('First client')
    const second = controlPlane.createClient('Second client')
    const tunnel = controlPlane.createTunnel(first.id, { protocol: 'http', hostname: 'APP.example.com', localPort: 3000, enabled: false })
    expect(() => controlPlane.createTunnel(second.id, { protocol: 'http', hostname: 'app.example.com', localPort: 3001 })).toThrow('already reserved')
    controlPlane.deleteTunnel(tunnel.id)
    expect(controlPlane.createTunnel(second.id, { protocol: 'http', hostname: 'app.example.com', localPort: 3001 }).hostname).toBe('app.example.com')
    database.close()
  })

  test('allocates the lowest free port independently per transport protocol', () => {
    const { database, controlPlane } = fixture()
    const first = controlPlane.createClient('First client')
    const second = controlPlane.createClient('Second client')
    const tcp = controlPlane.createTunnel(first.id, { protocol: 'tcp', localPort: 5432 })
    const udp = controlPlane.createTunnel(second.id, { protocol: 'udp', localPort: 53 })
    const nextTcp = controlPlane.createTunnel(second.id, { protocol: 'tcp', localPort: 5433 })
    expect(tcp.serverPort).toBe(20000)
    expect(udp.serverPort).toBe(20000)
    expect(nextTcp.serverPort).toBe(20001)
    expect(() => controlPlane.createTunnel(first.id, { protocol: 'tcp', serverPort: 20001, localPort: 1 })).toThrow('already reserved')
    database.close()
  })

  test('increments Desired Revision atomically for every mutation and bounds Applied Revision', () => {
    const { database, controlPlane } = fixture()
    const client = controlPlane.createClient('DNS client')
    const tunnel = controlPlane.createTunnel(client.id, { protocol: 'udp', localPort: 53 })
    expect(controlPlane.getClient(client.id).desiredRevision).toBe(1)
    controlPlane.updateTunnel(tunnel.id, { enabled: false })
    expect(controlPlane.getClient(client.id).desiredRevision).toBe(2)
    controlPlane.recordAppliedRevision(client.id, 1)
    expect(controlPlane.getClient(client.id).lastAppliedRevision).toBe(1)
    expect(() => controlPlane.recordAppliedRevision(client.id, 3)).toThrow('cannot exceed')
    controlPlane.deleteTunnel(tunnel.id)
    expect(controlPlane.getClient(client.id).desiredRevision).toBe(3)
    database.close()
  })

  test('cascade deletion frees reservations and removes the client', () => {
    const { database, controlPlane } = fixture()
    const first = controlPlane.createClient('First client')
    controlPlane.createTunnel(first.id, { protocol: 'tcp', serverPort: 20002, localPort: 22 })
    controlPlane.deleteClient(first.id)
    expect(() => controlPlane.getClient(first.id)).toThrow('not found')
    const second = controlPlane.createClient('Second client')
    expect(controlPlane.createTunnel(second.id, { protocol: 'tcp', serverPort: 20002, localPort: 22 }).serverPort).toBe(20002)
    database.close()
  })
})
