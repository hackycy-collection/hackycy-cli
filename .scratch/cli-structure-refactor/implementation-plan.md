# CLI Structure Refactor Implementation Plan

## Source Decisions

- [`map.md`](map.md) closes the Wayfinder and fixes the destination, invariants, and exclusions.
- [`issues/01-research-github-cli-project-structure.md`](issues/01-research-github-cli-project-structure.md) and [`research/github-cli-structure.md`](research/github-cli-structure.md) establish the upstream GitHub CLI layout evidence.
- [`issues/02-research-ycy-current-tree-and-ownership.md`](issues/02-research-ycy-current-tree-and-ownership.md) and [`research/02-ycy-current-tree-and-ownership.md`](research/02-ycy-current-tree-and-ownership.md) inventory the current package, test, platform, Web, build, and generated-path ownership.
- [`issues/03-choose-command-package-visibility-and-composition-root.md`](issues/03-choose-command-package-visibility-and-composition-root.md) fixes the entry chain, Factory/Options/`runF` model, command-package visibility, and dependency direction.
- [`issues/04-choose-domain-command-tree-and-file-names.md`](issues/04-choose-domain-command-tree-and-file-names.md) fixes the command-token tree, leaf ownership, and filename rules.
- [`issues/05-choose-shared-module-seams-and-dependency-direction.md`](issues/05-choose-shared-module-seams-and-dependency-direction.md) fixes the shared Module seams, platform ownership, Factory boundary, and forbidden imports.
- [`issues/06-choose-repository-support-area-layout.md`](issues/06-choose-repository-support-area-layout.md) fixes the support-area roles, generated-output boundaries, and project-layout documentation requirement.
- [`issues/07-choose-test-fixture-and-acceptance-topology.md`](issues/07-choose-test-fixture-and-acceptance-topology.md) fixes package-local, tagged acceptance, Web, PTY, native, and fixture evidence ownership.
- [`issues/08-prototype-target-tree-and-migration-slices.md`](issues/08-prototype-target-tree-and-migration-slices.md) accepts the prototype's target tree, exact Factory, four new internal Modules, root-first transition, and Slice 0-7 sequence.
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md) supplies the normative target inventory, source-to-destination map, transitional structure, and slice boundaries.
- [`issues/09-approve-structural-migration-gates.md`](issues/09-approve-structural-migration-gates.md) fixes the no-behavior evidence matrix, native-host and release evidence, stop contract, and rollback policy.

## Outcome

The repository has one thin `cmd/ycy/main.go` entry that delegates process orchestration to `internal/ycycmd`, root command behavior to `pkg/cmd/root`, and complete command vertical slices to `pkg/cmd/<domain>/<leaf>`. Shared behavior remains in named deep `internal` Modules, cross-process evidence is explicitly tagged under `acceptance/`, and repository support areas retain their approved roles. The public CLI, build outputs, Web/generated artifacts, and all observable behavior remain unchanged.

```text
cmd/ycy -> internal/ycycmd -> pkg/cmd/root -> pkg/cmd/<domain>/<leaf>
                         \-> pkg/cmd/factory -> pkg/cmdutil -> internal/*
```

## Non-Negotiable Rules

- Preserve every command token, argument, flag and default, alias, hidden/deprecated marker, help and completion output, prompt order, stdout/stderr placement, machine output, exit code, cancellation/signal result, PTY behavior, hidden worker marker, side-effect order, Web route/asset, version injection, binary path, and generated-output policy.
- `cmd/ycy` remains the only binary entry and the only caller of `os.Exit`; its final Go source is only `main.go`.
- The final dependency direction is `cmd/ycy -> internal/ycycmd -> pkg/cmd/root -> pkg/cmd/<domain>/<leaf> -> internal/*`, with `pkg/cmd/factory` and `pkg/cmdutil` limited to their approved composition roles.
- The final `pkg/cmdutil.Factory` field set is exactly Version, IOStreams, Terminal, Logging, Environment, EnvironmentLookup, WorkingDirectory, HTTPClient, Now, lazy/memoized ConfigStore, and lazy GitRunner. Leaf-only dependencies stay in leaf `Options`.
- No forwarding package, re-export, handler alias, sibling-leaf helper, duplicate old/new implementation, generic `utils`/`common`/`services`/`platform` bucket, or Go SDK compatibility layer may survive a completed checkpoint.
- The lifted root's temporary handler `Dependencies` and architecture allowlist may only shrink; both disappear before composition-root completion.
- Every migrated leaf and every completed Slice is a green rollback checkpoint. Frozen command-surface files are comparison-only after creation and may not be updated to absorb a migration difference.
- Ordinary `go test ./...` and `make check` keep their package-level scope. Tagged acceptance remains explicit through `make acceptance`; finite browser acceptance uses `make acceptance-web`; `make web-browser-harness` is preview-only and never green evidence.
- `legacy/bun` remains a frozen read-only reference. `web`, `mock`, `tools`, `scripts`, `build`, `public`, 7-Zip payloads, and ignored/generated outputs keep their approved ownership and behavior.
- Any unexplained behavior difference stops the current leaf immediately. Do not weaken, delete, skip, regenerate, or replace evidence to make a structural move green; stabilize a pre-existing bug or intentional behavior change in a separate task and commit.
- Native Darwin, Linux, and Windows execution evidence is mandatory in addition to six-target cross-build evidence. This effort adds no CI workflow.

## Gate Overview

