import type { DownloadTask } from '../api'
import type { ActivityTask } from '../types'
import { Check, CircleStop, LoaderCircle, RotateCcw, TriangleAlert, X } from 'lucide-react'
import { Button } from '../../../../shared/web/components/ui/button'
import { ScrollArea } from '../../../../shared/web/components/ui/scroll-area'

export function formatBytes(bytes: number): string {
  if (bytes < 1024)
    return `${bytes} B`
  const units = ['KiB', 'MiB', 'GiB', 'TiB']
  let value = bytes / 1024
  let unit = units[0]!
  for (let index = 1; index < units.length && value >= 1024; index++) {
    value /= 1024
    unit = units[index]!
  }
  return `${value.toFixed(value >= 10 ? 0 : 1)} ${unit}`
}

export function downloadDetail(task: DownloadTask): string {
  if (task.error)
    return task.error
  if (task.status === 'queued')
    return 'Waiting to download'
  if (task.status === 'cancelled')
    return 'Cancelled'
  const size = task.totalBytes === undefined
    ? formatBytes(task.bytesDownloaded)
    : `${formatBytes(task.bytesDownloaded)} / ${formatBytes(task.totalBytes)}`
  if (task.status === 'done')
    return `${size} · Complete`
  return `${size}${task.speedBytesPerSecond ? ` · ${formatBytes(task.speedBytesPerSecond)}/s` : ''}`
}

export function ActivityCenter({
  tasks,
  downloads,
  onClear,
  onCancelDownload,
  onRetryDownload,
  onClearDownloads,
}: {
  tasks: ActivityTask[]
  downloads: DownloadTask[]
  onClear: () => void
  onCancelDownload: (id: string) => void
  onRetryDownload: (id: string) => void
  onClearDownloads: () => void
}): React.JSX.Element | null {
  if (tasks.length === 0 && downloads.length === 0)
    return null
  const pending = tasks.some(task => task.status === 'queued' || task.status === 'running') || downloads.some(task => task.status === 'queued' || task.status === 'running')
  const complete = tasks.filter(task => task.status === 'done').length + downloads.filter(task => task.status === 'done').length
  const allTasks: Array<{ kind: 'download', task: DownloadTask } | { kind: 'local', task: ActivityTask }> = [
    ...downloads.map(task => ({ kind: 'download' as const, task })),
    ...tasks.map(task => ({ kind: 'local' as const, task })),
  ]

  return (
    <aside className="activity-center" aria-label="Activity center">
      <header className="activity-header">
        <span className="text-xs font-semibold">Activity</span>
        <span className="text-[11px] text-muted-foreground">{pending ? 'Working' : `${complete}/${allTasks.length} complete`}</span>
        {!pending && (
          <Button
            className="ml-auto size-7"
            size="icon"
            variant="ghost"
            aria-label="Clear activity"
            onClick={() => {
              onClear()
              onClearDownloads()
            }}
          >
            <X className="size-3.5" />
          </Button>
        )}
      </header>
      <ScrollArea className="max-h-72">
        <div className="divide-y divide-border/70">
          {allTasks.map(({ kind, task }) => {
            const download = kind === 'download' ? task : undefined
            const local = kind === 'local' ? task : undefined
            const status = task.status
            const label = download ? (download.filename ?? download.url) : local!.label
            const detail = download ? downloadDetail(download) : local!.detail ?? local!.status
            const progress = download ? download.progress : local!.progress
            const indeterminate = download !== undefined && status === 'running' && progress === undefined
            return (
              <div key={`${kind}:${task.id}`} className="activity-row">
                {status === 'done' && <Check className="size-4 text-emerald-600" />}
                {(status === 'error' || status === 'cancelled') && <TriangleAlert className="size-4 text-red-600" />}
                {status === 'queued' && <span className="size-2 justify-self-center rounded-full bg-zinc-400" />}
                {status === 'running' && <LoaderCircle className="size-4 animate-spin text-accent" />}
                <span className="min-w-0">
                  <span className="block truncate text-xs font-medium" title={download?.url}>{label}</span>
                  <span className="block truncate text-[10px] text-muted-foreground">{detail}</span>
                </span>
                <span className="text-right text-[10px] tabular-nums text-muted-foreground">
                  {progress === undefined ? '' : `${progress}%`}
                </span>
                {download && (status === 'running' || status === 'queued') && (
                  <span className="activity-actions">
                    <Button className="size-6" size="icon" variant="ghost" aria-label="Cancel download" onClick={() => onCancelDownload(download.id)}><CircleStop className="size-3.5" /></Button>
                  </span>
                )}
                {download && (status === 'error' || status === 'cancelled') && (
                  <span className="activity-actions">
                    <Button className="size-6" size="icon" variant="ghost" aria-label="Retry download" onClick={() => onRetryDownload(download.id)}><RotateCcw className="size-3.5" /></Button>
                  </span>
                )}
                {(progress !== undefined || indeterminate) && (
                  <span className={`activity-progress${indeterminate ? ' indeterminate' : ''}`}>
                    <span className={status === 'error' || status === 'cancelled' ? 'bg-red-500' : 'bg-accent'} style={progress === undefined ? undefined : { width: `${progress}%` }} />
                  </span>
                )}
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </aside>
  )
}
