# Terminal Experience Modernization Implementation Plan

## Source Decisions

- [Map](map.md): defines the destination, non-negotiable terminal boundaries,
  command inventory, and the resolved route through all command experiences.
- [Latest compatible Charm v2 stack](research/01-latest-charm-v2-stack.md) and
  [its research decision](issues/01-research-latest-charm-v2-stack.md): pin the
  mutually compatible Bubble Tea, Bubbles, Huh, Lip Gloss, and Log versions.
- [Exemplary CLI terminal experiences](research/02-exemplary-cli-terminal-experiences.md)
  and [its research decision](issues/02-research-exemplary-cli-terminal-experiences.md):
  establish the live/final state, form, progress, and lifecycle-log references.
- [Current terminal surface inventory](inventories/03-current-command-terminal-surfaces.md)
  and [its decision](issues/03-inventory-every-command-terminal-surface.md):
  establish current streams, modes, losses, and evidence gaps.
- [Charm v2 and logging boundaries](issues/04-choose-charm-v2-and-logging-boundaries.md),
  [vivid visual system](issues/05-prototype-vivid-terminal-visual-system.md),
  and [shared Experience contract](issues/06-choose-shared-terminal-experience-contract.md):
  define the one renderer stack, semantic phases, Transcript, leases, and stream
  ordering.
- [Root/help/error presentation](issues/07-choose-root-help-and-error-presentation.md):
  preserves discovery, help, version, completion, diagnostics, and root exits.
- [Export env](issues/08-choose-export-env-presentation.md), [fork list](issues/09-choose-config-fork-list-presentation.md),
  [fork add](issues/10-choose-config-fork-add-presentation.md), and [fork remove](issues/11-choose-config-fork-remove-presentation.md):
  define the Fork configuration list, setup, and removal slices.
- [CM list](issues/12-choose-config-cm-list-presentation.md), [CM add](issues/13-choose-config-cm-add-presentation.md),
  [CM use](issues/14-choose-config-cm-use-presentation.md), [CM set](issues/15-choose-config-cm-set-presentation.md),
  [CM remove](issues/16-choose-config-cm-remove-presentation.md), and [CM test](issues/17-choose-config-cm-test-presentation.md):
  define the provider-profile read, setup, mutation, selection, and test slices.
- [Diff Lifecycle Log](issues/18-choose-diff-lifecycle-log-presentation.md) and
  [FS Lifecycle Log](issues/19-choose-fs-lifecycle-log-presentation.md): define
  service startup, refresh/task, failure, shutdown, and NDJSON boundaries.
- [Git Heat](issues/20-choose-git-heat-presentation.md), [Git Pulse](issues/21-choose-git-pulse-presentation.md),
  [Git Fork](issues/22-choose-git-fork-presentation.md), and [Git CM](issues/23-choose-git-cm-presentation.md):
  define read-only reports, fetch omissions, archive acquisition, and commit
  workflows.
- [RM](issues/24-choose-rm-presentation.md), [Run handoff](issues/25-choose-run-handoff-presentation.md),
  [Tunnel connect](issues/26-choose-tunnel-connect-lifecycle-log-presentation.md),
  [Tunnel server](issues/27-choose-tunnel-server-lifecycle-log-presentation.md),
  [ZIP](issues/28-choose-zip-presentation.md), and [Upgrade](issues/29-choose-upgrade-presentation.md):
  define destructive, child-process, service, archive, and detached-update
  behavior.
- [Rollout and acceptance decision](issues/30-approve-rollout-and-acceptance-plan.md):
  fixes the risk-ordered batches, no-feature-flag policy, evidence gates,
  behavior baselines, visual review, and release stop conditions.

## Outcome

Every root surface and command leaf has an independently owned terminal
presentation built on one Charm v2 semantic Experience. Finite work shows
truthful phases in Rich, replays a bounded redacted Transcript after AltScreen,
and emits its unchanged Command Result once. Service Commands emit bounded
Lifecycle Logs, while Automation and redirected streams remain control-free and
machine-compatible. The rollout is individually verifiable and reversible at
command-slice boundaries.

## Non-Negotiable Rules

- Existing command names, flags, defaults, prompts, confirmation semantics,
  side effects, stdout results, JSON/NDJSON schemas, exit codes, signals, state
  files, process ownership, and secret boundaries remain unchanged unless a
  source decision explicitly permits a human-readable TTY log change.
