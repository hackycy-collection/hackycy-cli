import type { OverlayRenderProps } from 'react-photo-view/dist/types'
import type { DirectoryEntry, TextPreview } from '../api'
import { File as CodeFile } from '@pierre/diffs/react'
import { Download, ExternalLink, FileQuestion, LoaderCircle, Maximize2, RefreshCcw, RotateCcw, RotateCw, X, ZoomIn, ZoomOut } from 'lucide-react'
import { useEffect, useState } from 'react'
import { PhotoSlider } from 'react-photo-view'
import { Button } from '../../../../shared/web/components/ui/button'
import { Sheet, SheetContent } from '../../../../shared/web/components/ui/sheet'
import { apiJson } from '../api'
import { formatFileSize } from './file-browser'

interface PreviewProps {
  entry: DirectoryEntry
  theme: 'light' | 'dark'
  imageViewerOpen: boolean
  onImageViewerOpenChange: (open: boolean) => void
  onClose: () => void
}

export function PreviewPane({ entry, theme, imageViewerOpen, onImageViewerOpenChange, onClose }: PreviewProps): React.JSX.Element {
  return (
    <aside className="preview-pane" aria-label={`Preview ${entry.name}`}>
      <PreviewHeader entry={entry} onOpenImage={() => onImageViewerOpenChange(true)} onClose={onClose} />
      <PreviewBody entry={entry} theme={theme} imageViewerOpen={imageViewerOpen} onImageViewerOpenChange={onImageViewerOpenChange} />
    </aside>
  )
}

export function PreviewSheet({ entry, theme, imageViewerOpen, onImageViewerOpenChange, onClose }: Omit<PreviewProps, 'entry'> & { entry?: DirectoryEntry }): React.JSX.Element {
  return (
    <Sheet open={entry !== undefined} modal={!imageViewerOpen} onOpenChange={open => !open && !imageViewerOpen && onClose()}>
      {entry && (
        <SheetContent
          side="right"
          title={entry.name}
          description="File preview and download actions"
          closeLabel="Close preview"
          className="mobile-sheet preview-sheet flex w-[min(96vw,760px)] flex-col"
          onEscapeKeyDown={event => imageViewerOpen && event.preventDefault()}
        >
          <PreviewHeader entry={entry} onOpenImage={() => onImageViewerOpenChange(true)} onClose={onClose} hideClose />
          <PreviewBody entry={entry} theme={theme} imageViewerOpen={imageViewerOpen} onImageViewerOpenChange={onImageViewerOpenChange} />
        </SheetContent>
      )}
    </Sheet>
  )
}

function PreviewHeader({ entry, onOpenImage, onClose, hideClose = false }: { entry: DirectoryEntry, onOpenImage: () => void, onClose: () => void, hideClose?: boolean }): React.JSX.Element {
  return (
    <header className="preview-header">
      <div className="min-w-0 flex-1">
        <h2 className="truncate text-sm font-semibold" title={entry.name}>{entry.name}</h2>
        <p className="truncate text-xs text-muted-foreground">{`${entry.mimeType ?? 'Unknown type'} · ${formatFileSize(entry.size)}`}</p>
      </div>
      {entry.previewKind === 'image' && entry.fileUrl && (
        <Button variant="ghost" size="icon" title="Open full screen" aria-label="Open full screen image preview" onClick={onOpenImage}><Maximize2 className="size-4" /></Button>
      )}
      {entry.fileUrl && (
        <a href={entry.fileUrl} target="_blank" rel="noreferrer" aria-label="Open in new tab">
          <Button variant="ghost" size="icon"><ExternalLink className="size-4" /></Button>
        </a>
      )}
      {entry.downloadUrl && (
        <a href={entry.downloadUrl} download aria-label="Download file">
          <Button variant="ghost" size="icon"><Download className="size-4" /></Button>
        </a>
      )}
      {!hideClose && <Button variant="ghost" size="icon" aria-label="Close preview" onClick={onClose}><X className="size-4" /></Button>}
    </header>
  )
}

