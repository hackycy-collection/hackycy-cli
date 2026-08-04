import type { Path, UseFormRegister } from 'react-hook-form'
import type { ClientView, TunnelView } from './api'
import type { TunnelEditorStep, TunnelFormValues } from './tunnel-form'
import type { ConfirmAction } from './ui'
import { zodResolver } from '@hookform/resolvers/zod'
import { ArrowLeft, ArrowRight, ChevronDown, ChevronRight, Pencil, Plus, RefreshCw, RotateCcw, Trash2 } from 'lucide-react'
import { Fragment, useCallback, useEffect, useRef, useState } from 'react'
import { Controller, useFieldArray, useForm, useWatch } from 'react-hook-form'
import { z } from 'zod'
import { apiJson, jsonRequest } from './api'
import { FormError, FormField, FormMessage } from './form'
import { DialogShell, FormScrollArea, SegmentedControl, Select, Tabs } from './primitives'
import { buildTunnelPayload, createTunnelSchema, draftToTunnelForm, stepForTunnelField, tunnelStepFields } from './tunnel-form'
import { ConfirmationDialog, ErrorState, IconButton, LoadingState, navigate, PageHeader, Spinner, Status, Switch, Token, useFeedback } from './ui'

const clientRemarkSchema = z.object({ remark: z.string().max(100, 'Client remark must be 100 characters or fewer') })

type ClientRemarkValues = z.infer<typeof clientRemarkSchema>

function message(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause)
}

