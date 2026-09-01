# Choose the diff Lifecycle Log presentation

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What line-oriented startup, address, baseline/target, refresh, warning,
failure, and shutdown Lifecycle Log should the `diff` Service Command emit at
each log level and in text versus NDJSON modes, without introducing a custom
full-screen UI or high-volume request logging?

## Answer

`diff` remains a foreground Service Command with no stdin, form, Live View,
AltScreen, Work Phase, or Interaction Transcript. It keeps its command name,
two directory arguments, flags/defaults, fixed Comparison Workspace, browser
and MCP behavior, asynchronous Refresh semantics, stdout startup document,
normal signal success, errors, and exit codes. It adds a scoped `diff`
Lifecycle Log on stderr through the repository logging Runtime.

### Streams and startup

The existing Command Result remains one stdout document, emitted after the
listener is bound and the initial Comparison Refresh Attempt is accepted:

```text
Directory diff: <local URL>
MCP endpoint:   <local URL>/mcp
Baseline: <resolved baseline directory>
Target:   <resolved target directory>
```

In public mode, the existing ordered `Network:` / `Network MCP:` pairs remain
between the MCP and Baseline lines. All fields, ordering, redirection behavior,
and emission time are unchanged. Lifecycle records never move into stdout, and
`--log-format=json` does not convert the startup document to JSON. It uses the
Service Command result checkpoint operation with stable ID `diff-startup`,
submitted at most once; it does not claim that the long-running service has
already reached its terminal outcome.

The command obtains `factory.Logging.Logger("diff")` before service work. It
does not emit a speculative `Starting` record. Workspace validation, embedded
asset loading, listener binding, network-interface discovery, or initial-
Refresh scheduling failures return through the existing root error path with
no duplicate Lifecycle Error.

Once all startup prerequisites succeed, Info records have this fixed order:

1. `Directory diff started` with `localURL`, `public`, and actual `port`.
2. `Diff endpoints available` with `mcpURL` and ordered `networkURLs`.
3. `Comparison workspace configured` with resolved `baselineDirectory` and
   `targetDirectory`.
4. `Initial comparison refresh started` with `refresh=1` and
   `source=initial`.
5. The unchanged startup Command Result is submitted once to stdout.

Lifecycle observation is installed before the initial Refresh starts. Phase
and terminal events are sequenced after the stdout submission, so even an
immediate Refresh cannot lose or reorder its semantic lifecycle. This is a
call-order guarantee across the two writers, not a promise about how a shell
interleaves independently redirected stdout and stderr bytes.

### Refresh Attempts

Every accepted Comparison Refresh Attempt receives a service-local,
monotonically increasing `refresh` ordinal. The initial attempt is 1; accepted
REST and MCP attempts use `source=rest` and `source=mcp`. Rejected concurrent
attempts, invalid HTTP methods/origins, no-op cancellations, SSE connections,
and ordinary HTTP/MCP requests produce their existing protocol responses but
no Lifecycle record.

An accepted later attempt emits Info `Comparison refresh started`. Each
attempt then has exactly one terminal event:

- Info `Comparison snapshot ready` includes `refresh`, `source`, non-negative
  integer `durationMs`, `snapshotID`, `added`, `deleted`, `modified`,
  `unchanged`, `issues`, `totalEntries`, and available `comparedBytes`.
- A ready Snapshot with `issues > 0` remains successful. Its ready record is
  immediately followed by Warn `Comparison snapshot contains issues` with
  `refresh`, `snapshotID`, and `issues`; individual issue paths/messages remain
  in the browser/API.
- Info `Comparison refresh cancelled` includes the original `refresh` and
  start `source`, `durationMs`, and `cancelSource`. REST cancellation uses
  `cancelSource=rest`. Cancellation retains any previously published Snapshot.
- Error `Comparison refresh failed` includes `refresh`, `source`,
  `durationMs`, and a safe `error`. The server remains available and later
  attempts can succeed; no failure Command Result is written to stdout and the
  process exit status does not change.

Cancelled and failed attempts include their last available bounded progress
fields and `hasPreviousSnapshot`; `previousSnapshotID` is present only when a
published Snapshot exists. They never expose an unpublished Snapshot ID or
partial comparison as a published result.

Debug adds one `Comparison refresh phase` record only on the first transition
into each of `discovering`, `comparing`, and `publishing`. It carries
`refresh`, `source`, `phase`, and whichever of `discoveredEntries`,
`comparedEntries`, `totalEntries`, `comparedBytes`, `totalBytes`, and `issues`
are known at that transition. It does not sample by time or record subsequent
progress ticks.

