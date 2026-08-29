# Choose shared Module seams and dependency direction

Type: grilling
Status: resolved
Blocked by: 01, 02, 03

## Comments

- Claimed for the shared Module seam and Factory dependency decision on
  2026-08-29.
- Continuing the human decision round on 2026-08-29 after the command-tree
  decision accepted Factory/Options/`runF`.

## Question

Classify current shared packages and repeated helpers into deep Modules with
clear ownership. Decide where CLI parsing, terminal experience, logging,
configuration/state, file sessions, platform adapters, embedded web serving,
and process/signal helpers live after the command packages move. For each
candidate seam, state the narrow interface, owning layer, permitted imports,
and the test seam. Reject shallow `utils`/`services` buckets and avoid moving
code merely because it is used by more than one command.

The result must include a dependency-direction rule that can be enforced by
package tests or static checks while preserving all existing behavior.

## Answer

Keep the existing deep shared Modules and their ownership names:

```text
internal/appconfig
internal/filesession
internal/logging
internal/terminal
internal/terminaltest   # test-only
internal/sevenzipmanifest
internal/windowsacl
```

They retain configuration persistence and encryption, file sessions and
platform replacement, diagnostics and redaction, terminal Sessions and
Experience Runs, test-only terminal evidence, 7-Zip manifests, and Windows
ACL behavior respectively. No generic `utils`, `common`, `services`, or
`platform` bucket is introduced. `internal/terminaltest` remains outside
production Factory construction.

The four Git leaves' shared process behavior moves from `cmd/ycy` into a new
`internal/gitprocess` Module. Its narrow Interface owns external Git process
execution, captured output, Unix/Windows process groups, and signal behavior;
Git command arguments, provider behavior, archive/clone semantics, and
command-specific result types remain in their owning leaves. `run` process
handling remains leaf-owned, and root signal setup belongs to
`internal/ycycmd`.

`pkg/cmdutil.Factory` is bounded to process-level capabilities shared by at
least two leaves. It may contain application version, IO streams, the shared
Terminal Experience, logging runtime, environment lookup, lazy config store,
lazy `gitprocess.Runner`, HTTP client, clock, and browser capabilities as
needed. It must not contain handlers, business Modules, flags, results, or
one-off leaf dependencies; the final field set may be smaller than this
candidate list. Leaf-only dependencies live in leaf `Options`.

The root `web` package remains an independent frontend/embedding/static-route
package. It keeps Vite sources, `go:embed`, shell validation, readiness
handlers, `web/dist`, and `web/node_modules` rules. Diff, FS, and Tunnel leaf
packages may call its narrow constructors; `internal/ycycmd` may validate
assets at startup; the Factory does not own Web routing.

Platform implementations stay with their owning Module or leaf:

```text
internal/gitprocess/process_{unix,windows}.go
internal/ycycmd/signals_{unix,windows}.go
pkg/cmd/run/process_{unix,windows}.go
internal/filesession/*_{unix,windows}.go
pkg/cmd/upgrade/*_{unix,windows}.go
```

No repository-wide `internal/platform` package is created. A platform seam
must be justified by cross-leaf reuse and a deep Interface.

The enforced dependency direction is:

```text
pkg/cmdutil -> internal/*
pkg/cmd/root -> pkg/cmd/<domain> -> pkg/cmdutil and internal/*
internal/ycycmd -> pkg/cmd/root, pkg/cmd/factory, pkg/cmdutil, internal/*
```

`internal/*` never imports `pkg/cmd*`, `cmd/ycy`, or `legacy/bun`; leaf
packages never import sibling leaves; production code never imports the
frozen Bun tree. The architecture tests replace the blanket `pkg` ban with
these ownership and direction rules.
