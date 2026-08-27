# Terminal Visual Language Review (Superseded)

This initial exploration was rejected during review on 2026-08-26 because its
three directions still read as a generic colored-log page or a dashboard. It
is retained only as a record of the rejected exploration. Continue with the
GitHub CLI-inspired, command-local prototype in
`../github-cli-journeys/`.

---

This is a throwaway prototype for the Wayfinder ticket `Prototype the terminal
visual language`. It is not production ycy code and makes no commitment about
automation-input flag syntax.

Run the interactive gallery from this directory with `go run .`. Left/Right
or `h`/`l` changes direction; `q` exits. The three Huh demonstrations are
available through `go run . --form selection`, `--form confirm`, and
`--form secret`.

## Directions

| Direction | Structure | Best fit | Cost / risk |
| --- | --- | --- | --- |
| Signal | Spacious sectioned journey: question, work, outcome, report. | Default interactive command flow and small multi-step forms. | Can feel verbose if every short command renders all sections. |
| Ledger | Time-ordered transcript with minimal decoration. | Long-running operations whose progress should remain useful in scrollback. | Gives forms less spatial guidance and is less visibly distinct from plain logs. |
| Workbench | Controls and live state side-by-side on wide terminals; stacked below 80 columns. | Bubble Tea-only dynamic work such as Git Pulse or Git Fork. | Too much screen ownership for ordinary commands, child-process passthrough, browser URLs, or service logs. |

## Shared Grammar To Evaluate

- One compact `ycy` label is the whole brand treatment. There is no banner or
  startup animation.
- Color has semantic roles, not decoration: cyan/blue for focus, green for
  success, amber for warning, red/pink for error, and muted gray for context.
  `NO_COLOR` leaves the hierarchy and words intact.
- `OK`, `WARN`, and `ERROR` are short status labels. Long detail moves below
  the status on narrow terminals instead of being truncated.
- Interaction uses the familiar `>` focus marker, `enter` to confirm, and
  `esc` / `ctrl+c` to cancel. Unicode gives way to words and layout when color
  is unavailable; there are no decorative glyphs required to understand a
  result.
- Reports are compact tables; a 60-column terminal stacks dynamic work rather
  than hiding or clipping it.
- Automation Session examples deliberately remain plain, stream-oriented
  lines. They never open a form, prompt, animate, or read an implicit secret.
  The exact command-specific Automation Input surface remains pending in
  `Choose automation inputs for prompted commands`.

## Proposed Default

Use **Signal** as the shared interactive grammar. Borrow **Ledger**'s durable
plain progress lines for non-transient and plain interactive output. Reserve
**Workbench** for the small set of commands that later earn a Bubble Tea live
view; it should not become a global CLI shell.
