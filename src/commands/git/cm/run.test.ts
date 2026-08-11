import type { ResolvedCmProfile } from '../../../config/types'
import { stripVTControlCharacters } from 'node:util'
import { describe, expect, test } from 'bun:test'
import { formatGeneratedMessage } from './run'

const profile: ResolvedCmProfile = {
  name: 'deepseek',
  baseURL: 'https://api.deepseek.com',
  model: 'deepseek-chat',
  apiKey: 'test-key',
  temperature: 0.2,
  timeoutMs: 30_000,
  maxOutputTokens: 200,
}

describe('formatGeneratedMessage', () => {
  test('prints a compact result without changed file details', () => {
    const output = stripVTControlCharacters(formatGeneratedMessage(
      'feat(cm): streamline generated output',
      profile,
      {
        promptTokens: 1_234,
        completionTokens: 42,
        totalTokens: 1_276,
      },
      false,
    ))

    expect(output).toBe([
      'feat(cm): streamline generated output',
      '',
      'Profile: deepseek (deepseek-chat)',
      'Tokens: 1,234 prompt / 42 completion / 1,276 total',
    ].join('\n'))
    expect(output).not.toContain('Changed files')
    expect(output).not.toContain('raw diffs were compressed')
  })

  test('preserves a multiline commit body and handles unavailable usage', () => {
    const output = stripVTControlCharacters(formatGeneratedMessage(
      'fix(cm): preserve commit body\n\nKeep the generated context intact.',
      profile,
      undefined,
      false,
    ))

    expect(output).toContain('fix(cm): preserve commit body\n\nKeep the generated context intact.')
    expect(output).toContain('Tokens: unavailable')
  })

  test('shows unknown partial usage values and the truncation warning', () => {
    const output = stripVTControlCharacters(formatGeneratedMessage(
      'feat(cm): streamline generated output',
      profile,
      { completionTokens: 42 },
      true,
      101,
    ))

    expect(output).toContain('Tokens: unknown prompt / 42 completion / unknown total')
    expect(output).toContain('Commit scope: 101 changed files; raw diffs were compressed to fit the prompt budget.')
    expect(output).toContain('This does not affect which files are committed.')
  })
})
