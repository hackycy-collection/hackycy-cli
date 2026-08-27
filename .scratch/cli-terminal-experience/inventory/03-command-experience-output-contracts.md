# ycy Command Experience and Output Inventory

Date: 2026-08-26

This inventory records the observable terminal behavior of the active Go CLI.
It is a compatibility baseline for the CLI-experience map, not a proposal to
preserve every visual detail. `legacy/bun/` is outside this inventory except
where an existing active-Go test already makes a behavior observable.

## Evidence Basis

- [`cmd/ycy/main.go`](../../../cmd/ycy/main.go) is the composition root. It
  passes `os.Stdin` and `os.Stdout` into nearly every command adapter; only
  `run` receives stderr separately, `upgrade` receives both output streams,
  and the logging Runtime writes to stderr.
- [`internal/cliapp/app.go`](../../../internal/cliapp/app.go) owns Cobra,
  public argument grammar, help, global `--log-level`, error reporting, and
  exit-code conversion. Cobra help and version output use stdout; ordinary
  errors use stderr and exit code `1`.
- Current terminal adapters live in `cmd/ycy/`. They are mostly handwritten
  `bufio.Reader` loops that write prompts, validation, progress, and results to
  the same output writer. They are constructed even when streams are
  redirected, except for the remembered Tunnel selector.
- [`internal/logging/logging.go`](../../../internal/logging/logging.go) emits
  redacted timestamped text records to stderr. In practice it is materially
  used by Tunnel; it is not the common presenter for command status.
- The command and standalone tests named below are the current output
  compatibility evidence. A `CombinedOutput` test proves visible text and exit
  behavior but does not prove stdout/stderr ownership.

## Cross-Cutting Findings

### Stream topology today

| Surface | Current stream behavior | Compatibility implication |
| --- | --- | --- |
| Cobra help, completion, and version | stdout | Completion is script data. Help/version remain stable plain Command Results in an Automation Session. |
| Cobra errors and panic diagnostics | stderr, except panic handling writes one blank line to stdout first | Normal errors already fit the future Diagnostic Event direction. The blank-line preflight/panic behavior is tested and needs an explicit migration decision. |
| Most command Prompters and Presenters | stdout | Prompt, progress, status, JSON, reports, and success text are currently mixed. They cannot be globally moved to stderr without per-command result analysis. |
| `run` child process | child stdout to stdout; child stderr to stderr | This split is explicitly tested and must not be wrapped in a transient renderer. |
| `diff` and `fs` startup | stdout, then foreground wait | Integration tests consume stdout incrementally to find browser URLs. |
| Tunnel lifecycle | structured text logs on stderr | Tunnel is the existing model closest to Diagnostic Events, but its format is not JSON and its color decision is separate from other presenters. |
| `upgrade` | success/current status to stdout; classified failure writes an error to stderr and `Update aborted.` to stdout | Several error classes deliberately exit `0`; this is a public exception, not presentation noise. |

### Terminal and cancellation gaps

- The current `terminal(*os.File)` check only tests a character device. It does
  not apply the newly approved all-stream, `CI`, and conservative terminal
  capability contract.
- Configuration secret fields use `term.ReadPassword` only when stdin is a
  terminal, then silently fall back to ordinary line input. That fallback is
  incompatible with the approved Automation Session secret policy.
- Most prompt cancellation (`q`, `quit`, `cancel`, EOF, or an empty answer in
  command-specific cases) becomes a successful command result with a printed
  cancellation message. Signal cancellation flows through the root context;
  `diff`, `fs`, and Tunnel use it to close foreground servers cleanly.
- Existing output has no pseudo-terminal coverage. The only explicit color
  evidence is `NO_COLOR` for Git Heat and Git Pulse.

## Public Command Inventory

### Global and command discovery

| Command | Current adapter and output | Interaction and cancellation | Existing evidence | Likely experience archetype |
| --- | --- | --- | --- | --- |
| `ycy`, `--help`, command-group help | Cobra writes usage/help to stdout. Invoking bare `ycy` prints help but returns `1`; explicit help returns `0`. | No prompt. | `internal/cliapp/app_test.go` and each command binder test assert command visibility, help text, and error routing. | Plain result; later help styling must preserve Automation Session text. |
| `--version` / `-V` | One version line on stdout; it bypasses startup update-result output. | No prompt. | `internal/cliapp/app_test.go`, standalone binary tests. | Plain result; machine-safe. |
| `completion` | Cobra-generated shell script on stdout. | No prompt. | Cobra behavior is inherited; no ycy presenter. | Plain result only; never style or prefix it. |
| Unknown command, invalid flags, invalid `--log-level` | `cliapp` normalizes selected Cobra errors and writes `error: ...` to stderr. | No prompt; exit `1`. | `internal/cliapp/app_test.go` and per-leaf binder tests. | Plain actionable error; later error work must retain syntax/exit behavior. |

### Configuration and export