| Gate | Name | Unlock condition | Outcome |
| --- | --- | --- | --- |
| G0 | Stabilize pre-migration baseline | Start | A reproducible green source baseline and rollback commit are recorded before production movement. |
| G1 | Freeze surface and establish acceptance architecture | G0 Exit all satisfied | Slice 0 freezes the CLI surface and installs tagged acceptance plus precise architecture enforcement. |
| G2 | Lift root and introduce the bounded Factory | G1 Exit all satisfied | Slice 1 removes `internal/cliapp` while preserving root behavior through the new Factory boundary. |
| G3 | Migrate simple command vertical slices | G2 Exit all satisfied | Slice 2 moves `rm`, `export env`, `run`, and `zip` into complete leaf packages. |
| G4 | Migrate nested Config groups | G3 Exit all satisfied | Slice 3 moves all `config fork` and `config cm` parents and leaves without changing the nested CLI tree. |
| G5 | Migrate Git leaves and shared process runtime | G4 Exit all satisfied | Slice 4 establishes `internal/gitprocess` and moves all four Git leaves with native signal evidence. |
| G6 | Migrate long-running and worker-backed commands | G5 Exit all satisfied | Slice 5 moves Diff, FS, Tunnel, and Upgrade and establishes their four approved shared runtime Modules. |
| G7 | Finalize the process composition root | G6 Exit all satisfied | Slice 6 leaves a thin binary, complete `internal/ycycmd`, and no transitional handler graph. |
| G8 | Publish and audit the final structure | G7 Exit all satisfied | Slice 7 updates repository paths and documentation and proves the exact final package/dependency inventory. |
| G9 | Collect native and release evidence | G8 Exit all satisfied | Darwin, Linux, Windows, cross-build, browser, and clean-checkout release evidence close the effort. |

## G0: Stabilize pre-migration baseline

### Purpose

Prevent a known Tunnel supervisor timeout or another pre-existing instability from being mistaken for a structural regression.

### Inputs

- [`issues/09-approve-structural-migration-gates.md`](issues/09-approve-structural-migration-gates.md)
- `Makefile`
- `internal/commands/tunnel/frp_supervisor_unix_test.go`
- the current Git commit and worktree status

### Objective

Record the rollback commit and prove the unchanged ordinary suite three times plus the known Tunnel supervisor case twenty times before any production path is moved.

### Scope boundary

This Gate records baseline evidence only. It does not create acceptance infrastructure, move code, regenerate Web/7-Zip artifacts for tracking, or change production behavior. A failing pre-existing test is handled in a separate task and commit.

### Constraints

- Use the current pre-migration command tree and assertions.
- Preserve unrelated worktree changes and record them alongside the rollback commit.
- A retry after a failure does not satisfy this Gate.

### Slice policy

Treat the rollback-commit record, the three-run ordinary suite, and the twenty-run focused case as separate evidence slices. Stop on the first failure.

### Verification

#### Directed

- Before running tests, record `git rev-parse HEAD`, `git status --short`, OS/architecture, and `go version` in the G0 Progress Log.
- Inspect every run result; one failure invalidates the whole repetition set.

#### Repository

1. `for run in 1 2 3; do GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./... || exit 1; done`
2. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=20 ./internal/commands/tunnel -run '^TestFRPSupervisorRejectsAnActivationExitAndKeepsConfigurationFailuresStopped$'`
3. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 requires the recorded Git/worktree/environment facts. Exit 2 requires three distinct zero-exit ordinary-suite results. Exit 3 requires one zero-exit `-count=20` result with no skipped or retried failure. Exit 4 requires `git diff --check` and confirmation that no production migration occurred.

### Stop conditions

- Any ordinary-suite or focused-case failure.
- The current commit or worktree changes while the repetition set is running.
- Required Go/Web/7-Zip prerequisites cannot be prepared without changing tracked source.

### Rollback

There is no migration rollback yet. Preserve the recorded commit and worktree, stop, and stabilize the failure in a separate task and commit before restarting G0.

### Exit conditions

- E1. The rollback commit, worktree status, OS/architecture, and Go version are recorded.
- E2. The ordinary Go suite completes successfully three consecutive times.
- E3. The named Tunnel supervisor test completes successfully twenty consecutive times.
- E4. The worktree contains no production migration from this effort and passes `git diff --check`.

## G1: Freeze surface and establish acceptance architecture

### Purpose

Create comparison evidence and structural enforcement before the first production move.

### Inputs

- G0 evidence in `goal-runbook.md`
- [`issues/07-choose-test-fixture-and-acceptance-topology.md`](issues/07-choose-test-fixture-and-acceptance-topology.md)
- [`issues/09-approve-structural-migration-gates.md`](issues/09-approve-structural-migration-gates.md)
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), Slice 0 and test migration map
- `cmd/ycy/architecture_test.go`
- existing `cmd/ycy` black-box, standalone, PTY, signal, and integration tests
- `Makefile`

### Objective

Freeze the complete pre-migration CLI surface and establish tagged top-level acceptance plus architecture checks that permit only the approved, monotonically shrinking transition.

### Scope boundary

May add `acceptance/`, `acceptance/testdata/command-surface/`, Make targets, and `internal/architecture`; may move existing black-box tests and their helpers. It must not move or refactor production command implementation.

### Constraints

- `manifest.json` covers every command path, `Use`, aliases, hidden/deprecated state, and all local/inherited flag names, shorthands, types, and defaults.
- Normalize and track help for every command plus deterministic Bash, Zsh, Fish, and PowerShell completions using a fixed version and controlled environment.
- Golden generation has an explicit update mode used only for initial creation; ordinary comparison never updates files.
- Every Go file under `acceptance/` carries `//go:build acceptance`.
- `make acceptance` is equivalent to `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 -tags=acceptance ./acceptance/...`.
- `make check` does not invoke tagged acceptance.