- `internal/terminal` is the only owner of Charm modules, terminal capability
  policy, renderer leases, semantic phase mechanics, Transcript storage, and
  stream ordering. Commands own wording, information hierarchy, phase catalogs,
  and result documents.
- There is one Bubble Tea v2 root per finite Rich run, no v1 compatibility
  stack, no parallel form implementation, and no long-lived legacy/new UI
  feature flag.
- Rich owns stderr and AltScreen only while running. Completion freezes semantic
  state, restores the primary screen, replays the safe Transcript, flushes
  deferred diagnostics, and emits stdout last. Plain and Automation never
  duplicate or contaminate machine output.
- Transcript and presentation projections are bounded, control-free, and
  redacted. Secrets, raw errors, full URLs, absolute paths, checksums, child
  output, large results, and internal arguments never enter the wrong stream.
- Service Commands remain line-oriented; JSON/NDJSON contracts remain stable.
  No high-volume request, heartbeat, polling, or access-log stream is added at
  normal levels.
- Failed verification never claims success. Rollback and cleanup are explicit,
  and unrelated worktree changes are preserved.

## Gate Overview

| Gate | Name | Unlock condition | Outcome |
| --- | --- | --- | --- |
| G0 | Baseline and dependency lock | Start | Behavior baselines and the pinned v2 module boundary are recorded. |
| G1 | Shared semantic terminal foundation | G0 Exit conditions all satisfied | `Finish`, `Milestone`, phases, Transcript, leases, and capability degradation are production-ready. |
| G2 | Logging and root boundary | G1 Exit conditions all satisfied | Log v2 is isolated and root/help/discovery preserve their contracts. |
| G3 | Low-risk finite command slices | G2 Exit conditions all satisfied | Read-only/config-selection commands prove the complete finite Experience path. |
| G4 | External-read and service-startup slices | G3 Exit conditions all satisfied | Provider/network reads, `upgrade` parent presentation, and initial `diff`/`fs` logs are covered. |
| G5 | Mutating, destructive, and archive slices | G4 Exit conditions all satisfied | Configuration changes, `rm`, and `zip` preserve safety and side-effect boundaries. |
| G6 | Git mutation and process handoff | G5 Exit conditions all satisfied | Git acquisition/commit and `run` release-before-exec are verified. |
| G7 | Service lifecycles and detached updater | G6 Exit conditions all satisfied | Tunnel logs and the complete `upgrade` replacement/rollback chain are verified. |
| G8 | Release candidate and final acceptance | G7 Exit conditions all satisfied | All commands pass the release checklist on supported targets. |

## G0: Baseline and dependency lock

### Purpose

Establish an auditable pre-change behavior baseline and make the single Charm
v2 module graph explicit before any command presentation changes.

### Inputs

- `CONTEXT.md`
- `docs/project-layout.md`
- `docs/platform-boundaries.md`
- `go.mod`, `go.sum`, `Makefile`
- `internal/architecture/architecture_test.go`
- `issues/01-research-latest-charm-v2-stack.md`, `research/01-latest-charm-v2-stack.md`
- `issues/03-inventory-every-command-terminal-surface.md`
- `issues/30-approve-rollout-and-acceptance-plan.md`

### Objective

Record behavior baselines for root and every command family, pin the approved
Charm v2 dependencies, and enforce the import boundary without changing
command behavior.

### Scope boundary

May change module requirements, architecture checks, baseline/evidence
fixtures, and documentation links. It does not change command runners,
presentation behavior, result bytes, or business logic; those belong to later
Gates.

### Constraints

- Use Bubble Tea `v2.0.9`, Bubbles `v2.2.1`, Huh `v2.0.3`, Lip Gloss `v2.0.6`,
  and Log `v2.0.0` only.
- Do not add v1 wrappers, user-visible feature flags, or a second renderer.
- Baselines must include help, streams, exits, schemas, signals, side effects,
  state files, permissions, and process boundaries.

### Slice policy

One slice captures one command family baseline or one dependency/import change.
Keep module migration and behavior capture independently reversible.

### Verification

#### Directed

- Capture current command-surface, representative stdout/stderr, exit, signal,
  side-effect, and redaction evidence for each command family listed in the
  inventory.
- Verify the module graph resolves the five approved v2 versions and that
  command packages do not import Charm or Log directly.

