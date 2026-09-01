# Modernize every ycy command's terminal experience

Label: wayfinder:map
Status: resolved

## Destination

Produce an execution-ready terminal-experience specification for the root
surface and every ycy leaf command. The specification must define each
command's Live View or Lifecycle Log, Work Phases, Interaction Transcript,
Command Result, failure/cancellation presentation, latest Charm v2 migration,
rollout order, and acceptance evidence without changing existing CLI behavior.

## Notes

- Domain: Go CLI terminal interaction, finite-work progress, long-running
  service logs, durable output, diagnostics, redaction, and PTY evidence.
- Use the vocabulary in [`CONTEXT.md`](../../CONTEXT.md). Consult
  `domain-modeling` when those terms change, `research` for external library
  and upstream CLI facts, `prototype` for visual interaction decisions, and
  `codebase-design` when selecting shared terminal or logging interfaces.
- Preserve command names, arguments, flags, defaults, selection and
  confirmation semantics, cancellation meaning, exit codes, side effects,
  stdout machine output, JSON/NDJSON schemas, and sensitive-data redaction.
- Finite interactive work uses a vivid, high-contrast latest-Charm-v2 Live
  View in the alternate screen. Every real Work Phase is shown even when it
  completes quickly; do not invent percentages for unmeasurable work.
- Leaving the alternate screen replays a semantic Interaction Transcript on
  stderr: completed questions and redacted answers, meaningful Work Phase
  final states, cancellation/failure location, and final outcome. It never
  replays keystrokes, filters, invalid secret input, or animation frames.
- The complete Command Result is emitted once on stdout. Large results are not
  duplicated into the Interaction Transcript. Automation mode emits no extra
  transcript and remains free of terminal control sequences.
- `run` exits the Live View, replays its selection summary, and releases the
  terminal before the selected child process starts. It does not decorate or
  intercept the child process terminal session.
- `fs`, `diff`, `tunnel connect`, and `tunnel server` are Service Commands.
  They do not receive a custom full-screen UI; their user experience is a
  styled, line-oriented Lifecycle Log. Text logs may be redesigned while JSON
  logs retain their existing NDJSON schema.
- Use the latest mutually compatible stable Charm v2 modules, including an
  explicit decision on `charm.land/log/v2`; do not retain v1 merely to reduce
  migration effort.
- The first implementation may rely on the repository's current Rich,
  Plain Interactive, and Automation classification and normal terminal
  degradation. Expanding into a broad terminal-specific compatibility matrix
  is not a prerequisite.
- Every command owns its information structure and wording projection. Shared
  code is limited to proven visual primitives, state/transcript mechanics,
  capability handling, and logging facilities.
- This map plans the work. It does not implement the terminal redesign.

## Decisions so far

<!-- Closed child tickets are indexed here by name. -->

