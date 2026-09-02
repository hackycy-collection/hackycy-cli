# Choose the upgrade terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact release-resolution, download, checksum, candidate verification, and
scheduling Work Phases; already-current state; startup transaction result;
failure/rollback presentation; redacted Interaction Transcript; and final
Command Result should `upgrade` use while preserving every existing unusual
exit and self-update contract?

## Answer

`upgrade` is a finite command with no user form. In a capable Rich TTY it uses
one vivid Upgrade Live View in AltScreen, then replays a compact semantic
Interaction Transcript after the primary screen is restored. Plain Interactive
uses the same semantic phases as control-free linear diagnostics. Automation and
redirected execution emit no phases or Transcript. All existing arguments,
release sources, file effects, state-file names, stdout text, exit codes, and
parent/hidden-updater boundaries remain unchanged.

### Work Phase catalog and execution

The command owns this ordered catalog. Every phase actually reached submits
`Active` and one terminal state (`Completed`, `Failed`, or `Cancelled`), even
when it completes too quickly to be perceived. No artificial sleep, file-level
spinner, ETA, or invented percentage is added.

1. `consume-startup-transaction`: inspect and consume a completed prior
   transaction. A pending transaction is reported and blocks the command using
   the existing error. `--version`/`-V` and the hidden updater entry skip this
   consumption exactly as today.
2. `resolve-release`: validate the current semantic version, request and parse
   latest release metadata, validate JSON/tag/version, and compare versions.
3. `resolve-artifact`: select the target platform/architecture artifact and
   obtain its checksum from the release digest or `SHA256SUMS` manifest.
4. `download-candidate`: download the candidate beside the install target,
   checking HTTP status, body completeness, and non-empty content.
5. `verify-candidate`: verify SHA-256, executable permissions/quarantine
   handling, and the candidate's plain `--version` self-check.
6. `stage-updater`: copy and protect the detached updater executable.
7. `publish-pending`: atomically publish the pending Go-owned transaction
   state.
8. `schedule-updater`: start the detached updater and release its process
   handle. The parent does not wait for, capture, or report child output.
9. `complete`: commit the already-current, scheduled, aborted, or ordinary
   failure outcome.

Already-current completes after release resolution (and artifact identity is
known) without downloading, staging, writing state, or spawning a child. A
scheduled result is emitted only after pending state publication and detached
spawn both succeed; it never claims that the new binary is already installed.

### Resolution, candidate, and scheduling failures

Release metadata and artifact resolution remain strict. Invalid current
version, malformed `v` semantic tag, unsupported platform/architecture,
missing or duplicate asset, malformed digest/manifest, and invalid HTTP/JSON
responses fail at their real phase; the command never guesses a missing value.

Download and verification retain their current ordering and cleanup:
Content-Length truncation, empty content, checksum mismatch, permission or
quarantine failure, and candidate version mismatch stop before scheduling and
remove the staged candidate. Full URL, headers, checksum, temporary path, and
self-check output are diagnostics-only and are not copied into the Transcript.

Copy/chmod, pending-state publication, and detached-spawn errors fail at their
corresponding stage. Candidate, updater, and state temporary files are cleaned
with the existing retry policy; no second upgrade attempt is started.

The deliberate exit matrix is preserved. Resolution/download/verification
failures set `Aborted=true`, write one redacted diagnostic, emit the existing
`Update aborted.` stdout result, and preserve the owned `ExitCodeError` (HTTP,
checksum, and empty-file classes remain exit code 0; other classified aborts
remain exit code 1). Stage/publish/schedule failures do not fabricate
`Update aborted.`; root emits the existing diagnostic and exit code 1.
Context cancellation or SIGINT/SIGTERM marks the active phase `Cancelled`,
requests underlying cancellation at most once, drains the operation, and
preserves the existing context/signal error semantics. It is never rewritten as
a user decline and does not introduce automatic retries.

### Hidden updater and startup transaction

