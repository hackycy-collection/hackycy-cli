import type { FormEvent, ReactNode } from 'react'
import { Activity, Gauge, LogOut, Network, Play, RefreshCw, Server, Square, Users } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { ApiError, apiJson, jsonRequest } from './api'
import { ClientDetailPage, ClientsPage } from './client-pages'
import { IconButton, navigate, PageHeader, Status } from './ui'

interface StateView {
  frps: { state: string, pid?: number, error?: { message: string } }
  counts: { clients: number, connected: number, tunnels: number, pending: number, errors: number }
  settings: {
    address: string
    controlPort: number
    frpPort: number
    httpPort: number
    portRange: { start: number, end: number }
    advertiseFrpAddress: { host: string, port: number } | null
    dataDir: string
    adminUser: string
  }
}

type Page = { name: 'overview' | 'clients' | 'server' } | { name: 'client', id: string }

function currentPage(): Page {
  const client = /^\/clients\/([^/]+)$/.exec(location.pathname)?.[1]
  if (client)
    return { name: 'client', id: decodeURIComponent(client) }
  if (location.pathname === '/clients')
    return { name: 'clients' }
  if (location.pathname === '/server')
    return { name: 'server' }
  return { name: 'overview' }
}

function Login({ onLogin }: { onLogin: () => void }): React.JSX.Element {
  const [error, setError] = useState('')
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    try {
      await apiJson('/api/session', jsonRequest('POST', { username: form.get('username'), password: form.get('password') }))
      onLogin()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
  }
  return (
    <main className="login-shell">
      <form className="login-panel" onSubmit={event => void submit(event)}>
        <div className="brand">
          <Network size={18} />
          <span>HACKYCY TUNNEL</span>
        </div>
        <h1>Control Plane</h1>
        <label>
          Username
          <input name="username" autoComplete="username" required autoFocus />
        </label>
        <label>
          Password
          <input name="password" type="password" autoComplete="current-password" required />
        </label>
        {error && <p className="form-error">{error}</p>}
        <button className="primary" type="submit">Sign in</button>
      </form>
    </main>
  )
}

function Layout({ page, children, onLogout }: { page: Page, children: ReactNode, onLogout: () => void }): React.JSX.Element {
  const item = (name: Page['name'], path: string, icon: ReactNode, label: string): React.JSX.Element => (
    <button type="button" className={page.name === name ? 'nav-item active' : 'nav-item'} onClick={() => navigate(path)}>
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
          {item('server', '/server', <Server size={16} />, 'Server')}
        </nav>
        <button type="button" className="nav-item logout" onClick={onLogout}>
          <LogOut size={16} />
          <span>Sign out</span>
        </button>
      </aside>
      <div className="content-shell">{children}</div>
    </div>
  )
}

function Overview({ state }: { state: StateView }): React.JSX.Element {
  const metrics = [
    ['Trusted clients', state.counts.clients],
    ['Connected', state.counts.connected],
    ['Tunnel definitions', state.counts.tunnels],
    ['Pending', state.counts.pending],
    ['Errors', state.counts.errors],
  ] as const
  return (
    <>
      <PageHeader title="Overview" />
      <section className="metric-grid">
        {metrics.map(([label, value]) => (
          <div className="metric" key={label}>
            <span>{label}</span>
            <strong>{value}</strong>
          </div>
        ))}
      </section>
      <section className="section-band">
        <div className="section-title">
          <h2>Runtime</h2>
          <Status value={state.frps.state} />
        </div>
        <dl className="detail-grid">
          <dt>frps process</dt>
          <dd>{state.frps.pid ? `PID ${state.frps.pid}` : 'No active process'}</dd>
          <dt>Control listener</dt>
          <dd>
            {state.settings.address}
            :
            {state.settings.controlPort}
          </dd>
          <dt>FRP bind</dt>
          <dd>
            {state.settings.address}
            :
            {state.settings.frpPort}
          </dd>
          <dt>HTTP vhost</dt>
          <dd>
            {state.settings.address}
            :
            {state.settings.httpPort}
          </dd>
        </dl>
        {state.frps.error && <p className="runtime-error">{state.frps.error.message}</p>}
      </section>
    </>
  )
}

