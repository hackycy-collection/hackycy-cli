# Choose the git pulse terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact date and author forms, repository-scan/fetch Work Phases,
partial-failure treatment, semantic Interaction Transcript, empty/cancelled
states, and potentially large commit-tree Command Result should `git pulse`
use without duplicating the report or changing filtering semantics?

## Answer

`git pulse` is a finite workspace-inspection command with two conditional
interactions and one potentially large Git Pulse Report. It uses four truthful
Work Phases, preserves the current scan/date/fetch/author order, restores the
primary screen, replays a bounded semantic Interaction Transcript, and submits
one complete Command Result. The redesign does not change directory discovery,
Git concurrency, date boundaries, author filtering, report ordering, prompt
cancellation, stdout documents, signal outcomes, or exit codes.

### Workflow and Work Phases

Cobra argument-count and `--days` parsing errors remain before the Terminal
Experience. Once the leaf begins, declare and execute these phases in order as
their prerequisites become available:

1. `prepare-workspace` / `Prepare workspace` resolves the default, relative,
   or absolute root, validates that it is a directory, and verifies Git with
   the existing `git --version` call. Use `Checking workspace and Git` while
   active and `Workspace ready` when completed.
2. `scan-repositories` / `Scan repositories` walks the root with the existing
   exclusions and discovery rules. Its replaceable detail is
   `Found <N> repositories` and may include only the most recently discovered
   safe relative repository path. Its completed detail retains the count.
3. `fetch-commits` / `Fetch commits` starts only after a date is available.
   Keep at most five Git log children active. Its replaceable detail is
   `[<done>/<total>] Reading <safe relative repository>` and its completed
   detail is `Read <successful> of <total> repositories`.
4. `build-commit-tree` / `Build commit tree` starts only after any required
   author selection succeeds. Use `Grouping commits by repository` while
   active and `Built report with <C> commits in <R> repositories` when
   completed.

Every reached phase gets at least one real Rich model/view cycle even when it
completes quickly; no sleep is inserted. Directory traversal ticks, every
visited path, individual commit records, process arguments, raw errors, and
timings are not progress. No-repository and no-commit branches stop before
their downstream phases instead of pretending those phases ran.

The current behavioral order is intentional and remains exact: scan before
asking for a date so an empty workspace needs no interaction; fetch before
asking for authors because the available identities come from fetched commits.

### Date interaction

When `--days` is absent, use a Huh v2 Select labelled `Select date range:` with
these values and order:

1. `Today` / `1`
2. `Yesterday` / `2`
3. `Last 3 days` / `3`
4. `Last 7 days` / `7`
5. `Last 30 days` / `30`

Do not add a custom range, date picker, or implicit default. A completed answer
shows the selected label and resulting local calendar boundary. Preserve the
inclusive start-of-day calculation in the injected clock's location, including
DST behavior.

An explicit `--days` skips the form. Preserve the permissive decimal-prefix
parser and the accepted positive, zero, and negative values exactly; the
presentation records the normalized integer and calculated boundary without
reinterpreting it as one of the preset labels.

### Author interaction

- Zero or one distinct author skips the form and retains every commit. More
  than one uses a searchable Huh v2 MultiSelect labelled
  `Filter by authors:`.
- Preserve American English collation for option order. With two or three
  authors, every option starts selected. With more than three, none starts
  selected. At least one author remains required; an empty submission or
  malformed Plain selection retries without entering the Transcript.
- Interaction values remain the exact original author strings and filtering
  remains exact equality. Labels use bounded, control-free display projections.
  If distinct identities collapse to the same safe label, append stable
  ordinals to distinguish the labels without merging values.
- Record completed selected labels in selection order. Never record unchecked
  authors, cursor movement, search text, invalid retries, or partial input.

Cancelling either form submits the existing `Operation cancelled.\n` Command
Result with `Finish(Cancelled, document)`, returns nil, and exits 0. The
Transcript distinguishes `Date range selection cancelled` from
`Author filter cancelled`; no fetch occurs after date cancellation and no
partial report is built after author cancellation.

### Partial discovery and fetch warnings

Unreadable directories remain non-fatal and scanning continues. Count them and
retain a deterministic, bounded projection of safe relative paths for
presentation. `Scan repositories` still enters the shared `Completed` state;
a separate warning milestone says `Skipped <N> unreadable directories` and
lists up to five paths followed by `... and <N> more` when needed. Deliberately
excluded directory names, symlinks, `.git` files, and bare repositories remain
normal discovery rules and never become warnings.

A Git startup error or nonzero exit for one discovered repository remains a
Repository Fetch Omission, not a command failure. Preserve successful commits,
the five-child concurrency limit, exit code 0, and the existing stdout result.
`Fetch commits` completes with `Read <successful> of <total> repositories`,
then a separate warning milestone reports the omission count and up to five
deterministically ordered safe relative repository paths. It never exposes raw
Git stderr or per-repository error text. `Completed with warning` is a visual
description of a Completed phase plus this warning milestone, not a new shared
phase state.

If every repository fetch is omitted, preserve the successful
`No commits found in the specified date range.\n` result while clearly stating
that all repositories were skipped. Rich records the warning in its
Transcript, Plain writes it as a safe stderr warning, and Automation writes one
control-free Diagnostic warning; none changes stdout or the exit code.

### Git Pulse Report

- Rich uses the eyebrow `YCY / git pulse`, title
  `Workspace commit activity`, and the existing commit/repository count
  summary. It outputs the complete report to the primary screen after
  Transcript replay, never inside a post-work AltScreen viewport.
- Preserve existing report membership and ordering: repository groups use the
  current American English path collation, and commits within a repository use
  descending formatted date with stable ties. Author selection changes only
  membership through the existing exact filter.
