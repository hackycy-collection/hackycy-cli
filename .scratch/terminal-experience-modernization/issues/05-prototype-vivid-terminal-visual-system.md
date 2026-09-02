# Prototype the vivid terminal visual system

Type: prototype
Status: resolved
Blocked by: 01, 02, 04

## Question

What concrete high-contrast Charm v2 theme, typography hierarchy, status
symbols, spacing, form controls, phase display, error state, and post-exit
Interaction Transcript should ycy use for text, secret, select, multiselect,
confirm, loading, success, cancellation, and failure states, as judged through
a runnable PTY prototype by the user?

## Prototype

The throwaway Charm v2 PTY prototype is linked at
[`internal/terminal/prototype-vivid`](../../../internal/terminal/prototype-vivid/README.md).
Run `make prototype-terminal`; use F2/F3 to compare historical A / Signal Rail,
current B / Ops Console, and C / Focus Flow against the same interaction, F4 to
select the success, failure, and cancellation journey, and F5 to restart.
Direct launch arguments are documented with the prototype.

## Answer

Adopt variant B, **OPS CONSOLE**, as the current Rich visual contract for the
first implementation. Variant A (Signal Rail) is superseded and variant C
(Focus Flow) remains available only as comparison material in the throwaway
prototype. This decision changes presentation hierarchy only; command
arguments, results, exits, interaction semantics, redaction, and the shared
Experience API remain unchanged.

The Ops Console direction is a dense, operator-oriented dark-terminal
treatment for finite Rich commands:

- Start with a compact top command/status bar. It identifies the command,
  safe target/profile context, and current outcome or lifecycle status in one
  stable line; it is not a hero heading and does not contain secrets.
- Follow with a metadata row for safe command context such as workspace,
  provider, scope, or selected range. Metadata is bounded, single-line, and
  control-free; command adapters own which fields are meaningful.
- Render an aligned status table with the columns `STATE`, `PHASE`, and
  `DETAIL`. Keep every reached form step or Work Phase in catalog order, use
  fixed column widths on wide terminals, and attach the active/final detail to
  the corresponding row. Do not invent percentages for unknown totals.
- Place the active Huh form or final result in the content region below the
  table. The form/result area is the only region that changes interaction
  focus; completed status rows remain stable above it. Large result documents
  still belong to stdout and are not copied into the view or Transcript.
- Use amber (`#FFB454`) as the primary action/status color and cyan
  (`#4CC9F0`) as the accent. Green means completed/successful, yellow means
  warning/cancellation, red means failure, and muted gray is reserved for
  descriptions and pending work. Pair every color with a state symbol and
  label: active `◆`, completed `✓`, pending `○`, warning/cancellation `!` or
  `⊘`, and failure `✕`.
- Huh v2 Input, Password, Select, MultiSelect, and Confirm controls retain
  visible descriptions, searchable long lists, and explicit selected/
  unselected markers. B uses a bottom focus rule for the active form rather
  than a persistent left focus rail. Password values stay masked in Live View
  and become `[redacted]` in the Transcript; option descriptions do not alter
  submitted semantic values.
- Active finite work shows one spinner beside the current phase and a
  sentence-length detail. Completed phases stay in the table, and failure or
  cancellation is attached to the phase where it occurred.
- After AltScreen closes, replay a compact semantic Interaction Transcript on
  stderr: completed answers (with secrets redacted), meaningful final Work
  Phase states, the cancellation/failure location, and one final outcome line.
  Do not replay keystrokes, filters, invalid secret input, animation frames,
  or a large Command Result.
- Keep no persistent left-side rail. On narrow or short terminals, collapse
  the table and active region into a legible single-column layout while
  preserving state labels, symbols, phase order, and safe details. Use the
  repository's capability policy and normal Charm downsampling; do not add a
  terminal-specific compatibility matrix.
- Service Commands (`diff`, `fs`, `tunnel connect`, and `tunnel server`) stay
  line-oriented Lifecycle Logs and do not use this full-screen Console.

### Amendment / decision history

- 2026-09-02: variant B / `OPS CONSOLE` supersedes variant A / Signal Rail as
  the current production Rich visual contract. A's persistent rail and cyan/
  magenta focus treatment are historical comparison material only; C / Focus
  Flow also remains comparison material.
- The existing A PTY acceptance entries in `goal-runbook.md` are retained as
  historical evidence and are not rewritten as B evidence. No new B PTY
  acceptance is claimed by this decision; B visual re-acceptance is a later
  runbook gate before command-slice work proceeds.
