import type { UploadTask } from '../types'
import { Check, LoaderCircle, TriangleAlert, X } from 'lucide-react'
import { Button } from '../../../../shared/web/components/ui/button'
import { ScrollArea } from '../../../../shared/web/components/ui/scroll-area'

export function UploadQueue({ tasks, active, onClear }: { tasks: UploadTask[], active: boolean, onClear: () => void }): React.JSX.Element | null {
  if (tasks.length === 0)
    return null
  const completed = tasks.filter(task => task.status === 'done').length
  return (
    <aside className="fixed bottom-3 right-3 z-30 flex w-[min(calc(100vw-24px),360px)] flex-col overflow-hidden rounded-md border border-border bg-background shadow-xl">
      <header className="flex h-10 shrink-0 items-center gap-2 border-b border-border px-3">
        <span className="text-xs font-semibold">Uploads</span>
        <span className="text-[11px] text-muted-foreground">{`${completed}/${tasks.length} complete`}</span>
        {!active && <Button className="ml-auto size-7" size="icon" variant="ghost" aria-label="Clear uploads" onClick={onClear}><X className="size-3.5" /></Button>}
      </header>
      <ScrollArea className="max-h-64">
        <div className="divide-y divide-border/70">
          {tasks.map(task => (
            <div key={task.id} className="grid grid-cols-[20px_minmax(0,1fr)_40px] items-center gap-2 px-3 py-2">
              {task.status === 'done' && <Check className="size-4 text-emerald-600" />}
              {task.status === 'error' && <TriangleAlert className="size-4 text-red-600" />}
              {task.status === 'queued' && <span className="size-2 justify-self-center rounded-full bg-zinc-400" />}
              {task.status === 'uploading' && <LoaderCircle className="size-4 animate-spin text-cyan-600" />}
              <span className="min-w-0">
                <span className="block truncate text-xs font-medium">{task.filename}</span>
                <span className="block truncate text-[10px] text-muted-foreground">{task.detail ?? task.status}</span>
              </span>
              <span className="text-right text-[10px] tabular-nums text-muted-foreground">{task.status === 'done' ? '100%' : task.status === 'error' ? '-' : `${task.progress}%`}</span>
              <span className="col-start-2 col-end-4 h-1 overflow-hidden rounded-full bg-muted">
                <span className={`block h-full ${task.status === 'error' ? 'bg-red-500' : 'bg-cyan-500'}`} style={{ width: `${task.progress}%` }} />
              </span>
            </div>
          ))}
        </div>
      </ScrollArea>
    </aside>
  )
}
