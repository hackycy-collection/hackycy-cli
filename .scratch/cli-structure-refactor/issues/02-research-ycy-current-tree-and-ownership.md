# Research the current ycy tree and ownership

Type: research
Status: resolved

## Question

Inventory the active Go CLI repository as it exists today. For every public
command leaf and every top-level Go package, record its current path,
responsibility, direct dependencies, command-facing versus shared role,
platform-specific files, test/fixture location, and whether it is part of
composition, domain behavior, terminal presentation, persistence, web
serving, or build tooling. Identify long or command-prefixed filenames that
obscure discovery and the import edges that make a mechanical move risky.

Capture the inventory at
`.scratch/cli-structure-refactor/research/02-ycy-current-tree-and-ownership.md`.
This is a facts-only inventory: do not propose behavior changes or move
files.

## Comments

- Claimed for local source, package, dependency, test, and build-ownership
  research on 2026-08-28.

## Answer

The current-tree inventory is recorded in
[`02-ycy-current-tree-and-ownership.md`](../research/02-ycy-current-tree-and-ownership.md).
The active repository contains 30 Go packages: one binary package, one CLI
assembly package, 15 command packages, seven shared/foundation packages, five
tool packages, and one embedded-Web package. `cmd/ycy` currently flattens 35
production and 42 test files into one composition package, while
`internal/cliapp` owns Cobra registration plus a per-leaf typed handler graph.
The command packages themselves already have clean domain ownership and no
cross-domain production imports.

The inventory identifies eleven mechanical migration risks. The most
important are the shared Git/process helpers in `cmd/ycy`, direct Web asset
composition in Diff/FS/Tunnel, embedded 7-Zip payload paths, configuration
filename ownership, hard-coded build/test paths, and the current architecture
test's explicit ban on any `pkg` path segment. Adopting `pkg/cmd` therefore
requires replacing that old structural rule with the approved command-package
rule while preserving its other dependency and ownership checks. No
production code was modified.
