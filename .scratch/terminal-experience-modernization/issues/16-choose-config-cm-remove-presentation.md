# Choose the config cm remove terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact confirmation, profile/default context, removal Work Phases,
Interaction Transcript, and success/cancellation/missing/failure result should
`config cm remove` use while preserving its default-negative confirmation and
default-profile reassignment behavior?

## Answer

`config cm remove <profile>` is a parameter-driven destructive interaction
with one pre-confirmation read, one default-negative confirmation, and one
conditional mutation. It keeps the existing argument syntax, validation order,
confirmation wording, cancellation meaning, exit codes, default reassignment,
and result text. Rich uses one Bubble Tea root with Huh v2 Confirm and Ops
Console status-table phases through the shared Terminal Experience; Plain and Automation retain
their established stream boundaries.

### Validation and confirmation

- The command remains exactly `remove <profile>` with `cobra.ExactArgs(1)`.
  It never adds profile selection, confirmation by typed name, or stdin reads
  before validation succeeds.
- Rich uses eyebrow `YCY / config cm remove`, title `Remove CM profile`, and
  subtitle `Delete one stored commit message provider`. The requested profile
  is shown only in a bounded, safe detail area.
- Before confirmation, the command runs one real `Validate CM profile` phase
  covering store construction and one `ListCMProfiles` call. Its stable ID is
  `validate-cm-profile`. Rich shows a spinner; Plain writes
  `Checking CM profile...` to stderr; Automation remains silent. A missing
  profile fails here and never reads stdin.
- A successful validation records whether the target was the current default.
  The Rich Confirm message is `Remove CM profile "<safe name>"?`, defaults to
  `No`, and uses an explicit destructive `!` visual. For a current default it
  additionally shows:

  `Removing the default selects the first remaining stored profile, or clears the default when none remain.`

  No Base URL, model, API key, or other profile rows are shown. The writer
  still receives the original profile argument.

### Removal phase and persistence

- Confirming `Yes` creates the second stable phase
  `remove-cm-profile` / `Remove CM profile`. Rich and Plain show its spinner/
  loading line (`Removing CM profile...`) even when deletion is immediate; no
  artificial delay or percentage is introduced.
- The writer is called at most once. `RemoveCMProfile` remains the atomic
  mutation boundary: deleting a non-default profile retains the current
  default; deleting the default chooses the first remaining persisted profile;
  deleting the last profile clears the default. The Command Result never
  reports the reassigned name or a `default cleared` variant.
- If the confirmation-time writer returns `removed=false`, the removal phase
  is Failed and the command returns the existing `CM profile not found: <name>`
  error. This CM command does not convert the race into idempotent success.

### Safety projection and outcomes

- Validation and writer lookup use the exact profile argument. Rich Confirm,
  phase details, Transcript, and diagnostics use a bounded, single-line,
  control-free name projection; unsafe names fall back to `Profile configured`
  or a generic `Remove CM profile?` prompt. The shared output boundary strips
  controls from the compatibility stdout result.
- The one-shot Finish mapping is:

  | Branch | Finish outcome | Command Result |
  | --- | --- | --- |
  | Writer returns `removed=true` | `Succeeded` | `Profile <name> removed` |
  | Confirmation Esc/context cancel or Confirm No | `Cancelled` | `Cancelled` |
  | Store/read failure, confirmation error, writer error, or `removed=false` | `Failed` | no stdout document |

  Success keeps the existing `Profile <name> removed` wording and never adds
  the new default name. `Finish` is called at most once; `Close` is cleanup
  only and cannot emit a second result.

### Transcript and streams

After Rich restores the primary screen, the compact semantic Transcript is
replayed on stderr. A successful default removal records:

1. `Validate CM profile` completed, detail `Profile: <safe name>` and
   `Role: Current default`
2. `Remove CM profile "<safe name>": confirmed`
3. `Remove CM profile` completed
4. `Succeeded`

Non-default success uses `Role: Configured profile`. Confirmation Esc records
`Validate CM profile` completed, `Confirmation cancelled`, and `Cancelled`;
Confirm No records `Removal declined` and `Cancelled`. Validation, writer, or
confirmation failure records only safe phase details and `Failed`. It never
records the reassigned default name, API key, Base URL, model, paths, raw
errors, spinner frames, keystrokes, or the complete result.

Plain emits `Checking CM profile...` before validation. After confirmation it
emits `Removing CM profile...` before the single stdout success result;
validation/confirmation failures leave stdout empty, while cancellation and
decline emit only stdout `Cancelled`. Plain does not replay the Transcript.

Automation preserves the existing validation-first behavior without UI:
missing/store failure returns the original error; an existing profile returns
`config cm remove requires an interactive terminal` before stdin or mutation.
Automation emits no loading, Transcript, styling, or terminal controls.

Context cancellation during `ListCMProfiles` preserves the original context
error and does not fabricate stdout `Cancelled`. Cancellation during Confirm
is a user cancellation. Once mutation starts, the writer's real success,
failure, or `removed=false` determines the outcome; the terminal never retries
or relabels it. If deletion/default reassignment succeeds but stdout writing
fails, phase and Transcript remain Succeeded and the output error is returned
without rollback or a second writer call. Rich may fall back to Plain only
before AltScreen or mutation begins.

### Evidence

Acceptance covers exact arguments/help, validation-before-stdin ordering,
Rich titles and safe names, validation/removal spinners and phase finals,
default-No destructive confirmation and default-rule guidance, default/non-
default/last-profile persistence, concurrent validation-to-removed=false
failure, all cancellation/decline locations, missing/store/confirm/writer/
context/renderer/stdout errors, Transcript order and redaction, Plain stream
ordering, Automation missing/existing side effects, at-most-once reader/
writer/Finish behavior, primary-screen restoration, control stripping, and
unchanged result/error/exit contracts.
