# ycy target tree and migration slices prototype

> THROWAWAY PROTOTYPE: this document tests whether the approved GitHub CLI
> structure can absorb the current ycy files without changing observable CLI
> behavior. It is a planning artifact, not production documentation. The
> validated rules will later be condensed into `docs/project-layout.md` and
> the implementation plan.

## Question answered

Can the current flat `cmd/ycy`, monolithic `internal/cliapp`, and
`internal/commands` tree be moved to the approved
`cmd/ycy -> internal/ycycmd -> pkg/cmd/root -> pkg/cmd/<leaf>` chain in
reviewable, testable slices without a forwarding package, import-path alias,
or external behavior change?

Prototype verdict: yes. The key is to move the current root implementation
first, then replace its typed handler fields one command domain at a time.
This keeps every completed slice buildable while the root's temporary
`Dependencies` value monotonically shrinks to zero. No old package forwards
to a new package, and `internal/cliapp` is deleted in the first production
slice rather than retained as a compatibility shell.

The inventory also exposes four implementation seams that the high-level
tree did not name explicitly:

- `internal/updater`: shared by the public `upgrade` leaf and the pre-Cobra
  hidden updater/startup transaction path;
- `internal/fsthumbnail`: shared by the `fs` leaf and the pre-Cobra private
  thumbnail worker path;
- `internal/tunnelruntime`: shared FRP/protocol runtime used by both
  `tunnel server` and `tunnel connect`;
- `internal/sevenzipruntime`: the existing deep 7-Zip materialization Module,
  moved out of the deleted `internal/commands` namespace.

These are real seams with at least two consumers or a separately dispatched
process mode. They prevent `internal/ycycmd` from importing command leaves and
prevent sibling leaves from importing one another.

## Normative versus illustrative content

Normative if this prototype is accepted:

- the final directory ownership and command-token tree;
- the final entry chain, dependency direction, and forbidden imports;
- the bounded `cmdutil.Factory` field set and the rule that leaf-only
  dependencies stay in leaf `Options`;
- the source-to-destination ownership table;
- deletion of `internal/cliapp` and `internal/commands` with no forwarding
  packages or aliases;
- the four newly exposed internal Modules above;
- test/fixture ownership and the explicit `acceptance` build tag;
- the invariant that each completed migration slice has a named verification
  gate and preserves the external CLI contract.

Illustrative and adjustable during implementation:

- exact commit count and which adjacent low-risk leaves share a commit;
- private helper filenames inside a leaf when `command.go`, `run.go`, and the
  ownership rules remain clear;
- the pseudocode below, including unexported helper names;
- focused `go test` package lists inside a slice. The final required gates are
  decided by the later structural-migration-gates ticket.

## Final repository tree

The tree omits unchanged non-Go detail below `web`, `mock`, and `legacy/bun`,
but every active ownership root is shown.

