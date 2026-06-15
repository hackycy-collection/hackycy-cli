export type CommitLanguage = 'en' | 'zh'

export interface AiProfile {
  baseURL: string
  model: string
  apiKey: string // encrypted
  temperature?: number
  timeoutMs?: number
  maxOutputTokens?: number
}

export interface AiConfig {
  defaultProfile?: string
  profiles: Record<string, AiProfile>
}

export interface ResolvedAiProfile {
  name: string
  baseURL: string
  model: string
  apiKey: string
  temperature: number
  timeoutMs: number
  maxOutputTokens: number
}

export interface CmOptions {
  profile?: string
  lang?: CommitLanguage
  history?: boolean
  staged?: boolean
  commit?: boolean
  stageAll?: boolean
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
