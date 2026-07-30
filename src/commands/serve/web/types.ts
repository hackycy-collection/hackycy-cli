export type ViewMode = 'list' | 'grid'
export type SortKey = 'name' | 'size' | 'modified'
export type SortDirection = 'asc' | 'desc'

export interface UploadTask {
  id: string
  filename: string
  status: 'queued' | 'uploading' | 'done' | 'error'
  progress: number
  detail?: string
}