| Command | Current adapter and output | Interaction and cancellation | Existing evidence | Likely experience archetype |
| --- | --- | --- | --- | --- |
| `config fork list` | `internal/commands/config/fork.Render` writes the list directly to stdout. | No prompt or mutation. | `cmd/ycy/configfork_test.go`, `internal/cliapp/configfork_test.go`. | Static styled report in Rich Interactive; plain result otherwise. |
| `config fork add` | Handwritten text, provider-choice, and password prompts plus success/cancel text all write stdout. Password falls back to visible line input when stdin is not a terminal. | Multi-field create flow; validation loops; cancellation returns success without mutation. | `cmd/ycy/configforkadd_test.go` checks prompts, selection errors, secret non-disclosure, cancel, and success. | Huh form, including secure secret input only in an Interactive Session. |
| `config fork remove` | Select and confirmation prompts plus status/outcome all write stdout. | Select an instance, then destructive confirmation; decline/cancel returns success. | `cmd/ycy/configforkremove_test.go`. | Huh selection followed by Huh confirmation. |
| `config cm list` | `internal/commands/config/cm.Render` writes profile text directly to stdout. | No prompt or mutation. | `cmd/ycy/configcm_test.go`, `internal/cliapp/configcm_test.go`. | Static styled report in Rich Interactive; plain result otherwise. |
| `config cm add` | Handwritten text and password prompts plus validation/success/cancel write stdout. Password has the same non-terminal visible-input fallback as Fork add. | Multi-field create flow; cancellation returns success without mutation. | `cmd/ycy/configcmadd_test.go` and standalone config tests prove prompts, redaction, cancel, and success. | Huh form, including secure secret input only in an Interactive Session. |
| `config cm use` and `config cm set` | Positional arguments select the profile/change; success text writes stdout. | No prompt. Errors go through Cobra stderr. | `cmd/ycy/configcmuse_test.go`, `cmd/ycy/configcmset_test.go`, binder tests. | Plain result or compact static success notice. |
| `config cm remove` | A handwritten `[y/N]` confirmation and outcome write stdout. | Destructive confirmation; decline/cancel returns success. | `cmd/ycy/configcmremove_test.go`, `internal/cliapp/configcmremove_test.go`. | Huh confirmation. |
| `config cm test [profile]` | Provider response and profile diagnostic are printed to stdout; a failing provider test prints safe profile data before returning a redacted error to Cobra stderr. | No prompt; network request; no special cancellation UI. | `cmd/ycy/configcmtest_test.go` locks the response/diagnostic layout and redaction. | Static styled result with a short non-transient wait indicator at most. |
| `export env [dir]` | Without `--env`, a handwritten environment selector writes prompts to stdout. Success without `--out` writes `Exported variables:` then JSON to stdout; `--out` writes a status line and writes JSON to a file. | Selector cancellation is successful. `--env` avoids selection. | `cmd/ycy/exportenv_test.go`, module tests, binder tests. | Huh single-select only for Interactive Session; JSON result remains plain/machine-safe. |

### Local commands

| Command | Current adapter and output | Interaction and cancellation | Existing evidence | Likely experience archetype |
| --- | --- | --- | --- | --- |
| `rm [paths...]` | A handwritten confirmation or smart-cleanup action/target selector writes the plan, progress labels, notices, and final outcome to stdout. `--force` bypasses prompts. | Explicit deletion asks `[y/N]`; smart mode chooses action and targets. Prompt cancellation is successful and avoids deletion. | `cmd/ycy/rm_test.go` has an exact presenter snapshot and standalone destructive-flow checks; `internal/cliapp/rm_test.go` locks parsing. | Huh confirmation/selection with static status. No existing granular progress evidence justifies a Bubble Tea view by itself. |
| `run [path]` | Script/package-manager selection and informational intro write stdout. The selected external child inherits stdin and streams stdout/stderr separately. | Selector cancellation is successful. Child exit codes are propagated. | `cmd/ycy/run_test.go` and `cmd/ycy/run_process_test.go` explicitly assert child stdout/stderr separation and signal/exit behavior. | Huh selection before execution; then ordinary child-process passthrough, never an alternate-screen view. |
| `zip [directory]` | Handwritten package/source/glob/output-name wizard and plan/progress/outro write stdout. | Multi-step selection; cancellation is successful; archive creation/opening occurs after planning. | `cmd/ycy/zip_test.go` has prompt and exact presenter output coverage plus standalone cancellation/success checks. | Huh multi-step form with static status. |

### Git commands

