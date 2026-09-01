# Choose the rm terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact explicit confirmation, smart action and target forms, discovered
path context, deletion Work Phases, partial-failure treatment, semantic
Interaction Transcript, and final result should `rm` use while preserving all
safety defaults, force behavior, cancellation, and partial-commit semantics?

## Answer

`rm` is a finite destructive cleanup command with two command-owned routes:
explicit path removal and smart project-artifact cleanup. The redesign keeps
path resolution, discovery, force behavior, confirmation defaults, concurrent
deletion, partial-success semantics, stdout documents, errors, exit codes, and
side effects unchanged while adding truthful Rich phases, modern Huh forms,
safe path projections, and semantic Transcript replay.

### Explicit paths

With operands, the command resolves each relative path against the absolute
working directory and cleans each absolute operand directly. It preserves input
order and legacy permissiveness for duplicate paths, `.`, directories, files,
symlinks, dangling symlinks, and paths outside the working directory. Planning
classifies missing paths without failing. Missing operands continue to emit the
existing `not found, skipping` notices; if no existing target remains, the
successful result is `No valid paths to delete.` and no prompt or mutation is
performed.

Without `--force`, a Huh v2 destructive Confirm keeps the default-negative
behavior and `[y/N]` Plain syntax. Rich shows the normalized absolute targets,
their file/directory/symlink type where known, target count, and an explicit
warning that recursive deletion removes all contents. It adds a prominent
warning for cwd or parent-scope targets, outside-cwd targets, and duplicate
operands, but does not reject or rewrite those targets. `--force` bypasses this
confirmation completely and deletes the planned targets directly.

### Smart cleanup

With no operands, the first interaction remains a single-select over the six
existing actions in their exact order: root `dist`, root `node_modules`,
recursive monorepo `dist`, recursive monorepo `node_modules`, root lockfiles,
and AI-agent configuration directories. The first action remains the default.
The scan preserves the existing root existence checks, recursive depth default
of 5, explicit depth parsing and negative-depth behavior, and the rules that
skip hidden directories, VCS directories, and `__pycache__` while retaining
only directories for recursive matches.

When targets are found and `--force` is absent, a searchable Huh v2 MultiSelect
keeps discovery order and selects every target by default. Labels are safe
paths relative to cwd; internal values remain the original absolute paths.
Plain keeps numbered comma-separated selection, `all`, `none`, and invalid
input retries. `--force` skips target selection but does not skip action
selection or scanning. An empty scan remains the successful pair
`No targets found.` and `Nothing to clean.` with no deletion.

### Work Phases

The command uses conditional, truthful phases. Interactions are not phases:

- explicit route: `resolve-explicit-targets` / `Resolve explicit targets`,
  then `delete-selected-paths` / `Delete selected paths` after confirmation;
- smart route: `scan-cleanup-targets` / `Scan cleanup targets`, then
  `delete-selected-paths` / `Delete selected paths` after optional target
  selection.

The action select, target multiselect, and explicit confirmation are semantic
interactions attached to their command-owned route. Each phase enters Active
and a terminal Completed, Failed, or Cancelled state whenever reached,
including immediate operations, without artificial sleep, per-path spinners,
scan ticks, goroutine counts, or fabricated percentages. Phase details contain
only route, action, target counts, and bounded safe path summaries. They do not
expose cwd internals, hidden traversal paths, implementation identifiers, or
raw filesystem errors.

### Deletion and partial success

Deletion retains the current concurrent `RemovePath` calls and deterministic
input-order result aggregation. A completed deletion phase reports requested,
succeeded, and failed counts. Individual failures remain partial deletion
records rather than command failures: Plain and Automation preserve the
existing `Deleted N item(s)`, `skipped`, and `Done!` document and exit 0,
including when every target fails. Rich presents the same facts under
`YCY / rm`, with `Paths removed` for explicit mode or `Cleanup complete` for
smart mode. Stable failure categories such as `permission`, `not-found`,
`path`, and `filesystem` replace raw errors in Rich and Transcript; raw errors
remain available only to the existing redacted diagnostic path.

Successful `RemovePath` means only that the remover returned nil; the command
does not claim more detailed filesystem state. A failed target is described as
possibly still present. There is no retry, rollback, serializing of the
existing work, or change to the number of remover calls.

### Outcomes and stream ordering

The durable result remains route-specific and is submitted once:

- all explicit operands missing: `No valid paths to delete.`;
- smart scan empty: `No targets found.` followed by `Nothing to clean.`;
- empty smart selection: `Nothing selected.`;
- user cancellation or decline: `Cancelled.`;
- deletion success or partial success: `Deleted N item(s)` and `Done!`;
- all other planning/scan/remover orchestration failures: no success-looking
  result and the original error is returned.

Rich owns one Bubble Tea root in AltScreen. It freezes phases and the semantic
ledger, restores the primary screen, replays the bounded redacted Transcript
to stderr, then submits the complete Command Result to stdout exactly once.
Plain Interactive emits control-free loading/status lines and the existing
durable result without replaying a second Transcript. Automation emits no
AltScreen, forms, phases, Transcript, styling, or terminal controls.

### Transcript and safety projection

The Rich Transcript records only semantic events: route, selected smart action,
target and missing counts, bounded safe target summaries, confirmation or
selection outcomes, scan completion, deletion counts/failure categories, and
the final Succeeded, Cancelled, or Failed outcome. It does not copy complete
path lists, cwd, hidden traversal, keystrokes, cursor/search activity, invalid
answers, animation frames, or raw errors.

Mutation always receives the original resolved absolute paths. Presentation
uses an independent safe projection: control characters such as newline, tab,
and ESC become visible escapes; ordinary Unicode remains intact; long values
wrap or use bounded labels; and unsafe relative conversion falls back to a
generic safe label. Colors and `!` symbols emphasize risk but never replace
explicit destructive wording.

### Cancellation, context, and signals

If context is already cancelled before the command starts, it preserves the
current no-work behavior: no introduction, cwd lookup, prompt, or mutation.
Cancellation during path planning, smart scanning, or an interaction stops
before later phases and returns the existing context/signal outcome without a
success result. Esc/Ctrl+C/`q` from a user form is a normal exit-0 interaction
cancellation with `Cancelled.` and no deletion. A context cancellation observed
after confirmation but before deletion prevents the delete phase from starting.

Once concurrent deletion has started, already-issued remover calls are not
claimed stoppable. The command waits for the current batch to return, merges
the real success/failure facts, and retains the existing deletion result
contract. It never rewrites a signal as user decline or introduces rollback.
When a real remover or scan error races with cancellation, the real operation
error takes precedence. Typed signal outcomes and their existing exit codes
remain intact.

### Automation and acceptance evidence

Automation keeps the current boundaries: explicit `--force` deletes directly;
explicit non-force returns `rm requires an interactive terminal` before stdin
or deletion; no-operand mode requires the smart action interaction even with
`--force`, while `--force` only skips the target multiselect. Missing explicit
paths and empty smart results retain their current successful stdout behavior.

Acceptance covers explicit and smart routes, all path kinds and risk scopes,
duplicates, missing operands, force/default-negative confirmation, six action
options, depth edges, discovery skip rules, target multiselect defaults and
Plain grammar, empty/no-selection/cancelled outcomes, concurrent partial
deletion, failure categories and safe projections, Rich PTY AltScreen restore
and Transcript ordering, cancellation at every stage, signal behavior, Plain
result bytes, Automation zero-side-effect and control-free behavior, exactly-
once result submission, and unchanged help, errors, exit codes, and side
effects.