#### Repository

1. Run `make command-surface` before and after baseline fixture creation.
2. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/architecture ./internal/logging ./internal/terminal`.
3. Run `git diff --check`.

#### Manual acceptance

无。

### Evidence rule

The baseline artifacts, module-resolution output, architecture test result, and
repository checks together prove each G0 Exit condition.

### Stop conditions

- A baseline cannot be captured without changing behavior or leaking a secret.
- The approved module graph cannot resolve under the repository toolchain.
- Existing v1 imports or a conflicting dependency are discovered and no source
  decision explains their replacement.

### Rollback

Revert the module/import/baseline slice only; preserve unrelated worktree
changes and existing command implementations.

### Exit conditions

1. Baselines exist for root/help and every command family in the inventory.
2. The module graph contains only the approved Charm v2 paths and versions.
3. Architecture tests reject direct Charm/Log imports from command packages.
4. Baseline capture and all G0 repository checks pass without behavior drift.

## G1: Shared semantic terminal foundation

### Purpose

Deliver the mechanism that every finite command can use without owning terminal
implementation details.

### Inputs

- `issues/04-choose-charm-v2-and-logging-boundaries.md`
- `issues/05-prototype-vivid-terminal-visual-system.md`
- `issues/06-choose-shared-terminal-experience-contract.md`
- `CONTEXT.md`
- Existing `internal/terminal`, `internal/terminaltest`, and their tests

### Objective

Implement and test one semantic Experience with validated interactions,
ordered Work Phases, `Finish`/`Milestone`, bounded redacted Transcript replay,
renderer leases, and Rich/Plain/Automation degradation.

### Scope boundary

May change `internal/terminal`, `internal/terminaltest`, and shared test
fixtures. It must not migrate a command or alter command-owned result wording.

### Constraints

- One Bubble Tea root and at-most-once AltScreen entry/exit per Rich run.
- `Finish` commits once; `Close` is idempotent cleanup and never invents an
  outcome.
- Valid phase transitions, cancellation draining, transcript limits, and
  sensitive interaction handling are enforced centrally.
- Automation never reads stdin or emits styles, Transcript, or control codes.

### Slice policy

Implement one semantic mechanism at a time: outcome ordering, milestones,
phase protocol, transcript ledger, interaction adapter, lease/replay, then
capability degradation. Each slice gets focused tests before the next.

### Verification

#### Directed

- Exercise `Finish`, `Milestone`, `Close`, phase validation, cancellation, and
  transcript truncation with `internal/terminaltest` recorders.
- Exercise Text, Secret, Select, MultiSelect, and Confirm through Huh v2 child
  models under one Rich root.
- Exercise Rich PTY teardown and Plain/Automation/redirected control-free
  output, including narrow terminals and `NO_COLOR`.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal ./internal/terminaltest`.
2. Run `make check` after shared files are formatted and dependency-locked.
3. Run `git diff --check`.

#### Manual acceptance

Use the local terminal prototype/binary at wide and narrow dimensions with a
colored TTY and `NO_COLOR`; confirm focus, state symbols, loading visibility,
AltScreen restoration, Transcript-before-diagnostics-before-stdout ordering,
and the absence of secret/large-result replay. Record the scenario and explicit
human response `通过` or `未通过: <reason>`.

### Evidence rule

Directed semantic/PTY tests prove mechanism behavior; repository checks prove
integration and import stability; the recorded visual review proves the
human-visible Signal Rail contract.

### Stop conditions

- Renderer failure cannot restore terminal modes or preserve the original
  error.
- A semantic operation would require command packages to import Charm directly.
- A phase, Transcript, or result rule conflicts with the shared contract.

### Rollback

Revert the individual shared mechanism slice. Do not retain a partially switched
dual renderer; commands remain on their pre-migration presentation.

### Exit conditions

1. All shared semantic methods and protocol errors have focused tests.
2. Rich/Plain/Automation/redirected behavior and all five form families pass.
3. Transcript, lease, teardown, cancellation, and exactly-once behavior pass.
4. Manual visual review is recorded with explicit acceptance.

## G2: Logging and root boundary

### Purpose

Integrate Log v2 behind the existing logging Runtime and migrate root/help/
discovery without changing machine-facing output.

### Inputs

- `issues/04-choose-charm-v2-and-logging-boundaries.md`
- `issues/07-choose-root-help-and-error-presentation.md`
- `internal/logging`, `pkg/cmd/root`, `docs/project-layout.md`