- [Research the latest compatible Charm v2 stack](issues/01-research-latest-charm-v2-stack.md): the compatible latest stable set is Bubble Tea 2.0.9, Bubbles 2.2.1, Huh 2.0.3, Lip Gloss 2.0.6, and Log 2.0.0; Log can only sit behind the existing redaction and stable-NDJSON facade.
- [Research exemplary open-source CLI terminal experiences](issues/02-research-exemplary-cli-terminal-experiences.md): use Dagger for live/final state separation, Gum and GitHub CLI for vivid forms, Pulumi for progress/failure ordering, and Docker Compose for Lifecycle Logs without copying their product-specific layouts.
- [Inventory every current command terminal surface](issues/03-inventory-every-command-terminal-surface.md): the root and all 22 leaves are catalogued by stream, mode, state, loss/duplication, and evidence; Rich has no durable transcript, only three Git leaves use typed phases, and Service Command logging is split between stdout results and Tunnel lifecycle logs.
- [Choose the Charm v2 and logging boundaries](issues/04-choose-charm-v2-and-logging-boundaries.md): isolate the pinned v2 stack behind `internal/terminal`, use embedded Huh forms and one Bubble Tea root, and keep Log v2 as an atomic text-only Adapter behind the existing redaction and stable-NDJSON logging Interface.
- [Prototype the vivid terminal visual system](issues/05-prototype-vivid-terminal-visual-system.md): adopt Signal Rail (A): persistent semantic workflow rail, vivid cyan/magenta focus colors with symbol-paired states, Huh v2 forms, phase-attached outcomes, and compact post-AltScreen transcript replay.
- [Choose the shared terminal Experience contract](issues/06-choose-shared-terminal-experience-contract.md): use a semantic Finish/Milestone seam with ordered phase catalogs, an append-only redacted transcript ledger, strict stream/lifecycle ordering, at-most-once cancellation, and command-owned wording/results; Service Commands keep Logger-backed Lifecycle Logs.
- [Choose the root, help, discovery, and error presentation](issues/07-choose-root-help-and-error-presentation.md): keep discovery as durable stdout without AltScreen, preserve version/completion and exit contracts, style only TTY output, enforce safe single-line root errors, propagate discovery write failures, and keep startup/diagnostic controls ordered and isolated.
- [Choose the export env terminal experience](issues/08-choose-export-env-presentation.md): use a safe Huh selection and fixed discovery/export phases, preserve heading plus pretty JSON, delay file-success output until after writing, replay only metadata and phase finals, and keep cancellation/Automation side-effect rules unchanged.
- [Choose the config fork list terminal experience](issues/09-choose-config-fork-list-presentation.md): show an always-visible finite loading phase, preserve ordered safe instance rows and encrypted previews, use responsive TTY tables without transcript duplication, and keep empty/read-failure/result contracts intact.
- [Choose the config fork add terminal experience](issues/10-choose-config-fork-add-presentation.md): keep the five-step Huh form and validation contract, add Signal Rail field/save phases, redact credentials and unsafe host details, preserve silent alias overwrite, and submit exactly one outcome result.
- [Choose the config fork remove terminal experience](issues/11-choose-config-fork-remove-presentation.md): preserve ordered first-default selection and default-No destructive confirmation while adding loading/removal phases, distinct redacted cancellation Transcript markers, safe labels, one-shot outcomes, and Automation no-side-effect boundaries.
- [Choose the config cm list terminal experience](issues/12-choose-config-cm-list-presentation.md): add an early loading phase spanning store/read, preserve ordered profile results and default semantics, use responsive Rich layouts with safe URL projections, and keep Plain/Automation result compatibility and failure boundaries.
- [Choose the config cm add terminal experience](issues/13-choose-config-cm-add-presentation.md): preserve the four-field Huh form and validation/overwrite semantics while adding Signal Rail collection and save phases, safe profile/URL projections, redacted cancellation transcripts, and one-shot success/failure outcomes.
- [Choose the config cm use terminal experience](issues/14-choose-config-cm-use-presentation.md): keep the parameter-only atomic default selection, represent lookup/persistence as one truthful phase, carry the safe target in phase detail, and preserve Plain/Automation results, missing-profile errors, and at-most-once mutation.
- [Choose the config cm set terminal experience](issues/15-choose-config-cm-set-presentation.md): preserve the three-argument atomic update and all legacy key parsers while adding one truthful phase, key-classified successful details, universal failed-value suppression, API-key redaction, and unchanged Plain/Automation results.
- [Choose the config cm remove terminal experience](issues/16-choose-config-cm-remove-presentation.md): preserve validation-first parameter deletion, default-No confirmation, and automatic default reassignment while adding separate validation/removal phases, safe destructive prompts, distinct cancellation Transcript markers, and Automation no-side-effect boundaries.
- [Choose the config cm test terminal experience](issues/17-choose-config-cm-test-presentation.md): preserve profile resolution, the fixed provider request, timeout/error semantics, and stdout documents while adding two truthful phases, safe provider projections, bounded Rich response/usage views, redacted Transcript replay, and Plain/Automation-compatible degradation.
- [Choose the diff Lifecycle Log presentation](issues/18-choose-diff-lifecycle-log-presentation.md): preserve the stdout startup document and foreground service semantics while adding bounded `diff` startup, Refresh, issue, failure, and shutdown Lifecycle records with stable levels, fields, text symbols, NDJSON compatibility, and exactly-once ordering.
- [Choose the fs Lifecycle Log presentation](issues/19-choose-fs-lifecycle-log-presentation.md): preserve startup/stopped stdout checkpoints and foreground service behavior while adding safe `fs` startup, capability, authentication, Managed Task, warning, failure, and shutdown Lifecycle records with bounded volume and stable text/NDJSON projections.
- [Choose the git heat terminal experience](issues/20-choose-git-heat-presentation.md): add three truthful Git inspection phases, a responsive vivid hot-path report, safe Rich path/query projections, compact Transcript replay, and phase-specific failure/cancellation while preserving every existing result and exit contract.
- [Choose the git pulse terminal experience](issues/21-choose-git-pulse-presentation.md): preserve scan/date/fetch/author semantics while adding four truthful phases, modern Huh forms, visible partial-discovery/fetch warnings, a responsive complete commit tree, semantic Transcript replay, and unchanged result/signal boundaries.
- [Choose the git fork terminal experience](issues/22-choose-git-fork-presentation.md): preserve archive-first acquisition, overwrite and fallback semantics while adding truthful phases, a vivid Huh confirmation, safe partial-disk accounting, semantic Transcript replay, and one-shot Rich/Plain/Automation results.
- [Choose the git cm terminal experience](issues/23-choose-git-cm-presentation.md): preserve the full Git CM flag matrix and mutation order while adding searchable file selection, complete message review, truthful generation/commit/push phases, partial-Git accounting, semantic Transcript replay, and exactly-once results.
- [Choose the rm terminal experience](issues/24-choose-rm-presentation.md): preserve explicit and smart cleanup safety, force behavior, concurrent partial deletion, and result semantics while adding destructive Huh forms, truthful phases, safe path projections, semantic Transcript replay, and control-free Automation boundaries.
- [Choose the run handoff terminal experience](issues/25-choose-run-handoff-presentation.md): preserve project/script/manager discovery and raw child IO/exit ownership while adding modern selection forms, truthful handoff phases, safe parent Transcript replay, and release-before-exec guarantees.
- [Choose the tunnel connect Lifecycle Log presentation](issues/26-choose-tunnel-connect-lifecycle-log-presentation.md): keep remembered-connection selection as the only bounded Rich interaction, then use a redacted line-oriented session log with stable event IDs, throttled reconnects, explicit FRP/reconciliation states, and exactly-once shutdown records.
- [Choose the tunnel server Lifecycle Log presentation](issues/27-choose-tunnel-server-lifecycle-log-presentation.md): separate control-plane availability from managed FRPS state, suppress default access logs, aggregate agent warnings, expose committed changes safely, and finish with exactly-once server failure/stopped records.
- [Choose the zip terminal experience](issues/28-choose-zip-presentation.md): preserve four-step archive planning, default glob behavior, filename sanitization, and archive semantics while adding Huh v2 forms, truthful collection/compression/publication phases, bounded Transcript replay, and exactly-once results.
- [Choose the upgrade terminal experience](issues/29-choose-upgrade-presentation.md): keep strict release resolution, candidate checksum/version verification, detached scheduling, startup transaction consumption, unusual exit codes, and stdout compatibility while adding truthful phases and redacted Rich Transcript replay.
- [Approve the terminal experience rollout and acceptance plan](issues/30-approve-rollout-and-acceptance-plan.md): implement the shared Charm v2 foundation first, deliver command-sized risk-ordered batches without runtime dual stacks, and require layered semantic/stream/PTY/visual/cross-platform gates before release.

## Not yet specified

- None. The open child tickets cover the currently visible route; new command
  states discovered by the inventory may graduate into additional tickets.

## Out of scope

- Implementing or shipping the redesigned terminal experience while this map
  is being resolved.
- Changing command behavior, parser contracts, business rules, external
  effects, result schemas, exit semantics, or secrets policy.
- Building custom full-screen dashboards for Service Commands or rendering
  high-volume request, heartbeat, or polling activity at normal log levels.
- Redesigning the browser applications served by `fs`, `diff`, or Tunnel.
- Guaranteeing bespoke rendering for every terminal emulator, shell, screen
  reader, or legacy console in the first implementation.
