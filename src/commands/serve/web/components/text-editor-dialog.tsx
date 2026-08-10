import type { Extension } from '@codemirror/state'
import type { DirectoryEntry, TextPreview, TextSaveResult } from '../api'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { css } from '@codemirror/lang-css'
import { html } from '@codemirror/lang-html'
import { javascript } from '@codemirror/lang-javascript'
import { json } from '@codemirror/lang-json'
import { markdown } from '@codemirror/lang-markdown'
import { python } from '@codemirror/lang-python'
import { yaml } from '@codemirror/lang-yaml'
import { bracketMatching, defaultHighlightStyle, indentOnInput, syntaxHighlighting } from '@codemirror/language'
import { EditorState } from '@codemirror/state'
import { drawSelection, EditorView, highlightActiveLine, highlightActiveLineGutter, keymap, lineNumbers } from '@codemirror/view'
import * as Dialog from '@radix-ui/react-dialog'
import { AlertTriangle, Download, LoaderCircle, Save, X } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { Button } from '../../../../shared/web/components/ui/button'
import { ApiError, apiJson } from '../api'
import { formatFileSize } from './file-browser'

export type ReadyTextPreview = Extract<TextPreview, { status: 'ready' }>

export interface TextEditorTarget {
  entry: DirectoryEntry
  preview: ReadyTextPreview
}