### Objective

Provide atomic redacted text Lifecycle Log projection and preserve root
diagnostic controls, help, discovery, version, completion, and error exits.

### Scope boundary

May change `internal/logging`, root presentation adapters, and related tests.
It does not migrate leaf command business logic or add new root output.

### Constraints

- Runtime remains the owner of filtering, redaction, timestamps, formats,
  serialization, and writers; Log v2 is text-only and private.
- JSON/NDJSON fields and boundaries remain exact.
- Discovery remains durable stdout without AltScreen; root errors remain safe
  single-line diagnostics.

### Slice policy

Separate the Log adapter, redaction/atomic-write integration, and each root
surface (help, discovery, version/completion, parser errors) into independently
revertible slices.

### Verification

#### Directed

- Verify text records, JSON/NDJSON records, dynamic level/format changes,
  recursive redaction, injected writers, and lease-aware ordering.
- Verify root help/discovery/version/completion and unknown-command/flag recovery
  against the captured baseline.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/logging ./pkg/cmd/root ./internal/architecture`.
2. Run `make command-surface`.
3. Run `make check` and `git diff --check`.

#### Manual acceptance

无。

### Evidence rule

Logging unit tests and root command-surface/regression tests prove all machine
and diagnostic contracts; no manual acceptance is required for this gate.

### Stop conditions

- Any JSON/NDJSON schema, filtering, timestamp, redaction, or discovery stdout
  drift is observed.
- The adapter would receive unredacted data or write outside Runtime.

### Rollback

Revert the logging adapter or root adapter slice independently; keep G1 intact.

### Exit conditions

1. Log v2 is private behind Runtime and architecture checks enforce the boundary.
2. Text and JSON/NDJSON logging tests pass with redaction and ordering intact.
3. Root/help/discovery/version/completion behavior matches its baseline.
4. G2 repository checks pass.

## G3: Low-risk finite command slices

### Purpose

Prove the full finite Experience flow on bounded read-only and selection
commands before introducing broad side effects.

### Inputs

- `issues/09-choose-config-fork-list-presentation.md`
- `issues/12-choose-config-cm-list-presentation.md`
- `issues/14-choose-config-cm-use-presentation.md`
- `issues/20-choose-git-heat-presentation.md`
- `internal/terminal` contracts and G0 behavior baselines

### Objective

Migrate `config fork list`, `config cm list`, `config cm use`, and `git heat`
with command-owned phases, safe projections, one-shot results, and complete
capability-mode evidence.

### Scope boundary

Only these four command presentation adapters and their tests may change. Keep
storage, Git inspection, result schemas, and defaults intact.

### Constraints

- Each command keeps its own information hierarchy and phase catalog.
- Loading is always represented when the decision specifies it; no artificial
  delay or fabricated percentage is introduced.
- Large lists/results are not copied into Transcript.

### Slice policy

One command adapter per slice; within a command separate loading, projection,
Transcript, and stream tests when possible.

### Verification

#### Directed

- Run focused package tests for each command's success, empty, failure,
  cancellation, redaction, result, and side-effect boundaries.
- Run Rich PTY tests for the relevant Select/Text/Confirm controls and narrow/
  wide/no-color views; run Plain/Automation/redirected fixtures.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/config/... ./pkg/cmd/git/heat ./internal/terminal`.
2. Run `make command-surface` and `make check`.
3. Run `git diff --check`.

#### Manual acceptance

For each command, run its local binary in a colored TTY and `NO_COLOR` at wide
and narrow dimensions. Confirm safe labels, focus/loading, phase finals,
Transcript bounds, and unchanged result. Reply once with
`通过: <commands>` or `未通过: <command and reason>`.

### Evidence rule

Each command's directed tests and visual record cover its Exit conditions;
shared and repository checks cover cross-command compatibility.

### Stop conditions

- Any command needs a shared-layer exception for its information hierarchy.
- A list/result or provider value crosses the Transcript/redaction boundary.
- A command's baseline changes outside approved TTY presentation.

### Rollback

Revert one command adapter and its tests; leave other G3 slices and G1/G2
available.

### Exit conditions

1. All four command adapters preserve their baseline behavior and result bytes.
2. Rich/Plain/Automation/redirected tests pass for every command and relevant
   form family.
