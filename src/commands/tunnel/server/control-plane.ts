import type { ClientRecord, TunnelDefinition, TunnelProtocol, TunnelSnapshot } from '../types'
import type { TunnelDatabase } from './database'
import { randomBytes, randomUUID } from 'node:crypto'
import { isIP } from 'node:net'
import { domainToASCII } from 'node:url'
import { TunnelError } from '../types'

interface ClientRow {
  internal_id: string
  owner_account_id: string
  remark: string
  token: string
  desired_revision: number
  last_applied_revision: number
  revocation_pending: number
  created_at: string
  rotated_at: string | null
}

interface TunnelRow {
  id: string
  client_internal_id: string
  protocol: TunnelProtocol
  hostname: string | null
  server_port: number | null
  local_host: string
  local_port: number
  enabled: number
  created_at: string
  updated_at: string
}

export type ControlPlaneEvent
  = | { type: 'desired_state', clientId: string, ownerAccountId: string }
    | { type: 'client_created', clientId: string, ownerAccountId: string }
    | { type: 'client_rotated', clientId: string, ownerAccountId: string }
    | { type: 'client_deleted', clientId: string, ownerAccountId: string }
    | { type: 'client_updated', clientId: string, ownerAccountId: string }

export interface TunnelMutationInput {
  protocol: TunnelProtocol
  hostname?: string | null
  serverPort?: number | null
  localHost?: string
  localPort: number
  enabled?: boolean
}

export interface TunnelPatchInput {
  protocol?: TunnelProtocol
  hostname?: string | null
  serverPort?: number | null
  localHost?: string
  localPort?: number
  enabled?: boolean
}

function clientRecord(row: ClientRow): ClientRecord {
  return {
    id: row.internal_id,
    ownerAccountId: row.owner_account_id,
    remark: row.remark,
    token: row.token,
    desiredRevision: row.desired_revision,
    lastAppliedRevision: row.last_applied_revision,
    revocationPending: row.revocation_pending === 1,
    createdAt: row.created_at,
    rotatedAt: row.rotated_at,
  }
}

