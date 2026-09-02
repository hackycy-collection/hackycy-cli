# Choose the run handoff terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact script and package-manager forms, project/command context,
post-selection Interaction Transcript, cancellation/failure treatment, and
terminal handoff should `run` use before releasing all streams to the raw child
process, without decorating child output or changing its exit code?

## Answer

`run` is a finite selection-and-handoff command. It discovers a package
project, obtains a script and package-manager choice, then releases all
terminal ownership before starting the selected child. The redesign preserves
project resolution, script and lockfile ordering, prompt grammar, child argv,
inherited streams, working directory, exit codes, and error semantics while
making the parent lifecycle explicit and the forms modern.

### Discovery and selection order

`run [path]` still accepts at most one project path, defaulting to cwd. Relative
and absolute path resolution is unchanged. The command reads `package.json`,
requires an object-valued `scripts` property, filters empty/non-string entries,
and preserves declaration order. Existing errors remain exact:
`No package.json found in current directory.`, `No scripts found in
package.json.`, `No runnable scripts found in package.json.`, and `Failed to
parse package.json.`

The order is fixed:

1. resolve the project and runnable scripts;
2. select one script;
3. inspect lockfiles and calculate package-manager order;
4. select one package manager;
5. build the child request `{manager, ["run", script], project directory}`;
6. release the terminal and start the child.

The manager order remains `pnpm`, `npm`, `bun`, `yarn`, with the first matching
lockfile manager moved to the front. A lockfile existence error fails the
manager-resolution step; it does not silently fall back to the default order.
The command never auto-selects a manager merely because a lockfile exists.

### Work phases and interactions

The command uses conditional truthful phases:

- `resolve-project` / `Resolve project`, covering path normalization,
  package.json reading, and script discovery;
- `resolve-package-manager` / `Resolve package manager`, covering lockfile
  checks and ordering;
- `prepare-child-command` / `Prepare child command`, covering the safe command
  summary and exact child request construction;
- `release-terminal` / `Release terminal`, the real handoff milestone that
  restores the primary screen, exits AltScreen, releases the renderer lease,
  and restores terminal modes before exec.

Script and package-manager selection remain interactions, not spinner phases.
Every reached phase exposes active and final semantic state, including fast
operations, without artificial sleeps, script-content progress, or fabricated
percentages. Phase and Transcript details contain only a safe project basename
or relative path, script label, manager, and `run <script>` summary. Control
characters are projected as visible escapes; long commands wrap or use bounded
labels.

### Rich, Plain, and Automation forms

Rich uses one Bubble Tea root in AltScreen with Huh v2 Select controls and the
Ops Console visual system. The script list keeps package.json order; each item
shows a safe script name and its command as a bounded description. The manager
list keeps the calculated order and may show the detected lockfile source or a
`default order` hint. The Live View identifies the safe project context and
script count without exposing unnecessary absolute cwd details.

Plain preserves numbered selection, `Invalid selection` retries, and the
existing `> ` prompt. A completed script selection is recorded before manager
resolution; if manager selection is cancelled, the Transcript retains the
script but not a manager. Automation never reads stdin for this command: when
either selection would be needed it returns `run requires an interactive
terminal` before child startup and without additional result output.

### Handoff and child process ownership

After both selections, the parent records a safe project/script/manager summary
and the exact `manager run script` command. It completes `Release terminal`
before invoking the child runner. The child inherits the original stdin,
stdout, stderr, and selected project working directory. The parent emits no
Bubble Tea control sequences after release, does not capture, reorder, color,
wrap, redact, or copy child output, and never inserts child logs into its
Interaction Transcript.

If terminal release fails, the child is not called; the terminal is restored as
far as possible and the original release error follows the root diagnostic
path. A missing executable or other child start failure preserves the existing
filesystem/start error contract and never claims that the script ran.

Child normal, non-zero, and signal exits remain owned by the child runner. The
parent returns ordinary exit codes unchanged, maps signals through the existing
`128 + signal` behavior, and does not append a parent `error:` diagnostic for a
child exit result. A child that starts before context cancellation is stopped
and reaped by the existing runner logic; the parent does not retry or re-enter
the Live View.

### Transcript, cancellation, and results

Rich freezes the bounded semantic ledger, restores the primary screen, replays
the Transcript to stderr, then completes any parent-owned result handling. The
Transcript contains safe project context, selected script, manager and
lockfile/default-order context, `Prepare child command` completion, and
`Release terminal` completion. It may end with `Child started`; it never
contains child stdout/stderr, full command output, keystrokes, invalid retries,
or raw package contents/absolute cwd.

Esc/Ctrl+C/`q` in the script form records `Script selection cancelled`; the
same input in the manager form records `Package manager selection cancelled`.
Both preserve the established `Operation cancelled.\n` result, exit 0, and no
child launch. Context cancellation or SIGTERM is not rewritten as a user
decline: it returns the original context/signal outcome, emits no success
result, and leaves the child unstarted when cancellation occurs before
handoff. A cancellation observed after terminal release is delegated to the
existing child runner.

`run` has no parent success document after handoff. The child owns all output
and its exit result; `run` must not add `Run completed`, duplicate a child
summary, or submit a second result. `Close` and release are idempotent and
cannot reopen AltScreen or emit new output.

### Failure and acceptance evidence

Project discovery failures stop before forms and child startup; manager
resolution failures stop before manager selection's downstream work; release
or child-start failures stop before child execution and retain root diagnostics.
Rich/Transcript use stable safe categories such as missing package, invalid
package, no runnable scripts, lockfile read, terminal release, executable
missing, and child start. Raw errors remain on the existing redacted diagnostic
path.

Acceptance covers argument/help parsing, path modes, package.json errors,
script order/filtering, lockfile priority/default order and existence errors,
Huh PTY rendering, Plain selection grammar and both cancellation positions,
context/signal cancellation, safe long/control-bearing labels, phase order,
release-before-exec, inherited stdin/stdout/stderr/cwd and byte identity,
missing executable/start errors, normal/non-zero/signal exit codes and process
cleanup, bounded redacted Transcript without child logs, Automation's
interactive boundary, exactly-once release/close, and unchanged help,
errors, output, exit, and side-effect contracts.