3. Transcript, phase, cancellation, redaction, and visual-review evidence is
   recorded for each command.
4. G3 repository checks pass.

## G4: External-read and service-startup slices

### Purpose

Extend the Experience to provider/network reads and establish bounded service
logging without mixing service lifecycles with finite full-screen UI.

### Inputs

- `issues/08-choose-export-env-presentation.md`
- `issues/17-choose-config-cm-test-presentation.md`
- `issues/21-choose-git-pulse-presentation.md`
- `issues/29-choose-upgrade-presentation.md`
- `issues/18-choose-diff-lifecycle-log-presentation.md`
- `issues/19-choose-fs-lifecycle-log-presentation.md`
- `pkg/cmd/{export,git/pulse,upgrade,diff,fs}` and `internal/updater`

### Objective

Migrate `export env`, `config cm test`, `git pulse`, the parent `upgrade`
presentation, and initial `diff`/`fs` Lifecycle Log surfaces with truthful
network/read phases and unchanged result/error contracts.

### Scope boundary

Do not change detached updater replacement, tunnel service behavior, provider
protocols, archive writes, or existing service ownership; those belong to G5/G7.

### Constraints

- Automation must stop before stdin, scanning, network, or side effects where
  the command decision requires it.
- Service logs are line-oriented and bounded; no AltScreen for `diff` or `fs`.
- `upgrade` retains strict resolution, candidate checks, unusual abort exits,
  and stdout text.

### Slice policy

One command or service startup cluster per slice. Keep result projection,
phase/lifecycle events, and redaction tests separable from business changes.

### Verification

#### Directed

- Exercise all provider/network success, timeout, malformed response, partial
  result, cancellation, Automation, and redaction paths.
- Exercise `upgrade` resolution/already-current/download/verify/schedule parent
  presentation and exact exit/result mapping.
- Exercise `diff`/`fs` startup, capability, task/refresh, failure, text and
  NDJSON event projection at each configured level.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/export ./pkg/cmd/config/cm/... ./pkg/cmd/git/pulse ./pkg/cmd/upgrade ./pkg/cmd/diff ./pkg/cmd/fs`.
2. Run `make command-surface` and `make check`.
3. Run `git diff --check`.

#### Manual acceptance

Review one colored and one no-color run for each finite command and one
line-oriented text/NDJSON session for each service command. Check truthful
loading, safe provider/error projections, log density, filtering, and result
ordering. Reply once with `通过: <scenarios>` or `未通过: <scenario and reason>`.

### Evidence rule

Directed command and service event tests prove semantics and schemas; PTY and
manual records prove finite visual ordering and service readability.

### Stop conditions

- A service command requires AltScreen or changes its NDJSON schema.
- Network/provider failures alter existing exit/result semantics.
- Any raw provider credential, URL, response body, or filesystem detail leaks.

### Rollback

Revert the affected command/service adapter while preserving G3 and the shared
foundation.

### Exit conditions

1. All finite commands in G4 pass their result, phase, cancellation, capability,
   and Transcript evidence.
2. `diff` and `fs` pass bounded text/NDJSON lifecycle and shutdown tests.
3. `upgrade` parent behavior and unusual exit/result contracts remain exact.
4. Manual review and G4 repository checks pass.

## G5: Mutating, destructive, and archive slices

### Purpose

Migrate commands that collect confirmation or mutate configuration/files while
preserving validation-first safety and partial-failure semantics.

### Inputs

- `issues/10-choose-config-fork-add-presentation.md`
- `issues/11-choose-config-fork-remove-presentation.md`
- `issues/13-choose-config-cm-add-presentation.md`
- `issues/15-choose-config-cm-set-presentation.md`
- `issues/16-choose-config-cm-remove-presentation.md`
- `issues/24-choose-rm-presentation.md`
- `issues/28-choose-zip-presentation.md`
- Relevant `pkg/cmd/config`, `pkg/cmd/rm`, `pkg/cmd/zip` code and tests

### Objective

Deliver modern forms, confirmations, phase state, Transcript, and safe result
projection for all G5 commands without changing mutation order or side effects.

### Scope boundary

Presentation adapters, shared form usage, and focused tests only. Do not alter
configuration schemas, deletion/archive algorithms, filename/glob semantics,
or confirmation defaults.

### Constraints

- Validation happens before mutation exactly as today.
- Default-No and cancellation semantics remain distinct.
- Automation never reads stdin and performs no later side effect when the
  command decision requires an interactive terminal.
- Secrets, credentials, absolute paths, raw errors, and complete file lists
  stay out of Transcript and inappropriate streams.

### Slice policy

One command per slice, with form collection, validation/mutation phase, and
result/replay tests independently reviewable.

### Verification

#### Directed

- Exercise every Huh field/confirm/multiselect path, invalid retry, cancellation,
  default, overwrite, partial mutation, no-file, and failure case in each
  command ticket.
- Verify archive bytes, deletion effects, configuration ordering, and output
  result bytes remain unchanged.
- Run Rich PTY scenarios for Text, Secret, Select, MultiSelect, and Confirm;
  run Plain/Automation/redirected redaction fixtures.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/config/... ./pkg/cmd/rm ./pkg/cmd/zip ./internal/terminal`.
