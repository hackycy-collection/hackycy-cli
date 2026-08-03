import type { FormEvent, ReactNode } from 'react'
import type { CurrentAccount } from './api'
import { Gauge, KeyRound, LogOut, Network, Play, RefreshCw, Server, Shield, Square, Users, X } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { AccountsPage } from './account-pages'
import { ApiError, apiJson, jsonRequest } from './api'
import { ClientDetailPage, ClientsPage } from './client-pages'
import { ErrorState, IconButton, navigate, PageHeader, Spinner, Status, useFeedback } from './ui'

interface ServerProjection {
  frps: { state: string, pid?: number, error?: { message: string } }
  settings: {
    address: string
    controlPort: number
    frpPort: number
    httpPort: number
    portRange: { start: number, end: number }
    advertiseFrpAddress?: { host: string, port: number }
    dataDir: string
    adminUser: string
  }
}

interface StateView {
  account: CurrentAccount
  counts: { clients: number, connected: number, tunnels: number, pending: number, errors: number }
  server?: ServerProjection
}

type Page = { name: 'overview' | 'clients' | 'accounts' | 'server' } | { name: 'client', id: string }

function currentPage(): Page {
  const client = /^\/clients\/([^/]+)$/.exec(location.pathname)?.[1]
  if (client)
    return { name: 'client', id: decodeURIComponent(client) }
  if (location.pathname === '/clients')
    return { name: 'clients' }
  if (location.pathname === '/accounts')
    return { name: 'accounts' }
  if (location.pathname === '/server')
    return { name: 'server' }
  return { name: 'overview' }
}

function Login({ onLogin }: { onLogin: () => void }): React.JSX.Element {
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const { notify } = useFeedback()
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    setError('')
    setSubmitting(true)
    try {
      await apiJson('/api/session', jsonRequest('POST', { username: form.get('username'), password: form.get('password') }))
      notify('Signed in')
      onLogin()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setSubmitting(false)
    }
  }
  return (
    <main className="login-shell">
      <form className="login-panel" aria-busy={submitting} onSubmit={event => void submit(event)}>
        <div className="brand">
          <Network size={18} />
          <span>HACKYCY TUNNEL</span>
        </div>
        <label>
          Username
          <input name="username" autoComplete="username" required autoFocus disabled={submitting} />
        </label>
        <label>
          Password
          <input name="password" type="password" autoComplete="current-password" required disabled={submitting} />
        </label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <button className="primary" type="submit" disabled={submitting}>
          {submitting && <Spinner />}
          {submitting ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
    </main>
  )
}

function PasswordEditor({ onClose, onChanged }: { onClose: () => void, onChanged: () => void }): React.JSX.Element {
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const { notify } = useFeedback()
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const newPassword = String(form.get('newPassword'))
    if (newPassword !== form.get('confirmation')) {
      setError('New passwords do not match')
      return
    }
    setSaving(true)
    try {
      await apiJson('/api/session/password', jsonRequest('PUT', { currentPassword: form.get('currentPassword'), newPassword }))
      notify('Password changed')
      onChanged()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setSaving(false)
    }
  }
  return (
    <div className="modal-backdrop" role="presentation">
      <form className="modal" role="dialog" aria-modal="true" aria-labelledby="password-title" onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2 id="password-title">Change password</h2>
          <IconButton label="Close" onClick={onClose}><X size={16} /></IconButton>
        </div>
        <label>
          Current password
          <input name="currentPassword" type="password" autoComplete="current-password" required autoFocus />
        </label>
        <label>
          New password
          <input name="newPassword" type="password" minLength={5} maxLength={256} autoComplete="new-password" required />
        </label>
        <label>
          Confirm password
          <input name="confirmation" type="password" minLength={5} maxLength={256} autoComplete="new-password" required />
        </label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>
            {saving && <Spinner />}
            {saving ? 'Saving...' : 'Change'}
          </button>
        </div>
      </form>
    </div>
  )
}

function Layout({ page, account, children, loggingOut, onLogout, onChangePassword }: { page: Page, account: CurrentAccount, children: ReactNode, loggingOut: boolean, onLogout: () => void, onChangePassword: () => void }): React.JSX.Element {
  const item = (name: Page['name'], path: string, icon: ReactNode, label: string): React.JSX.Element => (
    <button
      type="button"
      className={page.name === name ? 'nav-item active' : 'nav-item'}
      aria-label={label}
      aria-current={page.name === name ? 'page' : undefined}
      title={label}
      onClick={() => navigate(path)}
    >
      {icon}
      <span>{label}</span>
    </button>
  )
  return (
    <div className="app-shell">
      <aside>
        <div className="brand">
          <Network size={17} />
          <span>HACKYCY TUNNEL</span>
        </div>
        <nav>
          {item('overview', '/', <Gauge size={16} />, 'Overview')}
          {item('clients', '/clients', <Users size={16} />, 'Clients')}
          {account.role === 'admin' && item('accounts', '/accounts', <Shield size={16} />, 'Accounts')}
          {account.role === 'admin' && item('server', '/server', <Server size={16} />, 'Server')}
        </nav>
        <div className="account-block">
          <div className="account-meta">
            <strong>{account.username}</strong>
            <span>{account.role}</span>
          </div>
          {!account.managedByEnvironment && <IconButton label="Change password" onClick={onChangePassword}><KeyRound size={15} /></IconButton>}
          <IconButton label="Sign out" loading={loggingOut} onClick={onLogout}><LogOut size={15} /></IconButton>
        </div>
      </aside>
      <div className="content-shell">{children}</div>
    </div>
  )
}

