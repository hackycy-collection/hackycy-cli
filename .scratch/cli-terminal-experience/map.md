# Chart the ycy terminal experience redesign

Label: wayfinder:map
Status: resolved

## Destination

Produce an execution-ready design and migration plan for a friendly, visually
coherent ycy terminal experience built around the existing Cobra command tree
and Charmbracelet tooling. The plan must improve every existing interactive
command and long-running command while preserving stable automation behavior,
machine-readable results, public command syntax, and exit semantics.

## Notes

- Domain: ycy command-line interaction, terminal capability negotiation,
  visual presentation, structured diagnostics, and command compatibility.
- Cobra remains the command-tree and argument-parsing Module in
  `internal/cliapp`. Charmbracelet may improve the terminal-facing Adapters;
  it must not leak into domain command Modules under `internal/commands`.
- The composition root currently wires `os.Stdin`, `os.Stdout`, and
  `os.Stderr` directly into terminal-specific Prompters and Presenters in
  `cmd/ycy`. The map must decide a deep replacement Module and its seam.
- An Interactive Session may use Huh and Lip Gloss by default. Bubble Tea is
  reserved for complex selection, live progress, or cancellable dynamic views.
- An Automation Session must use stable plain output: no animation, cursor
  control, alternate screen, or unsolicited prompt. Existing command names,
  flags, positional grammar, exit codes, and machine-consumed output stay
  compatible unless a later decision explicitly documents an exception.
- Command Results go to stdout. Diagnostic Events go to stderr. The desired
  public logging surface includes the current `--log-level`, a JSON log
  format, and deliberate quiet/verbose behavior; exact spellings and semantics
  remain a decision.
- English remains the terminal language for this effort. `NO_COLOR` and
  cross-platform terminal behavior are compatibility requirements, not visual
  polish details.
- Consult `domain-modeling` whenever CLI-experience terminology changes,
  `codebase-design` for Module interfaces and seams, `research` for external
  library/runtime facts, and `prototype` when a visual or interaction decision
  needs a concrete artifact.
- The existing Go migration map permits human-facing layout, color, wording,
  and diagnostics to change only when they are not machine-consumed. Its
  established output and standalone tests are evidence to preserve, not an
  invitation to change command behavior.

## Decisions so far

<!-- Closed child tickets are indexed here by name. -->

- [Research the Charmbracelet terminal runtime contract](issues/01-research-charmbracelet-terminal-runtime.md): pin Huh `v1.0.0`, Bubble Tea `v1.3.10`, and Lip Gloss `v1.1.0`; keep terminal-mode selection, plain automation output, and cancellation/logging policy owned by ycy.
- [Define the terminal capability and automation contract](issues/02-define-terminal-capability-and-automation-contract.md): classify sessions from all three standard streams and `CI`; use rich UI only for positively recognized terminals, conservatively degrade unknown terminals, and make automation fail explicitly rather than prompt or open a controlling TTY.
- [Inventory command experience and output contracts](issues/03-inventory-command-experience-and-output-contracts.md): ycy has distinct result, prompt, child-stream, browser-readiness, service-log, and update-output contracts; migrate terminal presentation command by command rather than moving all status output at once.
- [Prototype the terminal visual language](issues/04-prototype-the-terminal-visual-language.md): use the legacy Bun CLI as the Rich Interactive baseline for hierarchy, stage order, wording, title behavior, and semantic colors; use Charm-native control chrome rather than byte-for-byte Clack output, while keeping Automation plain.
- [Choose the terminal experience Module seam](issues/05-choose-the-terminal-experience-module-seam.md): use a shared deep `internal/terminal` Module with thin `cmd/ycy` semantic Adapters, an Experience Run Interface, and a lease-aware diagnostic stream; keep command behavior and Raw Child I/O outside its renderer.
- [Define the command-result and diagnostic contract](issues/06-define-output-and-diagnostic-contract.md): keep durable results and Raw Child I/O stream-compatible, add explicit text/NDJSON Diagnostic Configuration, preserve plain User-Actionable Errors, and apply recursive redaction with rich-only semantic log color.
- [Choose the long-running operation model](issues/07-choose-the-long-running-operation-model.md): use inline Bubble Tea only for Git Pulse, Git CM, and Git Fork; keep other journeys Huh/static/raw/service-oriented, with command-owned Operation Phases and cooperative cleanup.
- [Define help, error, and completion experience](issues/08-define-help-error-and-completion-experience.md): keep Cobra as the grammar and raw-completion owner, render static rich Help only through the terminal Module, and give one-line errors conservative recovery guidance.
- [Define automation fallback for prompted commands](issues/10-choose-automation-inputs-for-prompted-commands.md): preserve existing input grammar; let already-resolved paths proceed and fail Prompt-Dependent Paths before side effects in Automation Sessions.
- [Approve the CLI experience migration and acceptance plan](issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md): migrate via a shared terminal foundation and command-owned vertical slices, with `tunnel server` as an explicit continuous-log journey and a terminal-specific acceptance matrix.

## Not yet specified

- None. The approved wrap-or-stack behavior and the final width/resize PTY
  evidence leave component-level layout tuning to the implementation slices;
  it is not a pre-implementation product decision.

## Out of scope

- A new full-screen command center, dashboard, or replacement top-level
  command. This effort improves existing command journeys.
- Changing command names, flag spellings, positional grammar, business
  behavior, destructive-action semantics, or exit codes merely to support a
  prettier terminal UI.
- Adding automation-only flags, positional inputs, secret sources, implicit
  selections, or workflow shortcuts for commands that currently require a
  prompt. Those paths instead use the approved Automation Session failure
  behavior.
- Localizing the CLI or adding a language selector; English remains the
  supported terminal language for this effort.
- Redesigning the browser-based `diff`, `fs`, or Tunnel web applications.
- Telemetry, remote log shipping, and a general application-observability
  platform.
