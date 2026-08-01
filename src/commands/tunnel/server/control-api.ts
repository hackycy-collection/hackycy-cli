import type { FrpSupervisor } from '../frp/supervisor'
import type { ServerTunnelConfig, TunnelErrorCode } from '../types'
import type { AgentGateway, AgentSocketData } from './agent-gateway'
import type { TunnelControlPlane, TunnelMutationInput, TunnelPatchInput } from './control-plane'
import { z } from 'zod'
import { TunnelError } from '../types'
import { AdminSessions } from './admin-sessions'
import { clientView, tunnelState } from './views'

const API_HEADERS = {
  'Cache-Control': 'no-store',
  'Content-Security-Policy': 'default-src \'none\'; frame-ancestors \'none\'',
  'Referrer-Policy': 'no-referrer',
  'X-Content-Type-Options': 'nosniff',
}

const sessionSchema = z.object({ username: z.string(), password: z.string() }).strict()
const clientCreateSchema = z.object({ remark: z.string().optional() }).strict()
const clientRemarkSchema = z.object({ remark: z.string() }).strict()
const tunnelSchema = z.object({
  protocol: z.enum(['http', 'tcp', 'udp']),
  hostname: z.string().nullable().optional(),
  serverPort: z.number().int().nullable().optional(),
  localHost: z.string().optional(),
  localPort: z.number().int(),
  enabled: z.boolean().optional(),
}).strict()
const tunnelPatchSchema = tunnelSchema.partial()

const TUNNEL_ERROR_STATUS = {
  ACTIVATION_FAILED: 500,
  AUTHENTICATION_FAILED: 401,
  CLIENT_STOPPED: 409,
  CONFIGURATION_FAILED: 500,
  DATABASE_TOO_NEW: 500,
  FRP_INSTALL_FAILED: 500,
  INCOMPATIBLE_CLIENT: 409,
  INSTANCE_ACTIVE: 409,
  INVALID_CLIENT_REMARK: 400,
  INVALID_CONFIG: 400,
  INVALID_FRP_ARCHIVE: 500,
  INVALID_FRP_BINARY: 500,
  INVALID_FRP_VERSION: 500,
  INVALID_HOSTNAME: 400,
  INVALID_LOCAL_ENDPOINT: 400,
  INVALID_PROTOCOL: 400,
  INVALID_REVISION: 400,
  INVALID_TUNNEL: 400,
  LOCK_UNAVAILABLE: 500,
  NOT_FOUND: 404,
  PORT_OUTSIDE_POOL: 400,
  PORT_POOL_EXHAUSTED: 409,
  RESOURCE_RESERVED: 409,
  UNSUPPORTED_PLATFORM: 500,
} as const satisfies Record<TunnelErrorCode, number>

function json(data: unknown, status = 200, headers?: HeadersInit): Response {
  return Response.json(data, { status, headers: { ...API_HEADERS, ...headers } })
}

function error(code: string, message: string, status: number): Response {
  return json({ version: 1, error: { code, message } }, status)
}

function errorResponse(cause: unknown): Response {
  if (!(cause instanceof TunnelError))
    return error('INTERNAL_ERROR', cause instanceof Error ? cause.message : String(cause), 500)
  return error(cause.code, cause.message, TUNNEL_ERROR_STATUS[cause.code])
}

async function requestJson<T>(request: Request, schema: z.ZodType<T>): Promise<T | Response> {
  if (request.headers.get('Content-Type')?.split(';')[0]?.trim().toLowerCase() !== 'application/json')
    return error('UNSUPPORTED_MEDIA_TYPE', 'Request body must use JSON', 415)
  try {
    const parsed = schema.safeParse(await request.json())
    return parsed.success ? parsed.data : error('INVALID_REQUEST', 'Request body is invalid', 400)
  }
  catch {
    return error('INVALID_REQUEST', 'Request body must be valid JSON', 400)
  }
}

function sameOrigin(request: Request): boolean {
  const origin = request.headers.get('Origin')
  if (!origin)
    return true
  try {
    return new URL(origin).host === new URL(request.url).host
  }
  catch {
    return false
  }
}

function parseId(pathname: string, pattern: RegExp): string | undefined {
  const value = pattern.exec(pathname)?.[1]
  if (!value)
    return undefined
  try {
    return decodeURIComponent(value)
  }
  catch {
    return undefined
  }
}

export interface TunnelControlApiOptions {
  config: ServerTunnelConfig
  controlPlane: TunnelControlPlane
  gateway: AgentGateway
  frps: FrpSupervisor
  frpsConfigPath: string
}

