# Prove the Go CLI compatibility approach

Type: prototype
Status: resolved
Blocked by: 01, 04, 05, 09

## Question

Which thin Go CLI composition can reproduce the difficult parts of the current command interface without leaking framework details through every command module? Prototype and compare the selected candidates against optional flag values such as `--push [remote]`, repeatable options, nested commands, `run` passthrough arguments, parser errors and exit codes, hidden updater invocation, global log configuration, TTY prompts, cancellation, and signal propagation. Present the interface and observed gaps for human selection; the prototype is disposable.

## Comments

- 2026-08-22, scope correction: retain the selected CLI App architecture, but treat the prototype's corrected `run` grammar, exact-first-token updater route, Git safety dispositions, and proposed global exit mappings as capability demonstrations rather than first-release product decisions. Legacy argv and outcome behavior remains the parity baseline.

## Answer

The selected Interface is the deep CLI App plus a fixed, strongly typed command manifest documented by the [Go CLI compatibility prototype results](../prototypes/go-cli-compat/RESULTS.md). `cmd/ycy` constructs command Modules with their domain dependencies, supplies their typed handlers to `internal/cliapp`, calls `Execute(context.Context, []string) Outcome`, and is the only code that maps an outcome through `os.Exit`. Command Modules receive context and typed input; they do not import Cobra/pflag, inspect framework flag state, call `os.Exit`, or receive a repository-wide `Invocation`/`Session` dependency bag.

`internal/cliapp` owns Cobra as an implementation dependency, a fresh command tree per execution, handwritten command-specific binders, global log parsing, diagnostics, and outcome classification. This keeps Commander compatibility local: command-scoped normalization can reproduce optional-value spellings, repeatables use `StringArray` so commas and occurrence order survive, and updater routing can live outside the public Cobra tree. The first release must bind these mechanisms to the legacy inventories: `run` retains its currently observable argv acceptance and rejection, the hidden updater retains its legacy detection behavior, and Git option oddities are reproduced rather than corrected. A focused parity failure may reopen only the affected binder as a compatibility exception.

Exit handling uses typed outcomes rather than error strings, and the composition root remains the only code that calls `os.Exit`. The prototype's outcome categories are available implementation tools, not a new global exit policy: each command maps cancellation, validation, action failure, child exit, graceful stop, and interruption according to its legacy contract tests. The foreground child status remains directly representable. Any command whose legacy signal or status behavior cannot be reproduced must surface that exact mismatch rather than inherit a convenient global default.

The disposable probe demonstrated nested groups, optional and repeatable flags, delimiter capability, global flag placement, an updater route outside Cobra, TTY/non-TTY prompt behavior, cancellation, child exit propagation, and SIGINT/SIGTERM propagation. With Go 1.26.7 and `CGO_ENABLED=0`, it also cross-builds for all six required darwin/linux/windows amd64/arm64 targets. Production work still requires legacy-derived argv tables, real prompt and Windows Console tests, process/signal parity, and updater invocation parity. These are command migration tests, not permission to redesign behavior.