### Slice policy

Create one evidence layer at a time: command-surface capture, architecture-test relocation/rules, black-box test relocation, then Make targets. Verify each layer against the unchanged binary before proceeding.

### Verification

#### Directed

- Run the command-surface capture once in explicit update mode against the G0 binary, review all tracked artifacts, then run the ordinary comparison path and require no diff.
- Run moved black-box cases from `acceptance/` and confirm their assertions, timeouts, stream placement, and standalone-binary behavior are unchanged.
- Run `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture` after every architecture-rule change.

#### Repository

1. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...`
2. `make acceptance`
3. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
4. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 is proven by the reviewed tracked manifest/help/completion set and a clean comparison run. Exit 2 is proven by ordinary and tagged suite results. Exit 3 is proven by the architecture package result and inspection of its transition allowlist. Exit 4 is proven by Makefile inspection plus `make acceptance`/`make check` dependency separation. Exit 5 is proven by the production-source diff.

### Stop conditions

- The pre-migration binary cannot deterministically reproduce its surface under a controlled environment.
- A moved black-box assertion changes for a reason other than path/build-tag ownership.
- The architecture rule cannot express the approved transition without a broad `pkg` exemption.
- Any production implementation change appears in this Gate.

### Rollback

Return to the G0 rollback commit and remove only G1 evidence/test-ownership changes. Never update a frozen artifact to match unexplained output.

### Exit conditions

- E1. The complete frozen command surface is tracked under `acceptance/testdata/command-surface/` and compares cleanly.
- E2. Ordinary package tests and explicit tagged acceptance both pass against the unchanged binary.
- E3. `internal/architecture` enforces the final rules plus an explicit shrinking transition allowlist and the old blanket `pkg` ban is gone.
- E4. `make acceptance` has the fixed execution contract and `make check` retains ordinary scope.
- E5. No production command implementation changed.

## G2: Lift root and introduce the bounded Factory

### Purpose

Establish the target command construction seam without retaining `internal/cliapp` as a compatibility layer.

### Inputs

- [`issues/03-choose-command-package-visibility-and-composition-root.md`](issues/03-choose-command-package-visibility-and-composition-root.md)
- [`issues/05-choose-shared-module-seams-and-dependency-direction.md`](issues/05-choose-shared-module-seams-and-dependency-direction.md)
- [`issues/08-prototype-target-tree-and-migration-slices.md`](issues/08-prototype-target-tree-and-migration-slices.md)
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), Factory prototype and Slice 1
- `internal/cliapp/`
- `cmd/ycy/main.go`

### Objective

Move root construction/execution directly into `pkg/cmd/root`, add the exact `pkg/cmdutil.Factory` and `pkg/cmd/factory.New`, and route the current binary through the lifted root with no `internal/cliapp` path.

### Scope boundary

May move root/global behavior and tests, add Factory packages, and temporarily retain only unmigrated typed handlers in the root's shrinking `Dependencies`. Command implementations and their current Adapters remain for later Gates.

### Constraints

- The Factory field set and lazy behavior exactly match the accepted prototype.
- Command construction must not read configuration, start Git, perform network access, or introduce other eager side effects.
- `internal/cliapp` is deleted, not wrapped, aliased, or forwarded.
- `cmd/ycy` remains the temporary process composition root only until G7.

### Slice policy

Separate root move, Factory type, default Factory construction, and temporary binary wiring into independently reviewed slices; close the Gate only after they work together and the old path is absent.

### Verification

#### Directed

- `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/root ./pkg/cmd/factory ./pkg/cmdutil`
- Constructor tests prove exact Factory fields, lazy/memoized ConfigStore, lazy GitRunner, deterministic process facts, and unchanged root outcome/error/help behavior.
- `rg -n 'internal/cliapp|github.com/hackycy/hackycy-cli/internal/cliapp' cmd internal pkg tools web Makefile` returns no result.
- `make acceptance` records an unchanged command-surface comparison.

#### Repository

1. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...`
2. `make acceptance`
3. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
4. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 is proven by root/factory tests and the zero-result search. Exit 2 is proven by exact-field and laziness tests. Exit 3 is proven by architecture results and inspection of the transitional allowlist. Exit 4 is proven by ordinary/tagged suites and frozen-surface comparison.

### Stop conditions

- Root execution, discovery, diagnostics, help, version, error normalization, or exit outcome differs.
- Factory construction introduces an eager persistence, Git, network, or process side effect.
- Removing `internal/cliapp` would require a wrapper, alias, or duplicate implementation.
- Any frozen command-surface or black-box acceptance difference appears.

### Rollback

Revert G2 to the G1 green checkpoint as one root/Factory boundary. Do not retain a partial wrapper or restore the old broad architecture rule.

### Exit conditions

- E1. Root behavior and tests live under `pkg/cmd/root`; no source, import, or build reference to `internal/cliapp` remains.
- E2. `pkg/cmdutil.Factory` and `pkg/cmd/factory.New` have the exact approved capabilities and verified lazy semantics.
- E3. The root's temporary `Dependencies` contains only unmigrated handler capabilities and has an explicit shrinking architecture allowlist.
- E4. Ordinary tests, tagged acceptance, architecture checks, command-surface comparison, and diff hygiene pass.

## G3: Migrate simple command vertical slices

### Purpose

Prove the leaf migration pattern on bounded commands before moving nested, shared-process, or long-running domains.

