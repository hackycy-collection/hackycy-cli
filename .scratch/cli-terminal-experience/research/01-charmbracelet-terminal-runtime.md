# Charmbracelet Terminal Runtime Research

Date: 2026-08-26

## Decision

Use a small, pinned Charmbracelet surface at the terminal adapter boundary:

```go
require (
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/huh v1.0.0
	github.com/charmbracelet/lipgloss v1.1.0
)
```

`golang.org/x/term v0.45.0` is already a direct ycy dependency and remains
the application-owned API for TTY detection and terminal dimensions. Add
`github.com/charmbracelet/bubbles v1.0.0` only in the slice that imports a
Bubbles component for a complex Bubble Tea view. Do not add
`github.com/muesli/termenv`, `github.com/mattn/go-isatty`,
`github.com/charmbracelet/x/term`, or `github.com/charmbracelet/log` as direct
production dependencies initially.

This is a CGO-free graph on the supported build targets, but it is **not** a
license for libraries to choose ycy's automation, stdout/stderr, logging, or
cancellation semantics. ycy must select its terminal mode before constructing
a Huh form or Bubble Tea program.

## Version And Graph Evidence

| Module | Pin | Why it is direct | Upstream compatibility evidence |
| --- | --- | --- | --- |
| `github.com/charmbracelet/huh` | `v1.0.0` | Forms, selects, confirms, and text/password input | The tag declares Go `1.23.0`, Bubble Tea `v1.3.6`, and Lip Gloss `v1.1.0`. [Source][huh-mod] |
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | Complex selection, live progress, and cancellable dynamic views | The tag declares Go `1.24.0` and Lip Gloss `v1.1.0`. [Source][tea-mod] |
| `github.com/charmbracelet/lipgloss` | `v1.1.0` | Shared styling for non-TUI presenters and selected Bubble Tea views | The tag declares Go `1.18` and termenv `v0.16.0`. [Source][lipgloss-mod] |
| `github.com/charmbracelet/bubbles` | `v1.0.0`, conditional | Use only when ycy directly imports a component such as a spinner, list, or progress model | The tag declares Go `1.24.2`, Bubble Tea `v1.3.10`, and Lip Gloss `v1.1.0`. [Source][bubbles-mod] |
| `golang.org/x/term` | existing `v0.45.0` | Application-owned `IsTerminal` and `GetSize` capability checks | The tag declares Go `1.25.0`; its public API covers TTY detection, raw state, restore, and size. [Module][go-term-mod] [API][go-term] |

The highest declared Go version in the selected graph is `1.25.0`, so the
repository's `go 1.26.4` / `toolchain go1.26.7` is sufficient. Bubble Tea
internally uses **`github.com/charmbracelet/x/term`**, which is a different
module from ycy's `golang.org/x/term`; ycy should not import the internal
choice merely to duplicate its own capability checks.

Huh `v1.0.0` declares an earlier pre-`v1` Bubbles revision. If a complex view
needs Bubbles, adding the direct `v1.0.0` pin causes Go's minimal-version
selection to use that exact stable release for both. That combination compiled
and Huh's package tests passed with Go `1.26.7`, but the Huh `go.mod` does not
itself promise Bubbles `v1` compatibility. Keep the pin explicit and retain a
small Huh smoke test when the graph changes.

### Local Verification

In a temporary module containing the exact direct pins above, the following
were successful with `GOTOOLCHAIN=go1.26.7`:

- `CGO_ENABLED=0 go build` for Huh, Bubble Tea, Bubbles, and Lip Gloss on the
  native Darwin target, plus Linux/amd64 and Windows/amd64 cross-compiles.
- `go list -deps` reported zero packages with `CgoFiles` for the selected
  dependency closure on Darwin, Linux/amd64, and Windows/amd64.
- `CGO_ENABLED=0 go test github.com/charmbracelet/huh` passed on the native
  target after adding Huh's test dependencies.

This proves the pinned graph compiles without CGO on those target builds; it
does not guarantee that a future dependency upgrade remains CGO-free. The
release build should keep `CGO_ENABLED=0`, and the dependency update workflow
should rerun the same checks.

## Runtime Contract

### I/O And Non-TTY Behavior

Bubble Tea's public options default output to stdout and input to stdin.
`WithInput(nil)` disables input. Its *default* input behavior is deliberately
interactive: when redirected stdin is not a TTY, it opens `/dev/tty` on Unix
or `CONIN$` on Windows. [Program options][tea-options] [Unix TTY source][tea-tty-unix]
[Windows TTY source][tea-tty-windows]

