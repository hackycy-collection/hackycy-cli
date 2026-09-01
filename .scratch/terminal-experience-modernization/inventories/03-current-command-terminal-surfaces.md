# Current command terminal surfaces

Date: 2026-08-31

This inventory is the current-state evidence for
[Inventory every current command terminal surface](../issues/03-inventory-every-command-terminal-surface.md).
It covers the root/process surface and all 22 registered leaf commands. It
describes behavior; it does not propose the redesigned presentation.

## Reading the inventory

- **stdin** means the inherited terminal input read by a ycy interaction or a
  handed-off child process.
- **stdout / Command Result** means durable output written through
  `ExperienceRun.Result`, Cobra discovery/completion, the version path, or a
  raw child stdout stream.
- **stderr / Live View** means Rich alternate-screen rendering or Plain
  Interactive prompts, notices, and phases.
- **stderr / Diagnostic Record** means root errors or diagnostic logger
  records.
- **stderr / Lifecycle Log** means the long-running Tunnel logger. Diff and FS
  do not currently use Lifecycle Logs even though the map classifies them as
  Service Commands.
- **lost** means visible in the alternate screen but absent from normal
  terminal history after that screen is restored. A PTY byte capture can still
  contain those bytes; that is not an Interaction Transcript.

Unless a command says otherwise, a returned error becomes one
`error: <message>` Diagnostic Record on stderr and exit code 1. Cobra argument
and flag validation happens before command work. Command Results emitted before
an error remain on stdout.

## Shared terminal runtime

The shared behavior is implemented by
[`internal/terminal`](../../../internal/terminal/experience.go) and classified
in [`session.go`](../../../internal/terminal/session.go).

| Concern | Current behavior |
| --- | --- |
| Mode selection | Automation is selected when stdin or stderr is not a TTY, or `CI` is set. Rich Interactive additionally requires a recognized `TERM`; otherwise the mode is Plain Interactive. Stdout does not decide interaction mode. |
| Rich availability | Rich presentation also requires stdin and stderr to be real `*os.File` terminals with a readable size. Failure falls back to Plain Interactive for that run. |
| AltScreen lifecycle | Rich starts lazily on the first `Ask`, `Notice`, or `Track`, uses one Bubble Tea v1 alternate-screen program on stderr, and stops on the first `Result` or `Close`. A command that only calls `Result` never opens AltScreen. |
| Forms | Text, secret, and confirm use Huh v1. Select and multiselect use the repository's custom ASCII `richListForm` with `>`, `[ ]`, `[x]`, `/` filtering, and a textual key-help footer. |
| Plain interactions | Questions, choices, validation retries, and notices are line-oriented stderr. Answers are read from stdin. Secrets require a real TTY and use `term.ReadPassword`. |
| Automation interactions | `Ask` returns `ErrAutomationInteraction` without reading stdin or writing a prompt. `Notice` writes nothing. `Track` consumes all updates and requests cancellation when the context ends, but renders nothing. |
| Work phases | Only pending, active, completed, cancelled, and failed exist. Plain prints every phase update to stderr. Rich replaces phases by name in the Live View and preserves only a final phase inside that same alternate-screen model. There is no skipped state or transcript ledger. |
| Command Result | `Result` first restores the main screen, then writes stdout. A TTY stdout receives styled/wrapped output; redirected stdout receives plain, control-free output. The first result forbids later `Ask`/`Notice`/`Track`, but later `Result` calls are allowed. |
| Diagnostics during Rich | The renderer lease defers concurrent diagnostics and flushes them to stderr after AltScreen closes. This protects rendering order but does not replay Live View content. |
| Redirection | Redirecting stdout alone preserves Rich interaction on stderr. Redirecting stdin or stderr selects Automation. Results never move to stderr just because stdout is redirected. |
| Color | Semantic result roles map to cyan title/active, green success, yellow warning, red error, and faint gray muted text. Plain and redirected rendering strip control sequences. |
| Current dependencies | Bubble Tea `v1.3.10`, Huh `v1.0.0`, Lip Gloss `v1.1.0`, and an indirect Bubbles pre-release are installed. `charm.land/log/v2` is not installed. |