### Inputs

- [`issues/04-choose-domain-command-tree-and-file-names.md`](issues/04-choose-domain-command-tree-and-file-names.md)
- [`issues/07-choose-test-fixture-and-acceptance-topology.md`](issues/07-choose-test-fixture-and-acceptance-topology.md)
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), leaf Interface and Slice 2
- current `internal/commands/{rm,exportenv,run,zip}` and matching `cmd/ycy`/root files and tests

### Objective

Move `rm`, `export env`, `run`, and `zip` in that order into complete `pkg/cmd` vertical slices using `Options`, `NewCmdX(factory, runF)`, and private default runners.

### Scope boundary

Only the four named leaves, their grammar, Adapters, tests/fixtures, binary acceptance, and matching root/composition fields may move. Nested Config, Git, Diff, FS, Tunnel, Upgrade, and process-root ownership remain unchanged.

### Constraints

- Each leaf owns Cobra grammar, Options, runner, command-specific Module/Adapter, platform files, and package tests.
- `run` child-process handling remains leaf-owned; no generic process package is introduced.
- Each leaf removes its old implementation, flat Adapter, handler field, wiring, and transition allowlist entry before its checkpoint.
- No command token, prompt, output, side effect, or exit behavior changes.

### Slice policy

One leaf is one independently verifiable and revertible slice, in the order `rm`, `export env`, `run`, `zip`. Do not combine leaves in one checkpoint.

### Verification

#### Directed

- After `rm`: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/rm`, then `make acceptance`.
- After `export env`: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/export/env`, then `make acceptance`.
- After `run`: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/run`, then `make acceptance`.
- After `zip`: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/zip`, then `make acceptance`.
- For each checkpoint, record constructor/runner tests, touched Module tests, the leaf black-box case, and unchanged command-surface comparison.

#### Repository

1. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...`
2. `make acceptance`
3. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
4. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 requires a complete source/test ownership diff for each leaf. Exit 2 requires each leaf's constructor/runner, black-box, and surface evidence. Exit 3 requires architecture evidence showing four removed handler fields/allowlist entries and no old/new pairs. Exit 4 requires the final repository commands.

### Stop conditions

- A leaf cannot be cut over without retaining both old and new implementations at its checkpoint.
- Any constructor parse, prompt, process I/O, filesystem side effect, machine output, or exit result differs.
- A leaf requires a Factory capability outside the exact approved field set.
- Any package, acceptance, architecture, or frozen-surface check fails.

### Rollback

Return only the current leaf to the immediately preceding green leaf/G2 checkpoint. Preserve prior leaf checkpoints; never keep a partial alias or duplicate path.

### Exit conditions

- E1. All four leaves and their owned tests/Adapters live at the approved `pkg/cmd` paths, with no corresponding old implementation or flat Adapter.
- E2. Every leaf has green constructor, runner/Module, black-box, and frozen command-surface evidence.
- E3. Four handler capabilities and their architecture allowlist entries are removed.
- E4. The complete ordinary, tagged acceptance, architecture, and diff-hygiene gate passes.

## G4: Migrate nested Config groups

### Purpose

Move nested command groups while preserving every public token and configuration/prompt behavior at group cutover.

### Inputs

- [`issues/04-choose-domain-command-tree-and-file-names.md`](issues/04-choose-domain-command-tree-and-file-names.md)
- [`issues/05-choose-shared-module-seams-and-dependency-direction.md`](issues/05-choose-shared-module-seams-and-dependency-direction.md)
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), Config ownership map and Slice 3
- current `internal/commands/config/{fork,cm}`, matching root/`cmd/ycy` files, `internal/appconfig`, and tests

### Objective

Move `config fork` and `config cm` into parent/leaf packages where parents only register children and every current leaf owns its grammar, narrow Options/Store interfaces, implementation, Adapter, and tests.

### Scope boundary

May change only Config group/leaf paths, registration, command-specific Adapters/tests/fixtures, and corresponding transitional fields. `internal/appconfig` remains the persistence owner and other domains do not move.

### Constraints

- Preserve the exact `config fork {list,add,remove}` and `config cm {list,add,use,set,remove,test}` vocabulary and full help tree.
- A group switches only when all of its children are present; no intermediate checkpoint may omit or duplicate a public token.
- Parent packages contain registration only. Store Interfaces are declared at the narrowest owning leaf; configuration filenames remain owned by `internal/appconfig`.
- Prefixed filenames split by token into short responsibility names; no cross-leaf import is allowed.

### Slice policy

Use two rollback slices: complete `config fork`, then complete `config cm`. Within each group implement and test one leaf at a time, but declare the checkpoint only after the full group cuts over with no duplicate path.

### Verification

#### Directed

- `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/config/... ./internal/appconfig`
- After each group cutover, run `make acceptance` and record black-box evidence for every child plus an unchanged full nested help/command-surface comparison.
- Inspect architecture results for parent-to-direct-child imports only, no sibling imports, and removal of the group's handler fields/allowlist entries.

#### Repository

1. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...`
2. `make acceptance`
3. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
4. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 is proven by the final package/source inventory. Exit 2 requires per-leaf constructor, persistence/prompt, black-box, and surface results. Exit 3 requires architecture evidence for ownership/import direction and removal of transitional fields. Exit 4 requires the final repository commands.

### Stop conditions

- A group cutover loses, duplicates, renames, or reorders any public command/flag behavior.
- Persistence, encryption, prompt order, HTTP provider, or config filename ownership changes.
- A shared Config helper would require a sibling import or generic shared bucket.
- Any ordinary, tagged, architecture, or surface comparison fails.