function Overview({ state, refreshing, reload }: { state: StateView, refreshing: boolean, reload: () => void }): React.JSX.Element {
  const metrics = [['Trusted clients', state.counts.clients], ['Connected', state.counts.connected], ['Tunnel definitions', state.counts.tunnels], ['Pending', state.counts.pending], ['Errors', state.counts.errors]] as const
  return (
    <>
      <PageHeader title="Overview" actions={<IconButton label="Refresh overview" loading={refreshing} onClick={reload}><RefreshCw size={15} /></IconButton>} />
      <section className="metric-grid">
        {metrics.map(([label, value]) => (
          <div className="metric" key={label}>
            <span>{label}</span>
            <strong>{value}</strong>
          </div>
        ))}
      </section>
      {state.server && (
        <section className="section-band">
          <div className="section-title">
            <h2>Runtime</h2>
            <Status value={state.server.frps.state} />
          </div>
          <dl className="detail-grid">
            <dt>frps process</dt>
            <dd>{state.server.frps.pid ? `PID ${state.server.frps.pid}` : 'No active process'}</dd>
            <dt>Control listener</dt>
            <dd>
              {state.server.settings.address}
              :
              {state.server.settings.controlPort}
            </dd>
            <dt>FRP bind</dt>
            <dd>
              {state.server.settings.address}
              :
              {state.server.settings.frpPort}
            </dd>
            <dt>HTTP vhost</dt>
            <dd>
              {state.server.settings.address}
              :
              {state.server.settings.httpPort}
            </dd>
          </dl>
          {state.server.frps.error && <p className="runtime-error">{state.server.frps.error.message}</p>}
        </section>
      )}
    </>
  )
}

function ServerView({ server, reload }: { server: ServerProjection, reload: () => Promise<void> }): React.JSX.Element {
  const [error, setError] = useState('')
  const [pending, setPending] = useState<'start' | 'stop' | 'restart'>()
  const { notify } = useFeedback()
  const action = async (value: 'start' | 'stop' | 'restart'): Promise<void> => {
    setPending(value)
    try {
      await apiJson(`/api/server/frp/${value}`, { method: 'POST' })
      setError('')
      notify(`frps ${value} completed`)
      await reload()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setPending(undefined)
    }
  }
  return (
    <>
      <PageHeader
        title="Tunnel Server"
        actions={(
          <>
            <IconButton label="Start frps" loading={pending === 'start'} disabled={Boolean(pending)} onClick={() => void action('start')}><Play size={15} /></IconButton>
            <IconButton label="Stop frps" loading={pending === 'stop'} disabled={Boolean(pending)} onClick={() => void action('stop')}><Square size={14} /></IconButton>
            <IconButton label="Restart frps" loading={pending === 'restart'} disabled={Boolean(pending)} onClick={() => void action('restart')}><RefreshCw size={15} /></IconButton>
          </>
        )}
      />
      {error && <p className="runtime-error" role="alert">{error}</p>}
      <section className="section-band">
        <div className="section-title">
          <h2>frps</h2>
          <Status value={server.frps.state} />
        </div>
        <dl className="detail-grid">
          <dt>Process</dt>
          <dd>{server.frps.pid ? `PID ${server.frps.pid}` : 'Stopped'}</dd>
        </dl>
      </section>
      <section className="section-band">
        <div className="section-title"><h2>Deployment settings</h2></div>
        <dl className="detail-grid">
          <dt>Control listener</dt>
          <dd>
            {server.settings.address}
            :
            {server.settings.controlPort}
          </dd>
          <dt>FRP bind</dt>
          <dd>
            {server.settings.address}
            :
            {server.settings.frpPort}
          </dd>
          <dt>HTTP vhost</dt>
          <dd>
            {server.settings.address}
            :
            {server.settings.httpPort}
          </dd>
          <dt>Server Port Pool</dt>
          <dd>
            {server.settings.portRange.start}
            -
            {server.settings.portRange.end}
            {' '}
            TCP/UDP
          </dd>
          <dt>Advertised FRP</dt>
          <dd>{server.settings.advertiseFrpAddress ? `${server.settings.advertiseFrpAddress.host}:${server.settings.advertiseFrpAddress.port}` : 'Derived from agent request'}</dd>
          <dt>Data directory</dt>
          <dd className="mono break">{server.settings.dataDir}</dd>
          <dt>Deployment Administrator</dt>
          <dd>{server.settings.adminUser}</dd>
        </dl>
      </section>
    </>
  )
}

