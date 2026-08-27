# Approve the CLI experience migration and acceptance plan

Type: grilling
Status: resolved
Blocked by: 05, 06, 07, 08, 10

## Question

What implementation order upgrades the existing ycy terminal experience while
keeping each integrated state buildable and compatible? Decide the migration
phases, command grouping, dependency additions, removal strategy for the
handwritten terminal adapters, and acceptance evidence required before each
phase lands.

The plan must include focused Module tests, pseudo-terminal and redirected-I/O
tests, exact Command Result regression coverage, Diagnostic Event/redaction
tests, signal/cancellation tests, cross-platform terminal checks, and a final
standalone-binary smoke matrix. It must name any command-specific exceptions
found by the inventory instead of smoothing them over.

## Answer

Approved on 2026-08-27. The terminal experience migrates in compatible,
vertical command slices. A slice keeps the repository buildable, has a single
command journey or shared terminal concern, and is independently revertible.
This plan does not add terminal-only command grammar, implicit Automation
inputs, hidden mode switches, or alternate command workflows. It implements
only the logging grammar already approved by [Define the command-result and
diagnostic contract](06-define-output-and-diagnostic-contract.md).

### Delivery and dependency rules

- `internal/terminal` is the sole shared terminal Module. `cmd/ycy` owns thin
  semantic Adapters and composition; `internal/commands` imports neither that
  Module nor Charmbracelet.
- Each migrated command replaces and deletes its handwritten
  `terminal...Prompter` and `terminal...Presenter` implementation in the same
  integration change. There is no feature flag, dual renderer, or deferred
  production-code cleanup.
- The foundation adds Huh `v1.0.0` and Lip Gloss `v1.1.0`. Bubble Tea
  `v1.3.10` is added only with the first tracked-operation slice. Bubbles is
  added only if a completed view actually imports it. `go mod tidy` must leave
  the dependency graph clean after each dependency addition.
- Existing exact text remains authoritative for Automation Sessions, Plain
  Interactive Sessions, Raw Child I/O, machine-consumed Command Results, and
  intentional command-specific exit behavior. Rich Interactive rendering has
  its own semantic and PTY screen evidence; it is not a byte-for-byte copy of
  the old handwritten UI.
- The old Bun implementation remains visual and behavioral evidence only. No
  production path, test, or fallback executes it.

### Migration phases

#### Phase 0: characterize the protected behavior

Before changing terminal behavior, add reusable test helpers for injected
terminal facts, recording Experience Runs, redirected streams, and controlled
PTY subprocesses. Capture the existing stdout, stderr, exit, mutation,
cancellation, and redaction contracts for every command journey in the
matrix below. This phase introduces no Rich Interactive behavior.

#### Phase 1: establish the terminal Module

Create `internal/terminal` with the approved Session classifier, visual
roles, Presentation Document renderer, Interaction Request handling,
Experience Run lifecycle, Renderer Lease, and lease-aware diagnostic writer.
Its pure classifier, output routing, width behavior, redaction boundary, and
cleanup logic receive focused Module tests. The Module remains private to the
terminal-facing composition layer and exposes no Cobra or business-command
types.

#### Phase 2: wire the root and common diagnostics

Construct one Experience at `cmd/ycy` from inherited streams and environment.
Route Cobra discovery presentation, plain User-Actionable Errors, and the
already-approved diagnostic configuration through the appropriate terminal
seams. Implement text and NDJSON Diagnostic Records, recursive redaction,
`NO_COLOR`, and lease deferral before command-specific Rich rendering begins.
Help, completion, version, syntax, and exit-code ownership remain in Cobra.

#### Phase 3: migrate static and continuous-output journeys

`tunnel server` is an explicit command migration in this phase, not merely an
example of the logging foundation. It continues to own a continuous stderr
Diagnostic Event stream, with Rich Interactive semantic level and scope
color, Plain/Automation unstyled line records, JSON NDJSON, redaction, and
signal-driven stopping. It never acquires a Renderer Lease or opens Huh or
Bubble Tea.

The same phase migrates static Rich presentation for `config fork list`,
`config cm list`, `config cm use`, `config cm set`, `config cm test`, `git
heat`, and `upgrade`. It makes only safe human-facing layout and color changes
to `diff` and `fs` lifecycle presentation while preserving their stdout
readiness URLs and foreground signal behavior. It also adds the preservation
tests that prove the eventual `run` selector migration cannot wrap its child
process's Raw Child I/O.

#### Phase 4: migrate ordinary forms one command at a time

Use the shared Huh Adapter primitives, but integrate one existing command
journey per change, in this order:

1. `config fork add`
2. `config fork remove`
3. `config cm add`
4. `config cm remove`
5. `export env`
6. `rm`
7. `run` selection before Raw Child I/O
8. `zip`
9. `tunnel connect` selection before its continuous client diagnostics

Every slice preserves the command's current validation, cancellation,
destructive-action, side-effect, and Automation fallback behavior. Secrets
are accepted only through an Interactive Session's secure terminal input;
Automated prompt-dependent paths use [Define automation fallback for prompted
commands](10-choose-automation-inputs-for-prompted-commands.md).