| Command | Current adapter and output | Interaction and cancellation | Existing evidence | Likely experience archetype |
| --- | --- | --- | --- | --- |
| `git heat` | A report renderer writes stdout. It is the only presenter with an injected color boolean; `NO_COLOR` is honored here. | No prompt. | `cmd/ycy/githeat_test.go` locks an exact tabular report; `cmd/ycy/githeat_integration_test.go` proves `NO_COLOR` contains no ANSI. | Static styled report in Rich Interactive; preserve plain report for Automation Session. |
| `git pulse [directory]` | Handwritten day/author selectors, repository-scan and fetch progress, and tree report all write stdout. | Prompt cancellation is successful; root context can cancel scanning/fetching. | `cmd/ycy/gitpulse_test.go` locks selectors and semantic progress/report; integration tests prove `NO_COLOR` and missing-Git behavior. | Huh selections plus a Bubble Tea candidate for live scan/fetch progress. |
| `git fork <repo> [dest]` | Handwritten overwrite confirmation and a sequence of fetch/archive/clone milestones write stdout. External Git output is captured by its adapter rather than streamed directly. | Decline/cancel is successful; root context reaches archive/clone work. | `cmd/ycy/gitfork_test.go` locks milestones, cancellation, redaction, and process behavior; standalone tests cover completion. | Huh confirmation plus a Bubble Tea candidate for multi-stage progress. |
| `git cm` | Handwritten multi-file staging selector and commit confirmation write stdout. Generated commit text, provider/token/evidence metadata, failures, and final outcomes also write stdout. | Stage selection and confirmation cancellation are successful; provider/Git failures return errors. | `cmd/ycy/gitcm_test.go` and integration tests protect prompt/output, cancellation, mutations, and secret redaction. | Huh multi-select and confirmation plus a static styled generated-message report. |

### Browser, Tunnel, and update lifecycles

| Command | Current adapter and output | Interaction and cancellation | Existing evidence | Likely experience archetype |
| --- | --- | --- | --- | --- |
| `diff <baseline> <target>` | A presenter prints local/network browser and MCP URLs plus directories to stdout, then the command waits for process context/server completion. | Signal cancellation closes the server and returns normally. | `cmd/ycy/diff_test.go` locks startup text; `cmd/ycy/diff_integration_test.go` consumes stdout line-by-line to discover readiness and checks signals/errors. | Static startup/lifecycle result. Keep URLs visible in ordinary scrollback; do not hide them behind an alternate screen. |
| `fs [directory]` | A presenter prints browser URLs and server configuration to stdout, then prints stopped status after lifecycle completion. | Signal cancellation closes the server and returns normally. | `cmd/ycy/fs_integration_test.go` consumes stdout incrementally for startup and checks signal behavior; binder tests lock flags/errors. | Static startup/lifecycle result. Keep URLs visible in ordinary scrollback. |
| `tunnel server` | No command presenter. Server, FRP, and failure events flow through the scoped logging Runtime to stderr. | Foreground server shuts down through root context. | `cmd/ycy/tunnel_integration_test.go`, `internal/cliapp/tunnel_test.go`, and Tunnel runtime tests cover configuration, lifecycle, and logs. | Plain structured diagnostics; future work may style only Rich Interactive diagnostics without obstructing service logs. |
| `tunnel connect` | If remembered connections require a choice and stdin/stdout are character devices, a handwritten selector writes stdout and masks tokens. Client lifecycle uses stderr scoped logging. | Selector cancellation is successful; client exits on context cancellation or connection failure. | `cmd/ycy/tunnel_test.go`, `cmd/ycy/tunnel_integration_test.go`, and client runtime tests. | Huh selection in an Interactive Session, then plain structured diagnostics. Existing two-stream TTY gate must be replaced by the approved three-stream session contract. |
| `upgrade` | Success/current/update-state messages write stdout. Classified failed upgrades write a detailed error to stderr and `Update aborted.` to stdout; selected HTTP/integrity failure paths intentionally return exit `0`. | No prompt. The updater is detached; startup may present a prior transaction result. | `cmd/ycy/upgrade_test.go`, `internal/cliapp/upgrade_test.go`, and updater integration tests. | Plain result and diagnostics only; preserve status/output/exit exceptions exactly. |

## Test and Migration Constraints

1. Preserve all Cobra command names, flags, positional grammar, help behavior,
   and exit handling. UI replacement begins below the `internal/cliapp` command
   tree rather than replacing Cobra.
2. Introduce session classification before command-module construction. Today
   every input-dependent adapter except Tunnel is created with raw stdin; the
   configuration secret fallback is the highest-priority contract violation.
3. Split Command Results from Diagnostic Events one command family at a time.
   The existing single-writer presenters make a global stdout-to-stderr move
   unsafe, especially for `export env`, Git reports, browser URLs, and
   Upgrade's deliberate mixed output.
4. Retain direct child process streams for `run`. Preserve stdout readiness
   lines and foreground signal paths for `diff` and `fs`. Preserve stderr log
   streaming for Tunnel.
5. Replace presenter snapshots with tests at the new terminal-experience seam
   only after their observable Automation Session text, redaction, exit, and
   mutation contracts are represented elsewhere. Do not discard current exact
   output tests prematurely.
6. Add pseudo-terminal coverage for Rich and Plain Interactive Sessions,
   redirected-stream and `CI` Automation Sessions, `NO_COLOR`, secret echo,
   cancellation cleanup, and no-control-sequence assertions. Current tests do
   not cover this matrix.

## Newly Specified Decision

The inventory exposes a decision not visible when the map was charted:
several existing public commands have no explicit non-interactive input grammar
because they were built around a raw terminal prompt. The approved Automation
Session contract requires a command-by-command decision about whether each
command gains a narrow explicit input surface or fails before side effects.
That question now lives in [Choose automation inputs for prompted
commands](../issues/10-choose-automation-inputs-for-prompted-commands.md).
