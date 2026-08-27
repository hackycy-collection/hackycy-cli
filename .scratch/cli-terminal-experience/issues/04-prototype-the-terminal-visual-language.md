# Prototype the terminal visual language

Type: prototype
Status: resolved
Blocked by: 01, 02

## Question

What should a cohesive ycy Interactive Session look and feel like in a normal
terminal? Produce a deliberately cheap but concrete prototype covering a Huh
selection, a confirmation, secret entry, a long-running progress/cancellation
view, a success result, a warning, an error, and a compact report. Use it to
decide hierarchy, spacing, color roles, status vocabulary, width behavior,
Unicode use, and how much branding is appropriate without making a developer
tool slow or noisy.

The prototype must include an Automation Session counterpart for each relevant
state so the human-facing design does not conceal a script-breaking behavior.
The resolution records the user-approved visual grammar and links the prototype
asset.

## Comments

- 2026-08-26: The user selected the legacy Bun output as the visual baseline,
  with GitHub CLI as a general ergonomics reference and visibly semantic
  colors. A non-final feasibility assessment is recorded in
  [`Bun Baseline Feasibility`](../prototype/bun-baseline/README.md). It finds
  the baseline implementable through the existing Go Prompter/Presenter
  interfaces; `git cm` needs explicit phase events, and the logger needs a
  small formatter correction. Huh can reproduce the interaction and visual
  hierarchy, but its stock prompt chrome is not byte-for-byte Clack output.

## Answer

Approved on 2026-08-26. The Rich Interactive Session uses the legacy Bun CLI
as its visual and behavioral baseline. The parity target is not byte-for-byte
Clack output: ycy preserves the information architecture, wording, ordering,
and semantic visual roles, while Huh and Lip Gloss may render their native
control chrome.

### Rich Interactive grammar

- Commands remain command-local; ycy gains no dashboard, persistent shell, or
  generic startup screen.
- Commands that used the Bun `printTitle()` convention retain its compact
  clear-screen `HACKYCY CLI` title in a Rich Interactive Session only. It is
  not emitted to Plain Interactive or Automation Sessions.
- Cyan identifies the title, active question, and current focus. Green records
  a completed stage, generated commit message, or successful outro. Yellow
  marks a warning or reduced-evidence note. Red marks errors and cancellation.
  Gray carries metadata, paths, guide text, timestamps, and tree connectors.
  Git Pulse retains its readable repository/date/author/message distinction.
- Spacing follows the Bun journeys: an intro before an interactive workflow,
  transient work in place, then a durable report or outro in scrollback.
  Results wrap or stack on narrow terminals rather than truncate essential
  text. Functional Unicode may have an ASCII-accessible representation.
- Huh owns ordinary selection, confirmation, and text/secret forms. Bubble Tea
  remains reserved for a later-approved dynamic operation; it is not needed to
  imitate Clack's default checkbox, radio, guide-bar, or spinner glyphs.

### Compatibility boundary

Automation Sessions keep their approved plain, newline-delimited behavior:
no title clear, Huh form, Bubble Tea program, animation, ANSI styling, or
implicit terminal access. `NO_COLOR` removes color while retaining Rich
Interactive layout and safe transient controls.

The assessment asset is
[`Bun Baseline Feasibility`](../prototype/bun-baseline/README.md). It maps the
four representative journeys to their existing Go seams and identifies only
two implementation gaps: explicit phase events for `git cm`, and a closer
human-readable logger formatter for Tunnel.
