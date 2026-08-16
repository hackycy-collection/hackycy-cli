export type ViewMode = 'list' | 'grid'
export type SortKey = 'name' | 'size' | 'modified'
export type SortDirection = 'asc' | 'desc'

export interface ActivityTask {
  id: string
  label: string
  status: 'queued' | 'running' | 'done' | 'error'
  progress?: number
  detail?: string
}