```text
.
|-- acceptance/
|   |-- helpers_test.go
|   |-- standalone_test.go
|   |-- config_test.go
|   |-- diff_test.go
|   |-- diff_unix_test.go
|   |-- fs_test.go
|   |-- git_cm_test.go
|   |-- git_heat_test.go
|   |-- git_pulse_test.go
|   |-- signal_unix_test.go
|   |-- tunnel_test.go
|   |-- upgrade_test.go
|   |-- testdata/<shared-scenario>/
|   `-- web/
|       `-- browser_test.go
|-- cmd/
|   `-- ycy/
|       `-- main.go
|-- docs/
|   `-- project-layout.md
|-- internal/
|   |-- appconfig/
|   |-- architecture/
|   |   `-- architecture_test.go
|   |-- filesession/
|   |-- fsthumbnail/
|   |   |-- codec.go
|   |   |-- pool.go
|   |   |-- protocol.go
|   |   |-- worker.go
|   |   `-- *_test.go
|   |-- gitprocess/
|   |   |-- process.go
|   |   |-- process_unix.go
|   |   |-- process_windows.go
|   |   `-- *_test.go
|   |-- logging/
|   |-- sevenzipmanifest/
|   |-- sevenzipruntime/
|   |   |-- payload.go
|   |   |-- payload_<goos>_<goarch>.go
|   |   |-- payload/<target>/...
|   |   |-- runtime.go
|   |   `-- *_test.go
|   |-- terminal/
|   |-- terminaltest/
|   |-- tunnelruntime/
|   |   |-- manifest.go
|   |   |-- protocol.go
|   |   |-- runtime.go
|   |   |-- supervisor.go
|   |   |-- supervisor_unix.go
|   |   |-- supervisor_windows.go
|   |   |-- toml.go
|   |   `-- *_test.go
|   |-- updater/
|   |   |-- candidate.go
|   |   |-- release.go
|   |   |-- transaction.go
|   |   |-- replace.go
|   |   |-- process_<goos>.go
|   |   |-- permissions_<goos>.go
|   |   `-- *_test.go
|   |-- windowsacl/
|   `-- ycycmd/
|       |-- main.go
|       |-- dispatch.go
|       |-- signals.go
|       |-- signals_unix.go
|       |-- signals_windows.go
|       |-- startup_update.go
|       |-- terminal.go
|       `-- *_test.go
|-- pkg/
|   |-- cmdutil/
|   |   |-- factory.go
|   |   `-- iostreams.go
|   `-- cmd/
|       |-- factory/
|       |   |-- factory.go
|       |   `-- factory_test.go
|       |-- root/
|       |   |-- root.go
|       |   |-- execute.go
|       |   |-- diagnostics.go
|       |   |-- discovery.go
|       |   |-- errors.go
|       |   `-- *_test.go
|       |-- config/
|       |   |-- config.go
|       |   |-- fork/
|       |   |   |-- fork.go
|       |   |   `-- {list,add,remove}/
|       |   |       |-- command.go
|       |   |       |-- run.go
|       |   |       `-- *_test.go
|       |   `-- cm/
|       |       |-- cm.go
|       |       `-- {list,add,use,set,remove,test}/
|       |           |-- command.go
|       |           |-- run.go
|       |           `-- *_test.go
|       |-- export/
|       |   |-- export.go
|       |   `-- env/
|       |       |-- command.go
|       |       |-- run.go
|       |       `-- *_test.go
|       |-- git/
|       |   |-- git.go
|       |   `-- {heat,pulse,fork,cm}/
|       |       |-- command.go
|       |       |-- run.go
|       |       |-- process.go
|       |       `-- *_test.go
|       |-- tunnel/
|       |   |-- tunnel.go
|       |   |-- server/
|       |   |   |-- command.go
|       |   |   |-- run.go
|       |   |   `-- *_test.go
|       |   `-- connect/
|       |       |-- command.go
|       |       |-- run.go
|       |       `-- *_test.go
|       |-- diff/
|       |-- fs/
|       |-- rm/
|       |-- run/
|       |-- upgrade/
|       `-- zip/
|           |-- command.go
|           |-- run.go
|           `-- *_test.go
|-- web/                    # unchanged ownership
|-- mock/                   # unchanged ownership
|-- legacy/bun/             # unchanged frozen reference
|-- tools/                  # independent tools
|-- scripts/                # user-facing installers
|-- build/                  # generated build output
`-- public/                 # versioned static source
```

`pkg/cmd/diff`, `pkg/cmd/fs`, and the other single-token leaves use the same
`command.go`, `run.go`, responsibility-file, and colocated-test pattern shown
for `zip`. The tree deliberately has no `pkg/cmd/shared`, `internal/platform`,
`utils`, `common`, `services`, or `adapters` directory.

## Final entry chain

```text
cmd/ycy/main.go
  os.Exit(ycycmd.Main(version))
    -> internal/ycycmd
       pre-Cobra updater/thumbnail dispatch
       stream + Session + signal setup
       factory.New(...)
       webassets.Validate()
       root.Execute(ctx, factory, args)
    -> pkg/cmd/root
       global flags, diagnostics, help, error/exit normalization
       top-level command registration
    -> pkg/cmd/<domain>
       child registration only
    -> pkg/cmd/<domain>/<leaf>
       Cobra grammar + Options + runF + implementation + Adapters
    -> internal/*
       shared deep Modules only
```

`cmd/ycy/main.go` contains only the linker-injected `version`, `main`, the
call to `ycycmd.Main`, and `os.Exit`. No command token, Cobra import, worker
protocol, terminal setup, or application dependency construction remains in
the binary directory.

## Bounded Factory prototype

The final Factory contains process-level capabilities used by at least two
leaves. It is not a registry of commands and does not contain flags, leaf
inputs, handlers, Modules, results, or one-off OS Adapters.

