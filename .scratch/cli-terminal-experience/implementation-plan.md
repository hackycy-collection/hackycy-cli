# CLI Terminal Experience Implementation Plan

## Source Decisions

- .scratch/cli-terminal-experience/map.md: destination, effort boundary, and resolved Wayfinder entry.
- CONTEXT.md: canonical terminal-experience terminology and ownership language.
- .scratch/cli-terminal-experience/issues/01-research-charmbracelet-terminal-runtime.md: permitted dependency graph and application-owned runtime responsibilities.
- .scratch/cli-terminal-experience/research/01-charmbracelet-terminal-runtime.md: version, cleanup, capability, platform, and test evidence behind the runtime decision.
- .scratch/cli-terminal-experience/issues/02-define-terminal-capability-and-automation-contract.md: session classifier and Automation fallback contract.
- .scratch/cli-terminal-experience/issues/03-inventory-command-experience-and-output-contracts.md: compatibility inventory decision.
- .scratch/cli-terminal-experience/inventory/03-command-experience-output-contracts.md: command-by-command stream, cancellation, and regression baseline.
- .scratch/cli-terminal-experience/issues/04-prototype-the-terminal-visual-language.md: approved Rich Interactive visual grammar.
- .scratch/cli-terminal-experience/prototype/bun-baseline/README.md: concrete Bun-baseline parity constraints and identified implementation gaps.
- .scratch/cli-terminal-experience/issues/05-choose-the-terminal-experience-module-seam.md: internal/terminal ownership and semantic adapter seam.
- .scratch/cli-terminal-experience/issues/06-define-output-and-diagnostic-contract.md: result, diagnostic, error, redaction, and logging configuration contract.
- .scratch/cli-terminal-experience/issues/07-choose-the-long-running-operation-model.md: tracked-operation classification, phase, cancellation, and cleanup contract.
- .scratch/cli-terminal-experience/issues/08-define-help-error-and-completion-experience.md: Cobra discovery, recovery, and completion boundary.
- .scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md: approved phase ordering, migration matrix, verification cadence, and stop policy.
- .scratch/cli-terminal-experience/issues/10-choose-automation-inputs-for-prompted-commands.md: existing-input-only Automation matrix.
- .scratch/go-migration/map.md: established current-Go compatibility boundary.
- .scratch/go-migration/acceptance.md: canonical build and artifact evidence for the completed Go migration.

## Outcome

ycy retains its existing Cobra grammar, command semantics, exit behavior, and
machine-consumed streams while every approved terminal journey is rendered
through one terminal-facing Module. A Rich Interactive Session receives the
approved command-local Huh, Lip Gloss, and selected inline Bubble Tea
experience; a Plain Interactive Session remains line-oriented; an Automation
Session remains deterministic, unprompted, and stream-safe.

## Non-Negotiable Rules

- Classify a Session once from all inherited standard streams and CI before any
  Huh or Bubble Tea object is constructed. Unknown capability is Plain
  Interactive; a redirected stream or CI is Automation.
- internal/terminal owns terminal capability, styles, terminal cleanup,
  renderer leases, and Charmbracelet lifecycle. cmd/ycy owns thin semantic
  adapters and composition. internal/commands owns business behavior and
  imports neither internal/terminal nor Charmbracelet.
- Command Results remain on stdout. Diagnostic Events use the lease-aware
  stderr writer. User-Actionable Errors are one plain stderr error line. Raw
  Child I/O remains directly inherited.
- Preserve command names, flags, positional grammar, defaults, exit semantics,
  existing explicit input surfaces, durable result formats, and deliberate
  command-specific output exceptions. No global color, yes, non-interactive,
  or Automation-only input flag is introduced.
- A command migration deletes its handwritten terminal adapter in the same
  integration slice. No feature flag, hidden renderer fallback, or concurrent
  production renderer is retained.
- Rich rendering is command-local and semantic. NO_COLOR removes color only.
  Plain Interactive and Automation rendering never emits ANSI cursor control,
  animation, an alternate screen, a title clear, or an unsolicited prompt.
- cmd/ycy remains the sole OS-signal owner. Terminal cleanup precedes return
  to Cobra, and a renderer never owns retries, domain work, inferred progress,
  or Raw Child I/O.
