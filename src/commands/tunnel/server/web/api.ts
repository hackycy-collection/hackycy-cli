import type { AccountKind, AccountRole, ClientRecord, ClientRuntimeState, TunnelDefinition, TunnelPresentationState } from '../../types'

export interface CurrentAccount {
  id: string
  kind: AccountKind
  username: string
  role: AccountRole
  createdAt: string
  updatedAt: string
  managedByEnvironment: boolean
}

export interface AccountView extends CurrentAccount {
  clientCount: number
}

export interface ClientView extends Omit<ClientRecord, 'ownerAccountId'> {
  owner: { id: string, username: string }
  runtime: ClientRuntimeState
  tunnelCounts: { total: number, enabled: number, applied: number, pending: number, error: number }
}

export type TunnelView = TunnelDefinition & { state: TunnelPresentationState }

export class ApiError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message)
  }
}

export async function apiJson<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  const body = response.status === 204 ? undefined : await response.json()
  if (!response.ok && response.status === 401)
    window.dispatchEvent(new Event('tunnel-authentication-required'))
  if (!response.ok)
    throw new ApiError(response.status, body?.error?.message ?? `Request failed (${response.status})`)
  return body as T
}

export function jsonRequest(method: string, body?: unknown): RequestInit {
  return {
    method,
    headers: { 'Content-Type': 'application/json' },
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  }
}