The framework does not automatically become a safe plain renderer when output
is redirected. It constructs the normal renderer unless `WithoutRenderer` is
chosen; terminal setup still hides the cursor, and the standard renderer writes
cursor and redraw escape sequences to its configured writer. `WithoutRenderer`
uses a nil renderer, so it is not a useful plain presentation layer because it
does not render `View()` at all. [Terminal initialization][tea-tty]
[Renderer][tea-renderer] [Nil renderer][tea-nil-renderer]

**ycy-owned constraint:** classify the session before creating Huh or Bubble
Tea. An Automation Session must never invoke either framework to achieve its
plain mode. It must use its own deterministic presenter, accept only explicit
flags or supplied input, and preserve machine-readable stdout. A full
capability policy belongs to the terminal-capability ticket, but it must at
least distinguish these cases:

| Session | Required ycy behavior |
| --- | --- |
| Advanced interactive | Only when the input and diagnostic-output endpoints are genuine TTYs and the terminal is capable of redraws. Pass explicit input/output rather than relying on library defaults. |
| Plain interactive / accessible | A TTY may still need line-oriented prompting, but it must be explicitly selected. `TERM=dumb` does not authorize an Automation Session to block for input. |
| Automation | If either needed endpoint is non-TTY, run no Huh or Bubble Tea prompt/TUI, no animation, no cursor control, and no implicit `/dev/tty` fallback. Missing required input becomes the command's stable non-interactive error. |

Huh's normal-form default is stderr, while accessible mode falls back to stdout
when no output is supplied. It exposes `WithInput`, `WithOutput`, and
`WithProgramOptions`; the latter **replaces** rather than appends the form's
existing Bubble Tea options. [Huh form construction and `TERM=dumb`][huh-form]
[Huh I/O options][huh-io] [Huh execution][huh-run]

Therefore every ycy form must explicitly configure its session input and
diagnostic writer. When ycy later chooses to own OS signal handling, a wrapper
that changes Huh's program options must preserve both paths, for example by
setting `WithInput`/`WithOutput` for accessible mode and supplying equivalent
`tea.WithInput`/`tea.WithOutput` values alongside
`tea.WithoutSignalHandler()` for normal mode.

### Alternate Screen, Cursor, And Child Processes

`tea.WithAltScreen()` starts a full-window alternate buffer; Bubble Tea exits
it on normal program shutdown. `EnterAltScreen` is asynchronous and upstream
explicitly says not to use it from `Init`; use the option when a view starts
full-screen. [Alternate-screen option][tea-altscreen] [Screen API][tea-screen]

At normal shutdown Bubble Tea stops the renderer, disables bracketed paste,
mouse, and focus reporting, shows the cursor, exits the alternate screen, and
restores saved terminal state. Default panic recovery is part of that cleanup;
do not opt into `WithoutCatchPanics`. [Shutdown][tea-shutdown]
[Terminal restoration][tea-restore] [Panic option][tea-panic]

**ycy-owned constraint:** use an alternate screen only for the genuinely
full-screen Bubble Tea views the product selects. Do not use it for ordinary
Huh forms, one-line progress, or automation. Cleanup covers `Run`'s controlled
exit and Bubble Tea's default panic recovery, not `SIGKILL`, power loss, or an
application that bypasses the program lifecycle.

For a foreground editor or other interactive child process, Bubble Tea offers
`ReleaseTerminal`/`RestoreTerminal` and `ExecProcess`. Release/restore retains
alternate-screen, bracketed-paste, and focus state, but not mouse mode, so
mouse support must be re-enabled and tested if ycy ever uses it. [Release and
restore][tea-release] [Child process helper][tea-exec]

### Cancellation And Signals

By default Bubble Tea registers handlers for `SIGINT` and `SIGTERM`; raw TTY
mode usually delivers Ctrl-C as a `KeyMsg`, while non-TTY Ctrl-C reaches the
signal handler. It translates `SIGINT` to `InterruptMsg` and `SIGTERM` to a
graceful `QuitMsg`. `WithContext` exits on context cancellation, and
`WithoutSignalHandler` exists for applications that own their own signal
policy. [Signal handler][tea-signals] [Context and signal options][tea-options]

Huh binds Ctrl-C to quit. Its normal form maps that interruption to
`huh.ErrUserAborted`. More subtly, its `RunWithContext` maps every Bubble Tea
`ErrProgramKilled` to `huh.ErrTimeout`, including cancellation of a caller's
context. [Huh Ctrl-C keymap][huh-keymap] [Huh cancellation mapping][huh-run]

