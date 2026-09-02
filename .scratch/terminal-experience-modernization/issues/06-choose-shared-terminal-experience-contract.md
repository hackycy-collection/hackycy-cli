# Choose the shared terminal Experience contract

Type: grilling
Status: resolved
Blocked by: 03, 04, 05

## Question

What semantic requests, Work Phase state model, transcript ledger, stream
ownership, alternate-screen lifecycle, result/failure ordering, cancellation
cleanup, and Service Command logging primitives should be shared, and which
presentation decisions must remain command-owned so all commands can be
independently designed without duplicating mechanics?

## Answer

The shared seam is a semantic terminal Experience. Commands supply semantic
requests, ordered Work Phase definitions and updates, explicit milestones, and
their own result documents; `internal/terminal` owns rendering, terminal
capabilities, state, transcript storage, and stream/lifecycle mechanics. A
command never imports Bubble Tea, Bubbles, Huh, Lip Gloss, or Log v2 and never
hands the terminal a Bubble Tea model.

### ExperienceRun contract

- Keep `Experience.Open(context.Context)` and `DiagnosticWriter()` as the
  invocation boundary. A run's semantic methods are serialized; concurrent
  calls from a command are not supported, while diagnostic logging may remain
  concurrent through the injected writer.
- Replace the public durable-result operation for finite commands with
  `Finish(FinishOutcome, *PresentationDocument)`. The outcome is exactly one
  of `Succeeded`, `Cancelled`, or `Failed`; a non-nil document is emitted to
  stdout once, including for a cancelled or failed command when that is part
  of its existing contract. A nil document means that the command has no
  stdout result. Existing finite commands that currently call `Result` more
  than once must compose one command-owned document first, preserving their
  established bytes and ordering. A Service Command may use the narrow,
  stable-ID result checkpoint Interface operation only when its command-
  specific ticket requires time-sensitive durable stdout; see
  [Choose the diff Lifecycle Log presentation](18-choose-diff-lifecycle-log-presentation.md)
  and [Choose the fs Lifecycle Log presentation](19-choose-fs-lifecycle-log-presentation.md).
- `Finish` atomically commits the run as finished before attempting output.
  A second call returns `ErrExperienceRunFinished` and never retries a partial
  write. Renderer, transcript, stdout, and cleanup errors are joined without
  changing the command's outcome or exit semantics.
- `Close` is idempotent cleanup only. It must restore terminal modes, freeze
  and safely replay any already-recorded Rich ledger, release the renderer
  lease, and emit no synthetic outcome or stdout result. Commands must call
  `Finish` on success, cancellation, and known failure paths; panic/root
  recovery still uses `Close` and the existing Diagnostic path.
- Add `Milestone(PresentationDocument)`. It is an explicit command-owned
  durable checkpoint: Rich renders it and records it, Plain Interactive writes
  it once as a control-free line, and Automation drops it. `Notice` remains
  transient Live View context and is never copied into the ledger.

### Interactions and redaction

- Validate an `InteractionRequest` synchronously before starting or changing a
  Rich view. Unknown kinds, duplicate option values, invalid defaults, and
  malformed required fields return `ErrInvalidInteractionRequest` without
  reading stdin or writing a ledger entry.
- Add optional `TranscriptLabel` and `Sensitive` request metadata. Only a
  successfully completed Ask is recorded as an answer. Secret interactions
  and any Sensitive request record `[redacted]`; Select/MultiSelect record
  option labels in selection order, never internal values; Text and Confirm
  record normalized semantic answers. Invalid retries and partial keystrokes
  are excluded.
- A cancelled Ask records only a labelled `cancelled` position marker, never
  a partial value. The terminal preserves the original interaction/context
  error for the command; only the visual/outcome projection is normalized to
  cancellation.
- `PresentationBlock` has an explicit `Sensitive` flag. Sensitive blocks are
  shown as `[redacted]` in Live View and Transcript. The terminal strips
  control sequences but does not run generic `logging.Redact` over ordinary
  business text; command adapters must construct safe provider errors, URLs,
  paths, and file-content summaries, while logging Runtime retains full
  Diagnostic redaction responsibility.

