# Latest compatible Charm v2 stack

Research cutoff: 2026-08-31. This note uses upstream release pages, tagged
source, tagged `go.mod` files, and Go's module specification. It does not treat
the repository default branches as released APIs.

## Executive findings

1. The coherent latest stable set is Bubble Tea `v2.0.9`, Bubbles `v2.2.1`,
   Huh `v2.0.3`, Lip Gloss `v2.0.6`, and `charm.land/log/v2` `v2.0.0`. The
   effective Go floor is `1.25.8`, set by Huh and Log; this repository's
   `go 1.26.4` satisfies it.
2. Bubble Tea v2 still supports explicit `io.Reader`/`io.Writer` injection, but
   alternate-screen state moved from `tea.WithAltScreen()` to
   `tea.View.AltScreen`. This is the API change that lets one root model choose
   full-screen or inline behavior per command/state.
3. Exiting an alternate screen restores the primary screen; it does not copy
   the TUI into scrollback. Bubble Tea's durable `Println`/`Printf` output is
   explicitly suppressed while AltScreen is active. Therefore the map's
   Interaction Transcript must be rendered after the alternate screen closes,
   not inferred from renderer history.
4. Huh v2 remains the best-aligned form layer inside this stack. It includes
   input/password, text area, select, multi-select, confirm, note, and file
   picker controls; its select controls already cover filtering, scrolling,
   half-page movement, first/last movement, and select-all/none. Its theme model
   exposes all form/field states as Lip Gloss styles.
5. Lip Gloss v2 removed `Renderer`. Styles are pure values and emit
   full-fidelity ANSI; output/profile downsampling moved to the writer or Bubble
   Tea. Both current `lipgloss.NewRenderer(...)` call sites must be redesigned,
   not merely have their imports changed.
6. Log v2 offers attractive text output, JSON/logfmt, levels, structured
   key/value fields, `With`, output injection, and `slog.Handler`. It has no
   credential-redaction facility and no pluggable formatter interface; its
   formatter is a closed three-value enum. A direct replacement for
   `internal/logging` would violate the existing redaction and NDJSON schema
   contracts. Any adoption must remain behind the existing logging facade and
   retain pre-format redaction plus the existing JSON projection.
7. All five modules are MIT licensed.

## Stable version and compatibility matrix

