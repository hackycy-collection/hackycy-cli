import type { DirectoryEntry, DirectoryListing } from './api'
import type { SortDirection, SortKey, UploadTask, ViewMode } from './types'
import { ArrowDown, ArrowUp, FolderOpen, Grid2X2, HardDrive, List, LoaderCircle, Moon, RefreshCw, Sun, Upload } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Button } from '../../../shared/web/components/ui/button'
import { Tooltip } from '../../../shared/web/components/ui/tooltip'
import { cn } from '../../../shared/web/lib/utils'
import { apiJson, uploadFile } from './api'
import { FileBrowser } from './components/file-browser'
import { PreviewSheet } from './components/preview-sheet'
import { UploadQueue } from './components/upload-queue'

interface RouteState {
  directoryPath: string
  previewPath?: string
  error?: string
}

function routeState(): RouteState {
  const pathname = window.location.pathname
  const encodedPath = pathname === '/' || pathname === '/browse'
    ? ''
    : pathname.startsWith('/browse/') ? pathname.slice('/browse/'.length) : ''
  try {
    return {
      directoryPath: encodedPath.split('/').filter(Boolean).map(decodeURIComponent).join('/'),
      previewPath: new URLSearchParams(window.location.search).get('preview') ?? undefined,
    }
  }
  catch {
    return { directoryPath: '', error: 'The browser path is not valid URL encoding.' }
  }
}

function browserUrl(relativePath: string): string {
  if (!relativePath)
    return '/'
  return `/browse/${relativePath.split('/').map(encodeURIComponent).join('/')}`
}

function useStoredValue<T>(key: string, fallback: () => T): [T, (value: T) => void] {
  const [value, setValue] = useState<T>(() => {
    try {
      const stored = localStorage.getItem(key)
      return stored === null ? fallback() : JSON.parse(stored) as T
    }
    catch {
      return fallback()
    }
  })
  const update = (next: T): void => {
    setValue(next)
    localStorage.setItem(key, JSON.stringify(next))
  }
  return [value, update]
}

function entryRank(entry: DirectoryEntry): number {
  return entry.kind === 'directory' ? 0 : entry.kind === 'file' ? 1 : 2
}

function sortedEntries(entries: DirectoryEntry[], key: SortKey, direction: SortDirection): DirectoryEntry[] {
  const multiplier = direction === 'asc' ? 1 : -1
  return [...entries].sort((left, right) => {
    const rank = entryRank(left) - entryRank(right)
    if (rank !== 0)
      return rank
    if (key === 'name')
      return multiplier * (left.name.localeCompare(right.name, undefined, { sensitivity: 'base' }) || left.name.localeCompare(right.name))
    if (key === 'size')
      return multiplier * ((left.size ?? -1) - (right.size ?? -1)) || left.name.localeCompare(right.name)
    return multiplier * ((Date.parse(left.modifiedAt ?? '') || 0) - (Date.parse(right.modifiedAt ?? '') || 0)) || left.name.localeCompare(right.name)
  })
}

