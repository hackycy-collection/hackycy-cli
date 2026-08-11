import type { ResolvedCmProfile } from './types'
import { afterEach, describe, expect, test } from 'bun:test'
import { createChatCompletionWithUsage } from './client'

const profile: ResolvedCmProfile = {
  name: 'test',
  baseURL: 'https://provider.test',
  model: 'test-model',
  apiKey: 'test-key',
  temperature: 0.2,
  timeoutMs: 30_000,
  maxOutputTokens: 321,
}

const originalFetch = globalThis.fetch

afterEach(() => {
  globalThis.fetch = originalFetch
})

describe('createChatCompletionWithUsage', () => {
  test('keeps the profile output limit unless a request override is supplied', async () => {
    const requests: Array<{ max_tokens?: number }> = []
    globalThis.fetch = (async (_input, init) => {
      requests.push(JSON.parse(String(init?.body)) as { max_tokens?: number })
      return new Response(JSON.stringify({ choices: [{ message: { content: 'ok' } }] }))
    }) as typeof fetch
    const messages = [{ role: 'system' as const, content: 'Return ok' }]

    await createChatCompletionWithUsage(profile, messages)
    await createChatCompletionWithUsage(profile, messages, { maxOutputTokens: 80 })

    expect(requests.map(request => request.max_tokens)).toEqual([321, 80])
  })
})
