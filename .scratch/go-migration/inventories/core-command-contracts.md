# Global, configuration, and local command compatibility inventory

Inventory date: 2026-08-22
Legacy baseline: `78358c0201b71891e36603d6abb8d7c87d54ad57`
Scope: global CLI behavior and 13 leaf commands: `export env`; `config fork add/remove/list`; `config cm add/list/use/remove/set/test`; `rm`; `run`; and `zip`.

## First-release scope

This inventory records observed facts for a parity-first port. For the first Go release, the actual Bun behavior remains the test baseline even where this document labels it a legacy defect or recommends a correction. Those labels form a post-parity hardening backlog and do not authorize new validation, prompts, limits, atomicity, or exit semantics. Only a focused implementation or test proving that Go cannot reproduce a behavior under the migration's hard constraints may create a narrow compatibility exception.

## Contract classification

This inventory separates observable behavior into three classes so later hardening remains distinguishable from first-release parity:

- **Compatibility contract**: command and option names, defaults, successful effects, persistent formats, machine-readable data, external process/network requests, and cancellation/exit meaning that existing users or callers can rely on.
- **Presentation freedom**: ANSI styling, screen clearing, spinner frames, column widths, exact prompt decoration, and incidental whitespace may change. Prompt choices and safety decisions remain contracts even when their rendering changes.
- **Legacy defect**: dangerous, internally contradictory, undocumented, or nonfunctional behavior. Reproduce its observable effect for first-release parity, record it explicitly in tests or migration notes, and defer correction unless a proven Go incompatibility makes exact reproduction impossible.

No separate golden or black-box fixture corpus is required. Go tests should encode the contracts directly, using `legacy/bun/` as the reference while each command is ported. Small inline compatibility vectors are still required for encryption and serialized data.

## Existing test baseline

The inspected tree has tests only for environment export, two CM resolution/client cases, and the shared logger. There are no tests for the root command tree, config storage/locking/crypto, any config UI leaf, `rm`, `run`, or `zip` discovery/planning/archive behavior.

The following legacy tests passed during this inventory:

```text
15 pass, 0 fail
src/commands/export/env/env.test.ts
src/config/cm.test.ts
src/config/client.test.ts
src/shared/log/index.test.ts
```

This is evidence that the inspected baseline is runnable, not a requirement to run Bun tests after the source is archived.

## Global CLI contract

### Command surface and parser behavior

- Program name is `ycy` and the version comes from root `package.json`; both `-V` and `--version` print the plain version, currently `0.0.69`, and exit 0.
- Global option `--log-level <level>` accepts `debug`, `info`, `warn`, or `error`; it is accepted before or after a subcommand. CLI value wins over `YCY_LOG_LEVEL`, whose fallback is `info`. Values are trimmed and case-folded.
- Root registration order is `export`, `diff`, `config`, `git`, `rm`, `fs`, `tunnel`, `zip`, `run`, `upgrade`. Help ordering is presentation, but names and nesting are contracts.
- Root or a command group invoked without a leaf prints that group's help and exits 1. Explicit `--help` exits 0.
- An unknown root command prints `error: unknown command '<name>'` and exits 1. Commander reports missing required operands and excess arguments on stderr with exit 1.
- `parseIntArg` uses base-10 `parseInt`: nonnumeric input fails, but a numeric prefix such as `3oops` is accepted as `3`, and negative values are accepted. This permissiveness is a legacy defect that remains part of first-release parser parity.
- A hidden updater argument is detected anywhere in `argv` before Commander parsing. Pending updater status may block normal command execution before help/version parsing. Its details belong to `Inventory upgrade and release-artifact compatibility contracts`.

### Errors and logging