| Module | Latest stable at cutoff | Module path / Go directive | Direct Charm-v2 requirements | License |
| --- | --- | --- | --- | --- |
| Bubble Tea | [`v2.0.9`, 2026-08-19](https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.9) | `charm.land/bubbletea/v2`; `go 1.25.0` ([tagged `go.mod`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/go.mod)) | None of the other four modules | [MIT](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/LICENSE) |
| Bubbles | [`v2.2.1`, 2026-08-24](https://github.com/charmbracelet/bubbles/releases/tag/v2.2.1) | `charm.land/bubbles/v2`; `go 1.25.0` ([tagged `go.mod`](https://github.com/charmbracelet/bubbles/blob/v2.2.1/go.mod)) | Bubble Tea `v2.0.8`, Lip Gloss `v2.0.5` | [MIT](https://github.com/charmbracelet/bubbles/blob/v2.2.1/LICENSE) |
| Huh | [`v2.0.3`, 2026-03-10](https://github.com/charmbracelet/huh/releases/tag/v2.0.3) | `charm.land/huh/v2`; `go 1.25.8` ([tagged `go.mod`](https://github.com/charmbracelet/huh/blob/v2.0.3/go.mod)) | Bubbles `v2.0.0`, Bubble Tea `v2.0.2`, Lip Gloss `v2.0.1` | [MIT](https://github.com/charmbracelet/huh/blob/v2.0.3/LICENSE) |
| Lip Gloss | [`v2.0.6`, 2026-08-11](https://github.com/charmbracelet/lipgloss/releases/tag/v2.0.6) | `charm.land/lipgloss/v2`; `go 1.25.0` ([tagged `go.mod`](https://github.com/charmbracelet/lipgloss/blob/v2.0.6/go.mod)) | None of the other four modules | [MIT](https://github.com/charmbracelet/lipgloss/blob/v2.0.6/LICENSE) |
| Log | [`v2.0.0`, 2026-03-09](https://github.com/charmbracelet/log/releases/tag/v2.0.0) | `charm.land/log/v2`; `go 1.25.8` ([tagged `go.mod`](https://github.com/charmbracelet/log/blob/v2.0.0/go.mod)) | Lip Gloss `v2.0.1` | [MIT](https://github.com/charmbracelet/log/blob/v2.0.0/LICENSE) |

Pinning all five versions above is module-graph compatible: Bubbles' minimums
are the highest transitive requirements, while the selected Bubble Tea and Lip
Gloss versions are newer releases within the same v2 module paths. Go minimal
version selection chooses the highest required version of each module
([Go Modules Reference](https://go.dev/ref/mod#minimal-version-selection)).
The `go 1.25.8` directives from Huh and Log dominate the set. The repository's
[`go.mod`](../../../go.mod) already declares `go 1.26.4` and toolchain
`go1.26.7`, so no Go downgrade or toolchain exception is required.

This was also validated rather than inferred only from the graph: a temporary
module declaring `go 1.26.4` and `toolchain go1.26.7`, directly requiring all
five exact versions, selected the same five versions and passed `go build`
under `GOTOOLCHAIN=go1.26.7`.

The Log `v2.0.0` tagged README still shows the old GitHub `go get` and import
path in its early usage section. Its tagged `go.mod`, release announcement,
and [v2 upgrade guide](https://github.com/charmbracelet/log/blob/v2.0.0/UPGRADE_GUIDE_V2.md#L27-L41)
all agree that `charm.land/log/v2` is authoritative.

## Bubble Tea v2

### Input and output ownership

- `tea.WithInput(io.Reader)` and `tea.WithOutput(io.Writer)` remain program
  options. Passing nil input disables input; output defaults to stdout
  ([`options.go`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/options.go#L28-L44)).
- If no input is supplied and process stdin is not a terminal, `Run` tries to
  open the controlling TTY. Explicit injection therefore remains important for
  the repository's stream ownership and tests
  ([`tea.go`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/tea.go#L1015-L1025)).
- `WithoutSignalHandler`, `WithoutRenderer`, `WithContext`, `WithColorProfile`,
  and `WithWindowSize` remain available. `WithoutRenderer` is a non-TUI mode,
  not an accessible form renderer
  ([`options.go`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/options.go#L19-L102)).

This preserves the current architecture in which the rich renderer reads the
injected stdin terminal and writes only to the injected diagnostic/stderr
terminal, leaving stdout for the durable Command Result.

### AltScreen and renderer lifecycle

Bubble Tea v2 makes terminal features declarative. A model now returns
`tea.View`, and `View.AltScreen` replaces `tea.WithAltScreen()` and the old
enter/exit commands
([v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/UPGRADE_GUIDE_V2.md#L39-L102),
[`tea.View`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/tea.go#L82-L173)).
The important consequences are:

- `AltScreen == false` is inline mode; `true` is full-window mode.
- The model can change the value across renders, so the terminal Experience can
  make this a per-command display policy without constructing a second program.
- On shutdown, the renderer restores terminal state and exits AltScreen; its
  source explicitly disables AltScreen before closing
  ([`cursed_renderer.go`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/cursed_renderer.go#L171-L209)).
- `Program.ReleaseTerminal` stops the renderer, cancels input, and restores the
  terminal; `RestoreTerminal` reinitializes input and repaints. These APIs are
  for temporary external-process/editor ownership, not for transcript replay
  ([`tea.go`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/tea.go#L1326-L1365)).

### What actually persists in scrollback

`tea.Println`/`tea.Printf` (and the equivalent `Program` methods) insert
unmanaged lines above an inline renderer and promise that those lines persist
across renders. Their API explicitly says nothing is printed while AltScreen is
active
([`renderer.go`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/renderer.go#L59-L92),
[`Program.Println`](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/tea.go#L1377-L1398)).

Therefore there are two valid but different lifecycle patterns:

| Pattern | During interaction | Durable history |
| --- | --- | --- |
| Inline form/log | Keep only the active control in `View`; commit completed semantic entries with `tea.Println` | Completed entries remain above the renderer |
| Alternate-screen Live View | Set `View.AltScreen = true`; render freely | Close the program first, then write a separately accumulated Interaction Transcript to stderr |

The wayfinder map has already chosen the second pattern for finite interactive
commands. Bubble Tea supplies cleanup, but the repository must own the semantic
transcript state and replay ordering. Renderer frames, keystrokes, filters, and
secret input must never become that transcript.

### Repository API breakpoints

The v2 upgrade is more than import rewriting:

- `View() string` becomes `View() tea.View`.
- v1 `tea.KeyMsg` becomes an interface; normal key handling should switch to
  `tea.KeyPressMsg`. `Type`, `Runes`, and `Alt` become `Code`, `Text`, and
  modifiers respectively
  ([upgrade guide key table](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/UPGRADE_GUIDE_V2.md#L470-L486)).
- `tea.WithAltScreen()` and `tea.WithoutBracketedPaste()` disappear; set
  `View.AltScreen` and `View.DisableBracketedPasteMode`
  ([removed-options table](https://github.com/charmbracelet/bubbletea/blob/v2.0.9/UPGRADE_GUIDE_V2.md#L285-L297)).

Those changes directly affect `internal/terminal/rich.go`,
`internal/terminal/rich_form.go`, and their tests.

## Bubbles v2

Bubbles is a component set, not a form orchestrator. `v2.2.1` includes cursor,
file picker, help, key bindings, list, paginator, progress, spinner, stopwatch,
table, textarea, text input, timer, tree, and viewport packages. The official
README documents the main controls and their intended composition inside
Bubble Tea
([README](https://github.com/charmbracelet/bubbles/blob/v2.2.1/README.md#L21-L198)).

Relevant v2 constraints:

- Imports move to `charm.land/bubbles/v2/...`; the tagged module requires the
  v2 Bubble Tea and Lip Gloss module paths
  ([`go.mod`](https://github.com/charmbracelet/bubbles/blob/v2.2.1/go.mod)).
- Renderer-specific constructors were removed because Lip Gloss v2 has no
  renderer. Several mutable sizing fields became setters/getters, and key map
  variables became functions
  ([upgrade guide](https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md#L77-L150)).
- Styles that adapt to background now take an explicit `isDark` value; the
  upgrade guide shows `DefaultStyles(isDark)` and explicit style assignment
  ([upgrade guide](https://github.com/charmbracelet/bubbles/blob/v2.2.1/UPGRADE_GUIDE_V2.md#L170-L204)).

This repository currently consumes Bubbles only through Huh. Direct adoption is
most useful for command-specific displays such as long lists, tables, progress,
or viewports; it should not become a universal command presentation abstraction.

## Huh v2

### Controls and behavior

Huh v2 provides `Input`, multiline `Text`, `Select`, `MultiSelect`, `Confirm`,
`Note`, and `FilePicker` fields. Input still provides `EchoModePassword` and
`EchoModeNone` for secrets
([field reference](https://github.com/charmbracelet/huh/blob/v2.0.3/README.md#L134-L236),
[`field_input.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/field_input.go#L176-L196)).

The upstream select implementation already supplies most behavior duplicated
by this repository's custom `richListForm`:

- select filtering with `/`, j/k, arrows, ctrl+n/ctrl+p, dynamic options, and a
  height-limited scrolling viewport
  ([`field_select.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/field_select.go#L60-L66),
  [`Height`](https://github.com/charmbracelet/huh/blob/v2.0.3/field_select.go#L241-L265));
- both select key maps include half-page, first/last, filter, and submit actions;
  multi-select also includes toggle and select-all/none
  ([`keymap.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/keymap.go#L35-L70));
- multi-select supports filtering, a selection limit, width/height, and
  validation
  ([`field_multiselect.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/field_multiselect.go#L176-L217)).

Huh options have a label (`Key`) and typed `Value`, but not this repository's
separate `InteractionOption.Description`
([`option.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/option.go#L5-L37)).
Replacing `richListForm` is therefore viable only after deciding how per-option
descriptions should be projected and PTY testing the largest command lists.

### Theme surface

`Theme` is an interface that resolves styles from a light/dark boolean. `Styles`
contains form, group, separator, blurred/focused field, and help styles; field
styles separately expose title, description, errors, select cursor/options,
multi-select prefixes/states, input cursor/prompt/text/placeholder, confirm
buttons, notes, and file-picker entries
([`theme.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/theme.go#L9-L90)).
Built-ins are Charm, Dracula, Catppuccin, Base16, and Base; custom themes can
start from `ThemeBase`
([README themes](https://github.com/charmbracelet/huh/blob/v2.0.3/README.md#L257-L279),
[`theme.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/theme.go#L97-L189)).
Because `WithTheme` accepts the `Theme` interface, a repository theme function
with the built-in signature is passed as `huh.ThemeFunc(repositoryTheme)`; the
named `ThemeFunc` type supplies the interface method
([`theme.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/theme.go#L9-L20)).

This is sufficient for a repository-owned vivid visual system without forking
Huh. Keep the palette and semantic visual roles in repository helpers, then
project them into Huh styles per command/form.

### I/O, accessibility, and embedding

- A normal standalone form defaults to stderr; `WithInput` and `WithOutput`
  append Bubble Tea input/output options. Accessible mode defaults to stdout
  unless output is explicitly supplied
  ([`form.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go#L104-L134),
  [`WithInput`/`WithOutput`](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go#L328-L341)).
- `WithProgramOptions` replaces Huh's option slice instead of appending to it.
  A caller using it must include every required output/input option
  ([`form.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go#L350-L353)).
- `WithAccessible(true)` disables Bubble Tea redraw and runs line-oriented
  prompts for screen readers. `TERM=dumb` enables it automatically
  ([`form.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go#L129-L132),
  [`WithAccessible`](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go#L229-L237)).
- Standalone `Run` wraps Huh's string-view compatibility model in a Bubble Tea
  v2 `tea.View`; `WithViewHook` can then set AltScreen or other view properties
  ([`form.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go#L653-L707),
  [`internal/compat/model.go`](https://github.com/charmbracelet/huh/blob/v2.0.3/internal/compat/model.go#L6-L38)).

The repository embeds Huh inside its own root model instead of calling
`Form.Run`. Preserve that ownership: adapt Huh's string `View` as content of the
root `tea.View`, and let only the root choose AltScreen, focus reporting,
bracketed paste, transcript, and stderr lifecycle. Automation should continue
to bypass Huh entirely; Huh accessible mode is interactive fallback, not
automation mode.

## Lip Gloss v2

Lip Gloss v2 is intentionally pure. The `Renderer` type,
`NewRenderer`, renderer-bound `NewStyle`, and renderer color-profile methods
were removed. `Style` is a plain value, and `Style.Render` always emits
full-fidelity ANSI
([upgrade guide](https://github.com/charmbracelet/lipgloss/blob/v2.0.6/UPGRADE_GUIDE_V2.md#L216-L249)).

Output behavior is now split by context:

- Bubble Tea v2 down-samples styled view output itself.
- Standalone output should use Lip Gloss writer functions or a configured
  `colorprofile.Writer`; the package-level `lipgloss.Writer` is global
  ([upgrade guide](https://github.com/charmbracelet/lipgloss/blob/v2.0.6/UPGRADE_GUIDE_V2.md#L245-L276)).
- Background selection is explicit via `LightDark(isDark)`; standalone
  detection accepts input/output streams, while Bubble Tea supplies
  `tea.BackgroundColorMsg`
  ([upgrade guide](https://github.com/charmbracelet/lipgloss/blob/v2.0.6/UPGRADE_GUIDE_V2.md#L280-L330)).

Repository constraints:

1. Replace the `*lipgloss.Renderer` field in `richRootModel` with pure style
   values/functions. Bubble Tea owns profile conversion for Live Views.
2. Replace `WriteRich`'s per-writer `lipgloss.NewRenderer(output)` path. Do not
   switch to the global `lipgloss.Writer`, because this repository deliberately
   injects streams per invocation. Use an explicitly owned output/profile layer,
   or preserve the current no-color/color decision when only ANSI base colors
   are used.
3. Preserve the rule that Plain Interactive and Automation output contain no
   terminal controls. A v2 style emits ANSI regardless of destination, so the
   existing semantic branch that creates unstyled styles when `Color == false`
   remains necessary.

## `charm.land/log/v2`

### Capability matrix

| Requirement | Log v2 fact | Consequence here |
| --- | --- | --- |
| Text output | Default human-readable formatter, customizable Lip Gloss v2 styles; styling only affects text ([README](https://github.com/charmbracelet/log/blob/v2.0.0/README.md#L185-L225)) | Suitable for redesigned Service Command text Lifecycle Logs. Output will not match the current exact text schema without deliberate style/key choices. |
| JSON | Built-in newline-terminated JSON formatter ([`formatter.go`](https://github.com/charmbracelet/log/blob/v2.0.0/formatter.go#L3-L14), [`json.go`](https://github.com/charmbracelet/log/blob/v2.0.0/json.go#L10-L30)) | It emits flat keys named `time`, `level`, `prefix`, `msg` plus caller keyvals, not the current `timestamp`, `level`, `scope`, `message`, nested `context` schema. |
| Logfmt | Built-in third formatter ([`logfmt.go`](https://github.com/charmbracelet/log/blob/v2.0.0/logfmt.go#L11-L31)) | Extra capability, not a current public CLI format; adding it would be a separate behavior decision. |
| Levels | Debug, Info, Warn, Error, Fatal with threshold filtering ([`level.go`](https://github.com/charmbracelet/log/blob/v2.0.0/level.go#L10-L64)) | Current four levels map directly; do not expose `Fatal` through the facade because it calls `os.Exit(1)`. |
| Structured fields | All log calls take flat key/value pairs; `With` clones a child logger with persistent fields ([README](https://github.com/charmbracelet/log/blob/v2.0.0/README.md#L69-L75), [`logger.go`](https://github.com/charmbracelet/log/blob/v2.0.0/logger.go#L330-L343)) | An adapter must convert current maps, preserve nested context, and sanitize values before calling Log. A child is a snapshot, so later level/formatter changes on its parent do not propagate as they do in the current Runtime. |
| Writer injection | `New(io.Writer)` / `NewWithOptions` and `SetOutput` accept an `io.Writer` ([`pkg.go`](https://github.com/charmbracelet/log/blob/v2.0.0/pkg.go#L41-L82), [`logger.go`](https://github.com/charmbracelet/log/blob/v2.0.0/logger.go#L276-L297)) | It can target the current lease-aware diagnostic writer. Default color detection reads `os.Environ`, so deterministic injected environment/capability behavior requires explicitly setting the color profile. |
| Time/caller | Options include time transform/format, timestamp, caller, prefix, fields, formatter ([`options.go`](https://github.com/charmbracelet/log/blob/v2.0.0/options.go#L39-L61)) | The time function transforms an internally sampled `time.Now`; the current Runtime's direct `Now func() time.Time` is a stronger deterministic seam and should stay at the facade. |
| Redaction | `Logger.handle` stringifies the message, appends fields unchanged, and immediately formats them; there is no redactor option or hook ([`logger.go`](https://github.com/charmbracelet/log/blob/v2.0.0/logger.go#L83-L145)) | Current recursive key/value, bearer, error, byte-slice, cycle, and unencodable-value redaction must remain before Log sees a record. Redacting serialized output afterward is too late and loses structure. |
| Formatter extensibility | `Formatter` is a `uint8` enum with only Text, JSON, and Logfmt; unknown values fall back to Text ([`formatter.go`](https://github.com/charmbracelet/log/blob/v2.0.0/formatter.go#L3-L14), [`logger.go`](https://github.com/charmbracelet/log/blob/v2.0.0/logger.go#L124-L135)) | The current stable NDJSON projection cannot be installed as a custom Log formatter. It must stay in the facade or be implemented as a separate handler/projection. |
| Handler integration | `*log.Logger` implements all four `slog.Handler` methods ([`logger_121.go`](https://github.com/charmbracelet/log/blob/v2.0.0/logger_121.go#L23-L72)) and can create a standard-library `*log.Logger` adapter ([README](https://github.com/charmbracelet/log/blob/v2.0.0/README.md#L310-L340)) | Useful as a leaf handler/adapter. Log does not accept an arbitrary handler internally, so middleware/redaction must wrap it externally or stay in `internal/logging`. |

### Diagnostic Record compatibility

The repository currently guarantees, and tests, all of the following in
[`internal/logging`](../../../internal/logging/logging.go):

- exact NDJSON keys and nesting: `timestamp`, `level`, optional `scope`,
  `message`, optional `context`;
- UTC millisecond timestamps and one record per line;
- recursive redaction by credential-shaped keys and strings, including nested
  maps/slices/structs/errors and cycle/unencodable handling;
- stable text projection and explicit color enablement;
- runtime level/format updates visible to already-created scoped loggers;
- injected clock and writer;
- diagnostics only on the chosen stderr stream.

Log v2 does not preserve these by direct substitution. Three adoption scopes
are technically possible for the later boundary decision:

1. **Keep `internal/logging` unchanged.** Lowest contract risk; Charm v2 is used
   only for terminal/form presentation.
2. **Use Log v2 only for text projection behind the current facade.** Keep
   current filtering, clock, scope, redaction, JSON projection, stream, and
   configuration. This can improve Service Command text appearance while JSON
   stays byte-compatible, but it creates two formatter implementations.
3. **Adapt both formats behind the facade.** Pre-redact and normalize a complete
   record before handing it to Log, and retain a separate exact JSON projection.
   Because Log has no custom formatter hook, this still does not eliminate the
   repository formatter boundary and offers little simplification over option 2.

Using Log v2 as a direct exported dependency of commands is incompatible with
the map: it would let commands bypass the repository's scope, redaction,
format-selection, lease, and schema policies.

## Repository migration constraints

| Existing boundary | Current implementation | v2 constraint |
| --- | --- | --- |
| Terminal streams | `ExperienceOptions` injects stdin, stdout, stderr; rich UI uses stderr while Result uses stdout ([`experience.go`](../../../internal/terminal/experience.go)) | Continue passing explicit Bubble Tea input/output. Do not let Huh accessible defaults or Lip Gloss globals reroute output. |
| Renderer ownership | One long-lived `richController` acquires an exclusive diagnostic writer lease ([`rich.go`](../../../internal/terminal/rich.go), [`lease.go`](../../../internal/terminal/lease.go)) | Keep one root renderer. Set terminal modes in its `tea.View`. Close it before transcript replay and before launching a child process. Release the lease before ordinary diagnostics/transcript are written, or explicitly write the transcript through the lease owner in a defined order. |
| Full-screen behavior | `tea.WithAltScreen()` is unconditional today | Replace with an explicit root-model display policy. For the map's finite commands it remains true during Live View, then a separately stored transcript is emitted after close. Service Commands should never start this controller. |
| Huh embedding | Text/secret/confirm use Huh; select/multi-select use custom `richListForm` ([`rich_form.go`](../../../internal/terminal/rich_form.go)) | Change model/key types for Bubble Tea v2. Evaluate Huh v2 select/multi-select against description rendering and PTY cases before deleting the custom list. Apply one repository theme through `WithTheme`. |
| Rich result rendering | Two `lipgloss.NewRenderer` call sites bind styles to writers ([`presentation.go`](../../../internal/terminal/presentation.go), [`rich.go`](../../../internal/terminal/rich.go)) | Remove renderer fields/parameters. Bubble Tea handles Live View color conversion; standalone Result/Transcript writers need explicit per-stream profile/no-color handling. |
| Plain/Automation | Both use terminal-control-free line output; Automation never prompts | Do not replace this with Huh accessible mode or `tea.WithoutRenderer`. Keep the semantic non-TUI renderer and existing capability classification. |
| Diagnostics | Exact text/NDJSON plus recursive redaction behind `internal/logging` | Any Log v2 use stays private behind this facade. JSON and redaction are acceptance constraints, not formatter preferences. |

## Evidence required during implementation

The library APIs remove uncertainty about feasibility, but not terminal behavior
in this repository. The v2 implementation should preserve or add focused tests
for:

- redirected stdin/stdout/stderr and `TERM=dumb` behavior;
- PTY entry/exit sequences, cursor restoration, and no remaining AltScreen;
- semantic transcript presence after full-screen exit and absence in Automation;
- secret answers never entering transcript or diagnostic output;
- large/select/multi-select filtering, descriptions, resize, and cancellation;
- exact existing NDJSON bytes and recursive redaction;
- stdout containing exactly one complete Command Result;
- diagnostic ordering around the renderer lease and transcript replay.

These are repository acceptance tests, not responsibilities that Bubble Tea,
Huh, Lip Gloss, or Log can satisfy automatically.
