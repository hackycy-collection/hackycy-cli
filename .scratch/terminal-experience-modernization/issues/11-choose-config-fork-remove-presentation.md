# Choose the config fork remove terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact instance selection, confirmation, empty state, removal Work Phases,
Interaction Transcript, and success/cancellation/failure result should
`config fork remove` use while preserving selection and default-negative
confirmation semantics?

## Answer

`config fork remove` is a finite destructive interaction with one read phase,
two ordered questions, and one conditional mutation phase. It uses the shared
terminal Experience contract and owns its option labels, safety projections,
phase wording, transcript milestones, and result document. Existing selection,
default-negative confirmation, mutation, exit, and stdout semantics remain
unchanged.

### Presentation and interaction

- Rich uses one Bubble Tea root with a Signal Rail and Huh v2 child forms. The
  eyebrow is `YCY / config fork remove`, the title is `Remove fork provider
  instance`, and the subtitle is `Choose a configured provider connection to
  remove`.
- The command always starts the `Load fork provider instances` phase before
  reading configuration. Rich shows a spinner and Plain Interactive emits
  `Loading fork provider instances...`; Automation is silent. The phase ends
  immediately in Completed or Failed without an artificial delay.
- After a successful non-empty read, Huh Select keeps the exact persisted
  order, uses `Select instance to remove`, and defaults to the first item.
  Each choice uses the instance name as its primary label and a safe Host
  projection as its description. Rich uses Huh/Bubbles scrolling and wrapping
  for long lists and narrow terminals; it adds no sorting, filtering,
  pagination, or row limit. Capability degradation falls back to the existing
  control-free Plain selection flow.
- Confirmation remains a separate Huh Confirm with message
  `Remove instance "<name>"?`, default `No`, and an explicit destructive
  visual treatment. No additional review or confirmation step is introduced.

### Safety and identity

- The writer receives the selected persisted name exactly as returned by the
  reader. Display labels, confirmation text, Transcript fields, and diagnostic
  context use a command-owned safe projection that strips control characters,
  enforces single-line bounded output, and removes unsafe Host URL userinfo,
  query, and fragment components. If a name or Host cannot be safely
  projected, the presentation uses a generic label such as `Selected instance`
  or `Instance removed` instead of echoing it.
- The command consumes only `appconfig.ForkInstance`; it never decrypts,
  displays, or records token material, ciphertext, or token previews.

### Work Phases and outcomes

- A confirmed removal starts `Remove provider instance`, always showing Active
  with a spinner and then Completed or Failed. Fast operations still submit
  both observable states without sleeping. A writer result of `removed=false`
  remains a successful idempotent removal and produces the existing success
  wording.
- The one-shot `Finish` mapping is:

  | Branch | Outcome | Command Result |
  | --- | --- | --- |
  | Non-empty, writer succeeds (`removed` true or false) | `Succeeded` | `Instance <name> removed` |
  | Empty list | `Succeeded` | `Nothing to remove` |
  | Selection Esc, confirmation Esc, or Confirm No | `Cancelled` | `Cancelled` |
  | Read or writer failure | `Failed` | no stdout document |

  `Finish` is called at most once; `Close` is cleanup-only and cannot emit a
  second result.
- An empty list is a successful non-mutating branch. Rich records the load
  completion and `No instances configured`, while stdout emits only
  `Nothing to remove`; no selection, confirmation, or config creation occurs.
- Selection cancellation, confirmation cancellation, and explicit decline all
  avoid the writer and retain stdout `Cancelled`. Rich distinguishes them as
  `Selection cancelled`, `Confirmation cancelled`, and `Removal declined`,
  respectively. A context cancellation during the read marks the load phase
  Cancelled and preserves the original context error without fabricating a
  cancellation result. A context cancellation after mutation begins remains
  the writer's real failure/result, not a user-cancel projection.

### Transcript and streams

After Rich freezes the ledger and exits AltScreen, it replays only semantic,
redacted events. Successful non-empty removal is ordered as:

1. `Load fork provider instances` completed
2. `Selected instance: <safe name>`
3. `Host: <safe host>`
4. `Remove instance "<safe name>": confirmed`
5. `Remove provider instance` completed
6. `Succeeded`

The empty, cancellation, decline, read-failure, and writer-failure branches
retain the corresponding load/interaction/phase final states and end with
`Succeeded`, `Cancelled`, or `Failed`. Spinner frames, keystrokes, invalid
input, partial answers, raw errors, paths, and secret-derived values never
enter the ledger. Plain has already emitted durable lines and never replays
the Transcript; Automation emits no prompts, Transcript, or control
sequences.

Plain Interactive emits `Loaded fork provider instances` after the loading
line, and after confirmation emits `Removing provider instance...` before the
single existing success result. Empty and cancelled branches retain their
existing result lines; diagnostics remain separate and control-free. Read or
write failures submit no success-looking stdout and leave root to emit its
single redacted error.

### Evidence

Acceptance covers Rich PTY loading and phase completion, persisted ordering,
first-item default, safe labels, long/narrow lists, default-No destructive
confirmation, all cancellation/decline locations, empty success,
`removed=false` idempotence, read/write/context failures, transcript ordering
and redaction, exactly one Finish, primary-screen restoration, Plain loading
and mutation ordering, Automation zero-side-effect behavior, control-sequence
stripping, and unchanged help/exit/config/result contracts.
