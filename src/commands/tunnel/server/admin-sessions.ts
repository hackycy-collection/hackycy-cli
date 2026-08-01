import { Buffer } from 'node:buffer'
import { randomBytes, timingSafeEqual } from 'node:crypto'

const MAX_ADMIN_SESSIONS = 32
const SESSION_LIFETIME_MS = 12 * 60 * 60 * 1000

function cookie(request: Request, name: string): string | undefined {
  return request.headers.get('Cookie')?.split(';').map(value => value.trim()).find(value => value.startsWith(`${name}=`))?.slice(name.length + 1)
}

export class AdminSessions {
  private readonly sessions = new Map<string, number>()

  constructor(private readonly username: string, private readonly password: string) {}

  authenticate(username: string, password: string): string | undefined {
    const expected = Buffer.from(`${this.username}\0${this.password}`)
    const supplied = Buffer.from(`${username}\0${password}`)
    if (expected.length !== supplied.length || !timingSafeEqual(expected, supplied))
      return undefined
    const now = Date.now()
    for (const [token, expires] of this.sessions) {
      if (expires < now)
        this.sessions.delete(token)
    }
    while (this.sessions.size >= MAX_ADMIN_SESSIONS) {
      const oldest = this.sessions.keys().next().value
      if (oldest === undefined)
        break
      this.sessions.delete(oldest)
    }
    const value = randomBytes(32).toString('base64url')
    this.sessions.set(value, now + SESSION_LIFETIME_MS)
    return value
  }

  valid(request: Request): boolean {
    const value = cookie(request, 'ycy_tunnel_session')
    const expires = value ? this.sessions.get(value) : undefined
    if (!value || !expires || expires < Date.now()) {
      if (value)
        this.sessions.delete(value)
      return false
    }
    this.sessions.delete(value)
    this.sessions.set(value, expires)
    return true
  }

  delete(request: Request): void {
    const value = cookie(request, 'ycy_tunnel_session')
    if (value)
      this.sessions.delete(value)
  }
}
