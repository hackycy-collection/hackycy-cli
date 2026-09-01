# Choose the tunnel server Lifecycle Log presentation

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What line-oriented configuration, listener, control-plane, managed-FRP,
warning, failure, and shutdown Lifecycle Log should `tunnel server` emit at
each log level and in text versus NDJSON modes without exposing credentials,
adding a full-screen UI, or logging high-volume requests at Info?

## Answer

`tunnel server` is a Service Command with no AltScreen or custom full-screen
dashboard. It owns a line-oriented Lifecycle Log for one Tunnel Server Session.
The control-plane listener and the managed FRPS data plane are separate status
domains: the control plane may be available while FRPS is stopped, recovering,
or failed. Existing flag/environment precedence, protocol/API behavior,
resource ownership, stdout behavior, exit codes, and side effects remain
unchanged.

### Startup boundary and phases

The command keeps its existing configuration resolution before runtime
construction. The semantic lifecycle is:

1. `server.starting` (Info) begins the resolved server invocation.
2. `state.opened` (Info) follows data-directory lock acquisition, session-store
   and SQLite opening, account/session/control-plane composition, and internal
   FRP token resolution. Safe substeps may be Debug.
3. `control.listening` (Info) follows successful binding of the control HTTP
   listener.
4. `server.started` (Info) follows `control.listening`; it does not wait for
   FRPS.
5. `frps.preparing` (Debug) covers pinned artifact/runtime preparation,
   configuration publication, and verification.
6. `frps.running` (Info) or `frps.failed` (Error) reports the actual managed
   FRPS result. FRPS startup runs independently after the listener is usable.
7. The session remains in the foreground until context cancellation, signal,
   listener failure, or an unrecoverable resource/cleanup failure.

`server.started` therefore means that the control plane is serving, not that
the FRP data plane is healthy. A failed FRPS preparation or activation keeps
the listener and browser/API control surface alive whenever the existing
runtime permits it.

### Safe configuration and listener projection

Info records expose only operationally useful, non-secret fields: listener
kind (`control`, `frp`, `http-vhost`), port, port-pool bounds, an address class
(`loopback`, `private`, `public`, or `unspecified`), whether an advertised FRP
address is configured, advertised host class/port, session idle lifetime, and
whether the FRP token was explicit or generated/reused. They never expose the
complete URL, data directory, SQLite/WAL path, administrator username or
password, Internal FRP Token, generated TOML, or any credential field. Debug
may identify only a safe path category, never an absolute path.

The existing numerical parsing, credentials, port collision, and pool-boundary
validation remain unchanged. Configuration failures occur before a runtime is
owned and are reported through the existing root diagnostic path.

### Failure phases and cleanup

Startup failures retain their phase and safe class rather than collapsing into
an opaque message:

- `config.resolution_failed` for command configuration resolution;
- `state.open_failed` for lock, session-store, or SQLite initialization;
- `control.composition_failed` for accounts, sessions, control plane, or HTTP
  handler composition;
- `frps.preparation_failed` for pinned FRP runtime preparation or verification;
- `control.bind_failed` for the listener bind; and
- `control.listener_failed` for an unexpected Serve termination.

Each startup or serving failure produces at most one final `server.failed`
Error with `phase`, fixed `failureClass`, and `cleanup=succeeded|failed`.
Already-owned resources release in the current reverse order, with release
errors folded into that one outcome. `http.ErrServerClosed` is normal and does
not produce an Error. A request cancellation, SSE disconnect, or client socket
close does not alter the Session outcome.

### Managed FRPS lifecycle

The FRPS supervisor is the only source of process state. It emits:

- `frps.running` (Info) after a child is actually running;
- `frps.stopped` (Info, `reason=admin|shutdown|configuration`) for a real stop;
- `frps.recovering` (Warn) after an unexpected exit;
- `frps.recovered` (Info) after automatic recovery;
- `frps.failed` (Error, `failureClass=configuration|activation|frps`) for
  verification, start, or activation failure; and
- `frps.restarted` (Info) only after an administrator restart actually
  succeeds.

Administrator start/stop/restart requests are Debug intent records; final
status transitions are emitted only after the supervisor confirms them.
Automatic recovery and administrator operations are ordered by actual state
transitions. Canceled or superseded recovery timers do not create synthetic
failures. FRPS child stdout/stderr is never normal-level Lifecycle Log: bounded,
control-escaped, redacted `frps.child_output` lines are Debug-only. Scanner or
child-ownership errors are one safe Warn without PID, path, or raw process
output.

### HTTP, agent, and control-plane changes

There is no default access log. Info never records every HTTP request, static
asset, health check, SSE poll, WebSocket ping/pong, or control frame. Debug may
record a bounded request summary with method, stable route template,
status-class, and duration, omitting raw path/query/body/header/cookie/origin,
remote address, user agent, and response content. Expected 4xx validation,
authorization, and not-found results stay Debug; unexpected internal failures
are one `control.request_failed` Error without raw error text.

