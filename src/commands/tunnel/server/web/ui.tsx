import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { Check, CheckCircle2, Clipboard, LoaderCircle, Plus, RefreshCw, Trash2, X, XCircle } from 'lucide-react'
import { createContext, useCallback, useContext, useRef, useState } from 'react'

interface Notification {
  id: number
  kind: 'success' | 'error'
  message: string
}

interface Feedback {
  notify: (message: string, kind?: Notification['kind']) => void
}

const FeedbackContext = createContext<Feedback | null>(null)

export function FeedbackProvider({ children }: { children: ReactNode }): React.JSX.Element {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const dismiss = useCallback((id: number): void => setNotifications(values => values.filter(value => value.id !== id)), [])
  const notify = useCallback((message: string, kind: Notification['kind'] = 'success'): void => {
    const id = Date.now() + Math.random()
    setNotifications(values => [...values.slice(-3), { id, kind, message }])
    setTimeout(() => dismiss(id), 3600)
  }, [dismiss])
  return (
    <FeedbackContext value={{ notify }}>
      {children}
      <div className="notification-region" aria-live="polite" aria-atomic="false">
        {notifications.map(notification => (
          <div className={`notification notification-${notification.kind}`} key={notification.id} role={notification.kind === 'error' ? 'alert' : 'status'}>
            {notification.kind === 'success' ? <CheckCircle2 size={16} /> : <XCircle size={16} />}
            <span>{notification.message}</span>
            <IconButton label="Dismiss notification" onClick={() => dismiss(notification.id)}><X size={14} /></IconButton>
          </div>
        ))}
      </div>
    </FeedbackContext>
  )
}

export function useFeedback(): Feedback {
  const value = useContext(FeedbackContext)
  if (!value)
    throw new Error('FeedbackProvider is required')
  return value
}

