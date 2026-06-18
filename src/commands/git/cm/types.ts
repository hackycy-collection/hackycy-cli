export type CommitLanguage = 'en' | 'zh'

export interface CmOptions {
  profile?: string
  lang?: CommitLanguage
  history?: boolean
  staged?: boolean
  stage?: boolean
  stageAll?: boolean
  push?: boolean | string
  stagePush?: boolean | string
  dryRun?: boolean
  body?: boolean
}

export interface FileChange {
  path: string
  originalPath?: string
  status: string
  indexStatus: string
  worktreeStatus: string
  binary: boolean
  sensitive: boolean
  diffSkippedReason?: string
  diff?: string
}

export interface ChangeSummary {
  repoRoot: string
  files: FileChange[]
  promptText: string
  truncated: boolean
}
