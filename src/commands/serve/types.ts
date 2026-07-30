export interface ServeOptions {
  directory: string
  port: number
  address: string
  manage: boolean
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
  browseUrl?: string
  fileUrl?: string
  thumbnailUrl?: string
  downloadUrl?: string
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
  = | { status: 'ready', text: string, encoding: 'utf-8' | 'utf-16le' | 'utf-16be', size: number }
    | { status: 'too_large', size: number, maxBytes: number }
    | { status: 'binary', size: number }

export interface ServeUploadResult {
  filename: string
  path: string
  size: number
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
  uploadFile: (directoryPath: string, file: File) => Promise<ServeUploadResult>
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
    | 'NAME_EXHAUSTED'
    | 'ALREADY_EXISTS'
    | 'ROOT_IMMUTABLE'
    | 'UNAVAILABLE'

export class ServeWorkspaceError extends Error {
  constructor(
    readonly code: ServeErrorCode,
    message: string,
  ) {
    super(message)
    this.name = 'ServeWorkspaceError'
  }
}