**ycy-owned constraint:** define exactly one OS-signal owner. The recommended
direction is the Cobra composition root: derive a root context from signals,
pass it to all command work, then use `tea.WithContext(ctx)` plus
`tea.WithoutSignalHandler()` in advanced UI wrappers. Map Ctrl-C key events to
the same operation cancellation path, and inspect `ctx.Err()` before treating
Huh's `ErrTimeout` as a genuine form timeout. A Bubble Tea `Kill` or UI quit
does not cancel a long-running domain worker: Bubble Tea commands run in
goroutines and cannot be forcibly cancelled by the framework. Domain work must
honor ycy's context. [Command execution][tea-commands] [Program termination][tea-kill]

### Color, `NO_COLOR`, And `TERM=dumb`

Lip Gloss creates a renderer for a supplied writer and derives its color
profile from termenv. termenv returns ASCII for a non-TTY, honors non-empty
`NO_COLOR`, and also understands `CLICOLOR` / `CLICOLOR_FORCE`. [Lip Gloss
renderer][lipgloss-renderer] [termenv environment policy][termenv-env]
[termenv Unix profile][termenv-unix]

Huh notices `TERM=dumb` at form construction, fixes a default width of 80, and
uses its accessible, non-redrawing prompt path. It does not independently test
TTY eligibility. [Huh form construction][huh-form] [Accessible execution][huh-run]

`NO_COLOR` must mean "remove color," not "turn a terminal session into
automation." Conversely, `TERM=dumb` must be treated by ycy's capability
policy as no advanced redraw UI, even though termenv's color profile itself
looks at `COLORTERM` before falling through to `TERM`. A plain/accessible path
may still be offered only for a real interactive terminal.

There is one integration detail to keep explicit: Huh themes create Lip Gloss
styles through the package-global default renderer, whereas a Huh form writes
to its separately configured Bubble Tea output. Lip Gloss's default renderer
is initialized with termenv's default output (stdout). [Lip Gloss default
renderer][lipgloss-renderer] [termenv default output][termenv-output]
[Huh theme source][huh-theme]

For ycy's own presenters, construct a renderer against the diagnostic writer
and retain it in the terminal-session adapter. If ycy changes Lip Gloss's
global renderer so Huh colors match stderr, do it once at process setup, do
not mutate it during a live UI, and restore/serialize it in tests because it
is process-global. Import termenv directly only if a later `--color` policy
needs typed, forced profile values; pin its already-resolved `v0.16.0` then.

### Width, Windows, And Logging

Bubble Tea's automatic initial size probe and Unix resize listener are enabled
only when it has a TTY output. Windows has no `SIGWINCH` listener; its console
input reader can emit window-buffer-size events, but ycy must not promise a
resize event on every Windows terminal. Models must handle an initial size and
tolerate a later size change. [Resize initialization][tea-resize] [Windows
resize source][tea-resize-windows] [Windows console input][tea-key-windows]
[Window-size API][tea-screen]

On Windows Bubble Tea saves console state and enables virtual-terminal input
and output. Its Windows input implementation warns that cancellation of an
outstanding console read is not fully reliable. termenv also has Windows color
fallbacks for older console versions. [Bubble Tea Windows terminal setup][tea-tty-windows]
[Bubble Tea Windows reader][tea-input-windows] [termenv Windows profile][termenv-windows]

**ycy-owned constraint:** keyboard-first interaction is the supported baseline
on Windows. Keep mouse optional, use a native Windows smoke test for cursor
restoration and Ctrl-C, and make resize layout resilient rather than relying
on continuous resize signals.

`tea.Println` and `tea.Printf` place unmanaged output above an inline program,
but deliberately suppress it in the alternate screen. Direct stdout/stderr
writes can corrupt a live renderer. Bubble Tea's own logging helper writes to
a file; it is not a structured diagnostic framework. [Inline printing][tea-print]
[Bubble Tea logging][tea-logging]

**ycy-owned constraint:** treat interactive UI state and diagnostic logs as
separate sinks. A live full-screen model should receive user-facing progress
as messages; diagnostics should go to a selected file/buffer or be presented
after the UI exits. Plain mode can keep command results on stdout and
diagnostics on stderr through ycy's existing logging module. The exact
`--log-format`, quiet, and verbose behavior remains the output-contract
decision.

## Test Strategy Enabled By The APIs

- Test the capability classifier independently by injecting TTY/size/environment
  facts. Do not depend on the test process's actual terminal.
- Test Bubble Tea model reducers directly with `tea.KeyMsg` and
  `tea.WindowSizeMsg`; use `WithInput`, `WithOutput`, `WithContext`, and
  `WithoutSignals` with buffers for program-lifecycle tests. [Upstream test
  examples][tea-tests]
- Test Huh's accessible path with explicit buffer I/O. For normal forms, test
  the ycy wrapper's option construction and cancellation translation; reserve
  PTY tests for a small end-to-end suite.