- The current Go migration Acceptance Ledger remains the source for
  build/artifact evidence. This effort does not create or modify a parallel
  acceptance ledger.

## Gate Overview

| Gate | Name | Unlock condition | Outcome |
| --- | --- | --- | --- |
| G0 | Characterize Protected Behavior | Start | Reusable test seams and explicit stream, mutation, cancellation, redaction, and exit baselines exist before terminal behavior changes. |
| G1 | Establish Terminal Foundation | G0 Exit conditions met | internal/terminal provides the approved semantic terminal boundary and is independently testable. |
| G2 | Wire Root and Diagnostics | G1 Exit conditions met | Composition, Cobra discovery, errors, and diagnostics use the shared terminal boundary without changing command behavior. |
| G3 | Migrate Static and Continuous Journeys | G2 Exit conditions met | Static reports, browser lifecycles, Upgrade, and Tunnel service diagnostics obey the new presentation and stream contracts. |
| G4 | Migrate Ordinary Form Journeys | G3 Exit conditions met | Every approved Huh-form journey uses the shared adapter while preserving its existing Automation fallback and side effects. |
| G5 | Migrate Tracked Git Journeys | G4 Exit conditions met | Git Pulse, Fork, and CM use approved inline tracked phases with native Rich terminal evidence. |
| G6 | Audit Completion and Accept | G5 Exit conditions met | All residue is removed and the complete terminal-specific acceptance matrix is demonstrated. |

## G0: Characterize Protected Behavior

### Purpose

Close the regression risk created by replacing handwritten terminal adapters by
making each protected observable contract executable before production terminal
behavior changes.

### Inputs

- .scratch/cli-terminal-experience/inventory/03-command-experience-output-contracts.md
- .scratch/cli-terminal-experience/issues/02-define-terminal-capability-and-automation-contract.md
- .scratch/cli-terminal-experience/issues/06-define-output-and-diagnostic-contract.md
- .scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md
- cmd/ycy, internal/cliapp, internal/logging, and their existing tests
- .scratch/go-migration/acceptance.md

### Objective

Provide injected terminal-fact, recording Experience Run, redirected-stream,
and controlled PTY test support, then capture the protected output, exit,
mutation, cancellation, and redaction expectations for every journey in the
approved matrix.

### Scope boundary

This Gate may add test-only helpers and focused compatibility tests around the
current command seams. It does not add Rich rendering, alter production
adapters, introduce internal/terminal production behavior, or change public
output.

### Constraints

- Preserve existing exact-output tests until the same observable Automation,
  stream, redaction, exit, and mutation behavior is represented at the new
  terminal seam.
- Cover Help, version, completion, parser errors, configuration, export,
  local, Git, browser, Tunnel, and Upgrade journeys according to the inventory.
- Test destructive and prompt-dependent paths for cancellation or Automation
  failure before side effects.

### Slice policy

One reusable helper or one command journey contract is one slice. A slice may
add only the assertions needed to expose that concern; it does not refactor a
terminal adapter or pre-stage a production renderer.

### Verification

#### Directed

- After each helper or compatibility slice, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./cmd/ycy ./internal/cliapp ./internal/logging

  The slice's focused tests must prove the stated stream, exit, cancellation,
  mutation, or redaction behavior rather than only a combined-output result.
- Exercise injected terminal facts and redirected streams without consulting
  the test process terminal. PTY coverage must distinguish Rich and Plain
  Interactive behavior, including NO_COLOR and cleanup.

#### Repository

1. At Gate completion, run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- 无

### Evidence rule

Focused test output proves each newly introduced helper and every protected
journey contract. The three repository commands prove the phase remains
buildable, embeds its required assets, and produces the existing six
CGO-free build targets.

### Stop conditions

- A protected stream, exit, mutation, cancellation, redaction, or
  compatibility contract cannot be characterized from the inventory and
  existing behavior.
- A required test needs an undocumented external service, destructive target,
  or input policy that cannot be isolated.

### Rollback

Revert the current test-helper or test-contract slice only; leave all
pre-existing production adapters and unrelated test evidence intact.

### Exit conditions

- The test suite has reusable injected-fact, recording-run, redirected-stream,
  and controlled-PTY seams.