export function navigate(path: string): void {
  history.pushState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

function statusClass(value: string): string {
  if (['connected', 'running', 'Applied'].includes(value))
    return 'status status-good'
  if (['recovering', 'Pending', 'revocation_pending'].includes(value))
    return 'status status-warn'
  if (['incompatible', 'configuration_failed', 'Error'].includes(value))
    return 'status status-error'
  return 'status status-muted'
}

export function Status({ value }: { value: string }): React.JSX.Element {
  return <span className={statusClass(value)}>{value.replaceAll('_', ' ')}</span>
}

export function Spinner({ size = 15 }: { size?: number }): React.JSX.Element {
  return <LoaderCircle className="spinner" size={size} aria-hidden="true" />
}

export function IconButton({ label, children, loading = false, disabled, ...props }: { label: string, children: ReactNode, loading?: boolean } & ButtonHTMLAttributes<HTMLButtonElement>): React.JSX.Element {
  return <button type="button" className="icon-button" aria-label={label} title={label} aria-busy={loading || undefined} disabled={disabled || loading} {...props}>{loading ? <Spinner /> : children}</button>
}

export function Switch({ label, checked, disabled, loading = false, onChange }: { label: string, checked: boolean, disabled?: boolean, loading?: boolean, onChange: (checked: boolean) => void }): React.JSX.Element {
  return (
    <button
      type="button"
      className={`switch${checked ? ' switch-on' : ''}`}
      role="switch"
      aria-checked={checked}
      aria-label={label}
      aria-busy={loading || undefined}
      disabled={disabled || loading}
      onClick={() => onChange(!checked)}
    >
      <span className="switch-thumb">{loading && <Spinner size={11} />}</span>
    </button>
  )
}

export function LoadingState({ label }: { label: string }): React.JSX.Element {
  return (
    <div className="load-state" role="status" aria-live="polite">
      <Spinner size={18} />
      <span>{label}</span>
    </div>
  )
}

export function ErrorState({ message, retrying = false, onRetry }: { message: string, retrying?: boolean, onRetry?: () => void }): React.JSX.Element {
  return (
    <div className="error-state" role="alert">
      <XCircle size={17} />
      <span>{message}</span>
      {onRetry && (
        <button type="button" disabled={retrying} onClick={onRetry}>
          {retrying ? <Spinner /> : <RefreshCw size={14} />}
          Retry
        </button>
      )}
    </div>
  )
}

export interface ValueFieldRow {
  id: number
  value: string
}

export interface KeyValueFieldRow {
  id: number
  name: string
  value: string
}

let fieldRowId = 0
export function valueFieldRow(value = ''): ValueFieldRow {
  return { id: ++fieldRowId, value }
}

export function keyValueFieldRow(name = '', value = ''): KeyValueFieldRow {
  return { id: ++fieldRowId, name, value }
}

export function ValueFieldRows({ label, rows, placeholder, minimum = 0, onChange }: { label: string, rows: ValueFieldRow[], placeholder?: string, minimum?: number, onChange: (rows: ValueFieldRow[]) => void }): React.JSX.Element {
  const container = useRef<HTMLFieldSetElement>(null)
  const add = (): void => {
    onChange([...rows, valueFieldRow()])
    requestAnimationFrame(() => container.current?.querySelector<HTMLInputElement>('[data-new-row="true"]')?.focus())
  }
  return (
    <fieldset className="structured-field" ref={container}>
      <legend>{label}</legend>
      <div className="structured-rows">
        {rows.map((row, index) => (
          <div className="value-field-row" key={row.id}>
            <input
              data-new-row={index === rows.length - 1 ? 'true' : undefined}
              aria-label={`${label} ${index + 1}`}
              placeholder={placeholder}
              value={row.value}
              required={minimum > 0}
              onChange={event => onChange(rows.map(value => value.id === row.id ? { ...value, value: event.target.value } : value))}
            />
            <IconButton label={`Remove ${label} ${index + 1}`} disabled={rows.length <= minimum} onClick={() => onChange(rows.filter(value => value.id !== row.id))}><Trash2 size={14} /></IconButton>
          </div>
        ))}
      </div>
      <button className="add-row" type="button" onClick={add}>
        <Plus size={14} />
        Add
        {' '}
        {label.toLowerCase()}
      </button>
    </fieldset>
  )
}

export function KeyValueFieldRows({ label, rows, onChange }: { label: string, rows: KeyValueFieldRow[], onChange: (rows: KeyValueFieldRow[]) => void }): React.JSX.Element {
  const container = useRef<HTMLFieldSetElement>(null)
  const add = (): void => {
    onChange([...rows, keyValueFieldRow()])
    requestAnimationFrame(() => container.current?.querySelector<HTMLInputElement>('[data-new-row="true"]')?.focus())
  }
  return (
    <fieldset className="structured-field" ref={container}>
      <legend>{label}</legend>
      <div className="structured-rows">
        {rows.map((row, index) => (
          <div className="key-value-field-row" key={row.id}>
            <input
              data-new-row={index === rows.length - 1 ? 'true' : undefined}
              aria-label={`${label} name ${index + 1}`}
              placeholder="Header name"
              value={row.name}
              required
              onChange={event => onChange(rows.map(value => value.id === row.id ? { ...value, name: event.target.value } : value))}
            />
            <input aria-label={`${label} value ${index + 1}`} placeholder="Value" value={row.value} onChange={event => onChange(rows.map(value => value.id === row.id ? { ...value, value: event.target.value } : value))} />
            <IconButton label={`Remove ${label} ${index + 1}`} onClick={() => onChange(rows.filter(value => value.id !== row.id))}><Trash2 size={14} /></IconButton>
          </div>
        ))}
      </div>
      <button className="add-row" type="button" onClick={add}>
        <Plus size={14} />
        Add header
      </button>
    </fieldset>
  )
}

export interface ConfirmationRequest {
  message: string
  action: () => Promise<void>
  successMessage?: string
}

export function Confirmation({ request, onClose }: { request: ConfirmationRequest, onClose: () => void }): React.JSX.Element {
  const { notify } = useFeedback()
  const [error, setError] = useState('')
  const [running, setRunning] = useState(false)
  const confirmAction = async (): Promise<void> => {
    setError('')
    setRunning(true)
    try {
      await request.action()
      if (request.successMessage)
        notify(request.successMessage)
      onClose()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setRunning(false)
    }
  }
  return (
    <div className="modal-backdrop" role="presentation">
      <div className="modal" role="alertdialog" aria-modal="true" aria-labelledby="confirmation-title">
        <div className="modal-title">
          <h2 id="confirmation-title">Confirm action</h2>
          <IconButton label="Close" onClick={onClose}><X size={16} /></IconButton>
        </div>
        <p>{request.message}</p>
        {error && <p className="form-error" role="alert">{error}</p>}
        <div className="modal-actions">
          <button type="button" disabled={running} onClick={onClose}>Cancel</button>
          <button className="danger" type="button" disabled={running} onClick={() => void confirmAction()}>
            {running && <Spinner />}
            {running ? 'Working...' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  )
}

export function PageHeader({ title, actions }: { title: string, actions?: ReactNode }): React.JSX.Element {
  return (
    <header className="page-header">
      <h1>{title}</h1>
      <div className="actions">{actions}</div>
    </header>
  )
}

export function Token({ value }: { value: string }): React.JSX.Element {
  const [copied, setCopied] = useState(false)
  const { notify } = useFeedback()
  const copy = async (): Promise<void> => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 1200)
    }
    catch (cause) {
      notify(cause instanceof Error ? cause.message : 'Could not copy Client Token', 'error')
    }
  }
  return (
    <div className="token">
      <code>{value}</code>
      <IconButton label="Copy Client Token" onClick={() => void copy()}>{copied ? <Check size={15} /> : <Clipboard size={15} />}</IconButton>
    </div>
  )
}
