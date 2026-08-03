import type { FrpSupervisor } from '../frp/supervisor'
import type { AccountKind, AccountRecord, AccountRole, ClientRecord, PublicTunnelDefinition, ServerTunnelConfig, TunnelDefinition } from '../types'
import type { AgentGateway } from './agent-gateway'
import type { TunnelControlPlane, TunnelMutationInput, TunnelPatchInput } from './control-plane'
import type { TunnelDatabase } from './database'
import { randomBytes, randomUUID } from 'node:crypto'
import { redactTunnelDefinition } from '../definition'
import { TunnelError } from '../types'
import { clientView, tunnelState } from './views'

const ENVIRONMENT_ADMIN_ID = 'environment-admin'
const SESSION_LIFETIME_MS = 12 * 60 * 60 * 1000
const MAX_ACCOUNT_SESSIONS = 8
const MAX_SESSIONS = 128
const PASSWORD_ALGORITHM = { algorithm: 'argon2id', memoryCost: 65_536, timeCost: 3 } as const

interface AccountRow {
  internal_id: string
  kind: AccountKind
  username: string
  username_key: string
  role: AccountRole
  password_hash: string | null
  created_at: string
  updated_at: string
}

interface SessionEntry {
  accountId: string
  expiresAt: number
}

export interface CurrentAccount extends AccountRecord {
  managedByEnvironment: boolean
}

export interface AccountView extends CurrentAccount {
  clientCount: number
}

export interface ManagedClientView extends Omit<ReturnType<typeof clientView>, 'ownerAccountId'> {
  owner: { id: string, username: string }
}

export interface ClientDetailView {
  client: ManagedClientView
  tunnels: Array<PublicTunnelDefinition & { state: ReturnType<typeof tunnelState> }>
}

export type WorkspaceEvent = 'changed' | 'session_revoked'

export interface OverviewView {
  account: CurrentAccount
  counts: { clients: number, connected: number, tunnels: number, pending: number, errors: number }
  server?: ServerView
}

export interface ServerView {
  frps: ReturnType<FrpSupervisor['state']>
  settings: Omit<ServerTunnelConfig, 'adminPassword'>
}

export interface TunnelManagementOptions {
  database: TunnelDatabase
  controlPlane: TunnelControlPlane
  gateway: AgentGateway
  frps: FrpSupervisor
  frpsConfigPath: string
  serverConfig: ServerTunnelConfig
  sessionLifetimeMs?: number
}

function username(value: string): { display: string, key: string } {
  const display = value
  if (!/^[\w.-]{1,64}$/.test(display))
    throw new TunnelError('INVALID_ACCOUNT', 'Username must contain 1-64 ASCII letters, numbers, dots, underscores, or hyphens')
  return { display, key: display.toLowerCase() }
}

function validPassword(value: string): void {
  if (value.length < 8 || value.length > 256)
    throw new TunnelError('INVALID_ACCOUNT', 'Password must contain 8-256 characters')
}

function accountRecord(row: AccountRow): CurrentAccount {
  return {
    id: row.internal_id,
    kind: row.kind,
    username: row.username,
    role: row.role,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    managedByEnvironment: row.kind === 'environment',
  }
}

function isUniqueConstraint(cause: unknown): boolean {
  return cause instanceof Error && /UNIQUE constraint failed/i.test(cause.message)
}

export class TunnelManagement {
  private readonly sessions = new Map<string, SessionEntry>()
  private readonly observers = new Map<string, Set<(event: WorkspaceEvent) => void>>()
  private readonly unsubscribers: Array<() => void> = []
  private environmentPasswordHash = ''
  private expirationTimer: ReturnType<typeof setTimeout> | undefined

  private constructor(private readonly options: TunnelManagementOptions) {}

  static async create(options: TunnelManagementOptions): Promise<TunnelManagement> {
    const management = new TunnelManagement(options)
    await management.initializeEnvironmentAdministrator()
    management.connectEvents()
    return management
  }

  private connectEvents(): void {
    this.unsubscribers.push(
      this.options.controlPlane.subscribe(event => this.notifyResource(event.ownerAccountId)),
      this.options.gateway.observe((clientId) => {
        if (!clientId)
          return this.notifyAdministrators()
        try {
          this.notifyResource(this.options.controlPlane.getClient(clientId).ownerAccountId)
        }
        catch (cause) {
          if (!(cause instanceof TunnelError) || cause.code !== 'NOT_FOUND')
            throw cause
        }
      }),
      this.options.frps.observe(() => this.notifyAdministrators()),
    )
  }