### Rollback

Return the current whole Config group to the previous green group/G3 checkpoint. Do not leave a partly registered parent or path-only test copy.

### Exit conditions

- E1. Both Config groups and all nine leaves occupy the approved package tree with complete local ownership.
- E2. Constructor, persistence/prompt, standalone black-box, and frozen surface evidence is green for every leaf.
- E3. Config parents only register direct children; sibling imports, old implementations, flat Adapters, handler fields, and matching allowlist entries are absent.
- E4. The complete ordinary, tagged acceptance, architecture, and diff-hygiene gate passes.

## G5: Migrate Git leaves and shared process runtime

### Purpose

Separate cross-leaf Git process lifecycle from leaf-owned Git semantics while preserving cancellation and native signal behavior.

### Inputs

- [`issues/05-choose-shared-module-seams-and-dependency-direction.md`](issues/05-choose-shared-module-seams-and-dependency-direction.md)
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), Git ownership map and Slice 4
- current `cmd/ycy/git*_process*.go`, `internal/commands/git/{heat,pulse,fork,cm}`, and tests
- `internal/terminaltest` native PTY/signal helpers

### Objective

Create `internal/gitprocess`, then move Heat, Pulse, Fork, and CM into sibling leaf packages with leaf-owned result adapters and a registration-only `pkg/cmd/git` parent.

### Scope boundary

May move shared external Git process execution/capture/platform-group signaling and the four Git vertical slices. Git arguments, provider/archive/clone semantics, result types, presentation, and tracking remain leaf-owned. Other command domains do not move.

### Constraints

- `internal/gitprocess` owns only process execution, captured output, Unix/Windows groups, and signal behavior.
- Leaves never import siblings; `internal/gitprocess` never imports command packages.
- Real Git behavior, provider HTTP behavior, progress/tracking, raw child I/O, cancellation, and errors remain observable-equivalent.
- Each leaf removes its handler field, flat files, old package, and allowlist entry before its checkpoint.

### Slice policy

First establish and verify `internal/gitprocess`; then migrate Heat, Pulse, Fork, and CM as separate leaf checkpoints. One process Module or one leaf is the maximum slice.

### Verification

#### Directed

- After process extraction: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/gitprocess` including applicable Unix/Windows and cancellation tests.
- After each leaf: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/git/...`, then `make acceptance`; record the matching real-Git black-box and command-surface result.
- On a native Unix host, run `make acceptance` and record the Git signal/cancellation case specifically.

#### Repository

1. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...`
2. `make acceptance`
3. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
4. `git diff --check`

#### Manual acceptance

- None; this Gate must execute on a native Unix environment capable of running the accepted signal case.

### Evidence rule

Exit 1 requires process Module tests and dependency inspection. Exit 2 requires four leaf constructor/Module/real-Git/surface records. Exit 3 requires the native Unix signal result. Exit 4 requires architecture evidence for parent/child and sibling restrictions plus transition shrinkage. Exit 5 requires the final repository commands.

### Stop conditions

- No native Unix environment is available for the required signal evidence.
- Process grouping, cancellation, signal forwarding, raw I/O, error mapping, provider behavior, or Git side-effect order differs.
- A leaf requires importing another Git leaf or placing command semantics in `internal/gitprocess`.
- Any ordinary, tagged, architecture, or surface check fails.

### Rollback

Return the current Module/leaf to the preceding green checkpoint. If shared process extraction itself cannot remain green, return the entire Gate to G4; do not create a temporary generic process layer.

### Exit conditions

- E1. `internal/gitprocess` has the approved narrow ownership and green platform/cancellation tests.
- E2. Heat, Pulse, Fork, and CM are complete sibling leaf packages with green constructor, Module, real-Git, and frozen-surface evidence.
- E3. Native Unix Git signal/cancellation acceptance passes.
- E4. The Git parent only registers children; no sibling import, old Git package/flat file, handler field, or matching allowlist entry remains.
- E5. The complete ordinary, tagged acceptance, architecture, and diff-hygiene gate passes.

## G6: Migrate long-running and worker-backed commands

### Purpose

Move the highest-lifecycle-risk commands while preserving Web, worker, generated-payload, process, and transaction behavior behind approved deep Modules.

### Inputs

- [`issues/05-choose-shared-module-seams-and-dependency-direction.md`](issues/05-choose-shared-module-seams-and-dependency-direction.md)
- [`issues/06-choose-repository-support-area-layout.md`](issues/06-choose-repository-support-area-layout.md)
- [`issues/07-choose-test-fixture-and-acceptance-topology.md`](issues/07-choose-test-fixture-and-acceptance-topology.md)
- [`issues/08-prototype-target-tree-and-migration-slices.md`](issues/08-prototype-target-tree-and-migration-slices.md)
- [`issues/09-approve-structural-migration-gates.md`](issues/09-approve-structural-migration-gates.md), Web acceptance and stop contract
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), Slice 5 and ownership maps
- current Diff, FS, Tunnel, Upgrade, Web, 7-Zip, worker, and release-path implementation/tests

### Objective

Move Diff, FS, Tunnel, and Upgrade into their final command packages; establish `internal/fsthumbnail`, `internal/sevenzipruntime`, `internal/tunnelruntime`, and `internal/updater`; and remove the root's temporary handler graph.

### Scope boundary

Only these four command domains, their approved shared runtime seams, owned tests/acceptance, Web consumers, 7-Zip prepare/release paths, and corresponding composition fields may change. Web application layout, generated-output policy, release workflow, and command behavior do not change.

### Constraints

- Diff retains leaf-owned Start/Wait/Close lifecycle and Web construction.
- FS owns command/server behavior; thumbnail worker/runtime and 7-Zip runtime move to their named internal Modules, while typed FS error mapping remains leaf-owned.
- Tunnel server/connect remain siblings; FRP/protocol shared behavior belongs only to `internal/tunnelruntime`.
- Upgrade command/presentation stays in the leaf; release/transaction/replacement and hidden updater behavior belong to `internal/updater`.
- Update 7-Zip prepare/release paths in the FS slice so no checkpoint references a stale payload path.
- `web/dist` and all generated/ignored boundaries remain unchanged. `make web-browser-harness` is not test evidence.

### Slice policy

Use four command checkpoints in order: Diff; FS plus `fsthumbnail`/`sevenzipruntime`; Tunnel plus `tunnelruntime`; Upgrade plus `updater`. Within a command, separate shared-Module extraction, leaf cutover, and acceptance repair, but do not checkpoint an old/new pair.

### Verification

#### Directed

- After each command, run its package and touched Module tests, then `make acceptance`; record its standalone/worker/surface cases.
- For Diff, FS, and Tunnel, also run `make check-web`, `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./web`, and the finite `make acceptance-web` target after it is introduced.
- At Gate close: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./pkg/cmd/diff ./pkg/cmd/fs ./pkg/cmd/tunnel/... ./pkg/cmd/upgrade ./internal/fsthumbnail ./internal/sevenzipruntime ./internal/tunnelruntime ./internal/updater ./web`.
- Inspect worker markers, bounded shutdown, browser console/resource checks, Web routes, update transaction, payload selectors, and platform files in the recorded acceptance output.

