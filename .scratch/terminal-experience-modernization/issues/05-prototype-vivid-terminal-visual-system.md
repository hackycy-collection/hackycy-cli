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
Run `make prototype-terminal`; use F2/F3 to compare Signal Rail, Ops Console,
and Focus Flow against the same interaction, F4 to select the success, failure,
or cancellation journey, and F5 to restart. Direct launch arguments are
documented with the prototype.

## Answer

Adopt variant A, **Signal Rail**, as the visual direction for the first
implementation. The user reviewed the runnable PTY prototype and selected A.
Variants B (Ops Console) and C (Focus Flow) remain only as comparison material
in the throwaway prototype; their layouts are not part of the production
specification.

The Signal Rail direction is a vivid but restrained dark-terminal treatment:

- Use a compact `YCY / <command>` eyebrow, one clear command title, and a
  muted one-line context description. Keep title and body typography separate;
  do not use hero-scale text inside command forms.
- Keep a persistent left workflow rail for finite interactive commands. It
  shows numbered semantic steps, a cyan active marker (`◆`), green completed
  markers (`✓`), muted pending markers (`○`), and explicit warning/error or
  cancellation symbols (`!`, `✕`, `⊘`). The rail is a progress explanation,
  not a percentage meter.
- Use high-contrast cyan as the primary action color and magenta as the
  secondary focus color. Green means completed/successful, yellow means
  warning/cancellation, red means failure, and muted gray is reserved for
  descriptions, pending work, and help. Keep color paired with symbols and
  labels so meaning survives color degradation.
- Render Huh v2 Input, Password, Select, MultiSelect, and Confirm controls with
  a colored left focus rule, visible descriptions, searchable long lists, and
  explicit selected/unselected markers. Password values are masked in the Live
  View and always become `[redacted]` in the transcript. Option descriptions
  remain visible and searchable but do not alter the submitted semantic value.
- Represent active finite work with one spinner beside the current Work Phase
  and a sentence-length detail. Never invent a percentage for unknown totals.
  Preserve completed phases in place and attach failure or cancellation to the
  phase where it occurred.
- After AltScreen closes, replay a compact semantic Interaction Transcript on
  stderr: completed answers (with secrets redacted), meaningful final Work
  Phase states, the cancellation/failure location, and one final outcome line.
  Do not replay keystrokes, filters, invalid secret input, animation frames,
  or a large Command Result.
- Emit a complete successful Command Result once on stdout. Failure and
  cancellation in this prototype intentionally emit no result, matching the
  separate-result ownership decision; production commands retain their
  existing result and exit contracts.
- Keep the layout responsive: use the rail on normal terminal sizes and a
  compact single-column form when width or height is constrained. Use the
  repository's explicit stream capability for color and normal Charm
  downsampling; do not add terminal-specific layout branches.

The prototype's observed PTY lifecycle validates that the chosen direction
restores the primary screen before transcript replay, keeps stdout separate,
and never exposes the synthetic API token. The throwaway source remains linked
above as the primary visual reference for implementation and later PTY
acceptance evidence.
