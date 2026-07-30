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
  browseUrl?: string
  fileUrl?: string
  downloadUrl?: string
}

export interface DirectoryListing {
  version: 1
  rootName: string
  path: string
  parentPath?: string
  uploadEnabled: boolean
  maxUploadBytes: number
  entries: DirectoryEntry[]
}

export type TextPreview
  = | { version: 1, status: 'ready', text: string, encoding: string, size: number }
    | { version: 1, status: 'too_large', size: number, maxBytes: number }
    | { version: 1, status: 'binary', size: number }

export interface UploadResult {
  version: 1
  filename: string
  path: string
  size: number
}

interface ErrorBody {
  error?: { message?: string }
}

export async function apiJson<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  const body = await response.json() as T & ErrorBody
  if (!response.ok)
    throw new Error(body.error?.message ?? `Request failed (${response.status})`)
  return body
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
      const body = request.response as ErrorBody | null
      reject(new Error(body?.error?.message ?? `Upload failed (${request.status})`))
    }
    request.onerror = () => reject(new Error('Upload failed: network error'))
    const body = new FormData()
    body.set('file', file)
    request.send(body)
  })
}