```go
// pkg/cmdutil/iostreams.go
type IOStreams struct {
    In     io.Reader
    Out    io.Writer
    ErrOut io.Writer // raw inherited stderr; diagnostics use Terminal
}

// pkg/cmdutil/factory.go
type Factory struct {
    Version            string
    IOStreams          IOStreams
    Terminal           *terminal.Runtime
    Logging            *logging.Runtime
    Environment        func(string) string
    EnvironmentLookup  func(string) (string, bool)
    WorkingDirectory   func() (string, error)
    HTTPClient         *http.Client
    Now                func() time.Time
    ConfigStore        func() (*appconfig.Store, error)
    GitRunner          func() *gitprocess.Runner
}
```

This is the complete final field set, not a starting wishlist. `ConfigStore`
is lazy and memoized so command construction has no persistence side effect.
`GitRunner` is lazy so help/version paths do not resolve or start Git.
`HTTPClient`, `Now`, and `WorkingDirectory` cover multiple existing leaves.
Network-interface discovery, ZIP reveal, Run child execution, file deletion,
and Tunnel-specific WebSocket/FRP dependencies remain leaf-owned because they
are not shared Factory capabilities.

`pkg/cmd/factory.New` accepts explicit `Version`, `IOStreams`, terminal
`Session`, and environment lookup options. It constructs the Terminal,
Logging, config, HTTP, clock, working-directory, and Git process Adapters.
`internal/ycycmd` owns gathering the real process facts passed into it; tests
pass deterministic facts.

## Leaf Interface prototype

Every leaf follows the same shape. `rm` is the simple representative:

```go
type Options struct {
    Context context.Context

    Paths []string
    Force bool
    Depth *int

    WorkingDirectory func() (string, error)
    Terminal         *terminal.Runtime
    Remover          PathRemover
}

func NewCmdRM(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command
func runRM(options *Options) error
```

`NewCmdRM` owns `Use`, help, args, flags, and translating Cobra values into
`Options`. A nil `runF` selects `runRM`; constructor tests inject a `runF` and
assert parsed Options without touching disk. `runRM` creates the command's
Terminal Adapter and Module, then returns an error or process-independent
result through root normalization. The existing prompt order, presentation,
side effects, error text, and exit behavior remain unchanged.

Nested parents have no business behavior:

```go
func config.NewCmdConfig(f *cmdutil.Factory) *cobra.Command
func cm.NewCmdCM(f *cmdutil.Factory) *cobra.Command
func add.NewCmdAdd(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command
```

`config` registers `fork` and `cm`; `cm` registers `list`, `add`, `use`,
`set`, `remove`, and `test`. Parents may import children; children never
import parents or siblings.

Long-running leaves retain lifecycle ownership. For example, Diff's default
`runDiff` creates the existing Module, calls `Start`, presents startup facts,
waits on the returned Operation, and closes it on presentation failure. The
move does not fold server lifecycle into root or Factory.

## Source-to-destination map

### `cmd/ycy`

| Current source | Final destination |
| --- | --- |
| `main.go` | rewritten as the thin `cmd/ycy/main.go`; terminal/logging/dispatch content moves to `internal/ycycmd` |
| `signals.go`, `signals_unix.go`, `signals_windows.go` | `internal/ycycmd/` with the same responsibility names |
| `discovery.go` | `pkg/cmd/root/discovery.go` |
| `git_process.go`, `git_process_unix.go`, `git_process_windows.go` | `internal/gitprocess/process*.go` |
| `githeat_process.go`, `gitpulse_process.go`, `gitfork_process.go`, `gitcm_process.go` | matching Git leaf `process.go`, adapting `gitprocess.Output` to leaf-owned result types |
| `run_process.go`, `run_process_unix.go`, `run_process_windows.go` | `pkg/cmd/run/process*.go` |
| `process_errors.go` | its Git use moves to `internal/gitprocess/errors.go`; its Run use moves to `pkg/cmd/run/process_errors.go`; no generic process package is introduced |
| `upgrade.go` public command/presentation portion | `pkg/cmd/upgrade/run.go` and `presentation.go` |
| `upgrade.go` hidden dispatch/startup-consumption portion | `internal/ycycmd/startup_update.go`, backed by `internal/updater` |
| `configcm.go`, `configcmadd.go`, `configcmremove.go`, `configcmtest.go` | split by token into `pkg/cmd/config/cm/<leaf>/`; directory context removes the `configcm` prefix |
| `configfork.go`, `configforkadd.go`, `configforkremove.go` | split into `pkg/cmd/config/fork/<leaf>/` |
| `exportenv.go` | `pkg/cmd/export/env/{run,terminal,filesystem}.go` |
| `githeat.go`, `gitpulse.go`, `gitfork.go`, `gitcm.go` | matching `pkg/cmd/git/<leaf>/` responsibility files |
| `tunnel.go` | split into `pkg/cmd/tunnel/server/` and `connect/` |
| `diff.go`, `fs.go`, `rm.go`, `run.go`, `zip.go` | matching single-token leaf; Adapter names shorten by directory context |