- Uncaught exceptions and unhandled rejections print a blank line to stdout, then the error message to stderr, and exit 1.
- If `DEBUG` is nonempty or `NODE_ENV=development`, the stack is appended. Exact blank lines and stack formatting are presentation, but nonzero failure and opt-in diagnostics must remain.
- Structured logs use stderr, ISO timestamps, padded uppercase levels, optional dotted scope, JSON context, and terminal color only when stderr is a TTY. Messages replace embedded newlines with `\n` in formatted output.
- Context keys matching authorization/cookie/password/secret/token and matching assignments or bearer credentials in messages/errors are redacted. Redaction is a security contract.
- These local commands mostly use direct prompt output rather than the structured logger. Global log parsing must still work consistently for the whole command tree.

### Required Go tests

1. Root/group no-argument, help, version, unknown command, missing operand, and extra operand exit behavior.
2. Global option before/after subcommand; CLI/environment/default precedence; invalid value; help/version bypass behavior.
3. Panic/error mapping with and without debug diagnostics, with stdout/stderr captured separately.
4. Logger level filtering, shared runtime reconfiguration, child scope/context, ANSI TTY selection, and secret redaction.
5. Signal cancellation and one composition-root exit path; command modules should return results/errors rather than call `os.Exit` internally.

## Shared configuration persistence contract

All `config` leaves share `~/.ycy-cli/config.json` with `git fork`, `git cm`, and tunnel commands. It must be treated as an existing cross-command data contract, not as command-local storage.

### Location and schema

- Config directory is `(USERPROFILE || HOME || os.UserHomeDir())/.ycy-cli` on every platform. The unusual unconditional `USERPROFILE` precedence is part of locating existing data.
- Config file is `config.json`; the lock directory is `.config.lock`.
- Current normalized shape is:

```json
{
  "salt": "<base64 32 random bytes>",
  "fork": {
    "instances": {
      "<name>": {
        "host": "github.com",
        "scheme": "https",
        "type": "github",
        "token": "<encrypted>"
      }
    }
  },
  "cm": {
    "defaultProfile": "<optional name>",
    "profiles": {
      "<name>": {
        "baseURL": "https://provider.example/v1",
        "model": "<model>",
        "apiKey": "<encrypted>",
        "temperature": 0.2,
        "timeoutMs": 300000,
        "maxOutputTokens": 1000
      }
    }
  },
  "tunnel": {
    "connections": {}
  }
}
```

- `cm` and `tunnel` are optional. Profile tuning fields are optional on disk and receive defaults on resolution.
- Missing files read as an empty in-memory config with a newly generated salt. Reads do not create the file; the first update or explicit ensure operation publishes it.
- Legacy top-level `instances` is normalized into `fork.instances`; legacy `ai` is normalized into `cm`; an instance `host` containing a URL is split into `scheme` and host. These read migrations must remain available.
- Malformed JSON fails. A syntactically valid but unexpected root silently normalizes toward an empty config; unknown fields are discarded on the next write. Silent reset/data loss is a legacy defect retained for first-release parity unless direct-read implementation evidence proves the Go path cannot reproduce it.

### Encryption compatibility

- Key derivation is PBKDF2-HMAC-SHA256, 100,000 iterations, 32-byte output. Salt is base64-decoded from `config.salt`.
- Passphrase is `<machine-id>:<os username>`.
- Machine ID is `IOPlatformUUID` from `ioreg -rd1 -c IOPlatformExpertDevice` on macOS, trimmed `/etc/machine-id` on Linux, and `MachineGuid` from the registry query on Windows. Fallback is `<hostname>-<username>`.
- Secret encryption is AES-256-GCM with a random 16-byte IV, no additional authenticated data, and serialized as `<base64 iv>:<base64 auth tag>:<base64 ciphertext>`.
- Fork tokens and CM API keys use the same salt/key. The Go implementation must decrypt existing values byte-for-byte before any optional migration is considered.
- Moving a config file to a different machine or OS username may already make it undecryptable. That limitation is an existing fact, but the Go port must not create additional incompatibility on the same machine/user.

### Locking and publication