#### Repository

1. `make check`
2. `make acceptance`
3. `make acceptance-web`
4. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
5. `git diff --check`

#### Manual acceptance

- None. `make acceptance-web` must start, drive, and stop its environment within a bounded timeout.

### Evidence rule

Exit 1 requires source/dependency and Module test evidence for all four internal Modules. Exit 2 requires per-command package, standalone/worker, Web where applicable, and frozen-surface results. Exit 3 requires finite browser acceptance with initial page, critical resources, console, and shutdown evidence. Exit 4 requires zero transitional handler fields/allowlist and no old/new paths. Exit 5 requires the repository commands.

### Stop conditions

- Any service lifecycle, readiness, route/asset, worker protocol, FRP behavior, signal, update transaction, platform permission, payload, or release-path result differs.
- Browser acceptance cannot finish within its bounded timeout or reports a resource/console failure.
- The known Tunnel supervisor timeout recurs.
- A shared behavior cannot fit one of the four approved Modules without a sibling import or expanded Factory.
- Any full, tagged, Web, architecture, or surface check fails.

### Rollback

Return the current command and its Module/path changes to the preceding green command/G5 checkpoint. Keep prior command checkpoints; never solve a failure by changing frozen output or retaining duplicate runtime paths.

### Exit conditions

- E1. The four approved internal runtime Modules exist with their exact ownership and green package/platform tests.
- E2. Diff, FS, Tunnel server/connect, and Upgrade occupy final command packages with green command, standalone/worker, Web where applicable, and frozen-surface evidence.
- E3. Finite browser acceptance covers Diff, FS, and Tunnel and shuts down cleanly.
- E4. The temporary root handler `Dependencies` and its allowlist are deleted; no old command implementation or flat Adapter for these domains remains.
- E5. `make check`, tagged acceptance, Web acceptance, architecture, and diff hygiene all pass.

## G7: Finalize the process composition root

### Purpose

Complete the target entry chain and remove all remaining transitional command composition from the binary and old command namespace.

### Inputs

- [`issues/03-choose-command-package-visibility-and-composition-root.md`](issues/03-choose-command-package-visibility-and-composition-root.md)
- [`issues/05-choose-shared-module-seams-and-dependency-direction.md`](issues/05-choose-shared-module-seams-and-dependency-direction.md)
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), final entry chain and Slice 6
- remaining `cmd/ycy` process setup/dispatch files and tests
- all migrated command packages and internal Modules

### Objective

Move process facts, terminal/logging setup, signals, pre-Cobra worker/updater dispatch, startup update consumption, and Web validation into `internal/ycycmd`, leaving only version injection, `main`, `ycycmd.Main(version)`, and `os.Exit` in `cmd/ycy/main.go`.

### Scope boundary

May move only remaining process orchestration and its tests, finalize Factory assembly, remove old paths, and delete transitional fields. Command leaf behavior and support-area paths are unchanged.

### Constraints

- `internal/ycycmd` owns process orchestration but not command business behavior.
- Only `cmd/ycy` calls `os.Exit`; only command packages import Cobra/pflag.
- Hidden updater/thumbnail dispatch occurs before Cobra exactly as before.
- `internal/commands` and all transitional root fields are deleted, not forwarded.

### Slice policy

Separate process-fact/terminal setup, signal handling, hidden dispatch/startup transaction, Web validation/Factory assembly, and final deletion. The Gate checkpoint is the complete thin-entry chain.

### Verification

#### Directed

- `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/ycycmd ./pkg/cmd/root ./pkg/cmd/factory ./pkg/cmdutil`
- `make acceptance` must record worker, startup update, standalone, PTY, signal, version, exit, and frozen command-surface results.
- Inspect `cmd/ycy/main.go` and the architecture output for the exact import/`os.Exit`/Cobra ownership rules.

#### Repository

