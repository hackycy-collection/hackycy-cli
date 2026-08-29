# Prototype the target tree and migration slices

Type: prototype
Status: resolved
Blocked by: 03, 04, 05, 06, 07

## Comments

- Claimed for the target-tree and migration-slice prototype on 2026-08-29.

## Prototype

- [ycy target tree and migration slices prototype](../prototypes/08-target-tree-and-migration-slices.md)

## Resolution

Accept the prototype's normative target tree, file-ownership map, bounded
Factory, test destinations, and Slice 0-7 migration sequence. Private helper
filenames and the grouping of adjacent low-risk changes remain illustrative,
but every completed slice must retain its named verification gate.

The accepted Factory has exactly these process-level capabilities: Version,
IOStreams, Terminal, Logging, Environment, EnvironmentLookup,
WorkingDirectory, HTTPClient, Now, a lazy/memoized ConfigStore, and a lazy
GitRunner. Network-interface discovery, ZIP reveal, Run child execution,
deletion, and Tunnel-specific dependencies remain in leaf Options.

Add four evidence-backed internal Modules exposed by the concrete target
tree: `internal/updater`, `internal/fsthumbnail`,
`internal/tunnelruntime`, and `internal/sevenzipruntime`. They respectively
separate public/pre-Cobra update execution, FS/private thumbnail worker
execution, Tunnel server/connect shared runtime, and the existing embedded
7-Zip runtime from command leaf ownership. They supplement the previously
approved shared-Module list without creating a generic shared bucket.

Use the root-first transition described by the prototype. Move
`internal/cliapp` directly into `pkg/cmd/root` and delete its old path in the
first production slice. The lifted root's existing handler Dependencies may
remain only as migration-branch state and must shrink as each leaf adopts
Factory/Options/`runF`; it is deleted before the composition-root slice
passes. No forwarding package, re-export, handler alias, sibling helper, or
import-path compatibility shim is allowed.

The final production tree deletes both `internal/cliapp` and
`internal/commands`, leaves only `main.go` under `cmd/ycy`, preserves the
existing external CLI and support-area behavior, and enforces the dependency
rules recorded in the prototype.

## Question

Produce a concrete, reviewable target directory tree and a small set of
mechanical migration slices based on the resolved package, domain, support,
and test rules. The prototype should show representative moves for one
simple command, one nested command group, one long-running command, one
shared Module, and the composition root. It must show import-path updates,
temporary rename handling, and the point at which `go test`, `make check`,
and structural checks run, while leaving behavior untouched.

Link the prototype from this ticket and record which portions are illustrative
versus normative.
