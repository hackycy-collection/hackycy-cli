import type { FormEvent } from 'react'
import type { AccountRole } from '../../types'
import type { AccountView } from './api'
import type { ConfirmationRequest } from './ui'
import { KeyRound, Pencil, Plus, RefreshCw, Trash2, X } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { apiJson, jsonRequest } from './api'
import { Confirmation, ErrorState, IconButton, LoadingState, PageHeader, Spinner, Status, useFeedback } from './ui'

function RoleSelector({ value, onChange }: { value: AccountRole, onChange: (role: AccountRole) => void }): React.JSX.Element {
  return (
    <div className="segmented roles">
      {(['user', 'admin'] as const).map(role => (
        <button type="button" className={value === role ? 'active' : ''} key={role} onClick={() => onChange(role)}>
          {role === 'admin' ? 'Administrator' : 'User'}
        </button>
      ))}
    </div>
  )
}

function AccountEditor({ account, onClose, onSaved }: { account?: AccountView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<AccountRole>(account?.role ?? 'user')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const { notify } = useFeedback()
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    setSaving(true)
    try {
      if (account)
        await apiJson(`/api/accounts/${encodeURIComponent(account.id)}`, jsonRequest('PATCH', { role }))
      else
        await apiJson('/api/accounts', jsonRequest('POST', { username, password, role }))
      notify(account ? 'Account role saved' : 'Account created')
      onSaved()
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
      <form className="modal" role="dialog" aria-modal="true" aria-labelledby="account-editor-title" onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2 id="account-editor-title">{account ? 'Change account role' : 'Create account'}</h2>
          <IconButton label="Close" onClick={onClose}><X size={16} /></IconButton>
        </div>
        {!account && (
          <>
            <label>
              Username
              <input value={username} maxLength={64} autoComplete="off" required autoFocus onChange={event => setUsername(event.target.value)} />
            </label>
            <label>
              Password
              <input value={password} type="password" minLength={5} maxLength={256} autoComplete="new-password" required onChange={event => setPassword(event.target.value)} />
            </label>
          </>
        )}
        <label>
          Role
          <RoleSelector value={role} onChange={setRole} />
        </label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>
            {saving && <Spinner />}
            {saving ? 'Saving...' : account ? 'Save' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  )
}

function PasswordReset({ account, onClose, onSaved }: { account: AccountView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const { notify } = useFeedback()
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    setSaving(true)
    try {
      await apiJson(`/api/accounts/${encodeURIComponent(account.id)}/password`, jsonRequest('PUT', { password }))
      notify('Account password reset')
      onSaved()
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
      <form className="modal" role="dialog" aria-modal="true" aria-labelledby="password-reset-title" onSubmit={event => void submit(event)}>
        <div className="modal-title">
          <h2 id="password-reset-title">Reset password</h2>
          <IconButton label="Close" onClick={onClose}><X size={16} /></IconButton>
        </div>
        <label>
          New password
          <input value={password} type="password" minLength={5} maxLength={256} autoComplete="new-password" required autoFocus onChange={event => setPassword(event.target.value)} />
        </label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>
            {saving && <Spinner />}
            {saving ? 'Saving...' : 'Reset'}
          </button>
        </div>
      </form>
    </div>
  )
}

export function AccountsPage({ currentAccountId, refreshSequence, onSessionEnded }: { currentAccountId: string, refreshSequence: number, onSessionEnded: () => void }): React.JSX.Element {
  const [accounts, setAccounts] = useState<AccountView[]>([])
  const [editing, setEditing] = useState<AccountView | null | undefined>()
  const [resetting, setResetting] = useState<AccountView>()
  const [confirmation, setConfirmation] = useState<ConfirmationRequest>()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const loaded = useRef(false)
  const requestId = useRef(0)
  const load = useCallback(async () => {
    const currentRequest = ++requestId.current
    loaded.current ? setRefreshing(true) : setLoading(true)
    try {
      const result = (await apiJson<{ accounts: AccountView[] }>('/api/accounts')).accounts
      if (currentRequest !== requestId.current)
        return
      setAccounts(result)
      setError('')
      loaded.current = true
    }
    catch (cause) {
      if (currentRequest === requestId.current)
        setError(cause instanceof Error ? cause.message : String(cause))
    }
    finally {
      if (currentRequest === requestId.current) {
        setLoading(false)
        setRefreshing(false)
      }
    }
  }, [])
  useEffect(() => void load(), [load, refreshSequence])
  const remove = async (account: AccountView): Promise<void> => {
    await apiJson(`/api/accounts/${encodeURIComponent(account.id)}`, { method: 'DELETE' })
    if (account.id === currentAccountId)
      onSessionEnded()
    else
      await load()
  }
  return (
    <>
      <PageHeader
        title="Control Plane Accounts"
        actions={(
          <>
            <IconButton label="Refresh accounts" loading={refreshing} onClick={() => void load()}><RefreshCw size={15} /></IconButton>
            <button className="primary" type="button" onClick={() => setEditing(null)}>
              <Plus size={15} />
              Create account
            </button>
          </>
        )}
      />
      {loading && !loaded.current
        ? <LoadingState label="Loading accounts" />
        : error && !loaded.current
          ? <ErrorState message={error} retrying={loading} onRetry={() => void load()} />
          : (
              <>
                {error && <ErrorState message={error} retrying={refreshing} onRetry={() => void load()} />}
                <section className="table-wrap" aria-busy={refreshing}>
                  <table>
                    <thead>
                      <tr>
                        <th>Username</th>
                        <th>Role</th>
                        <th>Source</th>
                        <th>Clients</th>
                        <th aria-label="Actions" />
                      </tr>
                    </thead>
                    <tbody>
                      {accounts.map(account => (
                        <tr key={account.id}>
                          <td><strong>{account.username}</strong></td>
                          <td><Status value={account.role} /></td>
                          <td>{account.managedByEnvironment ? 'Environment' : 'Local'}</td>
                          <td>{account.clientCount}</td>
                          <td>
                            <div className="row-actions">
                              <IconButton label="Change role" disabled={account.managedByEnvironment} onClick={() => setEditing(account)}><Pencil size={15} /></IconButton>
                              <IconButton label="Reset password" disabled={account.managedByEnvironment} onClick={() => setResetting(account)}><KeyRound size={15} /></IconButton>
                              <IconButton label={account.clientCount ? 'Delete owned clients first' : 'Delete account'} disabled={account.managedByEnvironment || account.clientCount > 0} onClick={() => setConfirmation({ message: `Delete account ${account.username}?`, successMessage: 'Account deleted', action: () => remove(account) })}><Trash2 size={15} /></IconButton>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {!accounts.length && <div className="empty-row">No accounts</div>}
                </section>
              </>
            )}
      {editing !== undefined && (
        <AccountEditor
          account={editing ?? undefined}
          onClose={() => setEditing(undefined)}
          onSaved={() => {
            setEditing(undefined)
            if (editing?.id === currentAccountId)
              onSessionEnded()
            else
              void load()
          }}
        />
      )}
      {resetting && (
        <PasswordReset
          account={resetting}
          onClose={() => setResetting(undefined)}
          onSaved={() => {
            const self = resetting.id === currentAccountId
            setResetting(undefined)
            if (self)
              onSessionEnded()
            else
              void load()
          }}
        />
      )}
      {confirmation && <Confirmation request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}