- Construct a local Lip Gloss renderer with an explicit profile in styling
  tests; never depend on the cached global profile. [Lip Gloss renderer
  tests][lipgloss-tests]
- Maintain non-TTY subprocess tests that assert no escape sequences, no prompt,
  correct stdout/stderr separation, and stable exits. Add a Windows build and
  native smoke lane for raw-mode restoration and cancellation.
- Bubble Tea examples use the separate experimental
  `github.com/charmbracelet/x/exp/teatest` package for higher-fidelity terminal
  tests. It is not a required or stable runtime dependency, so prefer ycy's
  adapter tests unless a view needs ANSI golden testing. [Example test][tea-teatest]

## Sources

All behavioral sources below are first-party repository source at the cited
release commit/tag, or Go's official package documentation.

[huh-mod]: https://github.com/charmbracelet/huh/blob/9dc45e34a40badf1dc3e68a5b06573e815324024/go.mod#L1-L42
[tea-mod]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/go.mod#L1-L26
[bubbles-mod]: https://github.com/charmbracelet/bubbles/blob/4824effc3f91c9517c776d8200ef99a1207136e0/go.mod#L1-L39
[lipgloss-mod]: https://github.com/charmbracelet/lipgloss/blob/f0e45475a64ee60d712b81145172d3739db36a93/go.mod#L1-L27
[go-term-mod]: https://github.com/golang/term/blob/v0.45.0/go.mod#L1-L5
[go-term]: https://pkg.go.dev/golang.org/x/term@v0.45.0
[tea-options]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/options.go#L17-L113
[tea-tty-unix]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tty_unix.go#L11-L38
[tea-tty-windows]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tty_windows.go#L14-L64
[tea-tty]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tty.go#L25-L74
[tea-renderer]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/standard_renderer.go#L160-L420
[tea-nil-renderer]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/nil_renderer.go#L1-L29
[huh-form]: https://github.com/charmbracelet/huh/blob/9dc45e34a40badf1dc3e68a5b06573e815324024/form.go#L98-L129
[huh-io]: https://github.com/charmbracelet/huh/blob/9dc45e34a40badf1dc3e68a5b06573e815324024/form.go#L330-L355
[huh-run]: https://github.com/charmbracelet/huh/blob/9dc45e34a40badf1dc3e68a5b06573e815324024/form.go#L656-L720
[tea-altscreen]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/options.go#L95-L113
[tea-screen]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/screen.go#L1-L49
[tea-shutdown]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tea.go#L804-L917
[tea-restore]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tty.go#L39-L115
[tea-panic]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/options.go#L77-L85
[tea-release]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tea.go#L862-L917
[tea-exec]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/exec.go#L101-L133
[tea-signals]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tea.go#L273-L312
[huh-keymap]: https://github.com/charmbracelet/huh/blob/9dc45e34a40badf1dc3e68a5b06573e815324024/keymap.go#L106-L115
[tea-commands]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tea.go#L331-L366
[tea-kill]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tea.go#L792-L835
[lipgloss-renderer]: https://github.com/charmbracelet/lipgloss/blob/f0e45475a64ee60d712b81145172d3739db36a93/renderer.go#L10-L108
[termenv-output]: https://github.com/muesli/termenv/blob/2e6fa35162bb1c735367736319813b2dc01a77a4/output.go#L9-L86
[termenv-env]: https://github.com/muesli/termenv/blob/2e6fa35162bb1c735367736319813b2dc01a77a4/termenv.go#L28-L115
[termenv-unix]: https://github.com/muesli/termenv/blob/2e6fa35162bb1c735367736319813b2dc01a77a4/termenv_unix.go#L21-L76
[huh-theme]: https://github.com/charmbracelet/huh/blob/9dc45e34a40badf1dc3e68a5b06573e815324024/theme.go#L1-L153
[tea-resize]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tty.go#L119-L140
[tea-resize-windows]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/signals_windows.go#L1-L10
[tea-key-windows]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/key_windows.go#L36-L94
[tea-input-windows]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/inputreader_windows.go#L66-L73
[termenv-windows]: https://github.com/muesli/termenv/blob/2e6fa35162bb1c735367736319813b2dc01a77a4/termenv_windows.go#L14-L44
[tea-print]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/standard_renderer.go#L620-L672
[tea-logging]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/logging.go#L11-L52
[tea-tests]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/tea_test.go#L63-L100
[lipgloss-tests]: https://github.com/charmbracelet/lipgloss/blob/f0e45475a64ee60d712b81145172d3739db36a93/renderer_test.go#L1-L58
[tea-teatest]: https://github.com/charmbracelet/bubbletea/blob/9edf69c677c7353eca5fae6d3ea3986af39717b7/examples/simple/main_test.go#L14-L76