function tunnelDefinition(row: TunnelRow): TunnelDefinition {
  return {
    id: row.id,
    protocol: row.protocol,
    hostname: row.hostname,
    serverPort: row.server_port,
    localHost: row.local_host,
    localPort: row.local_port,
    enabled: row.enabled === 1,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

export function normalizeExactHostname(input: string): string {
  const candidate = input.trim().replace(/\.$/, '')
  if (!candidate || candidate.length > 253 || candidate.includes('*') || candidate.includes('://') || /[\s/?#@:[\]]/.test(candidate) || isIP(candidate))
    throw new TunnelError('INVALID_HOSTNAME', 'HTTP Tunnel hostname must be one exact DNS hostname without scheme, path, port, IP address, or wildcard')
  const ascii = domainToASCII(candidate).toLowerCase()
  if (!ascii || ascii.length > 253) {
    throw new TunnelError('INVALID_HOSTNAME', 'HTTP Tunnel hostname is not a valid internationalized DNS hostname')
  }
  const labels = ascii.split('.')
  if (labels.length < 2 || labels.some(label => !label || label.length > 63 || !/^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/.test(label)))
    throw new TunnelError('INVALID_HOSTNAME', 'HTTP Tunnel hostname must contain valid DNS labels and a suffix')
  return ascii
}

function localEndpoint(localHost: string | undefined, localPort: number): { localHost: string, localPort: number } {
  const host = localHost?.trim() || '127.0.0.1'
  if (host.length > 253 || !/^[\w.:-]+$/.test(host))
    throw new TunnelError('INVALID_LOCAL_ENDPOINT', 'Local Endpoint host must be an IP address, hostname, or container service name')
  if (!Number.isSafeInteger(localPort) || localPort < 1 || localPort > 65535)
    throw new TunnelError('INVALID_LOCAL_ENDPOINT', 'Local Endpoint port must be an integer from 1 through 65535')
  return { localHost: host, localPort }
}

function protocol(value: string): TunnelProtocol {
  if (!['http', 'tcp', 'udp'].includes(value))
    throw new TunnelError('INVALID_TUNNEL', 'Tunnel protocol must be http, tcp, or udp')
  return value as TunnelProtocol
}

function token(): string {
  return `ycy_${randomBytes(32).toString('base64url')}`
}

function clientRemark(value: string): string {
  const remark = value.trim()
  if (remark.length > 100)
    throw new TunnelError('INVALID_CLIENT_REMARK', 'Client Remark must contain no more than 100 characters')
  return remark
}

function now(): string {
  return new Date().toISOString()
}

function constraintError(cause: unknown): never {
  const message = cause instanceof Error ? cause.message : String(cause)
  if (/UNIQUE constraint failed.*hostname|tunnels_unique_http_hostname/i.test(message))
    throw new TunnelError('RESOURCE_RESERVED', 'HTTP Tunnel hostname is already reserved')
  if (/UNIQUE constraint failed.*(?:protocol|server_port)|tunnels_unique_transport_port/i.test(message))
    throw new TunnelError('RESOURCE_RESERVED', 'Port Tunnel protocol and server port are already reserved')
  throw cause
}

export class TunnelControlPlane {
  private readonly listeners = new Set<(event: ControlPlaneEvent) => void>()

  constructor(
    private readonly database: TunnelDatabase,
    private readonly portPool: { start: number, end: number },
  ) {}

  subscribe(listener: (event: ControlPlaneEvent) => void): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  private emit(event: ControlPlaneEvent): void {
    for (const listener of this.listeners)
      listener(event)
  }

  private incrementDesiredRevision(clientId: string): void {
    this.database.sqlite.query('UPDATE clients SET desired_revision = desired_revision + 1 WHERE internal_id = ?').run(clientId)
  }

  internalFrpToken(): string {
    const existing = this.database.sqlite.query<{ value: string }, []>('SELECT value FROM meta WHERE key = \'internal_frp_token\'').get()
    if (existing)
      return existing.value
    const value = randomBytes(32).toString('base64url')
    this.database.sqlite.query('INSERT INTO meta(key, value) VALUES(\'internal_frp_token\', ?)').run(value)
    return value
  }

  listClients(): ClientRecord[] {
    return this.database.sqlite.query<ClientRow, []>('SELECT * FROM clients ORDER BY created_at, internal_id').all().map(clientRecord)
  }

  listClientsForOwner(ownerAccountId: string): ClientRecord[] {
    return this.database.sqlite.query<ClientRow, [string]>('SELECT * FROM clients WHERE owner_account_id = ? ORDER BY created_at, internal_id').all(ownerAccountId).map(clientRecord)
  }

  getClient(id: string): ClientRecord {
    const row = this.database.sqlite.query<ClientRow, [string]>('SELECT * FROM clients WHERE internal_id = ?').get(id)
    if (!row)
      throw new TunnelError('NOT_FOUND', 'Trusted Tunnel Client was not found')
    return clientRecord(row)
  }

  getClientForOwner(id: string, ownerAccountId: string): ClientRecord {
    const row = this.database.sqlite.query<ClientRow, [string, string]>('SELECT * FROM clients WHERE internal_id = ? AND owner_account_id = ?').get(id, ownerAccountId)
    if (!row)
      throw new TunnelError('NOT_FOUND', 'Trusted Tunnel Client was not found')
    return clientRecord(row)
  }

  findClientByToken(value: string): ClientRecord | undefined {
    const row = this.database.sqlite.query<ClientRow, [string]>('SELECT * FROM clients WHERE token = ?').get(value)
    return row ? clientRecord(row) : undefined
  }

  createClient(ownerAccountId: string, remark = ''): ClientRecord {
    const id = randomUUID()
    const createdAt = now()
    this.database.sqlite.query('INSERT INTO clients(internal_id, owner_account_id, remark, token, created_at) VALUES(?, ?, ?, ?, ?)').run(id, ownerAccountId, clientRemark(remark), token(), createdAt)
    const created = this.getClient(id)
    this.emit({ type: 'client_created', clientId: id, ownerAccountId })
    return created
  }

  updateClientRemark(id: string, remark: string): ClientRecord {
    this.getClient(id)
    this.database.sqlite.query('UPDATE clients SET remark = ? WHERE internal_id = ?').run(clientRemark(remark), id)
    const updated = this.getClient(id)
    this.emit({ type: 'client_updated', clientId: id, ownerAccountId: updated.ownerAccountId })
    return updated
  }

  rotateClientToken(id: string): ClientRecord {
    const rotate = this.database.sqlite.transaction(() => {
      this.getClient(id)
      this.database.sqlite.query('UPDATE clients SET token = ?, revocation_pending = 1, rotated_at = ? WHERE internal_id = ?').run(token(), now(), id)
      return this.getClient(id)
    })
    const rotated = rotate.immediate()
    this.emit({ type: 'client_rotated', clientId: id, ownerAccountId: rotated.ownerAccountId })
    return rotated
  }

  acknowledgeReplacementToken(id: string): void {
    const result = this.database.sqlite.query('UPDATE clients SET revocation_pending = 0 WHERE internal_id = ? AND revocation_pending = 1').run(id)
    if (result.changes > 0)
      this.emit({ type: 'client_updated', clientId: id, ownerAccountId: this.getClient(id).ownerAccountId })
  }

  deleteClient(id: string): void {
    let ownerAccountId = ''
    const remove = this.database.sqlite.transaction(() => {
      ownerAccountId = this.getClient(id).ownerAccountId
      this.database.sqlite.query('DELETE FROM clients WHERE internal_id = ?').run(id)
    })
    remove.immediate()
    this.emit({ type: 'client_deleted', clientId: id, ownerAccountId })
  }

  listTunnels(clientId: string): TunnelDefinition[] {
    this.getClient(clientId)
    return this.database.sqlite.query<TunnelRow, [string]>('SELECT * FROM tunnels WHERE client_internal_id = ? ORDER BY created_at, id').all(clientId).map(tunnelDefinition)
  }

  getTunnel(id: string): TunnelDefinition & { clientId: string } {
    const row = this.database.sqlite.query<TunnelRow, [string]>('SELECT * FROM tunnels WHERE id = ?').get(id)
    if (!row)
      throw new TunnelError('NOT_FOUND', 'Tunnel Definition was not found')
    return { ...tunnelDefinition(row), clientId: row.client_internal_id }
  }

  getTunnelForOwner(id: string, ownerAccountId: string): TunnelDefinition & { clientId: string } {
    const row = this.database.sqlite.query<TunnelRow, [string, string]>(`
      SELECT tunnels.* FROM tunnels
      JOIN clients ON clients.internal_id = tunnels.client_internal_id
      WHERE tunnels.id = ? AND clients.owner_account_id = ?
    `).get(id, ownerAccountId)
    if (!row)
      throw new TunnelError('NOT_FOUND', 'Tunnel Definition was not found')
    return { ...tunnelDefinition(row), clientId: row.client_internal_id }
  }

  snapshot(clientId: string): TunnelSnapshot {
    const client = this.getClient(clientId)
    return { clientKey: client.id, revision: client.desiredRevision, tunnels: this.listTunnels(clientId) }
  }

  private availablePort(tunnelProtocol: 'tcp' | 'udp'): number {
    const rows = this.database.sqlite.query<{ server_port: number }, [string, number, number]>(
      'SELECT server_port FROM tunnels WHERE protocol = ? AND server_port BETWEEN ? AND ? ORDER BY server_port',
    ).all(tunnelProtocol, this.portPool.start, this.portPool.end)
    const reserved = new Set(rows.map(row => row.server_port))
    for (let candidate = this.portPool.start; candidate <= this.portPool.end; candidate++) {
      if (!reserved.has(candidate))
        return candidate
    }
    throw new TunnelError('PORT_POOL_EXHAUSTED', `No ${tunnelProtocol.toUpperCase()} server port is available in ${this.portPool.start}-${this.portPool.end}`)
  }

  private values(input: TunnelMutationInput): Omit<TunnelDefinition, 'id' | 'createdAt' | 'updatedAt'> {
    const tunnelProtocol = protocol(input.protocol)
    const endpoint = localEndpoint(input.localHost, input.localPort)
    if (tunnelProtocol === 'http') {
      if (!input.hostname)
        throw new TunnelError('INVALID_HOSTNAME', 'HTTP Tunnel hostname is required')
      return { protocol: tunnelProtocol, hostname: normalizeExactHostname(input.hostname), serverPort: null, ...endpoint, enabled: input.enabled ?? true }
    }
    const selectedPort = input.serverPort == null ? this.availablePort(tunnelProtocol) : input.serverPort
    if (!Number.isSafeInteger(selectedPort) || selectedPort < this.portPool.start || selectedPort > this.portPool.end)
      throw new TunnelError('PORT_OUTSIDE_POOL', `Server port must be inside ${this.portPool.start}-${this.portPool.end}`)
    return { protocol: tunnelProtocol, hostname: null, serverPort: selectedPort, ...endpoint, enabled: input.enabled ?? true }
  }

  createTunnel(clientId: string, input: TunnelMutationInput): TunnelDefinition {
    const id = randomUUID()
    const timestamp = now()
    const create = this.database.sqlite.transaction(() => {
      this.getClient(clientId)
      const value = this.values(input)
      try {
        this.database.sqlite.query(`
          INSERT INTO tunnels(id, client_internal_id, protocol, hostname, server_port, local_host, local_port, enabled, created_at, updated_at)
          VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `).run(id, clientId, value.protocol, value.hostname, value.serverPort, value.localHost, value.localPort, value.enabled ? 1 : 0, timestamp, timestamp)
      }
      catch (cause) {
        constraintError(cause)
      }
      this.incrementDesiredRevision(clientId)
      const { clientId: _, ...created } = this.getTunnel(id)
      return created
    })
    const created = create.immediate()
    this.emit({ type: 'desired_state', clientId, ownerAccountId: this.getClient(clientId).ownerAccountId })
    return created
  }

  updateTunnel(id: string, patch: TunnelPatchInput): TunnelDefinition {
    let clientId = ''
    const update = this.database.sqlite.transaction(() => {
      const current = this.getTunnel(id)
      clientId = current.clientId
      const value = this.values({
        protocol: patch.protocol ?? current.protocol,
        hostname: patch.hostname === undefined ? current.hostname : patch.hostname,
        serverPort: patch.serverPort === undefined ? current.serverPort : patch.serverPort,
        localHost: patch.localHost ?? current.localHost,
        localPort: patch.localPort ?? current.localPort,
        enabled: patch.enabled ?? current.enabled,
      })
      try {
        this.database.sqlite.query(`
          UPDATE tunnels SET protocol = ?, hostname = ?, server_port = ?, local_host = ?, local_port = ?, enabled = ?, updated_at = ?
          WHERE id = ?
        `).run(value.protocol, value.hostname, value.serverPort, value.localHost, value.localPort, value.enabled ? 1 : 0, now(), id)
      }
      catch (cause) {
        constraintError(cause)
      }
      this.incrementDesiredRevision(clientId)
      const { clientId: _, ...changed } = this.getTunnel(id)
      return changed
    })
    const changed = update.immediate()
    this.emit({ type: 'desired_state', clientId, ownerAccountId: this.getClient(clientId).ownerAccountId })
    return changed
  }

  deleteTunnel(id: string): void {
    let clientId = ''
    let ownerAccountId = ''
    const remove = this.database.sqlite.transaction(() => {
      const current = this.getTunnel(id)
      clientId = current.clientId
      ownerAccountId = this.getClient(clientId).ownerAccountId
      this.database.sqlite.query('DELETE FROM tunnels WHERE id = ?').run(id)
      this.incrementDesiredRevision(clientId)
    })
    remove.immediate()
    this.emit({ type: 'desired_state', clientId, ownerAccountId })
  }

  recordAppliedRevision(clientId: string, revision: number): void {
    if (!Number.isSafeInteger(revision) || revision < 0)
      throw new TunnelError('INVALID_REVISION', 'Applied Revision must be a non-negative integer')
    const client = this.getClient(clientId)
    if (revision > client.desiredRevision)
      throw new TunnelError('INVALID_REVISION', 'Applied Revision cannot exceed Desired Revision')
    if (revision > client.lastAppliedRevision) {
      this.database.sqlite.query('UPDATE clients SET last_applied_revision = ? WHERE internal_id = ?').run(revision, clientId)
      this.emit({ type: 'client_updated', clientId, ownerAccountId: this.getClient(clientId).ownerAccountId })
    }
  }
}