- Updates acquire `.config.lock` by atomically creating a directory, then create `owner.json` with random UUID, PID, and ISO start time.
- Acquisition retries every 25 ms for up to 10 seconds. A missing/invalid/dead owner is stale; an owner-less directory receives a one-second publication grace period. PID existence treats permission-denied as alive.
- Stale locks are first renamed to a unique stale path, then recursively removed. Release removes the lock only when the owner UUID still matches.
- Writes create a mode-0600 candidate beside the target and rename it over `config.json`; the config directory requests mode 0700. Permission modes are meaningful on Unix and generally advisory/ignored on Windows.
- Every read-modify-write occurs while holding the directory lock. The Go port must preserve concurrent-update serialization, atomic publication, stale-owner handling, and cleanup after errors. It may improve the mechanism only if cross-process tests prove equivalent safety on all three operating systems.

### Required Go tests

1. Missing, current, legacy `instances`, legacy `ai`, URL-host, malformed, and unknown-field documents.
2. Inline deterministic PBKDF2/AES-GCM vectors matching the legacy algorithm; real machine-ID adapters tested separately from pure crypto.
3. macOS/Linux/Windows machine-ID parsing and fallback, username binding, and config-path precedence.
4. Concurrent independent field updates, live-lock timeout, owner publication grace, stale/dead owner recovery, PID permission behavior, and owner-safe release.
5. Candidate cleanup and atomic replacement failures, Unix permissions, and Windows replacement behavior.

## `export env [dir]`

### Inputs and behavior

- Options: `-e, --env <name>`, `--merge`, and `-o, --out <file>`. Input directory defaults to cwd and is resolved absolutely.
- Discovery scans only direct files matching `.env*`, includes dotfiles, takes basenames, and sorts names lexically.
- `.envrc`, `.env.example`, and `.env.sample` are excluded by exact suffix logic. Names such as `.env.example.local` are currently included.
- `--env prod` requires exact `.env.prod`. Without `--merge`, only the selected environment file is parsed. With `--merge`, base `.env` is parsed first and the selected environment overrides duplicate keys.
- Without `--env`, base `.env` is selectable only when not merging. A lone base file is selected automatically. Other available choices use suffix labels and are presented interactively in sorted order.
- With `--merge` and only a base `.env`, the base is exported without a selection. With one non-base environment and no base, the command still prompts.
- Cancellation reports `Cancelled` and exits 0.
- Parsing is the `dotenv` package's semantics, including quoting, comments, `export`, escapes, and multiline handling. A Go parser must be selected and tested against representative syntax rather than replaced with ad hoc splitting.
- Result is an object of strings serialized as two-space JSON. `--out` is resolved relative to process cwd, not the input directory; it overwrites the target and does not create parent directories. Without `--out`, JSON is printed to stdout after prompt output.

### Failure and dependencies

- No usable files or a missing named environment throws and reaches the global exit-1 handler.
- Active dependencies to replace are `Bun.Glob`, `Bun.file`, `Bun.write`, `dotenv`, and `@clack/prompts`.
- JSON object key order is not semantically significant, but output must be deterministic, valid JSON, contain only strings, use two-space indentation, and avoid Go's default HTML escaping if it changes values relative to `JSON.stringify`.

### Required Go tests

1. No files; base only; environment only; base plus several environments; exact exclusions; nested names; files versus directories.
2. Explicit environment with and without merge, duplicate-key precedence, selection ordering, single-choice branches, and cancellation.
3. Representative dotenv grammar, empty values, Unicode, CRLF, comments, quoted multiline content, and malformed lines.
4. Stdout versus output file, cwd-relative output path, overwrite, missing parent, permissions error, and deterministic JSON.

Risk: **low**, except for dotenv parser parity and interactive selection.

## `config fork` leaves

### `config fork add`

- No CLI operands; always interactive.
- Prompts in order for alias, host, provider type, protocol, and password-masked token. Alias is nonempty with no whitespace; host and token are nonempty. Provider choices are GitLab then GitHub; protocol choices are HTTPS then HTTP.
- Cancellation exits 0. Save failure logs the message and exits 1.
- Success encrypts the token and writes `{host, scheme, type, token}` at the alias. Existing aliases are silently overwritten. The next read also normalizes a host entered with `http[s]://` even though input validation does not.
- Silent overwrite and failure to validate/trim a host are legacy defects retained for first-release parity and later hardening.