All corresponding package-local tests follow these owners. CLI-level
standalone, signal, and PTY cases follow the acceptance rules below.

### `internal/cliapp`

`internal/cliapp` is deleted during the root-lift slice. Its files are moved,
not wrapped:

| Current source | Final destination |
| --- | --- |
| `app.go` | `pkg/cmd/root/{root,execute}.go` plus `pkg/cmd/root/*_test.go` |
| `diagnostics.go` | `pkg/cmd/root/diagnostics.go` |
| `errors.go` | `pkg/cmd/root/errors.go` |
| `discovery.go` | `pkg/cmd/root/discovery.go` |
| `config.go` | `pkg/cmd/config/config.go` |
| `configcm.go` | `pkg/cmd/config/cm/cm.go` plus each CM leaf's `command.go` |
| `configfork.go` | `pkg/cmd/config/fork/fork.go` plus each Fork leaf's `command.go` |
| `githeat.go`, `gitpulse.go`, `gitfork.go`, `gitcm.go` | `pkg/cmd/git/git.go` plus matching leaf `command.go` |
| `tunnel.go` | `pkg/cmd/tunnel/tunnel.go`, `server/command.go`, and `connect/command.go` |
| `exportenv.go` | `pkg/cmd/export/export.go` and `env/command.go` |
| `rm.go`, `run.go`, `zip.go`, `diff.go`, `fs.go`, `upgrade.go` | matching leaf `command.go` |

Root execution tests follow `pkg/cmd/root`; flag/argument/command grammar
tests follow the leaf whose constructor owns that grammar. Tests that only
verified conditional handler registration are replaced by assertions over
the complete production tree; the final root always registers every public
command.

### `internal/commands`

The parent directory is deleted after these direct moves:

- `config/cm`: `input.go`, `read.go`, and `run.go` become the `list` leaf;
  each `add_*`, `use*`, `set*`, `remove_*`, and `test_*` family moves to its
  matching leaf. Narrow Store Interfaces are declared locally when two leaves
  currently share the old package-level `Reader` name.
- `config/fork`: `input.go`, `read.go`, and `run.go` become `list`; `add_*`
  and `remove_*` move to their matching leaves.
- `git/{heat,pulse,fork,cm}`: each complete package moves to the matching
  `pkg/cmd/git/<leaf>` and gains its Cobra/Terminal/OS Adapter files.
- `exportenv`: moves as a complete implementation to `pkg/cmd/export/env`.
- `diff`, `rm`, `run`, and `zip`: each complete package moves to its matching
  leaf; existing short responsibility filenames remain short.
- `fs`: moves to `pkg/cmd/fs`, except `thumbnail.go`,
  `thumbnail_worker.go`, and `thumbnail_pool.go`, which form
  `internal/fsthumbnail`; `thumbnail_service.go` remains leaf-owned and maps
  the internal Module's typed errors to the existing FS error vocabulary.
- `fs/sevenzipruntime`: moves intact to `internal/sevenzipruntime`, including
  generated payload selectors and tests.
- `tunnel`: `client_*.go` moves to `pkg/cmd/tunnel/connect`; `server_*.go`,
  `database*.go`, `file_permissions*.go`, and `state*.go` move to
  `pkg/cmd/tunnel/server`; `frp_*.go` and `protocol.go` form
  `internal/tunnelruntime`. Tests follow the file owner.
- `upgrade`: the release/download/transaction/replacement implementation and
  its platform files move to `internal/updater`; the public Cobra binding,
  Terminal Adapter, and presentation remain in `pkg/cmd/upgrade`.

Existing shared packages `appconfig`, `filesession`, `logging`, `terminal`,
`terminaltest`, `sevenzipmanifest`, and `windowsacl` keep their paths.

## Test migration map

Every file under `acceptance/` starts with `//go:build acceptance`. Ordinary
`go test ./...` therefore remains offline and package-local.

- `cmd/ycy/*_integration_test.go` moves to the matching top-level acceptance
  file: Diff, FS, Git CM, Git Heat, Git Pulse, and Tunnel.
