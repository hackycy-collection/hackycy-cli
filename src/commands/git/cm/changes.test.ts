import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, test } from 'bun:test'
import { collectChangeSummary, MAX_PROMPT_CHARS, stageFiles } from './changes'

const temporaryDirectories: string[] = []

afterEach(async () => {
  await Promise.all(temporaryDirectories.splice(0).map(directory => rm(directory, { recursive: true, force: true })))
})

async function runGit(repoRoot: string, args: string[]): Promise<string> {
  const proc = Bun.spawn(['git', '-C', repoRoot, ...args], {
    stdout: 'pipe',
    stderr: 'pipe',
  })
  const [stdout, stderr, exitCode] = await Promise.all([
    new Response(proc.stdout).text(),
    new Response(proc.stderr).text(),
    proc.exited,
  ])

  if (exitCode !== 0)
    throw new Error(stderr.trim() || `git ${args.join(' ')} failed`)

  return stdout
}

async function createRepository(): Promise<string> {
  const repoRoot = await mkdtemp(path.join(tmpdir(), 'ycy-cm-changes-'))
  temporaryDirectories.push(repoRoot)
  await runGit(repoRoot, ['init', '-q'])
  await runGit(repoRoot, ['config', 'user.name', 'CM Test'])
  await runGit(repoRoot, ['config', 'user.email', 'cm-test@example.test'])
  return repoRoot
}

async function writeChangedFile(repoRoot: string, filePath: string, contents: string): Promise<void> {
  const absolutePath = path.join(repoRoot, filePath)
  await mkdir(path.dirname(absolutePath), { recursive: true })
  await writeFile(absolutePath, contents)
}

describe('collectChangeSummary', () => {
  test('keeps more than 80 files available for explicit staging', async () => {
    const repoRoot = await createRepository()
    const filePaths = Array.from({ length: 101 }, (_, index) => `src/file-${index}.ts`)
    await Promise.all(filePaths.map((filePath, index) => writeChangedFile(
      repoRoot,
      filePath,
      `export const file${index} = ${index}\n`,
    )))

    const summary = await collectChangeSummary({ cwd: repoRoot })

    expect(summary.files).toHaveLength(101)
    expect(summary.promptText).toContain('Changed files (101 total):')
    expect(summary.promptText).toContain('A src/file-100.ts')
    expect(summary.promptText).not.toContain('exceeded 80')

    await stageFiles(repoRoot, summary.files.map(file => file.path))
    const stagedPaths = (await runGit(repoRoot, ['diff', '--cached', '--name-only']))
      .trim()
      .split('\n')
      .filter(Boolean)

    expect(stagedPaths).toHaveLength(101)
  })

  test('compacts large diffs within the prompt budget across directories', async () => {
    const repoRoot = await createRepository()
    const directories = ['api', 'cli', 'web']
    const payload = 'x'.repeat(20_000)
    const filePaths = Array.from(
      { length: 101 },
      (_, index) => `${directories[index % directories.length]}/file-${index}.ts`,
    )
    await Promise.all(filePaths.map(filePath => writeChangedFile(
      repoRoot,
      filePath,
      `export const payload = '${payload}'\n`,
    )))

    const summary = await collectChangeSummary({ cwd: repoRoot })

    expect(summary.files).toHaveLength(101)
    expect(summary.truncated).toBe(true)
    expect(summary.promptText.length).toBeLessThanOrEqual(MAX_PROMPT_CHARS)
    expect(summary.promptText).toContain('# api')
    expect(summary.promptText).toContain('# cli')
    expect(summary.promptText).toContain('# web')
    expect(summary.promptText).toContain('Some raw diffs were compacted to fit the prompt budget.')
  })

  test('groups a path manifest only after it exceeds its own budget', async () => {
    const repoRoot = await createRepository()
    const directory = `src/${'long-directory-name-'.repeat(6)}`
    await Promise.all(Array.from({ length: 101 }, (_, index) => writeChangedFile(
      repoRoot,
      `${directory}/file-${index}.ts`,
      `export const file${index} = ${index}\n`,
    )))

    const summary = await collectChangeSummary({ cwd: repoRoot })
    const manifest = summary.promptText.split('\n\nDiffs:', 1)[0]!

    expect(summary.files).toHaveLength(101)
    expect(manifest).toContain('Changed files (101 total; paths grouped):')
    expect(manifest).toContain('A src (101 files)')
    expect(manifest).not.toContain(`${directory}/file-100.ts`)
    expect(summary.promptText.length).toBeLessThanOrEqual(MAX_PROMPT_CHARS)
  })
})