- Wide terminals use a compact tree row with timestamp, author, and subject.
  Narrow terminals put timestamp and author on one line and wrap the subject on
  the following indented line. Repository headings retain the name, parent
  path, and commit count. Every repository and commit is emitted; there is no
  pagination, maximum count, semantic truncation, or hidden narrow-mode field.
- Rich displays newline, tab, ESC, and other terminal control characters in
  repository paths, authors, and subjects as visible escapes. Ordinary Unicode
  remains intact, and long values wrap. The same safe projection is used in
  forms, warning milestones, and Transcript entries.
- Plain and Automation retain the exact established stdout projection,
  including its leading newline, summary, repository basename and parent path,
  ASCII connectors, commit fields, blank lines, ordering, and final newline.
  Rich safety and responsive layout do not mutate parsed data or these durable
  bytes.

The full report is composed before output and submitted once with
`Finish(Succeeded, document)`. No title, summary, repository, or commit is
written to stdout while scan, fetch, filtering, or grouping is incomplete.

### Empty results and Automation

- No discovered repositories completes Prepare and Scan, asks no questions,
  starts no Fetch or Build phase, calls
  `Finish(Succeeded, "No Git repositories found.\n")`, and exits 0.
- No fetched commits completes Fetch, asks no author question, starts no Build
  phase, calls
  `Finish(Succeeded, "No commits found in the specified date range.\n")`, and
  exits 0. Any scan/fetch warnings remain distinct stderr context.
- Form cancellation follows the cancellation result described above. These
  three established result documents are not combined into a generic empty
  result or moved exclusively into the Transcript.

Automation preserves the exact point at which an interaction becomes
necessary. With no repositories it succeeds even without `--days`. With
repositories and no `--days`, it scans and then returns
`git pulse requires an interactive terminal` at the date form boundary. With
explicit days and multiple authors, it fetches and returns the same error at
the author form boundary. Explicit days with at most one author completes
without interaction. Automation never reads stdin, invents defaults, renders
phases/forms, emits a Transcript, or writes terminal controls; safe partial
warnings remain Diagnostic Records and a rejected interaction emits no stdout
result.

### Interaction Transcript

After Rich restores the primary screen, replay one ordered, bounded semantic
Transcript containing:

- every reached Work Phase's final state and final safe detail;
- the selected date label, or explicit day integer, and calculated boundary;
- selected safe author labels, or `Author filter: All commits (<N> authors)`
  when the interaction was skipped after commits existed;
- unreadable-directory and Repository Fetch Omission warning counts with their
  bounded safe relative-path summaries;
- `Found <C> commits in <R> repositories`, the exact empty-state meaning, or
  the labelled cancellation position; and
- the final `Succeeded`, `Cancelled`, or `Failed` outcome.

Do not include absolute workspace paths, scan ticks, fetch completion order,
raw author values, unselected authors, Git argv/stderr, raw errors, full
repository/commit rows, animation frames, or the Command Result. The shared
ledger's event and field bounds handle unusually large author selections; it
truncates at semantic event boundaries without leaking partial unsafe text.

### Failures, signals, and stream ordering

- Working-directory, target-stat, non-directory, and Git-availability errors
  fail `Prepare workspace`. Context or renderer termination during traversal
  fails or cancels `Scan repositories`; ordinary unreadable children use the
  non-fatal warning rule. Whole-operation tracker/renderer errors fail
  `Fetch commits`; repository-local Git errors remain omissions. Grouping,
  sorting, or result-construction errors fail `Build commit tree`.
- A form renderer or input I/O error fails at the labelled date or author
  interaction. Invalid Plain answers remain retries. User form cancellation is
  the established exit-0 cancellation outcome, not a failure.
- Ctrl-C, SIGTERM, and context cancellation request cancellation at most once,
  stop the currently active Git process groups, wait for them, and stop
  scheduling more repositories. Restore the terminal and replay the safe
  cancellation position without emitting a partial report. Preserve the typed
  signal outcome, `128 + signal` exit code, ordinary context error semantics,
  and real-error precedence.
- On failure, call `Finish(Failed, nil)` at most once, restore the terminal,
  replay safe phase/form context, and return the original error for the root's
  single redacted Diagnostic Record. `Close` remains idempotent cleanup and
  cannot synthesize or retry a result.
- Rich completion follows the shared restore, Transcript, deferred-diagnostic,
  then stdout-result ordering. Plain writes its control-free introduction,
  workspace, reached phase lines, prompts/retries, and warnings to stderr and
  does not replay a Transcript. Automation follows the rules above.

### Evidence

Acceptance must cover Rich PTY execution across all four phases, the date
Select, searchable author MultiSelect, skipped interactions, form defaults,
required validation, cancellation positions, primary-screen restoration, and
Transcript order. Report evidence covers wide/narrow layouts, large results,
long fields, Unicode, terminal controls, repository and commit ordering, and
complete one-shot output.

Behavioral evidence covers default/relative/absolute roots, nested
repositories, exclusions, symlinks, bare layouts, unreadable-directory
warnings, some/all Repository Fetch Omissions, the five-child ceiling, and
cancellation without further scheduling or orphaned processes. Date evidence
covers presets, permissive prefixes, zero, negative values, local calendar
boundaries, and DST. Cross-mode evidence covers every empty/cancel result,
missing paths, unavailable Git, renderer/context/Ctrl-C/SIGTERM failures,
Plain ordering, Automation interaction timing and warning diagnostics, stdout
write failure, at-most-once Finish, unchanged help/stdout/error/exit contracts,
and absence of terminal controls where prohibited.
