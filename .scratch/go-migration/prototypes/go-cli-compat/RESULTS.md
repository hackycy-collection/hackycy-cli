# Go CLI compatibility prototype results

## Question

Can ycy preserve its difficult Commander-era command contracts while Cobra remains an implementation detail and each command Module receives typed input rather than framework state?

## Result

Yes. The throwaway probe demonstrated the required parsing and process-control mechanisms. The strongest production Interface is a deep CLI App Module with a fixed, strongly typed command manifest and handwritten command-specific binders hidden inside its Implementation.

The prototype does not settle Git behavior, updater routing semantics, platform-native child-tree behavior, `run` argv policy, or final signal exit codes. Those remain governed by the frozen Bun implementation unless a focused parity probe later proves an incompatibility.

## Observed behavior

| Concern | Commander baseline | Raw Cobra/pflag | Proven Go approach |
| --- | --- | --- | --- |
| `--push` | implicit `origin` | supported with `NoOptDefVal` | retain absent/implicit/explicit state in typed input |
| `--push upstream` / `-p upstream` | explicit remote | leaves `upstream` as an operand | command-scoped argv normalization to `--push=upstream` |
| `-pupstream` | explicit remote | parses `-u...` as shorthand flags and fails | command-scoped normalization to `--push=upstream` |
| `--push=upstream` | explicit remote | supported | direct pflag parsing |
| `-p=upstream` | literal `=upstream` | `upstream` | a command binder can preserve the Commander oddity when the legacy test requires it |
| Option-like remote | accepted and can reach Git option parsing | some equals forms are accepted | preserve the observed Git command behavior; escalate only if Cobra cannot represent it |
| Repeated value containing comma | one occurrence, comma retained | `StringSlice` would split it | use `StringArray`; order and commas remain intact |
| Nested groups | help plus exit 1 without a leaf | configurable | group `RunE` renders help and returns a typed usage failure |
| `run` passthrough | intended but currently unusable | delimiter available | preserve the current rejection behavior for the first release; the probe only proves a future delimiter form is technically possible |
| Global log option | before or after a leaf | persistent flag supports both | parse before invocation; tokens after `--` remain child arguments |
| Internal updater | legacy marker can be found anywhere | hidden commands remain discoverable internally | route outside the public Cobra tree while preserving the legacy marker detection contract |
| TTY prompt | interactive flow; cancellation commonly succeeds | not a Cobra concern | prompt Adapter checks terminal facts and reports typed cancellation |
| Child status | `run` propagates it | not a Cobra concern | typed child outcome preserves the exact code |
| Signals | command-dependent legacy behavior | context can carry cancellation | composition root owns signals; final mapping remains a human choice |

The probe manually produced these outcomes: parser/action failure `1`, prompt cancellation `0`, child status `7`, SIGINT `130`, SIGTERM `143`, non-TTY prompt failure `1`, and help/version `0`. Raw-Cobra mode confirms the optional-value incompatibilities are real rather than assumptions.

With Go 1.26.7 and `CGO_ENABLED=0`, the same source cross-builds successfully for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, and `windows/arm64`.

## Interface designs

### 1. Direct Cobra registration

Each command constructor returns or mutates `*cobra.Command` and reads flags in `RunE`.

- Smallest initial implementation.
- Cobra types, flag state, hook ordering, output, and error conventions become part of every command Module's Interface.
- Cross-cutting fixes for logging, errors, cancellation, and compatibility lose Locality.
- Rejected unless minimizing first-day code is more important than the migration's technical-debt constraint.

### 2. Generic declarative command specification

Command Modules build a framework-neutral tree of `Group`, `Leaf`, `OptionalFlag`, `RepeatedFlag`, codecs, validation callbacks, and typed bindings; a compiler emits Cobra commands.

- Centralizes grammar validation and provides good leverage if commands are dynamically extended.
- Its public Interface is almost as complex as the parser problem it describes.
- Optional values and `run` passthrough require policy knobs and escape hatches, turning the Module into a second CLI framework.
- Rejected for ycy's fixed command tree.

### 3. Deep CLI App plus typed command manifest

The composition root supplies a fixed typed set of command handlers. The CLI App owns the tree, command-specific binders, normalization, global flags, diagnostics, and outcome classification.

```go
package cliapp

type Commands struct {
    GitCM func(context.Context, gitcm.Input) error
    Diff  func(context.Context, diff.Input) error
    FS    func(context.Context, fs.Input) error
    Run   func(context.Context, run.Input) (run.Result, error)
    // One explicit field per remaining leaf, grouped by domain where useful.
}

type Dependencies struct {
    Host     Host
    Commands Commands
    Updater  InternalUpdater
}

func New(BuildInfo, Dependencies) (*App, error)
func (*App) Execute(context.Context, []string) Outcome
```

Each command Module owns `Run(context.Context, Input) ...`, captures its own Git, HTTP, config, prompt, or process ports, and imports neither Cobra nor pflag. `Commands` is an honest compile-time composition manifest, not a generic registry or dependency bag.

Inside `internal/cliapp`, separate handwritten `bind_*.go` files retain imperative freedom for the few compatibility anomalies. Every `Execute` builds fresh parser state. A central invocation wrapper resolves logging, decodes typed input, invokes the handler, classifies the result, and emits a diagnostic exactly once. `main` is the only place that calls `os.Exit`.

This design has the greatest Depth for callers, the best Locality for compatibility fixes, and no hypothetical parser seam. Its cost is one explicit handler field and one hidden binder per leaf, which is appropriate for a fixed public CLI.

## Dependency placement

| Dependency | Category | Placement |
| --- | --- | --- |
| Cobra, pflag, validators, argv normalizers, outcome classifier | in-process | hidden inside `internal/cliapp`; tested through `Execute` |
| Streams, environment, cwd, terminal facts | local-substitutable | production Host Adapter and test Host Adapter |
| Prompts and child processes | local-substitutable | command-owned semantic ports with production and scripted Adapters |
| Tunnel transport | remote but owned | port owned by the tunnel Module, not the CLI Module |
| GitHub, model providers, external Git | true external | ports owned by their command Modules with production and test Adapters |

Do not pass a repository-wide `Invocation` or `Session` service locator into every command. Command constructors capture domain dependencies; the CLI passes only context and typed command input.

## Selected decision

The human selected the deep CLI App plus typed command manifest. Direct Cobra registration and the generic declarative command specification are rejected.

Only the deep CLI App architecture is selected for production. Typed signal, cancellation, failure, and child outcomes remain useful internally, but their numeric mapping is established command by command from legacy tests. The composition root is the only place that maps those typed outcomes to process exit codes; the prototype does not authorize a new global behavior policy.

## Remaining production gates

- Exhaustive argv tables and fuzzing for optional values, repeats, global placement, and delimiter grammar.
- Exact stdout/stderr snapshots for help, version, usage, and action failures.
- Real Huh behavior for PTY, redirected I/O, `TERM=dumb`, Escape, Ctrl-C, and Windows Console.
- Unix process groups and Windows Job Objects for child/grandchild cancellation and no-orphan guarantees.
- Legacy-compatible updater bootstrap and state handling, with a narrow compatibility exception only if a required target cannot reproduce them.
- Sequential repeat execution without stale pflag state; declare concurrent execution unsupported unless a real caller appears.
