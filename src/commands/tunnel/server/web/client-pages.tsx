import type { FormEvent } from 'react'
import type { ClientView, TunnelView } from './api'
import type { ConfirmationRequest } from './ui'
import { ArrowLeft, ArrowRight, Pencil, Plus, RefreshCw, RotateCcw, Trash2, X } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { apiJson, jsonRequest } from './api'
import { Confirmation, IconButton, navigate, PageHeader, Status, Token } from './ui'

function ClientRemarkEditor({ client, onClose, onSaved }: { client?: ClientView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [remark, setRemark] = useState(client?.remark ?? '')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    setSaving(true)
    try {
      await apiJson(client ? `/api/clients/${encodeURIComponent(client.id)}` : '/api/clients', jsonRequest(client ? 'PATCH' : 'POST', { remark }))
      onSaved()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setSaving(false)
    }
  }
  const title = client ? 'Edit Client Remark' : 'Create client'
  return (
    <div className="modal-backdrop" role="presentation">
      <form className="modal" role="dialog" aria-modal="true" aria-labelledby="client-editor-title" onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2 id="client-editor-title">{title}</h2>
          <IconButton label="Close" onClick={onClose}><X size={16} /></IconButton>
        </div>
        <label>
          Client Remark
          <textarea value={remark} maxLength={100} autoFocus rows={4} onChange={event => setRemark(event.target.value)} />
        </label>
        {error && <p className="form-error">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>{saving ? 'Saving...' : client ? 'Save' : 'Create'}</button>
        </div>
      </form>
    </div>
  )
}

export function ClientsPage({ refreshSequence, showOwner }: { refreshSequence: number, showOwner: boolean }): React.JSX.Element {
  const [clients, setClients] = useState<ClientView[]>([])
  const [error, setError] = useState('')
  const [confirmation, setConfirmation] = useState<ConfirmationRequest>()
  const [editing, setEditing] = useState<ClientView | null | undefined>()
  const load = useCallback(async () => {
    try {
      setClients((await apiJson<{ clients: ClientView[] }>('/api/clients')).clients)
      setError('')
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
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
  return (
    <>
      <PageHeader
        title="Trusted Tunnel Clients"
        actions={(
          <button className="primary" type="button" onClick={() => setEditing(null)}>
            <Plus size={15} />
            Create client
          </button>
        )}
      />
      {error && <p className="runtime-error">{error}</p>}
      <section className="table-wrap">
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
                    <IconButton label="Rotate Client Token" onClick={() => setConfirmation({ message: 'Rotate this Client Token? The active client will stop.', action: () => rotate(client.id) })}><RotateCcw size={15} /></IconButton>
                    <IconButton label="Delete client" onClick={() => setConfirmation({ message: 'Delete this trusted client and all of its tunnel definitions?', action: () => remove(client.id) })}><Trash2 size={15} /></IconButton>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!clients.length && <div className="empty-row">No trusted clients</div>}
      </section>
      {editing !== undefined && (
        <ClientRemarkEditor
          client={editing ?? undefined}
          onClose={() => setEditing(undefined)}
          onSaved={() => {
            setEditing(undefined)
            void load()
          }}
        />
      )}
      {confirmation && <Confirmation request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}

interface TunnelDraft {
  protocol: 'http' | 'tcp' | 'udp'
  hostname: string
  serverPort: string
  localHost: string
  localPort: string
  enabled: boolean
}

function TunnelEditor({ clientId, initial, onClose, onSaved }: { clientId: string, initial?: TunnelView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [draft, setDraft] = useState<TunnelDraft>({
    protocol: initial?.protocol ?? 'http',
    hostname: initial?.hostname ?? '',
    serverPort: initial?.serverPort?.toString() ?? '',
    localHost: initial?.localHost ?? '127.0.0.1',
    localPort: initial?.localPort.toString() ?? '',
    enabled: initial?.enabled ?? true,
  })
  const [error, setError] = useState('')
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    const body = {
      protocol: draft.protocol,
      hostname: draft.protocol === 'http' ? draft.hostname : null,
      serverPort: draft.protocol === 'http' || !draft.serverPort ? null : Number(draft.serverPort),
      localHost: draft.localHost,
      localPort: Number(draft.localPort),
      enabled: draft.enabled,
    }
    try {
      await apiJson(initial ? `/api/tunnels/${encodeURIComponent(initial.id)}` : `/api/clients/${encodeURIComponent(clientId)}/tunnels`, jsonRequest(initial ? 'PATCH' : 'POST', body))
      onSaved()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
  }
  return (
    <div className="modal-backdrop" role="presentation">
      <form className="modal" onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2>{initial ? 'Edit tunnel' : 'New tunnel'}</h2>
          <IconButton label="Close" onClick={onClose}><X size={16} /></IconButton>
        </div>
        <div className="segmented">{(['http', 'tcp', 'udp'] as const).map(value => <button type="button" className={draft.protocol === value ? 'active' : ''} key={value} onClick={() => setDraft({ ...draft, protocol: value })}>{value.toUpperCase()}</button>)}</div>
        {draft.protocol === 'http'
          ? (
              <label>
                Public hostname
                <input value={draft.hostname} onChange={event => setDraft({ ...draft, hostname: event.target.value })} required />
              </label>
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
            <input value={draft.localHost} onChange={event => setDraft({ ...draft, localHost: event.target.value })} required />
          </label>
          <label>
            Local port
            <input type="number" min="1" max="65535" value={draft.localPort} onChange={event => setDraft({ ...draft, localPort: event.target.value })} required />
          </label>
        </div>
        <label className="checkbox">
          <input type="checkbox" checked={draft.enabled} onChange={event => setDraft({ ...draft, enabled: event.target.checked })} />
          Enabled
        </label>
        {error && <p className="form-error">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit">Save</button>
        </div>
      </form>
    </div>
  )
}

export function ClientDetailPage({ id, refreshSequence, showOwner }: { id: string, refreshSequence: number, showOwner: boolean }): React.JSX.Element {
  const [client, setClient] = useState<ClientView>()
  const [tunnels, setTunnels] = useState<TunnelView[]>([])
  const [editing, setEditing] = useState<TunnelView | null | undefined>()
  const [error, setError] = useState('')
  const [confirmation, setConfirmation] = useState<ConfirmationRequest>()
  const load = useCallback(async () => {
    try {
      const body = await apiJson<{ client: ClientView, tunnels: TunnelView[] }>(`/api/clients/${encodeURIComponent(id)}`)
      setClient(body.client)
      setTunnels(body.tunnels)
      setError('')
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
  }, [id])
  useEffect(() => void load(), [load, refreshSequence])
  const mutate = async (action: () => Promise<void>): Promise<void> => {
    try {
      await action()
      setError('')
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
  }
  const toggle = async (tunnel: TunnelView): Promise<void> => {
    await mutate(async () => {
      await apiJson(`/api/tunnels/${encodeURIComponent(tunnel.id)}`, jsonRequest('PATCH', { enabled: !tunnel.enabled }))
      await load()
    })
  }
  const remove = async (tunnel: TunnelView): Promise<void> => {
    await apiJson(`/api/tunnels/${encodeURIComponent(tunnel.id)}`, { method: 'DELETE' })
    await load()
  }
  const restart = async (): Promise<void> => {
    await mutate(async () => {
      await apiJson(`/api/clients/${encodeURIComponent(id)}/restart`, { method: 'POST' })
    })
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
            <button type="button" onClick={() => void restart()}>
              <RefreshCw size={15} />
              Restart frpc
            </button>
            <button className="primary" type="button" onClick={() => setEditing(null)}>
              <Plus size={15} />
              New tunnel
            </button>
          </>
        )}
      />
      {error && <p className="runtime-error">{error}</p>}
      {client && (
        <section className="client-strip">
          <Token value={client.token} />
          <Status value={client.runtime.connectionState} />
          <Status value={client.runtime.processState} />
          {showOwner && (
            <span className="mono">
              Owner
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
      {client?.runtime.lastError && <p className="runtime-error">{client.runtime.lastError.message}</p>}
      <section className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Public mapping</th>
              <th>Local Endpoint</th>
              <th>Status</th>
              <th>Enabled</th>
              <th aria-label="Actions" />
            </tr>
          </thead>
          <tbody>
            {tunnels.map(tunnel => (
              <tr key={tunnel.id}>
                <td>
                  <strong>{tunnel.protocol.toUpperCase()}</strong>
                  {' '}
                  <span className="mono">{tunnel.hostname ?? tunnel.serverPort}</span>
                </td>
                <td className="mono">
                  {tunnel.localHost}
                  :
                  {tunnel.localPort}
                </td>
                <td><Status value={tunnel.state} /></td>
                <td><input aria-label={`Enable ${tunnel.id}`} type="checkbox" checked={tunnel.enabled} onChange={() => void toggle(tunnel)} /></td>
                <td>
                  <div className="row-actions">
                    <IconButton label="Edit tunnel" onClick={() => setEditing(tunnel)}><Pencil size={15} /></IconButton>
                    <IconButton label="Delete tunnel" onClick={() => setConfirmation({ message: 'Delete this Tunnel Definition?', action: () => remove(tunnel) })}><Trash2 size={15} /></IconButton>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!tunnels.length && <div className="empty-row">No tunnel definitions</div>}
      </section>
      {editing !== undefined && (
        <TunnelEditor
          clientId={id}
          initial={editing ?? undefined}
          onClose={() => setEditing(undefined)}
          onSaved={() => {
            setEditing(undefined)
            void load()
          }}
        />
      )}
      {confirmation && <Confirmation request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}
