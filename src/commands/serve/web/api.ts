export type EntryKind = 'directory' | 'file' | 'unavailable'
export type PreviewKind = 'image' | 'video' | 'audio' | 'pdf' | 'text' | 'none'

export interface DirectoryEntry {
  name: string
  path: string
  kind: EntryKind
  isSymlink: boolean
  size?: number
  modifiedAt?: string
  mimeType?: string
  previewKind: PreviewKind
  syntaxLanguage?: string
  browseUrl?: string
  fileUrl?: string
  thumbnailUrl?: string
  downloadUrl?: string
  extractable: boolean
}

export interface DirectoryListing {
  version: 1
  rootName: string
  path: string
  parentPath?: string
  managementEnabled: boolean
  maxUploadBytes: number
  entries: DirectoryEntry[]
}

export type SessionState
  = | { version: 1, authenticationEnabled: false }
    | { version: 1, authenticationEnabled: true, authenticated: false }
    | { version: 1, authenticationEnabled: true, authenticated: true, account: { username: string } }

export type TextPreview
  = | { version: 1, status: 'ready', text: string, encoding: 'utf-8' | 'utf-16le' | 'utf-16be', size: number, revision: string }
    | { version: 1, status: 'too_large', size: number, maxBytes: number }
    | { version: 1, status: 'binary', size: number }

export interface TextSaveResult {
  version: 1
  revision: string
  size: number
  modifiedAt: string
  encoding: 'utf-8' | 'utf-16le' | 'utf-16be'
}

export interface UploadResult {
  version: 1
  filename: string
  path: string
  size: number
}

export type OperationCommand
  = | { action: 'create-directory', parentPath: string, name: string }
    | { action: 'rename', path: string, newName: string }
    | { action: 'copy', paths: string[], destinationPath: string }
    | { action: 'move', paths: string[], destinationPath: string }
    | { action: 'delete', paths: string[] }

export type OperationItem
  = | { status: 'ok', sourcePath?: string, destinationPath?: string }
    | { status: 'error', sourcePath?: string, destinationPath?: string, error: { code: string, message: string } }

export interface OperationResult {
  version: 1
  action: OperationCommand['action']
  items: OperationItem[]
}

export type DownloadStatus = 'queued' | 'running' | 'done' | 'error' | 'cancelled'

export interface DownloadTask {
  id: string
  url: string
  directoryPath: string
  filename?: string
  status: DownloadStatus
  bytesDownloaded: number
  totalBytes?: number
  progress?: number
  speedBytesPerSecond?: number
  destinationPath?: string
  createdAt: string
  startedAt?: string
  finishedAt?: string
  error?: string
}

export interface DownloadList {
  version: 1
  tasks: DownloadTask[]
}

export interface DownloadResponse {
  version: 1
  task: DownloadTask
}

export type ExtractionStatus = 'queued' | 'running' | 'done' | 'error' | 'cancelled'

export interface ExtractionTask {
  id: string
  archivePath: string
  destinationPath?: string
  status: ExtractionStatus
  progress?: number
  uncompressedBytes?: number
  entryCount?: number
  createdAt: string
  startedAt?: string
  finishedAt?: string
  error?: string
}

export interface ExtractionList {
  version: 1
  tasks: ExtractionTask[]
}

export interface ExtractionBatchResponse {
  version: 1
  tasks: ExtractionTask[]
}

export interface ExtractionResponse {
  version: 1
  task: ExtractionTask
}

export class ApiError extends Error {
  constructor(readonly status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

interface ErrorBody {
  error?: { message?: string }
}

export async function apiJson<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  const body = response.status === 204 ? undefined : await response.json() as T & ErrorBody
  if (!response.ok) {
    if (response.status === 401 && url !== '/api/session')
      window.dispatchEvent(new Event('serve-authentication-required'))
    throw new ApiError(response.status, body?.error?.message ?? `Request failed (${response.status})`)
  }
  return body as T
}

export function applyOperation(operation: OperationCommand): Promise<OperationResult> {
  return apiJson<OperationResult>('/api/operations', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(operation),
  })
}

export function createDownload(url: string, directoryPath: string, filename?: string): Promise<DownloadResponse> {
  return apiJson<DownloadResponse>('/api/downloads', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url, directoryPath, ...(filename ? { filename } : {}) }),
  })
}

export function cancelDownload(id: string): Promise<DownloadResponse> {
  return apiJson<DownloadResponse>(`/api/downloads/${encodeURIComponent(id)}/cancel`, { method: 'POST' })
}

export function retryDownload(id: string): Promise<DownloadResponse> {
  return apiJson<DownloadResponse>(`/api/downloads/${encodeURIComponent(id)}/retry`, { method: 'POST' })
}

export async function clearDownloads(): Promise<void> {
  await apiJson<void>('/api/downloads?terminal=1', { method: 'DELETE' })
}

export function createExtractions(paths: string[]): Promise<ExtractionBatchResponse> {
  return apiJson<ExtractionBatchResponse>('/api/extractions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ paths }),
  })
}

export function cancelExtraction(id: string): Promise<ExtractionResponse> {
  return apiJson<ExtractionResponse>(`/api/extractions/${encodeURIComponent(id)}/cancel`, { method: 'POST' })
}

export function retryExtraction(id: string): Promise<ExtractionResponse> {
  return apiJson<ExtractionResponse>(`/api/extractions/${encodeURIComponent(id)}/retry`, { method: 'POST' })
}

export async function clearExtractions(): Promise<void> {
  await apiJson<void>('/api/extractions?terminal=1', { method: 'DELETE' })
}

export function uploadFile(
  directoryPath: string,
  file: File,
  onProgress: (progress: number) => void,
): Promise<UploadResult> {
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest()
    request.open('POST', `/api/upload?path=${encodeURIComponent(directoryPath)}`)
    request.responseType = 'json'
    request.upload.onprogress = (event) => {
      if (event.lengthComputable)
        onProgress(Math.round(event.loaded / event.total * 100))
    }
    request.onload = () => {
      if (request.status >= 200 && request.status < 300) {
        resolve(request.response as UploadResult)
        return
      }
      if (request.status === 401)
        window.dispatchEvent(new Event('serve-authentication-required'))
      const body = request.response as ErrorBody | null
      reject(new Error(body?.error?.message ?? `Upload failed (${request.status})`))
    }
    request.onerror = () => reject(new Error('Upload failed: network error'))
    const body = new FormData()
    body.set('file', file)
    request.send(body)
  })
}
