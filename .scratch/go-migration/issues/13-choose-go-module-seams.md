# Choose the Go module seams and project layout

Type: grilling
Status: resolved
Blocked by: 01, 04, 05, 06, 07, 08, 09, 10, 11, 14

## Question

Choose the standard Go project skeleton and dependency rules required before the port starts: the `cmd/ycy` composition root, the CLI framework boundary, command-local packages under `internal/`, shared configuration/persistence ownership, embedded-web ownership, platform-specific files, and the active `web/` plus frozen `legacy/bun/` boundaries. Define rules that prevent cycles and generic `utils`, `services`, `interfaces`, or speculative `pkg` layers. Do not predesign every command's final interfaces or dependency graph; command-local code is added during each port and shared modules are extracted only when demonstrated ownership or duplication justifies them.

## Comments

- 2026-08-22, grilling round 1: selected feature-first deep command modules under `internal/commands/`; selected evidence-backed sharing limited initially to CLI composition, global configuration, file sessions, logging, and embedded Web assets; selected `web/` as the colocated pnpm package and Go `webassets` package so `//go:embed dist` remains direct and unconditional; selected owner-local `_windows.go`, `_unix.go`, `_darwin.go`, and `_linux.go` implementations instead of a global platform package.
- 2026-08-22, grilling round 2: selected separate `config/fork`, `config/cm`, and four Git leaf modules plus one module for each other command-level behavior, with large command implementations split only when evidence appears; selected a semantic `appconfig` Interface that exclusively owns the shared document, crypto, locks, migrations, and publication; selected command-owned Input/Result types referenced by `cliapp`'s fixed handler manifest without a reverse dependency; selected consumer-owned external seams and owner-local production/test Adapters, promoted only after two real callers share the same invariants.
- 2026-08-22, grilling round 3: selected lazy package creation with no empty scaffolding; selected an explicitly composed logging runtime and scoped loggers instead of package-global lookup; selected a standard-library architecture test owned by `cmd/ycy` to enforce import and naming rules; selected owner-colocated Interface-level tests, private scripted Adapters, and narrow `testdata` without a global `testutil` or golden corpus.
- 2026-08-22, final confirmation: approved the complete project skeleton, ownership rules, dependency direction, platform placement, architecture enforcement, and test layout without further changes.

## Answer

Use one root product Go module pinned to the selected Go 1.26.7 toolchain, with `CGO_ENABLED=0` as the supported build target. Keep Lefthook in its already-selected isolated `tools/lefthook` Go module and do not add `go.work`, a public `pkg/`, another product module, or a second active frontend package. The target layout is a convention applied lazily as behavior is ported, not a request to create empty directories or placeholder types:

```text
cmd/ycy/                         # composition root; the only os.Exit owner
internal/
  cliapp/                        # Cobra, argv compatibility, fixed handler manifest
  commands/
    exportenv/
    config/
      fork/
      cm/
    git/
      heat/
      pulse/
      fork/
      cm/
    rm/
    run/
    zip/
    diff/
    fs/
    tunnel/
    upgrade/
  appconfig/                     # shared config.json, crypto, locking, migrations
  filesession/                   # FS/Tunnel session-v1 persistence
  logging/                       # explicit runtime and scoped loggers
web/                             # pnpm/Vite package and Go package webassets
  dist/                          # ignored generated output, embedded unconditionally
legacy/bun/                      # frozen read-only implementation reference
tools/lefthook/                  # isolated development-tool module
```

`internal/commands/`, `internal/commands/config/`, and `internal/commands/git/` are organizational directories, not generic Go packages. The two configuration domains and four Git leaves are separate command Modules. `exportenv`, `rm`, `run`, `zip`, `diff`, `fs`, `tunnel`, and `upgrade` each begin as one deep Module. A large Module such as Tunnel may gain owner-private implementation packages only when its real code makes that split useful; the skeleton does not predeclare server, client, repository, transport, or manager layers.

`cmd/ycy` is the only composition root. It creates the process context, signal handling, streams, terminal facts, logging runtime, concrete external Adapters, command Modules, and `cliapp.App`; it invokes `Execute(context.Context, []string)` and is the only production code that calls `os.Exit`. It may know every command in order to compose the fixed application. It must not contain command behavior, persistence logic, HTTP routing, framework-specific binders, or a repository-wide dependency bag.