The detached updater keeps the existing transaction protocol: it waits for the
parent to exit, replaces the target, re-applies permissions/quarantine rules,
re-hashes and self-checks the installed binary, rolls back on failure, and
persists exactly one of `succeeded`, `succeeded_with_cleanup_warning`, or
`failed`. A cleanup warning is distinct from replacement failure. A rollback
failure retains the original failure plus a bounded `rollback failed` class;
the updater never reports success after an unverified replacement.

On the next ordinary invocation, startup consumes a completed state exactly
once and removes the state/updater copy under the existing rules. The startup
result is independent from the new command result:

- `succeeded` shows that ycy was updated to the target version;
- `succeeded_with_cleanup_warning` shows the update succeeded with old-file
  cleanup warning;
- `failed` shows that the update failed and was rolled back, with rollback
  failure called out when applicable; and
- `pending` shows that an update is still being applied, then returns the
  existing blocking error.

When the defensive leaf path consumes a prior state, its summary is the first
Transcript entry and is not merged into a new scheduled result. Global startup
consumption continues to emit its established startup result before Cobra.

### Rich, Plain, and Automation projection

Rich owns stderr and one Bubble Tea root. The Live View uses the Ops Console
status table, symbol-paired states, and vivid but restrained status colors for active, completed, warning,
cancelled, and failed phases. Stdout remains empty until the one final durable
Command Result. On finish, the runtime freezes the ledger, restores the primary
screen, replays the safe Transcript, flushes deferred diagnostics in order, and
then emits stdout.

Plain Interactive emits control-free phase lines to stderr as they occur and
does not replay them a second time. Automation emits no prompts, phases,
Transcript, color, or terminal controls; it retains only the existing stdout
result and diagnostic/error behavior. Redirected streams follow the same
control-free policy.

### Interaction Transcript

There are no user answers. The Rich Transcript records, in order, only:

- a consumed prior transaction summary when the leaf path performed it;
- current and candidate versions plus safe target platform/architecture;
- artifact basename and checksum source (`release-digest` or `SHA256SUMS`);
- final states for reached resolution, artifact, download, verification,
  staging, state publication, and scheduling phases;
- already-current or scheduled outcome; or the real cancellation/failure phase
  and bounded failure class; and
- the final outcome as the last entry.

It never records keystrokes, animation frames, retries, invalid input,
complete URLs, full SHA-256 values, temporary/transaction paths, HTTP
headers/body, candidate stdout/stderr, hidden updater arguments, raw filesystem
errors, credentials, or large result documents. Values are control-free,
bounded, and independently projected from mutation values. Transcript replay
is exactly once for Rich, absent for Plain and Automation, and never duplicates
the Command Result.

### Command Result and compatibility

The existing stdout contract is immutable and submitted at most once:

- already-current retains the current/latest version text;
- scheduled retains the target version and "will finish after ycy exits" text;
- classified abort retains `Update aborted.` while its redacted diagnostic is
  on stderr; and
- ordinary stage/publish/schedule failures add no stdout result.

Startup transaction output remains a separate startup result. The hidden
updater emits no parent Command Result. No phase, Transcript, URL, checksum,
path, or child log is written to stdout.

### Acceptance evidence

Unit, integration, and PTY evidence must cover strict metadata/tag/version and
artifact selection; digest and manifest checksum paths; HTTP/rate-limit,
malformed JSON, empty/truncated/mismatched downloads; permission/quarantine and
candidate self-check failures; copy/chmod/state/spawn failures and cleanup;
already-current and all exit-code classifications; cancellation at each
cancelable network/download phase; hidden updater wait, replacement, second
verification, rollback, cleanup warning, state persistence, and exactly-once
startup consumption; unchanged help/arguments/state namespace/file modes and
detached process ownership; exact stdout bytes and single submission; Rich
AltScreen enter/exit once, phase visibility, Transcript-before-diagnostics-
before-stdout ordering, and terminal restoration; Plain/Automation/redirection
control-free output; and redaction of URLs, paths, checksums, raw errors,
headers, credentials, and candidate output across supported platforms.