The central loss mechanism is structural: the Rich model retains notices and
final tracked phases only in AltScreen memory. `stopRich` restores the main
screen and discards the model. No completed question, answer, notice, or phase
is projected to durable stderr.

## Root and process surface

The process chain is
[`internal/ycycmd`](../../../internal/ycycmd/main.go) ->
[`pkg/cmd/root`](../../../pkg/cmd/root/app.go).

| Path | stdin | stdout | stderr | Exit and notable behavior |
| --- | --- | --- | --- | --- |
| No arguments | none | custom discovery/help Command Result | none | Exit 1, unlike explicit help. |
| `--help` and command help | none | custom semantic discovery document | none | Exit 0. Parent commands invoked without a leaf also print help and exit 1. |
| `--version` / `-V` | none | exact version plus newline | none | Exit 0; skips startup-update consumption and diagnostic configuration. |
| `completion <shell>` | none | raw Cobra completion script | none | Exit 0; bypasses diagnostic configuration. |
| Parser/argument failure | none | none | one actionable `error:` line, sometimes with a single suggestion and help path | Exit 1; no automatic usage dump. |
| Command failure | command-dependent | any already-emitted results remain | one root `error:` line | Exit 1 unless the error owns an exit code. Ordinary root reporting does not itself call the redactor, so leaf errors must already be safe. |
| Exit-coded child/result | command-dependent | raw/result output | no automatic root error | Root returns the owned exit code. This is how `run` preserves child failures and Upgrade avoids a duplicate error. |
| Panic | none | one blank line | redacted `error:` and optional stack when debug is enabled | Exit 1. |
| Startup update state | none | completed update result before the requested command | pending state and then an `error:` line | Skipped for version and private routes. A prior result can therefore prefix unrelated command stdout. |
| Private updater | private arguments | none | `error:` on failure | Runs before Cobra. |
| Thumbnail worker | framed stdin protocol | framed stdout protocol | `thumbnail worker error:` on failure | Raw private machine channel; never terminal decoration. |
| Web/root construction failure | none | one blank line | `error:` | Exit 1 before Cobra. |

Global `--log-level`, `--log-format`, `--verbose`, and `--quiet` are parsed
before leaf effects. The logging facade writes injected, redacted text or
stable NDJSON to stderr. Help, version, completion, and parent discovery bypass
that configuration. A logger error followed by a returned error is not
deduplicated by root.

Evidence: root discovery, parser, stream, panic, and command-surface tests in
[`pkg/cmd/root`](../../../pkg/cmd/root); process and startup-update tests in
[`internal/ycycmd`](../../../internal/ycycmd); standalone private-route evidence
in [`acceptance/process_root_test.go`](../../../acceptance/process_root_test.go).

## Leaf summary

| Leaf | Current interactive surface | Durable stdout result | Current progress/log surface | Automation behavior |
| --- | --- | --- | --- | --- |
| `export env` | optional environment select | heading plus JSON, file-target message, or cancellation | none | works when selection is unambiguous; otherwise errors |
| `config fork list` | none | title/table/count or empty guidance | none | works |
| `config fork add` | five-field form | success or cancellation | none | rejected before config access |
| `config fork remove` | select plus confirm | success, cancellation, or empty outcome | empty notice only | empty works; nonempty errors at first question |
| `config cm list` | none | title/list or empty guidance | none | works |
| `config cm add` | four-field form | success or cancellation | none | rejected before config access |
| `config cm use` | none | success | none | works |
| `config cm set` | none | success | none | works |
| `config cm remove` | confirm | success or cancellation | none | missing profile errors first; existing profile requires interaction |
| `config cm test` | none | provider response or safe failure context | none | works |
| `diff` | none | startup addresses and directories | none | works; same stdout lifecycle |
| `fs` | none | startup facts and stopped result | none | works; same stdout lifecycle |
| `git heat` | none | report or empty result | none | works |
| `git pulse` | date select and optional author multiselect | report, empty, or cancellation | tracked scan/fetch phases | works only when prompts are avoided |
| `git fork` | overwrite confirm when needed | outcome/fallback detail or cancellation | tracked acquisition phases | works unless overwrite confirmation is needed |
| `git cm` | optional file select and commit confirm | generated message, normal/partial outcome, or safe failure context | tracked stage/collect/generate/commit/push phases | dry-run/non-prompt modes work; commit modes rejected |
| `rm` | explicit confirm or smart-clean selects | final/cancellation outcome | notices labeled as progress | only explicit `--force` is reliably noninteractive |
| `run` | script and package-manager selects | cancellation only | launch notice before handoff | rejected before child startup |
| `tunnel connect` | remembered-connection select when ambiguous | cancellation only | logger Lifecycle Log | explicit/unique configuration works; ambiguity errors |
| `tunnel server` | none | none | logger Lifecycle Log | works |
| `zip` | package/source/glob/name planning form | final or normal-failure outcome | notices labeled as progress | rejected at the first required question |
| `upgrade` | none | previous state, already-current, scheduled, or aborted result | none | works |

