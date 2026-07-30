import type { DirectoryEntry, TextPreview } from '../api'
import { Download, ExternalLink, FileQuestion, LoaderCircle } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Button } from '../../../../shared/web/components/ui/button'
import { ScrollArea } from '../../../../shared/web/components/ui/scroll-area'
import { Sheet, SheetContent } from '../../../../shared/web/components/ui/sheet'
import { apiJson } from '../api'
import { formatFileSize } from './file-browser'

export function PreviewSheet({ entry, onClose }: { entry?: DirectoryEntry, onClose: () => void }): React.JSX.Element {
  return (
    <Sheet open={entry !== undefined} onOpenChange={open => !open && onClose()}>
      {entry && (
        <SheetContent
          side="right"
          title={entry.name}
          description="File preview and download actions"
          closeLabel="Close preview"
          className="flex w-[min(96vw,760px)] flex-col"
        >
          <header className="flex min-h-14 shrink-0 items-center gap-3 border-b border-border px-4 pr-12">
            <div className="min-w-0 flex-1">
              <h2 className="truncate text-sm font-semibold" title={entry.name}>{entry.name}</h2>
              <p className="truncate text-xs text-muted-foreground">{`${entry.mimeType ?? 'Unknown type'} · ${formatFileSize(entry.size)}`}</p>
            </div>
            {entry.fileUrl && (
              <a href={entry.fileUrl} target="_blank" rel="noreferrer">
                <Button variant="ghost" size="icon" aria-label="Open in new tab"><ExternalLink className="size-4" /></Button>
              </a>
            )}
            {entry.downloadUrl && (
              <a href={entry.downloadUrl} download>
                <Button variant="ghost" size="icon" aria-label="Download file"><Download className="size-4" /></Button>
              </a>
            )}
          </header>
          <PreviewBody entry={entry} />
        </SheetContent>
      )}
    </Sheet>
  )
}

function PreviewBody({ entry }: { entry: DirectoryEntry }): React.JSX.Element {
  const [text, setText] = useState<TextPreview>()
  const [error, setError] = useState<string>()

  useEffect(() => {
    setText(undefined)
    setError(undefined)
    if (entry.previewKind !== 'text')
      return
    const controller = new AbortController()
    void apiJson<TextPreview>(`/api/text?path=${encodeURIComponent(entry.path)}`, { signal: controller.signal })
      .then(setText)
      .catch((cause) => {
        if (!controller.signal.aborted)
          setError(cause instanceof Error ? cause.message : String(cause))
      })
    return () => controller.abort()
  }, [entry])

  if (entry.previewKind === 'image' && entry.fileUrl) {
    return (
      <div className="image-preview-grid flex min-h-0 flex-1 items-center justify-center overflow-auto p-4">
        <img src={entry.fileUrl} alt={entry.name} className="max-h-full max-w-full object-contain" />
      </div>
    )
  }
  if (entry.previewKind === 'video' && entry.fileUrl) {
    return <div className="flex min-h-0 flex-1 items-center bg-black p-3"><video src={entry.fileUrl} controls className="max-h-full w-full" /></div>
  }
  if (entry.previewKind === 'audio' && entry.fileUrl) {
    return <div className="flex min-h-0 flex-1 items-center justify-center p-8"><audio src={entry.fileUrl} controls className="w-full max-w-lg" /></div>
  }
  if (entry.previewKind === 'pdf' && entry.fileUrl) {
    return <iframe src={entry.fileUrl} title={`Preview ${entry.name}`} className="min-h-0 flex-1 border-0 bg-white" />
  }
  if (entry.previewKind === 'text') {
    if (error)
      return <PreviewMessage title="Preview unavailable" detail={error} />
    if (!text)
      return <PreviewLoading />
    if (text.status === 'too_large')
      return <PreviewMessage title="Text preview is too large" detail={`${formatFileSize(text.size)} exceeds the ${formatFileSize(text.maxBytes)} preview limit.`} />
    if (text.status === 'binary')
      return <PreviewMessage title="This file is not valid supported text" detail="Open it in a new tab or download the original bytes." />
    return (
      <ScrollArea className="min-h-0 flex-1 bg-code" scrollbars="both">
        <pre className="min-w-max p-4 font-mono text-xs leading-5 text-foreground"><code>{text.text}</code></pre>
      </ScrollArea>
    )
  }
  return <PreviewMessage title="No inline preview" detail="Open the file in a new tab or download it to inspect the original content." />
}

function PreviewLoading(): React.JSX.Element {
  return (
    <div className="flex min-h-0 flex-1 items-center justify-center gap-2 text-sm text-muted-foreground">
      <LoaderCircle className="size-4 animate-spin" />
      <span>Loading preview</span>
    </div>
  )
}

function PreviewMessage({ title, detail }: { title: string, detail: string }): React.JSX.Element {
  return (
    <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 px-8 text-center">
      <FileQuestion className="size-8 text-muted-foreground" />
      <p className="text-sm font-medium">{title}</p>
      <p className="max-w-md text-xs leading-5 text-muted-foreground">{detail}</p>
    </div>
  )
}
