import type { ReactNode } from 'react'
import { Check, Clipboard, X } from 'lucide-react'
import { useState } from 'react'

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

export function IconButton({ label, children, ...props }: { label: string, children: ReactNode } & React.ButtonHTMLAttributes<HTMLButtonElement>): React.JSX.Element {
  return <button type="button" className="icon-button" aria-label={label} title={label} {...props}>{children}</button>
}

export interface ConfirmationRequest {
  message: string
  action: () => Promise<void>
}

export function Confirmation({ request, onClose }: { request: ConfirmationRequest, onClose: () => void }): React.JSX.Element {
  const [error, setError] = useState('')
  const [running, setRunning] = useState(false)
  const confirmAction = async (): Promise<void> => {
    setError('')
    setRunning(true)
    try {
      await request.action()
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
        {error && <p className="form-error">{error}</p>}
        <div className="modal-actions">
          <button type="button" disabled={running} onClick={onClose}>Cancel</button>
          <button className="danger" type="button" disabled={running} onClick={() => void confirmAction()}>{running ? 'Working...' : 'Confirm'}</button>
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
  const copy = async (): Promise<void> => {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 1200)
  }
  return (
    <div className="token">
      <code>{value}</code>
      <IconButton label="Copy Client Token" onClick={() => void copy()}>{copied ? <Check size={15} /> : <Clipboard size={15} />}</IconButton>
    </div>
  )
}