  private async initializeEnvironmentAdministrator(): Promise<void> {
    const identity = username(this.options.serverConfig.adminUser)
    validPassword(this.options.serverConfig.adminPassword)
    this.environmentPasswordHash = await Bun.password.hash(this.options.serverConfig.adminPassword, PASSWORD_ALGORITHM)
    const timestamp = new Date().toISOString()
    try {
      this.options.database.sqlite.query(`
        INSERT INTO accounts(internal_id, kind, username, username_key, role, password_hash, created_at, updated_at)
        VALUES(?, 'environment', ?, ?, 'admin', NULL, ?, ?)
        ON CONFLICT(internal_id) DO UPDATE SET
          username = excluded.username,
          username_key = excluded.username_key,
          role = 'admin',
          password_hash = NULL,
          updated_at = excluded.updated_at
      `).run(ENVIRONMENT_ADMIN_ID, identity.display, identity.key, timestamp, timestamp)
    }
    catch (cause) {
      if (isUniqueConstraint(cause))
        throw new TunnelError('INVALID_CONFIG', 'Environment administrator username conflicts with a local account')
      throw cause
    }
  }

  private account(id: string): CurrentAccount {
    const row = this.options.database.sqlite.query<AccountRow, [string]>('SELECT * FROM accounts WHERE internal_id = ?').get(id)
    if (!row)
      throw new TunnelError('AUTHENTICATION_REQUIRED', 'Authenticated session is required')
    return accountRecord(row)
  }

  private accountByUsername(key: string): AccountRow | undefined {
    return this.options.database.sqlite.query<AccountRow, [string]>('SELECT * FROM accounts WHERE username_key = ?').get(key) ?? undefined
  }

  private clearExpiredSessions(): void {
    const timestamp = Date.now()
    for (const [token, session] of this.sessions) {
      if (session.expiresAt <= timestamp)
        this.revokeSession(token, false)
    }
    this.scheduleSessionExpiration()
  }

  private scheduleSessionExpiration(): void {
    clearTimeout(this.expirationTimer)
    this.expirationTimer = undefined
    let nextExpiration = Number.POSITIVE_INFINITY
    for (const session of this.sessions.values())
      nextExpiration = Math.min(nextExpiration, session.expiresAt)
    if (!Number.isFinite(nextExpiration))
      return
    this.expirationTimer = setTimeout(() => {
      this.expirationTimer = undefined
      this.clearExpiredSessions()
    }, Math.max(0, nextExpiration - Date.now()))
    this.expirationTimer.unref?.()
  }

  private createSession(accountId: string): { token: string, expiresAt: string, account: CurrentAccount } {
    this.clearExpiredSessions()
    for (const [token, session] of this.sessions) {
      if (session.accountId === accountId && [...this.sessions.values()].filter(value => value.accountId === accountId).length >= MAX_ACCOUNT_SESSIONS)
        this.revokeSession(token)
    }
    while (this.sessions.size >= MAX_SESSIONS)
      this.revokeSession(this.sessions.keys().next().value!)
    const token = randomBytes(32).toString('base64url')
    const expiresAt = Date.now() + (this.options.sessionLifetimeMs ?? SESSION_LIFETIME_MS)
    this.sessions.set(token, { accountId, expiresAt })
    this.scheduleSessionExpiration()
    return { token, expiresAt: new Date(expiresAt).toISOString(), account: this.account(accountId) }
  }

  async signIn(input: { username: string, password: string }): Promise<{ token: string, expiresAt: string, account: CurrentAccount }> {
    let identity: { display: string, key: string } | undefined
    try {
      identity = username(input.username)
    }
    catch {}
    const row = identity ? this.accountByUsername(identity.key) : undefined
    const hash = row?.kind === 'local' ? row.password_hash! : this.environmentPasswordHash
    const verified = await Bun.password.verify(input.password, hash)
    if (!row || !verified)
      throw new TunnelError('AUTHENTICATION_FAILED', 'Account credentials are invalid')
    return this.createSession(row.internal_id)
  }

  resume(token: string | undefined): TunnelWorkspace | undefined {
    this.clearExpiredSessions()
    const session = token ? this.sessions.get(token) : undefined
    if (!token || !session)
      return undefined
    this.sessions.delete(token)
    this.sessions.set(token, session)
    try {
      return new TunnelWorkspace(this, this.options, token, session.accountId)
    }
    catch {
      this.sessions.delete(token)
      return undefined
    }
  }