1. `make check`
2. `make acceptance`
3. `make build`
4. `make cross-build`
5. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
6. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 requires source inspection and architecture evidence. Exit 2 requires process-root and worker/startup tests. Exit 3 requires zero old/transitional path results. Exit 4 requires check, tagged acceptance, native build, six-target cross-build, architecture, frozen-surface, and diff evidence.

### Stop conditions

- Startup ordering, environment/session classification, logging, signal ownership, worker dispatch, Web validation, version injection, or exit behavior differs.
- `internal/ycycmd` would need to import a command leaf or own command behavior.
- Any old namespace requires a forwarding package to keep the build green.
- Any full, tagged, build, cross-build, architecture, or surface check fails.

### Rollback

Return the current process-orchestration slice to the last G6/G7 green checkpoint. If the final chain cannot pass without an old namespace, return the whole Gate to G6 and diagnose the ownership error.

### Exit conditions

- E1. `cmd/ycy` contains only `main.go` with the exact thin-entry responsibilities; `internal/ycycmd` owns all approved process orchestration.
- E2. Worker, updater, startup, terminal/logging, signal, version, and exit evidence is green.
- E3. `internal/commands`, all transitional root fields, and all old composition files/imports are absent.
- E4. `make check`, tagged acceptance, native build, six-target cross-build, architecture, frozen-surface comparison, and diff hygiene pass.

## G8: Publish and audit the final structure

### Purpose

Align repository support paths and maintainer documentation with the completed tree and enforce the exact final inventory.

### Inputs

- [`issues/06-choose-repository-support-area-layout.md`](issues/06-choose-repository-support-area-layout.md)
- [`issues/07-choose-test-fixture-and-acceptance-topology.md`](issues/07-choose-test-fixture-and-acceptance-topology.md)
- [`issues/09-approve-structural-migration-gates.md`](issues/09-approve-structural-migration-gates.md), final structure gate
- [`prototypes/08-target-tree-and-migration-slices.md`](prototypes/08-target-tree-and-migration-slices.md), final tree, assertions, and Slice 7
- `Makefile`, `README.md`, `DEVELOPMENT.md`, build/release tools, and current architecture suite

### Objective

Update all approved build/tool/documentation paths, add and link `docs/project-layout.md`, and prove the exact final package inventory, dependency direction, and absence of obsolete source paths.

### Scope boundary

May update Make/tool paths, release/generated-payload checks, README/DEVELOPMENT links, project-layout documentation, and final architecture assertions. It does not redesign Web, mock, legacy, scripts, deployment, release, or generated-output behavior.

### Constraints

- `docs/project-layout.md` is canonical for entry chain, package ownership, dependency direction, test levels, frozen/reference areas, and generated outputs.
- `scripts/`, `web`, `mock`, `legacy/bun`, `tools`, `build`, and `public` retain their approved roots and roles.
- Architecture tests, not `rg` alone, enforce the exact package inventory and imports.
- No `.github/workflows` entry is added.

### Slice policy

Use separate slices for Make/tool/payload paths, documentation links/content, exact inventory assertions, and zero-old-path audit. Each slice is path/contract-only.

### Verification

#### Directed

- `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go list ./... | sort` matches the exact approved package inventory.
- `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go list -deps -test ./...` is recorded and audited against the approved dependency direction.
- `test ! -e internal/cliapp && test ! -e internal/commands`.
- `test "$(find cmd/ycy -maxdepth 1 -type f -name '*.go' | sort)" = 'cmd/ycy/main.go'`.
- `rg -n 'github.com/hackycy/hackycy-cli/internal/(cliapp|commands)' --glob '*.go' cmd internal pkg tools web` returns no result.
- Review `docs/project-layout.md`, README/DEVELOPMENT links, generated-output lists, 7-Zip paths, and architecture assertions against the accepted prototype.

#### Repository

1. `make check`
2. `make acceptance`
3. `make acceptance-web`
4. `make build`
5. `make cross-build`
6. `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./internal/architecture`
7. `git diff --check`

#### Manual acceptance

- None.

### Evidence rule

Exit 1 requires the exact `go list` inventory and architecture result. Exit 2 requires the dependency audit and zero-old-path checks. Exit 3 requires documentation/path review. Exit 4 requires clean frozen-surface comparison. Exit 5 requires all repository commands and confirmation that no CI workflow or support-area redesign appeared.

### Stop conditions

- The exact approved package inventory or dependency direction cannot be enforced.
- Any old source/import/flat command path remains, or a generic/shared compatibility layer appears.
- A Make/tool/documentation path change alters artifacts, Web behavior, installer behavior, or release semantics.
- Any full, tagged, Web, build, cross-build, architecture, or surface check fails.

### Rollback

Return the current path/documentation/audit slice to the last G7/G8 green checkpoint. Preserve the final production tree; do not loosen architecture assertions to admit a mismatch.

### Exit conditions

- E1. The exact approved target package inventory exists and architecture tests enforce it.
- E2. Dependency audit and zero-result checks prove no old source/import, transitional type/allowlist, forwarding path, sibling import, or flat command file remains.
- E3. `docs/project-layout.md` is complete and linked, and all Make/tool/generated-payload/release paths reference final owners.
- E4. The frozen command surface still compares unchanged.
- E5. Full ordinary, tagged, Web, build, cross-build, architecture, and diff-hygiene evidence is green with no CI or support-area redesign.

## G9: Collect native and release evidence

### Purpose

Prove runtime behavior on every required native platform and verify release artifacts from a separate clean checkout; cross-compilation alone is insufficient.

### Inputs

