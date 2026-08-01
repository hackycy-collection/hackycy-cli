import type { ClientRecord, ClientRuntimeState, TunnelDefinition, TunnelPresentationState } from '../../types'

export interface ClientView extends ClientRecord {
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