2. Run `make command-surface`, `make check`, and `git diff --check`.
3. Run `make acceptance` for the applicable tagged configuration, rm, and zip journeys.

#### Manual acceptance

Run each command in a colored and no-color PTY at wide and narrow dimensions,
including a cancellation and an Automation invocation. Check form focus,
destructive wording, phase/result order, and no secret/path leakage. Reply once
with `通过: <commands>` or `未通过: <command and reason>`.

### Evidence rule

Command directed tests prove safety and side effects; acceptance journeys prove
filesystem/configuration behavior; PTY and manual records prove presentation
and redaction.

### Stop conditions

- A UI change would require moving validation, confirmation, or mutation order.
- Automation or cancellation could trigger an unintended side effect.
- Archive/deletion/configuration bytes or schemas change without a new decision.

### Rollback

Revert only the affected command slice and its tests; preserve the shared
Experience and completed G3/G4 commands.

### Exit conditions

1. Every G5 command preserves validation, confirmation, side effects, result
   bytes, exits, and redaction boundaries.
2. All form families and cancellation/Automation paths have directed and PTY
   evidence.
3. Archive/deletion/configuration acceptance journeys pass.
4. Manual records and G5 repository checks pass.

## G6: Git mutation and process handoff

### Purpose

Verify workflows with irreversible Git effects and the terminal-to-child
process ownership boundary.

### Inputs

- `issues/22-choose-git-fork-presentation.md`
- `issues/23-choose-git-cm-presentation.md`
- `issues/25-choose-run-handoff-presentation.md`
- `pkg/cmd/git/fork`, `pkg/cmd/git/cm`, `pkg/cmd/run`, `internal/gitprocess`

### Objective

Migrate Git Fork/CM and Run so selection/review/phase summaries are durable,
Git mutation semantics remain exact, and the selected child receives a clean
terminal after parent release.

### Scope boundary

Presentation and handoff orchestration only. Do not change Git commands,
provider contracts, archive fallback, staging rules, child IO, cwd, or exit
mapping.

### Constraints

- No unstage/reset/amend/delete/retry/rollback behavior is introduced.
- `run` releases terminal ownership before child start and preserves inherited
  stdin/stdout/stderr/cwd and child exit code byte-for-byte.
- Partial Git failures and safe path projections remain command-owned.

### Slice policy

Implement Git Fork, Git CM, and Run as three independently revertible slices;
separate parent presentation from child/process tests.

### Verification

#### Directed

- Exercise archive-first/fallback and overwrite paths, complete message review,
  staged snapshot consistency, push/commit failures, cancellation, and safe
  Transcript summaries.