## Command details

### Export

#### `export env`

- Reads stdin only when discovery produces a real environment choice. The
  question is `Select environment`; Rich uses AltScreen and Plain writes the
  choices and retries to stderr.
- Success without `--out` calls `Result` twice, producing `Exported variables:`
  and the complete JSON on stdout. With `--out`, stdout says `Writing output to
  <target>` and the JSON is only written to the file.
- No `.env` files, an unknown `--env`, parse/encode errors, and filesystem
  errors produce only the root Diagnostic Record. There is no discovery/read/
  encode/write progress.
- Cancellation is a successful `Cancelled` Command Result on stdout. In Rich,
  the completed selection context disappears with AltScreen.
- Automation silently chooses the sole viable file. Ambiguous discovery returns
  `export env requires an interactive terminal` without reading stdin. A
  redirected stdout remains the established heading-plus-JSON format.
- The file-target message is emitted before `WriteOutput`; a failed write leaves
  a success-sounding stdout line followed by an error on stderr.
- Evidence: package behavior, selection, terminal, root-outcome, and
  [`acceptance/export_env_test.go`](../../../acceptance/export_env_test.go).

### Configuration

#### `config fork list`

- Never reads stdin or opens AltScreen. It reads configuration, then emits one
  stdout document: title, `NAME TYPE SCHEME HOST TOKEN`, token previews, and a
  count. Empty configuration emits warning plus add-command guidance.
- Read/configuration failure emits only stderr error. Automation and redirected
  stdout use the same plain content with no controls.
- Evidence: package terminal/read/run tests and
  [`acceptance/config_fork_test.go`](../../../acceptance/config_fork_test.go).

#### `config fork add`

- Reads five ordered questions from stdin: alias, host, provider, protocol, and
  secret access token. Validation stays inside the form; the token is password
  input and never appears in result text.
- Success and cancellation are stdout results. There is no durable summary of
  the four non-secret answers and no save phase. Form history is lost in Rich.
- Store/save failures close AltScreen and leave only the root stderr error.
  Automation is rejected before the store is resolved or input is read.
- Evidence: input/write/run/terminal tests and standalone add coverage in
  [`acceptance/config_fork_test.go`](../../../acceptance/config_fork_test.go).

#### `config fork remove`

- Reads a select and default-negative confirmation from stdin when instances
  exist. Selection and confirmation are lost after Rich exits.
- Empty configuration sends `No instances configured` as a Notice and `Nothing
  to remove` as stdout Result. Plain therefore has stderr plus stdout; Rich loses
  the notice; Automation suppresses the notice, leaving only stdout.
- Selection cancellation and a declined confirmation both become `Cancelled`
  on stdout. Success names the removed instance on stdout.
- Read/write/interaction failures emit only root stderr. Automation permits the
  empty branch but rejects a nonempty catalog before mutation.
- Evidence: remove input/confirmation/write/run/terminal tests and standalone
  removal coverage in `acceptance/config_fork_test.go`.

#### `config cm list`

- Never reads stdin or opens AltScreen. Stdout contains a title and rows with
  default marker, profile, model, and base URL; API keys are excluded. Empty
  state contains add-command guidance.
- Read/configuration failure is stderr only. Automation and redirection retain
  the plain result.
- Evidence: read/run/terminal tests and
  [`acceptance/config_test.go`](../../../acceptance/config_test.go).