- Every journey in the command acceptance matrix has focused protection for
  its applicable stream, result, error, mutation, cancellation, and redaction
  contract.
- Prompt-dependent Automation paths have before-side-effect tests, and
  Automation assertions reject implicit terminal access and control bytes.
- The Directed and Repository evidence is recorded.

## G1: Establish Terminal Foundation

### Purpose

Create the one deep terminal Module that removes duplicated terminal lifecycle
logic while preserving command-owned semantic interfaces.

### Inputs

- .scratch/cli-terminal-experience/issues/01-research-charmbracelet-terminal-runtime.md
- .scratch/cli-terminal-experience/research/01-charmbracelet-terminal-runtime.md
- .scratch/cli-terminal-experience/issues/02-define-terminal-capability-and-automation-contract.md
- .scratch/cli-terminal-experience/issues/04-prototype-the-terminal-visual-language.md
- .scratch/cli-terminal-experience/issues/05-choose-the-terminal-experience-module-seam.md
- G0 compatibility seams and evidence

### Objective

Create internal/terminal with a pure Session classifier, terminal-owned
semantic documents and requests, Experience and Experience Run lifecycle,
presentation rendering, interaction handling, width-aware visual roles,
renderer leasing, and a lease-aware diagnostic writer.

### Scope boundary

This Gate may add internal/terminal, Huh v1.0.0, and Lip Gloss v1.1.0 with
their directly required transitive graph. It may add foundation tests and
adapter fakes. It does not wire Cobra, alter a command journey, add Bubble
Tea or Bubbles, or move business interfaces into the terminal Module.

### Constraints

- Session classification uses all three standard streams, CI, TERM, and
  NO_COLOR exactly as approved; no dependency chooses Automation behavior.
- Experience Run exposes only semantic Ask, Present, and Track operations.
  It restores controlled terminal state on every return path.
- Present produces a Command Result on stdout. Ask and Track coordinate the
  diagnostic stream through a Renderer Lease.
- The Module does not import Cobra or command business types. Command Modules
  do not import the Module or Charmbracelet.
- Do not add Bubble Tea before G5. Do not add Bubbles unless a completed
  tracked view imports one.

### Slice policy

Implement and verify one foundation concern at a time: classifier, semantic
types, plain rendering, Rich visual roles and width behavior, interaction
wrapper, then lifecycle and lease coordination. Each concern remains
independently reversible.

### Verification

#### Directed

- After the first internal/terminal slice and after every subsequent
  foundation slice, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/terminal ./cmd/ycy

- When adding a direct dependency, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off go mod tidy
    GOTOOLCHAIN=go1.26.7 GOWORK=off go mod verify

  Record the pinned graph and prove the classifier, plain stream routing,
  NO_COLOR behavior, width behavior, and controlled cleanup through focused
  tests.

#### Repository

1. At Gate completion, run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- 无

### Evidence rule

The focused internal/terminal suite proves each foundation responsibility
without a process-terminal dependency. Module and dependency verification,
followed by the repository commands, proves the selected CGO-free foundation
is available to the application without changing public command behavior.

### Stop conditions

- The selected dependency graph cannot remain CGO-free for a required target
  or needs a package outside the approved direct dependency boundary.
- A classifier, cleanup, or stream-routing behavior cannot be proved without
  unsafe controlling-terminal access or a command-specific policy decision.

### Rollback

Revert the current internal/terminal foundation slice and its dependency
change as one unit; do not leave a partial shared abstraction or a duplicate
adapter path.

### Exit conditions

- internal/terminal owns the approved classifier, semantic operations, visual
  roles, width handling, controlled cleanup, renderer lease, and
  lease-aware diagnostic writer.
- Focused tests prove Rich, Plain Interactive, and Automation classification,
  NO_COLOR semantics, stdout/stderr routing, and cleanup boundaries.
- Huh and Lip Gloss are pinned as approved, the module graph is tidy and
  verified, and no Bubble Tea or Bubbles dependency has been added.
- The Directed and Repository evidence is recorded.

## G2: Wire Root and Diagnostics

### Purpose

Connect the shared terminal boundary at the composition root while making
diagnostics, discovery, errors, and logging configuration explicit without
moving command behavior into the terminal Module.

### Inputs