export function App(): React.JSX.Element {
  const initialRoute = useMemo(routeState, [])
  const [directoryPath, setDirectoryPath] = useState(initialRoute.directoryPath)
  const [previewPath, setPreviewPath] = useState(initialRoute.previewPath)
  const [routeError, setRouteError] = useState(initialRoute.error)
  const [listing, setListing] = useState<DirectoryListing>()
  const [loadError, setLoadError] = useState<string>()
  const [loading, setLoading] = useState(true)
  const [reloadVersion, setReloadVersion] = useState(0)
  const [viewMode, setViewMode] = useStoredValue<ViewMode>('ycy-serve-view', () => 'list')
  const [sortKey, setSortKey] = useStoredValue<SortKey>('ycy-serve-sort', () => 'name')
  const [sortDirection, setSortDirection] = useStoredValue<SortDirection>('ycy-serve-sort-direction', () => 'asc')
  const [theme, setTheme] = useStoredValue<'light' | 'dark'>('ycy-serve-theme', () => matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
  const [uploadTasks, setUploadTasks] = useState<UploadTask[]>([])
  const [uploadActive, setUploadActive] = useState(false)
  const [dragging, setDragging] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, [theme])

  useEffect(() => {
    const popstate = (): void => {
      const next = routeState()
      setDirectoryPath(next.directoryPath)
      setPreviewPath(next.previewPath)
      setRouteError(next.error)
    }
    window.addEventListener('popstate', popstate)
    return () => window.removeEventListener('popstate', popstate)
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    setLoadError(undefined)
    void apiJson<DirectoryListing>(`/api/directory?path=${encodeURIComponent(directoryPath)}`, { signal: controller.signal })
      .then((next) => {
        setListing(next)
        setRouteError(undefined)
      })
      .catch((cause) => {
        if (!controller.signal.aborted)
          setLoadError(cause instanceof Error ? cause.message : String(cause))
      })
      .finally(() => {
        if (!controller.signal.aborted)
          setLoading(false)
      })
    return () => controller.abort()
  }, [directoryPath, reloadVersion])

  const navigate = (path: string): void => {
    history.pushState({}, '', browserUrl(path))
    setDirectoryPath(path)
    setPreviewPath(undefined)
    setRouteError(undefined)
  }

  const openPreview = (entry: DirectoryEntry): void => {
    const url = new URL(window.location.href)
    url.searchParams.set('preview', entry.path)
    history.pushState({ servePreview: true }, '', url)
    setPreviewPath(entry.path)
  }

  const closePreview = (): void => {
    if (history.state?.servePreview) {
      history.back()
      return
    }
    const url = new URL(window.location.href)
    url.searchParams.delete('preview')
    history.replaceState({}, '', url)
    setPreviewPath(undefined)
  }

  const changeSort = (key: SortKey): void => {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
      return
    }
    setSortKey(key)
    setSortDirection('asc')
  }

  const updateUploadTask = (id: string, update: Partial<UploadTask>): void => {
    setUploadTasks(current => current.map(task => task.id === id ? { ...task, ...update } : task))
  }

  const enqueueUploads = useCallback(async (files: File[]) => {
    if (files.length === 0 || uploadActive || !listing?.uploadEnabled)
      return
    const tasks = files.map<UploadTask>(file => ({
      id: crypto.randomUUID(),
      filename: file.name,
      status: file.size > listing.maxUploadBytes ? 'error' : 'queued',
      progress: 0,
      detail: file.size > listing.maxUploadBytes ? 'File exceeds the 1 GiB limit' : undefined,
    }))
    setUploadTasks(tasks)
    setUploadActive(true)
    const pending = tasks.map((task, index) => ({ task, file: files[index]! })).filter(item => item.task.status === 'queued')
    let cursor = 0
    const worker = async (): Promise<void> => {
      while (cursor < pending.length) {
        const item = pending[cursor++]!
        updateUploadTask(item.task.id, { status: 'uploading', progress: 0, detail: 'Uploading' })
        try {
          const result = await uploadFile(directoryPath, item.file, progress => updateUploadTask(item.task.id, { progress }))
          updateUploadTask(item.task.id, {
            status: 'done',
            progress: 100,
            detail: result.filename === item.file.name ? 'Uploaded' : `Saved as ${result.filename}`,
          })
        }
        catch (cause) {
          updateUploadTask(item.task.id, {
            status: 'error',
            detail: cause instanceof Error ? cause.message : String(cause),
          })
        }
      }
    }
    await Promise.all(Array.from({ length: Math.min(3, pending.length) }, () => worker()))
    setUploadActive(false)
    setReloadVersion(version => version + 1)
  }, [directoryPath, listing, uploadActive])

  useEffect(() => {
    if (!listing?.uploadEnabled)
      return
    let dragDepth = 0
    const enter = (event: DragEvent): void => {
      if (!event.dataTransfer?.types.includes('Files'))
        return
      event.preventDefault()
      dragDepth++
      setDragging(true)
    }
    const leave = (event: DragEvent): void => {
      event.preventDefault()
      dragDepth = Math.max(0, dragDepth - 1)
      if (dragDepth === 0)
        setDragging(false)
    }
    const over = (event: DragEvent): void => event.preventDefault()
    const drop = (event: DragEvent): void => {
      event.preventDefault()
      dragDepth = 0
      setDragging(false)
      void enqueueUploads(Array.from(event.dataTransfer?.files ?? []))
    }
    document.addEventListener('dragenter', enter)
    document.addEventListener('dragleave', leave)
    document.addEventListener('dragover', over)
    document.addEventListener('drop', drop)
    return () => {
      document.removeEventListener('dragenter', enter)
      document.removeEventListener('dragleave', leave)
      document.removeEventListener('dragover', over)
      document.removeEventListener('drop', drop)
    }
  }, [enqueueUploads, listing?.uploadEnabled])

  const entries = useMemo(() => sortedEntries(listing?.entries ?? [], sortKey, sortDirection), [listing?.entries, sortDirection, sortKey])
  const previewEntry = listing?.entries.find(entry => entry.kind === 'file' && entry.path === previewPath)
  const visibleError = routeError ?? loadError

  return (
    <div className="flex h-dvh min-w-0 flex-col overflow-hidden bg-background text-foreground">
      <header className="flex h-12 shrink-0 items-center gap-2 border-b border-border bg-background px-3">
        <div className="flex min-w-0 items-center gap-2">
          <span className="flex size-7 shrink-0 items-center justify-center rounded-md bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-950"><HardDrive className="size-4" /></span>
          <div className="min-w-0 leading-tight">
            <div className="hidden text-[10px] font-semibold text-muted-foreground sm:block">HACKYCY CLI · FILE SERVER</div>
            <div className="truncate text-sm font-semibold" title={listing?.rootName}>{listing?.rootName ?? 'File Server'}</div>
          </div>
        </div>
        <div role="toolbar" aria-label="File browser controls" className="ml-auto flex items-center gap-0.5">
          <label className="flex items-center gap-1 rounded-md border border-border bg-background px-1 text-xs text-muted-foreground sm:px-2">
            <span className="hidden sm:inline">Sort</span>
            <select aria-label="Sort files" value={sortKey} className="h-7 w-14 bg-transparent text-foreground outline-none sm:w-auto" onChange={event => changeSort(event.target.value as SortKey)}>
              <option value="name">Name</option>
              <option value="size">Size</option>
              <option value="modified">Modified</option>
            </select>
          </label>
          <Tooltip label={`Sort ${sortDirection === 'asc' ? 'descending' : 'ascending'}`}>
            <Button aria-label={`Sort ${sortDirection === 'asc' ? 'descending' : 'ascending'}`} size="icon" variant="ghost" onClick={() => setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')}>
              {sortDirection === 'asc' ? <ArrowUp className="size-4" /> : <ArrowDown className="size-4" />}
            </Button>
          </Tooltip>
          <span className="mx-1 h-4 w-px bg-border" />
          <div className="flex rounded-md border border-border p-0.5">
            <Tooltip label="List view"><Button aria-label="List view" aria-pressed={viewMode === 'list'} className={cn('size-7', viewMode === 'list' && 'bg-muted')} size="icon" variant="ghost" onClick={() => setViewMode('list')}><List className="size-3.5" /></Button></Tooltip>
            <Tooltip label="Grid view"><Button aria-label="Grid view" aria-pressed={viewMode === 'grid'} className={cn('size-7', viewMode === 'grid' && 'bg-muted')} size="icon" variant="ghost" onClick={() => setViewMode('grid')}><Grid2X2 className="size-3.5" /></Button></Tooltip>
          </div>
          <Tooltip label="Refresh directory"><Button aria-label="Refresh directory" size="icon" variant="ghost" disabled={loading} onClick={() => setReloadVersion(version => version + 1)}><RefreshCw className={cn('size-4', loading && 'animate-spin')} /></Button></Tooltip>
          {listing?.uploadEnabled && (
            <>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={(event) => {
                  void enqueueUploads(Array.from(event.currentTarget.files ?? []))
                  event.currentTarget.value = ''
                }}
              />
              <Button aria-label="Upload files" className="ml-1" variant="default" disabled={uploadActive} onClick={() => fileInputRef.current?.click()}>
                <Upload className="size-4" />
                <span className="hidden sm:inline">Upload</span>
              </Button>
            </>
          )}
          <Tooltip label={`Use ${theme === 'light' ? 'dark' : 'light'} theme`}><Button aria-label={`Use ${theme === 'light' ? 'dark' : 'light'} theme`} size="icon" variant="ghost" onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>{theme === 'light' ? <Moon className="size-4" /> : <Sun className="size-4" />}</Button></Tooltip>
        </div>
      </header>

      <Breadcrumb path={listing?.path ?? directoryPath} rootName={listing?.rootName ?? 'Root'} onNavigate={navigate} />

      {visibleError && (
        <div role="alert" className="shrink-0 border-b border-red-300 bg-red-50 px-3 py-2 text-xs text-red-800 dark:border-red-950 dark:bg-red-950 dark:text-red-200">
          {visibleError}
        </div>
      )}

      <main className="min-h-0 flex-1 bg-feed">
        {loading && !listing && <CenteredState icon={<LoaderCircle className="size-5 animate-spin" />} label="Loading directory" />}
        {!loading && visibleError && !listing && <CenteredState icon={<FolderOpen className="size-6" />} label="Directory unavailable" />}
        {listing && entries.length === 0 && !loading && <CenteredState icon={<FolderOpen className="size-7" />} label="This directory is empty" />}
        {listing && entries.length > 0 && (
          <FileBrowser
            entries={entries}
            viewMode={viewMode}
            sortKey={sortKey}
            direction={sortDirection}
            onSort={changeSort}
            onOpenDirectory={entry => entry.path && navigate(entry.path)}
            onOpenFile={openPreview}
          />
        )}
      </main>

      <footer className="flex h-7 shrink-0 items-center border-t border-border bg-background px-3 text-[11px] text-muted-foreground">
        <span>{`${listing?.entries.length ?? 0} items`}</span>
        <span className="ml-auto truncate">{`/${listing?.path ?? directoryPath}`}</span>
      </footer>

      <PreviewSheet entry={previewEntry} onClose={closePreview} />
      <UploadQueue tasks={uploadTasks} active={uploadActive} onClear={() => setUploadTasks([])} />
      {dragging && (
        <div className="pointer-events-none fixed inset-2 z-[80] flex items-center justify-center rounded-md border-2 border-dashed border-cyan-500 bg-cyan-50/95 text-sm font-semibold text-cyan-900 dark:bg-cyan-950/95 dark:text-cyan-100">
          {`Drop files into /${listing?.path ?? ''}`}
        </div>
      )}
    </div>
  )
}

