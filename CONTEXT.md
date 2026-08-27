# ycy CLI Experience

The ycy CLI Experience is the human-facing terminal behavior of ycy while it
preserves reliable use from scripts, pipes, and CI.

## Language

**Interactive Session**:
An invocation classified as either a Rich Interactive Session or a Plain
Interactive Session. It is a capability state of one invocation, not a command
category.
_Avoid_: interactive command, TUI mode

**Rich Interactive Session**:
An invocation whose standard input, output, and error are terminals, which is
not running under `CI`, and whose terminal supports rich rendering. ycy may
ask questions and render Transient Views.
_Avoid_: TUI mode, color mode

**Plain Interactive Session**:
An invocation whose standard streams are terminals but whose terminal cannot
safely render rich terminal controls. ycy may use simple line prompts but not
ANSI styling, cursor control, animation, or an alternate screen.
_Avoid_: fallback mode, degraded TUI

**Automation Session**:
An invocation with a redirected standard stream or a set `CI` environment
variable. It receives stable stream-oriented behavior without cursor control
or unsolicited questions.
_Avoid_: non-interactive mode, headless mode

**Automation Input**:
A command-owned explicit value accepted in an Automation Session in place of
an Interactive Session prompt. It is never inferred from stdin.
_Avoid_: piped answer, non-interactive input

**Prompt-Dependent Path**:
A command path that cannot complete under its existing grammar without an
Interaction Request. In an Automation Session it fails before a side effect;
ycy does not invent a default or substitute an answer.
_Avoid_: interactive command, blocked command

**Command Result**:
The durable output a command intentionally returns to its caller or user.
Command Results belong on standard output and may be machine-consumed.
_Avoid_: log, status message

**Diagnostic Event**:
A redacted operational observation intended to explain progress, warnings, or
failures to a human operator. Diagnostic Events belong on standard error.
_Avoid_: command output, result

**Diagnostic Record**:
One redacted, timestamped rendering of a Diagnostic Event. It is emitted as
either one human-readable text line or one NDJSON object, never as a Command
Result.
_Avoid_: log string, JSON error

**Diagnostic Configuration**:
The process-wide selection of diagnostic threshold and record format. It never
changes Command Results, User-Actionable Errors, or panic-stack policy.
_Avoid_: output mode, command option

**User-Actionable Error**:
A single plain `error:` message on standard error that tells a caller what to
correct and preserves the command's established exit semantics. It is not a
structured Diagnostic Event, even when JSON diagnostic logging is selected.
_Avoid_: error log, JSON error

**Transient View**:
A terminal rendering that may be replaced in place, such as a form, spinner,
progress display, or live status view. It is available only in an Interactive
Session.
_Avoid_: output, log line

**Terminal Experience Module**:
The shared Module that determines an invocation's Session and consistently
provides terminal interaction, styling, and transient-render lifecycle without
owning command behavior.
_Avoid_: TUI runtime, UI framework

**Terminal Adapter**:
A command-specific Adapter that translates a command's semantic Prompt or
Presenter interface into the Terminal Experience Module.
_Avoid_: terminal implementation, command Module

**Experience Run**:
The per-invocation handle through which a Terminal Adapter performs a fixed
set of terminal interactions and presentation operations. It does not own the
command's business behavior or result semantics.
_Avoid_: Run, command execution

**Interaction Request**:
A typed request from a Terminal Adapter for a user answer, such as a choice,
text value, secret, or confirmation. It is command intent, not a particular
terminal control.
_Avoid_: Huh field, prompt implementation

**Presentation Document**:
A typed durable terminal presentation assembled from semantic visual roles.
It may render as a rich Command Result or as stable plain text according to
the Session.
_Avoid_: ANSI string, report template

**Command Discovery Document**:
A stable description of a command's grammar, available descendants, flags,
and safe examples. Cobra defines its content; terminal presentation may change
only its human-facing Rich Interactive rendering.
_Avoid_: help template, command screen

**Tracked Operation**:
A command operation whose human-facing state is updated while work is in
progress. It may use a Transient View only in a Rich Interactive Session.
_Avoid_: spinner, Bubble Tea program

**Operation Phase**:
A command-owned, externally meaningful unit of work within a Tracked
Operation. A Terminal Adapter translates Operation Phases for presentation;
the terminal must not infer them from logs, goroutines, or output text.
_Avoid_: UI step, spinner state

**Renderer Lease**:
Exclusive temporary ownership of the terminal diagnostic stream by an
Experience Run while it renders a Transient View. Direct diagnostic writes
must be presented through, or deferred until after, that lease.
_Avoid_: output lock, logging mode

**Lease-Aware Diagnostic Writer**:
The diagnostic-stream Adapter supplied by the Terminal Experience Module. It
serializes Diagnostic Events and defers them while a Renderer Lease is active.
_Avoid_: terminal logger, UI log sink

**Raw Child I/O**:
The unwrapped inherited standard streams used by a foreground child process.
It remains outside a Transient View and is not a Command Result or a
Diagnostic Event.
_Avoid_: terminal passthrough, renderer output