- Exercise Run selection, release-before-exec ordering, inherited streams,
  child signals, cwd, and exit code with a real fixture child.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/git/fork ./pkg/cmd/git/cm ./pkg/cmd/run ./internal/gitprocess`.
2. Run `make command-surface`, `make check`, `make acceptance`, and `git diff --check`.

#### Manual acceptance

Run Git Fork/CM with representative partial and success cases, then run a real
`run` child in a PTY. Confirm the parent Transcript ends before child startup,
the child owns the terminal, and no parent decoration appears. Reply once with
`通过: <scenarios>` or `未通过: <scenario and reason>`.

### Evidence rule

Focused Git/process tests and tagged acceptance prove side effects and process
ownership; PTY/manual evidence proves release ordering and readable summaries.

### Stop conditions

- Parent retains terminal ownership after child start or changes child IO/exit.
- A presentation change alters Git mutation order or creates an unapproved
  rollback/retry.

### Rollback

Revert the individual Git or Run slice; do not revert shared terminal state.

### Exit conditions

1. Git Fork and Git CM preserve all mutation, cancellation, result, and redaction
   contracts with complete directed evidence.
2. Run passes real child handoff, stream, cwd, signal, and exit tests.
3. Rich/Plain/Automation/redirected and PTY/manual evidence passes.
4. G6 repository checks pass.

## G7: Service lifecycles and detached updater

### Purpose

Complete long-lived Tunnel Lifecycle Logs and the highest-risk detached update
transaction while keeping service and child ownership explicit.

### Inputs

- `issues/26-choose-tunnel-connect-lifecycle-log-presentation.md`
- `issues/27-choose-tunnel-server-lifecycle-log-presentation.md`
- `issues/29-choose-upgrade-presentation.md`
- `pkg/cmd/tunnel`, `internal/tunnelruntime`, `pkg/cmd/upgrade`, `internal/updater`,
  `internal/ycycmd`

### Objective

Deliver bounded, redacted, exactly-once Tunnel logs and prove `upgrade` parent,
hidden updater, rollback, cleanup-warning, and startup-consumption behavior on
supported platforms.

### Scope boundary

May change service log adapters, updater presentation integration, and related
process/state tests. Do not alter tunnel protocols, browser applications, or
release artifact formats.

### Constraints

- No custom full-screen service dashboard; no default access-log flood.
- Control plane and FRPS state remain separate; shutdown/failure records are
  stable and exactly once, with `server.stopped`/`client.stopped` last.
- Detached updater never reports success before second verification and never
  writes a parent Command Result.
- Existing HTTP/checksum abort exit 0 and other classified abort exit 1 remain.

### Slice policy

Implement Tunnel Connect, Tunnel Server, and Upgrade detached transaction as
three independent high-risk slices. Within Upgrade, keep parent scheduling,
hidden replacement, and startup consumption separately testable.

### Verification

#### Directed

- Exercise remembered-connection selection, reconnect/failure windows, FRP and
  reconciliation states, agent aggregation, server control/FRPS independence,
  log filtering, NDJSON schema, and exactly-once shutdown.
- Exercise upgrade strict resolution, candidate checksum/version verification,
  copy/chmod/state/spawn cleanup, cancellation, detached wait, replacement,
  second verification, rollback, cleanup warning, and one-time startup result.
- Exercise Rich PTY transcript ordering for parent Upgrade; service text and
  NDJSON must be control-free/stable.

#### Repository

1. Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/tunnel/... ./pkg/cmd/upgrade ./internal/tunnelruntime ./internal/updater ./internal/ycycmd`.
2. Run `make acceptance` and `make check`.
3. Run `make command-surface`, `git diff --check`, and relevant OS-specific tests.

#### Manual acceptance

Run Tunnel Connect/Server in a TTY and inspect startup, recovery, warning, and
shutdown density at text levels and `NO_COLOR`. Run Upgrade through already-
current, scheduled, and startup-result scenarios, checking restored terminal
and Transcript order. Reply once with `通过: <scenarios>` or
`未通过: <scenario and reason>`.

### Evidence rule

Service event tests prove lifecycle ordering/schema/aggregation; updater unit,
integration, native, and process acceptance prove replacement safety; PTY and
manual records prove terminal restoration and bounded projection.

### Stop conditions

- Any service log floods normal levels, exposes credentials, or changes NDJSON.
- Detached replacement cannot prove rollback, state exactly-once consumption,
  or process ownership on a supported target.
- A cancellation or shutdown race can duplicate a terminal record.

### Rollback

Revert the affected Tunnel or Upgrade slice independently. If a shared logging
or terminal contract regresses, stop and revert that shared seam before any
further high-risk work.

### Exit conditions

1. Tunnel Connect and Server pass lifecycle, redaction, filtering, aggregation,
   shutdown, and cross-mode evidence.
2. Upgrade passes parent result/exit compatibility and full detached replacement,
   rollback, cleanup-warning, and startup-consumption evidence.
3. Supported-platform process/state tests and manual review pass.
4. G7 repository and tagged acceptance checks pass.

## G8: Release candidate and final acceptance

### Purpose

Prove that the complete effort is coherent across commands, capabilities,
platforms, streams, and release packaging before publishing.