- .scratch/cli-terminal-experience/issues/05-choose-the-terminal-experience-module-seam.md
- .scratch/cli-terminal-experience/issues/06-define-output-and-diagnostic-contract.md
- .scratch/cli-terminal-experience/issues/08-define-help-error-and-completion-experience.md
- cmd/ycy/main.go, internal/cliapp, internal/logging, and G1

### Objective

Construct one Experience from inherited streams and environment lookup,
connect the lease-aware diagnostic writer, and implement the approved text and
NDJSON diagnostic configuration, recursive redaction, Cobra discovery
document path, and one-line error behavior.

### Scope boundary

This Gate may change cmd/ycy, internal/cliapp, internal/logging, and
internal/terminal integration points. It does not migrate a command-specific
form, static report, tracked operation, browser lifecycle, or Raw Child I/O
path; those remain in later Gates.

### Constraints

- Cobra remains the grammar, Help, completion, direct-descendant, and parser
  recovery owner. internal/cliapp does not import Charmbracelet.
- Help, version, and completion remain stdout Command Results. Completion
  remains raw generated script bytes.
- An executing command validates the approved log-level, quiet, verbose, and
  log-format grammar before effects. Discovery paths retain their established
  behavior even when diagnostic controls are invalid.
- Diagnostic Records obey the approved text and NDJSON schemas, recursive
  redaction, threshold precedence, style policy, and lease ordering.
- A User-Actionable Error remains exactly one unstyled stderr error line; the
  separate DEBUG panic-stack behavior is unchanged.

### Slice policy

Use one root concern per slice: Experience construction, discovery document
translation, error normalization, diagnostic configuration parsing, formatter,
redaction, then lease-aware writer integration. No slice includes a
command-specific Rich renderer.

### Verification

#### Directed

- After each integration slice, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/terminal ./internal/cliapp ./internal/logging ./cmd/ycy

- Add focused cases for Help, version, raw completion, parser recovery, each
  diagnostic configuration precedence/conflict, text and NDJSON records,
  recursive redaction, NO_COLOR, and lease deferral.

#### Repository

1. At Gate completion, run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- 无

### Evidence rule

Focused package tests prove every discovery, error, configuration, formatter,
redaction, and lease contract. Existing Automation output and parser
regressions prove the root integration has not altered command-facing
semantics. Repository commands prove the root remains buildable across the
approved target set.

### Stop conditions

- A diagnostic change would alter a machine-consumed Command Result, raw
  completion bytes, User-Actionable Error shape, or an established exit code.
- Correct configuration timing for the special run parser cannot be shown by
  tests without changing its frozen grammar.

### Rollback

Revert the current root or diagnostic integration slice together with its
matching tests; retain the completed terminal foundation and all unrelated
command adapters.

### Exit conditions

- cmd/ycy constructs one Experience and supplies the lease-aware diagnostic
  writer to normal diagnostic paths.
- Cobra discovery, Help, completion, version, errors, and parser recovery
  retain their approved ownership and stream behavior.
- Text and NDJSON Diagnostic Records, recursive redaction, approved
  configuration precedence, NO_COLOR behavior, and lease deferral have
  focused evidence.
- The Directed and Repository evidence is recorded.

## G3: Migrate Static and Continuous Journeys

### Purpose

Move non-form human-facing journeys to the shared presentation boundary while
protecting service diagnostics, browser readiness streams, Upgrade exceptions,
and later Raw Child I/O migration.

### Inputs

- .scratch/cli-terminal-experience/issues/03-inventory-command-experience-and-output-contracts.md
- .scratch/cli-terminal-experience/inventory/03-command-experience-output-contracts.md
- .scratch/cli-terminal-experience/issues/04-prototype-the-terminal-visual-language.md
- .scratch/cli-terminal-experience/issues/06-define-output-and-diagnostic-contract.md
- .scratch/cli-terminal-experience/issues/07-choose-the-long-running-operation-model.md
- .scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md

### Objective

Migrate Tunnel server, static configuration and Git Heat reports, Upgrade,
and the safe human-facing presentation portions of Diff and FS. Add the
preservation evidence needed before the later Run selector migration.

### Scope boundary