function PreviewBody({ entry, theme, imageViewerOpen, onImageViewerOpenChange }: Omit<PreviewProps, 'onClose'>): React.JSX.Element {
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
      <>
        <div className="image-preview-grid flex min-h-0 flex-1 items-center justify-center overflow-auto p-4">
          <button type="button" className="image-preview-trigger" title="Open full screen" aria-label={`Open full screen preview of ${entry.name}`} onClick={() => onImageViewerOpenChange(true)}>
            <img src={entry.fileUrl} alt={entry.name} draggable={false} className="max-h-full max-w-full object-contain" />
          </button>
        </div>
        <PhotoSlider
          images={[{ key: entry.path, src: entry.fileUrl }]}
          visible={imageViewerOpen}
          loop={false}
          onClose={() => onImageViewerOpenChange(false)}
          toolbarRender={props => <ImageViewerToolbar {...props} />}
        />
      </>
    )
  }
  if (entry.previewKind === 'video' && entry.fileUrl)
    return <div className="flex min-h-0 flex-1 items-center bg-black p-3"><video src={entry.fileUrl} controls className="max-h-full w-full" /></div>
  if (entry.previewKind === 'audio' && entry.fileUrl)
    return <div className="flex min-h-0 flex-1 items-center justify-center p-8"><audio src={entry.fileUrl} controls className="w-full max-w-lg" /></div>
  if (entry.previewKind === 'pdf' && entry.fileUrl)
    return <iframe src={entry.fileUrl} title={`Preview ${entry.name}`} className="min-h-0 flex-1 border-0 bg-white" />
  if (entry.previewKind === 'text') {
    if (error)
      return <PreviewMessage title="Preview unavailable" detail={error} />
    if (!text)
      return <PreviewLoading />
    if (text.status === 'too_large')
      return <PreviewMessage title="Text preview is too large" detail={`${formatFileSize(text.size)} exceeds the ${formatFileSize(text.maxBytes)} preview limit.`} />
    if (text.status === 'binary')
      return <PreviewMessage title="This file is not supported text" detail="Open it in a new tab or download the original bytes." />
    if (entry.syntaxLanguage) {
      return (
        <div role="region" aria-label={`Code preview of ${entry.name}`} tabIndex={0} className="text-preview-scroll code-preview-scroll min-h-0 flex-1 bg-code">
          <CodeFile
            file={{
              name: entry.name,
              contents: text.text,
              lang: entry.syntaxLanguage,
              cacheKey: `${entry.path}:${entry.modifiedAt ?? ''}:${entry.size ?? text.size}`,
            }}
            options={{ disableFileHeader: true, overflow: 'scroll', themeType: theme }}
          />
        </div>
      )
    }
    return (
      <div role="region" aria-label={`Text preview of ${entry.name}`} tabIndex={0} className="text-preview-scroll min-h-0 flex-1 bg-code">
        <pre className="min-w-max p-4 font-mono text-xs leading-5 text-foreground"><code>{text.text}</code></pre>
      </div>
    )
  }
  return <PreviewMessage title="No inline preview" detail="Open the file in a new tab or download it to inspect the original content." />
}

function ImageViewerToolbar({ scale, rotate, onScale, onRotate, onClose }: OverlayRenderProps): React.JSX.Element {
  const tools = [
    { label: 'Zoom out', icon: <ZoomOut />, action: () => onScale(scale - 0.5) },
    { label: 'Zoom in', icon: <ZoomIn />, action: () => onScale(scale + 0.5) },
    { label: 'Rotate left', icon: <RotateCcw />, action: () => onRotate(rotate - 90) },
    { label: 'Rotate right', icon: <RotateCw />, action: () => onRotate(rotate + 90) },
    {
      label: 'Reset image',
      icon: <RefreshCcw />,
      action: () => {
        onScale(1)
        onRotate(0)
      },
    },
    { label: 'Close image preview', icon: <X />, action: () => onClose() },
  ]
  return (
    <div role="toolbar" aria-label="Image preview controls" className="image-viewer-toolbar">
      {tools.map(tool => (
        <button key={tool.label} type="button" className="image-viewer-tool" title={tool.label} aria-label={tool.label} onClick={tool.action}>{tool.icon}</button>
      ))}
    </div>
  )
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