### Inputs

- All prior Gate evidence and behavior baselines
- `issues/30-approve-rollout-and-acceptance-plan.md`
- `Makefile`, `CONTRIBUTING.md`, `docs/project-layout.md`,
  `docs/platform-boundaries.md`

### Objective

Produce a release candidate with every root surface and command leaf migrated,
verified, visually reviewed, and free of known semantic or security defects.

### Scope boundary

Only integration fixes, evidence normalization, release-candidate checks, and
slice-level rollback are allowed. No new command behavior or product decision
may be introduced.

### Constraints

- No beta flag, dual renderer, skipped test, weakened assertion, or unrecorded
  PTY limitation is acceptable.
- Cross-platform support is darwin/linux/windows on amd64/arm64 as defined by
  the existing build and platform-boundary contracts.
- Any machine-output, exit, signal, side-effect, state, permission,
  restoration, or security regression blocks release.

### Slice policy

Treat a failing command or platform as the smallest rollback unit. Keep final
integration fixes separate from command-owned presentation changes.

### Verification

#### Directed

- Compare all post-migration outputs and side effects with the captured
  Behavior Baselines.
- Run the full Rich/Plain/Automation/redirected matrix, all form families,
  Transcript/redaction checks, Service Log/NDJSON checks, `run` handoff, and
  detached `upgrade` journeys.
- Complete visual review records for wide/narrow, colored/no-color, loading,
  wrapping, status symbols, log density, and AltScreen replay.

#### Repository

1. Run `make check`.
2. Run `make check-terminal`.
3. Run `make acceptance` and `make acceptance-terminal`.
4. Run `make command-surface`.
5. Run `make cross-build`.
6. Run `git diff --check`.

#### Manual acceptance

Use the release candidate binary for a representative scenario from every
finite command family, every Service Command, `run`, and `upgrade`. Check the
minimum visual/replay/redaction checklist in the rollout decision at wide and
narrow dimensions and with `NO_COLOR`. The final response must be exactly
`通过: release candidate accepted` or `未通过: <scenario and reason>`.

### Evidence rule

All prior Gate Exit evidence, full repository commands, cross-build output, and
the final manual acceptance response are required. A missing or unsupported
PTY is recorded explicitly and cannot replace Plain/Automation/redirected
evidence.

### Stop conditions

- Any known semantic, compatibility, side-effect, terminal-restoration, or
  security defect remains.
- Any required Gate evidence, platform build, visual record, or manual response
  is missing or contradictory.

### Rollback

Revert the smallest affected command slice. Revert shared foundation changes
only for shared-contract or compatibility regressions; never erase unrelated
completed slices or user worktree changes.

### Exit conditions

1. G0-G7 have complete, traceable evidence.
2. Every root surface and command leaf has an implementation, decision link,
   tests, and visual-review record.
3. `make check`, `make check-terminal`, `make acceptance`,
   `make acceptance-terminal`, `make command-surface`, `make cross-build`, and
   `git diff --check` pass.
4. Capability, stream, redaction, Service Log/NDJSON, `run`, and `upgrade`
   release evidence passes on supported targets.
5. Final manual acceptance explicitly confirms the release candidate.

## Definition Of Done

- Every Gate G0-G8 has satisfied its Exit conditions and the runbook records
  the corresponding evidence.
- The shared Charm v2 and logging boundaries are stable, single-stack, and
  covered by architecture and capability tests.
- Every command's behavior baseline remains compatible, with only approved
  TTY visual or human-readable Lifecycle Log changes.
- Rich Transcript, Plain, Automation, redirected, Service Log, JSON/NDJSON,
  cancellation, redaction, process, rollback, and cross-platform evidence is
  complete.
- The release candidate satisfies the final automated checklist and explicit
  human acceptance.

## Explicitly Out Of Scope

- Changing command behavior, parser contracts, business rules, external
  effects, result schemas, exit semantics, or secrets policy.
- Building custom full-screen dashboards for Service Commands or logging
  high-volume request, heartbeat, or polling activity at normal levels.
- Redesigning browser applications served by `fs`, `diff`, or Tunnel.
- Guaranteeing bespoke rendering for every terminal emulator, shell, screen
  reader, or legacy console in the first implementation.
- Adding a beta mode, legacy UI flag, permanent dual renderer, or unrelated
  product feature during this effort.
