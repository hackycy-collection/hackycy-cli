import type { ChangeSummary, FileChange } from './types'
import path from 'node:path'

const MAX_TOTAL_CHARS = 24_000
const MAX_FILE_DIFF_CHARS = 6_000
const MAX_FILES = 80
const LARGE_FILE_BYTES = 200_000

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

export async function getRepoRoot(): Promise<string> {
  return (await runGit(['rev-parse', '--show-toplevel'])).trim()
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
  await runGit(['add', '-A', '--', ...filePaths], repoRoot)
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

function truncateDiff(diff: string): { text: string, truncated: boolean } {
  if (diff.length <= MAX_FILE_DIFF_CHARS)
    return { text: diff, truncated: false }

  return {
    text: `${diff.slice(0, MAX_FILE_DIFF_CHARS)}\n...[diff truncated for token budget]`,
    truncated: true,
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

async function enrichFile(repoRoot: string, file: FileChange, stagedOnly: boolean): Promise<boolean> {
  if (file.sensitive) {
    file.diffSkippedReason = 'sensitive file path'
    return false
  }

  if (file.binary) {
    file.diffSkippedReason = 'binary file'
    return false
  }

  if (isGeneratedPath(file.path)) {
    file.diffSkippedReason = 'generated or lock file'
    return false
  }

  const size = await fileSize(repoRoot, file.path)
  if (size !== undefined && size > LARGE_FILE_BYTES) {
    file.diffSkippedReason = 'large file'
    return false
  }

  const diff = await getFileDiff(repoRoot, file, stagedOnly)
  if (!diff.trim()) {
    file.diffSkippedReason = 'no textual diff'
    return false
  }

  if (/^Binary files .+ differ$/m.test(diff)) {
    file.diffSkippedReason = 'binary diff'
    file.binary = true
    return false
  }

  const truncated = truncateDiff(diff)
  file.diff = truncated.text
  if (truncated.truncated)
    file.diffSkippedReason = 'diff truncated'

  return truncated.truncated
}

export async function collectChangeSummary(options: { stagedOnly?: boolean } = {}): Promise<ChangeSummary> {
  const repoRoot = await getRepoRoot()
  const statusOutput = await runGit(['status', '--porcelain=v1', '-z', '--untracked-files=all'], repoRoot)
  let files = parseStatus(statusOutput)
  let omittedFileCount = 0

  if (options.stagedOnly) {
    files = files.filter(file => file.indexStatus !== ' ' && file.indexStatus !== '?')
  }

  files = await filterSubmodules(repoRoot, files)

  if (files.length > MAX_FILES) {
    omittedFileCount = files.length - MAX_FILES
    files = files.slice(0, MAX_FILES)
  }

  if (files.length === 0) {
    return {
      repoRoot,
      files,
      promptText: '',
      truncated: false,
    }
  }

  let truncated = omittedFileCount > 0
  for (const file of files)
    truncated = (await enrichFile(repoRoot, file, Boolean(options.stagedOnly))) || truncated

  const statusLines = files.map(file => file.status)
  const sections = [
    'Changed files:',
    ...statusLines,
    '',
    'Diffs:',
  ]

  let remaining = MAX_TOTAL_CHARS - sections.join('\n').length
  for (const file of files) {
    if (!file.diff) {
      sections.push(`\n# ${file.status}\n[diff omitted: ${file.diffSkippedReason ?? 'not available'}]`)
      continue
    }

    const section = `\n# ${file.status}\n${file.diff}`
    if (section.length > remaining) {
      sections.push(`\n# ${file.status}\n[diff omitted: total token budget exceeded]`)
      truncated = true
      continue
    }

    sections.push(section)
    remaining -= section.length
  }

  if (truncated)
    sections.push('\nSome diffs were omitted or truncated to save tokens.')
  if (omittedFileCount > 0)
    sections.push(`${omittedFileCount} changed file(s) were omitted because the file count exceeded ${MAX_FILES}.`)

  return {
    repoRoot,
    files,
    promptText: sections.join('\n'),
    truncated,
  }
}