#### `config cm add`

- Reads profile name, base URL, model, and secret API key. Validation is
  form-local and the key is masked.
- Success or cancellation is stdout only after AltScreen closes. Completed
  non-secret fields are not replayed. Store/save failure leaves only stderr.
- Automation is rejected before configuration access or stdin read.
- Evidence: input/write/run/terminal tests and standalone CM add coverage.

#### `config cm use`

- No stdin, Live View, or progress. Success writes `Default CM profile set to
  <name>` to stdout. Missing profile/storage errors write stderr only.
- Automation and redirected output work unchanged.
- Evidence: semantic/run/terminal tests and standalone use/list verification.

#### `config cm set`

- No stdin, Live View, or progress. Success writes `Profile <name> updated` to
  stdout without echoing the key or value. Invalid profile/key/value and storage
  failures write stderr only.
- Automation and redirection work. Secret-like values remain absent from the
  result.
- Evidence: semantic/run/terminal tests and standalone field-matrix/redaction
  coverage.

#### `config cm remove`

- Validates the named profile before opening the confirmation. Existing
  profiles read one default-negative confirm; missing profiles error without
  reading stdin.
- Cancel, decline, and terminal cancellation all produce `Cancelled` on stdout;
  success names the removed profile. Rich retains none of the confirmed prompt.
- Automation preserves missing-profile validation, but an existing profile
  fails with `config cm remove requires an interactive terminal` before write.
- Evidence: validation/confirmation/write/run/terminal tests and standalone
  missing/existing-profile coverage.

#### `config cm test`

- No stdin or progress. A successful provider call writes a titled response and
  `Done` to stdout.
- Provider failure writes a safe profile/base URL/model document to stdout and
  then the redacted provider error to stderr with exit 1. Resolution failures
  that occur before safe diagnostic construction write only stderr. An empty
  provider response is an operational error, not an empty result.
- Automation and redirected stdout are supported. API keys are redacted from
  response content and errors.
- Evidence: request/transport/response/run/terminal tests and local-provider
  standalone coverage.

### Service commands

#### `diff`

- Reads no stdin. After binding the server it writes local/network browser and
  MCP URLs plus baseline/target directories to stdout as a Command Result, then
  blocks in the foreground.
- It has no terminal Lifecycle Log for initial refresh, refresh failure,
  requests, readiness beyond binding, or shutdown. A failed asynchronous
  initial comparison is visible through the browser API, not the terminal.
- SIGINT/SIGTERM closes the server and exits successfully without a terminal
  stopped line. Startup/bind/validation/serve failures use root stderr.
- Automation and redirected stdout use the same startup result. Rich never
  opens AltScreen because the command only calls `Result`.
- Evidence: terminal, lifecycle, HTTP/MCP, signal, and
  [`acceptance/diff_test.go`](../../../acceptance/diff_test.go) coverage.

#### `fs`

- Reads no stdin. After binding it writes browser URLs, directory, bind,
  management, upload, HTML, authentication, and session facts to stdout, then
  blocks. Normal signal shutdown appends `File Browser stopped.` to stdout.
- There is no stderr Lifecycle Log for start, task state, warnings, server
  failure, or stop. Browser task events and errors live in HTTP/SSE state, not
  the terminal.
- Validation/start/serve failures use root stderr. Automation and redirection
  retain the stdout lifecycle; Rich never opens AltScreen.
- Evidence: terminal/runtime/server tests plus signal, authentication, managed
  operation, and browser acceptance journeys in
  [`acceptance/fs_test.go`](../../../acceptance/fs_test.go).

#### `tunnel connect`

- Explicit flags/environment or one remembered connection require no stdin.
  Multiple remembered connections open a masked select on stdin in interactive
  modes. Selection cancellation writes `Cancelled` to stdout and starts no
  client.
- The foreground client writes Lifecycle Logs to stderr: start, cleanup/remember
  warnings, terminal failure, stop, and debug reconnect scheduling. FRP child
  warnings also flow through the same logging facade; child streams are not
  handed through raw.
- Automation never presents the selector: a unique connection works; ambiguity
  returns an error. Text versus JSON and level filtering come from global
  diagnostics. Tokens are masked/redacted.