### Work Phases and tracking

- `TrackedOperation` declares one immutable, command-defined phase catalog in
  display order. Each `PhaseDefinition` has a stable ID and name; updates
  carry only the operation-scoped ID, state, and replaceable detail. Multiple
  Track calls are allowed sequentially, each with its own operation ID/label;
  only one Track is active at a time.
- Valid transitions are `Pending -> Active -> Completed|Failed|Cancelled`
  and `Pending -> Failed|Cancelled`. Terminal states cannot be rewritten,
  there is at most one Active phase, and no percentage is invented for work
  without a measurable total. Invalid IDs, transitions, or active-state
  violations return a recognizable protocol error; the runtime drains the
  update channel through closure before returning so producers cannot deadlock.
- The update stream carries state only, never a business error. Commands send
  a Failed phase and return the original business error; `Track` returns only
  renderer, protocol, I/O, or cancellation-cleanup errors.
- Cancellation calls `RequestCancel` at most once, then consumes updates until
  the producer closes the channel. The terminal adds no timeout and does not
  guess a final state. A real operation failure takes precedence over a
  cancellation signal.
- A real phase should submit Active and a terminal state whenever observable.
  Each update passes through at least one Bubble Tea model/view cycle without
  inserting artificial sleeps. The final phase state is retained even when a
  very fast Live View was not human-visible.

### Interaction Transcript ledger

The ledger is one append-only, monotonically numbered sequence per run. Ask
answer/cancellation markers and Milestones enter in call order. When a Track
closes, only its final phase states are appended, in catalog order; spinner
frames and intermediate details are never recorded. The final outcome is the
last ledger event. The ledger has fixed defaults of 64 events, 16 KiB total,
and 2 KiB per field; overflow drops later events at event boundaries and adds
one `... transcript truncated ...` marker. Text is normalized to UTF-8,
control-free lines before accounting.

Only Rich replays this compact semantic Transcript after AltScreen. Plain
Interactive already produced durable lines and does not replay them again;
Automation emits no Transcript, prompts, or extra controls. Large Command
Results are never copied automatically; commands use a small Milestone when a
summary is useful.

### Streams and lifecycle

Rich owns stderr and one AltScreen Bubble Tea root. Stdout remains exclusively
the Command Result. The renderer lease defers concurrent Diagnostic Records,
then flushes them FIFO after transcript replay. Completion is ordered as:

1. Freeze the ledger and commit the outcome.
2. Stop Bubble Tea and restore the primary screen/terminal modes.
3. Replay the safe Transcript while the renderer still owns stderr.
4. Release the lease and flush deferred diagnostics.
5. Emit the non-nil finite Command Result once on stdout. Service result
   checkpoints do not participate in an AltScreen completion sequence.
6. For `run`, release all terminal ownership before handing the selected child
   process the inherited terminal; child output and exit code are untouched.

Rich startup failures may fall back to Plain only before AltScreen starts and
before semantic state is committed. Once Rich has started, a renderer failure
must restore the terminal, replay the safe partial ledger with a control-free
writer, complete best-effort cleanup, and return the original renderer error
joined with later recovery errors. It must never repeat questions or work.

### Service Commands and ownership

`diff`, `fs`, `tunnel connect`, and `tunnel server` remain Service Commands:
they use the existing value `logging.Logger` and injected DiagnosticWriter for
line-oriented Lifecycle Logs, with no custom full-screen model or lifecycle
DSL. Log v2 is only a private text-only adapter behind logging Runtime; it
receives normalized/redacted records and does not filter, timestamp, redact,
choose formats, or write directly. Text logs remain line-oriented and use only
neutral, symbol-paired status markers and level-appropriate colors; they do
not adopt the full-screen Ops Console. JSON/NDJSON schema and record
boundaries remain exact.

Every command owns its information hierarchy, wording, phase catalog, prompt
order, option labels, result document, and command-specific failure/cancel
projection. Shared code is limited to semantic validation, Huh/Bubble Tea/Lip
Gloss visual primitives, capability/degradation policy, ledger mechanics,
phase validation, renderer leases, and logging facilities.
