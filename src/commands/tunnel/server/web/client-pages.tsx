import type { FormEvent } from 'react'
import type { TunnelHeader } from '../../types'
import type { ClientView, TunnelView } from './api'
import type { ConfirmationRequest, KeyValueFieldRow, ValueFieldRow } from './ui'
import { ArrowLeft, ArrowRight, ChevronDown, ChevronRight, Pencil, Plus, RefreshCw, RotateCcw, Trash2, X } from 'lucide-react'
import { Fragment, useCallback, useEffect, useRef, useState } from 'react'
import { apiJson, jsonRequest } from './api'
import {
  Confirmation,
  ErrorState,
  IconButton,
  keyValueFieldRow,
  KeyValueFieldRows,
  LoadingState,
  navigate,
  PageHeader,
  Spinner,
  Status,
  Switch,
  Token,
  useFeedback,
  valueFieldRow,
  ValueFieldRows,
} from './ui'

function message(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause)
}

function ClientRemarkEditor({ client, onClose, onSaved }: { client?: ClientView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [remark, setRemark] = useState(client?.remark ?? '')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const { notify } = useFeedback()
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    setError('')
    setSaving(true)
    try {
      await apiJson(client ? `/api/clients/${encodeURIComponent(client.id)}` : '/api/clients', jsonRequest(client ? 'PATCH' : 'POST', { remark }))
      notify(client ? 'Client remark saved' : 'Trusted client created')
      onSaved()
    }
    catch (cause) {
      setError(message(cause))
    }
    finally {
      setSaving(false)
    }
  }
  const title = client ? 'Edit Client Remark' : 'Create client'
  return (
    <div className="modal-backdrop" role="presentation">
      <form className="modal" role="dialog" aria-modal="true" aria-labelledby="client-editor-title" aria-busy={saving} onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2 id="client-editor-title">{title}</h2>
          <IconButton label="Close" disabled={saving} onClick={onClose}><X size={16} /></IconButton>
        </div>
        <label>
          Client Remark
          <textarea value={remark} maxLength={100} autoFocus rows={4} disabled={saving} onChange={event => setRemark(event.target.value)} />
        </label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <div className="modal-actions">
          <button type="button" disabled={saving} onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>
            {saving && <Spinner />}
            {saving ? 'Saving...' : client ? 'Save' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  )
}

export function ClientsPage({ refreshSequence, showOwner }: { refreshSequence: number, showOwner: boolean }): React.JSX.Element {
  const [clients, setClients] = useState<ClientView[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [confirmation, setConfirmation] = useState<ConfirmationRequest>()
  const [editing, setEditing] = useState<ClientView | null | undefined>()
  const loaded = useRef(false)
  const requestId = useRef(0)
  const load = useCallback(async () => {
    const currentRequest = ++requestId.current
    loaded.current ? setRefreshing(true) : setLoading(true)
    try {
      const result = (await apiJson<{ clients: ClientView[] }>('/api/clients')).clients
      if (currentRequest !== requestId.current)
        return
      setClients(result)
      setError('')
      loaded.current = true
    }
    catch (cause) {
      if (currentRequest === requestId.current)
        setError(message(cause))
    }
    finally {
      if (currentRequest === requestId.current) {
        setLoading(false)
        setRefreshing(false)
      }
    }
  }, [])
  useEffect(() => void load(), [load, refreshSequence])
  const rotate = async (id: string): Promise<void> => {
    await apiJson(`/api/clients/${encodeURIComponent(id)}/rotate`, { method: 'POST' })
    await load()
  }
  const remove = async (id: string): Promise<void> => {
    await apiJson(`/api/clients/${encodeURIComponent(id)}`, { method: 'DELETE' })
    await load()
  }
  const saved = (): void => {
    setEditing(undefined)
    void load()
  }
  return (
    <>
      <PageHeader
        title="Trusted Tunnel Clients"
        actions={(
          <>
            <IconButton label="Refresh clients" loading={refreshing} onClick={() => void load()}><RefreshCw size={15} /></IconButton>
            <button className="primary" type="button" onClick={() => setEditing(null)}>
              <Plus size={15} />
              Create client
            </button>
          </>
        )}
      />
      {loading && !loaded.current
        ? <LoadingState label="Loading trusted clients" />
        : error && !loaded.current
          ? <ErrorState message={error} retrying={loading} onRetry={() => void load()} />
          : (
              <>
                {error && <ErrorState message={error} retrying={refreshing} onRetry={() => void load()} />}
                <section className="table-wrap" aria-busy={refreshing}>
                  <table>
                    <thead>
                      <tr>
                        <th>Client Token</th>
                        <th>Client Remark</th>
                        {showOwner && <th>Owner</th>}
                        <th>Connection</th>
                        <th>Revision</th>
                        <th>Tunnels</th>
                        <th aria-label="Actions" />
                      </tr>
                    </thead>
                    <tbody>
                      {clients.map(client => (
                        <tr key={client.id}>
                          <td className="client-token"><Token value={client.token} /></td>
                          <td className="client-remark">{client.remark || 'Unlabeled client'}</td>
                          {showOwner && <td>{client.owner.username}</td>}
                          <td><Status value={client.runtime.connectionState} /></td>
                          <td className="mono">
                            {client.lastAppliedRevision}
                            {' '}
                            /
                            {' '}
                            {client.desiredRevision}
                          </td>
                          <td>{client.tunnelCounts.total}</td>
                          <td>
                            <div className="row-actions">
                              <IconButton label="Edit Client Remark" onClick={() => setEditing(client)}><Pencil size={15} /></IconButton>
                              <IconButton label="Open client" onClick={() => navigate(`/clients/${encodeURIComponent(client.id)}`)}><ArrowRight size={15} /></IconButton>
                              <IconButton label="Rotate Client Token" onClick={() => setConfirmation({ message: 'Rotate this Client Token? The active client will stop.', successMessage: 'Client Token rotated', action: () => rotate(client.id) })}><RotateCcw size={15} /></IconButton>
                              <IconButton label="Delete client" onClick={() => setConfirmation({ message: 'Delete this trusted client and all of its Tunnel Definitions?', successMessage: 'Trusted client deleted', action: () => remove(client.id) })}><Trash2 size={15} /></IconButton>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {!clients.length && <div className="empty-row">No trusted clients</div>}
                </section>
              </>
            )}
      {editing !== undefined && <ClientRemarkEditor client={editing ?? undefined} onClose={() => setEditing(undefined)} onSaved={saved} />}
      {confirmation && <Confirmation request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}

interface TunnelDraft {
  label: string
  protocol: 'http' | 'tcp' | 'udp'
  customDomains: ValueFieldRow[]
  location: string
  serverPort: string
  localHost: string
  localPort: string
  enabled: boolean
  useEncryption: boolean
  useCompression: boolean
  bandwidthEnabled: boolean
  bandwidthValue: string
  bandwidthUnit: 'KB' | 'MB'
  bandwidthMode: 'client' | 'server'
  proxyProtocolVersion: '' | 'v1' | 'v2'
  healthEnabled: boolean
  healthType: 'tcp' | 'http'
  healthInterval: string
  healthTimeout: string
  healthMaxFailed: string
  healthPath: string
  healthHeaders: KeyValueFieldRow[]
  authEnabled: boolean
  authUsername: string
  authPassword: string
  hostHeaderRewrite: string
  requestHeaders: KeyValueFieldRow[]
  responseHeaders: KeyValueFieldRow[]
}

function keyValueRows(headers: TunnelHeader[] | undefined): KeyValueFieldRow[] {
  return (headers ?? []).map(header => keyValueFieldRow(header.name, header.value))
}

function headerValues(rows: KeyValueFieldRow[]): TunnelHeader[] {
  return rows.map(row => ({ name: row.name.trim(), value: row.value.trim() })).filter(row => row.name)
}

function draftFor(initial?: TunnelView): TunnelDraft {
  const httpTunnel = initial?.protocol === 'http' ? initial : undefined
  const transport = initial?.options.transport
  const health = initial?.options.healthCheck
  const http = httpTunnel?.options.http
  return {
    label: initial?.label ?? '',
    protocol: initial?.protocol ?? 'http',
    customDomains: httpTunnel ? httpTunnel.customDomains.map(value => valueFieldRow(value)) : [valueFieldRow()],
    location: httpTunnel?.location ?? '',
    serverPort: initial?.serverPort?.toString() ?? '',
    localHost: initial?.localHost ?? '127.0.0.1',
    localPort: initial?.localPort.toString() ?? '',
    enabled: initial?.enabled ?? true,
    useEncryption: transport?.useEncryption ?? false,
    useCompression: transport?.useCompression ?? false,
    bandwidthEnabled: Boolean(transport?.bandwidthLimit),
    bandwidthValue: transport?.bandwidthLimit?.value.toString() ?? '1',
    bandwidthUnit: transport?.bandwidthLimit?.unit ?? 'MB',
    bandwidthMode: transport?.bandwidthLimit?.mode ?? 'client',
    proxyProtocolVersion: transport?.proxyProtocolVersion ?? '',
    healthEnabled: Boolean(health),
    healthType: health?.type ?? 'tcp',
    healthInterval: health?.intervalSeconds.toString() ?? '10',
    healthTimeout: health?.timeoutSeconds.toString() ?? '3',
    healthMaxFailed: health?.maxFailed.toString() ?? '3',
    healthPath: health?.type === 'http' ? health.path : '/health',
    healthHeaders: health?.type === 'http' ? keyValueRows(health.headers) : [],
    authEnabled: Boolean(http?.basicAuth),
    authUsername: http?.basicAuth?.username ?? '',
    authPassword: '',
    hostHeaderRewrite: http?.hostHeaderRewrite ?? '',
    requestHeaders: keyValueRows(http?.requestHeaders),
    responseHeaders: keyValueRows(http?.responseHeaders),
  }
}

type TunnelEditorStep = 'basics' | 'transport' | 'health' | 'http'

function validateTunnelDraft(draft: TunnelDraft, initial?: TunnelView): { message: string, step: TunnelEditorStep } | undefined {
  const localPort = Number(draft.localPort)
  if (!draft.localHost.trim())
    return { message: 'Local host is required', step: 'basics' }
  if (!Number.isSafeInteger(localPort) || localPort < 1 || localPort > 65535)
    return { message: 'Local port must be between 1 and 65535', step: 'basics' }
  if (draft.protocol === 'http') {
    if (!draft.customDomains.some(row => row.value.trim()))
      return { message: 'At least one custom domain is required', step: 'basics' }
    if (draft.location && (!draft.location.startsWith('/') || /\s/.test(draft.location)))
      return { message: 'Location must begin with / and contain no spaces', step: 'basics' }
  }
  else if (draft.serverPort) {
    const serverPort = Number(draft.serverPort)
    if (!Number.isSafeInteger(serverPort) || serverPort < 1 || serverPort > 65535)
      return { message: 'Server port must be between 1 and 65535', step: 'basics' }
  }
  if (draft.bandwidthEnabled && (!Number.isFinite(Number(draft.bandwidthValue)) || Number(draft.bandwidthValue) <= 0))
    return { message: 'Bandwidth limit must be greater than zero', step: 'transport' }
  if (draft.healthEnabled) {
    const values = [draft.healthInterval, draft.healthTimeout, draft.healthMaxFailed].map(Number)
    if (values.some(value => !Number.isSafeInteger(value) || value < 1))
      return { message: 'Health check timing values must be positive integers', step: 'health' }
    if (draft.healthType === 'http' && (!draft.healthPath.startsWith('/') || /\s/.test(draft.healthPath)))
      return { message: 'Health path must begin with / and contain no spaces', step: 'health' }
  }
  if (draft.protocol === 'http' && draft.authEnabled) {
    if (!draft.authUsername.trim())
      return { message: 'Basic Auth username is required', step: 'http' }
    if (!draft.authPassword && !initial?.options.http?.basicAuth)
      return { message: 'Basic Auth password is required', step: 'http' }
  }
}

function TunnelEditor({ clientId, initial, onClose, onSaved }: { clientId: string, initial?: TunnelView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [draft, setDraft] = useState<TunnelDraft>(() => draftFor(initial))
  const [activeStep, setActiveStep] = useState<TunnelEditorStep>('basics')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const formRef = useRef<HTMLFormElement>(null)
  const { notify } = useFeedback()
  const steps: Array<{ id: TunnelEditorStep, label: string }> = [
    { id: 'basics', label: 'Basics' },
    { id: 'transport', label: 'Transport' },
    { id: 'health', label: 'Health check' },
    ...(draft.protocol === 'http' ? [{ id: 'http' as const, label: 'HTTP' }] : []),
  ]
  const activeStepIndex = steps.findIndex(step => step.id === activeStep)
  const selectStep = (step: TunnelEditorStep): void => {
    setError('')
    setActiveStep(step)
  }
  const nextStep = (): void => {
    if (!formRef.current?.reportValidity())
      return
    const next = steps[activeStepIndex + 1]
    if (next)
      selectStep(next.id)
  }
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    if (activeStepIndex < steps.length - 1) {
      nextStep()
      return
    }
    const validation = validateTunnelDraft(draft, initial)
    if (validation) {
      setError(validation.message)
      setActiveStep(validation.step)
      return
    }
    setError('')
    setSaving(true)
    const healthCheck = draft.healthEnabled
      ? draft.healthType === 'http'
        ? { type: 'http' as const, path: draft.healthPath, intervalSeconds: Number(draft.healthInterval), timeoutSeconds: Number(draft.healthTimeout), maxFailed: Number(draft.healthMaxFailed), headers: headerValues(draft.healthHeaders) }
        : { type: 'tcp' as const, intervalSeconds: Number(draft.healthInterval), timeoutSeconds: Number(draft.healthTimeout), maxFailed: Number(draft.healthMaxFailed) }
      : null
    const basicAuth = draft.authEnabled
      ? { username: draft.authUsername, ...(draft.authPassword ? { password: draft.authPassword } : {}) }
      : null
    const body = {
      label: draft.label,
      protocol: draft.protocol,
      customDomains: draft.protocol === 'http' ? draft.customDomains.map(row => row.value.trim()).filter(Boolean) : undefined,
      location: draft.protocol === 'http' ? draft.location.trim() || null : null,
      serverPort: draft.protocol === 'http' || !draft.serverPort ? null : Number(draft.serverPort),
      localHost: draft.localHost,
      localPort: Number(draft.localPort),
      enabled: draft.enabled,
      options: {
        transport: {
          useEncryption: draft.useEncryption,
          useCompression: draft.useCompression,
          bandwidthLimit: draft.bandwidthEnabled ? { value: Number(draft.bandwidthValue), unit: draft.bandwidthUnit, mode: draft.bandwidthMode } : null,
          proxyProtocolVersion: draft.proxyProtocolVersion || null,
        },
        healthCheck,
        http: draft.protocol === 'http'
          ? {
              basicAuth,
              hostHeaderRewrite: draft.hostHeaderRewrite.trim() || null,
              requestHeaders: headerValues(draft.requestHeaders),
              responseHeaders: headerValues(draft.responseHeaders),
            }
          : null,
      },
    }
    try {
      await apiJson(initial ? `/api/tunnels/${encodeURIComponent(initial.id)}` : `/api/clients/${encodeURIComponent(clientId)}/tunnels`, jsonRequest(initial ? 'PATCH' : 'POST', body))
      notify(initial ? 'Tunnel Definition saved' : 'Tunnel Definition created')
      onSaved()
    }
    catch (cause) {
      setError(message(cause))
    }
    finally {
      setSaving(false)
    }
  }
  return (
    <div className="modal-backdrop" role="presentation">
      <form ref={formRef} className="modal tunnel-modal" role="dialog" aria-modal="true" aria-labelledby="tunnel-editor-title" aria-busy={saving} onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2 id="tunnel-editor-title">{initial ? 'Edit Tunnel Definition' : 'New Tunnel Definition'}</h2>
          <IconButton label="Close" disabled={saving} onClick={onClose}><X size={16} /></IconButton>
        </div>
        <div className="tunnel-steps" role="tablist" aria-label="Tunnel configuration steps">
          {steps.map((step, index) => (
            <button
              id={`tunnel-step-${step.id}`}
              type="button"
              role="tab"
              aria-controls={`tunnel-panel-${step.id}`}
              aria-selected={activeStep === step.id}
              disabled={saving}
              className={activeStep === step.id ? 'active' : ''}
              key={step.id}
              onClick={() => selectStep(step.id)}
            >
              <span className="step-number">{index + 1}</span>
              {step.label}
            </button>
          ))}
        </div>

        {error && <p className="form-error tunnel-form-error" role="alert">{error}</p>}

        <div className="tunnel-step-panel" id={`tunnel-panel-${activeStep}`} role="tabpanel" aria-labelledby={`tunnel-step-${activeStep}`}>
          {activeStep === 'basics' && (
            <div className="step-stack">
              <label>
                Display name
                <input value={draft.label} maxLength={100} autoFocus onChange={event => setDraft({ ...draft, label: event.target.value })} />
              </label>
              <div className="segmented" aria-label="Tunnel protocol">{(['http', 'tcp', 'udp'] as const).map(value => <button type="button" disabled={saving} className={draft.protocol === value ? 'active' : ''} key={value} onClick={() => setDraft({ ...draft, protocol: value })}>{value.toUpperCase()}</button>)}</div>
              {draft.protocol === 'http'
                ? (
                    <>
                      <ValueFieldRows label="Custom domains" rows={draft.customDomains} minimum={1} placeholder="routes.example.com" onChange={customDomains => setDraft({ ...draft, customDomains })} />
                      <label>
                        Location
                        <input value={draft.location} placeholder="All paths" onChange={event => setDraft({ ...draft, location: event.target.value })} />
                      </label>
                    </>
                  )
                : (
                    <label>
                      Server port
                      <input type="number" min="1" max="65535" value={draft.serverPort} placeholder="Automatic" onChange={event => setDraft({ ...draft, serverPort: event.target.value })} />
                    </label>
                  )}
              <div className="field-row">
                <label>
                  Local host
                  <input value={draft.localHost} required onChange={event => setDraft({ ...draft, localHost: event.target.value })} />
                </label>
                <label>
                  Local port
                  <input type="number" min="1" max="65535" value={draft.localPort} required onChange={event => setDraft({ ...draft, localPort: event.target.value })} />
                </label>
              </div>
              <div className="setting-row">
                <div><strong>Enabled</strong></div>
                <Switch label="Enable Tunnel Definition" checked={draft.enabled} onChange={enabled => setDraft({ ...draft, enabled })} />
              </div>
            </div>
          )}

          {activeStep === 'transport' && (
            <section className="form-section">
              <h3>Transport</h3>
              <div className="setting-row">
                <span>Encryption</span>
                <Switch label="Use encryption" checked={draft.useEncryption} onChange={useEncryption => setDraft({ ...draft, useEncryption })} />
              </div>
              <div className="setting-row">
                <span>Compression</span>
                <Switch label="Use compression" checked={draft.useCompression} onChange={useCompression => setDraft({ ...draft, useCompression })} />
              </div>
              <div className="setting-row">
                <span>Bandwidth limit</span>
                <Switch label="Limit bandwidth" checked={draft.bandwidthEnabled} onChange={bandwidthEnabled => setDraft({ ...draft, bandwidthEnabled })} />
              </div>
              {draft.bandwidthEnabled && (
                <div className="field-row three-fields">
                  <label>
                    Limit
                    <input type="number" min="0.01" step="any" value={draft.bandwidthValue} required onChange={event => setDraft({ ...draft, bandwidthValue: event.target.value })} />
                  </label>
                  <label>
                    Unit
                    <select value={draft.bandwidthUnit} onChange={event => setDraft({ ...draft, bandwidthUnit: event.target.value as 'KB' | 'MB' })}>
                      <option>KB</option>
                      <option>MB</option>
                    </select>
                  </label>
                  <label>
                    Limit at
                    <select value={draft.bandwidthMode} onChange={event => setDraft({ ...draft, bandwidthMode: event.target.value as 'client' | 'server' })}>
                      <option value="client">Client</option>
                      <option value="server">Server</option>
                    </select>
                  </label>
                </div>
              )}
              <label>
                Proxy Protocol
                <select value={draft.proxyProtocolVersion} onChange={event => setDraft({ ...draft, proxyProtocolVersion: event.target.value as '' | 'v1' | 'v2' })}>
                  <option value="">Off</option>
                  <option value="v1">v1</option>
                  <option value="v2">v2</option>
                </select>
              </label>
            </section>
          )}

          {activeStep === 'health' && (
            <section className="form-section">
              <div className="setting-row">
                <h3>Health check</h3>
                <Switch label="Enable health check" checked={draft.healthEnabled} onChange={healthEnabled => setDraft({ ...draft, healthEnabled })} />
              </div>
              {draft.healthEnabled && (
                <>
                  <div className="segmented two-segments">{(['tcp', 'http'] as const).map(value => <button type="button" className={draft.healthType === value ? 'active' : ''} key={value} onClick={() => setDraft({ ...draft, healthType: value })}>{value.toUpperCase()}</button>)}</div>
                  <div className="field-row three-fields">
                    <label>
                      Interval (s)
                      <input type="number" min="1" value={draft.healthInterval} required onChange={event => setDraft({ ...draft, healthInterval: event.target.value })} />
                    </label>
                    <label>
                      Timeout (s)
                      <input type="number" min="1" value={draft.healthTimeout} required onChange={event => setDraft({ ...draft, healthTimeout: event.target.value })} />
                    </label>
                    <label>
                      Max failed
                      <input type="number" min="1" value={draft.healthMaxFailed} required onChange={event => setDraft({ ...draft, healthMaxFailed: event.target.value })} />
                    </label>
                  </div>
                  {draft.healthType === 'http' && (
                    <>
                      <label>
                        Health path
                        <input value={draft.healthPath} required onChange={event => setDraft({ ...draft, healthPath: event.target.value })} />
                      </label>
                      <KeyValueFieldRows label="Health check headers" rows={draft.healthHeaders} onChange={healthHeaders => setDraft({ ...draft, healthHeaders })} />
                    </>
                  )}
                </>
              )}
            </section>
          )}

          {activeStep === 'http' && draft.protocol === 'http' && (
            <section className="form-section">
              <h3>HTTP</h3>
              <div className="setting-row">
                <span>Basic Auth</span>
                <Switch label="Enable HTTP Basic Auth" checked={draft.authEnabled} onChange={authEnabled => setDraft({ ...draft, authEnabled })} />
              </div>
              {draft.authEnabled && (
                <div className="field-row">
                  <label>
                    Username
                    <input value={draft.authUsername} required onChange={event => setDraft({ ...draft, authUsername: event.target.value })} />
                  </label>
                  <label>
                    Password
                    <input type="password" value={draft.authPassword} required={!initial?.options.http?.basicAuth} placeholder={initial?.options.http?.basicAuth ? 'Unchanged' : ''} autoComplete="new-password" onChange={event => setDraft({ ...draft, authPassword: event.target.value })} />
                  </label>
                </div>
              )}
              <label>
                Host Header Rewrite
                <input value={draft.hostHeaderRewrite} onChange={event => setDraft({ ...draft, hostHeaderRewrite: event.target.value })} />
              </label>
              <KeyValueFieldRows label="Request headers" rows={draft.requestHeaders} onChange={requestHeaders => setDraft({ ...draft, requestHeaders })} />
              <KeyValueFieldRows label="Response headers" rows={draft.responseHeaders} onChange={responseHeaders => setDraft({ ...draft, responseHeaders })} />
            </section>
          )}
        </div>

        <div className="modal-actions">
          <button type="button" disabled={saving} onClick={onClose}>Cancel</button>
          <div className="modal-step-actions">
            {activeStepIndex > 0 && (
              <button type="button" disabled={saving} onClick={() => selectStep(steps[activeStepIndex - 1]!.id)}>
                <ArrowLeft size={15} />
                Back
              </button>
            )}
            {activeStepIndex < steps.length - 1
              ? (
                  <button className="primary" type="button" disabled={saving} onClick={nextStep}>
                    Next
                    <ArrowRight size={15} />
                  </button>
                )
              : (
                  <button className="primary" type="submit" disabled={saving}>
                    {saving && <Spinner />}
                    {saving ? 'Saving...' : 'Save'}
                  </button>
                )}
          </div>
        </div>
      </form>
    </div>
  )
}

function TunnelDetails({ tunnel }: { tunnel: TunnelView }): React.JSX.Element {
  const transport = tunnel.options.transport
  const health = tunnel.options.healthCheck
  return (
    <div className="tunnel-details">
      <dl>
        <dt>Transport</dt>
        <dd>{[transport.useEncryption && 'Encrypted', transport.useCompression && 'Compressed', transport.proxyProtocolVersion && `Proxy Protocol ${transport.proxyProtocolVersion}`].filter(Boolean).join(', ') || 'Defaults'}</dd>
        <dt>Bandwidth</dt>
        <dd>{transport.bandwidthLimit ? `${transport.bandwidthLimit.value}${transport.bandwidthLimit.unit} at ${transport.bandwidthLimit.mode}` : 'Unlimited'}</dd>
        <dt>Health check</dt>
        <dd>{health ? `${health.type.toUpperCase()} every ${health.intervalSeconds}s` : 'Off'}</dd>
        {tunnel.protocol === 'http' && (
          <>
            <dt>Basic Auth</dt>
            <dd>{tunnel.options.http?.basicAuth ? `Configured for ${tunnel.options.http.basicAuth.username}` : 'Off'}</dd>
            <dt>Host rewrite</dt>
            <dd>{tunnel.options.http?.hostHeaderRewrite || 'Off'}</dd>
            <dt>Header sets</dt>
            <dd>
              {(tunnel.options.http?.requestHeaders.length ?? 0)}
              {' '}
              request /
              {' '}
              {(tunnel.options.http?.responseHeaders.length ?? 0)}
              {' '}
              response
            </dd>
          </>
        )}
      </dl>
    </div>
  )
}

export function ClientDetailPage({ id, refreshSequence, showOwner }: { id: string, refreshSequence: number, showOwner: boolean }): React.JSX.Element {
  const [client, setClient] = useState<ClientView>()
  const [tunnels, setTunnels] = useState<TunnelView[]>([])
  const [editing, setEditing] = useState<TunnelView | null | undefined>()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [pending, setPending] = useState<Set<string>>(new Set())
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [confirmation, setConfirmation] = useState<ConfirmationRequest>()
  const loaded = useRef(false)
  const requestId = useRef(0)
  const { notify } = useFeedback()
  const load = useCallback(async () => {
    const currentRequest = ++requestId.current
    loaded.current ? setRefreshing(true) : setLoading(true)
    try {
      const body = await apiJson<{ client: ClientView, tunnels: TunnelView[] }>(`/api/clients/${encodeURIComponent(id)}`)
      if (currentRequest !== requestId.current)
        return
      setClient(body.client)
      setTunnels(body.tunnels)
      setError('')
      loaded.current = true
    }
    catch (cause) {
      if (currentRequest === requestId.current)
        setError(message(cause))
    }
    finally {
      if (currentRequest === requestId.current) {
        setLoading(false)
        setRefreshing(false)
      }
    }
  }, [id])
  useEffect(() => void load(), [load, refreshSequence])
  const run = async (key: string, successMessage: string, action: () => Promise<void>): Promise<void> => {
    setPending(values => new Set(values).add(key))
    setError('')
    try {
      await action()
      notify(successMessage)
    }
    catch (cause) {
      const value = message(cause)
      setError(value)
      notify(value, 'error')
    }
    finally {
      setPending((values) => {
        const next = new Set(values)
        next.delete(key)
        return next
      })
    }
  }
  const toggle = async (tunnel: TunnelView): Promise<void> => run(`toggle:${tunnel.id}`, tunnel.enabled ? 'Tunnel Definition disabled' : 'Tunnel Definition enabled', async () => {
    await apiJson(`/api/tunnels/${encodeURIComponent(tunnel.id)}`, jsonRequest('PATCH', { enabled: !tunnel.enabled }))
    await load()
  })
  const remove = async (tunnel: TunnelView): Promise<void> => {
    await apiJson(`/api/tunnels/${encodeURIComponent(tunnel.id)}`, { method: 'DELETE' })
    await load()
  }
  const restart = async (): Promise<void> => run('restart', 'frpc restart requested', async () => {
    await apiJson(`/api/clients/${encodeURIComponent(id)}/restart`, { method: 'POST' })
  })
  const toggleDetails = (tunnelId: string): void => setExpanded((values) => {
    const next = new Set(values)
    next.has(tunnelId) ? next.delete(tunnelId) : next.add(tunnelId)
    return next
  })
  const saved = (): void => {
    setEditing(undefined)
    void load()
  }
  return (
    <>
      <PageHeader
        title={client?.remark || 'Unlabeled client'}
        actions={(
          <>
            <button type="button" onClick={() => navigate('/clients')}>
              <ArrowLeft size={15} />
              Clients
            </button>
            <IconButton label="Refresh client" loading={refreshing} onClick={() => void load()}><RefreshCw size={15} /></IconButton>
            <button type="button" disabled={pending.has('restart')} aria-busy={pending.has('restart')} onClick={() => void restart()}>
              {pending.has('restart') ? <Spinner /> : <RefreshCw size={15} />}
              Restart frpc
            </button>
            <button className="primary" type="button" onClick={() => setEditing(null)}>
              <Plus size={15} />
              New tunnel
            </button>
          </>
        )}
      />
      {loading && !loaded.current
        ? <LoadingState label="Loading client configuration" />
        : error && !loaded.current
          ? <ErrorState message={error} retrying={loading} onRetry={() => void load()} />
          : (
              <>
                {error && <ErrorState message={error} retrying={refreshing} onRetry={() => void load()} />}
                {client && (
                  <section className="client-strip">
                    <Token value={client.token} />
                    <Status value={client.runtime.connectionState} />
                    <Status value={client.runtime.processState} />
                    {showOwner && (
                      <span className="mono">
                        Owner
                        {' '}
                        {client.owner.username}
                      </span>
                    )}
                    <span className="mono">
                      Revision
                      {' '}
                      {client.lastAppliedRevision}
                      {' '}
                      /
                      {' '}
                      {client.desiredRevision}
                    </span>
                  </section>
                )}
                {client?.runtime.lastError && <p className="runtime-error" role="alert">{client.runtime.lastError.message}</p>}
                <section className="table-wrap" aria-busy={refreshing}>
                  <table className="tunnel-table">
                    <thead>
                      <tr>
                        <th aria-label="Details" />
                        <th>Name</th>
                        <th>Public mapping</th>
                        <th>Local Endpoint</th>
                        <th>Status</th>
                        <th>Enabled</th>
                        <th aria-label="Actions" />
                      </tr>
                    </thead>
                    <tbody>
                      {tunnels.map(tunnel => (
                        <Fragment key={tunnel.id}>
                          <tr className={`tunnel-row${expanded.has(tunnel.id) ? ' is-expanded' : ''}`} key={tunnel.id}>
                            <td data-label="Details"><IconButton label={expanded.has(tunnel.id) ? 'Hide advanced options' : 'Show advanced options'} onClick={() => toggleDetails(tunnel.id)}>{expanded.has(tunnel.id) ? <ChevronDown size={15} /> : <ChevronRight size={15} />}</IconButton></td>
                            <td data-label="Name"><strong>{tunnel.label || 'Unlabeled tunnel'}</strong></td>
                            <td data-label="Public mapping">
                              <div className="public-mapping">
                                <span>
                                  <strong>{tunnel.protocol.toUpperCase()}</strong>
                                  {' '}
                                  <span className="mono">{tunnel.protocol === 'http' ? tunnel.customDomains.join(', ') : tunnel.serverPort}</span>
                                </span>
                                {tunnel.protocol === 'http' && <span className="mono mapping-paths">{tunnel.location ?? 'All paths'}</span>}
                              </div>
                            </td>
                            <td className="mono" data-label="Local endpoint">
                              {tunnel.localHost}
                              :
                              {tunnel.localPort}
                            </td>
                            <td data-label="Status"><Status value={tunnel.state} /></td>
                            <td data-label="Enabled"><Switch label={`${tunnel.enabled ? 'Disable' : 'Enable'} ${tunnel.label || tunnel.id}`} checked={tunnel.enabled} loading={pending.has(`toggle:${tunnel.id}`)} onChange={() => void toggle(tunnel)} /></td>
                            <td data-label="Actions">
                              <div className="row-actions">
                                <IconButton label="Edit tunnel" onClick={() => setEditing(tunnel)}><Pencil size={15} /></IconButton>
                                <IconButton label="Delete tunnel" onClick={() => setConfirmation({ message: 'Delete this Tunnel Definition?', successMessage: 'Tunnel Definition deleted', action: () => remove(tunnel) })}><Trash2 size={15} /></IconButton>
                              </div>
                            </td>
                          </tr>
                          {expanded.has(tunnel.id) && <tr className="details-row" key={`${tunnel.id}:details`}><td colSpan={7}><TunnelDetails tunnel={tunnel} /></td></tr>}
                        </Fragment>
                      ))}
                    </tbody>
                  </table>
                  {!tunnels.length && <div className="empty-row">No Tunnel Definitions</div>}
                </section>
              </>
            )}
      {editing !== undefined && <TunnelEditor clientId={id} initial={editing ?? undefined} onClose={() => setEditing(undefined)} onSaved={saved} />}
      {confirmation && <Confirmation request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}