export class TunnelControlApi {
  private readonly sessions: AdminSessions
  private readonly eventListeners = new Set<() => void>()
  private readonly unsubscribers: Array<() => void>

  constructor(private readonly options: TunnelControlApiOptions) {
    this.sessions = new AdminSessions(options.config.adminUser, options.config.adminPassword)
    const broadcast = (): void => {
      for (const listener of this.eventListeners)
        listener()
    }
    this.unsubscribers = [
      options.controlPlane.subscribe(broadcast),
      options.gateway.observe(broadcast),
      options.frps.observe(broadcast),
    ]
  }

  async handle(request: Request, server: Bun.Server<AgentSocketData>): Promise<Response | undefined> {
    const url = new URL(request.url)
    if (url.pathname === '/healthz')
      return request.method === 'GET' ? json({ status: 'ok' }) : error('METHOD_NOT_ALLOWED', 'Use GET', 405)

    if (url.pathname === '/api/agent') {
      if (request.method !== 'GET')
        return error('METHOD_NOT_ALLOWED', 'Use GET with WebSocket upgrade', 405)
      const authorization = this.options.gateway.authorize(request)
      if (authorization instanceof Response)
        return authorization
      if (server.upgrade(request, { data: authorization.data }))
        return undefined
      this.options.gateway.cancelAuthorization(authorization.data.clientId)
      return error('UPGRADE_REQUIRED', 'WebSocket upgrade is required', 426)
    }

    if (url.pathname === '/api/session' && request.method === 'POST') {
      const body = await requestJson(request, sessionSchema)
      if (body instanceof Response)
        return body
      const token = this.sessions.authenticate(body.username, body.password)
      if (!token)
        return error('AUTHENTICATION_FAILED', 'Administrator credentials are invalid', 401)
      return json({ version: 1, authenticated: true }, 200, {
        'Set-Cookie': `ycy_tunnel_session=${token}; HttpOnly; SameSite=Strict; Path=/; Max-Age=43200`,
      })
    }

    if (!url.pathname.startsWith('/api/'))
      return error('NOT_FOUND', 'Route not found', 404)
    if (!this.sessions.valid(request))
      return error('AUTHENTICATION_REQUIRED', 'Administrator session is required', 401)
    if (!['GET', 'HEAD'].includes(request.method) && !sameOrigin(request))
      return error('ORIGIN_FORBIDDEN', 'Mutation requests must be same-origin', 403)

    try {
      if (url.pathname === '/api/session' && request.method === 'DELETE') {
        this.sessions.delete(request)
        return new Response(null, { status: 204, headers: { ...API_HEADERS, 'Set-Cookie': 'ycy_tunnel_session=; HttpOnly; SameSite=Strict; Path=/; Max-Age=0' } })
      }
      if (url.pathname === '/api/events' && request.method === 'GET') {
        server.timeout(request, 0)
        const encoder = new TextEncoder()
        let dispose: (() => void) | undefined
        const stream = new ReadableStream<Uint8Array>({
          start: (controller) => {
            const publish = (): void => controller.enqueue(encoder.encode(`data: ${JSON.stringify({ version: 1, changed: true })}\n\n`))
            this.eventListeners.add(publish)
            publish()
            dispose = () => this.eventListeners.delete(publish)
          },
          cancel() {
            dispose?.()
          },
        })
        return new Response(stream, { headers: { ...API_HEADERS, 'Content-Type': 'text/event-stream; charset=utf-8', 'X-Accel-Buffering': 'no' } })
      }
      if (url.pathname === '/api/state' && request.method === 'GET') {
        const clients = this.options.controlPlane.listClients()
        const views = clients.map(client => clientView(this.options.controlPlane, this.options.gateway, client))
        return json({
          version: 1,
          frps: this.options.frps.state(),
          counts: {
            clients: clients.length,
            connected: clients.filter(client => this.options.gateway.state(client.id).connectionState === 'connected').length,
            tunnels: views.reduce((sum, view) => sum + view.tunnelCounts.total, 0),
            pending: views.reduce((sum, view) => sum + view.tunnelCounts.pending, 0),
            errors: views.reduce((sum, view) => sum + view.tunnelCounts.error, 0),
          },
          settings: {
            address: this.options.config.address,
            controlPort: this.options.config.controlPort,
            frpPort: this.options.config.frpPort,
            httpPort: this.options.config.httpPort,
            portRange: this.options.config.portRange,
            advertiseFrpAddress: this.options.config.advertiseFrpAddress ?? null,
            dataDir: this.options.config.dataDir,
            adminUser: this.options.config.adminUser,
          },
        })
      }
      if (url.pathname === '/api/clients') {
        if (request.method === 'GET')
          return json({ version: 1, clients: this.options.controlPlane.listClients().map(client => clientView(this.options.controlPlane, this.options.gateway, client)) })
        if (request.method === 'POST') {
          const body = await requestJson(request, clientCreateSchema)
          if (body instanceof Response)
            return body
          const client = this.options.controlPlane.createClient(body.remark)
          return json({ version: 1, client: clientView(this.options.controlPlane, this.options.gateway, client) }, 201)
        }
        return error('METHOD_NOT_ALLOWED', 'Use GET or POST', 405)
      }
      const clientId = parseId(url.pathname, /^\/api\/clients\/([^/]+)$/)
      if (clientId) {
        if (request.method === 'GET') {
          const client = this.options.controlPlane.getClient(clientId)
          const runtime = this.options.gateway.state(clientId)
          return json({
            version: 1,
            client: clientView(this.options.controlPlane, this.options.gateway, client),
            tunnels: this.options.controlPlane.listTunnels(clientId).map(tunnel => ({ ...tunnel, state: tunnelState(tunnel, client, runtime) })),
          })
        }
        if (request.method === 'PATCH') {
          const body = await requestJson(request, clientRemarkSchema)
          if (body instanceof Response)
            return body
          const client = this.options.controlPlane.updateClientRemark(clientId, body.remark)
          return json({ version: 1, client: clientView(this.options.controlPlane, this.options.gateway, client) })
        }
        if (request.method === 'DELETE') {
          this.options.controlPlane.deleteClient(clientId)
          return new Response(null, { status: 204, headers: API_HEADERS })
        }
        return error('METHOD_NOT_ALLOWED', 'Use GET, PATCH, or DELETE', 405)
      }
      const rotateClientId = parseId(url.pathname, /^\/api\/clients\/([^/]+)\/rotate$/)
      if (rotateClientId && request.method === 'POST') {
        const client = this.options.controlPlane.rotateClientToken(rotateClientId)
        return json({ version: 1, client: clientView(this.options.controlPlane, this.options.gateway, client) })
      }
      const restartClientId = parseId(url.pathname, /^\/api\/clients\/([^/]+)\/restart$/)
      if (restartClientId && request.method === 'POST') {
        if (!this.options.gateway.restartClient(restartClientId))
          return error('CLIENT_OFFLINE', 'Trusted Tunnel Client is not connected', 409)
        return json({ version: 1, accepted: true }, 202)
      }
      const tunnelsClientId = parseId(url.pathname, /^\/api\/clients\/([^/]+)\/tunnels$/)
      if (tunnelsClientId) {
        if (request.method === 'GET')
          return json({ version: 1, tunnels: this.options.controlPlane.listTunnels(tunnelsClientId) })
        if (request.method === 'POST') {
          const body = await requestJson(request, tunnelSchema)
          if (body instanceof Response)
            return body
          return json({ version: 1, tunnel: this.options.controlPlane.createTunnel(tunnelsClientId, body as TunnelMutationInput) }, 201)
        }
        return error('METHOD_NOT_ALLOWED', 'Use GET or POST', 405)
      }
      const tunnelId = parseId(url.pathname, /^\/api\/tunnels\/([^/]+)$/)
      if (tunnelId) {
        if (request.method === 'PATCH') {
          const body = await requestJson(request, tunnelPatchSchema)
          if (body instanceof Response)
            return body
          return json({ version: 1, tunnel: this.options.controlPlane.updateTunnel(tunnelId, body as TunnelPatchInput) })
        }
        if (request.method === 'DELETE') {
          this.options.controlPlane.deleteTunnel(tunnelId)
          return new Response(null, { status: 204, headers: API_HEADERS })
        }
        return error('METHOD_NOT_ALLOWED', 'Use PATCH or DELETE', 405)
      }
      const action = /^\/api\/server\/frp\/(start|stop|restart)$/.exec(url.pathname)?.[1]
      if (action && request.method === 'POST') {
        if (action === 'start')
          await this.options.frps.start(this.options.frpsConfigPath)
        else if (action === 'stop')
          await this.options.frps.stop()
        else
          await this.options.frps.restart()
        return json({ version: 1, frps: this.options.frps.state() })
      }
      return error('NOT_FOUND', 'Route not found', 404)
    }
    catch (cause) {
      return errorResponse(cause)
    }
  }

  stop(): void {
    for (const unsubscribe of this.unsubscribers)
      unsubscribe()
    this.eventListeners.clear()
  }
}
