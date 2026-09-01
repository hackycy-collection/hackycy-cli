# Choose the config fork list terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact loading state, empty guidance, provider-instance layout, ciphertext
preview treatment, Command Result, and failure presentation should
`config fork list` use while preserving order and secret non-disclosure?

## Answer

`config fork list` is a finite, read-only command. It uses a minimal loading
Live View when a terminal is available, then emits one complete Command Result
to stdout. It does not ask questions, mutate configuration, decrypt tokens, or
copy the instance table into the Interaction Transcript. The command adapter
owns the list wording and safe projection; `internal/terminal` owns the
loading phase, responsive layout primitives, and capability degradation.

### Loading lifecycle

- Rich TTY starts a compact Signal Rail view immediately and always shows
  `Loading fork provider instances` with one spinner and no percentage. After
  the read completes, the phase becomes `Completed`, the primary screen is
  restored, and the result is written to stdout.
- Plain Interactive writes one control-free
  `Loading fork provider instances...` diagnostic line before the final result.
  Automation writes no loading line or Transcript and emits only the final
  result. No artificial delay is introduced to make loading visible.
- The semantic phase is `Load fork provider instances`. A successful read with
  zero rows is still a completed success, not a warning failure.

### Result layout

- Keep the title `Fork provider instances`. Rich adds the eyebrow
  `YCY / config fork list` and the subtitle `Configured providers for git fork
  operations`; it does not show configuration paths, salts, read timings, or
  other internal metadata.
- Wide Rich TTYs use an aligned five-column table: `NAME`, `TYPE`, `SCHEME`,
  `HOST`, and `TOKEN`. Rows retain the exact `ListForkInstances` order and the
  result ends with the existing singular/plural configured count.
- Narrow Rich terminals use a compact single-column block per instance with
  explicit field labels and wrapped values; they do not horizontally truncate
  names, hosts, or providers. A Bubbles viewport may allow browsing the Live
  View for very large lists, but the final stdout result always contains every
  row. There is no pagination or row limit.
- Plain and Automation preserve the existing tabwriter-based durable text and
  field order. Styling, spacing, lock symbols, and responsive wrapping are
  terminal projections only and never change the redirected semantic fields.

### Token preview and safe data

- Continue displaying the existing `TokenPreview` bytes (for example,
  `MDEy***`) to preserve the current result contract. The Rich header may say
  `TOKEN` with an adjacent explanation that it is an encrypted preview, and a
  lock symbol may reinforce that meaning; no new secret-derived value is
  introduced.
- Never decrypt, display plaintext, expose complete ciphertext, infer token
  length/checksums, or copy previews into the Transcript. Empty or short
  previews remain the appconfig `***` projection. Terminal rendering strips
  controls and applies the normal field-size safety limit; malformed values
  produce a safe warning without echoing the raw value.
- The appconfig reader remains the security boundary: the command consumes
  only its secret-safe `ForkInstance` projection and does not access
  credentials directly.

### Empty, failure, and cancellation states

- An empty successful result contains the title, `No instances configured.`,
  and `Run "ycy config fork add" to add one.` Rich uses yellow warning
  emphasis paired with the words, then records `Loaded 0 fork provider
  instances`; exit remains 0 and no extra interaction is started.
- A configuration read error marks `Load fork provider instances` as Failed in
  Rich, replays only a safe phase detail after AltScreen, and returns the
  original error. No partial table, title, or other success-looking stdout is
  submitted; root emits the existing single redacted error diagnostic and exits
  1. Plain may already have emitted the loading diagnostic, while Automation
  emits no loading line.
- If a context cancellation occurs before or during the read, the phase may be
  marked Cancelled for the Live View/Transcript, but the command preserves the
  existing cancellation/error return contract and does not fabricate a
  `Cancelled` stdout result for this non-interactive read.

### Transcript and completion

After Rich exits AltScreen, the compact Transcript contains only the loading
phase final state, `Loaded N fork provider instances`, the empty-state guidance
when applicable, and the final outcome. It never repeats host, provider,
scheme, names, token previews, or the full table. Plain has already emitted
durable lines and does not replay them; Automation has no Transcript.

The adapter calls `Finish(Succeeded, document)` exactly once after a successful
read, including the empty case. On read failure it calls `Finish(Failed, nil)`
at most once and returns the original error; `Close` remains cleanup-only and
cannot produce a second result. The result is complete and independent from
diagnostics.

### Evidence

Acceptance must cover Rich PTY loading/completion/primary-screen restoration,
wide and narrow layouts plus optional viewport browsing, Plain loading/result
ordering, Automation's zero extra loading output, empty-state guidance,
persistent instance order, long fields, safe TokenPreview behavior and
control stripping, read/context failures with no partial stdout, one complete
Finish submission, Transcript count/phase/outcome limits, and unchanged
command-surface/help snapshots.
