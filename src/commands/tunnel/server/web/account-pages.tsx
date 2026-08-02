import type { FormEvent } from 'react'
import type { AccountRole } from '../../types'
import type { AccountView } from './api'
import type { ConfirmationRequest } from './ui'
import { KeyRound, Pencil, Plus, Trash2, X } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { apiJson, jsonRequest } from './api'
import { Confirmation, IconButton, PageHeader, Status } from './ui'

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
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    setSaving(true)
    try {
      if (account)
        await apiJson(`/api/accounts/${encodeURIComponent(account.id)}`, jsonRequest('PATCH', { role }))
      else
        await apiJson('/api/accounts', jsonRequest('POST', { username, password, role }))
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
              <input value={password} type="password" minLength={8} maxLength={256} autoComplete="new-password" required onChange={event => setPassword(event.target.value)} />
            </label>
          </>
        )}
        <label>
          Role
          <RoleSelector value={role} onChange={setRole} />
        </label>
        {error && <p className="form-error">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>{saving ? 'Saving...' : account ? 'Save' : 'Create'}</button>
        </div>
      </form>
    </div>
  )
}

function PasswordReset({ account, onClose, onSaved }: { account: AccountView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault()
    setSaving(true)
    try {
      await apiJson(`/api/accounts/${encodeURIComponent(account.id)}/password`, jsonRequest('PUT', { password }))
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
          <input value={password} type="password" minLength={8} maxLength={256} autoComplete="new-password" required autoFocus onChange={event => setPassword(event.target.value)} />
        </label>
        {error && <p className="form-error">{error}</p>}
        <div className="modal-actions">
          <button type="button" onClick={onClose}>Cancel</button>
          <button className="primary" type="submit" disabled={saving}>{saving ? 'Saving...' : 'Reset'}</button>
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
  const load = useCallback(async () => {
    try {
      setAccounts((await apiJson<{ accounts: AccountView[] }>('/api/accounts')).accounts)
      setError('')
    }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
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
          <button className="primary" type="button" onClick={() => setEditing(null)}>
            <Plus size={15} />
            Create account
          </button>
        )}
      />
      {error && <p className="runtime-error">{error}</p>}
      <section className="table-wrap">
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
                    <IconButton label={account.clientCount ? 'Delete owned clients first' : 'Delete account'} disabled={account.managedByEnvironment || account.clientCount > 0} onClick={() => setConfirmation({ message: `Delete account ${account.username}?`, action: () => remove(account) })}><Trash2 size={15} /></IconButton>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!accounts.length && <div className="empty-row">No accounts</div>}
      </section>
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
