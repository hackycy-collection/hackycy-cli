# Approve the terminal experience rollout and acceptance plan

Type: grilling
Status: resolved
Blocked by: 07, 08, 09, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29

## Question

In what dependency and risk order should the shared Charm v2 foundation and
every independently specified command experience be implemented, and what
unit, renderer, golden, PTY, redirected-stream, automation, redaction,
cross-platform, visual-review, and no-behavior-change evidence must gate each
slice and the final rollout?

## Answer

Implementation proceeds in command-sized slices over one shared Charm v2
foundation. Every slice preserves the command's existing behavior and ships
its implementation, tests, specification link, and visual-review record
together. There is no long-lived legacy/new UI switch or parallel renderer.
Commands that have not yet migrated continue to use their existing
presentation until their slice is complete.

### Dependency and risk order

The rollout is divided into the following batches. Work within a batch may be
parallel once its prerequisite gates pass.

**Batch 0: shared foundation and contracts**

- Upgrade to the pinned Bubble Tea `v2.0.9`, Bubbles `v2.2.1`, Huh `v2.0.3`,
  Lip Gloss `v2.0.6`, and Log `v2.0.0` modules.
- Deliver `internal/terminal` `Finish`/`Milestone`, Work Phase validation,
  Transcript ledger, renderer lease, Rich/Plain/Automation capability policy,
  and control-free degradation.
- Deliver the Log v2 text-only adapter behind `internal/logging`, plus the
  architecture import rule that keeps Charm and Log out of command packages.

**Batch 1: baseline and low-risk read-only slices**

Root/help/discovery, `config cm use`, `config cm list`, `config fork list`, and
`git heat` validate result routing, safe projections, loading phases, and the
first complete Rich/Plain/Automation path.

**Batch 2: external reads and finite operations**

`export env`, `config cm test`, the parent-process presentation of `upgrade`,
`git pulse`, and the initial Lifecycle Log work for `diff` and `fs` extend the
same contracts to network/provider reads and bounded reports.

**Batch 3: configuration and destructive finite work**

`config fork add/remove`, `config cm add/set/remove`, `rm`, and `zip` cover
forms, confirmation, mutation boundaries, archive phases, cancellation, and
partial failures.

**Batch 4: Git and process handoff**

`git fork`, `git cm`, and `run` cover archive/commit side effects and the
release-before-exec terminal handoff. `run` must not decorate, capture, or
rewrite the selected child's streams or exit code.

**Batch 5: service lifecycles and highest-risk update paths**

`tunnel connect` and `tunnel server` complete their Lifecycle Logs. The
`upgrade` hidden updater, detached replacement, rollback, cleanup warning, and
startup transaction consumption are then verified as a separate high-risk
chain, even though its parent presentation was introduced in Batch 2.

**Batch 6: release candidate and final verification**

Run the complete command matrix, cross-platform builds, acceptance journeys,
visual review, and release checklist before publishing.

### No runtime feature flag

The new experience is selected by the existing Rich, Plain Interactive,
Automation, terminal, and color capabilities only. No `--legacy-ui`,
environment fallback, terminal-specific matrix, or dual implementation is
introduced. Renderer fallback is limited to the already-approved startup
failure behavior: Rich may fall back to Plain before AltScreen or semantic
state begins; after startup, it restores the terminal and returns the original
failure without repeating work.

### Required evidence for each slice

Every slice must pass `gofmt`, `go vet`, package tests, architecture import
checks, `git diff --check`, help/command-surface checks, and stdout/stderr/
exit-code regression tests before its batch can advance. The command's existing
parameters, defaults, result schemas, signal behavior, side effects, state
files, permissions, and redaction rules are part of that regression.

Finite commands additionally require:

- command-owned state-machine, Work Phase, result, cancellation, side-effect
  ordering, and safe projection tests;
- Rich, Plain Interactive, Automation, and redirected stream tests;
- Transcript ordering, boundedness, truncation, sensitive-value suppression,
  and `Finish`/`Close` exactly-once tests;