  signOut(token: string | undefined): void {
    if (token)
      this.revokeSession(token)
  }

  revokeAccountSessions(accountId: string): void {
    for (const [token, session] of this.sessions) {
      if (session.accountId === accountId)
        this.revokeSession(token)
    }
  }

  private revokeSession(token: string, scheduleExpiration = true): void {
    this.sessions.delete(token)
    const listeners = this.observers.get(token)
    if (listeners) {
      for (const listener of listeners)
        listener('session_revoked')
      this.observers.delete(token)
    }
    if (scheduleExpiration)
      this.scheduleSessionExpiration()
  }

  private notifyResource(ownerAccountId: string): void {
    for (const [token, listeners] of this.observers) {
      try {
        const account = this.activeAccount(token)
        if (account.role === 'admin' || account.id === ownerAccountId) {
          for (const listener of listeners)
            listener('changed')
        }
      }
      catch {
        this.revokeSession(token)
      }
    }
  }

  notifyAdministrators(): void {
    for (const [token, listeners] of this.observers) {
      try {
        if (this.activeAccount(token).role === 'admin') {
          for (const listener of listeners)
            listener('changed')
        }
      }
      catch {
        this.revokeSession(token)
      }
    }
  }

  observe(token: string, accountId: string, listener: (event: WorkspaceEvent) => void): () => void {
    this.activeAccount(token, accountId)
    const listeners = this.observers.get(token) ?? new Set()
    listeners.add(listener)
    this.observers.set(token, listeners)
    return () => {
      listeners.delete(listener)
      if (!listeners.size)
        this.observers.delete(token)
    }
  }

  activeAccount(token: string, expectedAccountId?: string): CurrentAccount {
    this.clearExpiredSessions()
    const session = this.sessions.get(token)
    if (!session || (expectedAccountId && session.accountId !== expectedAccountId))
      throw new TunnelError('AUTHENTICATION_REQUIRED', 'Authenticated session is required')
    return this.account(session.accountId)
  }

  stop(): void {
    for (const unsubscribe of this.unsubscribers.splice(0))
      unsubscribe()
    this.sessions.clear()
    this.observers.clear()
    clearTimeout(this.expirationTimer)
    this.expirationTimer = undefined
  }
}

export class TunnelWorkspace {
  constructor(
    private readonly management: TunnelManagement,
    private readonly options: TunnelManagementOptions,
    private readonly sessionToken: string,
    private readonly accountId: string,
  ) {}

  get account(): CurrentAccount {
    return this.management.activeAccount(this.sessionToken, this.accountId)
  }

  private requireClient(clientId: string): ClientRecord {
    return this.account.role === 'admin'
      ? this.options.controlPlane.getClient(clientId)
      : this.options.controlPlane.getClientForOwner(clientId, this.account.id)
  }

  private clientView(client: ClientRecord): ManagedClientView {
    const owner = this.options.database.sqlite.query<AccountRow, [string]>('SELECT * FROM accounts WHERE internal_id = ?').get(client.ownerAccountId)!
    const { ownerAccountId: _, ...view } = clientView(this.options.controlPlane, this.options.gateway, client)
    return { ...view, owner: { id: owner.internal_id, username: owner.username } }
  }

  listClients(): ManagedClientView[] {
    const clients = this.account.role === 'admin'
      ? this.options.controlPlane.listClients()
      : this.options.controlPlane.listClientsForOwner(this.account.id)
    return clients.map(client => this.clientView(client))
  }

  getClient(clientId: string): ClientDetailView {
    const client = this.requireClient(clientId)
    const runtime = this.options.gateway.state(client.id)
    return {
      client: this.clientView(client),
      tunnels: this.options.controlPlane.listTunnels(client.id).map(tunnel => ({ ...redactTunnelDefinition(tunnel), state: tunnelState(tunnel, client, runtime) })),
    }
  }

  createClient(input: { remark?: string }): ManagedClientView {
    return this.clientView(this.options.controlPlane.createClient(this.account.id, input.remark))
  }

