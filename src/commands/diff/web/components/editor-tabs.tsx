import type { Entry } from '../api'
import { FileCode2, FileWarning, Link as LinkIcon, X } from 'lucide-react'
import { cn } from '../lib/utils'

const statusClass = {
  added: 'bg-green-500',
  deleted: 'bg-red-500',
  modified: 'bg-amber-500',
  unchanged: 'bg-zinc-400',
  issue: 'bg-cyan-500',
}

const statusLabel = {
  added: 'Added',
  deleted: 'Deleted',
  modified: 'Modified',
  unchanged: 'Unchanged',
  issue: 'Issue',
}

function fileName(filePath: string): string {
  return filePath.split('/').at(-1) ?? filePath
}

function tabLabel(entry: Entry, tabs: Entry[]): string {
  const name = fileName(entry.path)
  return tabs.some(other => other.id !== entry.id && fileName(other.path) === name) ? entry.path : name
}

export function EditorTabs({
  tabs,
  activeId,
  onSelect,
  onClose,
}: {
  tabs: Entry[]
  activeId?: number
  onSelect: (entry: Entry) => void
  onClose: (entry: Entry) => void
}): React.JSX.Element {
  return (
    <div role="tablist" aria-label="Open files" className="flex h-10 shrink-0 items-stretch overflow-x-auto border-b border-border bg-muted/30">
      {tabs.length === 0 && <div className="flex items-center px-3 text-xs text-muted-foreground">No open files</div>}
      {tabs.map((entry) => {
        const active = entry.id === activeId
        const Icon = entry.status === 'issue' ? FileWarning : entry.kind === 'symlink' ? LinkIcon : FileCode2
        return (
          <div
            key={entry.id}
            className={cn(
              'group flex min-w-36 max-w-56 shrink-0 items-center border-r border-t-2 border-border text-xs',
              active ? 'border-t-cyan-600 bg-background text-foreground' : 'border-t-transparent text-muted-foreground hover:bg-muted',
            )}
          >
            <button
              type="button"
              role="tab"
              aria-selected={active}
              className="flex h-full min-w-0 flex-1 items-center gap-1.5 px-3 text-left outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-cyan-500"
              aria-label={`Open ${entry.path}`}
              onClick={() => onSelect(entry)}
              title={entry.path}
            >
              <Icon className="size-3.5 shrink-0" />
              <span className="min-w-0 truncate font-mono">{tabLabel(entry, tabs)}</span>
              <span className={cn('size-1.5 shrink-0 rounded-full', statusClass[entry.status])} title={statusLabel[entry.status]} />
            </button>
            <button
              type="button"
              className="mr-1 flex size-6 shrink-0 items-center justify-center rounded-sm opacity-60 outline-none hover:bg-muted hover:opacity-100 focus-visible:ring-2 focus-visible:ring-cyan-500"
              aria-label={`Close ${entry.path}`}
              onClick={() => onClose(entry)}
            >
              <X className="size-3.5" />
            </button>
          </div>
        )
      })}
    </div>
  )
}
