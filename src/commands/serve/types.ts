export interface ServeOptions {
  directory: string
  port: number
  address: string
  upload: boolean
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

export interface ServeWorkspace {
  listDirectory: (relativePath: string) => Promise<ServeDirectoryListing>
  openFile: (relativePath: string) => Promise<ServeFile>
  readTextPreview: (relativePath: string) => Promise<ServeTextPreview>
  uploadFile: (directoryPath: string, file: File) => Promise<ServeUploadResult>
}

export type ServeErrorCode
  = | 'INVALID_PATH'
    | 'INVALID_UPLOAD'
    | 'PATH_FORBIDDEN'
    | 'NOT_FOUND'
    | 'NOT_DIRECTORY'
    | 'NOT_FILE'
    | 'TOO_LARGE'
    | 'NAME_EXHAUSTED'
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
