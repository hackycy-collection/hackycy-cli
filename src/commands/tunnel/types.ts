export const TUNNEL_PROTOCOL_VERSION = 1 as const

export type TunnelProtocol = 'http' | 'tcp' | 'udp'
export type ClientConnectionState = 'connected' | 'disconnected' | 'incompatible' | 'revocation_pending'
export type FrpProcessState = 'stopped' | 'running' | 'recovering' | 'configuration_failed'
export type TunnelPresentationState = 'Disabled' | 'Pending' | 'Applied' | 'Error'
export type TunnelErrorCode
  = | 'ACTIVATION_FAILED'
    | 'AUTHENTICATION_FAILED'
    | 'CLIENT_STOPPED'
    | 'CONFIGURATION_FAILED'
    | 'DATABASE_TOO_NEW'
    | 'FRP_INSTALL_FAILED'
    | 'INCOMPATIBLE_CLIENT'
    | 'INSTANCE_ACTIVE'
    | 'INVALID_CLIENT_REMARK'
    | 'INVALID_CONFIG'
    | 'INVALID_FRP_ARCHIVE'
    | 'INVALID_FRP_BINARY'
    | 'INVALID_FRP_VERSION'
    | 'INVALID_HOSTNAME'
    | 'INVALID_LOCAL_ENDPOINT'
    | 'INVALID_PROTOCOL'
    | 'INVALID_REVISION'
    | 'INVALID_TUNNEL'
    | 'LOCK_UNAVAILABLE'
    | 'NOT_FOUND'
    | 'PORT_OUTSIDE_POOL'
    | 'PORT_POOL_EXHAUSTED'
    | 'RESOURCE_RESERVED'
    | 'UNSUPPORTED_PLATFORM'

export interface TunnelDefinition {
  id: string
  protocol: TunnelProtocol
  hostname: string | null
  serverPort: number | null
  localHost: string
  localPort: number
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface TunnelSnapshot {
  clientKey: string
  revision: number
  tunnels: TunnelDefinition[]
}

export interface FrpcDesiredConfiguration {
  advertisedFrpHost: string
  advertisedFrpPort: number
  internalFrpToken: string
  snapshot: TunnelSnapshot
}

export interface StructuredRuntimeError {
  code: string
  message: string
  revision?: number
}

export interface ClientRecord {
  id: string
  remark: string
  token: string
  desiredRevision: number
  lastAppliedRevision: number
  revocationPending: boolean
  createdAt: string
  rotatedAt: string | null
}

export interface ClientRuntimeState {
  connectionState: ClientConnectionState
  processState: FrpProcessState
  lastError?: StructuredRuntimeError
}

export interface FrpArtifactDescription {
  version: string
  archive: string
  url: string
  sha256: string
  frpcSha256: string
}

export interface AgentHelloMessage {
  type: 'hello'
  tunnelProtocolVersion: number
  ycyVersion: string
  platform: string
  architecture: string
  lastAppliedRevision: number
}

export interface AgentWelcomeMessage {
  type: 'welcome'
  tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION
  requiredFrpVersion: string
  artifact: FrpArtifactDescription
  advertisedFrpHost: string
  advertisedFrpPort: number
  internalFrpToken: string
  snapshot: TunnelSnapshot
}

export interface DesiredStateMessage {
  type: 'desired_state'
  tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION
  snapshot: TunnelSnapshot
}

export interface ApplyResultMessage {
  type: 'apply_result'
  tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION
  revision: number
  success: boolean
  error?: StructuredRuntimeError
}

export interface ProcessStateMessage {
  type: 'process_state'
  tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION
  state: FrpProcessState
  error?: StructuredRuntimeError
}

export type AgentToServerMessage = AgentHelloMessage | ApplyResultMessage | ProcessStateMessage

export type ServerToAgentMessage
  = | AgentWelcomeMessage
    | DesiredStateMessage
    | { type: 'restart_frpc', tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION }
    | { type: 'revoke', tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION, reason: 'rotated' | 'deleted' }
    | { type: 'incompatible', tunnelProtocolVersion: typeof TUNNEL_PROTOCOL_VERSION, message: string }

export interface ServerTunnelConfig {
  address: string
  controlPort: number
  frpPort: number
  httpPort: number
  portRange: { start: number, end: number }
  advertiseFrpAddress?: { host: string, port: number }
  dataDir: string
  adminUser: string
  adminPassword: string
}

export interface ClientTunnelConfig {
  server: URL
  token: string
  stateDir: string
}

export class TunnelError extends Error {
  constructor(public readonly code: TunnelErrorCode, message: string) {
    super(message)
    this.name = 'TunnelError'
  }
}