### `config fork remove`

- No CLI operand; loads configured aliases, then interactively selects and confirms one.
- Empty config reports nothing to remove and exits 0. Selection cancellation exits 0; negative confirmation returns normally.
- Confirmed removal updates the shared file atomically. Concurrent updates must not be lost.

### `config fork list`

- No operands and no mutation. Empty config prints an add hint.
- Nonempty output preserves config iteration order and shows name, provider type, defaulted scheme, host, and a preview derived from the first four characters of the **encrypted ciphertext**, not the decrypted token. The Ink table auto-unmounts after roughly 100 ms.
- Exact table widths, colors, timing, and ciphertext preview are presentation artifacts; the Go version must never reveal plaintext tokens. A stable noninteractive-friendly listing is permitted by the top-level compatibility policy if fields remain equivalent.

### Dependencies and tests

Active UI/runtime dependencies are Clack, Ink, React, Ansis, Node crypto/fs/os/path, and Bun subprocess/file behavior in machine-ID lookup. Tests must cover legacy validation/cancellation, silent add/overwrite behavior, encryption and persisted shape, empty/list ordering, removal confirmation, concurrent updates, plaintext non-disclosure, HTTP scheme, and all machine-ID adapters.

Risk: **high** because direct readability of encrypted credentials is mandatory.

## `config cm` leaves

### Shared profile resolution

- Defaults are temperature `0.2`, timeout `300000` ms, and max output tokens `1000`.
- Profile selection precedence is explicit CLI name, `YCY_CM_PROFILE`, stored default, then first stored profile.
- `YCY_CM_BASE_URL`, `YCY_CM_MODEL`, and `YCY_CM_API_KEY` independently override selected stored values. A fully environment-defined profile is named `env` only when no stored/selected name exists.
- Base URL is trimmed and trailing slashes removed. Stored API keys are decrypted only when the environment does not provide one.
- Numeric environment values use any finite JavaScript number; invalid values fall back. Command timeout override wins over environment and stored/default values. Missing base URL/model/key fails with an actionable message.
- Environment values currently bypass the stored range/integer validation. Negative/fractional timeouts or token counts are therefore possible; this is a legacy validation defect.

### `config cm add`

- No operands; interactively collects nonempty, no-whitespace profile name, nonempty OpenAI-compatible base URL, model, and password-masked API key.
- Base URL loses trailing slashes, model is trimmed, API key is encrypted, and the first added profile becomes default.
- Existing names are silently overwritten while retaining the current default selection. Silent overwrite and lack of URL validation are legacy defects retained for first-release parity and later hardening.

### `config cm list`

- No operands. Empty state prints an add hint. Nonempty state lists each profile in stored insertion order with `*` on the default plus model and base URL. API keys are never printed.

### `config cm use <profile>`

- Sets an existing profile as default. Missing profiles print `CM profile not found: <name>` and exit 1.

### `config cm remove <profile>`

- Always asks for confirmation. Cancel/no exits successfully without mutation.
- A confirmed missing profile exits 1. Removing the default selects the first remaining profile; removing the last leaves no default.

### `config cm set <profile> <key> <value>`

- Exact supported keys are `baseURL`, `model`, `apiKey`, `temperature`, `timeoutMs`, and `maxOutputTokens`.
- API key is encrypted. Temperature must be finite in `[0,2]`. Timeout must parse to an integer at least 1000; max tokens must parse to an integer at least 32.
- Integer parsing currently accepts numeric prefixes such as `1000oops`; model may become empty; base URL is not structurally validated. These are defects, not desired Go parser behavior.
- Unsupported key, missing profile, or validation failure exits 1.

### `config cm test [profile]`