`internal/cliapp` owns Cobra/pflag, construction of a fresh command tree per execution, the fixed typed handler manifest, handwritten command binders, command-scoped argv normalization, global option parsing, help/diagnostics, and outcome classification. Each command Module owns its `Input`, `Result`, and errors. `cliapp` may import those command packages to type its fixed manifest and invoke supplied handlers; command packages never import `cliapp`, Cobra, or pflag. There is no generic registry, declarative mini-framework, `Invocation`, or `Session` service locator.

The allowed dependency direction is:

```text
cmd/ycy -> cliapp + command Modules + shared Modules + production Adapters
cliapp -> command-owned Input/Result types
command Module -> its owner-private implementation + proven shared Modules
shared Module -> standard library or its own implementation dependencies
```

A command Module must not import a sibling command Module. Except for `internal/cliapp` referencing command-owned Input/Result types in its fixed manifest, a shared Module must not import `internal/commands`. When a second real caller needs behavior already owned by one command, extract it only if both callers share the same invariants and the resulting Interface removes meaningful complexity; place it at their narrowest common owner. Similar-looking helpers or a single production Adapter do not justify a seam. Do not create `utils`, `services`, `interfaces`, generic `adapters`, `common`, or speculative `pkg` layers.

`internal/appconfig` is the sole owner of the existing cross-command `~/.ycy-cli/config.json` contract. Its semantic Interface owns path precedence, current and legacy schema decoding, normalization, PBKDF2/AES-GCM compatibility, machine-ID lookup, secret handling, locking, atomic publication, and operations over Fork instances, CM profiles, and remembered Tunnel connections. Callers do not receive a mutable root document, acquire the lock, encrypt fields, or write the file themselves. Provider HTTP behavior is not part of persistence; it remains with the narrowest command owner and may be extracted when the second real Go caller proves a shared provider contract.

`internal/filesession` owns the exact session-v1 directory, key, lock, record, expiry, credential-revision, issue/revoke, and observation behavior shared by FS and Tunnel. It does not own cookies, HTTP routes, command startup, or product-specific authorization decisions. `internal/logging` owns parsing, filtering, redaction, formatting, sinks, clocks, and child scopes through concrete `Runtime` and `Logger` types. `cmd/ycy` creates one runtime, `cliapp` applies the parsed global level, and only commands that log receive scoped loggers; there is no package-global logger lookup.

Because Go embed patterns cannot traverse `..` and the selected Vite output remains `web/dist`, `web/` deliberately contains both the pnpm/Vite application and the root product module's Go package named `webassets`. That Module directly and unconditionally embeds `dist`, validates fixed app selection, and owns generated-asset existence, MIME serving, immutable asset headers, and common shell headers through the previously proven `Site` Interface. Diff, FS, and Tunnel command Modules import it but retain their own API/MCP/file/WebSocket precedence, method rules, CSP selection, and shell fallback. No thin embedded-FS provider, copied output tree, build-tag stub, external asset directory, or second Go module is added to bridge the two toolchains.

Platform variation stays with the Module that owns the behavior. Use ordinary Go filename/build constraints such as `machineid_darwin.go`, `machineid_linux.go`, `machineid_windows.go`, `replace_unix.go`, and `replace_windows.go`; keep the common Interface in that owner and test platform-independent parsing separately from native execution. Do not centralize `GOOS` switches in `internal/platform`, expose platform flags through command Interfaces, or use feature build tags to change product contents.

The active tree may consult `legacy/bun/` while writing each migration test, but active Go and frontend code must not import, execute, dispatch into, generate from, or package legacy. `web/` contains the retained active React applications; their old copies under `legacy/bun/` are reference-only. The root product build and test lifecycle remains frontend-first because every compiled ycy contains the unconditional embedded output.

Tests live with their owner and exercise the same Interface callers use. Command-specific fake or scripted Adapters stay in `_test.go` files or that owner's `testdata/`; cross-command fixture libraries, a global `testutil`, internal-state assertions across Modules, and a separately maintained golden/black-box corpus are rejected. Native filesystem, process, replacement, and protocol tests remain with the responsible Module, while final whole-binary and artifact checks belong at the composition/release level.

Add a standard-library-only architecture test owned by `cmd/ycy` and run it through the normal Go test gate. It parses the active Go package/import structure and rejects Cobra/pflag outside `internal/cliapp`, sibling-command imports, shared-to-command imports other than `cliapp`'s fixed typed manifest, active legacy references, forbidden generic package names, and unauthorized `webassets` consumers. Go itself rejects import cycles; the architecture test makes the remaining ownership rules executable without adding another linter. This layout introduces no compatibility exception and does not predesign the final internal graph of commands that have not yet been ported.