The Gate includes Tunnel server; config fork list; config cm list, use, set,
and test; Git Heat; Upgrade; Diff; FS; and Run Raw Child I/O preservation
tests. It excludes all prompted forms, Tunnel connect selection, Raw Child
I/O wrapping, and Bubble Tea tracked work.

### Constraints

- Tunnel server remains a continuous stderr Diagnostic Event stream and never
  acquires a Renderer Lease, Huh form, or Bubble Tea view.
- Diff and FS keep their readiness URLs discoverable on stdout while their
  foreground lifecycle runs. They do not gain a competing transient renderer.
- Upgrade retains its deliberate mixed result/error and exit exceptions.
- Static Rich presentation may change only safe human-facing layout and
  semantic color. Plain Interactive and Automation text, stream, and exit
  behavior remain protected.
- Each migrated journey deletes its matching handwritten terminal adapter in
  the same slice. Run receives preservation tests only in this Gate.

### Slice policy

One named journey is one slice. Tunnel server is a standalone continuous-log
slice; Diff, FS, and Run preservation are separate slices. Do not combine
unrelated reports merely because they share a renderer.

### Verification

#### Directed

- After each journey slice, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/terminal ./internal/logging ./internal/cliapp ./cmd/ycy ./internal/commands/config/fork ./internal/commands/config/cm ./internal/commands/git/heat ./internal/commands/diff ./internal/commands/fs ./internal/commands/tunnel ./internal/commands/upgrade ./internal/commands/run

- Tunnel tests must cover text and NDJSON records, redaction, NO_COLOR,
  redirection, shutdown, and lease ordering. Diff, FS, and Run tests must
  retain independent stdout/stderr, readiness, child-process, and signal
  evidence.

#### Repository

1. At Gate completion, run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- 无

### Evidence rule

The named journey tests prove the approved static presentation or continuous
diagnostic behavior and all protected exceptions. The build commands prove
the integrated command set remains available in the existing standalone
artifact shape.

### Stop conditions

- A static presentation would hide a browser readiness URL, capture a child
  stream, alter Upgrade's intentional result/exit behavior, or corrupt a
  Tunnel service log.
- A journey cannot delete its legacy adapter without proving equivalent
  Automation and Plain Interactive output.

### Rollback

Revert the current named journey slice, including its adapter deletion and
tests, without reintroducing a feature-flagged old renderer.

### Exit conditions

- Every listed static or continuous journey uses the shared terminal boundary
  where approved, and its handwritten adapter is removed.
- Tunnel server retains continuous stderr diagnostics with text, JSON,
  redaction, color, and shutdown evidence.
- Diff and FS preserve stdout readiness and lifecycle behavior; Run's Raw
  Child I/O preservation proof is complete.
- The Directed and Repository evidence is recorded.

## G4: Migrate Ordinary Form Journeys

### Purpose

Replace every approved ordinary prompt journey with shared Huh adapter
primitives while preserving existing validation, cancellation, side effects,
and Automation fallback behavior.

### Inputs

- .scratch/cli-terminal-experience/issues/02-define-terminal-capability-and-automation-contract.md
- .scratch/cli-terminal-experience/issues/05-choose-the-terminal-experience-module-seam.md
- .scratch/cli-terminal-experience/issues/07-choose-the-long-running-operation-model.md
- .scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md
- .scratch/cli-terminal-experience/issues/10-choose-automation-inputs-for-prompted-commands.md
- G3 Run preservation evidence

### Objective

Migrate the approved form journeys in order: config fork add, config fork
remove, config cm add, config cm remove, export env, rm, Run selection, zip,
then Tunnel connect selection.

### Scope boundary

The Gate includes only those nine command journeys and shared Huh adapter
primitives needed by them. It excludes static journeys, tracked Git work,
Tunnel service rendering after selection, new input grammar, and business
workflow changes.

### Constraints

- Secrets are entered only through terminal-only echo-disabled input in an
  Interactive Session. A pipe is never a secret channel.
- Automation uses only already-resolved command input, configuration, and
  defaults. Every Prompt-Dependent Path follows its approved failure matrix
  before a side effect and emits no partial Command Result.
- Existing destructive confirmation, cancellation, validation, and mutation
  semantics remain command-owned.
- Run releases the form before executing the child and preserves inherited
  stdin, stdout, stderr, exit, and signal behavior without a renderer.