#### Phase 5: migrate tracked Git operations

Add Bubble Tea only here, then integrate `git pulse`, `git fork`, and `git
cm` separately. Each view is inline on stderr, disables Bubble Tea signal
handling, takes a Renderer Lease only around its tracked segment, and leaves
a compact final phase in scrollback before its Command Result renders. No
view uses an alternate screen, captures Raw Child I/O, invents a percentage
or retry, or changes the command-owned cancellation or partial-success
semantics.

#### Phase 6: prove completion and remove residue

Audit that each migrated leaf removed its old terminal Adapter, that no
terminal control bytes reach an Automation Session, and that no
Charmbracelet import entered `internal/commands`. Execute the final matrix,
native terminal checks, and standalone-binary smoke suite. This phase does
not introduce a release workflow or require `make release-candidate`.

### Command acceptance matrix

Every named journey receives Rich Interactive, Plain Interactive, and
Automation evidence where its behavior applies. Rich-only expectations never
become Automation output requirements.

| Journey | Required evidence and exception |
| --- | --- |
| Help, version, completion, parser errors | Rich Help may be static and styled; Automation Help/version/completion remain plain stdout; errors remain one unstyled stderr `error:` line with existing exit behavior. Completion remains raw script bytes. |
| `config fork list`; `config cm list/use/set/test`; `git heat`; `upgrade` | Exact Plain/Automation results, stable stream/exit contracts, Rich semantic presentation, `NO_COLOR`, and redaction. Upgrade retains its deliberate mixed result/error and exit exceptions. |
| `config fork add/remove`; `config cm add/remove` | Form order and validation, masked secrets, successful persistence, decline/cancel behavior, destructive confirmation, and Automation failure before writes. |
| `export env`; `rm`; `zip` | Existing JSON/file/deletion/archive results and side effects, cancellation before mutation, `--env`/`--force` existing paths, and no inferred Automation answer. |
| `run` | Script and package-manager selection in an Interactive Session; Automation prompt fallback; exact child argv, inherited stdin, stdout, stderr, exit, and signal behavior without a renderer. |
| `git pulse`; `git fork`; `git cm` | Approved Huh and tracked-phase sequence, inline cleanup, cancellation, errors, durable report/result, no alternate-screen bytes, and partial facts such as a successful commit followed by failed push. |
| `tunnel server` | Continuous stderr Diagnostic Records for lifecycle, retry, warning, failure, and shutdown; Rich semantic colors, Plain/Automation text, JSON NDJSON, recursive redaction, `NO_COLOR`, redirection, and SIGINT cleanup. |
| `tunnel connect` | Existing resolution precedence and ambiguity errors; selection only before client diagnostics start; identical continuous log stream requirements after connection; Automation ambiguity fails before client launch. |
| `diff`; `fs` | Readiness URLs remain discoverable on stdout while the process is alive; no transient renderer competes with foreground lifecycle; signal shutdown and browser-facing startup contract stay intact. |

### Verification cadence

For every implementation slice, run focused command Module tests, recording
Adapter tests, and the named command's relevant Rich/Plain/Automation
evidence. A slice that changes a prompt or destructive path additionally
proves cancellation and failure before side effect. A slice that changes
diagnostics proves text and JSON serialization, recursive redaction, and
lease ordering.

At the end of each phase, run `make check`, `make build`, `make cross-build`,
and a current-host standalone-binary smoke test for the completed journeys.
The ticket's matrix is the terminal-specific acceptance record; the existing
Go Migration Acceptance Ledger remains the source for build and artifact
evidence rather than gaining a parallel ledger.

The final gate additionally requires:

- pure classifier, formatter, redaction, and width tests;
- PTY tests for Rich Interactive and Plain Interactive behavior, including
  `NO_COLOR`, resize/narrow layout, Ctrl-C/Esc cleanup, and no alternate
  screen;
- redirected-stream and `CI` Automation tests for every affected prompt path,
  with no controlling-terminal access, ANSI, cursor control, partial Command
  Result, or side effect;
- host standalone smoke coverage for static reports, Huh forms, all three
  Bubble Tea journeys, service diagnostics, browser URL lifecycle, Raw Child
  I/O, and command-specific exit exceptions;
- six-target `CGO_ENABLED=0` cross-builds at phase boundaries; and
- native Rich terminal verification on macOS, Linux, and Windows after the
  first Bubble Tea integration and again before accepting the final matrix.

Native terminal execution is required evidence because cross-compilation does
not prove terminal mode, signal, or console cleanup behavior. It does not
require a new CI workflow or a three-platform manual gate for every static
presentation change.

### Rollback and stop conditions

The rollback unit is the current command slice or shared-foundation change.
Revert it rather than retaining a hidden renderer fallback. Stop the migration
when a protected stream, exit, mutation, redaction, cancellation, or
cross-platform terminal contract cannot be demonstrated; resolve the narrow
compatibility question before integrating a later phase.