- Resolves the named/default/environment profile, then POSTs `<baseURL>/chat/completions` with bearer token and JSON containing model, temperature, `max_tokens`, and fixed system/user messages requesting exactly `ok`.
- For `api.deepseek.com` models whose lowercase name starts `deepseek-v4-`, request JSON also disables thinking.
- Timeout aborts and reports the effective milliseconds. Non-2xx errors include status/status text and response body. Empty/malformed choice content fails with finish reason and a bounded response summary.
- Success reports provider response. Failure also prints provider name, base URL, and model, but never the API key, then exits 1.

### Dependencies and tests

Active dependencies are shared config/crypto, Clack, Ansis, and standards-shaped fetch/AbortController. Required tests cover every leaf's validation and cancellation, default selection changes, overwrite disposition, supported set keys and strict numeric parsing, full environment precedence matrix, decryption failures, request URL/headers/body, DeepSeek condition, timeout, HTTP/JSON/empty response errors, usage normalization, and secret non-disclosure.

Risk: **high** for encrypted storage and **medium** for provider semantics.

## `rm [paths...]`

### Explicit path mode

- Every operand is resolved against cwd. Existing files, directories, and symlinks are accepted; missing paths print a warning and are skipped.
- With no remaining path, the command reports no valid paths and currently exits 0. Without `--force`, it displays absolute paths and asks a default-negative confirmation. Cancel/no makes no change and exits 0.
- Deletions run concurrently with recursive, forceful removal. Partial failures are printed as skipped; successful deletions remain committed, and the command still reports done without a nonzero status.

### Smart mode

