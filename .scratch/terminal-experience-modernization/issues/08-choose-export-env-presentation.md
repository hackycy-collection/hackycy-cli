# Choose the export env terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact environment selection Live View, discovery/export Work Phases,
redacted Interaction Transcript, cancellation/failure treatment, JSON Command
Result, and output-file success presentation should `export env` use for every
current branch without contaminating stdout?

## Answer

`export env` uses a command-owned Ops Console Live View for finite discovery,
selection, parsing, encoding, and optional file writing. The command preserves
its existing dotenv parsing, merge precedence, output bytes, overwrite
behavior, and Automation rules. `internal/terminal` supplies forms, phases,
ledger mechanics, and capability-aware rendering; the export adapter owns all
wording and branch decisions.

### Live View and selection

- The Rich title is `Export environment` with subtitle `Export .env file
  contents as JSON`; the eyebrow is `YCY / export env`. The context shows the
  normalized input directory and `Merge base .env: on/off`. Plain Interactive
  does not add decorative title text.
- Ambiguous candidates use Huh v2 Select titled `Select environment`. Each
  option keeps the existing environment label (`default`, `local`,
  `production`, etc.) and displays its filename as description. Values remain
  the original filenames and are never shown in the transcript. Candidate
  order is the stable discovery order; `.env` is first when it is selectable,
  while merge mode excludes it from the select and reports it as the automatic
  base file. No heuristic reordering or new default is introduced.
- The selection view never shows dotenv contents, variable values, absolute
  directories, file sizes, or modification times. A long list is searchable
  and uses the normal terminal capability/degradation policy.

### Work Phases

The adapter declares this fixed phase order, omitting only a phase that is not
applicable to the branch:

1. `Resolve directory`
2. `Discover environment files`
3. `Select environment` (only when a real interactive choice is required)
4. `Read selected files`
5. `Parse and merge values`
6. `Encode JSON`
7. `Write output file` (only with `--out`)

Every observable phase submits Active and a terminal state, even when it
completes quickly; no artificial sleep or invented percentage is used.
Details may include safe relative filenames, variable counts, and the target
path, but never values or file contents. Automation consumes the same semantic
phase stream without rendering it.

### Selection, merge, and transcript

- Explicit `--env <name>` validates and selects `.env.<name>` without a form;
  a sole candidate is also selected automatically. A safe Milestone identifies
  the source as `--env`, `unique candidate`, or `user selection`, reports the
  final file order, and reports `merge=on/off`. The existing merge order is
  unchanged: `.env` is read first and the selected environment file overrides
  duplicate keys.
- Rich transcript replay retains, in event order, `Selected environment:
  <label>`, `Files: <relative names>`, `Merge: on/off`, each phase's final
  state, `Exported <N> variables`, and a final `Succeeded` outcome. With
  `--out`, the final success event is `Wrote output to <target>` instead.
  Absolute paths may be normalized but are not expanded with extra directory
  details. The transcript never contains JSON, key/value data, internal
  option values, filters, or absolute working-directory context.
- A cancelled selection records only the labelled cancellation position and
  produces the existing successful `Cancelled` result. No file is read or
  written. User cancellation, selection context cancellation, Esc, and
  `q`/`quit`/`cancel` are visually unified; cancellation after reading or
  writing has begun remains an operational failure with its original error.

### Command Result and failure rules

- Without `--out`, stdout remains one complete result containing the existing
  `Exported variables:` heading followed by the deterministic pretty JSON.
  With `--out`, the file is written first and stdout then contains one
  `Wrote output to <target>` line. TTY styling may color or symbol-emphasize
  the heading/status, but it never decorates the JSON; Plain, redirected, and
  Automation bytes remain control-free and compatible.
- The success message for `--out` is not emitted before the write. A write
  failure therefore leaves no success-looking stdout and does not promise a
  rollback of an existing target. Relative and absolute target resolution,
  directory creation behavior, and overwrite semantics stay unchanged.
- Any discovery, explicit-environment validation, read, parse, encode, or
  write failure marks its owning phase Failed and records a safe operation/path
  detail in Rich. It submits no partial JSON or success heading. Root emits
  exactly one existing, redacted `error:` Diagnostic while preserving the
  original error and exit behavior.
- An empty parsed environment is valid: `Encode` produces `{}`, stdout keeps
  the normal heading plus JSON, and the transcript may state `Exported 0
  variables` without copying the JSON.

### Branch matrix

| Branch | Selection | Live View / Transcript | stdout |
| --- | --- | --- | --- |
| No usable `.env` files | Fail during discovery | `Discover environment files` failed | Empty |
| Only `.env` | Automatic | Auto-selected base, parse/encode finals | Heading + pretty JSON |
| Only `.env.<name>` | Automatic | Auto-selected environment, parse/encode finals | Heading + pretty JSON |
| `.env` plus multiple environments, no `--merge` | Select base or environment | Selected label and one-file order | Heading + pretty JSON |
| `.env` plus multiple environments, `--merge` | Select environment; base is automatic | Base + selected order and merge on | Heading + pretty JSON |
| `--env <name>` | Explicit validated selection | Explicit source and file order | Heading + pretty JSON |
| Any successful branch with `--out` | Write before success result | Write phase and target | `Wrote output to <target>` |
| Any interactive cancellation | Stop before file I/O | Cancelled at selection | `Cancelled` |

### Adapter and evidence

The old multi-call `Presenter.Outro/Print/Cancel` projection is replaced by
semantic milestones and one `Finish` call. Successful stdout content is
assembled by the command adapter before submission; cancellation is one
cancelled finish; phases and selection summaries are not duplicated into the
result. Required acceptance evidence includes Rich PTY selection/search/
cancellation and transcript replay, Plain validation retry and no-control
output, Automation unique-selection success and ambiguous no-side-effect
failure, explicit `--env`/`--merge` ordering, empty JSON, every failure phase,
target preservation on write failure, stable heading+JSON bytes, secret/value
non-disclosure, and exactly one result submission.
