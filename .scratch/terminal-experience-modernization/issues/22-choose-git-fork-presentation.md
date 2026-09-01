# Choose the git fork terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact repository resolution, default-branch, archive, extraction, and
overwrite Work Phases; repository/destination context; cancellation/failure
treatment; Interaction Transcript; and final Command Result should `git fork`
use without changing provider, filesystem, or overwrite semantics?

## Answer

`git fork` is a finite, potentially destructive repository-acquisition command.
It keeps its archive-first and shallow-clone fallback semantics, but makes each
real stage visible in a vivid Rich Live View, preserves a semantic redacted
Interaction Transcript after AltScreen exits, and submits its existing durable
Command Result exactly once. Plain Interactive and Automation remain compatible
with the current stdout, error, exit, and side-effect contracts.

### Work Phases and route

The command declares one ordered phase catalog and enters only phases whose
prerequisites are met:

1. `resolve-repository` / `Resolve repository` parses the repository reference
   and resolves its provider, host, owner, name, and optional explicit ref.
2. `inspect-destination` / `Inspect destination` resolves the destination from
   the current working directory and checks whether it has entries.
3. `replace-destination` / `Replace destination` runs only for a nonempty
   destination after confirmation and represents the confirmed recursive
   removal of all existing contents.
4. `resolve-default-branch` / `Resolve default branch` runs only when no
   explicit `#ref` was supplied.
5. `download-archive` / `Download archive` runs when a ref is available.
6. `extract-archive` / `Extract archive` runs only after archive download
   succeeds and retains the existing strip-one and entry-type behavior.
7. `clone-fallback` / `Clone fallback` runs after default-branch lookup fails,
   or after archive download/extraction fails, using the same resolved ref (or
   the remote default when no ref was resolved).
8. `remove-git-metadata` / `Remove Git metadata` runs after a successful clone
   and removes `<destination>/.git` exactly as today.

`Project ready` is a completion milestone, not a fabricated Work Phase. Every
entered phase has an active and final state, including fast phases, without
sleeping or inventing an unmeasurable percentage. Phase details are safe and
low-noise: repository owner/name, provider, ref, and destination label only.
They never expose HTTP URLs, authentication data, Git argv, TAR entries, raw
stderr, timings, or request-level activity.

The route remains archive-first. An explicit ref skips default-branch lookup.
Without a ref, a default-branch failure goes directly to a clone using the
remote default. Download or extraction failure falls back to a clone at the
same ref. A successful clone is followed by Git metadata removal; no new
rollback or cleanup policy is introduced.

### Destination and overwrite interaction

The destination defaults to the repository name. Relative destinations are
resolved against cwd and absolute destinations are cleaned directly. Empty or
missing destinations are not confirmed or pre-removed. A `ReadDir` error keeps
the existing permissive behavior (continue as if confirmation is unnecessary)
but emits a safe warning in interactive presentation.

For a nonempty destination, Huh v2 displays the normalized absolute path and a
clear warning that recursive replacement deletes all existing contents. The
default remains Yes with `[Y/n]` semantics. A No answer is recorded as
`Destination replacement declined`; Esc, Ctrl+C, `q`, or equivalent prompt
termination is recorded as `Destination replacement cancelled`. Both preserve
the current outcome: stdout contains `Cancelled`, exit code is 0, no removal or
network request occurs, and the destination is recorded as `Destination
unchanged`.

Automation never reads stdin. If a destination is known to be nonempty, it
returns `git fork requires an interactive terminal` before deletion, network
access, or result output. Empty, missing, or unreadable destinations retain the
existing non-interactive execution boundary.

### Rich, Plain, and Automation presentation

Rich uses one Bubble Tea root in AltScreen with Huh v2 for the overwrite form
and the Signal Rail visual system. Loading is always visible while a phase is
active. On the final result it uses the eyebrow `YCY / git fork` and title
`Project acquired`, with a compact result view showing the safe repository and
provider projection, final ref, `archive` or `clone` acquisition, fallback
category when applicable, whether history was removed, and the destination.

Plain Interactive may use control-free line-oriented phase status and warning
messages, but preserves the existing redacted stdout document and error text.
Automation emits no AltScreen, forms, phases, Transcript, or terminal control
sequences and preserves its current stdout/error behavior.

### Interaction Transcript and durable result

When Rich leaves AltScreen, it replays one ordered, bounded, semantic
Interaction Transcript to stderr. It contains completed overwrite questions and
redacted answers when present; final states for reached phases; fallback
milestones; the final outcome; and the applicable disk fact. It never replays
keystrokes, cursor movement, invalid answers, secret input, animation frames,
full result rows, Git argv, URLs containing credentials, or raw provider/Git
errors. Only safe host, owner/name, provider, alias, ref, and destination
projections are allowed.

Failure and cancellation entries describe facts rather than promises:

- `Destination unchanged` when mutation has not happened.
- `Existing destination removed` after confirmed replacement completes.
- `Destination may contain partially extracted files` after extraction can
  have written before failing or cancellation.
- `Destination may contain a partial clone` when clone did not complete.
- `Project files created; Git metadata remains` when clone succeeds but `.git`
  removal fails.

The complete Command Result is written to stdout only once, after the final
phase has been settled, AltScreen has been restored, and the Rich Transcript
has been replayed. Successful output retains the established `Resolved`,
conditional `Branch`, redacted fallback diagnostics, acquisition outcome, and
`Done! Project created at ...` document. Large output is never duplicated into
the Transcript. A declined/cancelled overwrite submits the existing one-shot
`Cancelled` result. A failure or signal does not submit a success-looking
result; it returns the original error through the existing root diagnostic path.

### Cancellation, failures, and signals

Only a phase that has actually started is marked cancelled or failed; later
phases remain pending and are not displayed as completed. Ctrl+C, SIGTERM, and
context cancellation preserve the typed signal outcome and `128 + signal`
exit code, stop the active Git process group, and wait for it without leaving
children behind. A real operation error wins over a simultaneous cancellation
request. Renderer, Transcript, cleanup, and stdout errors follow the shared
Experience ordering and joining rules, with idempotent close and at-most-once
result submission.

### Acceptance evidence

Acceptance covers Rich PTY phase visibility and primary-screen restoration,
fast transitions, Huh overwrite defaults and decline/cancel paths, explicit
and implicit refs, default-branch/archive/extraction fallback, clone metadata
removal, each partial-disk state, cancellation and signal process-group
behavior, safe redaction with credential sentinels, Transcript ordering,
exactly-once stdout results, Plain loading/result ordering, Automation's
control-free interactive boundary, and unchanged command help, errors, exit
codes, side effects, and output bytes.
