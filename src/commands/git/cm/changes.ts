import type { ChangeSummary, FileChange } from './types'
import { lstat } from 'node:fs/promises'
import path from 'node:path'

export const MAX_PROMPT_CHARS = 24_000
const MAX_MANIFEST_CHARS = 8_000
const LARGE_FILE_BYTES = 200_000
const COMPACTION_NOTICE = 'Some raw diffs were compacted to fit the prompt budget.'

const BINARY_EXTS = new Set([
  '.png',
  '.jpg',
  '.jpeg',
  '.gif',
  '.webp',
  '.ico',
  '.pdf',
  '.zip',
  '.gz',
  '.tgz',
  '.7z',
  '.rar',
  '.sqlite',
  '.db',
  '.woff',
  '.woff2',
  '.ttf',
  '.otf',
  '.mp3',
  '.mp4',
  '.mov',
])

const SKIP_DIFF_BASENAMES = new Set([
  'bun.lock',
  'package-lock.json',
  'pnpm-lock.yaml',
  'yarn.lock',
])

async function pathExists(filePath: string): Promise<boolean> {
  try {
    await lstat(filePath)
    return true
  }
  catch (err) {
    if ((err as NodeJS.ErrnoException).code === 'ENOENT')
      return false

    throw err
  }
}

async function runGit(args: string[], cwd?: string, allowFailure = false): Promise<string> {
  const proc = Bun.spawn(['git', ...args], {
    cwd,
    stdout: 'pipe',
    stderr: 'pipe',
  })
  const [stdout, stderr, exitCode] = await Promise.all([
    new Response(proc.stdout).text(),
    new Response(proc.stderr).text(),
    proc.exited,
  ])

  if (exitCode !== 0 && !allowFailure)
    throw new Error(stderr.trim() || `git ${args.join(' ')} failed with exit code ${exitCode}`)

  return stdout
}

export async function getRepoRoot(cwd?: string): Promise<string> {
  return (await runGit(['rev-parse', '--show-toplevel'], cwd)).trim()
}

export async function stageAllChanges(repoRoot: string): Promise<void> {
  await runGit(['add', '-A'], repoRoot)
}

export async function hasStagedChanges(repoRoot: string): Promise<boolean> {
  const proc = Bun.spawn(['git', 'diff', '--cached', '--quiet'], { cwd: repoRoot })
  const exitCode = await proc.exited
  return exitCode === 1
}

export async function stageFiles(repoRoot: string, filePaths: string[]): Promise<void> {
  const existingPaths: string[] = []
  const missingPaths: string[] = []

  for (const filePath of filePaths) {
    const exists = await pathExists(path.join(repoRoot, filePath))
    if (exists)
      existingPaths.push(filePath)
    else
      missingPaths.push(filePath)
  }

  if (existingPaths.length > 0)
    await runGit(['add', '-A', '--', ...existingPaths], repoRoot)
  if (missingPaths.length > 0)
    await runGit(['update-index', '--remove', '--', ...missingPaths], repoRoot)
}

export async function unstageFiles(repoRoot: string, filePaths: string[]): Promise<void> {
  await runGit(['restore', '--staged', '--', ...filePaths], repoRoot)
}

export async function commitWithMessage(repoRoot: string, message: string): Promise<void> {
  const parts = message.split('\n').map(line => line.trimEnd())
  const subject = parts[0] ?? message
  const body = parts.slice(1).join('\n').trim()
  const args = ['commit', '-m', subject]
  if (body)
    args.push('-m', body)

  await runGit(args, repoRoot)
}

async function getCurrentBranch(repoRoot: string): Promise<string> {
  const branch = (await runGit(['branch', '--show-current'], repoRoot)).trim()
  if (!branch)
    throw new Error('Cannot push from detached HEAD. Check out a branch first.')

  return branch
}

export async function pushChanges(repoRoot: string, remote = 'origin'): Promise<void> {
  const branch = await getCurrentBranch(repoRoot)
  await runGit(['push', '-u', remote, branch], repoRoot)
}

export async function getRecentCommitSubjects(repoRoot: string, limit = 20): Promise<string[]> {
  const output = await runGit(['log', `-${limit}`, '--pretty=%s'], repoRoot, true)
  return output.split('\n').map(line => line.trim()).filter(Boolean)
}

function parseStatus(output: string): FileChange[] {
  const chunks = output.split('\0').filter(Boolean)
  const files: FileChange[] = []

  for (let i = 0; i < chunks.length; i++) {
    const chunk = chunks[i]!
    const indexStatus = chunk[0] ?? ' '
    const worktreeStatus = chunk[1] ?? ' '
    const filePath = chunk.slice(3)
    let originalPath: string | undefined

    if (indexStatus === 'R' || indexStatus === 'C') {
      originalPath = chunks[++i]
    }

    files.push({
      path: filePath,
      originalPath,
      status: formatStatus(indexStatus, worktreeStatus, filePath, originalPath),
      indexStatus,
      worktreeStatus,
      binary: isBinaryPath(filePath),
      sensitive: isSensitivePath(filePath),
    })
  }

  return files
}