- Tunnel connect finishes selection before continuous client diagnostics start;
  ambiguous remembered connections retain their established specific error.
- Each slice removes the matching handwritten terminal adapter immediately.

### Slice policy

One ordered command journey is one independently reversible slice. Shared Huh
primitive work belongs only to the first slice that needs it. Do not combine a
read, write, confirmation, selection, or presentation responsibility from
different journeys.

### Verification

#### Directed

- After each form journey slice, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/terminal ./internal/cliapp ./cmd/ycy ./internal/commands/config/fork ./internal/commands/config/cm ./internal/commands/exportenv ./internal/commands/rm ./internal/commands/run ./internal/commands/zip ./internal/commands/tunnel

- For the current journey, add or retain focused Rich, Plain Interactive, and
  Automation cases for prompt order, cancellation, validation, redaction,
  stream separation, existing resolved paths, and before-side-effect failure.
- For every destructive or writing journey, prove decline, cancel, validation
  failure, and Automation failure leave the protected state unchanged.

#### Repository

1. At Gate completion, run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- 无

### Evidence rule

Focused form and command tests prove each migration retains its semantic
request, result, side-effect, cancellation, and fallback contract. The
repository commands prove the complete form set stays buildable and
cross-compilable.

### Stop conditions

- A journey requires an Automation-only flag, implicit value, controlling
  terminal, visible secret input, or changed confirmation semantics to
  complete.
- A migration cannot preserve child-process stream ownership, Tunnel
  resolution precedence, or mutation-before-failure guarantees.

### Rollback

Revert the current command journey slice, including its Huh adapter and old
adapter deletion, while preserving completed earlier journey slices.

### Exit conditions

- The nine journeys are migrated in the approved order and each has removed
  its handwritten terminal adapter.
- Every migrated journey has Rich, Plain Interactive, and applicable
  Automation evidence for its approved behavior.
- Secret, destructive, child-process, and Tunnel-specific boundaries are
  demonstrated without new grammar or inferred input.
- The Directed and Repository evidence is recorded.

## G5: Migrate Tracked Git Journeys

### Purpose

Add the narrow Bubble Tea surface only where command-owned live state benefits
from it, then prove terminal cleanup and native Rich behavior.

### Inputs

- .scratch/cli-terminal-experience/issues/01-research-charmbracelet-terminal-runtime.md
- .scratch/cli-terminal-experience/issues/04-prototype-the-terminal-visual-language.md
- .scratch/cli-terminal-experience/issues/05-choose-the-terminal-experience-module-seam.md
- .scratch/cli-terminal-experience/issues/07-choose-the-long-running-operation-model.md
- .scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md
- G4 form and cancellation evidence

### Objective

Add Bubble Tea v1.3.10 only with the first tracked slice, then separately
migrate Git Pulse, Git Fork, and Git CM to the approved Huh and inline tracked
phase sequence.

### Scope boundary

This Gate includes the terminal tracked-operation renderer and the three Git
journeys. It excludes a general TUI runtime, Bubbles unless a completed view
uses it, alternate-screen rendering, Raw Child I/O, generic retry controls,
and unrelated Git behavior changes.

### Constraints

- Every tracked view is inline on stderr under a Renderer Lease and leaves a
  compact final phase in scrollback before its durable stdout result.
- Bubble Tea signal handling is disabled. cmd/ycy retains signal ownership;
  Ctrl-C and Esc use the command-owned cooperative cancellation path.
- Pulse, Fork, and CM expose typed Operation Phases. The terminal does not
  infer phases, percentages, ETAs, retries, or progress from logs or
  goroutines.
- A lease is released before every Huh form and before durable results or
  deferred diagnostics render.
- No tracked view uses the alternate screen or changes partial-success,
  failure, cancellation, external-Git, or redaction semantics.

### Slice policy

Add the dependency and renderer capability with the first Git Pulse slice,
then migrate Git Fork and Git CM as separate slices. Within a journey, one
contiguous command-owned phase segment is one slice; forms, phases, and
durable results remain separate responsibilities.

### Verification

#### Directed

- After each tracked journey slice, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/terminal ./internal/cliapp ./cmd/ycy ./internal/commands/git/pulse ./internal/commands/git/fork ./internal/commands/git/cm ./internal/commands/git/heat