- Resolution and runtime failures are often logged and then returned, so root
  can append a second error line for the same failure. Normal context shutdown
  logs stop and returns success.
- Evidence: resolver/client lifecycle tests, terminal selection and masking
  tests, root diagnostic tests, and Tunnel acceptance validation.

#### `tunnel server`

- Reads no stdin and writes no stdout. All normal lifecycle is the stderr
  logger: control-plane start, FRP listener configuration, state directory,
  stopping, stopped, warnings, and failures.
- Text logs are readable records; JSON mode is stable NDJSON. Global levels can
  hide info lifecycle lines. Sensitive values are excluded or redacted.
- Configuration errors log `Could not resolve tunnel server configuration` and
  are then returned, producing a second root `error:` line. Runtime failures can
  have the same logged-plus-root duplication.
- Context cancellation follows ordered shutdown and normally exits 0.
- Evidence: configuration/runtime/logger/root tests and
  [`acceptance/tunnel_test.go`](../../../acceptance/tunnel_test.go).

### Git

#### `git heat`

- No stdin, Live View, or progress. Repository discovery and log collection are
  silent. Success writes the complete ranked report, summary, headers, and
  legend to stdout; query matches only change TTY styling.
- No commits/files yields `No changed files found in the selected range.` on
  stdout. Flag/input/repository/Git failures write stderr only.
- Automation and redirected stdout are fully supported and control-free.
- Evidence: input/Git/log/aggregate/report/terminal/root tests and
  [`acceptance/git_heat_test.go`](../../../acceptance/git_heat_test.go).

#### `git pulse`

- Writes introduction and workspace as a Notice, tracks `Scanning repositories`
  and `Fetching commits`, asks for a date range unless `--days` is supplied, and
  asks for authors only when more than one author exists.
- Success writes the complete repository/commit tree to stdout. No repositories,
  no commits, and form cancellation write stdout results. Per-repository fetch
  failures are counted in `Result.FailedRepositories` but not presented, so a
  partial report does not explain omissions.
- Plain leaves notices, every phase update, prompts, and retries on stderr.
  Rich loses all of them when stdout Result restores the screen. A scan/fetch
  failure or tracked cancellation can therefore collapse to only root stderr.
- Automation suppresses notices/phases. It works when `--days` is supplied and
  there are at most one author; otherwise it may scan and fetch before failing
  at the author prompt.
- Evidence: discovery/fetch/report/tracking/terminal/root tests and
  [`acceptance/git_pulse_test.go`](../../../acceptance/git_pulse_test.go).

#### `git fork`

- Starts with a Git Fork Notice, asks a default-yes overwrite confirmation only
  for a nonempty destination, then tracks resolve, default branch, archive,
  clone fallback, and ready phases.
- Cancellation/decline writes `Cancelled` to stdout. Success writes resolved
  repository facts, safe fallback errors/warnings, acquisition outcome, and
  destination. Default-branch/archive failures are therefore represented both
  as failed transient phases and as final fallback detail, but only the stdout
  detail survives Rich.
- Final clone or filesystem failure emits no Command Result; the last failed
  phase is lost with AltScreen and root prints stderr. The destination may
  already have been removed after overwrite confirmation.
- Automation works if no confirmation is needed and suppresses all phases. A
  nonempty destination errors before removal or stdin read.
- Evidence: input/provider/archive/clone/tracking/terminal/root tests and local
  provider standalone coverage in
  [`acceptance/git_fork_test.go`](../../../acceptance/git_fork_test.go).

#### `git cm`

- Depending on flags, asks for staged files and/or default-yes commit approval.
  It tracks stage, collect, provider generation, commit, and push in separate
  tracked segments.
- Dry-run and generation-only success writes the generated message, profile,
  provider token usage, and local evidence coverage to stdout. No changes,
  nothing selected, cancellation, commit, and push outcomes are stdout results.
- Before commit confirmation the generated message is a Notice. If the user
  declines in Rich, the valuable generated message disappears and only
  `Cancelled` remains on stdout. On successful commit, stdout normally retains
  only `Commit created` or `Commit created and pushed`, not the generated
  message that was visible in AltScreen.