- at least one real Rich PTY journey for each form family (Text, Secret,
  Select, MultiSelect, Confirm), including cancellation, narrow terminals, and
  no-color behavior; and
- structured view tests for status symbols, roles, block order, field
  projection, wrapping, and narrow/wide layouts. Golden files cover only
  normalized semantic views or finite fragments with ANSI, clocks, paths, and
  other dynamic data removed.

Service Commands require event-recorder tests for Lifecycle Log ordering,
levels, filtering, aggregation/throttling, text projection, stable event IDs,
NDJSON schema, and exactly-once shutdown. `run` requires a real child-process
handoff test. `upgrade` requires detached replacement, rollback, startup
consumption, and supported-platform process/state tests.

All commands must prove that Automation never reads stdin and emits no
Transcript, styling, or terminal controls; redirected output is ANSI-free;
secrets, paths, URLs, tokens, checksums, raw errors, headers, child output, and
large result documents do not leak into the wrong stream.

### Test entry points and CI gates

The existing Makefile remains the single task entry point:

- `make check` keeps lock, web, Go format/vet/unit, and architecture checks;
- `make acceptance` runs tagged black-box, process, signal, native, and PTY
  journeys;
- `make command-surface` verifies help and argument snapshots;
- add `make check-terminal` for shared terminal, terminaltest, presentation,
  redaction, redirected, Automation, and PTY package tests; and
- add `make acceptance-terminal` for real PTY, signal, child-process, service,
  and detached-updater journeys. `make cross-build` remains a release gate,
  not a requirement for every local unit run.

A slice must pass its relevant terminal gate, package tests, and
command-surface check. Shared-contract changes rerun all migrated-command
regressions. A missing PTY on a platform is recorded as explicitly unsupported
while Plain/Automation/redirected evidence still runs; it is never treated as
silent success.

### Behavior baseline and visual review

Before changing a command, capture its current help, stdout, stderr, exit code,
JSON/NDJSON, signal, side-effect, state-file, and child-process behavior. The
post-change comparison must show that only the approved TTY Rich/Plain visual
layer and human-readable text Lifecycle Log projection changed. Machine result
bytes, Automation/redirected output, JSON/NDJSON schema, exits, and external
effects remain compatible.

Each Rich or Service slice carries a minimal visual-review record with the
scenario, input path, capability mode, terminal dimensions, expected semantic
states, and known differences. Review covers colored and `NO_COLOR`/Plain
output, Signal Rail hierarchy, symbols, focus, always-visible loading, long
value wrapping, log density, and post-AltScreen Transcript ordering. Screenshots
and PTY captures are review evidence, not cross-terminal byte contracts.
Golden changes require a reason and an explicit statement that stdout, exits,
side effects, and redaction did not change. Command adapters own their visual
fixes; shared terminal changes are reserved for repeated mechanism defects.

### Batch gates and final release checklist

Batch 0 blocks all command migrations. A command slice blocks only dependent
work until its own gate passes. Shared-contract regressions rerun every
migrated command. Any stdout, JSON/NDJSON, exit-code, signal, side-effect,
state-file, permission, terminal-restoration, or security regression is a
global release blocker. A local visual issue can be isolated to its slice, but
cannot be waived when it obscures a semantic state or leaks data.

The release candidate is publishable only when all root/help and command leaves
have an implementation, specification link, tests, and visual record; `make
check`, `make check-terminal`, `make acceptance`, `make acceptance-terminal`,
`make command-surface`, and `make cross-build` pass; all four capability modes
are control-free where required; Service Log and NDJSON contracts pass; `run`
handoff and the full `upgrade` replacement/rollback/startup chain pass on the
supported targets; and no known semantic or security defect remains. There is
no beta flag or permanent dual path. If production reveals a regression, revert
the affected command slice while preserving unrelated migrated slices; only a
shared-contract or compatibility regression requires a broader rollback.
