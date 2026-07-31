import type { Layout, LayoutChangedMeta } from 'react-resizable-panels'
import type { DirectoryEntry, DirectoryListing, OperationCommand, OperationResult } from './api'
import type { ExplorerClipboard, ExplorerSelection, NavigationHistory } from './explorer-state'
import type { ActivityTask, SortDirection, SortKey, ViewMode } from './types'
import * as DropdownMenu from '@radix-ui/react-dropdown-menu'
import {
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  ArrowUpDown,
  ChevronRight,
  ClipboardPaste,
  Copy,
  FolderOpen,
  FolderPlus,
  Grid2X2,
  HardDrive,
  List,
  LoaderCircle,
  Menu,
  Moon,
  MoreHorizontal,
  PanelRight,
  Pencil,
  RefreshCw,
  Scissors,
  Search,
  Sun,
  Trash2,
  Upload,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Group, Panel, Separator, usePanelRef } from 'react-resizable-panels'
import { v4 as uuidv4 } from 'uuid'
import { Button } from '../../../shared/web/components/ui/button'
import { Sheet, SheetContent } from '../../../shared/web/components/ui/sheet'
import { Tooltip } from '../../../shared/web/components/ui/tooltip'
import { cn } from '../../../shared/web/lib/utils'
import { apiJson, applyOperation, uploadFile } from './api'
import { ActivityCenter } from './components/activity-center'
import { DeleteDialog } from './components/delete-dialog'
import { FileBrowser } from './components/file-browser'
import { NavigationPane } from './components/navigation-pane'
import { PreviewPane, PreviewSheet } from './components/preview-sheet'
import { RenameDialog } from './components/rename-dialog'
import {
  clipboardOperation,
  createNavigationHistory,
  entryNameError,
  moveNavigation,
  operationActivities,
  parentDirectoryPath,
  pushNavigation,
  selectEntry,
  settleClipboard,
  visibleEntries,
} from './explorer-state'
import {
  NAVIGATION_PANEL_MAX_WIDTH,
  NAVIGATION_PANEL_MIN_WIDTH,
  NAVIGATION_PANEL_WIDTH_STORAGE_KEY,
  navigationPanelWidth,
} from './layout-state'

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
  return relativePath ? `/browse/${relativePath.split('/').map(encodeURIComponent).join('/')}` : '/'
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

function useMobile(): boolean {
  const [mobile, setMobile] = useState(() => matchMedia('(max-width: 899px)').matches)
  useEffect(() => {
    const query = matchMedia('(max-width: 899px)')
    const change = (): void => setMobile(query.matches)
    query.addEventListener('change', change)
    return () => query.removeEventListener('change', change)
  }, [])
  return mobile
}