- Provider failure after profile resolution writes safe profile facts to stdout
  and a redacted error to stderr. Push failure after a successful commit writes
  `Commit created` to stdout and the push error to stderr, correctly retaining
  partial success. Earlier failures have only stderr.
- Automation preflights prompt-dependent modes before mutation. Generation-only
  and dry-run modes work while phases are silently consumed.
- Evidence: mode/snapshot/provider/commit/push/tracking/terminal/root tests and
  extensive standalone generation, prompt, hook, stale-snapshot, partial-push,
  and redaction coverage in
  [`acceptance/git_cm_test.go`](../../../acceptance/git_cm_test.go).

### Files, processes, archives, and updates

#### `rm`

- Explicit paths show missing-path notices and, unless `--force`, paths plus a
  default-negative confirmation. Smart mode asks for an action, emits scan
  notices, and may ask for target multiselection.
- `Scanning...`, found counts, deleting, deleted counts, skipped paths, and the
  introduction are all `Notice`, not `Track`. Plain prints them to stderr;
  Rich loses them; Automation suppresses them.
- Cancellation, no valid paths, nothing selected, nothing to clean, and `Done!`
  are stdout results. Per-path deletion failures are warnings followed by
  `Done!` and still exit 0. In Rich or Automation those skipped-path warnings
  can be absent from terminal history entirely.
- Automation works for explicit `--force`; missing-path and deletion warnings
  are suppressed. Non-forced explicit and all smart flows error at a prompt
  before deletion.
- Evidence: planning/delete/prompt/presentation/root tests and
  [`acceptance/rm_test.go`](../../../acceptance/rm_test.go).

#### `run`

- Discovers package scripts, asks for a script and package manager, and writes
  the selected child command as a Notice. There is no automatic single-option
  selection.
- Immediately before child startup it closes the terminal run, restoring
  AltScreen and releasing the renderer lease. The child inherits raw stdin,
  stdout, and stderr; ycy does not decorate or intercept them.
- Because there is no transcript, Rich loses the script/manager selections and
  launch command before the child starts. Cancellation alone writes `Operation
  cancelled.` to stdout.
- A child nonzero exit is mapped directly to its exit code with no ycy error
  line. Discovery/start failures use root stderr. Automation fails before child
  startup or stdin read.
- Evidence: discovery/manager/process/handoff/presentation/root tests, shared
  PTY handoff evidence, and
  [`acceptance/run_test.go`](../../../acceptance/run_test.go).

#### `zip`

- Planning may ask for package, source, glob multiselect, and output filename;
  it also emits project-detection notes. Collection, compression, writing, and
  save messages are Notices rather than tracked phases.
- Success writes `Done!` to stdout. Cancellation, missing/non-directory input,
  collection failure, no files, no valid files, compression failure, and write
  failure also produce warning/error prose through stdout `Result` calls.
- Those operational result kinds return nil from the command module, so they
  normally exit 0 despite failure text. Reveal/open failure is recorded only as
  `RevealFailed` and is not presented; stdout still says `Done!`.
- Plain preserves notes/progress on stderr, Rich loses them, and Automation
  suppresses them before failing at the first required question. No archive is
  created on the tested Automation path.
- Evidence: planning/archive/run/adapter/root tests and
  [`acceptance/zip_test.go`](../../../acceptance/zip_test.go). There is no
  command-specific Rich PTY or terminal-presentation test comparable to the Git
  commands.

#### `upgrade`

- Reads no stdin and has no progress surface while resolving, downloading, or
  scheduling an update. It never opens AltScreen.
- Stdout can contain a consumed previous transaction result followed by
  already-current or scheduled text. A classified abort writes a redacted
  stderr error and `Update aborted.` stdout result, then returns its owned exit
  code without root duplication. Other operational failures use root stderr.
- Startup consumption can emit a previous update result before any unrelated
  command. Pending state produces a state line plus an error on stderr and
  prevents command execution.
- Automation and redirection preserve the same result text.
- Evidence: updater unit tests, presentation/root tests, process startup tests,
  and replacement/rollback acceptance in
  [`acceptance/upgrade_test.go`](../../../acceptance/upgrade_test.go).