function Breadcrumb({ path, rootName, onNavigate }: { path: string, rootName: string, onNavigate: (path: string) => void }): React.JSX.Element {
  const segments = path.split('/').filter(Boolean)
  return (
    <nav aria-label="Breadcrumb" className="flex h-10 shrink-0 items-center gap-1 overflow-x-auto border-b border-border bg-muted/40 px-3 text-xs">
      <button type="button" className="shrink-0 font-medium hover:text-cyan-700 dark:hover:text-cyan-300" onClick={() => onNavigate('')}>{rootName}</button>
      {segments.map((segment, index) => {
        const segmentPath = segments.slice(0, index + 1).join('/')
        return (
          <span key={segmentPath} className="flex shrink-0 items-center gap-1">
            <span className="text-muted-foreground">/</span>
            <button type="button" className="max-w-48 truncate font-medium hover:text-cyan-700 dark:hover:text-cyan-300" title={segment} onClick={() => onNavigate(segmentPath)}>{segment}</button>
          </span>
        )
      })}
    </nav>
  )
}

function CenteredState({ icon, label }: { icon: React.ReactNode, label: string }): React.JSX.Element {
  return (
    <div className="flex h-full min-h-48 flex-col items-center justify-center gap-2 text-sm text-muted-foreground">
      {icon}
      <span>{label}</span>
    </div>
  )
}