export function App(): React.JSX.Element {
  const initialRoute = useMemo(routeState, [])
  const [initialNavigationPanelWidth] = useState(() => {
    try {
      return navigationPanelWidth(localStorage.getItem(NAVIGATION_PANEL_WIDTH_STORAGE_KEY))
    }
    catch {
      return navigationPanelWidth(null)
    }
  })
  const [directoryPath, setDirectoryPath] = useState(initialRoute.directoryPath)
  const [previewPath, setPreviewPath] = useState(initialRoute.previewPath)
  const [routeError, setRouteError] = useState(initialRoute.error)
  const [navigation, setNavigation] = useState<NavigationHistory>(() => createNavigationHistory(initialRoute.directoryPath))
  const navigationRef = useRef(navigation)
  const [listing, setListing] = useState<DirectoryListing>()
  const [loadError, setLoadError] = useState<string>()
  const [loading, setLoading] = useState(true)
  const [reloadVersion, setReloadVersion] = useState(0)
  const [viewMode, setViewMode] = useStoredValue<ViewMode>('ycy-serve-view', () => 'list')
  const [sortKey, setSortKey] = useStoredValue<SortKey>('ycy-serve-sort', () => 'name')
  const [sortDirection, setSortDirection] = useStoredValue<SortDirection>('ycy-serve-sort-direction', () => 'asc')
  const [theme, setTheme] = useStoredValue<'light' | 'dark'>('ycy-serve-theme', () => matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
  const [query, setQuery] = useState('')
  const [selection, setSelection] = useState<ExplorerSelection>({ paths: [] })
  const [clipboard, setClipboard] = useState<ExplorerClipboard>()
  const [creatingFolder, setCreatingFolder] = useState(false)
  const [renamingPath, setRenamingPath] = useState<string>()
  const [editingBusy, setEditingBusy] = useState(false)
  const [editingError, setEditingError] = useState<string>()
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [operationBusy, setOperationBusy] = useState(false)
  const [activities, setActivities] = useState<ActivityTask[]>([])
  const [uploadActive, setUploadActive] = useState(false)
  const [dragging, setDragging] = useState(false)
  const [mobileNavigation, setMobileNavigation] = useState(false)
  const [imageViewerOpen, setImageViewerOpen] = useState(false)
  const [toast, setToast] = useState<string>()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const activeResizeHandleRef = useRef<'navigation' | 'preview' | undefined>(undefined)
  const navigationPanelRef = usePanelRef()
  const mobile = useMobile()

  const saveNavigationPanelWidth = useCallback((_layout: Layout, meta: LayoutChangedMeta): void => {
    const activeHandle = activeResizeHandleRef.current
    activeResizeHandleRef.current = undefined
    if (!meta.isUserInteraction || activeHandle !== 'navigation')
      return
    const width = navigationPanelRef.current?.getSize().inPixels
    if (width === undefined)
      return
    try {
      localStorage.setItem(NAVIGATION_PANEL_WIDTH_STORAGE_KEY, JSON.stringify(Math.round(width)))
    }
    catch {}
  }, [navigationPanelRef])

  useEffect(() => {
    navigationRef.current = navigation
  }, [navigation])

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, [theme])

  useEffect(() => {
    history.replaceState({ serveCursor: 0 }, '', window.location.href)
    const popstate = (event: PopStateEvent): void => {
      const next = routeState()
      const cursor = typeof event.state?.serveCursor === 'number' ? event.state.serveCursor : undefined
      if (cursor !== undefined && navigationRef.current.paths[cursor] === next.directoryPath)
        setNavigation(current => ({ ...current, cursor }))
      else
        setNavigation(createNavigationHistory(next.directoryPath))
      setDirectoryPath(next.directoryPath)
      setPreviewPath(next.previewPath)
      setImageViewerOpen(false)
      setRouteError(next.error)
      setQuery('')
      setSelection({ paths: [] })
      setCreatingFolder(false)
      setRenamingPath(undefined)
      setEditingError(undefined)
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
        const available = new Set(next.entries.map(entry => entry.path))
        setSelection(current => ({
          paths: current.paths.filter(path => available.has(path)),
          anchorPath: current.anchorPath && available.has(current.anchorPath) ? current.anchorPath : undefined,
        }))
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

  useEffect(() => {
    if (!toast)
      return
    const timeout = window.setTimeout(() => setToast(undefined), 4000)
    return () => window.clearTimeout(timeout)
  }, [toast])

  const navigate = useCallback((path: string): void => {
    if (path === directoryPath) {
      setMobileNavigation(false)
      return
    }
    const next = pushNavigation(navigationRef.current, path)
    history.pushState({ serveCursor: next.cursor }, '', browserUrl(path))
    setNavigation(next)
    setDirectoryPath(path)
    setPreviewPath(undefined)
    setImageViewerOpen(false)
    setRouteError(undefined)
    setQuery('')
    setSelection({ paths: [] })
    setCreatingFolder(false)
    setRenamingPath(undefined)
    setEditingError(undefined)
    setMobileNavigation(false)
  }, [directoryPath])

  const moveHistory = (offset: -1 | 1): void => {
    const next = moveNavigation(navigationRef.current, offset)
    if (next.cursor === navigationRef.current.cursor)
      return
    offset === -1 ? history.back() : history.forward()
  }

  const openPreview = useCallback((entry: DirectoryEntry): void => {
    if (entry.kind !== 'file')
      return
    const url = new URL(window.location.href)
    url.searchParams.set('preview', entry.path)
    history.replaceState({ serveCursor: navigationRef.current.cursor }, '', url)
    setPreviewPath(entry.path)
    setImageViewerOpen(false)
    setSelection({ paths: [entry.path], anchorPath: entry.path })
  }, [])

  const openTreeFile = useCallback((entry: DirectoryEntry): void => {
    if (entry.kind !== 'file')
      return
    setMobileNavigation(false)
    const parentPath = parentDirectoryPath(entry.path)
    if (parentPath === directoryPath) {
      openPreview(entry)
      return
    }

    const next = pushNavigation(navigationRef.current, parentPath)
    const previewUrl = new URL(browserUrl(parentPath), window.location.origin)
    previewUrl.searchParams.set('preview', entry.path)
    history.pushState({ serveCursor: next.cursor }, '', previewUrl)
    setNavigation(next)
    setDirectoryPath(parentPath)
    setPreviewPath(entry.path)
    setImageViewerOpen(false)
    setRouteError(undefined)
    setQuery('')
    setSelection({ paths: [entry.path], anchorPath: entry.path })
    setCreatingFolder(false)
    setRenamingPath(undefined)
    setEditingError(undefined)
  }, [directoryPath, openPreview])

  const closePreview = useCallback((): void => {
    const url = new URL(window.location.href)
    url.searchParams.delete('preview')
    history.replaceState({ serveCursor: navigationRef.current.cursor }, '', url)
    setPreviewPath(undefined)
    setImageViewerOpen(false)
  }, [])

  const changeSort = (key: SortKey): void => {
    if (key === sortKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
      return
    }
    setSortKey(key)
    setSortDirection('asc')
  }

  const chooseSort = (key: SortKey): void => {
    if (key === sortKey)
      return
    setSortKey(key)
    setSortDirection('asc')
  }

  const updateActivity = (id: string, update: Partial<ActivityTask>): void => {
    setActivities(current => current.map(task => task.id === id ? { ...task, ...update } : task))
  }

  const runOperation = useCallback(async (label: string, command: OperationCommand): Promise<OperationResult | undefined> => {
    const id = uuidv4()
    setActivities(current => [{ id, label, status: 'running', detail: 'Working' }, ...current])
    setOperationBusy(true)
    try {
      const result = await applyOperation(command)
      const errors = result.items.filter(item => item.status === 'error')
      setActivities(current => current.flatMap(task => task.id === id ? operationActivities(id, label, result) : [task]))
      if (errors.length > 0)
        setToast(errors[0]!.error.message)
      setReloadVersion(version => version + 1)
      return result
    }
    catch (cause) {
      const message = cause instanceof Error ? cause.message : String(cause)
      updateActivity(id, { status: 'error', detail: message })
      setToast(message)
      return undefined
    }
    finally {
      setOperationBusy(false)
    }
  }, [])

  const entries = useMemo(
    () => visibleEntries(listing?.entries ?? [], query, sortKey, sortDirection),
    [listing?.entries, query, sortDirection, sortKey],
  )
  const entryMap = useMemo(() => new Map((listing?.entries ?? []).map(entry => [entry.path, entry])), [listing?.entries])
  const selectedSet = useMemo(() => new Set(selection.paths), [selection.paths])
  const selectedEntries = selection.paths.map(path => entryMap.get(path)).filter((entry): entry is DirectoryEntry => entry !== undefined)
  const previewEntry = entryMap.get(previewPath ?? '')
  const renamingEntry = entryMap.get(renamingPath ?? '')
  const managementEnabled = listing?.managementEnabled ?? false
  const visibleError = routeError ?? loadError

  const selectedPathsFor = (entry?: DirectoryEntry): string[] => entry && !selectedSet.has(entry.path) ? [entry.path] : selection.paths

  const startCreateFolder = (): void => {
    if (!managementEnabled || operationBusy)
      return
    setCreatingFolder(true)
    setRenamingPath(undefined)
    setEditingError(undefined)
    setSelection({ paths: [] })
  }

  const createFolder = async (name: string): Promise<void> => {
    const error = entryNameError(name)
    if (error) {
      setEditingError(error)
      return
    }
    setEditingBusy(true)
    const result = await runOperation(`Create ${name}`, { action: 'create-directory', parentPath: directoryPath, name })
    const item = result?.items[0]
    if (item?.status === 'ok') {
      setCreatingFolder(false)
      setEditingError(undefined)
      if (item.destinationPath)
        setSelection({ paths: [item.destinationPath], anchorPath: item.destinationPath })
    }
    else if (item?.status === 'error') {
      setEditingError(item.error.message)
    }
    setEditingBusy(false)
  }

  const startRename = (entry: DirectoryEntry): void => {
    if (!managementEnabled || operationBusy)
      return
    setSelection({ paths: [entry.path], anchorPath: entry.path })
    setCreatingFolder(false)
    setRenamingPath(entry.path)
    setEditingError(undefined)
  }

  const renameEntry = async (name: string): Promise<void> => {
    const entry = renamingEntry
    if (!entry)
      return
    const error = entryNameError(name)
    if (error) {
      setEditingError(error)
      return
    }
    if (name === entry.name) {
      setRenamingPath(undefined)
      return
    }
    setEditingBusy(true)
    const result = await runOperation(`Rename ${entry.name}`, { action: 'rename', path: entry.path, newName: name })
    const item = result?.items[0]
    if (item?.status === 'ok') {
      setRenamingPath(undefined)
      setEditingError(undefined)
      if (item.destinationPath)
        setSelection({ paths: [item.destinationPath], anchorPath: item.destinationPath })
      if (previewPath === entry.path)
        closePreview()
    }
    else if (item?.status === 'error') {
      setEditingError(item.error.message)
    }
    setEditingBusy(false)
  }

  const copySelection = (mode: 'copy' | 'move', entry?: DirectoryEntry): void => {
    const paths = selectedPathsFor(entry)
    if (managementEnabled && paths.length > 0)
      setClipboard({ mode, paths })
  }

  const pasteClipboard = async (destinationPath = directoryPath): Promise<void> => {
    if (!clipboard || operationBusy)
      return
    const result = await runOperation(clipboard.mode === 'copy' ? 'Copy items' : 'Move items', clipboardOperation(clipboard, destinationPath))
    if (!result)
      return
    setClipboard(settleClipboard(clipboard, result.items))
    const destinations = result.items.filter(item => item.status === 'ok' && item.destinationPath).map(item => item.destinationPath!)
    if (destinationPath === directoryPath && destinations.length > 0)
      setSelection({ paths: destinations, anchorPath: destinations.at(-1) })
  }

  const requestDelete = (entry?: DirectoryEntry): void => {
    const paths = selectedPathsFor(entry)
    if (managementEnabled && paths.length > 0) {
      setSelection({ paths, anchorPath: paths.at(-1) })
      setDeleteOpen(true)
    }
  }

  const confirmDelete = async (): Promise<void> => {
    const paths = selection.paths
    const result = await runOperation(`Delete ${paths.length} item${paths.length === 1 ? '' : 's'}`, { action: 'delete', paths })
    if (result) {
      const deleted = new Set(result.items.filter(item => item.status === 'ok').map(item => item.sourcePath))
      if (previewPath && deleted.has(previewPath))
        closePreview()
      setSelection(current => ({ paths: current.paths.filter(path => !deleted.has(path)) }))
      setDeleteOpen(false)
    }
  }

  const openEntry = useCallback((entry: DirectoryEntry): void => {
    if (entry.kind === 'directory')
      navigate(entry.path)
    else if (entry.kind === 'file')
      openPreview(entry)
  }, [navigate, openPreview])

  const selectFile = (entry: DirectoryEntry, modifiers: { toggle: boolean, range: boolean }): void => {
    setSelection(current => selectEntry(entries.map(item => item.path), current, entry.path, modifiers))
  }

  const enqueueUploads = useCallback(async (files: File[]) => {
    if (files.length === 0 || uploadActive || !listing?.managementEnabled)
      return
    const tasks = files.map(file => ({
      id: uuidv4(),
      file,
      activity: {
        id: uuidv4(),
        label: file.name,
        status: file.size > listing.maxUploadBytes ? 'error' as const : 'queued' as const,
        progress: 0,
        detail: file.size > listing.maxUploadBytes ? 'File exceeds the 1 GiB limit' : 'Waiting to upload',
      },
    }))
    setActivities(current => [...tasks.map(task => task.activity), ...current])
    setUploadActive(true)
    const pending = tasks.filter(task => task.activity.status === 'queued')
    let cursor = 0
    const worker = async (): Promise<void> => {
      while (cursor < pending.length) {
        const item = pending[cursor++]!
        updateActivity(item.activity.id, { status: 'running', progress: 0, detail: 'Uploading' })
        try {
          const result = await uploadFile(directoryPath, item.file, progress => updateActivity(item.activity.id, { progress }))
          updateActivity(item.activity.id, { status: 'done', progress: 100, detail: result.filename === item.file.name ? 'Uploaded' : `Saved as ${result.filename}` })
        }
        catch (cause) {
          updateActivity(item.activity.id, { status: 'error', detail: cause instanceof Error ? cause.message : String(cause) })
        }
      }
    }
    await Promise.all(Array.from({ length: Math.min(3, pending.length) }, () => worker()))
    setUploadActive(false)
    setReloadVersion(version => version + 1)
  }, [directoryPath, listing, uploadActive])

  useEffect(() => {
    if (!listing?.managementEnabled)
      return
    let dragDepth = 0
    let internalDrag = false
    const start = (): void => {
      internalDrag = true
      dragDepth = 0
      setDragging(false)
    }
    const end = (): void => {
      internalDrag = false
      dragDepth = 0
      setDragging(false)
    }
    const enter = (event: DragEvent): void => {
      if (internalDrag || !event.dataTransfer?.types.includes('Files'))
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
    const over = (event: DragEvent): void => {
      if (!internalDrag && event.dataTransfer?.types.includes('Files'))
        event.preventDefault()
    }
    const drop = (event: DragEvent): void => {
      if (internalDrag) {
        end()
        return
      }
      if (!event.dataTransfer?.types.includes('Files'))
        return
      event.preventDefault()
      dragDepth = 0
      setDragging(false)
      void enqueueUploads(Array.from(event.dataTransfer?.files ?? []))
    }
    document.addEventListener('dragstart', start)
    document.addEventListener('dragend', end)
    document.addEventListener('dragenter', enter)
    document.addEventListener('dragleave', leave)
    document.addEventListener('dragover', over)
    document.addEventListener('drop', drop)
    return () => {
      document.removeEventListener('dragstart', start)
      document.removeEventListener('dragend', end)
      document.removeEventListener('dragenter', enter)
      document.removeEventListener('dragleave', leave)
      document.removeEventListener('dragover', over)
      document.removeEventListener('drop', drop)
    }
  }, [enqueueUploads, listing?.managementEnabled])

  useEffect(() => {
    const keydown = (event: KeyboardEvent): void => {
      const target = event.target as HTMLElement | null
      if (target?.matches('input, textarea, select, [contenteditable="true"]'))
        return
      const modifier = event.metaKey || event.ctrlKey
      if (modifier && event.key.toLowerCase() === 'a') {
        event.preventDefault()
        const paths = entries.map(entry => entry.path)
        setSelection({ paths, anchorPath: paths.at(-1) })
      }
      else if (modifier && event.key.toLowerCase() === 'c' && managementEnabled) {
        event.preventDefault()
        copySelection('copy')
      }
      else if (modifier && event.key.toLowerCase() === 'x' && managementEnabled) {
        event.preventDefault()
        copySelection('move')
      }
      else if (modifier && event.key.toLowerCase() === 'v' && managementEnabled) {
        event.preventDefault()
        void pasteClipboard()
      }
      else if (event.key === 'F2' && selectedEntries.length === 1 && managementEnabled) {
        event.preventDefault()
        startRename(selectedEntries[0]!)
      }
      else if (event.key === 'Delete' && selection.paths.length > 0 && managementEnabled) {
        event.preventDefault()
        requestDelete()
      }
      else if (event.key === 'Enter' && selectedEntries.length === 1) {
        event.preventDefault()
        openEntry(selectedEntries[0]!)
      }
      else if (event.key === 'Escape') {
        if (creatingFolder || renamingPath) {
          setCreatingFolder(false)
          setRenamingPath(undefined)
          setEditingError(undefined)
        }
        else if (imageViewerOpen) {
          setImageViewerOpen(false)
        }
        else if (previewPath) {
          closePreview()
        }
        else {
          setSelection({ paths: [] })
        }
      }
    }
    window.addEventListener('keydown', keydown)
    return () => window.removeEventListener('keydown', keydown)
  }, [clipboard, closePreview, creatingFolder, entries, imageViewerOpen, managementEnabled, openEntry, previewPath, renamingPath, selectedEntries, selection.paths])

  const refresh = (): void => setReloadVersion(version => version + 1)
  const cancelEdit = (): void => {
    setCreatingFolder(false)
    setRenamingPath(undefined)
    setEditingError(undefined)
  }

  const fileArea = (
    <main className="flex h-full min-h-0 flex-1 flex-col bg-content">
      {loading && !listing && <CenteredState icon={<LoaderCircle className="size-5 animate-spin" />} label="Loading folder" />}
      {!loading && visibleError && !listing && <CenteredState icon={<FolderOpen className="size-6" />} label="Folder unavailable" />}
      {listing && (
        <FileBrowser
          entries={entries}
          viewMode={viewMode}
          sortKey={sortKey}
          direction={sortDirection}
          selectedPaths={selectedSet}
          managementEnabled={managementEnabled}
          canPaste={clipboard !== undefined}
          creatingFolder={creatingFolder}
          editingBusy={editingBusy}
          editingError={editingError}
          emptyLabel={entries.length === 0 && !creatingFolder && !loading ? (query ? 'No items match your search' : 'This folder is empty') : undefined}
          onSort={changeSort}
          onSelect={selectFile}
          onOpen={openEntry}
          onCreateFolder={name => void createFolder(name)}
          onCancelEdit={cancelEdit}
          onCut={entry => copySelection('move', entry)}
          onCopy={entry => copySelection('copy', entry)}
          onPaste={destination => void pasteClipboard(destination)}
          onStartRename={startRename}
          onDelete={requestDelete}
          onNewFolder={startCreateFolder}
          onRefresh={refresh}
        />
      )}
    </main>
  )

  const navigationPane = (
    <NavigationPane
      rootName={listing?.rootName ?? 'Root'}
      currentPath={directoryPath}
      previewPath={previewPath}
      revision={reloadVersion}
      onNavigate={navigate}
      onOpenFile={openTreeFile}
    />
  )

  return (
    <div className="explorer-shell">
      <div className="brand-line" />
      <header className="title-bar">
        <Button className="mobile-nav-trigger command-button" size="icon" variant="ghost" aria-label="Open folder navigation" onClick={() => setMobileNavigation(true)}><Menu className="size-4" /></Button>
        <span className="app-icon"><HardDrive className="size-4" /></span>
        <div className="min-w-0">
          <div className="brand-label">HACKYCY CLI · FILE SERVER</div>
          <div className="truncate text-xs font-semibold" title={listing?.rootName}>{listing?.rootName ?? 'File Server'}</div>
        </div>
        <span className={cn('mode-badge', managementEnabled && 'manage')}>{managementEnabled ? 'MANAGEMENT' : 'READ ONLY'}</span>
        <Tooltip label={`Use ${theme === 'light' ? 'dark' : 'light'} theme`}>
          <Button aria-label={`Use ${theme === 'light' ? 'dark' : 'light'} theme`} className="ml-auto command-button" size="icon" variant="ghost" onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
            {theme === 'light' ? <Moon className="size-4" /> : <Sun className="size-4" />}
          </Button>
        </Tooltip>
      </header>

      <div className="address-bar">
        <div className="navigation-buttons">
          <Tooltip label="Back"><Button className="command-button" size="icon" variant="ghost" aria-label="Back" disabled={navigation.cursor === 0} onClick={() => moveHistory(-1)}><ArrowLeft className="size-4" /></Button></Tooltip>
          <Tooltip label="Forward"><Button className="command-button" size="icon" variant="ghost" aria-label="Forward" disabled={navigation.cursor >= navigation.paths.length - 1} onClick={() => moveHistory(1)}><ArrowRight className="size-4" /></Button></Tooltip>
          <Tooltip label="Up"><Button className="command-button" size="icon" variant="ghost" aria-label="Up" disabled={!listing?.parentPath && directoryPath === ''} onClick={() => navigate(listing?.parentPath ?? '')}><ArrowUp className="size-4" /></Button></Tooltip>
        </div>
        <Breadcrumb path={listing?.path ?? directoryPath} rootName={listing?.rootName ?? 'Root'} onNavigate={navigate} />
        <label className="search-box">
          <Search className="size-3.5 shrink-0 text-muted-foreground" />
          <input
            aria-label="Search current folder"
            value={query}
            placeholder={`Search ${directoryPath.split('/').at(-1) || listing?.rootName || 'folder'}`}
            onChange={(event) => {
              setQuery(event.target.value)
              setSelection({ paths: [] })
            }}
          />
          {query && <button type="button" aria-label="Clear search" onClick={() => setQuery('')}><X className="size-3.5" /></button>}
        </label>
      </div>

      <div role="toolbar" aria-label="File commands" className="command-bar">
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
        {managementEnabled && <IconCommand icon={<FolderPlus />} label="New folder" disabled={operationBusy} onClick={startCreateFolder} />}
        {managementEnabled && <IconCommand icon={<Upload />} label="Upload" disabled={uploadActive} onClick={() => fileInputRef.current?.click()} />}
        {managementEnabled && <span className="command-separator desktop-command" />}
        {managementEnabled && (
          <div className="desktop-command flex items-center gap-0.5">
            <IconCommand icon={<Scissors />} label="Cut" disabled={selection.paths.length === 0} onClick={() => copySelection('move')} />
            <IconCommand icon={<Copy />} label="Copy" disabled={selection.paths.length === 0} onClick={() => copySelection('copy')} />
            <IconCommand icon={<ClipboardPaste />} label="Paste" disabled={!clipboard || operationBusy} onClick={() => void pasteClipboard()} />
            <IconCommand icon={<Pencil />} label="Rename" disabled={selectedEntries.length !== 1 || operationBusy} onClick={() => selectedEntries[0] && startRename(selectedEntries[0])} />
            <IconCommand icon={<Trash2 />} label="Delete" destructive disabled={selection.paths.length === 0 || operationBusy} onClick={() => requestDelete()} />
          </div>
        )}
        {managementEnabled && <span className="command-separator desktop-command" />}
        <div className="view-switch desktop-command">
          <IconCommand icon={<List />} label="Details view" pressed={viewMode === 'list'} onClick={() => setViewMode('list')} />
          <IconCommand icon={<Grid2X2 />} label="Grid view" pressed={viewMode === 'grid'} onClick={() => setViewMode('grid')} />
        </div>
        <IconCommand className="desktop-command" icon={<PanelRight />} label="Preview selected file" disabled={selectedEntries.length !== 1 || selectedEntries[0]?.kind !== 'file'} pressed={previewEntry !== undefined} onClick={() => selectedEntries[0] && openPreview(selectedEntries[0])} />
        <IconCommand className="desktop-command" icon={<RefreshCw className={loading ? 'animate-spin' : ''} />} label="Refresh" disabled={loading} onClick={refresh} />
        <MoreMenu
          mobile={mobile}
          managementEnabled={managementEnabled}
          hasSelection={selection.paths.length > 0}
          oneSelected={selectedEntries.length === 1}
          canPaste={clipboard !== undefined}
          sortKey={sortKey}
          direction={sortDirection}
          onCut={() => copySelection('move')}
          onCopy={() => copySelection('copy')}
          onPaste={() => void pasteClipboard()}
          onRename={() => selectedEntries[0] && startRename(selectedEntries[0])}
          onDelete={() => requestDelete()}
          onList={() => setViewMode('list')}
          onGrid={() => setViewMode('grid')}
          onPreview={() => selectedEntries[0] && openPreview(selectedEntries[0])}
          onRefresh={refresh}
          onSort={chooseSort}
          onDirection={setSortDirection}
        />
      </div>

      {visibleError && <div role="alert" className="error-banner">{visibleError}</div>}

      <div className="min-h-0 flex-1">
        {mobile
          ? fileArea
          : (
              <Group orientation="horizontal" className="h-full" onLayoutChanged={saveNavigationPanelWidth}>
                <Panel
                  id="navigation"
                  panelRef={navigationPanelRef}
                  defaultSize={initialNavigationPanelWidth}
                  minSize={NAVIGATION_PANEL_MIN_WIDTH}
                  maxSize={NAVIGATION_PANEL_MAX_WIDTH}
                  groupResizeBehavior="preserve-pixel-size"
                >
                  {navigationPane}
                </Panel>
                <Separator
                  className="resize-handle"
                  onPointerDown={() => activeResizeHandleRef.current = 'navigation'}
                  onKeyDown={() => activeResizeHandleRef.current = 'navigation'}
                >
                  <span />
                </Separator>
                <Panel id="files" minSize="360px">{fileArea}</Panel>
                {previewEntry && (
                  <>
                    <Separator
                      className="resize-handle"
                      onPointerDown={() => activeResizeHandleRef.current = 'preview'}
                      onKeyDown={() => activeResizeHandleRef.current = 'preview'}
                    >
                      <span />
                    </Separator>
                    <Panel id="preview" defaultSize="360px" minSize="300px" maxSize="48%">
                      <PreviewPane entry={previewEntry} theme={theme} imageViewerOpen={imageViewerOpen} onImageViewerOpenChange={setImageViewerOpen} onClose={closePreview} />
                    </Panel>
                  </>
                )}
              </Group>
            )}
      </div>

      <footer className="status-bar">
        <span>{`${listing?.entries.length ?? 0} ${(listing?.entries.length ?? 0) === 1 ? 'item' : 'items'}`}</span>
        {selection.paths.length > 0 && <span>{`${selection.paths.length} selected`}</span>}
        {query && <span>{`${entries.length} matches`}</span>}
        <span className="ml-auto truncate">{`/${listing?.path ?? directoryPath}`}</span>
      </footer>

      <Sheet open={mobileNavigation} onOpenChange={setMobileNavigation}>
        <SheetContent side="left" title="File navigation" description="Browse files and folders" className="mobile-sheet flex w-[min(88vw,340px)] flex-col pt-12">{navigationPane}</SheetContent>
      </Sheet>
      {mobile && <PreviewSheet entry={previewEntry} theme={theme} imageViewerOpen={imageViewerOpen} onImageViewerOpenChange={setImageViewerOpen} onClose={closePreview} />}
      <RenameDialog
        entry={renamingEntry}
        busy={editingBusy}
        serverError={editingError}
        onOpenChange={open => !open && cancelEdit()}
        onNameChange={() => setEditingError(undefined)}
        onConfirm={name => void renameEntry(name)}
      />
      <DeleteDialog paths={selection.paths} open={deleteOpen} busy={operationBusy} onOpenChange={setDeleteOpen} onConfirm={() => void confirmDelete()} />
      <ActivityCenter tasks={activities} onClear={() => setActivities([])} />
      {toast && (
        <div role="status" className="toast">
          <span>{toast}</span>
          <button type="button" aria-label="Dismiss" onClick={() => setToast(undefined)}><X className="size-3.5" /></button>
        </div>
      )}
      {dragging && (
        <div className="drop-overlay">
          <Upload className="size-6" />
          <span>{`Drop files into /${listing?.path ?? ''}`}</span>
        </div>
      )}
    </div>
  )
}

function Breadcrumb({ path, rootName, onNavigate }: { path: string, rootName: string, onNavigate: (path: string) => void }): React.JSX.Element {
  const segments = path.split('/').filter(Boolean)
  return (
    <nav aria-label="Breadcrumb" className="breadcrumb-box">
      <button type="button" title={rootName} onClick={() => onNavigate('')}>
        <HardDrive className="size-3.5 text-accent" />
        <span className="truncate">{rootName}</span>
      </button>
      {segments.map((segment, index) => {
        const segmentPath = segments.slice(0, index + 1).join('/')
        return (
          <span key={segmentPath} className="breadcrumb-segment">
            <ChevronRight className="size-3 text-muted-foreground" />
            <button type="button" title={segment} onClick={() => onNavigate(segmentPath)}>{segment}</button>
          </span>
        )
      })}
    </nav>
  )
}

function IconCommand({ icon, label, disabled, destructive = false, pressed, className, onClick }: { icon: React.ReactElement, label: string, disabled?: boolean, destructive?: boolean, pressed?: boolean, className?: string, onClick: () => void }): React.JSX.Element {
  return (
    <Tooltip label={label}>
      <Button aria-label={label} aria-pressed={pressed} className={cn('command-button', destructive && 'destructive', pressed && 'active', className)} size="icon" variant="ghost" disabled={disabled} onClick={onClick}>{icon}</Button>
    </Tooltip>
  )
}

function MoreMenu(props: {
  mobile: boolean
  managementEnabled: boolean
  hasSelection: boolean
  oneSelected: boolean
  canPaste: boolean
  sortKey: SortKey
  direction: SortDirection
  onCut: () => void
  onCopy: () => void
  onPaste: () => void
  onRename: () => void
  onDelete: () => void
  onList: () => void
  onGrid: () => void
  onPreview: () => void
  onRefresh: () => void
  onSort: (key: SortKey) => void
  onDirection: (direction: SortDirection) => void
}): React.JSX.Element {
  return (
    <DropdownMenu.Root modal={false}>
      <Tooltip label="More commands">
        <DropdownMenu.Trigger asChild><Button className="command-more command-button ml-auto" size="icon" variant="ghost" aria-label="More commands"><MoreHorizontal /></Button></DropdownMenu.Trigger>
      </Tooltip>
      <DropdownMenu.Portal>
        <DropdownMenu.Content className="menu-content" align="end" sideOffset={4}>
          {props.mobile && props.managementEnabled && (
            <DropdownMenu.Item className="menu-item" disabled={!props.hasSelection} onSelect={props.onCut}>
              <Scissors />
              <span>Cut</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && props.managementEnabled && (
            <DropdownMenu.Item className="menu-item" disabled={!props.hasSelection} onSelect={props.onCopy}>
              <Copy />
              <span>Copy</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && props.managementEnabled && (
            <DropdownMenu.Item className="menu-item" disabled={!props.canPaste} onSelect={props.onPaste}>
              <ClipboardPaste />
              <span>Paste</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && props.managementEnabled && (
            <DropdownMenu.Item className="menu-item" disabled={!props.oneSelected} onSelect={props.onRename}>
              <Pencil />
              <span>Rename</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && props.managementEnabled && (
            <DropdownMenu.Item className="menu-item destructive" disabled={!props.hasSelection} onSelect={props.onDelete}>
              <Trash2 />
              <span>Delete</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && props.managementEnabled && <DropdownMenu.Separator className="menu-separator" />}
          {props.mobile && (
            <DropdownMenu.Item className="menu-item" onSelect={props.onList}>
              <List />
              <span>Details view</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && (
            <DropdownMenu.Item className="menu-item" onSelect={props.onGrid}>
              <Grid2X2 />
              <span>Grid view</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && (
            <DropdownMenu.Item className="menu-item" disabled={!props.oneSelected} onSelect={props.onPreview}>
              <PanelRight />
              <span>Preview</span>
            </DropdownMenu.Item>
          )}
          {props.mobile && <DropdownMenu.Separator className="menu-separator" />}
          <SortSubmenu sortKey={props.sortKey} direction={props.direction} onSort={props.onSort} onDirection={props.onDirection} />
          {props.mobile && <DropdownMenu.Separator className="menu-separator" />}
          {props.mobile && (
            <DropdownMenu.Item className="menu-item" onSelect={props.onRefresh}>
              <RefreshCw />
              <span>Refresh</span>
            </DropdownMenu.Item>
          )}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  )
}

function SortSubmenu({ sortKey, direction, onSort, onDirection }: {
  sortKey: SortKey
  direction: SortDirection
  onSort: (key: SortKey) => void
  onDirection: (direction: SortDirection) => void
}): React.JSX.Element {
  return (
    <DropdownMenu.Sub>
      <DropdownMenu.SubTrigger className="menu-item">
        <ArrowUpDown />
        <span>Sort by</span>
        <ChevronRight className="menu-sub-chevron" />
      </DropdownMenu.SubTrigger>
      <DropdownMenu.Portal>
        <DropdownMenu.SubContent className="menu-content" sideOffset={6} alignOffset={-4}>
          <DropdownItem label="Name" checked={sortKey === 'name'} onSelect={() => onSort('name')} />
          <DropdownItem label="Date modified" checked={sortKey === 'modified'} onSelect={() => onSort('modified')} />
          <DropdownItem label="Size" checked={sortKey === 'size'} onSelect={() => onSort('size')} />
          <DropdownMenu.Separator className="menu-separator" />
          <DropdownItem label="Ascending" checked={direction === 'asc'} onSelect={() => onDirection('asc')} />
          <DropdownItem label="Descending" checked={direction === 'desc'} onSelect={() => onDirection('desc')} />
        </DropdownMenu.SubContent>
      </DropdownMenu.Portal>
    </DropdownMenu.Sub>
  )
}

function DropdownItem({ label, checked, onSelect }: { label: string, checked: boolean, onSelect: () => void }): React.JSX.Element {
  return (
    <DropdownMenu.Item className="menu-item" onSelect={onSelect}>
      <span className="menu-check">{checked ? '✓' : ''}</span>
      <span>{label}</span>
    </DropdownMenu.Item>
  )
}

function CenteredState({ icon, label }: { icon: React.ReactNode, label: string }): React.JSX.Element {
  return (
    <div className="centered-state">
      {icon}
      <span>{label}</span>
    </div>
  )
}
