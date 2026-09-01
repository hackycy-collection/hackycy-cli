# Choose the root, help, discovery, and error presentation

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What should the root and command-group discovery documents, help, version,
parser errors, global failures, startup update notices, and diagnostic controls
look like in terminal and redirected modes while preserving their existing
content, stream, exit, and machine-consumption contracts?

## Answer

Root presentation uses one semantic `DiscoveryDocument` and one root-error
projection. `pkg/cmd/root` continues to own Cobra discovery, parser recovery,
exit-code decisions, and private-route bypasses; `internal/terminal` owns only
the capability-aware styling and control-free rendering. No root or help path
starts AltScreen or creates an Interaction Transcript.

### Discovery and help

- Root, command groups, and explicit `--help` produce a durable stdout
  document. On a color-capable TTY it uses the Signal Rail visual vocabulary:
  a compact `YCY / <command>` eyebrow, command title, summary, and clearly
  separated `Usage`, `Commands`, `Options`, and `Examples` sections. The
  existing semantic fields and Cobra order remain intact.
- Commands list every available direct child with its summary. Options retain
  local and inherited/global flags, shorthand, usage, defaults, and examples;
  none are hidden for visual brevity. Local and inherited options may be
  grouped for scanning, but their values and order remain discoverable.
- A parent invoked without a leaf and the root invoked with no arguments keep
  their existing discovery output and exit code 1. Explicit `--help` exits 0.
  A small `Next: ycy <child> --help` hint may be appended when a group has
  descendants, but it never executes a command or changes stdout semantics.
- Wide TTYs use aligned columns. Narrow terminals switch to a single-column
  form with wrapped summaries and examples; command names, flag names, and
  required content are never truncated. Height does not trigger pagination or
  an interactive pager: the complete document is written once.
- Styling is applied only when stdout is a color-capable terminal. Plain,
  redirected, and Automation output contains no ANSI/control sequences and
  the same semantic content. Status emphasis always pairs color with words or
  symbols so it remains legible without color.
- `--version`/`-V` remains exactly the version string plus newline, and
  `completion <shell>` remains the raw Cobra-generated script. These machine
  surfaces bypass styling and diagnostic configuration, including invalid
  logging flags supplied alongside them.

The frozen command-surface test remains the compatibility gate. Intentional
help-layout changes update only the help snapshots in the same change;
the command/flag manifest and completion snapshots remain unchanged unless
Cobra's actual command contract changes. Additional semantic assertions must
continue to verify every Usage, command, flag, and example field independent of
visual layout.

### Parser and root errors

- Unknown commands, unknown flags, missing arguments, and suggestion recovery
  remain one actionable stderr record, exit 1, with stdout empty and no usage
  dump. A TTY may prefix the line with a red `✕`; the existing problem,
  `did you mean`, and `Run '<path> --help'` wording remains semantically
  unchanged. Redirected output is the same line without styling and is not
  converted into JSON/NDJSON merely because `--log-format=json` was present.
- Ordinary command, web/root construction, and startup failures produce one
  redacted root Error Diagnostic. The presenter removes terminal controls,
  normalizes CR/LF to spaces, and applies a fixed diagnostic length limit with
  an omission marker; it preserves the original error chain in `Outcome.Err`.
  Root does not append automatic help or duplicate a command-owned failure
  result.
- Exit-coded child/result errors continue to return their owned code without a
  second root error line. Panic handling keeps the existing compatibility
  behavior (blank stdout line, redacted stderr error, and a debug stack only
  when `DEBUG`/development mode is enabled), with TTY symbol/color enhancement
  only.
- Discovery presentation now propagates stdout write failures: its presenter
  returns `error`, root reports one diagnostic and exits 1, and it does not
  retry, move content to stderr, or claim success after a partial write.
- The no-argument/parent internal help marker remains private. It controls the
  established exit-1 distinction but never becomes an error line, JSON record,
  or panic presentation.

### Startup update and diagnostics

- Hidden updater transaction consumption still happens before Cobra and before
  the requested command. A completed update result is emitted to stdout first,
  preserving the current ordering even when it prefixes the later command's
  result; it uses no AltScreen, prompts, or Transcript.
- A pending transaction emits one warning Lifecycle line with the update
  state, then one short root error explaining that execution is blocked. A
  failed transaction emits one error line. The full state text is not repeated
  by the outer root reporter.
- `--log-level`, `--log-format`, `--verbose`, and `--quiet` retain their
  aliases, mutual-exclusion checks, environment precedence, and configure-
  before-effects ordering. They affect only Logger-backed Diagnostic Records.
  Help, version, completion, and parent discovery bypass invalid diagnostic
  configuration as they do today. JSON format changes logger records only;
  root/parser/help/result text stays human-readable or machine-raw according to
  its established contract.
- Hidden updater and thumbnail-worker routes bypass Terminal Experience and
  retain their private machine protocol, framing, and error prefixes.

### Ownership and evidence

Each command/group owns its wording and semantic discovery fields; root does
not maintain separate Rich and Plain documents. `terminalDiscoveryAdapter`
projects the one document through `Finish` and propagates errors. The rollout
must update frozen help snapshots deliberately, exercise TTY and redirected
streams, assert zero control sequences in non-TTY modes, verify redaction and
single-line errors, and preserve the command manifest/completion artifacts.