- `cmd/ycy/diff_integration_unix_test.go` becomes
  `acceptance/diff_unix_test.go` and retains its platform suffix.
- `cmd/ycy/standalone_binary_test.go` becomes
  `acceptance/standalone_test.go`.
- Standalone-binary test functions embedded in `configcm*_test.go`,
  `configfork*_test.go`, `rm_test.go`, `run_test.go`, and `zip_test.go` are
  split into command-named acceptance files; their non-binary assertions move
  with the leaf package.
- `internal/commands/upgrade/standalone_integration_test.go` becomes
  `acceptance/upgrade_test.go`.
- Shared binary build/run helpers are consolidated once in
  `acceptance/helpers_test.go`; they build `./cmd/ycy` into `t.TempDir()`.
- Implementation-level subprocess tests for `gitprocess`, Run process
  handling, Terminal, and `fsthumbnail` remain package-local because their
  subject is the Module Interface, not the assembled CLI.
- `cmd/ycy/architecture_test.go` becomes
  `internal/architecture/architecture_test.go` so it continues to run in the
  ordinary suite while no longer belonging to the binary package.
- Web Vitest and Go route/embed tests stay in `web`; only a real ycy/browser
  journey belongs in `acceptance/web`.

Fixtures used by only one moved package follow it into local `testdata/`.
The current runtime-created fixtures remain `t.TempDir()` setup. Only a
scenario shared by at least two acceptance files may enter
`acceptance/testdata/<scenario>`.

## Migration slices

Each slice is a review boundary, not necessarily one commit. Within a slice,
`git mv` operations and import rewrites may temporarily fail to compile; the
slice is complete only after its gate is green.

### Slice 0: freeze evidence and replace the obsolete structure check

1. Record the current help/output/exit and standalone evidence without
   changing assertions.
2. Move the repo-wide architecture test to `internal/architecture`.
3. Replace the blanket `pkg` ban with the approved dependency rules, while
   allowing the current source locations until their owning slice completes.
4. Create tagged `acceptance/` and move only already-black-box tests.
5. Add explicit Make targets for acceptance execution; do not add acceptance
   to `make check`.

Gate: ordinary `go test ./...` and the explicit acceptance command both pass
against the still-current binary.

### Slice 1: lift root and introduce the Factory

1. Move `internal/cliapp` implementation directly to `pkg/cmd/root` and
   update imports/package names.
2. Add `pkg/cmdutil.Factory` with the final field set and
   `pkg/cmd/factory.New`.
3. Let the lifted root temporarily retain only the existing unmigrated typed
   handlers in a shrinking `Dependencies` struct. Add the Factory alongside
   them; do not alias old types or leave an `internal/cliapp` wrapper.
4. Keep `cmd/ycy` as the temporary composition root and make it construct the
   Factory. Observable execution still passes through the lifted root.

Gate: root tests, all ordinary Go tests, current acceptance tests, and the
updated structural check pass. `rg 'internal/cliapp'` returns no active import
or source path.

### Slice 2: simple command vertical slices

Move `rm` first as the representative, then `export env`, `run`, and `zip`:

1. Move the implementation package from `internal/commands` to the target
   `pkg/cmd` leaf.
2. Move its Cobra grammar from `pkg/cmd/root` and its Adapter from `cmd/ycy`
   into that leaf.
3. Add `Options`, `NewCmdX(f, runF)`, and default `runX` without changing the
   underlying behavior.
4. Make root register the leaf constructor and remove the matching handler
   field/wiring from root and `cmd/ycy`.
5. Move package tests; split binary tests into `acceptance/`.

Gate after each leaf: constructor and leaf tests. Gate after the slice: all Go
tests, acceptance, and architecture checks. The temporary root Dependencies
has four fewer command capabilities.

### Slice 3: nested Config command groups

Move `config fork` and `config cm` one complete group at a time. Parent
packages only register children. Each old prefixed file is split by token,
and Store Interfaces are leaf-local. Move all leaves in a group before
removing the old group registration, so no public token disappears at an
intermediate gate.

Gate: leaf constructor tests cover every current argument/flag mapping;
package tests preserve persistence/prompt behavior; config standalone
acceptance proves the full nested help tree and real binary behavior.

### Slice 4: Git group and shared Git process Module

1. Move the shared process lifecycle to `internal/gitprocess` first and run
   its Unix/Windows and cancellation tests.
2. Move Heat, Pulse, Fork, and CM to sibling leaves. Each leaf retains its own
   Git result type and uses a small Adapter over `gitprocess.Output`.
