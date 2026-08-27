# GitHub CLI-Inspired Journey Review

This second throwaway prototype answers the visual-language ticket using the
user-selected GitHub CLI reference and four real ycy command shapes. It does
not propose a generic terminal shell, new command syntax, or Automation Input
flag grammar.

## Shared Grammar

- No persistent banner, dashboard, card frame, or alternate-screen ownership
  for ordinary commands. A command begins where its work begins.
- A cyan `?` identifies a decision; cyan is otherwise reserved for current
  focus and links/paths. Green `✓` records completed work. Amber is a warning;
  red is an error. Gray carries stable metadata and secondary context.
- Human-facing results use short verbs rather than a global `OK`/`WARN`/`ERROR`
  prefix. Long-running service diagnostics retain level labels because they
  are operational records, not command results.
- Functional Unicode is deliberately small: `?`, `>`, `✓`, and the spinner
  glyph. Each has enough surrounding wording to remain understandable with
  `NO_COLOR`; no symbol is decoration.
- A 60-column layout stacks local detail below the value it qualifies instead
  of clipping it. Wide reports gain alignment and, for Git Pulse, a compact
  weekly activity graph.

## Command-Specific Surfaces

| Journey | Rich Interactive proposal | Durable result or log proposal |
| --- | --- | --- |
| `git cm` | Huh multi-select for files, then a compact default-yes confirmation after generation. | The commit message is the visual focus; profile/context metadata is secondary; load and completion are terse verb lines. |
| `git pulse` | Huh date and author selection. A Bubble Tea transient view is a candidate only for live scan/fetch updates. | A small activity graph plus a compact repository table leads the result; the existing detailed commit report remains below it. |
| `tunnel connect` | Huh select of already-masked saved connections. | One success/result line and endpoint details; `ctrl+c` remains the familiar lifecycle control. |
| `tunnel server` | No form and no transient renderer once running. | Stderr-only, redacted, line-oriented logs with timestamp, colored severity, muted scope, message, and context. |
| `config cm add` | Sequential Huh form with echo-disabled API-key input. | One saved-profile result plus non-secret provider/model metadata. |

## Automation Counterpart

`--journey automation` demonstrates the required plain stream shape: no
prompt, form, spinner, alternate screen, or implicit stdin secret. It shows
failure-before-effects for missing decisions and ordinary progress/result/log
lines for commands that can proceed. The exact per-command Automation Input
surface remains pending in `Choose automation inputs for prompted commands`.

## Proposed Decision

Adopt this command-local grammar as the visual base. Apply a Bubble Tea
Transient View only where live work benefits from replacement-in-place
progress, initially Git Pulse and the generation phase of Git CM. Keep Tunnel
service output outside that renderer. This is intentionally a visual decision,
not approval to change current output contracts before the output/diagnostic
ticket resolves.
