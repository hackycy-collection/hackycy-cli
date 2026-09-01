# Choose the git cm terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact file-selection and commit-confirmation forms; stage, collect,
generate, commit, and push Work Phases; evidence/usage context; redacted
Interaction Transcript; provider/Git failure treatment; and generated-message
Command Result should `git cm` use without changing mutation or exit semantics?

## Answer

`git cm` is a scope-bound Git CM Workflow that may select and stage changes,
generate one validated commit message, create its commit, and optionally push
it. The redesign preserves the existing flag matrix, Git/provider requests,
mutation order, stdout documents, errors, exit codes, and partial-success
semantics while adding truthful Rich phases, modern Huh forms, semantic
Transcript replay, and exactly-once results.

### Mode matrix and phase catalog

The current modes remain exact:

- No mutation flag or `--dry-run` reads all uncommitted changes, generates a
  message, and emits it without interaction or mutation.
- `--staged` reads only staged changes, then asks whether to commit.
- `--stage` selects files and rewrites the index before generation and commit
  confirmation.
- `--stage-all` runs the existing `git add -A` before generation and commit
  confirmation.
- `--push` is valid only with `--stage`, `--staged`, or `--stage-all`.
- `--stage-push` selects, stages, generates, confirms, commits, and pushes.

All flag conflicts, optional remote parsing, default `origin`, scope rules,
and Automation boundaries remain unchanged. The command uses these conditional
real Work Phases in order:

1. `inspect-changes` / `Inspect changes` reads repository status and the
   selected scope.
2. `stage-selected-files` / `Stage selected files` runs only for `--stage` or
   `--stage-push` after the file form succeeds.
3. `stage-all-changes` / `Stage all changes` runs only for `--stage-all`.
4. `capture-commit-evidence` / `Capture commit evidence` captures the immutable
   snapshot and evidence used for generation.
5. `resolve-provider-profile` / `Resolve provider profile` resolves the CM
   profile and its safe diagnostic projection.
6. `generate-commit-message` / `Generate commit message` performs the single
   provider request and validates the Angular-style result.
7. `verify-unchanged-scope` / `Verify unchanged scope` runs immediately before
   commit and rejects a changed snapshot.
8. `create-commit` / `Create commit` runs only after confirmation.
9. `push-commit` / `Push commit` runs only when push was requested and commit
   creation succeeded.

File selection and commit confirmation are interactions, not Work Phases.
Git subprocesses, untracked-file reads, evidence clustering, HTTP requests,
and branch lookup remain implementation details inside their owning phase.
Every reached phase shows active and final state, even when fast; no artificial
sleep or unmeasurable percentage is introduced. Phase details are limited to
scope, file counts, safe profile/model, language/body mode, and safe remote
name. They never show diffs, hunks, evidence payloads, API keys, Authorization
headers, Git argv, raw provider response bodies, or raw stderr.

### File selection and commit confirmation

`--stage` and `--stage-push` use a searchable Huh v2 MultiSelect. It preserves
the existing Git status order, complete option set, and all-selected default.
Each option gets a vivid change-kind symbol and status label plus a safe path;
colors never carry the only meaning. Narrow terminals may wrap paths, ordinary
Unicode remains intact, and control characters become visible escapes.

Plain Interactive retains numbered selection, `all`, `none`, comma/space
indices, invalid-input retries, and the existing prompt wording. Empty
selection returns `Nothing selected.\n`, exits 0, and leaves the index
unchanged. Esc/Ctrl+C/`q` returns `Cancelled\n`, exits 0, and leaves the index
unchanged.

After generation, Rich keeps the complete subject and optional body visible in
the Live View with safe profile/model, provider token usage, local evidence
estimate, cluster/fact coverage, and compaction information. Huh v2 confirms
with default Yes using `Create this commit? [Y/n]`. A No records `Commit
creation declined`; Esc/Ctrl+C/`q` records `Commit creation cancelled`. Both
return the existing `Cancelled\n` result and do not create a commit. Any
previously successful staging is retained; the command never auto-unstages.