async function isSubmodulePath(repoRoot: string, filePath: string): Promise<boolean> {
  const output = await runGit(['ls-files', '-s', '--', filePath], repoRoot, true)
  return output.split('\n').some(line => line.startsWith('160000 '))
}

async function filterSubmodules(repoRoot: string, files: FileChange[]): Promise<FileChange[]> {
  const filtered: FileChange[] = []

  for (const file of files) {
    if (!(await isSubmodulePath(repoRoot, file.path)))
      filtered.push(file)
  }

  return filtered
}

function formatStatus(indexStatus: string, worktreeStatus: string, filePath: string, originalPath?: string): string {
  if (indexStatus === '?' && worktreeStatus === '?')
    return `A ${filePath}`
  if (indexStatus === 'R')
    return `R ${originalPath ?? filePath} -> ${filePath}`
  const status = `${indexStatus}${worktreeStatus}`.trim()
  return `${status || 'M'} ${filePath}`
}

function isBinaryPath(filePath: string): boolean {
  return BINARY_EXTS.has(path.extname(filePath).toLowerCase())
}

function isGeneratedPath(filePath: string): boolean {
  const normalized = filePath.replaceAll('\\', '/')
  const basename = path.basename(filePath)
  return SKIP_DIFF_BASENAMES.has(basename)
    || normalized.includes('/dist/')
    || normalized.startsWith('dist/')
    || normalized.includes('/build/')
    || normalized.startsWith('build/')
    || normalized.includes('/coverage/')
    || normalized.startsWith('coverage/')
    || basename.endsWith('.min.js')
    || basename.endsWith('.map')
}

function isSensitivePath(filePath: string): boolean {
  const basename = path.basename(filePath)
  const ext = path.extname(filePath).toLowerCase()
  return basename === '.env'
    || basename.startsWith('.env.')
    || basename === 'id_rsa'
    || basename === 'id_ed25519'
    || ext === '.pem'
    || ext === '.key'
    || ext === '.p12'
    || ext === '.pfx'
}

async function fileSize(repoRoot: string, filePath: string): Promise<number | undefined> {
  try {
    const file = Bun.file(path.join(repoRoot, filePath))
    return file.size
  }
  catch {
    return undefined
  }
}

async function getFileDiff(repoRoot: string, file: FileChange, stagedOnly: boolean): Promise<string> {
  if (file.indexStatus === '?' && file.worktreeStatus === '?')
    return readUntrackedFilePreview(repoRoot, file.path)

  if (stagedOnly)
    return runGit(['diff', '--cached', '--no-ext-diff', '--', file.path], repoRoot, true)

  const parts: string[] = []

  if (file.indexStatus !== ' ' && file.indexStatus !== '?') {
    const cached = await runGit(['diff', '--cached', '--no-ext-diff', '--', file.path], repoRoot, true)
    if (cached.trim())
      parts.push(`# staged diff\n${cached}`)
  }

  if (file.worktreeStatus !== ' ') {
    const unstaged = await runGit(['diff', '--no-ext-diff', '--', file.path], repoRoot, true)
    if (unstaged.trim())
      parts.push(`# unstaged diff\n${unstaged}`)
  }

  return parts.join('\n')
}

async function readUntrackedFilePreview(repoRoot: string, filePath: string): Promise<string> {
  const absolute = path.join(repoRoot, filePath)
  const file = Bun.file(absolute)
  const size = file.size
  if (size > LARGE_FILE_BYTES)
    return ''

  const text = await file.text()
  if (text.includes('\0'))
    return ''

  return [
    `diff --git a/${filePath} b/${filePath}`,
    'new file mode 100644',
    `--- /dev/null`,
    `+++ b/${filePath}`,
    '@@',
    ...text.split('\n').map(line => `+${line}`),
  ].join('\n')
}

async function enrichFile(repoRoot: string, file: FileChange, stagedOnly: boolean): Promise<void> {
  if (file.sensitive) {
    file.diffSkippedReason = 'sensitive file path'
    return
  }

  if (file.binary) {
    file.diffSkippedReason = 'binary file'
    return
  }

  if (isGeneratedPath(file.path)) {
    file.diffSkippedReason = 'generated or lock file'
    return
  }

  const size = await fileSize(repoRoot, file.path)
  if (size !== undefined && size > LARGE_FILE_BYTES) {
    file.diffSkippedReason = 'large file'
    return
  }

  const diff = await getFileDiff(repoRoot, file, stagedOnly)
  if (!diff.trim()) {
    file.diffSkippedReason = 'no textual diff'
    return
  }

  if (/^Binary files .+ differ$/m.test(diff)) {
    file.diffSkippedReason = 'binary diff'
    file.binary = true
    return
  }

  file.diff = diff
}

function getDirectoryGroup(filePath: string): string {
  const normalized = filePath.replaceAll('\\', '/')
  const separator = normalized.indexOf('/')
  return separator === -1 ? '.' : normalized.slice(0, separator)
}

