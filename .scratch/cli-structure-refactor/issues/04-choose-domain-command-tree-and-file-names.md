# Choose the domain command tree and file naming rules

Type: grilling
Status: resolved
Blocked by: 01, 02, 03

## Comments

- Claimed for the domain command tree and filename decision on 2026-08-29.

## Answer

Adopt a directory for every existing CLI token and keep the user's command
vocabulary unchanged. The target command tree is:

```text
pkg/cmd/config/config.go
pkg/cmd/config/fork/fork.go
pkg/cmd/config/fork/{list,add,remove}/<role>.go
pkg/cmd/config/cm/cm.go
pkg/cmd/config/cm/{list,add,use,set,remove,test}/<role>.go
pkg/cmd/git/git.go
pkg/cmd/git/{heat,pulse,fork,cm}/<role>.go
pkg/cmd/tunnel/tunnel.go
pkg/cmd/tunnel/{server,connect}/<role>.go
pkg/cmd/export/export.go
pkg/cmd/export/env/env.go
pkg/cmd/{rm,run,zip,diff,fs,upgrade}/<role>.go
```

Each leaf directory owns its `NewCmd`, `Options`, `runF`, command-specific
implementation, Adapter, and tests. Directory context replaces command-name
prefixes: `configcmadd.go` becomes files under
`pkg/cmd/config/cm/add/`, `gitpulse_process_unix.go` becomes
`pkg/cmd/git/pulse/process_unix.go`, and so on. Names such as `input.go`,
`run.go`, `report.go`, `prompt.go`, and `process_windows.go` are valid only
inside a leaf with clear ownership; generic `utils.go`, `common.go`, and
cross-domain catchalls are prohibited.

The migration moves command-specific implementation and tests from
`internal/commands/<domain>` and current `cmd/ycy` Adapters into the matching
leaf. Code remains under `internal` only when the shared-Module decision
proves that it serves multiple leaves behind a deep Interface. No speculative
`internal/commands/shared` or generic `internal/common` package is created.
`cm`, `fs`, `rm`, and other short CLI tokens remain unchanged even when their
English names could be expanded. `config/cm` and `git/cm` remain distinct
ownerships. Any command-token rename is outside this structural effort.

## Question

Decide how the existing ycy command tree maps to GitHub CLI's noun/verb
directories and short filenames. Cover `config fork`, `config cm`, `git`,
`tunnel`, `diff`, `fs`, `export env`, `rm`, `run`, `upgrade`, and `zip`,
including whether a command with a compound name gets a nested directory,
whether shared code stays beside the domain or moves to `internal`, and how
to rename files such as `configcmadd.go` without altering package APIs or
behavior.

The answer must produce naming and directory invariants that a mechanical
move can follow and must identify any vocabulary mismatch that needs a
separate decision rather than being silently renamed.
