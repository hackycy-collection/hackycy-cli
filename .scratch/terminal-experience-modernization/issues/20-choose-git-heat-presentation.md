# Choose the git heat terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact Git inspection Work Phases, empty state, ranked hot-path report,
legend, failure presentation, and Command Result should `git heat` use across
relative-time and absolute-time modes without changing report semantics?

## Answer

`git heat` is a finite, read-only inspection command with no questions. It
uses three truthful Work Phases, restores the primary screen, replays a compact
Interaction Transcript, and then submits one complete Git Heat Report as its
Command Result. The redesign changes Rich presentation and lifecycle
visibility without changing input normalization, Git inspection, aggregation,
sorting, time, query, empty-result, stdout, error, signal, or exit semantics.

### Work Phases

- Validate Cobra arguments and normalize the mutually exclusive commit/day
  range before opening a Live View. Invalid flags, targets, sorts, ranges, and
  unexpected arguments continue through the existing root error path without
  a phase or Transcript.
- Declare one immutable phase catalog in this order:
  `locate-repository` / `Locate Git repository`, `read-git-history` /
  `Read Git history`, and `rank-hot-paths` / `Rank hot paths`.
- `Locate Git repository` uses `Locating repository` while active and
  `Repository located` when completed. It does not expose the absolute root.
- `Read Git history` uses the normalized range, such as
  `Reading last 20 commits` or `Reading last 7 days`, while active and
  `Read <N> commits` when completed.
- `Rank hot paths` names the selected target and sort, such as
  `Ranking files by change count`, while active and uses
  `Ranked <N> files` or `Ranked <N> directories` when completed. It does not
  expose the raw query, Git arguments, timings, or per-record progress.
- Rich renders every reached phase, including fast operations, through at
  least one real model/view cycle without sleeping. Failed or cancelled work
  leaves later phases Pending and attaches the terminal state to the phase
  where it occurred.

Plain Interactive emits one control-free stderr loading line as each phase is
reached, for example `Locating repository...`, `Reading last 20 commits...`,
and `Ranking files by change count...`. It emits the applicable failed or
cancelled terminal state but does not repeat three completed lines before a
successful report. Automation emits no phases, Transcript, or terminal control
sequences.

### Rich report

- Use the eyebrow `YCY / git heat`, title `Repository heat`, and the existing
  summary semantics: safe repository name, selected range, actual commit count
  for a day range, and selected file or directory count. Do not add the
  repository root, duration, Git implementation details, or new statistics.
- A wide terminal uses an aligned table with the existing rank, `Changed at`,
  `M A D R C`, and `File` or `Directory` fields. A narrow terminal renders
  each row as a compact labelled multi-line record with the path on its own
  wrapping line. Every row remains present; there is no result pagination,
  row limit, horizontal semantic truncation, or hidden narrow-mode field.
- Preserve the current row order exactly. Path sorting retains its existing
  root/immediate-directory and locale behavior; count sorting retains its
  existing total-occurrence ordering and path tie-break. The presentation
  never derives a new heat score or bar from those totals.
- Mark the latest row with a green `▲ latest` projection and the earliest row
  with a yellow `▼ earliest` projection while retaining textual labels and the
  existing first-row-on-ties rule. When one row is both, continue to show only
  `latest`. Active `M`, `A`, `D`, `R`, and `C` letters receive distinct vivid
  colors; an absent kind remains a muted `-`. Color is never the sole signal.
- `--query` highlights only the exact matched path substrings using the
  secondary magenta focus treatment, bold text, and a contrasting background.
  Preserve case-insensitive, Unicode-boundary-safe, non-overlapping matching.
  A query never filters, hides, reorders, or changes a row's semantic value;
  no match leaves the complete report unchanged.
- Keep `Changed at` and the existing absolute or relative projection.
  Absolute mode retains Git's `YYYY-MM-DD HH:mm:ss` label. Relative mode uses
  the command's one captured result clock for every row, including current
  `just now`, past, and future wording; it does not update while rendering.
  An unknown time remains `-`.
- End with the complete existing legend for latest, earliest, modified, added,
  deleted, renamed, and copied. Keep the layout compact and unframed rather
  than surrounding report sections with decorative panels.

Rich renders ordinary Unicode path text as-is, but projects newline, tab, ESC,
and other control characters as visible escapes before terminal rendering.
Long paths wrap and remain semantically complete. Query matches are computed
against the original path and mapped onto its safe display projection. This
Rich-only safety projection does not alter parsed paths, aggregation, sorting,
or the established Plain/Automation result bytes.

### Command Result and empty state

- Plain and Automation preserve the current complete stdout document: the
  `HACKYCY CLI` title, range/count summary, tabular header and rows, and legend.
  Rich may use the responsive projection above, but it retains all the same
  semantic fields and rows.
- A successful non-empty command calls `Finish(Succeeded, document)` exactly
  once after all three phases complete. The complete report is written to
  stdout only after AltScreen restoration and Transcript replay; no partial
  title, summary, table, row, or legend is written while work is in progress.
- Zero commits or no changed file rows remains a successful empty result with
  exit code 0. All three phases complete, and stdout remains exactly
  `No changed files found in the selected range.\n`, including when the target
  is directories. Do not split empty commits from empty changes or introduce a
  different directory message.
- An empty Rich Live View uses a yellow, symbol-paired information state rather
  than presenting the absence of data as an operational failure. It still
  submits the same one-line Command Result once.

### Transcript, failure, and cancellation

The Rich Interaction Transcript contains the three reached phase final states,
one bounded result summary, and the final outcome. A non-empty summary uses
`Ranked <N> files from last 20 commits` or the directory/day equivalent; a day
range may also include the actual commit count. An empty success records
`Found 0 changed files` and `Succeeded`. It never copies report rows, paths,
the raw query, repository root, legend, Git arguments, raw errors, or the full
Command Result.

- Repository discovery errors fail `Locate Git repository`. Git process
  startup/exit errors and malformed log records fail `Read Git history`.
  Local aggregation or report-construction errors fail `Rank hot paths`.
- A failure restores the terminal, replays only safe phase context and
  `Failed`, calls `Finish(Failed, nil)` at most once, and returns the original
  error for the existing root diagnostic and exit behavior. It never emits a
  partial or success-looking stdout report.
- Ctrl-C, SIGTERM, or context cancellation marks the active phase Cancelled,
  stops the existing Git process group, restores the terminal, and replays the
  cancellation position plus `Cancelled`. It emits no partial report and does
  not replace the existing typed signal outcome, `128 + signal` exit code, or
  ordinary context-cancellation contract. A real operation failure continues
  to take precedence over a simultaneous cancellation request.
- Renderer, Transcript, cleanup, and stdout failures follow the shared
  Experience ordering and error-joining rules. `Close` remains idempotent
  cleanup and cannot manufacture or retry a result.

### Evidence

Acceptance must cover Rich PTY phase visibility, fast phase transitions,
primary-screen restoration, Transcript ordering, wide tables, narrow records,
long and control-bearing paths, files/directories, path/count sort,
limit/day ranges, absolute/relative time, latest/earliest ties, every supported
change kind, and Unicode query highlighting. It must also cover zero commits,
empty commits, no changed files, non-repository/unborn/bare repositories, Git
startup/exit/log-parse failures, Ctrl-C/SIGTERM/context cancellation without an
orphan or partial stdout, Plain loading/result ordering, Automation silence,
stdout write failure, at-most-once Finish, and unchanged command help, result,
error, signal, and exit contracts.
