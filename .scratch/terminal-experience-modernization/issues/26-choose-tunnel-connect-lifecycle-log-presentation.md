# Choose the tunnel connect Lifecycle Log presentation

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What connection-selection interaction followed by line-oriented resolution,
authentication, reconciliation, FRP supervision, reconnect, warning, failure,
and shutdown Lifecycle Log should `tunnel connect` use at each log level and in
text versus NDJSON modes without exposing tokens or adding a full-screen UI?

## Answer

`tunnel connect` is a Service Command with one bounded exception: ambiguous
remembered connections may use a temporary selection form before the long-lived
client starts. After selection, the command owns a line-oriented Lifecycle Log
for one Tunnel Client Session. Existing configuration precedence, protocol
behavior, reconnect behavior, result/exit contracts, and side effects remain
unchanged.

### Connection selection

`ResolveClientConfig` keeps the current source precedence and candidate order:
explicit `--server`/`--token`, their environment equivalents, token file,
remembered connections, and `DefaultTunnelServer` as already defined by the
command. A selector is invoked only when more than one remembered candidate
matches. One candidate remains an automatic resolution and an absent selector
continues to produce the existing non-interactive ambiguity error.

Rich uses a Huh v2 `Select` in a short-lived Bubble Tea AltScreen run. The
options preserve newest-first catalog order and contain a normalized server
origin, last-authenticated time, and, only when needed to distinguish equal
origins, a stable display ordinal. They never contain a token, token fragment,
connection ID, derived instance ID, or state path. The selected value remains
the opaque internal connection ID. The form exits before the client session
starts; its completed selection or cancellation is replayed as one semantic
summary on stderr.

Plain keeps the established numbered prompt, invalid-input retry, and
control-free output. Empty input, `q`, `quit`, and `cancel` remain user
cancellation. Automation never reads stdin: an ambiguous catalog returns the
existing ambiguity error without a form, transcript, or terminal controls.
Cancellation returns the established successful cancellation result exactly
once and does not acquire an instance or start a client.

### Session phases and event catalog

The session records meaningful state transitions, not requests, heartbeats,
polling, animation frames, or fabricated progress. The semantic event IDs are
stable and are emitted in this order when their stages are reached:

1. `client.starting` (Info) begins configuration-to-session startup.
2. `client.started` (Info) follows instance-lock acquisition and basic agent
   initialization. Debug-only records may describe safe source classes,
   remembered-state presence, lock ownership, and managed-runtime preparation.
3. Each control attempt may emit Debug records for `control.probe`,
   `control.handshake`, and welcome validation. Successful first authentication
   emits `control.authenticated` (Info); authentication after a transient
   outage emits `control.restored` (Info) instead.
4. A desired revision emits `state.applied` (Info), `state.skipped` (Debug), or
   `state.apply_failed` (Warn). A later successful revision after a failed one
   emits `state.recovered` (Info).
5. Real managed FRP transitions emit `frp.running`, `frp.stopped`,
   `frp.recovering`, `frp.recovered`, or `frp.restarted` at the levels below.
6. Transient control loss emits one `control.disconnected` (Warn) per failure
   window, followed by Debug `reconnect.scheduled` records and, on success,
   `control.restored`.
7. A revoke emits `control.revoked` (Warn), then follows the normal shutdown
   sequence.
8. Shutdown and terminal outcome emit `shutdown.requested` when applicable,
   optional `client.failed`, and exactly one final `client.stopped`.

The event IDs and ordering are the same in text and NDJSON. A failed desired
revision is reported to the control plane using the existing apply-result
protocol; it does not by itself terminate the session. Fatal startup or
session failures retain the current returned error.

### Levels and safe failure classes

Info is reserved for user-relevant lifecycle progress: starting, started,
authenticated/restored, applied/recovered state, real FRP running/stopped or
restarted transitions, and normal stopped outcome. Warn covers one transient
disconnect window, FRP unexpected exit/recovery, remembered-connection write
failure, stale-state cleanup failure, and reconciliation failure whose previous
state was restored. Error covers one final authentication, incompatibility,
protocol, runtime preparation, activation/rollback, or cleanup failure.

Debug may include safe source categories, probe/handshake/welcome stages,
attempt number, delay, revision skip reason, restart/revoke dispatch, and
bounded FRP child output. It never becomes a path for secrets or raw protocol
payloads. Failure details use fixed classes only:
`unauthorized`, `incompatible`, `protocol`, `transport`, `server`,
`local-callback`, `configuration`, `activation`, `rollback`, `cleanup`, and
`frp-child`.

Authentication failures caused by 401/403 or protocol/version incompatibility
are fatal. DNS/TLS/timeout/temporary server failures are transport or server
failures and enter reconnect. Before the first successful authentication,
repeated attempts remain Debug-only; after an authenticated connection drops,
the first disconnect is the single Warn for that failure window.

### Reconciliation and FRP supervision

