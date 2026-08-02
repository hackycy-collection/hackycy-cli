import { randomBytes } from 'node:crypto'

const SESSION_LIFETIME_MS = 12 * 60 * 60 * 1000
const MAX_ACCOUNT_SESSIONS = 8
const MAX_SESSIONS = 128
const PASSWORD_ALGORITHM = { algorithm: 'argon2id', memoryCost: 65_536, timeCost: 3 } as const

interface ServeAccount {
  username: string
  key: string
  passwordHash: string
}

interface SessionEntry {
  accountKey: string
  expiresAt: number
}

export interface ServeSession {
  account: { username: string }
  expiresAt: string
}

export interface ServeSessionGrant extends ServeSession {
  token: string
}

export interface ServeAuthenticationOptions {
  sessionLifetimeMs?: number
  maxAccountSessions?: number
  maxSessions?: number
}

function parseAccount(specification: string): { username: string, key: string, password: string } {
  const separator = specification.indexOf(':')
  if (separator === -1)
    throw new Error('Account must use <username>:<password>')

  const username = specification.slice(0, separator)
  if (!/^[\w.-]{1,64}$/.test(username))
    throw new Error('Username must contain 1-64 ASCII letters, numbers, dots, underscores, or hyphens')

  const password = specification.slice(separator + 1)
  if (password.length < 8 || password.length > 256)
    throw new Error('Password must contain 8-256 characters')

  return { username, key: username.toLowerCase(), password }
}

function usernameKey(value: string): string | undefined {
  return /^[\w.-]{1,64}$/.test(value) ? value.toLowerCase() : undefined
}

export class ServeAuthentication {
  private readonly sessions = new Map<string, SessionEntry>()
  private readonly observers = new Map<string, Set<() => void>>()
  private expirationTimer: ReturnType<typeof setTimeout> | undefined
  private closed = false

  private constructor(
    private readonly accounts: Map<string, ServeAccount>,
    private readonly options: Required<ServeAuthenticationOptions>,
  ) {}

  static async create(specifications: string[], options: ServeAuthenticationOptions = {}): Promise<ServeAuthentication | undefined> {
    if (specifications.length === 0)
      return undefined

    const parsedAccounts = specifications.map(parseAccount)
    const keys = new Set<string>()
    for (const account of parsedAccounts) {
      if (keys.has(account.key))
        throw new Error(`Username '${account.username}' is specified more than once`)
      keys.add(account.key)
    }

    const accounts = new Map<string, ServeAccount>()
    for (const account of parsedAccounts) {
      accounts.set(account.key, {
        username: account.username,
        key: account.key,
        passwordHash: await Bun.password.hash(account.password, PASSWORD_ALGORITHM),
      })
    }

    return new ServeAuthentication(accounts, {
      sessionLifetimeMs: Math.max(1, options.sessionLifetimeMs ?? SESSION_LIFETIME_MS),
      maxAccountSessions: Math.max(1, options.maxAccountSessions ?? MAX_ACCOUNT_SESSIONS),
      maxSessions: Math.max(1, options.maxSessions ?? MAX_SESSIONS),
    })
  }

  get accountCount(): number {
    return this.accounts.size
  }

  private revoke(token: string, scheduleExpiration = true): void {
    if (!this.sessions.delete(token))
      return
    const listeners = this.observers.get(token)
    this.observers.delete(token)
    if (listeners) {
      for (const listener of [...listeners])
        listener()
    }
    if (scheduleExpiration)
      this.scheduleExpiration()
  }

  private clearExpiredSessions(): void {
    const now = Date.now()
    for (const [token, session] of this.sessions) {
      if (session.expiresAt <= now)
        this.revoke(token, false)
    }
    this.scheduleExpiration()
  }

  private scheduleExpiration(): void {
    clearTimeout(this.expirationTimer)
    this.expirationTimer = undefined
    if (this.closed)
      return

    let nextExpiration = Number.POSITIVE_INFINITY
    for (const session of this.sessions.values())
      nextExpiration = Math.min(nextExpiration, session.expiresAt)
    if (!Number.isFinite(nextExpiration))
      return

    this.expirationTimer = setTimeout(() => this.clearExpiredSessions(), Math.max(0, nextExpiration - Date.now()))
    this.expirationTimer.unref?.()
  }

  private createSession(account: ServeAccount): ServeSessionGrant {
    this.clearExpiredSessions()
    const accountSessions = [...this.sessions].filter(([, session]) => session.accountKey === account.key)
    while (accountSessions.length >= this.options.maxAccountSessions) {
      const oldest = accountSessions.shift()
      if (oldest)
        this.revoke(oldest[0], false)
    }
    while (this.sessions.size >= this.options.maxSessions)
      this.revoke(this.sessions.keys().next().value!, false)

    const token = randomBytes(32).toString('base64url')
    const expiresAt = Date.now() + this.options.sessionLifetimeMs
    this.sessions.set(token, { accountKey: account.key, expiresAt })
    this.scheduleExpiration()
    return { token, expiresAt: new Date(expiresAt).toISOString(), account: { username: account.username } }
  }

  async signIn(input: { username: string, password: string }): Promise<ServeSessionGrant | undefined> {
    if (this.closed)
      return undefined
    const account = this.accounts.get(usernameKey(input.username) ?? '')
    const fallbackHash = this.accounts.values().next().value!.passwordHash
    const verified = await Bun.password.verify(input.password, account?.passwordHash ?? fallbackHash)
    return account && verified ? this.createSession(account) : undefined
  }

  resume(token: string | undefined): ServeSession | undefined {
    if (this.closed || !token)
      return undefined
    this.clearExpiredSessions()
    const session = this.sessions.get(token)
    const account = session ? this.accounts.get(session.accountKey) : undefined
    if (!session || !account)
      return undefined

    this.sessions.delete(token)
    this.sessions.set(token, session)
    return { expiresAt: new Date(session.expiresAt).toISOString(), account: { username: account.username } }
  }

  signOut(token: string | undefined): void {
    if (token)
      this.revoke(token)
  }

  observe(token: string, listener: () => void): () => void {
    if (!this.sessions.has(token))
      return () => {}
    const listeners = this.observers.get(token) ?? new Set()
    listeners.add(listener)
    this.observers.set(token, listeners)
    return () => {
      listeners.delete(listener)
      if (listeners.size === 0)
        this.observers.delete(token)
    }
  }

  close(): void {
    if (this.closed)
      return
    this.closed = true
    clearTimeout(this.expirationTimer)
    this.expirationTimer = undefined
    for (const token of [...this.sessions.keys()])
      this.revoke(token, false)
    this.observers.clear()
  }
}

export const createServeAuthentication = ServeAuthentication.create