- [`issues/09-approve-structural-migration-gates.md`](issues/09-approve-structural-migration-gates.md), native, Web, final structure, and release evidence contract
- the exact G8 commit
- native Darwin, Linux, and Windows execution environments
- a separate clean checkout for release-candidate execution
- `.scratch/cli-structure-refactor/evidence/`

### Objective

Record same-commit ordinary and tagged acceptance evidence on native Darwin/Linux/Windows, finite Web acceptance on Linux, six-target cross-build evidence, and verified `0.1.0` release artifacts from a clean checkout.

### Scope boundary

This Gate records and reviews evidence only. It does not change production, tests, frozen surface files, CI, release semantics, or support-area ownership. A failure returns to diagnosis or a separate fix before evidence collection restarts.

### Constraints

- Every record includes exact commit, OS, architecture, Go version, command, result, and relevant artifact/log path.
- All native results use the exact G8 commit; local uncommitted source changes invalidate the set.
- Darwin, Linux, and Windows each run the ordinary Go suite and `make acceptance` natively.
- Linux also runs `make acceptance-web`.
- `make cross-build` must produce all six approved targets.
- `make release-candidate RELEASE_VERSION=0.1.0` runs only in a separate clean checkout and its built-in artifact verification must pass.

### Slice policy

Collect one immutable evidence record per native host, then cross-build evidence, then clean-checkout release evidence. Do not mix results from different commits; any source correction invalidates and restarts the full G9 set.

### Verification

#### Directed

- On each native Darwin/Linux/Windows host: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...`.
- On each native Darwin/Linux/Windows host: `make acceptance`.
- On native Linux: `make acceptance-web`.
- On the designated build host: `make cross-build` and verify six target files.
- In a separate clean checkout: `make release-candidate RELEASE_VERSION=0.1.0` and retain the release-artifacts verification output.
- Audit all files under `.scratch/cli-structure-refactor/evidence/` for one exact commit and complete metadata.

#### Repository

1. `git diff --check`
2. `git status --short`

#### Manual acceptance

- Entry: `.scratch/cli-structure-refactor/evidence/` plus the G9 Progress Log summary.
- Automated boundary: all native commands, Linux browser acceptance, six cross-builds, and clean-checkout release verification must already have explicit results.
- Minimum checklist: one exact commit across all records; Darwin/Linux/Windows ordinary and tagged results; Linux finite browser result; six cross-build artifacts; clean-checkout `0.1.0` release artifacts; no unexplained behavior or generated-file tracking difference.
- Confirm with exactly `G9 验收通过；commit=<40-hex SHA>` or reject with `G9 验收未通过；host=<OS/arch>；command=<command>；evidence=<path>；reason=<reason>`.

### Evidence rule

Exit 1 requires three native ordinary-suite records. Exit 2 requires three native tagged-acceptance records. Exit 3 requires the Linux finite browser record. Exit 4 requires the cross-build log and six artifacts. Exit 5 requires the clean-checkout release log and artifact verification. Exit 6 requires the explicit manual response matching the format above.

### Stop conditions

- Any required native host or clean checkout is unavailable after the handoff contract has been prepared.
- Records refer to different commits, omit required metadata, or include uncommitted source changes.
- Any native, browser, cross-build, release, artifact, frozen-surface, or behavior check fails.
- A failure requires production/test changes or reveals a pre-existing bug; handle it separately and restart G9 on the resulting exact commit.

### Rollback

Discard only invalid G9 evidence and return to the G8 green commit. Do not change source within G9. After any separate source correction, rerun G8 and restart the entire same-commit native/release evidence set.

### Exit conditions

- E1. Native Darwin, Linux, and Windows ordinary Go suites pass on one exact commit.
- E2. Native Darwin, Linux, and Windows tagged acceptance suites pass on that commit.
- E3. Linux finite Web acceptance passes and shuts down cleanly.
- E4. All six release targets cross-build successfully and are recorded.
- E5. A separate clean checkout builds and verifies `make release-candidate RELEASE_VERSION=0.1.0` artifacts.
- E6. The user explicitly accepts the complete evidence set using the declared response format.

## Definition Of Done

- G0-G9 Exit conditions are all satisfied with evidence in `goal-runbook.md` and `.scratch/cli-structure-refactor/evidence/` where required.
- The final entry chain, exact Factory, command tree, shared Modules, support-area ownership, acceptance topology, and dependency direction match the accepted decisions and exact target inventory.
- `internal/cliapp`, `internal/commands`, the temporary handler graph/allowlist, old flat command files, forwarding/alias paths, and duplicate implementations are absent.
- The frozen CLI surface and all black-box behavior remain unchanged.
- Ordinary, tagged, Web, build, cross-build, architecture, native Darwin/Linux/Windows, and clean-checkout release evidence is green for the final exact commit.
- `docs/project-layout.md` is the linked canonical repository structure contract, and no CI workflow or unrelated support-area redesign was introduced.

## Explicitly Out Of Scope

- Changing commands, flags, arguments, prompts, output, exit codes, business rules, side effects, errors, cancellation, workers, Web routes, or other external contracts.
- Refactoring business behavior beyond the approved private Factory/Options/`runF` and ownership conversion.
- Providing a stable external Go SDK or preserving private Go Interface compatibility through wrappers.
- Rebuilding or reorganizing the Web applications, changing generated-output policies, or redesigning deployment/release workflows.
- Re-enabling or deleting `legacy/bun`, changing mock service behavior, or adding CI workflows.
- Dependency upgrades, feature work, bug fixes, new top-level commands, dashboards, or plugin systems.
