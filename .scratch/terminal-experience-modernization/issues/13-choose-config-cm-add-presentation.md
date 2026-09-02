# Choose the config cm add terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact profile, URL, model, and API-key form; validation display;
persistence Work Phases; redacted Interaction Transcript; and
success/cancellation/failure result should `config cm add` use without
changing prompt order, defaults, validation, or overwrite behavior?

## Answer

`config cm add` is a four-question interactive profile setup followed by one
conditional persistence phase. It uses one Bubble Tea root with Huh v2 child
forms and the Ops Console status-table layout through the shared Terminal
Experience. The command
owns prompt wording, validation, safe projections, overwrite semantics, and
the result document; the terminal layer owns form rendering, phase mechanics,
redaction enforcement, and capability degradation. Existing prompt order,
placeholders, input values, validation errors, cancellation meaning, default
profile behavior, encryption, and result text remain unchanged.

### Form structure

- Rich uses the title `Add commit message profile`, eyebrow
  `YCY / config cm add`, and subtitle `Configure an OpenAI-compatible
  provider`.
- The observable question order is fixed: `Profile name`,
  `OpenAI-compatible base URL`, `Model`, and `API key`. Rich advances through
  one Huh child form at a time and does not add a review confirmation or allow
  reordering/skipping.
- Placeholders remain exactly `e.g. openai, deepseek, work`,
  `https://api.openai.com/v1`, and `gpt-4.1-mini`; API key has no default or
  placeholder value. No environment variable or existing profile is
  auto-filled.
- The `STATE / PHASE / DETAIL` table rows are `Identity`, `Endpoint`, `Model`,
  and `Credential`. The active step uses `◆`, completed steps `✓`, and
  pending steps `○`. The active form uses B's bottom focus rule and there is
  no persistent left rail. API key input is always masked and never shows a
  preview, length, prefix, strength, or echo.

### Validation and safe projections

- Huh keeps validation on the current field and retries in place. The exact
  existing errors remain `Name is required`, `Name cannot contain spaces`,
  `Base URL is required`, `Model is required`, and `API key is required`.
  Invalid attempts, keystrokes, and partial values are not Transcript events,
  Work Phases, or writes.
- The writer receives the original input values. `appconfig` continues to
  trim Base URL whitespace, remove trailing `/`, encrypt the API key, set the
  first profile as default, preserve an existing default on replacement, and
  silently overwrite a same-name profile.
- Rich review text, Transcript fields, and diagnostic context use bounded,
  single-line, control-free projections. Profile names use a safe name
  projection; unsafe names fall back to `Profile configured`. Base URLs retain
  safe scheme/host/path while removing userinfo, query, and fragment; unsafe
  values become `Base URL configured`. Models use a safe single-line
  projection and fall back to `Model configured`. These projections never
  alter the stored profile identity or persistence values.

### Work Phases and lifecycle

- The Experience opens before store construction for Rich and Plain. Automation
  is rejected first, before store creation or stdin reads, with the unchanged
  error `config cm add requires an interactive terminal`.
- The immutable phase catalog contains exactly:

  | ID | Name |
  | --- | --- |
  | `collect-cm-profile-details` | `Collect CM profile details` |
  | `save-cm-profile` | `Save CM profile` |

  `Collect` enters Active before the first question and remains active while
  all four fields are answered. It completes only after the fourth valid
  answer; field validation does not change its state. A pre-save cancellation
  marks it Cancelled and never creates the save phase. A store-construction
  failure before the first question marks `Collect` Failed and starts no
  interaction or save phase.
- After the fourth valid answer, Rich may show a non-sensitive review summary
  (`Profile`, safe `Base URL`, safe `Model`, and `API key: [redacted]`) and
  immediately starts `Save CM profile`. The save phase always shows a spinner
  and ends in Completed or Failed without an artificial delay or preflight
  read.

### Outcomes and streams

The command submits exactly one semantic outcome:

| Branch | Finish outcome | Command Result |
| --- | --- | --- |
| Writer succeeds, including same-name replacement | `Succeeded` | `Profile <name> added` |
| Any field Esc/configured cancel/context cancel before save | `Cancelled` | `Cancelled` |
| Store, validation, or writer failure | `Failed` | no stdout document |

The success wording remains `Profile <name> added` even when the writer
replaced an existing profile. Rich may add a `✓` and color, but does not change
the semantic result or add URL/model fields to stdout. Plain emits the current
control-free prompts and validation retries, with lifecycle diagnostics
`Collecting CM profile details...` and `Saving CM profile...`; stdout contains
only the single success or cancellation result. Automation emits no prompts,
Transcript, or controls and is rejected before side effects.

If the writer fails, `Collect CM profile details` remains Completed,
`Save CM profile` is Failed, stdout stays empty, and the root emits one
redacted diagnostic while returning the original error. A context cancellation
after the writer starts is the writer's real result/failure, not a fabricated
user cancellation; a successful writer result remains success even if the
context is cancelled afterward. Rich restores the primary screen before
replaying the Transcript and then releases diagnostics and emits the result
once.

### Interaction Transcript

On successful Rich completion, the post-AltScreen Transcript records in order:

1. `Profile name: <safe name>`
2. `Base URL: <safe URL>`
3. `Model: <safe model>`
4. `API key: [redacted]`
5. `Collect CM profile details` completed
6. `Save CM profile` completed
7. `Succeeded`

For cancellation, completed safe fields are followed by the cancelled field
marker (for example `Model cancelled` or `API key cancelled`), then
`Collect CM profile details` cancelled and `Cancelled`. A partially typed API
key is never recorded. Save failure retains completed safe fields and the
redacted credential marker, records the completed collect phase, failed save
phase, and `Failed`. Store failure before form collection records only a safe
failure state. Raw errors, paths, invalid attempts, keystrokes, animation
frames, plaintext/ciphertext API keys, and unsafe URLs never enter the ledger.

### Evidence

Acceptance covers Automation's pre-store rejection, Rich PTY title and exact
four-question order, placeholders/defaults, Ops Console table state
transitions,
validation retries, password masking, safe review projections, fast save
loading/completion, first-default and same-name overwrite behavior,
field-specific cancellation, save/store/context failures, Transcript ordering
and redaction, one-shot Finish, primary-screen restoration, Plain diagnostic
and result ordering, control-free degradation, encrypted persistence, and
unchanged command/help/exit/result contracts.