- When Bubble Tea is first added, run:

    GOTOOLCHAIN=go1.26.7 GOWORK=off go mod tidy
    GOTOOLCHAIN=go1.26.7 GOWORK=off go mod verify

- PTY tests must prove inline rendering, NO_COLOR, narrow/resize behavior,
  Ctrl-C and Esc cleanup, no alternate-screen bytes, deferred diagnostics,
  durable stdout results, and Git CM partial facts after a failed push.

#### Repository

1. At Gate completion, run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- Entry: build the current binary with make build, then run the first
  completed tracked Git journey in a real Rich Interactive terminal against
  its disposable local repository fixture on macOS, Linux, and Windows.
- Scenario: exercise the Huh selection, live tracked phase, a narrow or
  resized terminal, NO_COLOR, Ctrl-C, and Esc; confirm that the cursor/input
  state returns and the final durable result follows the inline view.
- Checklist: no alternate screen, no corrupted diagnostic record, no
  unmasked credential, no lost stdout result, and command-owned cancellation
  or partial-success semantics.
- Confirmation format: Rich terminal verification: platform=<macOS|Linux|Windows>; terminal=<name>; journey=<name>; result=<pass|fail>; evidence=<command or captured test reference>; notes=<none or details>.

### Evidence rule

Focused model, adapter, command, PTY, and dependency tests prove the
automatable phase and cleanup conditions. One explicit native Rich terminal
confirmation for each required platform proves the behavior that
cross-compilation cannot establish. Repository commands prove the integrated
Git terminal surface remains buildable for all target artifacts.

### Stop conditions

- A tracked view needs an undocumented command phase, changes a cancellation
  or partial-success contract, emits an alternate screen, or cannot release
  the diagnostic stream safely.
- Native Rich terminal evidence is unavailable for macOS, Linux, or Windows,
  or it reveals cursor, input, resize, stream, redaction, or cleanup failure.

### Rollback

Revert the current tracked journey slice, including its Bubble Tea capability
or adapter deletion when coupled to that slice; do not retain a second
renderer or an alternate-screen fallback.

### Exit conditions

- Bubble Tea is pinned only where needed, and Bubbles is absent unless a
  completed view directly imports it.
- Git Pulse, Git Fork, and Git CM each use their approved form, tracked-phase,
  cancellation, and durable-result sequence with old adapters removed.
- Focused PTY evidence proves lease, terminal restoration, stream, resize,
  NO_COLOR, cancellation, and partial-success behavior.
- Native Rich terminal confirmations meet the stated checklist on macOS,
  Linux, and Windows.
- The Directed and Repository evidence is recorded.

## G6: Audit Completion and Accept

### Purpose

Close the migration only after all terminal-specific residue, compatibility
tests, cross-builds, standalone smoke evidence, and native Rich verification
are reconciled.

### Inputs

- .scratch/cli-terminal-experience/map.md
- .scratch/cli-terminal-experience/issues/02-define-terminal-capability-and-automation-contract.md
- .scratch/cli-terminal-experience/issues/06-define-output-and-diagnostic-contract.md
- .scratch/cli-terminal-experience/issues/08-define-help-error-and-completion-experience.md
- .scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md
- .scratch/go-migration/acceptance.md
- All G0 through G5 evidence

### Objective

Audit the completed terminal surface against the approved acceptance matrix,
remove terminal-adapter residue, and obtain the final native Rich terminal
evidence without adding release automation or a new product capability.

### Scope boundary

This Gate may remove verified obsolete terminal adapter residue and add missing
acceptance tests or evidence. It does not alter command grammar, business
behavior, release workflow, legacy Bun execution, web application design, or
the Go migration Acceptance Ledger.

### Constraints

- Every public journey receives applicable Rich, Plain Interactive, and
  Automation evidence. Rich-only rendering never becomes an Automation
  expectation.
- No terminal control byte reaches an Automation Session, no command Module
  imports Charmbracelet, and no migrated leaf retains a handwritten terminal
  adapter.
- Standalone binary smoke evidence covers static reports, Huh forms, all
  three tracked Git journeys, service diagnostics, browser URL lifecycle, Raw
  Child I/O, and command-specific exit exceptions.
