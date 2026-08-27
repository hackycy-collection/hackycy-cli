# Choose the long-running operation model

Type: grilling
Status: resolved
Blocked by: 03, 04, 05

## Question

Which existing ycy journeys warrant a Bubble Tea dynamic view, and what
behavior does each view own? Use the completed command inventory to distinguish
ordinary styled reports from operations with meaningful live progress,
selection, retry, cancellation, or cleanup. Decide how Bubble Tea interacts
with streaming child-process output, browser-server startup URLs, signal-driven
shutdown, operation completion, and error propagation without hiding durable
results or changing exit semantics.

The answer must give a command-by-command classification and a uniform
cancellation/cleanup contract for transient views.

## Answer

Approved on 2026-08-26. Bubble Tea is limited to three Rich Interactive
journeys with meaningful, command-owned live state: `git pulse`, `git cm`,
and `git fork`. All views render inline on stderr under a Renderer Lease;
none uses the alternate screen. They leave a compact final status in
scrollback, release the lease, and then allow durable Command Results to
render on stdout.

### Command classification

| Journey | Approved experience |
| --- | --- |
| `git pulse` | Huh selections around two tracked segments: scan repositories, choose a period, fetch commits, choose authors, then present the durable report. |
| `git cm` | Huh file selection, a tracked stage/generate segment, Huh commit confirmation, a tracked commit/push segment, then present the durable result. |
| `git fork` | Huh overwrite confirmation, then one tracked resolve/archive/clone segment followed by the durable destination result. |
| `config fork add/remove`, `config cm add/remove`, `export env`, `rm`, `zip`, `run` selection, and `tunnel connect` selection | Huh forms or selections plus static outcomes; they do not justify a Bubble Tea view. |
| `config fork list`, `config cm list/use/set/test`, `git heat`, and `upgrade` | Static Lip Gloss reports or plain Command Results. |
| `diff` and `fs` | Durable startup URLs and lifecycle text remain visible on stdout, outside a transient view. |
| `run` child process | Raw Child I/O retains inherited stdin, stdout, and stderr without a transient renderer. |
| `tunnel server` and connected Tunnel lifecycle | Continuous stderr diagnostics remain a service-log stream and never coexist with a Renderer Lease. |

Plain Interactive and Automation Sessions use their approved plain behavior;
they never construct Huh or Bubble Tea merely to imitate rich output.

### View behavior and phases

A tracked view exists only for one contiguous work segment. An Experience Run
must release its Renderer Lease before a Huh form appears, then may acquire a
new lease for the next segment. This preserves the existing command sequence:

- Pulse shows scan count/current repository, then fetch progress, before its
  report.
- CM shows explicit collect/stage, generate, commit, and optional push
  phases. Its confirmation remains between generate and commit.
- Fork shows resolve, default-branch/archive work, clone, and ready phases.

Each view is a semantic phase ledger: completed phases, one visibly colored
current phase, and one concise contextual detail. It must not invent a
percentage, ETA, raw Git stream, or inferred phase. Narrow terminals collapse
to the current phase and its context rather than truncate essential
information.

Command Modules define the Operation Phases through typed semantic callbacks.
Thin Terminal Adapters in `cmd/ycy` translate those callbacks into tracked
state. `internal/terminal` renders that state and requests cancellation, but
does not run business work, retry operations, parse logs, or infer progress.
The command root continues to own orchestration, result/error propagation,
and OS signals.

### Cancellation, cleanup, and outcomes

Bubble Tea runs with its signal handler disabled. During raw terminal input,
`Ctrl+C` is forwarded as an immediate cooperative cancellation request to the
`cmd/ycy`-owned controller; `Esc` asks for cancellation confirmation. The
view changes to a cancelling state and remains active until the domain work
acknowledges cancellation and completes its cleanup. It then releases the
Renderer Lease and restores cursor/input state before Cobra receives the
original outcome.

Completion, cancellation, and failure each produce an explicit final phase
state. Successful work uses a green completion state before its durable
result. Cancellation uses a red `Cancelled` state and does not print a
misleading success result. Failure uses a red failed-phase state, then
preserves the existing plain `error:` line and exit behavior. A partial
outcome, such as a created commit followed by a failed push, states both facts
without claiming full success.

No generic retry control, hidden retry, or Bubble Tea-owned error recovery is
introduced. Rerun behavior stays command-owned. Deferred diagnostic records
may render only after the tracked view has cleaned up, so they cannot corrupt
the inline renderer.
