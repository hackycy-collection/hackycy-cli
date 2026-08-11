import type { DiffHunk, GitChangeSnapshot, SnapshotFile } from './types'
import { describe, expect, test } from 'bun:test'
import {
  buildChangeClusters,
  compileEvidence,
  estimateInputTokens,
  extractEvidenceFacts,
  MAX_INPUT_TOKENS,
} from './evidence'

const system = 'Return a concise Angular commit message.'

function hunk(addedLines: string[], deletedLines: string[] = []): DiffHunk {
  return {
    id: 'worktree:example:0',
    source: 'worktree',
    oldStart: 1,
    oldLines: deletedLines.length,
    newStart: 1,
    newLines: addedLines.length,
    addedLines,
    deletedLines,
  }
}

function file(overrides: Partial<SnapshotFile> & Pick<SnapshotFile, 'path'>): SnapshotFile {
  return {
    id: overrides.path,
    status: `M ${overrides.path}`,
    indexStatus: ' ',
    worktreeStatus: 'M',
    role: 'source',
    contentPolicy: 'inspect',
    stats: { additions: 1, deletions: 0 },
    hunks: [hunk(['export const value = true'])],
    ...overrides,
  }
}

function snapshot(files: SnapshotFile[]): GitChangeSnapshot {
  return {
    repoRoot: '/repo',
    scope: 'all-uncommitted',
    snapshotId: 'snapshot',
    files,
    totals: files.reduce((total, item) => ({
      additions: total.additions + item.stats.additions,
      deletions: total.deletions + item.stats.deletions,
    }), { additions: 0, deletions: 0 }),
  }
}

describe('semantic evidence', () => {
  test('is byte-stable and represents every cluster in P0', () => {
    const files = [
      file({ path: 'src/commands/cm/engine.ts' }),
      file({ path: 'packages/web/src/view.ts' }),
      file({ path: 'docs/cm.md', role: 'docs' }),
    ]
    const first = compileEvidence(snapshot(files), system)
    const second = compileEvidence(snapshot([...files].reverse()), system)

    expect(first.text).toBe(second.text)
    expect(first.coverage).toEqual(second.coverage)
    expect(first.text).toContain('cluster=src/commands/cm')
    expect(first.text).toContain('cluster=packages/web')
    expect(first.coverage.representedClusters).toBe(first.coverage.totalClusters)
  })

  test('extracts package, declaration, test, behavior, and import facts by priority', () => {
    const packageFile = file({
      path: 'package.json',
      role: 'dependency',
      manifest: {
        before: JSON.stringify({ dependencies: { obsolete: '1.0.0', zod: '4.0.0' }, scripts: { test: 'bun test' } }),
        after: JSON.stringify({ dependencies: { zod: '4.4.3', vitest: '3.0.0' }, scripts: { test: 'bun test', lint: 'eslint .' } }),
      },
      hunks: [hunk([])],
    })
    const sourceFile = file({
      path: 'src/feature.ts',
      hunks: [hunk([
        'import { value } from \'./value\'',
        'export function enableFeature(): void {}',
        'throw new Error(\'Feature is unavailable\')',
        'const internal = value',
      ])],
    })
    const testFile = file({
      path: 'src/feature.test.ts',
      role: 'test',
      hunks: [hunk(['test(\'enables the feature\', () => expect(true).toBe(true))'])],
    })
    const facts = extractEvidenceFacts(buildChangeClusters(snapshot([packageFile, sourceFile, testFile])))

    expect(facts.some(item => item.priority === 1 && item.text.includes('dependency replacement add vitest@3.0.0; remove obsolete@1.0.0'))).toBe(true)
    expect(facts.some(item => item.priority === 1 && item.text.includes('symbol added enableFeature'))).toBe(true)
    expect(facts.some(item => item.priority === 1 && item.text.includes('test added "enables the feature"'))).toBe(true)
    expect(facts.some(item => item.priority === 2 && item.text.includes('Feature is unavailable'))).toBe(true)
    expect(facts.some(item => item.priority === 3 && item.text.startsWith('added const internal'))).toBe(true)
  })

  test('extracts a configuration key when a file has only low-priority implementation lines', () => {
    const configFile = file({
      path: 'tsconfig.json',
      role: 'config',
      hunks: [hunk(['    "types": ["bun", "./types.d.ts"],'])],
    })
    const facts = extractEvidenceFacts(buildChangeClusters(snapshot([configFile])))

    expect(facts).toContainEqual(expect.objectContaining({ priority: 1, text: 'config added types' }))
  })

  test('adds a deterministic type hint for workflow, Dockerfile, TypeScript config, docs, style, and release-only changes', () => {
    const cases: Array<[SnapshotFile, string]> = [
      [file({ path: '.github/workflows/release.yml', role: 'config' }), 'type=ci'],
      [file({ path: 'Dockerfile', role: 'unknown' }), 'type=build'],
      [file({ path: 'README.md', role: 'docs' }), 'type=docs'],
      [file({ path: 'styles.css', role: 'unknown' }), 'type=style'],
      [file({
        path: 'package.json',
        role: 'dependency',
        manifest: { before: JSON.stringify({ version: '1.0.0' }), after: JSON.stringify({ version: '1.0.1' }) },
      }), 'type=chore'],
    ]

    for (const [changedFile, expected] of cases)
      expect(compileEvidence(snapshot([changedFile]), system).text).toContain(expected)

    expect(compileEvidence(snapshot([
      file({ path: 'tsconfig.json', role: 'config' }),
      file({ path: 'types.d.ts', role: 'config' }),
    ]), system).text).toContain('type=build')
  })

  test('stays within the complete-message budget without truncating a fact or protected content', () => {
    const oversizedLine = `throw new Error('${'x'.repeat(8_000)}')`
    const publicFile = file({ path: 'src/public.ts', hunks: [hunk([oversizedLine, 'export const enabled = true'])] })
    const sensitiveFile = file({
      path: '.env',
      role: 'sensitive',
      contentPolicy: 'redacted',
      hunks: [hunk(['API_KEY=never-send-this'])],
    })
    const compiled = compileEvidence(snapshot([publicFile, sensitiveFile]), system)

    expect(estimateInputTokens(system, compiled.text)).toBeLessThanOrEqual(MAX_INPUT_TOKENS)
    expect(compiled.text).not.toContain('never-send-this')
    expect(compiled.text).not.toContain('x'.repeat(100))
    expect(compiled.coverage.omittedFacts).toBeGreaterThan(0)
    expect(compiled.text).toContain('symbol added enabled')
  })
})
