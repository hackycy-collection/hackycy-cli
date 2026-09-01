# Choose the fs Lifecycle Log presentation

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What line-oriented startup, address, capability, authentication, task,
warning, failure, and shutdown Lifecycle Log should the `fs` Service Command
emit at each log level and in text versus NDJSON modes, without introducing a
custom full-screen UI or high-volume request logging?

## Answer

`fs` remains a foreground Service Command with no stdin, form, Live View,
AltScreen, Work Phase, or Interaction Transcript. It keeps `fs [directory]`,
all flags/environment precedence/defaults, browser and HTTP behavior, stdout
startup and stopped documents, normal signal success, errors, and exit codes.
It adds a scoped `fs` Lifecycle Log on stderr through the repository logging
Runtime.

### Durable stdout checkpoints

The existing startup document remains byte- and order-compatible in Plain,
Automation, and redirected output:

```text
File Browser
Local: <local URL>
Network: <network URL>
Directory: <resolved directory>
Bind: <address>:<port>
Management: <true|false>
Chunked uploads: <true|false>
HTML execution: <true|false>
Authentication: <true|false>
Session storage: <resolved session directory>
```

`Network:` remains repeated only for discovered public IPv4 addresses and
`Session storage:` remains present only when authentication is enabled. No
account count, upload chunk size, status symbol, timestamp, or new field is
added. Normal successful shutdown retains the separate exact stdout line:

```text
File Browser stopped.
```

These two time-sensitive writes are Service Command result checkpoints, not a
finite command using `Finish` twice. `internal/terminal` provides the narrow
`ResultCheckpoint(id, document)` operation on the existing `ExperienceRun`
Interface, with stable IDs `fs-startup` and `fs-stopped`. It does not add a
parallel Service presenter Interface. The terminal implementation owns
injected-stdout routing, styling/control removal, checkpoint state, and write
errors. Each ID writes immediately at most once, never starts AltScreen, never
enters a Transcript, and rejects a duplicate without retrying. This exception
preserves an existing long-running stdout contract; finite commands still
submit one complete Command Result.

`--log-format=json` never changes either stdout checkpoint. Lifecycle records
never move into stdout.

### Startup and capabilities

The command obtains `factory.Logging.Logger("fs")` before service work. It does
not emit a speculative starting record. Argument/environment parsing,
workspace opening, authentication/session setup, thumbnail pool creation,
handler construction, listener binding, or network-interface discovery
failure returns through the existing single root error path with no duplicate
Lifecycle Error.

After the full Runtime has started, Info records have this fixed order:

1. `File Browser started` with `localURL`, ordered `networkURLs`,
   `bindingAddress`, and actual `port`.
2. `Browse root configured` with resolved `directory`.
3. `File Browser capabilities configured` with `managementEnabled`,
   `chunkedUploadsEnabled`, and `htmlExecutionEnabled`; it includes
   `uploadChunkSizeBytes` only when chunked uploads are enabled.
4. `File Browser authentication configured` with `authenticationEnabled`;
   when enabled it also includes `accountCount`, resolved `sessionDirectory`,
   and non-negative integer `sessionIdleDurationMs`.
5. When applicable, the unauthenticated-public warning defined below.
6. The `fs-startup` stdout checkpoint.

Lifecycle observers are installed before the listener can publish task state.
If a client reaches the bound listener before the startup checkpoint finishes,
the service-local sequencer buffers its task events and releases them only
after `fs-startup` succeeds. A failed startup checkpoint discards those task
records into the shutdown summary instead of presenting task work before a
usable service endpoint.

The log uses positive, user-facing capability names rather than internal
`SafeHTML` polarity. It does not record flag/environment provenance, route
lists, service object details, usernames, account specifications, passwords,
hashes, salts, credential revisions, session tokens, cookies, or login data.

If the binding address is not loopback and authentication is disabled, Warn
`File Browser is accessible without authentication` is emitted after the
configuration records and before stdout. It includes only `bindingAddress` and
`managementEnabled`; management-enabled text receives stronger visual
emphasis but remains the same single record. The warning does not prevent
startup or change exit status.

### Managed Task scope