An authenticated agent connection produces `agent.connected` (Info), one
`agent.disconnected` (Warn) per failure window, `agent.restored` (Info), or
`agent.revoked` (Warn, `reason=rotated|deleted`). Invalid token, duplicate
connection, protocol incompatibility, liveness timeout, and FRPS-unavailable
attempts are fixed-category, window-aggregated Warnings; detailed attempts are
Debug. Apply acknowledgements and process-state reports are logged only when a
revision or state actually changes. The display may use an escaped remark and
short non-reversible `clientRef`; it never logs client/account IDs, token,
address, IP, version payload, or raw frame.

Successful durable mutations emit one Info `control.change` after commit. The
projection contains action, object kind, bounded count, desired/applied
revision, enabled state, and optional `clientRef` for account, client, tunnel,
FRPS, or custom-404 changes. Input validation failures and rolled-back
transactions do not emit success changes; unexpected storage failures emit a
safe Error while the Session continues when possible.

### Levels, event schema, and text projection

The level policy is fixed:

- Info: startup, state opened, listener listening, server started/stopped, FRPS
  running/stopped/recovered/restarted, agent connect/restore, and committed
  control changes.
- Warn: FRPS recovering, one agent failure window, aggregated security or
  protocol warnings, recoverable FRPS/reconciliation issues, and individual
  shutdown cleanup failures.
- Error: configuration, lock/database/composition, bind/serve, FRPS
  preparation/verification/activation, and final cleanup failures.
- Debug: safe substeps, request summaries, retry/backoff, protocol stages,
  supervisor intent, state reports without change, and bounded child output.

Text and NDJSON share event IDs, levels, filtering, and ordering. Text is one
line in the fixed order `timestamp level scope symbol message details` with
`▶` entering, `✓` success/recovery, `↻` recovery in progress, `!` warning, `×`
failure, and `■` stop. Rich TTY colors reinforce these symbols (cyan, green,
yellow, red, muted gray); Plain and redirected output remove ANSI but keep
symbols and wording. NDJSON never contains ANSI.

NDJSON preserves the existing top-level `timestamp`, `level`, `scope`,
`message`, and `context` fields. New stable data lives in `context.event` and
bounded fields such as `category=server|control|agent|frps|change|shutdown`,
`phase`, `listener`, `addressClass`, `processState`, `agentState`, revisions,
counts, `durationMs`, `cleanup`, `reason`, `outcome`, and fixed
`failureClass=configuration|lock|database|composition|bind|transport|
authentication|protocol|frps|cleanup|unknown`. Fields are omitted when not
meaningful; free-form implementation details are not added. The Log v2
adapter remains behind the existing redaction and stable-schema facade.

### Shutdown and terminal result

Cancellation or SIGINT/SIGTERM emits `shutdown.requested` with
`reason=cancelled|signal`; listener failure may use `reason=listener-failure`.
The ordered close is: stop accepting requests and close the listener, stop
FRPS and await child ownership release, close HTTP/session/database state,
release the server lock, emit one `server.failed` if needed, then always emit
one final `server.stopped` with `outcome=succeeded|cancelled|failed`. The
`server.stopped` record is always last. `Close`/`Wait` races, duplicate signals,
and repeated cleanup are idempotent and cannot duplicate terminal records.
Existing context/signal return semantics and exit codes remain authoritative.

`tunnel server` adds no stdout Command Result document. Lifecycle records,
FRPS child output, and request summaries use the configured diagnostic stream.
The command never replays a Transcript after exit and never duplicates the
service log or child output. Automation remains free of forms, styles, and
terminal control sequences.

### Redaction and acceptance evidence

No log projection may expose administrator or client credentials, Internal FRP
Token, complete URL, absolute path, IDs, IP, headers, cookies, request/response
bodies, control frames, generated FRPS configuration, close text, or raw HTTP,
WebSocket, filesystem, or child-process errors. All free text is bounded,
redacted, and control-escaped; sensitive or unprovably safe child lines become
`suppressed`.

Acceptance covers configuration precedence and legacy numeric/credential
validation; lock, session, SQLite, token generation/reuse, and composition;
listener-before-FRPS startup; independent control-plane/FRPS availability;
bind/serve/runtime failures and reverse cleanup; administrator FRPS controls,
custom-404 and committed domain changes; agent authentication, duplicate,
protocol, liveness, disconnect/recovery/revoke behavior; supervisor recovery,
backoff and child-output isolation; HTTP request/log aggregation; context and
signal shutdown; Close/Wait races; exactly-once `server.failed` and
`server.stopped`; Text symbols/color degradation; NDJSON schema/event
goldens; complete redaction; redirected/no-extra-stdout checks; Automation
control-free behavior; and unchanged APIs, protocol, results, exit codes, and
side effects.
