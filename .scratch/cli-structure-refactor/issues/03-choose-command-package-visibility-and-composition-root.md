# Choose command package visibility and the composition root

Type: grilling
Status: resolved
Blocked by: 01, 02

## Question

After comparing the upstream model and the current tree, decide the target
visibility and responsibility of each command layer. Should ycy adopt
`pkg/cmd/<domain>/<leaf>` for command implementations, keep all shared
implementation under `internal`, and reduce `cmd/ycy` to process startup,
dependency construction, root command registration, and exit handling? Define
what (if anything) may be imported across command packages, whether
`internal/cliapp` remains a package or is split, and how the layout adoption
avoids accidentally promising a stable external Go API.

The answer must name the allowed dependency direction and the composition-root
responsibilities without changing command behavior.

## Evidence

- [Research the current ycy tree and ownership](02-research-ycy-current-tree-and-ownership.md)
  found that `cmd/ycy` currently combines 35 production files, process-level
  helpers, Terminal Adapters, and all command handler construction;
  `internal/cliapp` owns the Cobra tree and one typed handler per leaf.
- `cmd/ycy/architecture_test.go` currently forbids every `pkg` path segment
  alongside generic buckets such as `utils` and `services`. This old rule
  conflicts with the approved `pkg/cmd` destination and must be replaced by a
  precise command-package rule rather than silently deleted.

## Comments

- Claimed for the command-layer and composition-root decision on 2026-08-28.
- Round 1 initially selected structural parity while preserving ycy's typed
  handler graph. The user challenged that split on 2026-08-28, and the choice
  is superseded: carrying the graph into `pkg/cmd` would transplant the
  current dual ownership and force a second high-overlap migration later.
- The revised decision adopts GitHub CLI's Factory/Options/`runF` dependency
  model in this effort. Internal Interfaces may change; the invariant is the
  external command contract, including syntax, output, exit codes, side
  effects, cancellation, and observable initialization/error ordering.
- `pkg/cmd` may be imported by external Go modules as a technical consequence
  of its path, but ycy offers no Go SDK or compatibility promise for those
  packages. No wrapper layer will be added merely to hide them.
- Round 2 approved `cmd/ycy/main.go -> internal/ycycmd.Main(version) ->
  pkg/cmd/root -> pkg/cmd/<domain>/<leaf>`. The binary entry retains only
  build-version injection, the call into `ycycmd`, and `os.Exit`;
  `internal/ycycmd` owns process orchestration; `pkg/cmd/root` owns global
  Cobra behavior and execution outcome.
- Each leaf command package owns its complete vertical slice: Cobra
  construction, its command Interface, command-specific implementation,
  Terminal/OS/HTTP Adapters, and colocated tests. Shared
  terminal, configuration, logging, file-session, and similar Modules remain
  under `internal`.
- Round 3 approved a bounded `pkg/cmdutil.Factory` plus
  `pkg/cmd/factory.New` model. The Factory carries only process-level,
  cross-command capabilities, using concrete values or lazy functions; it
  does not store handlers, business Modules, flags, or results. Factory scope
  is deliberately narrower than a generic dependency container.
- Round 3 also approved the leaf constructor pattern used by GitHub CLI:
  each leaf exposes an `Options` value, `NewCmdX(f *cmdutil.Factory,
  runF func(*Options) error)`, and a private/default `runX`. Options expose
  only that leaf's inputs and narrow dependencies. The current per-leaf
  handler types are replaced, with no aliases or forwarding wrappers; the
  external command contract remains unchanged.
- The user explicitly accepted both Round 3 decisions on 2026-08-29.
- Round 4 approved two-level command registration: `pkg/cmd/root` registers
  top-level commands and global behavior; domain parent packages such as
  `pkg/cmd/config`, `pkg/cmd/git`, and `pkg/cmd/tunnel` register only their
  child commands; leaf packages own their own command logic. Parent packages
  may import their children, while sibling leaf imports are forbidden.
- Round 4 approved decomposing and deleting `internal/cliapp`. Its root
  execution model moves to `pkg/cmd/root`, domain and leaf registration moves
  to the matching `pkg/cmd` package, and tests follow the implementation.
  No forwarding wrapper, alias, or long-lived transition package remains.
- Round 4 approved the dependency direction `cmd/ycy -> internal/ycycmd ->
  pkg/cmd/root -> pkg/cmd/<domain>/<leaf> -> internal/*`, with
  `pkg/cmdutil` available to command packages and shared Modules. `internal`
  never imports command packages, `cmd/ycy`, or `legacy/bun`; only
  `internal/ycycmd` assembles production Factory Adapters. The old generic
  `pkg` ban in `architecture_test.go` is replaced with these precise rules.

## Answer

Adopt GitHub CLI's command-package architecture in full. The final execution
chain is:

```text
cmd/ycy/main.go
  -> internal/ycycmd.Main(version)
  -> pkg/cmd/root.NewCmdRoot(factory)
  -> pkg/cmd/<domain>.NewCmd<Domain>(factory)
  -> pkg/cmd/<domain>/<leaf>.NewCmd<Leaf>(factory, runF)
```

`cmd/ycy` retains only binary entry concerns and `os.Exit`. `internal/ycycmd`
owns process orchestration, stream and environment facts, Terminal/OS/network
Adapter construction, Factory assembly, hidden worker/updater dispatch, and
signal setup. `pkg/cmd/root` owns Cobra root construction, global flags,
discovery/help, completion, error normalization, version handling, and the
process-independent `Outcome` returned to `cmd/ycy`.

Each leaf command uses the approved Factory/Options/`runF` pattern. This is an
intentional internal Interface refactor: current per-leaf handler types and
the monolithic `internal/cliapp.Dependencies` graph are replaced rather than
aliased. The no-function-change constraint applies to the public command
contract and observable behavior, not to private Go Interfaces.

`pkg/cmd` and `pkg/cmdutil` are technically importable by other Go modules,
but ycy documents no Go SDK or compatibility guarantee. The structural
acceptance tests must enforce the dependency direction and command ownership,
including the new `pkg/cmd` exception, instead of retaining the old blanket
path-segment ban.