### Command Result and stream ownership

Rich uses one Bubble Tea root in AltScreen with the Signal Rail presentation,
then restores the primary screen and replays a bounded semantic Transcript to
stderr. The complete Command Result is submitted to stdout exactly once:

- generation-only: the full generated message and existing profile, usage, and
  evidence coverage document;
- no all-uncommitted changes: `No uncommitted changes.\n`;
- no staged changes: `No staged changes.\n`;
- empty file selection: `Nothing selected.\n`;
- interaction decline/cancellation: `Cancelled\n`;
- successful commit: `Commit created\n`;
- successful commit and push: `Commit created and pushed\n`.

On a push failure after commit creation, stdout still contains only
`Commit created\n` while the original push error controls the failed exit. Rich
may use a responsive result layout, but these established durable documents
and their ordering do not change. Plain writes generation previews and phase
lines to stderr and keeps the existing final stdout result. Automation emits no
forms, phases, Transcript, styling, or control sequences.

### Transcript and failure facts

After AltScreen exits, Rich replays phase final states, scope and file counts,
safe selected-path summaries, language/body mode, safe profile/model and usage,
the generation summary or complete generated message when a commit decision
was pending, the confirmation answer, mutation facts, and the final outcome.
Generation-only does not duplicate its complete message in the Transcript
because stdout already contains it. Large evidence/diff content, keystrokes,
search terms, invalid retries, unchecked options, credentials, URLs with
userinfo/query secrets, payloads, raw response bodies, Git argv, and raw
errors are excluded.

Failures use stable safe categories and preserve the original diagnostic:

- inspection/evidence: `Git capture`, `filesystem`, or `evidence`;
- profile: `store`, `selection`, `decrypt`, or `configuration`;
- generation: `timeout`, `HTTP status`, `response read/decode`,
  `empty response`, or `invalid model output`;
- recheck: `Git scope changed; commit not created`;
- commit: Git or hook failure;
- push: branch/detached-HEAD or remote rejection/transport failure.

The corresponding disk and Git facts are recorded without guessing: staging
failure may leave `Index may be partially updated`; later failure or
cancel after successful staging records `Staged changes retained`; provider,
recheck, or commit failure records `Commit not created`; commit success with
push failure or cancellation records `Commit created locally; push not
completed`. No unstage, reset, amend, deletion, retry, or rollback is added.

### Automation, cancellation, and signals

Automation preserves the current lazy-config order: the config store is
created/read first, then `RequiresInteraction` checks whether the mode would
reach file selection or commit confirmation. A store error wins unchanged.
When interaction would be required, the command returns
`git cm requires an interactive terminal` without reading stdin, invoking Git,
calling the provider, mutating the index/commit/push, or emitting a result.
Generation-only Automation still performs its normal Git/config/provider work
and emits the complete stdout document.

Ctrl+C, SIGTERM, and context cancellation mark only the active phase cancelled,
leave later phases pending, stop and wait for active Git process groups, and
preserve typed signal outcomes with `128 + signal`. Provider requests use the
existing timeout/cancellation behavior and are not retried. If commit already
succeeded and push is cancelled, the committed partial result remains. A real
operation failure takes precedence over a simultaneous cancellation.

### Acceptance evidence

Acceptance covers every flag mode and conflict, optional remote normalization,
Huh default/search/multiselect and Plain selection syntax, empty/declined/
cancelled interactions, index and partial-stage facts, provider precedence,
timeout/HTTP/read/decode/empty/invalid-output categories and API-key redaction,
generation-only and commit-path stream ordering, stale scope, hooks, push
rejection and detached HEAD, partial committed results, PTY AltScreen and
Transcript ordering, fast phases, renderer recovery, signal/process-group
cleanup, lazy-config Automation boundaries, no-control output, and unchanged
help, output bytes, errors, exit codes, and side effects.
