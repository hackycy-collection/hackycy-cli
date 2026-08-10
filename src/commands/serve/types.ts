export interface ServeOptions {
  directory: string
  port: number
  address: string
  manage: boolean
  safeHtml: boolean
  accounts: string[]
}

export type ServeEntryKind = 'directory' | 'file' | 'unavailable'
export type ServePreviewKind = 'image' | 'video' | 'audio' | 'pdf' | 'text' | 'none'

export interface ServeDirectoryEntry {
  name: string
  path: string
  kind: ServeEntryKind
  isSymlink: boolean
  size?: number
  modifiedAt?: string
  mimeType?: string
  previewKind: ServePreviewKind
  syntaxLanguage?: string
  browseUrl?: string
  fileUrl?: string
  thumbnailUrl?: string
  downloadUrl?: string
  extractable: boolean
}

export interface ServeDirectoryListing {
  rootName: string
  path: string
  parentPath?: string
  entries: ServeDirectoryEntry[]
}

export interface ServeFile {
  name: string
  size: number
  modifiedAt: Date
  mimeType: string
  body: Blob
}

export type ServeTextPreview
  = | { status: 'ready', text: string, encoding: 'utf-8' | 'utf-16le' | 'utf-16be', size: number, revision: string }
    | { status: 'too_large', size: number, maxBytes: number }
    | { status: 'binary', size: number }

export interface ServeTextSaveResult {
  revision: string
  size: number
  modifiedAt: Date
  encoding: 'utf-8' | 'utf-16le' | 'utf-16be'
}

export interface ServeUploadResult {
  filename: string
  path: string
  size: number
}

export interface ServeStreamWriteOptions {
  signal?: AbortSignal
  onProgress?: (bytesWritten: number) => void
}

export interface ServeArchiveExtractOptions {
  signal?: AbortSignal
  onProgress?: (progress: number) => void
  onInspect?: (details: { uncompressedBytes: number, entryCount: number }) => void
}

export interface ServeArchiveExtractResult {
  archivePath: string
  destinationPath: string
  uncompressedBytes: number
  entryCount: number
}

export type ServeExtractionStatus = 'queued' | 'running' | 'done' | 'error' | 'cancelled'

export interface ServeExtractionTask {
  id: string
  archivePath: string
  status: ServeExtractionStatus
  progress?: number
  uncompressedBytes?: number
  entryCount?: number
  destinationPath?: string
  createdAt: string
  startedAt?: string
  finishedAt?: string
  error?: string
}

export interface ServeExtractionManager {
  list: () => ServeExtractionTask[]
  enqueue: (paths: string[]) => Promise<ServeExtractionTask[]>
  cancel: (id: string) => ServeExtractionTask | undefined
  retry: (id: string) => Promise<ServeExtractionTask>
  clearTerminal: () => void
  subscribe: (listener: (tasks: ServeExtractionTask[]) => void) => () => void
  close: () => Promise<void>
}

export type ServeDownloadStatus = 'queued' | 'running' | 'done' | 'error' | 'cancelled'

export interface ServeDownloadRequest {
  url: string
  directoryPath: string
  filename?: string
}

export interface ServeDownloadTask {
  id: string
  url: string
  directoryPath: string
  filename?: string
  status: ServeDownloadStatus
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

export interface ServeDownloadManager {
  list: () => ServeDownloadTask[]
  enqueue: (request: ServeDownloadRequest) => Promise<ServeDownloadTask>
  cancel: (id: string) => ServeDownloadTask | undefined
  retry: (id: string) => Promise<ServeDownloadTask>
  clearTerminal: () => void
  subscribe: (listener: (tasks: ServeDownloadTask[]) => void) => () => void
  close: () => Promise<void>
}

export type ServeOperation
  = | { action: 'create-directory', parentPath: string, name: string }
    | { action: 'rename', path: string, newName: string }
    | { action: 'copy', paths: string[], destinationPath: string }
    | { action: 'move', paths: string[], destinationPath: string }
    | { action: 'delete', paths: string[] }

export type ServeOperationItem
  = | {
    status: 'ok'
    sourcePath?: string
    destinationPath?: string
  }
  | {
    status: 'error'
    sourcePath?: string
    destinationPath?: string
    error: { code: ServeErrorCode, message: string }
  }

export interface ServeOperationResult {
  action: ServeOperation['action']
  items: ServeOperationItem[]
}

export interface ServeWorkspace {
  listDirectory: (relativePath: string) => Promise<ServeDirectoryListing>
  openFile: (relativePath: string) => Promise<ServeFile>
  readTextPreview: (relativePath: string) => Promise<ServeTextPreview>
  saveTextFile: (relativePath: string, text: string, revision: string) => Promise<ServeTextSaveResult>
  uploadFile: (directoryPath: string, file: File) => Promise<ServeUploadResult>
  writeFileStream: (directoryPath: string, filename: string, stream: ReadableStream<Uint8Array>, options?: ServeStreamWriteOptions) => Promise<ServeUploadResult>
  extractArchive: (archivePath: string, options?: ServeArchiveExtractOptions) => Promise<ServeArchiveExtractResult>
  applyOperation: (operation: ServeOperation) => Promise<ServeOperationResult>
}

export type ServeErrorCode
  = | 'INVALID_PATH'
    | 'INVALID_UPLOAD'
    | 'INVALID_NAME'
    | 'INVALID_OPERATION'
    | 'PATH_FORBIDDEN'
    | 'NOT_FOUND'
    | 'NOT_DIRECTORY'
    | 'NOT_FILE'
    | 'TOO_LARGE'
    | 'PRECONDITION_REQUIRED'
    | 'REVISION_MISMATCH'
    | 'UNSUPPORTED_TEXT'
    | 'NAME_EXHAUSTED'
    | 'ALREADY_EXISTS'
    | 'ROOT_IMMUTABLE'
    | 'UNAVAILABLE'
    | 'UNSUPPORTED_ARCHIVE'
    | 'INVALID_ARCHIVE'
    | 'ENCRYPTED_ARCHIVE'
    | 'INSUFFICIENT_SPACE'

export class ServeWorkspaceError extends Error {
  constructor(
    readonly code: ServeErrorCode,
    message: string,
    options?: ErrorOptions,
  ) {
    super(message, options)
    this.name = 'ServeWorkspaceError'
  }
}