function appendWithinBudget(text: string, addition: string, budget: number): string | undefined {
  const next = `${text}${addition}`
  return next.length <= budget ? next : undefined
}

function buildManifest(files: FileChange[]): { text: string, compacted: boolean } {
  const header = `Changed files (${files.length} total):`
  const fullManifest = [header, ...files.map(file => file.status)].join('\n')
  if (fullManifest.length <= MAX_MANIFEST_CHARS)
    return { text: fullManifest, compacted: false }

  const groups = new Map<string, number>()
  for (const file of files) {
    const kind = file.status.split(' ', 1)[0] ?? 'M'
    const group = `${kind} ${getDirectoryGroup(file.path)}`
    groups.set(group, (groups.get(group) ?? 0) + 1)
  }

  const lines = [...groups]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([group, count]) => `${group} (${count} file${count === 1 ? '' : 's'})`)
  let text = `Changed files (${files.length} total; paths grouped):`
  let rendered = 0
  for (const line of lines) {
    const next = appendWithinBudget(text, `\n${line}`, MAX_MANIFEST_CHARS)
    if (!next)
      break
    text = next
    rendered += 1
  }

  if (rendered < lines.length) {
    const suffix = `\n... ${lines.length - rendered} additional path group(s)`
    const next = appendWithinBudget(text, suffix, MAX_MANIFEST_CHARS)
    if (next)
      text = next
  }

  return { text, compacted: true }
}

interface DiffGroup {
  name: string
  files: FileChange[]
}

function groupDiffFiles(files: FileChange[]): DiffGroup[] {
  const groups = new Map<string, FileChange[]>()
  for (const file of files) {
    if (!file.diff)
      continue
    const name = getDirectoryGroup(file.path)
    const group = groups.get(name) ?? []
    group.push(file)
    groups.set(name, group)
  }

  return [...groups]
    .map(([name, groupFiles]) => ({
      name,
      files: groupFiles.sort((left, right) => {
        const lengthDelta = (right.diff?.length ?? 0) - (left.diff?.length ?? 0)
        return lengthDelta || left.path.localeCompare(right.path)
      }),
    }))
    .sort((left, right) => left.name.localeCompare(right.name))
}

function buildGroupDiffSection(group: DiffGroup, budget: number): { text: string, compacted: boolean } {
  const header = `\n# ${group.name}`
  if (header.length >= budget)
    return { text: '', compacted: true }

  let text = header
  for (const file of group.files) {
    const sectionHeader = `\n\n## ${file.status}\n`
    const available = budget - text.length - sectionHeader.length
    if (available <= 0)
      return { text, compacted: true }

    const diff = file.diff!
    if (diff.length <= available) {
      text += `${sectionHeader}${diff}`
      continue
    }

    const marker = '\n...[diff truncated for prompt budget]'
    const previewLength = Math.max(0, available - marker.length)
    if (previewLength > 0)
      text += `${sectionHeader}${diff.slice(0, previewLength)}${marker}`
    return { text, compacted: true }
  }

  return { text, compacted: false }
}

function buildDiffContext(files: FileChange[], budget: number): { text: string, compacted: boolean } {
  const groups = groupDiffFiles(files)
  let remaining = budget
  let text = ''
  let compacted = false

  for (let index = 0; index < groups.length; index++) {
    const group = groups[index]!
    const groupBudget = Math.floor(remaining / (groups.length - index))
    const section = buildGroupDiffSection(group, groupBudget)
    if (!section.text) {
      compacted = true
      continue
    }

    text += section.text
    remaining -= section.text.length
    compacted ||= section.compacted
  }

  return { text, compacted }
}

export async function collectChangeSummary(options: { stagedOnly?: boolean, cwd?: string } = {}): Promise<ChangeSummary> {
  const repoRoot = await getRepoRoot(options.cwd)
  const statusOutput = await runGit(['status', '--porcelain=v1', '-z', '--untracked-files=all'], repoRoot)
  let files = parseStatus(statusOutput)

  if (options.stagedOnly) {
    files = files.filter(file => file.indexStatus !== ' ' && file.indexStatus !== '?')
  }

  files = await filterSubmodules(repoRoot, files)

  if (files.length === 0) {
    return {
      repoRoot,
      files,
      promptText: '',
      truncated: false,
    }
  }

  for (const file of files)
    await enrichFile(repoRoot, file, Boolean(options.stagedOnly))

  const manifest = buildManifest(files)
  const diffHeader = '\n\nDiffs:'
  const diffBudget = Math.max(0, MAX_PROMPT_CHARS - manifest.text.length - diffHeader.length - COMPACTION_NOTICE.length - 2)
  const diffs = buildDiffContext(files, diffBudget)
  const truncated = manifest.compacted || diffs.compacted
  const sections = [manifest.text, diffHeader, diffs.text]
  if (truncated)
    sections.push(`\n\n${COMPACTION_NOTICE}`)

  return {
    repoRoot,
    files,
    promptText: sections.join(''),
    truncated,
  }
}
