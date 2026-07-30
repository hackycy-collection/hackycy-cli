import type { DirectoryEntry } from '../api'
import type { SortDirection, SortKey, ViewMode } from '../types'
import { useVirtualizer } from '@tanstack/react-virtual'
import { ArrowDown, ArrowUp, File, FileText, Film, Folder, Image, Link as LinkIcon, Music, ShieldAlert } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { ScrollArea } from '../../../../shared/web/components/ui/scroll-area'
import { cn } from '../../../../shared/web/lib/utils'

export function formatFileSize(bytes: number | undefined): string {
  if (bytes === undefined)
    return '-'
  if (bytes === 0)
    return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const unit = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / 1024 ** unit
  return `${value.toFixed(unit === 0 ? 0 : value >= 10 ? 0 : 1)} ${units[unit]}`
}

function formatDate(value: string | undefined): string {
  if (!value)
    return '-'
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

function EntryGlyph({ entry, className }: { entry: DirectoryEntry, className?: string }): React.JSX.Element {
  const iconClass = cn('shrink-0', className)
  if (entry.kind === 'unavailable')
    return <ShieldAlert className={cn(iconClass, 'text-amber-600 dark:text-amber-400')} />
  if (entry.kind === 'directory')
    return <Folder className={cn(iconClass, 'text-cyan-600 dark:text-cyan-400')} />
  if (entry.previewKind === 'image')
    return <Image className={cn(iconClass, 'text-emerald-600 dark:text-emerald-400')} />
  if (entry.previewKind === 'video')
    return <Film className={cn(iconClass, 'text-rose-600 dark:text-rose-400')} />
  if (entry.previewKind === 'audio')
    return <Music className={cn(iconClass, 'text-violet-600 dark:text-violet-400')} />
  if (entry.previewKind === 'text')
    return <FileText className={cn(iconClass, 'text-zinc-600 dark:text-zinc-300')} />
  return <File className={cn(iconClass, 'text-zinc-500')} />
}

function EntryVisual({ entry, large = false }: { entry: DirectoryEntry, large?: boolean }): React.JSX.Element {
  if (entry.previewKind === 'image' && entry.fileUrl) {
    return (
      <span className={cn('image-thumbnail overflow-hidden border border-border bg-muted', large ? 'h-20 w-full rounded-md' : 'size-8 shrink-0 rounded')}>
        <img src={entry.fileUrl} alt="" loading="lazy" decoding="async" className="h-full w-full object-cover" />
      </span>
    )
  }
  return (
    <span className={cn('flex shrink-0 items-center justify-center bg-muted', large ? 'h-20 w-full rounded-md' : 'size-8 rounded')}>
      <EntryGlyph entry={entry} className={large ? 'size-8' : 'size-4'} />
    </span>
  )
}

function SortHeading({
  label,
  sortKey,
  activeKey,
  direction,
  className,
  onSort,
}: {
  label: string
  sortKey: SortKey
  activeKey: SortKey
  direction: SortDirection
  className?: string
  onSort: (key: SortKey) => void
}): React.JSX.Element {
  const active = activeKey === sortKey
  return (
    <button type="button" className={cn('flex h-full items-center gap-1 text-left hover:text-foreground', className)} onClick={() => onSort(sortKey)}>
      <span>{label}</span>
      {active && (direction === 'asc' ? <ArrowUp className="size-3" /> : <ArrowDown className="size-3" />)}
    </button>
  )
}

function EntryButton({
  entry,
  className,
  children,
  onOpenDirectory,
  onOpenFile,
  style,
}: {
  entry: DirectoryEntry
  className?: string
  children: React.ReactNode
  onOpenDirectory: (entry: DirectoryEntry) => void
  onOpenFile: (entry: DirectoryEntry) => void
  style?: React.CSSProperties
}): React.JSX.Element {
  const unavailable = entry.kind === 'unavailable'
  return (
    <button
      type="button"
      disabled={unavailable}
      title={unavailable ? 'This entry is outside the served directory or cannot be read' : entry.name}
      className={className}
      style={style}
      onClick={() => entry.kind === 'directory' ? onOpenDirectory(entry) : onOpenFile(entry)}
    >
      {children}
    </button>
  )
}

function ListView({ entries, sortKey, direction, onSort, onOpenDirectory, onOpenFile }: FileBrowserProps): React.JSX.Element {
  const viewportRef = useRef<HTMLDivElement>(null)
  const virtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => viewportRef.current,
    estimateSize: () => 49,
    overscan: 12,
  })

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="grid h-9 shrink-0 grid-cols-[minmax(0,1fr)_84px] border-b border-border px-3 text-[11px] font-semibold text-muted-foreground sm:grid-cols-[minmax(0,1fr)_100px_180px]">
        <SortHeading label="Name" sortKey="name" activeKey={sortKey} direction={direction} onSort={onSort} />
        <SortHeading label="Size" sortKey="size" activeKey={sortKey} direction={direction} className="justify-end" onSort={onSort} />
        <SortHeading label="Modified" sortKey="modified" activeKey={sortKey} direction={direction} className="hidden justify-end sm:flex" onSort={onSort} />
      </div>
      <ScrollArea className="min-h-0 flex-1" viewportRef={viewportRef}>
        <div className="relative w-full" style={{ height: `${virtualizer.getTotalSize()}px` }}>
          {virtualizer.getVirtualItems().map((row) => {
            const entry = entries[row.index]!
            return (
              <EntryButton
                key={entry.path}
                entry={entry}
                onOpenDirectory={onOpenDirectory}
                onOpenFile={onOpenFile}
                className="absolute left-0 grid w-full grid-cols-[minmax(0,1fr)_84px] items-center border-b border-border/65 px-3 text-left hover:bg-muted/70 disabled:cursor-not-allowed disabled:opacity-60 sm:grid-cols-[minmax(0,1fr)_100px_180px]"
                style={{ height: `${row.size}px`, transform: `translateY(${row.start}px)` }}
              >
                <span className="flex min-w-0 items-center gap-3">
                  <EntryVisual entry={entry} />
                  <span className="min-w-0">
                    <span className="flex min-w-0 items-center gap-1.5">
                      <span className="truncate text-sm font-medium">{entry.name}</span>
                      {entry.isSymlink && <LinkIcon aria-label="Symbolic link" className="size-3 shrink-0 text-muted-foreground" />}
                    </span>
                    <span className="block truncate text-[11px] text-muted-foreground sm:hidden">{formatDate(entry.modifiedAt)}</span>
                  </span>
                </span>
                <span className="text-right text-xs tabular-nums text-muted-foreground">{formatFileSize(entry.size)}</span>
                <span className="hidden text-right text-xs tabular-nums text-muted-foreground sm:block">{formatDate(entry.modifiedAt)}</span>
              </EntryButton>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}

function GridView({ entries, onOpenDirectory, onOpenFile }: FileBrowserProps): React.JSX.Element {
  const viewportRef = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(640)
  useEffect(() => {
    const element = viewportRef.current
    if (!element)
      return
    const observer = new ResizeObserver(([entry]) => setWidth(entry?.contentRect.width ?? 640))
    observer.observe(element)
    return () => observer.disconnect()
  }, [])
  const columns = width >= 1320 ? 7 : width >= 1080 ? 6 : width >= 820 ? 5 : width >= 600 ? 4 : width >= 410 ? 3 : 2
  const rowCount = Math.ceil(entries.length / columns)
  const virtualizer = useVirtualizer({
    count: rowCount,
    getScrollElement: () => viewportRef.current,
    estimateSize: () => 150,
    overscan: 5,
  })

  return (
    <ScrollArea className="h-full" viewportRef={viewportRef}>
      <div className="relative w-full px-3" style={{ height: `${virtualizer.getTotalSize() + 16}px` }}>
        {virtualizer.getVirtualItems().map((row) => {
          const rowEntries = entries.slice(row.index * columns, (row.index + 1) * columns)
          return (
            <div
              key={row.key}
              className="absolute left-3 right-3 grid gap-2"
              style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`, transform: `translateY(${row.start + 8}px)` }}
            >
              {rowEntries.map(entry => (
                <EntryButton
                  key={entry.path}
                  entry={entry}
                  onOpenDirectory={onOpenDirectory}
                  onOpenFile={onOpenFile}
                  className="grid h-[140px] min-w-0 grid-rows-[80px_1fr] gap-2 rounded-md border border-border bg-background p-2 text-left hover:border-cyan-500 hover:bg-muted/50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <EntryVisual entry={entry} large />
                  <span className="min-w-0 self-start">
                    <span className="flex min-w-0 items-center gap-1">
                      <span className="truncate text-xs font-medium">{entry.name}</span>
                      {entry.isSymlink && <LinkIcon aria-label="Symbolic link" className="size-3 shrink-0 text-muted-foreground" />}
                    </span>
                    <span className="mt-0.5 block text-[10px] text-muted-foreground">{entry.kind === 'directory' ? 'Folder' : formatFileSize(entry.size)}</span>
                  </span>
                </EntryButton>
              ))}
            </div>
          )
        })}
      </div>
    </ScrollArea>
  )
}

interface FileBrowserProps {
  entries: DirectoryEntry[]
  viewMode?: ViewMode
  sortKey: SortKey
  direction: SortDirection
  onSort: (key: SortKey) => void
  onOpenDirectory: (entry: DirectoryEntry) => void
  onOpenFile: (entry: DirectoryEntry) => void
}

export function FileBrowser(props: FileBrowserProps): React.JSX.Element {
  return props.viewMode === 'grid' ? <GridView {...props} /> : <ListView {...props} />
}