export function TextEditorDialog({ target, theme, onDirtyChange, onSaved, onClose }: {
  target?: TextEditorTarget
  theme: 'light' | 'dark'
  onDirtyChange: (dirty: boolean) => void
  onSaved: () => void
  onClose: () => void
}): React.JSX.Element {
  const [remote, setRemote] = useState<ReadyTextPreview>()
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const [conflict, setConflict] = useState<string>()
  const dirty = target !== undefined && remote !== undefined && draft !== remote.text

  useEffect(() => {
    if (!target) {
      setRemote(undefined)
      setDraft('')
      setConflict(undefined)
      setSaving(false)
      return
    }
    setRemote(target.preview)
    setDraft(target.preview.text)
    setConflict(undefined)
  }, [target?.entry.path, target?.preview.revision])

  useEffect(() => {
    onDirtyChange(Boolean(dirty))
    return () => onDirtyChange(false)
  }, [dirty, onDirtyChange])

  useEffect(() => {
    if (!dirty)
      return
    const beforeUnload = (event: BeforeUnloadEvent): void => {
      event.preventDefault()
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', beforeUnload)
    return () => window.removeEventListener('beforeunload', beforeUnload)
  }, [dirty])

  const close = (): void => {
    if (saving)
      return
    // eslint-disable-next-line no-alert
    if (dirty && !window.confirm('Discard unsaved changes?'))
      return
    onClose()
  }

  const save = async (): Promise<void> => {
    if (!target || !remote || saving || !dirty)
      return
    setSaving(true)
    setConflict(undefined)
    try {
      const result = await apiJson<TextSaveResult>(`/api/text?path=${encodeURIComponent(target.entry.path)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'text/plain; charset=utf-8', 'If-Match': remote.revision },
        body: draft,
      })
      setRemote({ version: 1, status: 'ready', text: draft, encoding: result.encoding, size: result.size, revision: result.revision })
      setConflict(undefined)
      onSaved()
    }
    catch (cause) {
      if (cause instanceof ApiError && cause.status === 412)
        setConflict('The file changed on the server. Your draft is still open.')
      else
        setConflict(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setSaving(false)
    }
  }

  const reloadRemote = async (): Promise<void> => {
    if (!target || saving)
      return
    try {
      const next = await apiJson<TextPreview>(`/api/text?path=${encodeURIComponent(target.entry.path)}`)
      if (next.status === 'ready') {
        setRemote(next)
        setDraft(next.text)
        setConflict(undefined)
      }
      else {
        setConflict(next.status === 'binary' ? 'The remote file is no longer text.' : 'The remote file is now too large to edit.')
      }
    }
    catch (cause) {
      setConflict(cause instanceof Error ? cause.message : String(cause))
    }
  }

  const downloadDraft = (): void => {
    if (!target)
      return
    const url = URL.createObjectURL(new Blob([draft], { type: 'text/plain;charset=utf-8' }))
    const link = document.createElement('a')
    link.href = url
    link.download = `${target.entry.name}.draft`
    link.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Dialog.Root open={target !== undefined} onOpenChange={open => !open && close()}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay text-editor-dialog-overlay" />
        {target && remote && (
          <Dialog.Content
            className="text-editor-dialog"
            onEscapeKeyDown={(event) => {
              event.preventDefault()
              close()
            }}
            onPointerDownOutside={(event) => {
              event.preventDefault()
              close()
            }}
            onOpenAutoFocus={event => event.preventDefault()}
          >
            <Dialog.Title className="sr-only">
              Edit
              {' '}
              {target.entry.name}
            </Dialog.Title>
            <Dialog.Description className="sr-only">Edit the text file and save changes</Dialog.Description>
            <header className="text-editor-dialog-header">
              <div className="min-w-0 flex-1">
                <h2 className="truncate text-sm font-semibold" title={target.entry.name}>{target.entry.name}</h2>
                <p className="truncate text-xs text-muted-foreground">{`${target.entry.mimeType ?? 'Unknown type'} · ${formatFileSize(remote.size)} · ${remote.encoding.toUpperCase()}`}</p>
              </div>
              <Button variant="ghost" size="icon" title="Save file" aria-label="Save file" disabled={saving || !dirty} onClick={() => void save()}>{saving ? <LoaderCircle className="size-4 animate-spin" /> : <Save className="size-4" />}</Button>
              <Button variant="ghost" size="icon" title="Close editor" aria-label="Close editor" disabled={saving} onClick={close}><X className="size-4" /></Button>
            </header>
            {conflict && (
              <div className="text-editor-conflict" role="alert">
                <AlertTriangle className="size-4 shrink-0" />
                <span>{conflict}</span>
                {conflict.startsWith('The file changed') && <Button variant="outline" size="default" onClick={() => void reloadRemote()}>Reload remote</Button>}
                {conflict.startsWith('The file changed') && (
                  <Button variant="outline" size="default" onClick={downloadDraft}>
                    <Download className="mr-1.5 size-3.5" />
                    Download draft
                  </Button>
                )}
              </div>
            )}
            <div className="text-editor-dialog-body">
              <CodeMirrorEditor value={draft} language={target.entry.syntaxLanguage} theme={theme} onChange={setDraft} onSave={() => void save()} />
            </div>
          </Dialog.Content>
        )}
      </Dialog.Portal>
    </Dialog.Root>
  )
}

function languageExtension(language: string | undefined): Extension | undefined {
  switch (language) {
    case 'javascript': return javascript()
    case 'typescript': return javascript({ typescript: true })
    case 'jsx': return javascript({ jsx: true })
    case 'tsx': return javascript({ jsx: true, typescript: true })
    case 'json': return json()
    case 'html': case 'xml': case 'svg': return html()
    case 'css': case 'scss': case 'less': return css()
    case 'markdown': return markdown()
    case 'python': return python()
    case 'yaml': return yaml()
    default: return undefined
  }
}

function CodeMirrorEditor({ value, language, theme, onChange, onSave }: { value: string, language?: string, theme: 'light' | 'dark', onChange: (value: string) => void, onSave: () => void }): React.JSX.Element {
  const hostRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const callbacks = useRef({ onChange, onSave })
  callbacks.current = { onChange, onSave }

  useEffect(() => {
    const parent = hostRef.current
    if (!parent)
      return
    const extensions: Extension[] = [
      lineNumbers(),
      highlightActiveLineGutter(),
      history(),
      drawSelection(),
      indentOnInput(),
      bracketMatching(),
      highlightActiveLine(),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab, {
        key: 'Mod-s',
        preventDefault: true,
        run: () => {
          callbacks.current.onSave()
          return true
        },
      }]),
      EditorView.lineWrapping,
      EditorView.updateListener.of(update => update.docChanged && callbacks.current.onChange(update.state.doc.toString())),
      EditorView.theme({
        '&': { height: '100%', fontSize: '13px', backgroundColor: 'var(--code)', color: 'var(--foreground)' },
        '.cm-scroller': { overflow: 'auto', fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace' },
        '.cm-gutters': { backgroundColor: 'var(--code)', color: 'var(--muted-foreground)', border: 'none' },
      }, { dark: theme === 'dark' }),
    ]
    const syntax = languageExtension(language)
    if (syntax)
      extensions.push(syntax)
    const view = new EditorView({ state: EditorState.create({ doc: value, extensions }), parent })
    viewRef.current = view
    view.focus()
    return () => {
      view.destroy()
      viewRef.current = null
    }
  }, [language, theme])

  useEffect(() => {
    const view = viewRef.current
    if (view && view.state.doc.toString() !== value)
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } })
  }, [value])

  return <div ref={hostRef} role="textbox" aria-label="Text editor" className="text-editor-surface" />
}