function ClientRemarkEditor({ client, onClose, onSaved }: { client?: ClientView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const { notify } = useFeedback()
  const form = useForm<ClientRemarkValues>({ resolver: zodResolver(clientRemarkSchema), defaultValues: { remark: client?.remark ?? '' } })
  const saving = form.formState.isSubmitting
  const submit = form.handleSubmit(async ({ remark }) => {
    form.clearErrors('root.server')
    try {
      await apiJson(client ? `/api/clients/${encodeURIComponent(client.id)}` : '/api/clients', jsonRequest(client ? 'PATCH' : 'POST', { remark }))
      notify(client ? 'Client remark saved' : 'Trusted client created')
      onSaved()
    }
    catch (cause) {
      form.setError('root.server', { message: message(cause) })
    }
  })
  return (
    <DialogShell open title={client ? 'Edit Client Remark' : 'Create client'} busy={saving} onOpenChange={open => !open && onClose()} onSubmit={submit}>
      <FormField label="Client Remark" error={form.formState.errors.remark}>
        <textarea {...form.register('remark')} maxLength={100} autoFocus rows={4} disabled={saving} aria-invalid={Boolean(form.formState.errors.remark)} />
      </FormField>
      <FormError error={form.formState.errors.root?.server} />
      <div className="modal-actions">
        <button type="button" disabled={saving} onClick={onClose}>Cancel</button>
        <button className="primary" type="submit" disabled={saving}>
          {saving && <Spinner />}
          {saving ? 'Saving...' : client ? 'Save' : 'Create'}
        </button>
      </div>
    </DialogShell>
  )
}

export function ClientsPage({ refreshSequence, showOwner }: { refreshSequence: number, showOwner: boolean }): React.JSX.Element {
  const [clients, setClients] = useState<ClientView[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [confirmation, setConfirmation] = useState<ConfirmAction>()
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
      {confirmation && <ConfirmationDialog request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}

function ValueFieldRows({ label, name, fields, placeholder, minimum = 0, register, error, onAdd, onRemove }: {
  label: string
  name: 'customDomains'
  fields: Array<{ id: string }>
  placeholder?: string
  minimum?: number
  register: UseFormRegister<TunnelFormValues>
  error?: { message?: string }
  onAdd: () => void
  onRemove: (index: number) => void
}): React.JSX.Element {
  return (
    <fieldset className="structured-field">
      <legend>{label}</legend>
      <div className="structured-rows">
        {fields.map((field, index) => (
          <div className="value-field-row" key={field.id}>
            <input {...register(`${name}.${index}.value` as Path<TunnelFormValues>)} aria-label={`${label} ${index + 1}`} placeholder={placeholder} />
            <IconButton label={`Remove ${label} ${index + 1}`} disabled={fields.length <= minimum} onClick={() => onRemove(index)}><Trash2 size={14} /></IconButton>
          </div>
        ))}
      </div>
      <FormMessage error={error} />
      <button className="add-row" type="button" onClick={onAdd}>
        <Plus size={14} />
        Add
        {' '}
        {label.toLowerCase()}
      </button>
    </fieldset>
  )
}

function KeyValueFieldRows({ label, name, fields, register, error, onAdd, onRemove }: {
  label: string
  name: 'healthHeaders' | 'requestHeaders' | 'responseHeaders'
  fields: Array<{ id: string }>
  register: UseFormRegister<TunnelFormValues>
  error?: { message?: string }
  onAdd: () => void
  onRemove: (index: number) => void
}): React.JSX.Element {
  return (
    <fieldset className="structured-field">
      <legend>{label}</legend>
      <div className="structured-rows">
        {fields.map((field, index) => (
          <div className="key-value-field-row" key={field.id}>
            <input {...register(`${name}.${index}.name` as Path<TunnelFormValues>)} aria-label={`${label} name ${index + 1}`} placeholder="Header name" />
            <input {...register(`${name}.${index}.value` as Path<TunnelFormValues>)} aria-label={`${label} value ${index + 1}`} placeholder="Value" />
            <IconButton label={`Remove ${label} ${index + 1}`} onClick={() => onRemove(index)}><Trash2 size={14} /></IconButton>
          </div>
        ))}
      </div>
      <FormMessage error={error} />
      <button className="add-row" type="button" onClick={onAdd}>
        <Plus size={14} />
        Add header
      </button>
    </fieldset>
  )
}

function TunnelEditor({ clientId, initial, onClose, onSaved }: { clientId: string, initial?: TunnelView, onClose: () => void, onSaved: () => void }): React.JSX.Element {
  const form = useForm<TunnelFormValues>({
    resolver: zodResolver(createTunnelSchema({ hasExistingBasicAuth: Boolean(initial?.options.http?.basicAuth) })),
    defaultValues: draftToTunnelForm(initial),
    shouldUnregister: false,
  })
  const [activeStep, setActiveStep] = useState<TunnelEditorStep>('basics')
  const { notify } = useFeedback()
  const protocol = useWatch({ control: form.control, name: 'protocol' })
  const customDomains = useFieldArray({ control: form.control, name: 'customDomains' })
  const healthHeaders = useFieldArray({ control: form.control, name: 'healthHeaders' })
  const requestHeaders = useFieldArray({ control: form.control, name: 'requestHeaders' })
  const responseHeaders = useFieldArray({ control: form.control, name: 'responseHeaders' })
  const saving = form.formState.isSubmitting
  const steps: Array<{ id: TunnelEditorStep, label: string }> = [
    { id: 'basics', label: 'Basics' },
    { id: 'transport', label: 'Transport' },
    { id: 'health', label: 'Health check' },
    ...(protocol === 'http' ? [{ id: 'http' as const, label: 'HTTP' }] : []),
  ]
  const activeStepIndex = steps.findIndex(step => step.id === activeStep)
  const selectStep = (step: TunnelEditorStep): void => {
    form.clearErrors('root.server')
    setActiveStep(step)
  }
  const nextStep = async (): Promise<void> => {
    if (!await form.trigger(tunnelStepFields[activeStep]))
      return
    const next = steps[activeStepIndex + 1]
    if (next)
      selectStep(next.id)
  }
  const submit = form.handleSubmit(async (values) => {
    form.clearErrors('root.server')
    try {
      await apiJson(initial ? `/api/tunnels/${encodeURIComponent(initial.id)}` : `/api/clients/${encodeURIComponent(clientId)}/tunnels`, jsonRequest(initial ? 'PATCH' : 'POST', buildTunnelPayload(values)))
      notify(initial ? 'Tunnel Definition saved' : 'Tunnel Definition created')
      onSaved()
    }
    catch (cause) {
      form.setError('root.server', { message: message(cause) })
    }
  }, (errors) => {
    const field = Object.keys(errors)[0]
    if (field)
      setActiveStep(stepForTunnelField(field))
  })
  const focusNewField = (name: Path<TunnelFormValues>): void => {
    requestAnimationFrame(() => form.setFocus(name))
  }

  return (
    <DialogShell open title={initial ? 'Edit Tunnel Definition' : 'New Tunnel Definition'} className="tunnel-modal" busy={saving} onOpenChange={open => !open && onClose()} onSubmit={submit}>
      <Tabs.Root className="tunnel-tabs" value={activeStep} onValueChange={value => selectStep(value as TunnelEditorStep)}>
        <Tabs.List className="tunnel-steps" aria-label="Tunnel configuration steps">
          {steps.map((step, index) => (
            <Tabs.Trigger value={step.id} disabled={saving} key={step.id}>
              <span className="step-number">{index + 1}</span>
              {step.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>

        <Tabs.Content value="basics" forceMount className="tunnel-step-panel">
          <FormScrollArea>
            <div className="step-stack">
              <FormField label="Display name" error={form.formState.errors.label}>
                <input {...form.register('label')} maxLength={100} autoFocus aria-invalid={Boolean(form.formState.errors.label)} />
              </FormField>
              <Controller name="protocol" control={form.control} render={({ field }) => <SegmentedControl label="Tunnel protocol" value={field.value} disabled={saving} onChange={field.onChange} options={[{ value: 'http', label: 'HTTP' }, { value: 'tcp', label: 'TCP' }, { value: 'udp', label: 'UDP' }]} />} />
              {protocol === 'http'
                ? (
                    <>
                      <ValueFieldRows
                        label="Custom domains"
                        name="customDomains"
                        fields={customDomains.fields}
                        minimum={1}
                        placeholder="routes.example.com"
                        register={form.register}
                        error={form.formState.errors.customDomains}
                        onAdd={() => {
                          customDomains.append({ value: '' })
                          focusNewField(`customDomains.${customDomains.fields.length}.value` as Path<TunnelFormValues>)
                        }}
                        onRemove={customDomains.remove}
                      />
                      <FormField label="Location" error={form.formState.errors.location}>
                        <input {...form.register('location')} placeholder="All paths" aria-invalid={Boolean(form.formState.errors.location)} />
                      </FormField>
                    </>
                  )
                : (
                    <FormField label="Server port" error={form.formState.errors.serverPort}>
                      <input {...form.register('serverPort')} type="number" min="1" max="65535" placeholder="Automatic" aria-invalid={Boolean(form.formState.errors.serverPort)} />
                    </FormField>
                  )}
              <div className="field-row">
                <FormField label="Local host" error={form.formState.errors.localHost}>
                  <input {...form.register('localHost')} aria-invalid={Boolean(form.formState.errors.localHost)} />
                </FormField>
                <FormField label="Local port" error={form.formState.errors.localPort}>
                  <input {...form.register('localPort')} type="number" min="1" max="65535" aria-invalid={Boolean(form.formState.errors.localPort)} />
                </FormField>
              </div>
              <div className="setting-row">
                <div><strong>Enabled</strong></div>
                <Controller name="enabled" control={form.control} render={({ field }) => <Switch label="Enable Tunnel Definition" checked={field.value} disabled={saving} onChange={field.onChange} />} />
              </div>
            </div>
          </FormScrollArea>
        </Tabs.Content>

        <Tabs.Content value="transport" forceMount className="tunnel-step-panel">
          <FormScrollArea>
            <section className="form-section">
              <h3>Transport</h3>
              <div className="setting-row">
                <span>Encryption</span>
                <Controller name="useEncryption" control={form.control} render={({ field }) => <Switch label="Use encryption" checked={field.value} disabled={saving} onChange={field.onChange} />} />
              </div>
              <div className="setting-row">
                <span>Compression</span>
                <Controller name="useCompression" control={form.control} render={({ field }) => <Switch label="Use compression" checked={field.value} disabled={saving} onChange={field.onChange} />} />
              </div>
              <div className="setting-row">
                <span>Bandwidth limit</span>
                <Controller name="bandwidthEnabled" control={form.control} render={({ field }) => <Switch label="Limit bandwidth" checked={field.value} disabled={saving} onChange={field.onChange} />} />
              </div>
              {form.watch('bandwidthEnabled') && (
                <div className="field-row three-fields">
                  <FormField label="Limit" error={form.formState.errors.bandwidthValue}>
                    <input {...form.register('bandwidthValue')} type="number" min="0.01" step="any" aria-invalid={Boolean(form.formState.errors.bandwidthValue)} />
                  </FormField>
                  <FormField label="Unit" error={form.formState.errors.bandwidthUnit}>
                    <Controller name="bandwidthUnit" control={form.control} render={({ field }) => <Select label="Bandwidth unit" value={field.value} disabled={saving} onChange={field.onChange} options={[{ value: 'KB', label: 'KB' }, { value: 'MB', label: 'MB' }]} />} />
                  </FormField>
                  <FormField label="Limit at" error={form.formState.errors.bandwidthMode}>
                    <Controller name="bandwidthMode" control={form.control} render={({ field }) => <Select label="Bandwidth limit position" value={field.value} disabled={saving} onChange={field.onChange} options={[{ value: 'client', label: 'Client' }, { value: 'server', label: 'Server' }]} />} />
                  </FormField>
                </div>
              )}
              <FormField label="Proxy Protocol" error={form.formState.errors.proxyProtocolVersion}>
                <Controller name="proxyProtocolVersion" control={form.control} render={({ field }) => <Select label="Proxy Protocol" value={field.value} disabled={saving} onChange={field.onChange} options={[{ value: '', label: 'Off' }, { value: 'v1', label: 'v1' }, { value: 'v2', label: 'v2' }]} />} />
              </FormField>
            </section>
          </FormScrollArea>
        </Tabs.Content>

        <Tabs.Content value="health" forceMount className="tunnel-step-panel">
          <FormScrollArea>
            <section className="form-section">
              <div className="setting-row">
                <h3>Health check</h3>
                <Controller name="healthEnabled" control={form.control} render={({ field }) => <Switch label="Enable health check" checked={field.value} disabled={saving} onChange={field.onChange} />} />
              </div>
              {form.watch('healthEnabled') && (
                <>
                  <Controller name="healthType" control={form.control} render={({ field }) => <SegmentedControl label="Health check type" className="two-segments" value={field.value} disabled={saving} onChange={field.onChange} options={[{ value: 'tcp', label: 'TCP' }, { value: 'http', label: 'HTTP' }]} />} />
                  <div className="field-row three-fields">
                    <FormField label="Interval (s)" error={form.formState.errors.healthInterval}>
                      <input {...form.register('healthInterval')} type="number" min="1" aria-invalid={Boolean(form.formState.errors.healthInterval)} />
                    </FormField>
                    <FormField label="Timeout (s)" error={form.formState.errors.healthTimeout}>
                      <input {...form.register('healthTimeout')} type="number" min="1" aria-invalid={Boolean(form.formState.errors.healthTimeout)} />
                    </FormField>
                    <FormField label="Max failed" error={form.formState.errors.healthMaxFailed}>
                      <input {...form.register('healthMaxFailed')} type="number" min="1" aria-invalid={Boolean(form.formState.errors.healthMaxFailed)} />
                    </FormField>
                  </div>
                  {form.watch('healthType') === 'http' && (
                    <>
                      <FormField label="Health path" error={form.formState.errors.healthPath}>
                        <input {...form.register('healthPath')} aria-invalid={Boolean(form.formState.errors.healthPath)} />
                      </FormField>
                      <KeyValueFieldRows
                        label="Health check headers"
                        name="healthHeaders"
                        fields={healthHeaders.fields}
                        register={form.register}
                        error={form.formState.errors.healthHeaders}
                        onAdd={() => {
                          healthHeaders.append({ name: '', value: '' })
                          focusNewField(`healthHeaders.${healthHeaders.fields.length}.name` as Path<TunnelFormValues>)
                        }}
                        onRemove={healthHeaders.remove}
                      />
                    </>
                  )}
                </>
              )}
            </section>
          </FormScrollArea>
        </Tabs.Content>

        {protocol === 'http' && (
          <Tabs.Content value="http" forceMount className="tunnel-step-panel">
            <FormScrollArea>
              <section className="form-section">
                <h3>HTTP</h3>
                <div className="setting-row">
                  <span>Basic Auth</span>
                  <Controller name="authEnabled" control={form.control} render={({ field }) => <Switch label="Enable HTTP Basic Auth" checked={field.value} disabled={saving} onChange={field.onChange} />} />
                </div>
                {form.watch('authEnabled') && (
                  <div className="field-row">
                    <FormField label="Username" error={form.formState.errors.authUsername}>
                      <input {...form.register('authUsername')} aria-invalid={Boolean(form.formState.errors.authUsername)} />
                    </FormField>
                    <FormField label="Password" error={form.formState.errors.authPassword}>
                      <input {...form.register('authPassword')} type="password" placeholder={initial?.options.http?.basicAuth ? 'Unchanged' : ''} autoComplete="new-password" aria-invalid={Boolean(form.formState.errors.authPassword)} />
                    </FormField>
                  </div>
                )}
                <FormField label="Host Header Rewrite" error={form.formState.errors.hostHeaderRewrite}>
                  <input {...form.register('hostHeaderRewrite')} aria-invalid={Boolean(form.formState.errors.hostHeaderRewrite)} />
                </FormField>
                <KeyValueFieldRows
                  label="Request headers"
                  name="requestHeaders"
                  fields={requestHeaders.fields}
                  register={form.register}
                  error={form.formState.errors.requestHeaders}
                  onAdd={() => {
                    requestHeaders.append({ name: '', value: '' })
                    focusNewField(`requestHeaders.${requestHeaders.fields.length}.name` as Path<TunnelFormValues>)
                  }}
                  onRemove={requestHeaders.remove}
                />
                <KeyValueFieldRows
                  label="Response headers"
                  name="responseHeaders"
                  fields={responseHeaders.fields}
                  register={form.register}
                  error={form.formState.errors.responseHeaders}
                  onAdd={() => {
                    responseHeaders.append({ name: '', value: '' })
                    focusNewField(`responseHeaders.${responseHeaders.fields.length}.name` as Path<TunnelFormValues>)
                  }}
                  onRemove={responseHeaders.remove}
                />
              </section>
            </FormScrollArea>
          </Tabs.Content>
        )}
      </Tabs.Root>

      <FormError error={form.formState.errors.root?.server} />
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
                <button className="primary" type="button" disabled={saving} onClick={() => void nextStep()}>
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
    </DialogShell>
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
  const [confirmation, setConfirmation] = useState<ConfirmAction>()
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
                      {' '}
                      /
                      {' '}
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
                          <tr className={`tunnel-row${expanded.has(tunnel.id) ? ' is-expanded' : ''}`}>
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
                          {expanded.has(tunnel.id) && <tr className="details-row"><td colSpan={7}><TunnelDetails tunnel={tunnel} /></td></tr>}
                        </Fragment>
                      ))}
                    </tbody>
                  </table>
                  {!tunnels.length && <div className="empty-row">No Tunnel Definitions</div>}
                </section>
              </>
            )}
      {editing !== undefined && <TunnelEditor clientId={id} initial={editing ?? undefined} onClose={() => setEditing(undefined)} onSaved={saved} />}
      {confirmation && <ConfirmationDialog request={confirmation} onClose={() => setConfirmation(undefined)} />}
    </>
  )
}
