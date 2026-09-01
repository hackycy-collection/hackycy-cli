# Choose the config cm list terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact loading state, empty guidance, profile/default layout, Command
Result, and failure presentation should `config cm list` use while preserving
stored order and API-key non-disclosure?

## Answer

`config cm list` is a finite, read-only discovery command. It uses one
`Load CM profiles` Work Phase, emits a complete Command Result exactly once,
and presents the command-owned profile/default projection without decrypting
API keys. The terminal redesign changes Rich layout and lifecycle visibility
only; stored order, result fields, empty guidance, exit behavior, and the
secret boundary remain intact.

### Lifecycle and loading

- The command opens the shared Terminal Experience before obtaining the store,
  so the loading phase covers both store construction and
  `ListCMProfiles`. Its immutable phase catalog contains only
  `load-cm-profiles` / `Load CM profiles`.
- Rich starts one spinner immediately and resolves the phase to Completed or
  Failed without an artificial delay. Plain Interactive writes one
  control-free diagnostic line, `Loading CM profiles...`, to stderr before
  reading; Automation emits no loading line, Transcript, or terminal control.
- A successful read with zero profiles is a completed success. A store or
  reader failure marks the phase Failed and returns the original error without
  emitting a partial result.

### Rich projection

- Rich uses the title `Commit message profiles`, eyebrow
  `YCY / config cm list`, and subtitle `Configured providers for commit
  message generation`.
- Non-empty Rich Live Views use a responsive table with columns
  `DEFAULT`, `PROFILE`, `MODEL`, and `BASE URL`. Wide terminals align the
  columns; narrow terminals render a scrollable field block per profile. Huh/
  Bubbles layout handles long lists and wrapping without sorting, filtering,
  pagination, or a row limit. The complete result remains available on
  stdout.
- The stored profile order is preserved exactly. A profile is marked default
  only when its name exactly matches the stored `DefaultProfile`; Rich uses a
  `✓` success marker. Empty, missing, or dangling default values do not cause
  the presentation layer to infer or mutate a replacement default.
- Rich Live View may use colors, symbols, column widths, and responsive
  spacing, but the Command Result retains the existing semantic text and
  field order. Plain and Automation continue to emit:

  ```text
  Commit message profiles
  PROFILE  MODEL  BASE URL
    work gpt-4.1-mini https://work.example/v1
  * personal deepseek-chat https://personal.example/v1
  ```

### Safety projection

- The command consumes only `appconfig.CMProfile`, whose fields are profile
  name, Base URL, and model. It never resolves, decrypts, reads, or displays
  an API key or ciphertext.
- Rich, Transcript, and diagnostic contexts use bounded, single-line,
  control-free projections. Base URLs retain safe scheme, host, and path while
  removing userinfo, query, and fragment; an unsafe value becomes `Base URL
  configured`. Unsafe names and models use `Profile` and `Model configured`
  placeholders. These projections never alter stored identities or the
  existing machine-result semantics.

### Empty, result, and failure behavior

- The empty Rich result contains `No CM profiles configured.` and
  `Run "ycy config cm add" to add one.` as separate warning/muted blocks.
  Plain and Automation preserve the existing combined stdout guidance:
  `No CM profiles configured. Run "ycy config cm add" to add one.` Exit is 0,
  no interaction starts, and no configuration is created or changed.
- Successful non-empty and empty branches call
  `Finish(Succeeded, document)` exactly once. The failed store/read branch
  calls `Finish(Failed, nil)` at most once; `Close` is cleanup-only and cannot
  emit a second result.
- Plain has already emitted its durable loading line and never replays a
  Transcript. Automation emits only the existing complete result and no
  diagnostics or control sequences. Rich restores the primary screen before
  replaying its compact Transcript and then emits the result once.

### Interaction Transcript

Rich replays only semantic summaries after AltScreen. A successful non-empty
run records, in order, `Load CM profiles` completed, `Loaded 1 CM profile` or
`Loaded <N> CM profiles`, an optional `Default profile: <safe name>` when the
stored default safely matches a listed profile, and `Succeeded`. An empty run
records `Load CM profiles` completed, `Loaded 0 CM profiles`, `No CM profiles
configured`, and `Succeeded`. A dangling or unsafe default records no name.
Read/store failure records the safe Failed phase and `Failed`; context
cancellation may use a Cancelled phase state but preserves its existing error
and does not fabricate stdout `Cancelled`.

Rows, Base URLs, models, API keys, ciphertext, full Command Results, invalid
raw fields, animation frames, paths, and raw errors never enter the ledger.

### Evidence

Acceptance covers opening the Experience before store/read, Rich loading and
phase completion, wide/narrow layouts, long-list scrolling and wrapping,
stored order, default/empty/dangling-default cases, safe field projections,
API-key non-disclosure, Transcript count/default/failure ordering, Plain
loading stream and exact result format, Automation silence, no partial stdout
on failures, primary-screen restoration, one-shot Finish, control-sequence
stripping, and unchanged help/exit/result contracts.