- No paths means interactive smart cleanup. Runtime depth defaults to 5; `--depth` uses the permissive global integer parser.
- User first chooses exactly one action: root `dist`, root `node_modules`, recursive `dist`, recursive `node_modules`, root lockfiles, or root agent directories.
- Lockfiles are `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `bun.lock`, and `bun.lockb`. Recognizing/removing another project's Bun lockfiles is retained command behavior, not an active ycy dependency.
- Agent directories are `.claude`, `.agents`, `.cursor`, `.copilot`, `.windsurf`, and `.aider`.
- Recursive scans ignore VCS directories, `__pycache__`, and every hidden directory; symlinked directories are not traversed. Read/permission errors are silently skipped. Match order is nondeterministic because concurrent scans append results.
- The depth check is off by one from an intuitive edge count: a scan with maximum depth `N` can match a target child at depth `N+1`.
- `--force` does not make smart mode noninteractive: it still asks which action, then deletes every discovered target without the multiselect. Without force, all targets start selected.

### Safety defects

- Explicit mode has no guard against `/`, a drive root, home, cwd, repository root, `.`/`..`, or paths outside the current project.
- Root cleanup checks existence rather than directory type, so a file named `dist` or `node_modules` is deletable.
- Invalid depth prefixes are accepted; negative depth silently yields few/no results.
- Missing targets, collection failures, and partial deletion failures do not produce reliable nonzero status.

These defects remain the first-release parity baseline. A dangerous-target policy, strict depth, deterministic display, and corrected failure status belong to post-parity hardening and do not block the Go port.

### Dependencies and tests

Active dependencies are Node filesystem/path/process, Clack, Ansis, and terminal screen helpers. Tests must use disposable directories and cover explicit files/dirs/symlinks, nonexistent mixes, default-negative confirmation, each smart action, legacy depth edges, hidden/VCS skips, unreadable paths, duplicate/nested targets, force semantics, cancellation, partial failures, and the observed lack of a protected-target deny table.

Risk: **medium** because implementation is small but destructive.

## `run [path]`

### Intended command behavior

- Resolve optional project path against cwd; default to cwd. Require a parseable `package.json` with at least one nonempty string-valued script.
- Script choices preserve JSON property order and display both name and command. Selection is always interactive; there is no direct script operand.
- Package manager detection only reorders a second interactive selector. Priority is `pnpm-lock.yaml`, `bun.lockb`, `bun.lock`, `yarn.lock`, then `package-lock.json`. Without a lock, order is pnpm, npm, Bun, Yarn.
- Retain Bun detection and execution for user projects. This does not authorize Bun anywhere in ycy's active build, hooks, source runtime, or dependency manifests.
- Spawn is argv-based, not a shell: `<pm> run <script>`, inherited stdin/stdout/stderr, selected project cwd. Intended passthrough appends `-- <args...>`. The CLI exits with the child exit code. Prompt cancellation exits 0.

### Confirmed parser defect

The current Commander declaration is `.command('run [path]').allowUnknownOption(true)` without variadic operands or pass-through mode. Diagnostics confirmed:

- `ycy run --flag` treats `--flag` as the optional **path**;
- `ycy run . --flag`, `ycy run --flag value`, `ycy run arg1 arg2`, and `ycy run -- arg1 arg2` fail before the action with excess-argument exit 1;
- therefore the action's passthrough calculation is not reachable for useful extra arguments.

Reproduce this observable parser behavior for the first Go release. The CLI prototype proves a future delimiter grammar is possible, but adding it is post-parity product work unless Cobra cannot reproduce the current rejection matrix.

### Other defects and tests

- Missing package messages say "current directory" even when an explicit path was supplied.
- A missing package-manager executable falls into the global error path without targeted remediation.
- No TTY guard or noninteractive selection path exists.

Required tests cover path resolution, package parse/schema failures, script filtering/order, every lockfile and priority collision, no-lock order, cancellation at both prompts, exact child argv/cwd/environment/stdio, external Bun selection, missing executable, exit-code/signal propagation, and the legacy passthrough rejection matrix.

Risk: **medium**, primarily CLI parsing and subprocess semantics.

## `zip [directory]`

### CLI and planning flow

- Directory defaults to cwd. `-w, --without-open` disables post-write reveal. `-d, --with-dir <dir>` prefixes archive entry names.
- The command is intentionally interactive and has no CI/noninteractive mode. It selects package when matched workspace packages exist, source directory, glob presets, and output filename.
- Workspace signals are `package.json` workspaces (array or `{packages}`), a lightweight line parser for `pnpm-workspace.yaml`, `turbo.json`, `nx.json`, and `packages/*`. Signals alone do not trigger package selection unless glob expansion finds packages. The workspace root is included as `.`.
- Project detection priority is uniapp H5, Vite, Webpack, root `index.html` frontend, then generic. Signals come from fixed config filenames, dependencies, and substring matches in scripts.
- Known output candidates include uniapp build/dev H5 paths and `dist`, `build`, `out`, `release`, and `public`. The engine scans directories to depth 2, ignoring VCS/editor/cache/node_modules paths, scores known names and `index.html`, then sorts by descending score and relative path.
- Recommendation confidence is high at score at least 92 with gap at least 18, medium at score at least 78 with gap at least 8, otherwise low. Package root is always a fallback; shallow directories are added for manual review when no alternative exists.
- Source selection is always confirmed even at high confidence. Glob choices are fixed presets: all, HTML, JS, CSS, `assets/**/*`, and `static/**/*`. Choosing all collapses other patterns; choosing none also falls back to all.
- Default output name is actually the input root's Git remote path (`owner-repo`, preferring `origin`) when available, then package name, then package/root basename. The README omits the Git-remote precedence.
- Output name keeps only the final slash-separated leaf, replaces illegal/control/path characters and whitespace with hyphens, collapses hyphens, strips leading/trailing dots, and falls back to `archive`.

### Archive behavior

- Output is `<selected-source>/<sanitized-name>.zip` and overwrites an existing file non-atomically.
- Bun glob collection is sequential by selected pattern and de-duplicates absolute paths. `**/*` includes regular non-dot files recursively. Diagnostics confirmed root matches for `**/*.html`, exclusion of dotfiles/nested dotfiles, and exclusion of symlinks. Empty directories and filesystem metadata are not archived.
- The output file itself is excluded when an old same-named archive was collected. Other existing archives remain eligible.
- Entry separators are normalized to `/`. Optional `withDir` is prepended verbatim. Every included file is read fully into memory; all entries are synchronously compressed together with fflate level 6, then the complete ZIP buffer is written.
- Reveal-file runs by default and failures are intentionally nonfatal. `-w` skips it.

### Defects and compatibility risks

- Directory existence/type is checked **after** the full interactive plan, so an invalid path can prompt before reporting failure.
- Invalid directory, collection errors, no matches, compression errors, and write errors call prompt cancellation and return normally, generally producing exit 0.
- `--with-dir` is unsanitized and can contain absolute or `..` archive paths. This zip-slip-producing defect remains first-release parity behavior and is a post-parity hardening priority.
- Whole-archive memory loading and synchronous compression can terminate or freeze the process on large inputs. A streaming Go implementation is permitted when it preserves the same selected entries, archive paths, metadata loss, output publication, and observable failure behavior; new limits are post-parity work.
- Output publication is not atomic. Permissions, modification times, symlinks, and empty directories are lost. The Go specification must state which metadata remains deliberately unsupported rather than inheriting library defaults accidentally.
- The pnpm workspace parser is not a YAML parser and project-script detection uses substring matching. These heuristics may be retained for compatibility initially, but the Go implementation should use a structured YAML parser and tests to ensure accepted legacy inputs do not regress.

### Dependencies and tests

Active dependencies are Bun Glob/File/Write/Spawn, Node fs/path, fflate, reveal-file, Clack, and Ansis. Required tests cover workspace formats/globs and root selection, all project-kind priorities, every known candidate and score tie, depth/ignore/fallback behavior, confidence thresholds, Git remote formats, filename sanitization, prompt/cancel sequence, glob normalization, dotfiles, symlinks, duplicate patterns, output self-exclusion, verbatim `withDir`, path separators, large inputs, legacy non-atomic overwrite, compression-readable contents, reveal failure, and legacy success-status error paths.

Risk: **high** because heuristics, archive semantics, resource usage, and interactive state all vary independently.

## Known defects retained for post-parity disposition

These findings are explicit so the port reproduces them deliberately and later hardening can address them separately:

| Area | Legacy behavior | First-release disposition |
| --- | --- | --- |
| Integer parsing | Accepts numeric prefixes and negative smart depth | Preserve and test; strict parsing is post-parity |
| `run` passthrough | Intended implementation is unreachable | Preserve the rejection matrix; a new `--` grammar is post-parity |
| Config overwrite | Add silently replaces same alias/profile | Preserve and test; confirmation is post-parity |
| Config malformed shape | May normalize valid-but-wrong JSON to empty and later overwrite | Preserve unless a direct-read probe proves a Go incompatibility |
| Config numeric environment | Accepts finite out-of-range/fractional values | Preserve and test; range tightening is post-parity |
| `rm` dangerous targets | Can recursively delete roots/home/cwd/repository | Preserve and test only in disposable roots; denial is post-parity |
| `rm` partial failures | Prints skipped but exits success | Preserve and test; corrected status is post-parity |
| ZIP error status | Many real failures exit 0 | Preserve and test; corrected status is post-parity |
| ZIP prefix | `--with-dir` permits unsafe archive paths | Preserve archive-name behavior in isolated tests; validation is post-parity |
| ZIP resources | Reads and compresses all bytes in memory | Streaming is implementation freedom only if all observable behavior remains compatible |

## Migration readiness summary

1. `export env` is the correct first vertical slice: small filesystem surface, useful interactive branch, deterministic machine-readable output, and one parser-parity risk.
2. Shared config persistence/crypto must precede any `config` or `git cm/fork` leaf. It deserves a deep module whose interface owns normalization, locking, atomic update, and secret compatibility.
3. `rm` is implementation-small but destructive; port its observed behavior with disposable-path tests and defer safety redesign until after parity.
4. `run` needs the CLI prototype result before implementation because its advertised passthrough path is currently broken.
5. `zip` should follow `run` and Git basics. Keep discovery/planning separate from archive execution; tests should use those interfaces rather than terminal snapshots.

This inventory resolves the local fact-finding ticket. It does not select Go libraries, approve the defect corrections, or implement any command.
