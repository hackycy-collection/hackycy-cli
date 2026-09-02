# Choose the config fork add terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact text, provider, protocol, and secret form; validation display;
persistence Work Phases; redacted Interaction Transcript; and
success/cancellation/failure result should `config fork add` use without
changing prompt order, defaults, validation, or overwrite behavior?

## Answer

`config fork add` is a finite interactive setup flow with five ordered
questions followed by one persistence phase. It uses one Bubble Tea root with
Huh v2 child forms and the Ops Console status-table layout; the command keeps ownership of prompt
wording, validation, provider/protocol values, overwrite behavior, and result
text. Automation remains rejected before configuration access or stdin reads.

### Form structure and controls

- Keep the exact observable question order: `Instance name (alias)`, `Host`,
  `Provider type`, `Protocol`, and `Access token`.
- Rich wraps these fields in one root model but advances through one Huh child
  form at a time. Users cannot reorder or jump between questions. Every field
  retains its current placeholder and validation callback, including
  `Name is required`, `Name cannot contain spaces`, `Host is required`,
  `Token is required`, `invalid provider type`, and `invalid protocol`.
- The Rich title is `Add fork provider instance` with subtitle `Store a
  provider connection for git fork operations` and eyebrow `YCY / config fork
  add`. The existing prompt messages and placeholders remain visible so this
  is a visual modernization, not a grammar change.
- Provider options remain `GitLab`/`GitHub` with internal values `gitlab`/
  `github`; GitLab remains the default. Protocol options remain `HTTPS` and
  `HTTP (self-hosted / no TLS)` with internal values `https`/`http`; HTTPS
  remains the default. Descriptions may clarify use but cannot add choices or
  alter submitted values.
- Access token uses Huh Password with masked input throughout. It never shows
  a preview, length, prefix, strength, or echo. Its Transcript value is always
  `[redacted]`.

### Ops Console status table and Work Phases

The `STATE / PHASE / DETAIL` table shows six semantic rows: `Identity`,
`Host`, `Provider`, `Protocol`, `Credential`, and `Save provider instance`.
The active field uses `◆`, completed fields use `✓`, and pending fields use
`○`; the credential row never displays its value. The active Huh form sits in
the content region below the table and uses B's bottom focus rule. Field
validation retries stay inside the current form and are not separate phases;
there is no persistent left rail.

The command declares two Work Phases: `Collect provider details` covers the
five prompts, and `Save provider instance` covers the writer call. The form
phase remains active while questions are answered. The save phase always shows
a spinner while the write is in progress, then a terminal Completed or Failed
state; no artificial delay is introduced for fast writes.

No review or confirmation page is added after the token. Once the fifth answer
is valid, the adapter may show a non-sensitive summary (alias, safe host,
provider, and protocol) and immediately begins persistence.

### Host and overwrite safety

- The saved Host value remains exactly as entered after existing validation;
  the terminal does not trim or normalize it for persistence.
- Rich Transcript and other durable summaries use a command-owned safe host
  projection: control characters are removed and URL userinfo, query, and
  fragment components are omitted. If safe parsing fails, the summary says
  `Host configured` rather than echoing an unsafe value. The persistent result
  keeps the existing alias/host wording without exposing the token.
- `SaveForkInstance` continues to silently replace an existing alias. The
  command does not perform a preflight read or add a confirmation. If a future
  writer can distinguish insertion from replacement, it may say `Added` or
  `Updated`; with the current writer the neutral success wording is
  `Instance <alias> saved successfully` (the existing success bytes remain
  the compatibility baseline).

### Transcript and outcomes

After Rich exits AltScreen, a successful Transcript records in prompt order:
`Instance name: <alias>`, safe `Host`, `Provider: <label>`, `Protocol:
<label>`, `Access token: [redacted]`, the completed save phase, and `Succeeded`.
It excludes invalid attempts, keystrokes, cursor movement, animation frames,
absolute config paths, plaintext credentials, and ciphertext.

Cancellation at any of the five form steps (Esc, configured cancel value, or
context cancellation before persistence) stops the remaining prompts and never
calls the writer. Rich records the field location and `cancelled` marker;
Plain/Automation retain their existing one-line `Cancelled` result and exit 0
for an interactive cancellation. A context cancellation after the save call
has begun is a save failure with its original error, not a fabricated user
cancellation.

Successful Plain and Automation-compatible result bytes preserve the existing
`Instance <alias> (<host>) added successfully` message; Rich may add a `✓` and
color as a projection, but never changes the semantic host or adds credentials.
The adapter submits exactly one `Finish(Succeeded, document)` for success or
one cancelled Finish for pre-save cancellation. Save failure submits
`Finish(Failed, nil)` at most once, leaves stdout free of success/partial
fields, retains non-sensitive completed context in the Rich Transcript, and
returns the original writer error for root's single redacted diagnostic.

### Mode behavior and evidence

Rich holds the diagnostic lease during the form and save, restores the primary
screen before replaying the Transcript, then flushes deferred diagnostics.
Plain emits its existing prompts and validation retries plus
`Saving provider instance...` and `Saved provider instance` around the writer;
all are control-free. Automation returns
`config fork add requires an interactive terminal` before constructing the
store, reading stdin, or writing state.

Acceptance must cover Rich PTY prompt order/defaults/descriptions, validation
retry, password masking, Ops Console table states, fast-save loading and
completion, successful/duplicate-alias/cancelled/save-failure journeys,
primary-screen restoration and Transcript redaction, Plain retry/loading
ordering, Automation's zero side effects, host safety projection, encrypted
storage, exactly one Finish call, and unchanged command-surface/help
snapshots.