  async overview(): Promise<OverviewView> {
    const account = this.account
    const clients = this.listClients()
    return {
      account,
      counts: {
        clients: clients.length,
        connected: clients.filter(client => client.runtime.connectionState === 'connected').length,
        tunnels: clients.reduce((sum, client) => sum + client.tunnelCounts.total, 0),
        pending: clients.reduce((sum, client) => sum + client.tunnelCounts.pending, 0),
        errors: clients.reduce((sum, client) => sum + client.tunnelCounts.error, 0),
      },
      ...(account.role === 'admin' ? { server: this.serverView() } : {}),
    }
  }

  private serverView(): ServerView {
    const { adminPassword: _, ...settings } = this.options.serverConfig
    return { frps: this.options.frps.state(), settings }
  }

  observe(listener: (event: WorkspaceEvent) => void): () => void {
    return this.management.observe(this.sessionToken, this.accountId, listener)
  }

  updateClientRemark(clientId: string, remark: string): ManagedClientView {
    this.requireClient(clientId)
    return this.clientView(this.options.controlPlane.updateClientRemark(clientId, remark))
  }

  rotateClientToken(clientId: string): ManagedClientView {
    this.requireClient(clientId)
    return this.clientView(this.options.controlPlane.rotateClientToken(clientId))
  }

  restartClient(clientId: string): void {
    this.requireClient(clientId)
    if (!this.options.gateway.restartClient(clientId))
      throw new TunnelError('CLIENT_OFFLINE', 'Trusted Tunnel Client is not connected')
  }

  deleteClient(clientId: string): void {
    this.requireClient(clientId)
    this.options.controlPlane.deleteClient(clientId)
  }

  createTunnel(clientId: string, input: TunnelMutationInput): PublicTunnelDefinition {
    this.requireClient(clientId)
    return redactTunnelDefinition(this.options.controlPlane.createTunnel(clientId, input))
  }

  private requireTunnel(tunnelId: string): TunnelDefinition & { clientId: string } {
    return this.account.role === 'admin'
      ? this.options.controlPlane.getTunnel(tunnelId)
      : this.options.controlPlane.getTunnelForOwner(tunnelId, this.account.id)
  }

  updateTunnel(tunnelId: string, patch: TunnelPatchInput): PublicTunnelDefinition {
    this.requireTunnel(tunnelId)
    return redactTunnelDefinition(this.options.controlPlane.updateTunnel(tunnelId, patch))
  }

  deleteTunnel(tunnelId: string): void {
    this.requireTunnel(tunnelId)
    this.options.controlPlane.deleteTunnel(tunnelId)
  }

  async changePassword(input: { currentPassword: string, newPassword: string }): Promise<void> {
    const account = this.account
    if (account.managedByEnvironment)
      throw new TunnelError('MANAGED_ACCOUNT', 'Deployment Administrator is managed by environment variables')
    validPassword(input.newPassword)
    const row = this.options.database.sqlite.query<AccountRow, [string]>('SELECT * FROM accounts WHERE internal_id = ?').get(account.id)!
    if (!await Bun.password.verify(input.currentPassword, row.password_hash!))
      throw new TunnelError('INVALID_CURRENT_PASSWORD', 'Current password is invalid')
    const passwordHash = await Bun.password.hash(input.newPassword, PASSWORD_ALGORITHM)
    this.management.activeAccount(this.sessionToken, account.id)
    this.options.database.sqlite.query('UPDATE accounts SET password_hash = ?, updated_at = ? WHERE internal_id = ?').run(passwordHash, new Date().toISOString(), account.id)
    this.management.revokeAccountSessions(account.id)
  }

  administration(): TunnelAdministration {
    if (this.account.role !== 'admin')
      throw new TunnelError('FORBIDDEN', 'Administrator role is required')
    return new TunnelAdministration(this.management, this.options, this.sessionToken, this.account.id)
  }
}

export class TunnelAdministration {
  constructor(
    private readonly management: TunnelManagement,
    private readonly options: TunnelManagementOptions,
    private readonly sessionToken: string,
    private readonly accountId: string,
  ) {}

  get account(): CurrentAccount {
    const account = this.management.activeAccount(this.sessionToken, this.accountId)
    if (account.role !== 'admin')
      throw new TunnelError('FORBIDDEN', 'Administrator role is required')
    return account
  }

  private localAccount(accountId: string): AccountRow {
    void this.account
    const row = this.options.database.sqlite.query<AccountRow, [string]>('SELECT * FROM accounts WHERE internal_id = ?').get(accountId)
    if (!row)
      throw new TunnelError('NOT_FOUND', 'Control Plane Account was not found')
    if (row.kind === 'environment')
      throw new TunnelError('MANAGED_ACCOUNT', 'Deployment Administrator is managed by environment variables')
    return row
  }