function ServerView({ state, reload }: { state: StateView, reload: () => void }): React.JSX.Element {
  const [error, setError] = useState('')
  const action = async (value: 'start' | 'stop' | 'restart'): Promise<void> => {
    try {
      await apiJson(`/api/server/frp/${value}`, { method: 'POST' })
      setError('')
      reload()
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    }
  }
  return (
    <>
      <PageHeader
        title="Tunnel Server"
        actions={(
          <>
            <IconButton label="Start frps" onClick={() => void action('start')}><Play size={15} /></IconButton>
            <IconButton label="Stop frps" onClick={() => void action('stop')}><Square size={14} /></IconButton>
            <IconButton label="Restart frps" onClick={() => void action('restart')}><RefreshCw size={15} /></IconButton>
          </>
        )}
      />
      {error && <p className="runtime-error">{error}</p>}
      <section className="section-band">
        <div className="section-title">
          <h2>frps</h2>
          <Status value={state.frps.state} />
        </div>
        <dl className="detail-grid">
          <dt>Process</dt>
          <dd>{state.frps.pid ? `PID ${state.frps.pid}` : 'Stopped'}</dd>
        </dl>
      </section>
      <section className="section-band">
        <div className="section-title"><h2>Deployment settings</h2></div>
        <dl className="detail-grid">
          <dt>Control listener</dt>
          <dd>
            {state.settings.address}
            :
            {state.settings.controlPort}
          </dd>
          <dt>FRP bind</dt>
          <dd>
            {state.settings.address}
            :
            {state.settings.frpPort}
          </dd>
          <dt>HTTP vhost</dt>
          <dd>
            {state.settings.address}
            :
            {state.settings.httpPort}
          </dd>
          <dt>Server Port Pool</dt>
          <dd>
            {state.settings.portRange.start}
            -
            {state.settings.portRange.end}
            {' '}
            TCP/UDP
          </dd>
          <dt>Advertised FRP</dt>
          <dd>{state.settings.advertiseFrpAddress ? `${state.settings.advertiseFrpAddress.host}:${state.settings.advertiseFrpAddress.port}` : 'Derived from agent request'}</dd>
          <dt>Data directory</dt>
          <dd className="mono break">{state.settings.dataDir}</dd>
          <dt>Administrator</dt>
          <dd>{state.settings.adminUser}</dd>
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
  const load = useCallback(async () => {
    try {
      setState(await apiJson<StateView>('/api/state'))
      setAuthenticated(true)
    }
    catch (cause) {
      if (cause instanceof ApiError && cause.status === 401)
        setAuthenticated(false)
    }
  }, [])
  useEffect(() => {
    const changed = (): void => setPage(currentPage())
    window.addEventListener('popstate', changed)
    return () => window.removeEventListener('popstate', changed)
  }, [])
  useEffect(() => void load(), [load])
  useEffect(() => {
    if (!authenticated)
      return
    const events = new EventSource('/api/events')
    events.onmessage = () => {
      setRefreshSequence(value => value + 1)
      void load()
    }
    return () => events.close()
  }, [authenticated, load])
  const logout = async (): Promise<void> => {
    await apiJson('/api/session', { method: 'DELETE' })
    setAuthenticated(false)
  }
  if (authenticated === undefined) {
    return (
      <div className="boot">
        <Activity size={18} />
        Loading
      </div>
    )
  }
  if (!authenticated)
    return <Login onLogin={() => void load()} />
  if (!state) {
    return (
      <div className="boot">
        <Activity size={18} />
        Loading
      </div>
    )
  }
  return (
    <Layout page={page} onLogout={() => void logout()}>
      {page.name === 'overview' && <Overview state={state} />}
      {page.name === 'clients' && <ClientsPage refreshSequence={refreshSequence} />}
      {page.name === 'client' && <ClientDetailPage id={page.id} refreshSequence={refreshSequence} />}
      {page.name === 'server' && <ServerView state={state} reload={() => void load()} />}
    </Layout>
  )
}