- Use existing make targets only. Do not introduce make release-candidate,
  publishing, tagging, CI, or deployment work.

### Slice policy

One audit concern is one slice: adapter-residue scan, import-boundary scan,
Automation matrix, standalone smoke group, cross-build evidence, then final
native terminal matrix. Do not mix remediation for independent journeys.

### Verification

#### Directed

- Run the complete terminal surface tests:

    GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/terminal ./internal/cliapp ./internal/logging ./cmd/ycy

- Confirm the command-layer import boundary:

    ! rg -n 'github.com/charmbracelet' internal/commands --glob '*.go'

- Confirm handwritten terminal adapter residue is absent:

    ! rg -n 'type terminal.*(Prompter|Presenter|Selector)' cmd/ycy --glob '*.go'

- Execute the final directed PTY, redirected-stream, CI, secret-redaction,
  error, completion, Raw Child I/O, browser lifecycle, and standalone smoke
  suites added by the prior Gates.

#### Repository

1. Run make check.
2. After make check succeeds, run make build.
3. After make build succeeds, run make cross-build.

#### Manual acceptance

- Entry: use the binary built by make build in a native Rich Interactive
  terminal on macOS, Linux, and Windows after the final test matrix is ready.
- Scenario: repeat the completed tracked Git journey checks and run the
  applicable static, form, Tunnel diagnostic, browser lifecycle, and Raw
  Child I/O smoke scenarios.
- Checklist: all controls clean up, NO_COLOR removes only color, Plain
  Interactive uses no rich controls, Automation has no prompt/control byte or
  side effect on Prompt-Dependent Paths, and durable stdout/stderr contracts
  remain observable.
- Confirmation format: Final terminal acceptance: platform=<macOS|Linux|Windows>; terminal=<name>; commands=<comma-separated>; result=<pass|fail>; evidence=<captured output or test reference>; exceptions=<none or details>.

### Evidence rule

The Directed checks prove the source-level ownership boundaries and the
automated acceptance matrix. The repository commands prove complete
buildability and six-target CGO-free output. The three explicit native
confirmations prove terminal behavior that compilation and PTY emulation do
not establish.

### Stop conditions

- Any terminal control byte, prompt, side effect, redaction failure, stream
  regression, exit regression, stale adapter, or forbidden import is found.
- The final native Rich terminal confirmation is unavailable or reports a
  failure on macOS, Linux, or Windows.

### Rollback

Revert only the smallest completed migration slice responsible for the failed
contract. Do not suppress the failing test, weaken the acceptance matrix, or
restore a hidden legacy renderer.

### Exit conditions

- The approved command acceptance matrix is complete with focused, PTY,
  redirected-stream, CI, and standalone evidence.
- No terminal control bytes reach Automation Sessions; no command Module
  imports Charmbracelet; and no migrated handwritten terminal adapter remains.
- make check, make build, and make cross-build all succeed.
- Final native Rich terminal confirmations meet the stated checklist on
  macOS, Linux, and Windows.
- The external Go migration Acceptance Ledger remains the canonical
  build/artifact reference, with no parallel acceptance ledger added.

## Definition Of Done

- G0 through G6 Exit conditions have evidence in the Goal Runbook.
- The terminal Module, root composition, diagnostics, static journeys, form
  journeys, and tracked Git journeys meet their approved ownership and
  compatibility boundaries.
- The complete terminal-specific acceptance matrix, repository build checks,
  cross-builds, standalone smoke evidence, and twice-required native Rich
  terminal verification are complete.
- The current Go CLI remains CGO-free and preserves its public grammar,
  machine-consumed output, command semantics, and established exceptions.

## Explicitly Out Of Scope

- A full-screen command center, dashboard, persistent shell, or replacement
  top-level command.
- Changes to command names, flags, positional grammar, business behavior,
  destructive-action semantics, or exit codes for terminal visual polish.
- Automation-only flags, positional inputs, secret sources, implicit
  selections, or workflow shortcuts for prompt-dependent paths.
- CLI localization or a language selector.
- Redesign of the Diff, FS, or Tunnel web applications.
- Telemetry, remote log shipping, or a general observability platform.
- Release automation, publishing, tagging, CI, deployment, or execution of
  the legacy Bun runtime.
