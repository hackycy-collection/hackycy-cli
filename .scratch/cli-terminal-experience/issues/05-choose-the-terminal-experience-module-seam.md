# Choose the terminal experience Module seam

Type: grilling
Status: resolved
Blocked by: 01, 02, 04

## Question

Where should the ycy terminal-experience Module live, and what is its smallest
deep Interface? Decide how it owns capability negotiation, input lifecycle,
style selection, transient-render cleanup, and adaptation of the existing
command-owned Prompter and Presenter interfaces without importing
Charmbracelet into `internal/commands` or turning `cmd/ycy` into a collection
of duplicated wrappers.

Compare concrete alternatives including command-local Charmbracelet adapters,
one shared terminal-experience Module with command-specific adapters, and a
full UI runtime. State which dependencies are injected, which stay private,
how tests cross the seam, and how normal stream output and child-process I/O
remain outside a transient renderer when required.

## Answer

Approved on 2026-08-26. ycy will use one shared, deep Terminal Experience
Module at `internal/terminal`. It owns terminal capability negotiation,
session-specific rendering behavior, Charmbracelet lifecycle and cleanup, and
the coordination required to keep diagnostics from corrupting a Transient
View. It does not own command behavior, command results, Cobra, or child
process execution.

### Seam and interface

The composition root constructs the Terminal Experience Module once from the
inherited standard streams and environment lookup. It classifies the immutable
Session before any Huh or Bubble Tea object is constructed. The externally
meaningful Interface is intentionally small:

- an `Experience` opens an `Experience Run` for one command Context;
- an `Experience Run` performs `Ask(Interaction Request)`,
  `Present(Presentation Document)`, and `Track(Tracked Operation)`; and
- every Experience Run is closed by its caller, normally through `defer`, so
  the Module can restore terminal state on every controlled return path.

`Ask` and `Track` acquire a Renderer Lease over the diagnostic stream.
`Present` writes the durable Command Result to stdout. The exact Go spelling
of these types is an implementation concern; the semantic roles and the three
intent-level operations are the stable Interface.

The Module also provides a lease-aware Diagnostic Writer to the composition
root. That writer is not part of a command Adapter's interaction Interface.
It is supplied to the logging Runtime and normal Cobra diagnostic path so that
diagnostics serialize normally outside a lease and defer, in order, until the
Transient View has been cleaned up during a lease. The later output-contract
decision owns log format, level, quiet, verbose, and JSON semantics.

### Ownership and package placement

`internal/terminal` privately owns Huh, Bubble Tea, Lip Gloss, palette
selection, width handling, raw-mode/cursor/screen cleanup, and the concrete
renderer implementation. Its construction inputs are only inherited process
streams and environment lookup; a pure Session-fact classifier remains
directly testable without the process terminal. The root command Context is
passed when an Experience Run begins.

Command-specific Terminal Adapters remain in `cmd/ycy`. Each is a thin
translation from an existing command-owned Prompter or Presenter interface to
the generic Interaction Request, Presentation Document, or Tracked Operation.
Those Adapters must not import Charmbracelet, inspect raw streams, classify a
Session, or reproduce cleanup logic. Business Modules in `internal/commands`
retain their semantic interfaces and must not import either Charmbracelet or
the Terminal Experience Module.

This rejects command-local Charmbracelet adapters because they duplicate
capability and lifecycle behavior, and rejects a full UI runtime because it
would force Cobra, child I/O, service logs, and unrelated command semantics
into one broad model.

### Cancellation, diagnostics, and direct I/O

`cmd/ycy` remains the single OS-signal owner. An Experience Run receives a
Context derived from that root Context; Huh and Bubble Tea wrappers disable
their own signal handlers. Esc and Ctrl-C map to each command's current
cancellation outcome, while tracked-work Ctrl-C cancels the same command
Context. Cleanup restores cursor, input, and screen state before control
returns to Cobra.

Tunnel service logs never acquire a Renderer Lease and remain a continuous
stderr diagnostic stream. A transient selection view finishes before Tunnel
service logging begins. Foreground child processes, including `run`, retain
their original unwrapped stdin, stdout, and stderr as Raw Child I/O; no
Transient View may coexist with that direct stream ownership. Browser-server
readiness and other durable stream contracts likewise remain outside a
Transient View unless a later command-specific decision explicitly changes
only its human-facing presentation.

### Test seam

The implementation will prove this Module through three layers:

1. pure Session classification tests with injected terminal facts and
   environment;
2. command Adapter tests using a recording Experience/Experience Run fake to
   assert semantic translation without a real terminal; and
3. a limited PTY/subprocess suite for rich rendering, Plain Interactive
   behavior, Automation behavior, Ctrl-C, and terminal restoration.

This decision adds the canonical terminology to `CONTEXT.md`: Terminal
Experience Module, Terminal Adapter, Experience Run, Interaction Request,
Presentation Document, Tracked Operation, Renderer Lease, Lease-Aware
Diagnostic Writer, and Raw Child I/O.
