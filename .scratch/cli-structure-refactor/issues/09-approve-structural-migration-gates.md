# Approve structural migration gates and no-behavior evidence

Type: grilling
Status: resolved
Blocked by: 07, 08

## Comments

- Claimed for the structural migration gate and no-behavior evidence decision
  on 2026-08-29.

## Resolution

Use committed CLI-surface evidence, focused leaf gates, complete Slice gates,
and native final evidence as the acceptance contract for the structural
migration. A green cross-build is compile evidence only; it never substitutes
for running behavior on a native host.

### Frozen command surface

Slice 0 creates `acceptance/testdata/command-surface/` from the pre-migration
binary. The tracked evidence contains:

- `manifest.json`, recording every command path, `Use` shape, aliases,
  hidden/deprecated state, and every local and inherited flag's name,
  shorthand, type, and default;
- normalized help output for every command path; and
- deterministic Bash, Zsh, Fish, and PowerShell completion output.

Generation uses a fixed version and controlled environment. Updating these
files requires an explicit update mode, which must not be used to absorb a
migration diff. During Slices 1-7 they are comparison-only evidence. Existing
and moved black-box tests continue to own prompts, stdout/stderr placement,
machine output, exit status, signals, PTY behavior, side-effect order, workers,
Web routes, and release behavior rather than turning all behavior into text
goldens.

### Gate matrix

| Checkpoint | Required evidence |
| --- | --- |
| Pre-migration baseline | Record the rollback commit; run the ordinary suite three times; run the known Tunnel supervisor timeout case twenty times; require every run to pass before Slice 0 changes production code. |
| Slice 0 | Frozen command-surface evidence; ordinary `go test ./...`; the new tagged acceptance suite; the replacement architecture checks. |
| Each migrated leaf | Constructor and runner tests, tests for every Module touched by the leaf, the leaf's black-box acceptance, and an unchanged command-surface comparison. |
| Each completed Slice | Full ordinary Go suite, full tagged acceptance, architecture suite, and `git diff --check`. |
| Web/FS/Tunnel work | The normal Slice gate plus `make check-web`, Go Web route/embed tests, and finite automated Web acceptance where relevant. |
| Composition-root completion | `make check`, tagged acceptance including PTY/signals/workers, `make build`, `make cross-build`, and architecture checks. |
| Final structure | Exact target package inventory, `go list -deps -test ./...` dependency audit, command-surface comparison, and zero active old paths/imports/flat command files. |
| Final native behavior | Darwin, Linux, and Windows each run the ordinary Go suite and tagged acceptance on a native host; all six release targets cross-build. |
| Final release | A separate clean checkout runs `make release-candidate RELEASE_VERSION=0.1.0` and verifies the resulting artifacts. |

The implementation adds `make acceptance`, whose Go execution is fixed to
the equivalent of:

```sh
GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 \
  go test -count=1 -tags=acceptance ./acceptance/...
```

`make check` retains its ordinary package-level scope and does not silently
gain tagged acceptance. Focused acceptance targets may select standalone,
PTY, native, or command cases with `-run`; they do not introduce additional
build-tag combinations.

### Native evidence without new CI

This refactor does not add `.github/workflows`, does not alter the current
workflow-directory prohibition in `tools/check-no-bun`, and does not redesign
release or deployment automation. Native evidence is gathered from existing
developer machines or temporary execution environments:

- one native Darwin host;
- one native Linux host; and
- one native Windows host.

Each result is recorded under
`.scratch/cli-structure-refactor/evidence/` with the exact commit, OS and
architecture, Go version, command, and result. Missing evidence from any one
operating system leaves the migration incomplete. The six-target
`make cross-build` remains required in addition to, not instead of, these
native runs.

### Web acceptance

`make web-browser-harness` remains a manual preview tool because it prints
URLs and waits for SIGINT. It is not counted as a green test. The migration
adds a finite `make acceptance-web` target that starts the real service or
accepted harness, drives Diff, FS, and Tunnel through a headless browser,
checks initial page and critical resource loading plus browser console errors,
and shuts down within a bounded timeout. Its browser dependency is test-only
and does not enter the product runtime. The Linux native evidence run executes
this target; no CI workflow is introduced.

### Allowed structural differences

The migration may change only:

- file locations and splits;
- package names, import paths, and names required by the approved
  Factory/Options/Interface/Adapter conversion;
- dependency construction and command registration wiring;
- test and fixture ownership;
- architecture enforcement;
- Make/tool/generated-payload paths needed by the accepted tree; and
- project-layout and migration documentation.

No Go API compatibility gate is imposed. The repository does not promise a Go
SDK, and the accepted design explicitly changes internal command Interfaces
and introduces `pkg/cmd` construction APIs. CLI behavior, command-surface
evidence, build outputs, and dependency direction are the compatibility
contract. `go list`, the AST architecture suite, and targeted searches audit
structure; they do not freeze incidental exported Go identifiers.

### Stop and rollback contract

Every completed leaf is a green checkpoint, even when multiple leaves later
form one Slice. Any unexplained difference in command tokens, arguments,
flags/defaults, help/completion, prompts, stdout/stderr, machine output, exit
codes, cancellation/signals, PTY behavior, hidden workers, Web routes/assets,
version injection, side effects, generated payloads, or release artifacts
stops the current leaf immediately.

The implementation returns to the last complete green checkpoint. It must not
make the same Slice green by updating a frozen golden, weakening or deleting
an assertion, retaining duplicate old/new implementations, or folding a
functional fix into the move. A genuine behavior change or pre-existing bug
is handled in a separate task and commit before the structural migration
continues.

The current baseline produced one timeout in
`TestFRPSupervisorRejectsAnActivationExitAndKeepsConfigurationFailuresStopped`
and then passed five focused repetitions. Before production migration starts,
that case must pass twenty repetitions and the full ordinary suite must pass
three repetitions. Any recurrence is stabilized separately before Slice 0;
there is no flaky-test allowlist and a retry does not turn a failed gate green.

### Transitional structure enforcement

The architecture suite owns an explicit, monotonically shrinking allowlist
for the lifted root's temporary handler `Dependencies`. Each leaf gate removes
the corresponding field and old path from that allowlist. A completed Slice
cannot retain a path-only test copy, an old/new implementation pair, a type
alias, or a forwarding package.

The final architecture gate asserts all of the following:

- `internal/cliapp` and `internal/commands` do not exist and have no active
  import or build reference;
- the temporary root handler `Dependencies` and its migration allowlist are
  absent;
- no old package re-exports or forwards to a new package;
- no command leaf imports a sibling leaf;
- the exact approved package inventory exists; and
- the entry-chain and internal/package dependency rules from the accepted
  target-tree prototype hold.

Targeted `rg` zero-result checks remain human-readable supporting evidence;
the AST architecture tests and `go list` inventory are the enforced checks.

## Question

Define the acceptance contract for executing the directory refactor. Decide
which checks prove that only paths, package names, imports, build wiring, and
documentation changed: command inventory snapshots, `go list` dependency
rules, `go test ./...`, `make check`, standalone/PTY tests, web checks,
platform builds, and git diff or API-surface audits. Specify the allowed
mechanical exceptions, rollback point, and how each migration slice remains
buildable without a compatibility shim that could survive unnoticed.

The answer must produce a finite gate matrix and an explicit stop condition
for any observed functional diff.
