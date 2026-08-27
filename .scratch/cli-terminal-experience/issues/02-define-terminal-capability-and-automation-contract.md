# Define the terminal capability and automation contract

Type: grilling
Status: resolved

## Question

What exact capability profile determines whether ycy enters an Interactive
Session or an Automation Session? Decide independently for stdin, stdout, and
stderr how ycy detects a usable terminal; the color policy for `NO_COLOR`,
`TERM=dumb`, CI, and any explicit color override; and the fallback behavior
when an existing prompt would otherwise be needed from a pipe or redirected
stream.

State the resulting contract for destructive confirmations, secret entry,
progress, cancellation, table-like reports, and error presentation. The
decision must preserve callers that consume stdout and must be testable through
real and pseudo-terminal scenarios on the supported target platforms.

## Comments

- Claimed for the first human decision round on 2026-08-26.

## Answer

Approved on 2026-08-26. ycy owns session classification before constructing
Huh, Bubble Tea, or any other terminal runtime. It never delegates this
decision to a dependency and never opens a controlling terminal to escape a
redirected standard stream.

### Session classification

Classify each invocation once from its inherited standard streams and
environment:

1. **Automation Session** applies when `CI` is set or any of stdin, stdout,
   or stderr is not a terminal. It is the conservative default.
2. **Plain Interactive Session** applies only when all three standard streams
   are terminals, `CI` is unset, and rich capability is not positively known.
   `TERM=dumb`, an empty `TERM`, or an unknown terminal type all produce this
   state.
3. **Rich Interactive Session** applies only when all three standard streams
   are terminals, `CI` is unset, and ycy positively recognizes a terminal that
   can safely render rich controls.

`NO_COLOR` is a style modifier, not a session classifier: it disables color
but does not demote a Rich Interactive Session or disable its safe transient
controls. ycy adds no `--color` flag. Exact terminal recognition mechanics are
an implementation detail, but a false positive is unacceptable: unknown
capability must become Plain Interactive.

### Interaction and fallback contract

- Only Interactive Sessions may prompt. Automation Sessions never read an
  implicit answer from stdin, never open `/dev/tty` or `CONIN$`, and never
  show a transient view, cursor control, animation, or alternate screen.
- A missing value that would have been prompted for causes an Automation
  Session to fail before any side effect. It writes one actionable plain-text
  error to stderr, writes no partial Command Result to stdout, and returns the
  existing ordinary error exit code `1`.
- Destructive confirmation remains explicit. An Automation Session may use an
  existing command-specific non-prompting option such as `--force`, but ycy
  must never infer confirmation or introduce a global `--yes` or
  `--non-interactive` flag. Whether a particular command later gains a narrow
  explicit input is a per-command decision.
- Secret entry is permitted only from a terminal with input echo disabled. A
  pipe is not a secret-input channel; Automation Session secret configuration
  requires an explicit command-owned surface decided with that command.
- Plain Interactive Sessions use only ordinary newline-delimited text and
  simple line prompts. Rich Interactive Sessions may use Huh, Lip Gloss, and
  approved Bubble Tea views. Automation Session reports and errors remain
  stable, unstyled, newline-delimited text. The Command Result and Diagnostic
  Event formats and their detailed streams remain owned by the later output
  contract decision.
- Every Rich Interactive transient view must restore cursor, screen, and input
  state before returning. `Ctrl+C`, Esc, and an in-view cancel option hand
  control back to each command's existing cancellation and exit semantics;
  this terminal decision does not redefine business cancellation behavior.

### Acceptance evidence

The later migration plan must prove the matrix with real or pseudo terminals:

| Session | Required evidence |
| --- | --- |
| Rich Interactive | all streams TTY, rich view may render, `NO_COLOR` removes only color, and all cancellation paths restore the terminal |
| Plain Interactive | all streams TTY with `TERM=dumb`, simple prompts work, and no ANSI, cursor, animation, or alternate-screen bytes are written |
| Automation | each redirected-stream permutation and `CI` avoid prompts/control-terminal access, preserve stdout, and fail missing requirements with stderr plus exit code `1` before effects |

This decision follows the completed [Research the Charmbracelet terminal
runtime contract](01-research-charmbracelet-terminal-runtime.md), which proves
that Bubble Tea is not itself a safe Automation Session fallback.