3. Add `pkg/cmd/git/git.go` to register all four children.
4. Remove all Git handler fields and flat `cmd/ycy/git*.go` files.

Gate: internal Git process tests, all four leaf suites, real Git acceptance,
signal acceptance on native Unix, then the full ordinary suite. The
architecture check rejects every sibling-leaf import.

### Slice 5: long-running and worker-backed commands

Move Diff, FS, Tunnel, and Upgrade:

- Diff keeps Start/Wait/Close in the leaf and retains Web ownership.
- FS extracts `internal/fsthumbnail` and `internal/sevenzipruntime`, then moves
  the server/command implementation to `pkg/cmd/fs`.
- Tunnel extracts `internal/tunnelruntime`, then splits server and connect
  without a sibling import.
- Upgrade extracts `internal/updater`, then leaves only command construction
  and presentation under `pkg/cmd/upgrade`.

Update 7-Zip prepare/release paths in the same FS sub-slice so no gate points
at a stale payload location. Preserve `web/dist` and generated artifact rules.

Gate after each command: package tests plus its existing standalone
acceptance. Gate after the slice: `make check`, acceptance, Web/browser
harness checks, and structural checks. Root's temporary handler Dependencies
is now empty and is deleted.

### Slice 6: finalize the process composition root

1. Move terminal/session/logging setup and signal files from `cmd/ycy` to
   `internal/ycycmd`.
2. Move hidden updater and thumbnail worker dispatch before Cobra execution.
3. Move startup-update consumption and Web validation into `internal/ycycmd`.
4. Make `cmd/ycy/main.go` only call `os.Exit(ycycmd.Main(version))`.
5. Remove any remaining command file from `cmd/ycy`, delete the now-empty
   `internal/commands`, and remove all transitional root fields.

Gate: full ordinary tests, tagged acceptance including workers/signals,
`make check`, native build, cross-build compile, and architecture checks.

### Slice 7: update paths and publish the layout contract

Update `Makefile`, tools, README/DEVELOPMENT links, release checks, and the
7-Zip payload path. Add `docs/project-layout.md` from the accepted rules in
this prototype. Check that `scripts/`, `web`, `mock`, `legacy/bun`, `public`,
and generated-output locations did not otherwise move.

Gate: the final gate set approved by the next ticket, plus searches proving
there are no active imports/references to `internal/cliapp`,
`internal/commands`, or obsolete flat command filenames.

## Temporary compatibility decision

No import-path compatibility shim is needed or allowed. In particular:

- no old package re-exports a new type;
- no forwarding `internal/cliapp` package remains;
- no handler alias preserves the old Interface;
- no leaf imports a temporary sibling helper.

The only transitional structure is the lifted root's shrinking handler
`Dependencies` during Slices 1-5. It is implementation state on the migration
branch, not a compatibility Interface: every migrated leaf deletes its field,
and the struct is deleted before the composition-root slice passes. It never
appears in the final tree or documentation.

## Structural assertions in the final tree

The architecture test must enforce at least these facts:

```text
cmd/ycy imports only internal/ycycmd and the standard library
internal/ycycmd may import pkg/cmd/root, pkg/cmd/factory, pkg/cmdutil,
  internal Modules, and web
pkg/cmd/factory may import pkg/cmdutil and internal Modules
pkg/cmdutil may import internal Modules but never pkg/cmd
pkg/cmd/root imports top-level command parents/leaves and pkg/cmdutil
command parents may import their direct children
command leaves may import pkg/cmdutil, internal Modules, web where approved,
  and external libraries
internal packages other than ycycmd never import pkg/cmd*, cmd/ycy, or
  legacy/bun; ycycmd has only the explicit composition imports above
leaf packages never import sibling leaves
only cmd/ycy calls os.Exit
only command packages import Cobra/pflag
no active production code imports legacy/bun
```

The Web consumer allowlist becomes the final owner paths (`internal/ycycmd`,
Diff, FS, Tunnel server, and the browser harness as applicable). The config
persistence literal rule remains owned by `internal/appconfig`.

## External behavior invariants

Every slice preserves the existing command tokens, flags, arguments, help,
prompt order, output streams, machine-readable output, diagnostic redaction,
exit codes, cancellation/signal results, side-effect order, hidden worker
markers, Web routes, embedded assets, build version injection, binary path,
and generated-output policy. A failing invariant stops the structural move;
it is not repaired by changing the expected output in the same slice unless
the expectation was purely an import or filesystem path assertion.