Only download, archive extraction, and chunked upload are FS Managed Tasks:
they have an independent identity and can span requests. Ordinary upload,
text edit, copy/move/delete/create operations, directory/file reads,
thumbnail conversion/cache work, authentication/session actions, SSE
connections, and HTTP requests are not Lifecycle Log tasks.

Every task record includes stable `taskType=download|extraction|chunkedUpload`
and full `taskID`. Text may additionally show a fixed-length ID prefix, while
NDJSON retains the full ID. Task IDs are not credentials, but still pass
through control removal and field bounds. Chunk owner and session identity are
always excluded.

Download has the fixed event sequence:

- Info `Download task accepted` for a new queued task.
- Info `Download task started` on the real queued-to-running transition.
- Exactly one Info `Download task completed` or `Download task cancelled`, or
  Warn `Download task failed`.

Download context never contains the original URL. It may include only parsed
`sourceScheme` and `sourceHost`, with userinfo, URL path, query, and fragment
removed, plus bounded workspace-relative `destinationPath` and final
`filename`. Completion includes `bytesDownloaded`, available `totalBytes`, and
non-negative integer `durationMs`. Cancellation includes `cancelSource=client`
and the last safe byte counts. Failure includes a stable public `code` and
safe public `error`, never redirect history, response headers, or an internal
cause.

Extraction has the fixed event sequence:

- Info `Extraction task accepted`.
- Info `Extraction task started`.
- Exactly one Info `Extraction task completed` or
  `Extraction task cancelled`, or Warn `Extraction task failed`.

Extraction records may include bounded workspace-relative `archivePath`.
Completion adds `destinationPath`, `entryCount`, `uncompressedBytes`, and
`durationMs`; cancellation includes `cancelSource=client` and last safe
statistics. Failure uses public `code` and safe public `error`, never archive
entry names, file contents, 7-Zip stdout/stderr, absolute staging paths, or
internal causes.

A download or extraction retry creates a new task with a new `taskID`, normal
accepted/started/terminal records, and `retryOf` pointing to the old task ID.
The old terminal event is immutable. Invalid retry, duplicate cancel, cancel
of a terminal task, clear-history, not-found, validation, and ordinary HTTP
errors remain silent in the Lifecycle Log.

Chunked upload has only real model states:

- Info `Chunked upload started` after Create succeeds, with safe `filename`,
  workspace-relative destination directory, `totalBytes`, and
  `chunkSizeBytes`.
- Info `Chunked upload completed` after atomic publication, with final safe
  `destinationPath`, `totalBytes`, and `durationMs`.
- Info `Chunked upload cancelled` after the first explicit cancellation, with
  `cancelSource=client` and last `uploadedBytes`.
- Info `Chunked upload expired` when existing prune behavior observes and
  removes an expired incomplete upload.

Append/Get and every chunk/progress update are silent. No timer is added to
change expiry behavior. The current model has no terminal error state:
Create/Append/Complete failures remain retryable HTTP errors and do not
fabricate `Chunked upload failed`. A future terminal error state requires a
future presentation decision.

### Volume, safety, and levels

Task accepted, started, completed, cancelled, failed, and expired states are
published even when transitions happen immediately; no artificial delay is
introduced. Download's 250ms updates, extraction percentages/inspection
callbacks, chunks, requests, methods/paths/statuses, client addresses,
thumbnail activity, and login/session activity are never logged, including at
Debug. Each task therefore has a fixed maximum number of records.

Task failure is Warn because the service remains available. The fixed Warn
messages are `Download task failed` and `Extraction task failed`; they expose
only stable public ServiceError `code` and bounded public `error`. Error is
reserved for a failure that prevents the File Browser service from continuing
or completing shutdown.

Debug currently adds no task progress events; it includes the same semantic
records admitted by lower-severity filtering. Info contains startup,
successful/cancelled task lifecycle, stopping, and stopped records. Warn shows
unauthenticated-public and task-failure warnings plus Error. Error shows only
service failure. Filtering remains owned by the shared logging Runtime.

All messages and nested context values pass through recursive credential
redaction, UTF-8 normalization, control/ANSI removal, single-line projection,
and fixed field-length bounds. Full resolved browse/session directories and
bounded workspace-relative task paths remain visible because they are
explicit operator-facing context; file contents and hidden transport data do
not.

