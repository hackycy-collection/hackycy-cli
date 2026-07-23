export type ComparisonStatus = 'added' | 'deleted' | 'modified' | 'unchanged'
export type ComparisonResultStatus = ComparisonStatus | 'issue'
export type ComparisonEntryKind = 'file' | 'symlink'
export type ComparisonItemKind = ComparisonEntryKind | 'issue'
export type ComparisonSide = 'baseline' | 'target'
export type WorkspacePhase = 'idle' | 'discovering' | 'comparing' | 'publishing' | 'ready' | 'canceled' | 'error'

export interface StatusCounts {
  added: number
  deleted: number
  modified: number
  unchanged: number
}

export interface SnapshotSummary {
  id: string
  baselineDirectory: string
  targetDirectory: string
  createdAt: string
  counts: StatusCounts
  issues: number
}

export interface ComparisonEntry {
  id: number
  path: string
  status: ComparisonStatus
  kind: ComparisonEntryKind
  baselineSize?: number
  targetSize?: number
}

export interface ComparisonIssue {
  id: number
  path: string
  status: 'issue'
  kind: 'issue'
  message: string
}

export type ComparisonListEntry = ComparisonEntry | ComparisonIssue

export interface EntryQuery {
  includeUnchanged?: boolean
  cursor?: string
  anchor?: number
  limit?: number
  statuses?: ComparisonResultStatus[]
  path?: string
  kinds?: ComparisonItemKind[]
}

export interface EntryPage {
  entries: ComparisonListEntry[]
  nextCursor?: string
}

export interface TreeQuery {
  path: string
}

export interface DirectoryTreeNode {
  kind: 'directory'
  name: string
  path: string
  counts: StatusCounts
  issues: number
}

export interface EntryTreeNode {
  kind: ComparisonItemKind
  name: string
  path: string
  id: number
  status: ComparisonResultStatus
  message?: string
}

export type TreeNode = DirectoryTreeNode | EntryTreeNode

export interface TreePage {
  children: TreeNode[]
}

export interface SearchPage {
  results: TreeNode[]
  truncated: boolean
}

export type EntryPresentation = 'text' | 'image' | 'binary' | 'symlink' | 'oversized' | 'stale' | 'issue'

export type EntryDetail
  = | (ComparisonEntry & {
    presentation: Exclude<EntryPresentation, 'issue'>
    baselineLinkTarget?: string
    targetLinkTarget?: string
  })
  | (ComparisonIssue & { presentation: 'issue' })

export type TextEncoding = 'utf-8' | 'utf-16le' | 'utf-16be'

export type TextContent
  = | { status: 'ready', text: string, encoding: TextEncoding, size: number, lineCount: number }
    | { status: 'guarded', size: number, lineCount: number }
    | { status: 'blocked', size: number, lineCount?: number }
    | { status: 'binary' | 'missing' | 'stale' }

export type BlobContent
  = | { status: 'ready', bytes: Uint8Array, mimeType: string, filename: string }
    | { status: 'missing' | 'stale' | 'unavailable' }

export interface WorkspaceProgress {
  discoveredEntries: number
  comparedEntries: number
  totalEntries?: number
  comparedBytes: number
  totalBytes?: number
  issues: number
}

export interface WorkspaceState {
  phase: WorkspacePhase
  snapshotId?: string
  error?: string
  progress?: WorkspaceProgress
}

export interface ComparisonSnapshot {
  readonly summary: SnapshotSummary
  list: (query: EntryQuery) => EntryPage
  tree: (query: TreeQuery) => TreePage
  search: (query: string, statuses?: ComparisonResultStatus[], limit?: number) => SearchPage
  detail: (entryId: number) => Promise<EntryDetail>
  content: (entryId: number, side: ComparisonSide, force?: boolean) => Promise<TextContent>
  blob: (entryId: number, side: ComparisonSide) => Promise<BlobContent>
}

export interface RefreshRun {
  result: Promise<ComparisonSnapshot>
  cancel: () => void
}

export interface ComparisonWorkspace {
  state: () => WorkspaceState
  refresh: () => RefreshRun
  snapshot: (id?: string) => ComparisonSnapshot | undefined
  observe: (listener: (state: WorkspaceState) => void) => () => void
}

export interface ComparisonWorkspaceOptions {
  baselineDirectory: string
  targetDirectory: string
  gitignore?: boolean
  exclusions?: string[]
}