`state.applied` contains only revision, total tunnel count, enabled count, and
the resulting `running` or `stopped` state. Older or duplicate revisions are
Debug `state.skipped`. `state.apply_failed` contains the safe reconciliation
code (`CONFIGURATION_FAILED`, `ACTIVATION_FAILED`, or `APPLY_FAILED`) and
`rollback=restored|not-required|failed`. It never includes generated TOML,
FRP host/port, internal FRP token, configuration path, or raw error text.

The supervisor is the source of truth for FRP state. `frp.running` is emitted
only after a child is actually running; desired-state disable emits
`frp.stopped` with `reason=desired-state`; an unexpected exit emits
`frp.recovering` (Warn) and a successful restart emits `frp.recovered` (Info).
An explicit `restart_frpc` request is Debug, with `frp.restarted` only after
the real restart succeeds. A restart with no enabled applied state is Debug
skipped. Configuration failures are not duplicated by a second FRP warning.

FRP child stdout/stderr is Debug-only `frp.child_output`, bounded to one safe
line at a time and marked with its stream. Every line passes the common
redactor and visible control-character escaping; suspected credentials,
configuration, URLs, or unprovably safe content are replaced with
`suppressed`. Scanner and child-ownership errors are one safe Warn each, with
no PID, path, or raw error.

### Reconnect and shutdown

The existing backoff schedule and retry behavior remain. `reconnect.scheduled`
is Debug and contains integer `attempt`, `delayMs`, and `backoffCapped`; the
attempt counter starts at one for each failure window. A successful restore
closes the window and resets the counter. Backoff waits observe context
cancellation immediately rather than sleeping through shutdown.

Cancellation or SIGINT/SIGTERM records `shutdown.requested` with
`reason=cancelled|signal`, stops new retries and work, closes the control
socket, stops reconciliation/FRP, and releases the instance. A server revoke
records `control.revoked` with `reason=rotated|deleted` and shuts down with
`reason=revoked`. Ordinary transport disconnect does not request shutdown.

The cleanup order is fixed and idempotent: close control connection; stop the
reconciler or supervisor and await owned-child release; release the client
instance; emit one `client.failed` (Error) if the main operation or cleanup
produced an error; and always emit one final `client.stopped` (Info) with
`outcome=succeeded|cancelled|revoked|failed`. `client.stopped` is the last
session record. Repeated signals, `Close`, or cleanup callbacks cannot create
duplicate terminal records. Existing context/signal exit codes and returned
errors take precedence over presentation wording.

### Text, NDJSON, and stream boundaries

Text records remain single-line and use the fixed order
`timestamp level scope symbol message details`. Symbols are `▶` for entering a
stage, `✓` for success/recovery, `↻` for reconnect/recovery in progress, `!`
for warning, `×` for failure, and `■` for stop. Rich TTY output uses cyan,
green, yellow, red, and muted gray to reinforce those states. Plain output,
redirected output, and terminals without color retain the symbols and wording
without ANSI. Debug is visually muted; Error is prominent. Symbols and color
are never required to understand the state.

NDJSON preserves the existing top-level `timestamp`, `level`, `scope`,
`message`, and `context` fields. New stable data is placed under
`context.event` plus bounded fields such as `category`, `reason`, `outcome`,
`failureClass`, `rollback`, `attempt`, `delayMs`, `backoffCapped`, `revision`,
`tunnelCount`, `enabledCount`, `state`, and `transition`. Fields are omitted
when not meaningful; enumerations and integer ranges are fixed. NDJSON never
contains ANSI or free-form process output. The Log v2 adapter remains behind
the existing redaction and stable-schema facade.

The only AltScreen surface is the bounded Rich connection selector. Once it
closes, its semantic summary is written to stderr and the Lifecycle Log owns
the remaining session. No long-lived full-screen dashboard is added. The
command does not add a stdout success document, duplicate a result, or copy
child/FRP output into stdout. Automation remains free of forms, transcripts,
styles, and terminal control sequences.

### Redaction and acceptance evidence

No selector, summary, text record, NDJSON record, diagnostic, or child-output
projection may expose Client Tokens, token fragments, Authorization headers,
Internal FRP Tokens, complete control-plane URLs outside the selection form,
derived IDs, state/configuration paths, complete FRP configuration, raw
control frames, response bodies, close text, or raw HTTP/WebSocket/filesystem
errors. Safe projections use source categories, bounded counts, revisions,
fixed failure classes, and visible escaping for controls.

Acceptance must cover configuration precedence and candidate order; Rich Huh
selection and AltScreen teardown; Plain numbered selection and cancellation;
Automation ambiguity; fake probe/WebSocket protocol and every failure class;
desired-state apply/skip/failure/rollback/recovery; restart, revoke, FRP
unexpected exit and recovery; reconnect-window throttling and cancellable
backoff; lock acquisition, stale-state cleanup, context and signal shutdown;
exactly-once `client.failed`/`client.stopped`; child ownership release; text
symbol/color degradation; NDJSON schema and event goldens; redaction of all
sensitive shapes; redirected-stream and no-extra-stdout checks; and unchanged
exit codes, side effects, protocol messages, and result behavior.