## Cross-command findings

1. **There is no Interaction Transcript.** Rich questions, non-secret answers,
   notices, final phase states, cancellation location, and failure location are
   not replayed after AltScreen. Plain happens to retain prompts and phase spam,
   but it is not a semantic or redacted transcript.
2. **Progress coverage is narrow.** Only `git pulse`, `git fork`, and `git cm`
   use typed `Track`. `rm` and `zip` call notices "progress"; Export, Config,
   Git Heat, Diff, FS, and Upgrade have silent work intervals.
3. **Service presentation is split.** Tunnel already has stderr text/NDJSON
   Lifecycle Logs. Diff and FS use stdout Command Results for startup/lifecycle
   and have no line-oriented logging facade.
4. **Some stdout results are intentionally partial or mixed.** CM Test and Git
   CM emit safe context/partial success before a stderr failure. Export writes a
   human heading before JSON. Upgrade startup state can prefix any command.
   Redesign must preserve these established stream contracts unless a separate
   behavior change is approved.
5. **Normal cancellation is usually stdout.** Interactive cancellations across
   Export, Config, Git, RM, Run, Tunnel Connect, and ZIP are modeled as
   successful Command Results, not diagnostics. A future transcript must not
   duplicate that complete result.
6. **Operational failures are not uniform.** ZIP maps several failures to exit
   0 result kinds; RM reports per-path failures but still says `Done!`; Git Pulse
   hides failed-repository counts; Git Fork and Git CM preserve selected
   fallback/partial-success facts; Tunnel may log and root-report the same
   error.
7. **Automation is effect-aware but command-specific.** Some leaves reject up
   front, while others do discovery or expensive read-only work before learning
   a prompt is required. Notice and phase suppression can remove warnings even
   from valid noninteractive paths such as `rm --force`.
8. **The current visual layer is semantic but shallow.** Results have only six
   colored roles; forms split between Huh and a custom ASCII list; progress uses
   textual `[active]`/`[done]` prefixes. There is no unified component theme,
   status icon system, loading primitive, or per-command layout.
9. **`run` already has the correct ownership boundary.** It restores the
   terminal before raw child handoff. The missing part is a durable, semantic
   selection/launch summary before that handoff.

## Evidence coverage and gaps

Existing evidence is unusually broad for stream preservation:

- Shared terminal unit and PTY tests cover classification, stdin/stderr forms,
  validation, cancellation, long-list navigation, resizing, AltScreen cleanup,
  stdout redirection, diagnostic deferral, tracked cancellation, no-color, and
  child handoff.
- Every leaf has package behavior tests. Most have terminal or root-outcome
  tests that separate stdout and stderr in Plain/Automation modes.
- All command families have tagged standalone acceptance coverage; Diff, FS,
  Git CM, Upgrade, and Tunnel additionally exercise process, signal, network,
  or replacement behavior.
- Root command-surface fixtures cover registered paths, flags, help, parser
  recovery, version, completion, diagnostics, and process ownership.

The specification and rollout tickets must add evidence for gaps the current
suite cannot express:

- no test asserts a semantic transcript outside AltScreen because no transcript
  exists;
- raw PTY captures prove terminal restoration and byte ordering, but do not
  prove what remains readable in human scrollback;
- no visual golden or screenshot compares hierarchy, symbols, color, spacing,
  form focus, validation, loading, and final transcript together;
- not every leaf has PTY coverage for success, cancellation, failure, and
  redirected stdout;
- Diff/FS have no Lifecycle Log contract, and Tunnel tests do not require
  deduplication between logger errors and root errors;
- ZIP lacks a command-specific terminal journey, and RM/ZIP notice-based
  progress has no durable-warning assertion in Rich or Automation;
- no cross-command test enforces exactly one complete Command Result on stdout
  plus exactly one semantic transcript on stderr.

No newly discovered decision requires another ticket. The existing shared
contract, visual prototype, Service Command, per-command, and rollout tickets
cover every gap above. ZIP/RM exit semantics, hidden Git Pulse failure counts,
and existing mixed-stream outcomes are behavioral contracts to preserve and
present clearly, not behavior changes inside this map.
