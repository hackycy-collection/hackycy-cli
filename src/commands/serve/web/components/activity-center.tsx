import type { ActivityTask } from '../types'
import { Check, LoaderCircle, TriangleAlert, X } from 'lucide-react'
import { Button } from '../../../../shared/web/components/ui/button'
import { ScrollArea } from '../../../../shared/web/components/ui/scroll-area'

export function ActivityCenter({ tasks, onClear }: { tasks: ActivityTask[], onClear: () => void }): React.JSX.Element | null {
  if (tasks.length === 0)
    return null
  const pending = tasks.some(task => task.status === 'queued' || task.status === 'running')
  const complete = tasks.filter(task => task.status === 'done').length

  return (
    <aside className="activity-center" aria-label="Activity center">
      <header className="activity-header">
        <span className="text-xs font-semibold">Activity</span>
        <span className="text-[11px] text-muted-foreground">{pending ? 'Working' : `${complete}/${tasks.length} complete`}</span>
        {!pending && <Button className="ml-auto size-7" size="icon" variant="ghost" aria-label="Clear activity" onClick={onClear}><X className="size-3.5" /></Button>}
      </header>
      <ScrollArea className="max-h-72">
        <div className="divide-y divide-border/70">
          {tasks.map(task => (
            <div key={task.id} className="activity-row">
              {task.status === 'done' && <Check className="size-4 text-emerald-600" />}
              {task.status === 'error' && <TriangleAlert className="size-4 text-red-600" />}
              {task.status === 'queued' && <span className="size-2 justify-self-center rounded-full bg-zinc-400" />}
              {task.status === 'running' && <LoaderCircle className="size-4 animate-spin text-accent" />}
              <span className="min-w-0">
                <span className="block truncate text-xs font-medium">{task.label}</span>
                <span className="block truncate text-[10px] text-muted-foreground">{task.detail ?? task.status}</span>
              </span>
              <span className="text-right text-[10px] tabular-nums text-muted-foreground">
                {task.progress === undefined ? '' : `${task.progress}%`}
              </span>
              {task.progress !== undefined && (
                <span className="activity-progress">
                  <span className={task.status === 'error' ? 'bg-red-500' : 'bg-accent'} style={{ width: `${task.progress}%` }} />
                </span>
              )}
            </div>
          ))}
        </div>
      </ScrollArea>
    </aside>
  )
}