Authentication success/failure, resume, logout, expiry, and session-store HTTP
errors are never request-logged. Authentication initialization or close errors
appear only when they cause startup or service shutdown failure.

### Shutdown and failure

Context cancellation, including normal SIGINT/SIGTERM, first stops accepting
new tasks and emits Info `File Browser stopping` with
`reason=context-cancelled` and a snapshot of `queuedDownloads`,
`activeDownloads`, `queuedExtractions`, `activeExtractions`, and
`incompleteChunkedUploads`. Manager shutdown cancels/removes outstanding work
and waits for resource release. These bulk cancellations do not emit one line
per task.

Clean shutdown then emits Info `File Browser stopped` with
`cancelledDownloads`, `cancelledExtractions`, and `removedChunkedUploads`,
followed by the one `fs-stopped` stdout checkpoint. A spontaneous clean server
stop uses `reason=server-stopped` and the same stopped/checkpoint order. Normal
shutdown preserves success and writes no other stdout.

Once stopping begins, late task callbacks update internal cleanup counts only;
they cannot emit new task events. `File Browser stopped` or
`File Browser failed` is the last Lifecycle record.

An unexpected serve, close, or Runtime resource-release failure emits one
Error `File Browser failed` with `stage=serve|close|release` and a bounded safe
`error`. Multiple causes remain joined internally but project to one record.
This replaces `File Browser stopped` and suppresses the stopped stdout
checkpoint. Root recognizes the already-reported error without duplicating it,
while the process Outcome retains the original error chain and exit code 1.

If the startup checkpoint write fails after binding, the command emits
`File Browser stopping`, closes and waits for all resources, then emits
`File Browser stopped` with `reason=startup-output-failed` only when cleanup
succeeds; it does not emit the stopped stdout checkpoint. If close/release also
fails, `File Browser failed` replaces stopped and safely projects the joined
write/cleanup error. Otherwise the original write error is returned once
through root. If the final stopped checkpoint write fails, the real service
outcome remains stopped and the write error is returned without restarting,
retrying, or copying result text to stderr. Cleanup calls and checkpoints are
at most once. A context cancelled before Runtime startup produces no stdout or
Lifecycle Log.

Logging-writer failures remain best effort: they do not stop the server,
change stdout or exit behavior, deadlock, panic, or recursively log another
failure.

### Text, NDJSON, and evidence

Text records remain one physical line with timestamp, level, `[fs]`, message,
and structured context. The vivid text Adapter uses `●` for started/accepted,
`✓` for completed/stopped, `⊘` for cancelled/expired, `!` for Warn, and `✕`
for Error. Color enhances symbols and levels only; redirected or disabled-
color output is ANSI-free and retains words/symbols. There are no boxes,
spinners, dashboards, animation rows, or multi-line records.

JSON retains the exact shared NDJSON envelope
`{timestamp, level, scope, message, context}` with `scope="fs"`; all FS fields
remain inside `context`, every event is one atomic JSON line, and no event
contains ANSI. Rich, Plain Interactive, and Automation share event semantics;
terminal mode changes only text coloring and never starts AltScreen.

A service-local sequencer orders manager callbacks, HTTP task creation,
server completion, result checkpoints, and shutdown. Each accepted download or
extraction has one accepted, at most one started, and exactly one explicit
terminal event unless service shutdown folds it into the service summary.
Each chunked upload has one started and at most one observed terminal event.
No task event follows stopping, and each logging/checkpoint write is atomic or
at most once according to its Interface contract.

Acceptance uses an injected clock, recording/failing logger and stdout writers,
controllable managers, and real temporary listeners. It covers flag/environment
precedence; private/public/port-zero startup; the four startup records, public-
without-auth warning, and startup stdout ordering/bytes; every capability and
authentication combination with zero username/password/token leakage; fast,
queued, retried, completed, explicitly cancelled, failed, and expired Managed
Tasks; safe URL/path/stat projections; absence of progress/chunk/request/auth/
thumbnail logs; shutdown folding/counts and concurrent exactly-once ordering;
serve/close/release/checkpoint/logger failures; context-before-start; all log
levels; text and NDJSON goldens; control stripping/redaction/field bounds; no
AltScreen; and unchanged arguments, HTTP behavior, stdout documents, errors,
and exit contracts.
