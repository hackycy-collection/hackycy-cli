# Define the command-result and diagnostic contract

Type: grilling
Status: resolved
Blocked by: 02, 03

## Question

How should ycy separate durable Command Results from Diagnostic Events across
all commands? Decide the exact behavior and flag grammar for `--log-level`,
JSON log formatting, quiet and verbose operation; timestamp, field, redaction,
and color rules; how user-actionable errors differ from operational diagnostics;
and which progress/status messages may appear on stderr.

The decision must explicitly protect existing JSON/text command results,
subprocess output passthrough, shell pipelines, and tests that currently merge
or separately capture stdout and stderr. It should define acceptance examples
for an Interactive Session and an Automation Session rather than prescribe a
logging implementation.

## Answer

Approved on 2026-08-26. ycy separates durable Command Results, Diagnostic
Events, User-Actionable Errors, and Raw Child I/O by purpose rather than by a
blanket stdout-to-stderr migration. Existing command-specific stream contracts
remain protected until their own migration slice explicitly reclassifies a
human-facing presentation.

### Stream contract

- A Command Result stays on stdout. This includes machine-consumed JSON,
  reports, browser readiness URLs, version/help/completion output, and the
  deliberate Upgrade result/exit exceptions.
- A Diagnostic Event goes to stderr through the lease-aware Diagnostic Writer.
  Tunnel lifecycle, retry, health, and operational-context records remain
  Diagnostics. Rich Interactive transient progress also uses stderr while it
  holds a Renderer Lease.
- A User-Actionable Error is exactly one plain `error: <message>` line on
  stderr and retains the command's existing exit semantics. It is not turned
  into JSON by `--log-format=json`.
- Raw Child I/O stays unwrapped: child stdout remains stdout and child stderr
  remains stderr. No Transient View may coexist with that direct ownership.

Automation Sessions do not gain generic start, progress, or completion records
for one-shot commands. They retain final Command Results, User-Actionable
Errors, and already-established service Diagnostics. This preserves pipelines
while leaving command-specific migration work free to move only safe,
human-facing progress into the approved stderr presentation path.

### Diagnostic configuration

The public grammar is:

```text
ycy [--log-level debug|info|warn|error | -v | -q]
    [--log-format text|json]
    <command>
```

- `--log-level` remains the global threshold. `-v`/`--verbose` select
  `debug`; `-q`/`--quiet` select `error`.
- `-v`, `-q`, and an explicit `--log-level` are mutually exclusive. Repeating
  either alias or combining the aliases is an error; ycy does not count
  verbosity or silently choose a winner. `--quiet` never suppresses a
  User-Actionable Error.
- Explicit CLI controls take precedence over `YCY_LOG_LEVEL` and the new
  `YCY_LOG_FORMAT`; environment values take precedence over `info` and `text`
  defaults. Diagnostic Configuration never changes Command Results or the
  separate `DEBUG=1` panic-stack behavior.
- Invalid level, format, or conflicting controls produce one User-Actionable
  Error and exit `1` before a command side effect. Validation happens only for
  an actual command-handler execution: version, help, completion, and other
  discovery paths retain stable output even with invalid diagnostic controls.
  The special `run` argument parser must honor this same grammar and timing.

### Record formats

`text` is the Bun-compatible human form:

```text
2026-08-26T12:34:56.789Z INFO  [tunnel.server] Tunnel started {"port":7000}
```

It uses a UTC ISO timestamp at millisecond precision, a padded uppercase
level, optional bracketed scope, a one-line message, and optional compact JSON
context. Newlines are escaped. `json` produces one NDJSON object per
Diagnostic Record:

```json
{"timestamp":"2026-08-26T12:34:56.789Z","level":"info","scope":"tunnel.server","message":"Tunnel started","context":{"port":7000}}
```

`scope` and `context` are omitted rather than represented as `null`. JSON
levels are lowercase and JSON never contains ANSI. The record schema applies
only to Logger-produced Diagnostic Events, not to User-Actionable Errors.

Rich Interactive text records use the approved visual roles: dim gray
timestamp and DEBUG, green INFO, yellow WARN, bold red ERROR, and cyan scope;
context stays uncolored. Plain Interactive, Automation, `NO_COLOR`, and JSON
records are entirely unstyled and never use cursor control or an alternate
screen.

### Safety and duplicate-failure policy

Text messages, fields, nested maps, arrays, and nested errors receive the same
recursive redaction before either formatter sees them. Field-name matching is
case-insensitive and separator-insensitive for credential-shaped data,
including authorization, cookies, passwords, secrets, tokens, API keys,
credentials, private keys, and request bodies. Credential-shaped text such as
`Bearer <value>` and `key=value` is redacted as well. Values that cannot be
safely encoded are replaced with a deterministic safe placeholder.

Logger-produced error context may accompany a failed command only when it
adds operational information. It must not repeat the final User-Actionable
Error message. Logger output does not gain stack traces; the existing explicit
panic stack policy remains separate.

### Acceptance evidence

The implementation and migration plan must demonstrate at least these cases:

| Case | Required observable behavior |
| --- | --- |
| Rich Interactive Tunnel service | stderr text records use the approved semantic colors and scope shape; stdout carries no log records; `NO_COLOR` removes color only. |
| Automation Tunnel service with `--log-format=json` | stderr is parseable NDJSON Diagnostic Records without ANSI; stdout remains free of log records; a fatal user error remains one plain `error:` line. |
| One-shot Automation command with a JSON Command Result | stdout is unchanged JSON with no progress prefix; no implicit prompt, cursor sequence, or new generic progress record is emitted. |
| Foreground child process | child stdout/stderr remain independently observable and are never captured by a renderer or converted to records. |
| Invalid diagnostic configuration | an executing command fails before effects with one plain error and code `1`; version, help, and completion still retain their normal output. |
| Redaction | text and JSON contain no direct or nested credential, request-body, or bearer-token values. |

This decision adds the canonical terms Diagnostic Record, Diagnostic
Configuration, and User-Actionable Error to `CONTEXT.md`.