### Shutdown and failure

Context cancellation, including normal SIGINT/SIGTERM handling, emits Info
`Directory diff stopping` with `reason=context-cancelled`, stops accepting new
Refresh Attempts, cancels and awaits active work, closes the server, then emits
Info `Directory diff stopped`. An active Refresh cancellation is folded into
this sequence rather than logged separately. Normal cancellation remains a
successful command outcome and appends nothing to stdout.

An unexpected server failure or close failure emits one Error
`Directory diff failed` with `stage=serve|close` and a safe `error`; it is the
last Lifecycle record. The returned error retains its original chain and exit
code 1. Root recognizes that the command has already reported this error and
does not duplicate it, while preserving the error in the process Outcome.
A spontaneous clean server stop records `Directory diff stopped` and preserves
the existing successful return.

If the startup stdout write fails after binding, the command emits
`Directory diff stopping` and `Directory diff stopped` with
`reason=startup-output-failed`, closes and waits for the server, and returns
the original write error. It never retries or copies the result to stderr. A
simultaneous close error is joined to the write error; this branch still uses
one root error rather than a second Lifecycle Error. Logging-writer failures
remain best effort: they do not stop the service, alter stdout/exit behavior,
deadlock, panic, or recursively log another failure.

### Levels, fields, and rendering

The stable message catalog is:

- `Directory diff started`
- `Diff endpoints available`
- `Comparison workspace configured`
- `Initial comparison refresh started`
- `Comparison refresh started`
- `Comparison refresh phase`
- `Comparison snapshot ready`
- `Comparison snapshot contains issues`
- `Comparison refresh cancelled`
- `Comparison refresh failed`
- `Directory diff stopping`
- `Directory diff stopped`
- `Directory diff failed`

Debug shows all records; Info suppresses only phase records; Warn shows issue
warnings and errors; Error shows errors only. Filtering, timestamps, scope,
serialization, redaction, injected writer selection, and format configuration
remain owned by `internal/logging`.

Context keys use lower camel case. Unknown optional values are absent rather
than represented by false zeroes or `null`; a private-mode `networkURLs` is an
explicit empty ordered array. Resolved full directory paths remain visible as
they are in the existing startup result, but all messages and nested context
values pass through recursive credential redaction, UTF-8 normalization,
control/ANSI removal, single-line projection, and fixed field-length bounds.
Logs never contain request/response bodies, headers, cookies, query strings,
file contents, symlink targets, per-entry paths, or network-client addresses.

Text records remain one physical line with timestamp, level, `[diff]`, message,
and structured context. The vivid text Adapter uses `●` for started, `✓` for
ready/stopped, `⊘` for cancelled, `!` for Warn, `✕` for Error, and `·` for
Debug phase. Color enhances the symbol and level only; redirected or disabled-
color output is ANSI-free and retains the symbol and words. There are no boxes,
spinners, animation rows, or multi-line records.

JSON continues to use the exact shared NDJSON envelope
`{timestamp, level, scope, message, context}` with `scope="diff"`; all
command-specific fields remain inside `context`. Each event is one atomic JSON
line with no ANSI. Rich, Plain Interactive, and Automation share the same
events; terminal capability changes only text coloring and never starts
AltScreen.

### Ordering and evidence

One service-local sequencer orders concurrent Workspace callbacks, REST/MCP
goroutines, server completion, and shutdown. For each accepted attempt it
guarantees one started event, at most one first-entry event per Debug phase,
and exactly one terminal event. A ready warning is adjacent to its ready event;
no Refresh start can appear after service stopping; `Directory diff stopped`
or `Directory diff failed` is the last Lifecycle record. The logging Runtime
continues to serialize each record as one atomic write.

Acceptance uses an injected clock, recording/failing writers, controllable
Workspace Refreshes, and real temporary listeners. It covers private/public
and port-zero startup; the three startup records, initial Refresh record, and
stdout ordering; immediate completion; phase deduplication; count/byte/duration
fields; issue warnings; REST/MCP sources and ordinals; silent active rejection;
explicit cancellation; shutdown cancellation folding; previous-Snapshot
retention; failure followed by successful Refresh; unexpected serve/close and
stdout/log-writer failure; context cancellation before startup; redaction,
control stripping, and field bounds; all log levels; text and NDJSON goldens;
concurrent exactly-once ordering; absence of request/per-entry logs; no
AltScreen/control sequences; and unchanged arguments, flags, startup result,
service behavior, errors, and exit contracts.