export function App(): React.JSX.Element {
  const [authenticated, setAuthenticated] = useState<boolean>()
  const [page, setPage] = useState<Page>(currentPage)
  const [state, setState] = useState<StateView>()
  const [refreshSequence, setRefreshSequence] = useState(0)
  const [changingPassword, setChangingPassword] = useState(false)
  const [stateLoading, setStateLoading] = useState(true)
  const [stateError, setStateError] = useState('')
  const [eventsConnected, setEventsConnected] = useState<boolean>()
  const [loggingOut, setLoggingOut] = useState(false)
  const { notify } = useFeedback()
  const sessionEnded = useCallback(() => {
    setAuthenticated(false)
    setState(undefined)
    setChangingPassword(false)
    setStateError('')
    setEventsConnected(undefined)
  }, [])
  const load = useCallback(async () => {
    setStateLoading(true)
    try {
      setState(await apiJson<StateView>('/api/state'))
      setAuthenticated(true)
      setStateError('')
    }
    catch (cause) {
      if (cause instanceof ApiError && cause.status === 401)
        sessionEnded()
      else
        setStateError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      setStateLoading(false)
    }
  }, [sessionEnded])
  useEffect(() => {
    const changed = (): void => setPage(currentPage())
    window.addEventListener('popstate', changed)
    window.addEventListener('tunnel-authentication-required', sessionEnded)
    return () => {
      window.removeEventListener('popstate', changed)
      window.removeEventListener('tunnel-authentication-required', sessionEnded)
    }
  }, [sessionEnded])
  useEffect(() => void load(), [load])
  useEffect(() => {
    if (state && state.account.role !== 'admin' && ['accounts', 'server'].includes(page.name))
      navigate('/')
  }, [page.name, state])
  useEffect(() => {
    if (!authenticated)
      return
    const events = new EventSource('/api/events')
    events.onopen = () => setEventsConnected(true)
    events.onerror = () => setEventsConnected(false)
    events.onmessage = (message) => {
      const event = JSON.parse(message.data) as { event: 'changed' | 'session_revoked' }
      if (event.event === 'session_revoked') {
        sessionEnded()
        return
      }
      setRefreshSequence(value => value + 1)
      void load()
    }
    return () => events.close()
  }, [authenticated, load, sessionEnded])
  const logout = async (): Promise<void> => {
    setLoggingOut(true)
    try {
      await apiJson('/api/session', { method: 'DELETE' })
      sessionEnded()
    }
    catch (cause) {
      notify(cause instanceof Error ? cause.message : String(cause), 'error')
    }
    finally {
      setLoggingOut(false)
    }
  }
  if (authenticated === undefined) {
    if (stateError) {
      return (
        <div className="boot boot-error">
          <ErrorState message={stateError} retrying={stateLoading} onRetry={() => void load()} />
        </div>
      )
    }
    return (
      <div className="boot">
        <Spinner size={18} />
        Loading session
      </div>
    )
  }
  if (!authenticated)
    return <Login onLogin={() => void load()} />
  if (!state) {
    return (
      <div className="boot">
        <Spinner size={18} />
        Loading workspace
      </div>
    )
  }
  return (
    <Layout page={page} account={state.account} loggingOut={loggingOut} onLogout={() => void logout()} onChangePassword={() => setChangingPassword(true)}>
      {eventsConnected === false && (
        <div className="sync-banner" role="status">
          <Spinner />
          Live updates disconnected. Reconnecting...
        </div>
      )}
      {stateError && <ErrorState message={stateError} retrying={stateLoading} onRetry={() => void load()} />}
      {page.name === 'overview' && <Overview state={state} refreshing={stateLoading} reload={() => void load()} />}
      {page.name === 'clients' && <ClientsPage refreshSequence={refreshSequence} showOwner={state.account.role === 'admin'} />}
      {page.name === 'client' && <ClientDetailPage id={page.id} refreshSequence={refreshSequence} showOwner={state.account.role === 'admin'} />}
      {page.name === 'accounts' && state.account.role === 'admin' && <AccountsPage currentAccountId={state.account.id} refreshSequence={refreshSequence} onSessionEnded={sessionEnded} />}
      {page.name === 'server' && state.server && <ServerView server={state.server} reload={load} />}
      {changingPassword && <PasswordEditor onClose={() => setChangingPassword(false)} onChanged={sessionEnded} />}
    </Layout>
  )
}