  private accountView(row: AccountRow): AccountView {
    const count = this.options.database.sqlite.query<{ count: number }, [string]>('SELECT count(*) AS count FROM clients WHERE owner_account_id = ?').get(row.internal_id)!.count
    return { ...accountRecord(row), clientCount: count }
  }

  listAccounts(): AccountView[] {
    void this.account
    return this.options.database.sqlite.query<AccountRow, []>('SELECT * FROM accounts ORDER BY kind, username_key, internal_id').all().map(row => this.accountView(row))
  }

  async createAccount(input: { username: string, password: string, role?: AccountRole }): Promise<AccountView> {
    void this.account
    const identity = username(input.username)
    validPassword(input.password)
    const role = input.role ?? 'user'
    const timestamp = new Date().toISOString()
    const passwordHash = await Bun.password.hash(input.password, PASSWORD_ALGORITHM)
    void this.account
    const id = randomUUID()
    try {
      this.options.database.sqlite.query(`
        INSERT INTO accounts(internal_id, kind, username, username_key, role, password_hash, created_at, updated_at)
        VALUES(?, 'local', ?, ?, ?, ?, ?, ?)
      `).run(id, identity.display, identity.key, role, passwordHash, timestamp, timestamp)
    }
    catch (cause) {
      if (isUniqueConstraint(cause))
        throw new TunnelError('USERNAME_TAKEN', 'Username is already in use')
      throw cause
    }
    const row = this.options.database.sqlite.query<AccountRow, [string]>('SELECT * FROM accounts WHERE internal_id = ?').get(id)!
    const created = this.accountView(row)
    this.management.notifyAdministrators()
    return created
  }

  changeAccountRole(accountId: string, role: AccountRole): AccountView {
    const row = this.localAccount(accountId)
    if (!['admin', 'user'].includes(role))
      throw new TunnelError('INVALID_ACCOUNT', 'Account role must be admin or user')
    if (row.role !== role) {
      const timestamp = new Date().toISOString()
      this.options.database.sqlite.query('UPDATE accounts SET role = ?, updated_at = ? WHERE internal_id = ?').run(role, timestamp, accountId)
      this.management.revokeAccountSessions(accountId)
      row.role = role
      row.updated_at = timestamp
    }
    const changed = this.accountView(row)
    this.management.notifyAdministrators()
    return changed
  }

  async resetAccountPassword(accountId: string, newPassword: string): Promise<void> {
    this.localAccount(accountId)
    validPassword(newPassword)
    const passwordHash = await Bun.password.hash(newPassword, PASSWORD_ALGORITHM)
    void this.account
    this.localAccount(accountId)
    this.options.database.sqlite.query('UPDATE accounts SET password_hash = ?, updated_at = ? WHERE internal_id = ?').run(passwordHash, new Date().toISOString(), accountId)
    this.management.revokeAccountSessions(accountId)
    this.management.notifyAdministrators()
  }

  deleteAccount(accountId: string): void {
    this.localAccount(accountId)
    const remove = this.options.database.sqlite.transaction(() => {
      const count = this.options.database.sqlite.query<{ count: number }, [string]>('SELECT count(*) AS count FROM clients WHERE owner_account_id = ?').get(accountId)!.count
      if (count > 0)
        throw new TunnelError('ACCOUNT_NOT_EMPTY', 'Control Plane Account still owns Trusted Tunnel Clients')
      this.options.database.sqlite.query('DELETE FROM accounts WHERE internal_id = ?').run(accountId)
    })
    remove.immediate()
    this.management.revokeAccountSessions(accountId)
    this.management.notifyAdministrators()
  }

  serverState(): ServerView {
    void this.account
    const { adminPassword: _, ...settings } = this.options.serverConfig
    return { frps: this.options.frps.state(), settings }
  }

  async controlFrps(action: 'start' | 'stop' | 'restart'): Promise<ServerView> {
    void this.account
    if (action === 'start')
      await this.options.frps.start(this.options.frpsConfigPath)
    else if (action === 'stop')
      await this.options.frps.stop()
    else
      await this.options.frps.restart()
    return this.serverState()
  }
}

export type { TunnelMutationInput, TunnelPatchInput }
