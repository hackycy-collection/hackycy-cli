# Diff command compatibility inventory

Inventory date: 2026-08-22
Legacy baseline: `78358c0201b71891e36603d6abb8d7c87d54ad57`
Scope: `ycy diff`, its immutable in-memory Comparison Workspace, REST/SSE/MCP protocols, and the active Diff React application.

## First-release scope

This inventory records observed facts for a parity-first port. The first Go release retains the Bun CLI, filesystem, REST, SSE, MCP, and React-facing behavior even where a section identifies unsafe, unbounded, or ambiguous policy. Those findings are post-parity backlog. Only a failed focused Go implementation or protocol test may create a narrow compatibility exception.

## Contract classification

This inventory separates the observed implementation into three classes:

- **Compatibility contract**: command and option names, defaults, directional comparison meaning, snapshot and query semantics, protocol fields and errors, security headers, resource limits, and core browser interactions that users or machine clients can rely on.
- **Implementation freedom**: Go package internals, traversal scheduling that does not change results, snapshot data structures, exact human-facing styling and wording, Vite-emitted asset names, and equivalent text-diff hunk choices where the result remains a valid bounded analysis-only Unified Diff.
- **Legacy defect or unresolved policy**: unsafe, contradictory, unbounded, or under-specified behavior retained as an explicit first-release parity fact and deferred for later correction.

The Go test suite must encode the selected contracts directly while consulting `legacy/bun/`. It must not invoke Bun or maintain a second legacy runner or separately owned golden corpus. Small inline compatibility tables for glob, gitignore, encoding, JSON, and protocol behavior are required because the replacement libraries otherwise choose behavior implicitly.

## Verification baseline

The complete current Diff suite passed during this inventory:

```text
64 pass, 0 fail, 187 assertions
src/commands/diff/index.test.ts
src/commands/diff/run.test.ts
src/commands/diff/workspace.test.ts
src/commands/diff/server.test.ts
src/commands/diff/mcp.test.ts
src/commands/diff/web/lib/content-cache.test.ts
```

A standalone Bun executable also built successfully. A production probe confirmed that its HTML shell and fallback responses use `Cache-Control: no-store`, emitted JS/CSS use `public, max-age=31536000, immutable`, `--port 0` prints the selected port, and `SIGINT` exits 0 after server shutdown. Source-mode Bun HTML bundling did not expose the same response headers in the probe, so production asset behavior is not proved by the current source-mode server tests. The Go/Vite acceptance gate must exercise the built standalone artifact rather than infer cache behavior from development mode.

Current coverage is unusually strong for workspace semantics, but it does not establish exact CLI parsing, arbitrary filename behavior, invalid glob parity, complete gitignore grammar, native Windows identity/no-follow behavior, REST Host/DNS-rebinding protection, unauthenticated public exposure, Blob memory bounds, SSE backpressure, or browser component flows.

## Domain contract

- A **Comparison Workspace** fixes one resolved Baseline Directory and one resolved Target Directory for the process lifetime. MCP and HTTP callers cannot select another path.
- A **Comparison Path** is a case-sensitive relative path represented with `/`, even on Windows. It is never accepted from a content caller and is not a rename identity.
- A **Comparison Snapshot** is an immutable, in-memory publication. It contains metadata and indexes, not complete contents or precomputed browser patches.
- A **Comparison Entry** is one file or symbolic link at an exact Comparison Path. Its Baseline and Target Entry States are independent.
- A **Comparison Issue** is a visible failure to establish comparison truth for one path. It is not silently treated as Added, Deleted, Modified, or Unchanged.
- A **Refresh** builds privately and atomically replaces the current snapshot only after discovery, comparison, source revalidation, and publication all succeed.
- Snapshot IDs, Entry IDs, and cursors are read references. Entry IDs are positive integers in the current JSON contracts, but their meaning is snapshot-local and opaque to clients.

There is no persistent Diff server state. Snapshots disappear on process exit. The browser persists only presentation preferences and panel layout in local storage; it must not persist filesystem paths or file contents. Consequently, Diff adds no file-format migration work.

## CLI and process lifecycle

### Command surface

```text
ycy diff [options] <baseline-directory> <target-directory>

-p, --port <number>       default 1205
    --public              default false
-x, --exclude <glob>      repeatable, default []
    --no-gitignore        default is to use Target .gitignore files
```

- Both directory operands are required and excess operands fail through the global parser. Options may be interspersed according to Commander behavior.
- The port parser accepts only ASCII decimal digits and then requires the result to be at most 65535. It accepts `0` and leading zeroes; it rejects signs, whitespace, decimal points, numeric prefixes, and non-ASCII digits. Invalid input exits 1 with either `'<value>' is not a valid port` or `Port must be between 0 and 65535`.
- Repeatable exclusions retain CLI order, although current matching is an OR and normally makes order irrelevant. There is no `--address` option and no environment-variable configuration for Diff.
- Both operands are resolved relative to cwd and passed through `realpath`. Missing/inaccessible paths, non-directories, and equal resolved path strings fail before the server starts. The workspace resolves them a second time and records each root's device/inode identity.
- Equal roots are rejected; nested roots are deliberately accepted. The Go port must test the resulting directional scope rather than add an undocumented nesting rejection.
- Default binding is `127.0.0.1`; `--public` binds `0.0.0.0`. Port `0` asks the kernel for an available port and all printed URLs use the selected port.
- The command does **not** launch or attempt to launch a browser.

### Human output and URLs

After binding, the current command prints:

```text
Directory diff: <local URL>
MCP endpoint:   <local URL>/mcp
[Network: <URL> ...]
[Network MCP: <URL>/mcp ...]
Baseline: <resolved absolute path>
Target:   <resolved absolute path>
```

- Local mode uses `http://127.0.0.1:<port>`. Public mode uses `http://localhost:<port>` locally and enumerates every non-internal IPv4 interface as a Network URL and Network MCP URL.
- IPv6 addresses are not printed. Interface order and duplicate handling follow Node's network-interface enumeration and are not a stable presentation contract.
- Exact labels are useful but human-facing spacing and ordering may change. Printing the usable browser URL, MCP URL, selected port, and both resolved roots before waiting is the compatibility requirement.

### Startup, signals, and exit meaning

- The HTTP server starts before the first Refresh. Initial Refresh is queued asynchronously, so callers can observe `idle`, `discovering`, and later phases.
- Initial Refresh failure leaves the server running in `error`; it does not terminate the CLI. A later Refresh may retry while the previous snapshot, if any, remains authoritative.
- `SIGINT` and `SIGTERM` each have a one-shot handler. It cancels the active server-owned Refresh, force-stops the Bun server, resolves its lifecycle promise, then calls `process.exit(0)`.
- Startup/parser/root/bind failures reach the global exit-1 handler. Ordinary Comparison Issues do not make the process fail.
- Go must centralize signal ownership at `cmd/ycy`, cancel a context, stop the server, and return exit meaning through one composition-root path. Command-internal `os.Exit` is not a compatibility requirement.

## Directional comparison and scope

### Status semantics

- **Added**: present only in Target.
- **Deleted**: present only in Baseline.
- **Modified**: present on both sides but kind, symlink target, size, or bytes differ.
- **Unchanged**: same kind and identical bytes or stored symlink target.
- Timestamps and permission/mode bits do not affect status. There is no rename or similarity detection; a case-only rename is Deleted plus Added where the filesystem permits both names.
- Directories organize the lazy tree but never receive independent entries or statuses. Empty-directory-only changes are invisible.
- Regular files and symbolic links are the only supported entry kinds. Sockets, devices, FIFOs, and other kinds become Comparison Issues.

### Hard exclusions

Hard exclusions win over every rule and apply at every depth. Despite README headings that describe macOS and Windows groups, the implementation applies all of these rules on every operating system:

- `.git` with exact case, whether file, directory, or symlink.
- `.DS_Store` and any basename beginning `._`, exact case.
- directories named `.Spotlight-V100` or `.Trashes`, exact case.
- files/symlinks named `Thumbs.db`, `ehthumbs.db`, or `Desktop.ini`, case-insensitively.
- directories named `$RECYCLE.BIN` or `System Volume Information`, case-insensitively.

The cross-platform implementation behavior is the migration baseline unless the corrected-contract decision deliberately narrows it. Dependencies, build output, editor metadata, and project-specific paths must not be added as hard exclusions.

### Target `.gitignore` policy

- Only Target `.gitignore` files define ignore scope, and that one hierarchical rule set is applied to both trees. A Baseline `.gitignore` is an ordinary compared file and never controls scope.
- Matchers are collected breadth-first from included real Target directories. Rules are relative to the directory containing each `.gitignore`; ancestor matchers apply in order and later nested negation can clear an ancestor file rule when traversal reached that directory.
- If Target has no corresponding Baseline-only subtree, that subtree still inherits the nearest Target ancestor rules. A directory ignored before descent cannot contribute a nested `.gitignore`.
- `--no-gitignore` disables all Target matcher discovery but does not disable hard or explicit exclusions.
- Missing `.gitignore` and a directory at the `.gitignore` path are ignored. Other read failures create an issue such as `Target Directory ignore rules could not be read (<code>)`, block that complete directory subtree on both sides, and fail closed.
- The current npm `ignore` library defines the actual grammar, including hierarchy and negation. Git's documentation expresses the intent, but a Go matcher cannot be assumed equivalent to either Git or npm `ignore` without differential tests.
- Current reads use Node's non-fatal UTF-8 string decoding. Invalid byte sequences can therefore become replacement characters rather than an explicit issue.

### Explicit exclusion glob policy

- Every `-x/--exclude` pattern becomes a `Bun.Glob`. Each Comparison Path is tested as written; a directory is additionally tested with a trailing `/`. A match on any pattern excludes that path and, for a directory, prevents descent on both sides.
- Matching uses case-sensitive POSIX-style Comparison Paths. Filesystem separators are introduced only when opening a local path.
- Bun semantics are not shell `filepath.Match` semantics. Observed compatibility vectors include: `*.tmp` matches only at the root; `**/*.tmp` matches root and nested files; `foo/**` matches `foo/` and descendants but not bare `foo`; brace alternatives work; backslash escapes metacharacters; an invalid `[` pattern silently matches nothing.
- Leading `!` is especially non-obvious: Bun treats it as glob negation. `!foo` matches everything in the tested set except `foo`, and because matches are exclusions it effectively excludes the complement. `!!foo` matches `foo`. This is not gitignore re-inclusion and must not leak directly from Doublestar defaults.
- The selected pure-Go baseline is `doublestar/v4` behind a command-owned adapter. The adapter must fix `/` protocol paths, directory/trailing-slash behavior, negation, invalid-pattern disposition, dot names, escaping/classes/braces, duplicates, and Windows input normalization before Diff consumes it.

Hard exclusions, gitignore, and explicit exclusions are checked in that order. A later mechanism cannot re-include a path excluded by an earlier one.

## Filesystem safety and comparison algorithm

### Discovery and identity

- Target ignore discovery precedes concurrent Baseline/Target discovery. Directory reads are bounded at 16 concurrent operations and operate breadth-first by directory level.
- Discovery uses `lstat`; symbolic links are recorded with `readlink` and are never traversed. Broken and cyclic links are valid entries.
- Each source record captures kind, size, device, inode, mtime, and ctime; symlinks also capture the stored target. Root identities capture device and inode.
- Read/inspection/readlink failures become side-labelled Comparison Issues. Messages merge with `; ` when both sides or ignore discovery report the same path.
- The union of Baseline paths, Target paths, and issue paths is sorted with JavaScript's default case-sensitive UTF-16 string order. Entry IDs are the resulting contiguous 1-based positions. Go byte ordering can differ for supplementary Unicode, so ordering must be made explicit before IDs are generated.

### Byte comparison and publication

- One-sided entries, size differences, kind changes, and symlink-target differences are classified without loading file contents.
- Same-size regular files are opened no-follow where the platform exposes `O_NOFOLLOW`, checked against the captured fingerprint, and compared in 64 KiB chunks. Comparison stops on the first byte difference.
- At most eight paths compare concurrently. Progress counts the maximum bytes read from either side for each paired chunk.
- A comparison error refreshes both source fingerprints and retries once. A second failure becomes a Comparison Issue unless cancellation caused it.
- After every classification completes, all retained source records are re-statted with concurrency 16 and both fixed roots are revalidated. Any one-sided or otherwise retained entry mutation at this stage fails the **whole Refresh** rather than publishing mixed metadata.
- Root replacement, including replacement by a symlink, fails every later Refresh while preserving the previous snapshot.
- Cancellation and failure never partially publish. A successful publish generates a UUID snapshot ID and UTC millisecond ISO timestamp, swaps one reference, then exposes phase `ready`.

### Stable content reads

- Content reads resolve only a snapshot-recorded root plus Comparison Path. They never accept an arbitrary path.
- A regular file is opened with read/no-follow flags, its real path and fingerprint are checked before reading, the complete requested bytes are read, and both checks repeat afterward. Any error or mismatch returns `stale`; external replacement bytes are not returned.
- `O_NOFOLLOW` is absent on some platforms, especially Windows. The realpath/fingerprint fallback and device/inode fields are not a sufficient Go design by themselves. Native Windows tests must prove file-ID checks and symlink/reparse-point containment using opened handles.
- The current Target `.gitignore` read is not held to the same rule: it can follow a `.gitignore` symlink outside Target, and rules can change between matcher collection and the later entry fingerprint. This can publish a scope inconsistent with the snapshot's `.gitignore` Entry State and is a security/consistency defect, not a requirement.
- Unix filenames are arbitrary byte sequences, while the current Node string APIs and JSON protocol assume Unicode Comparison Paths. The Go port must select and test a visible fail-closed policy for invalid UTF-8 instead of accidentally changing path identity or emitting invalid JSON.

## Snapshot state and queries

### State lifecycle

Workspace phases are exactly:

```text
idle -> discovering -> comparing -> publishing -> ready
                                      |            |
                                      +-> error    +-> subsequent Refresh
                                      +-> canceled
```

- State includes optional `snapshotId`, `error`, and progress. A failed or canceled replacement retains the prior `snapshotId`; without a prior publication those fields are absent.
- Progress contains `discoveredEntries`, `comparedEntries`, optional `totalEntries`, `comparedBytes`, optional `totalBytes`, and `issues`. Discovery/comparison progress publishes at most every 250 ms plus forced phase boundaries.
- Observers receive the current state immediately on subscription. Returned state/progress objects are copies.
- Only the newest published snapshot is addressable through the workspace. During Refresh the previous one remains readable. Successful replacement immediately invalidates every old HTTP/MCP Snapshot ID, Entry ID, and cursor.
- In-process objects are frozen against ordinary mutation; query results clone mutable directory counts. This is an internal safety property that Go should reproduce through ownership rather than reflection tricks.

### Flat listing

- Default page size is 100 and maximum effective size is 500. REST rejects limits below 1 but lets the workspace clamp values above 500; MCP validation rejects values above 500.
- Without an explicit status filter, Unchanged entries are excluded unless `includeUnchanged=true`; Comparison Issues remain included.
- Status filters are `added`, `deleted`, `modified`, `unchanged`, and `issue`. Kind filters are `file`, `symlink`, and `issue`; a file/symlink filter matches either side of a kind change.
- Path filtering trims the query and performs a case-insensitive substring match over the complete Comparison Path.
- A cursor is `base64url("entry-index:<zero-based absolute index>")`. It points to the next matching entry's position in the complete sorted array, not an offset in the filtered result.
- An anchor is a positive Entry ID, must itself match the active filters, and starts the page **at** that Entry. If both cursor and anchor are supplied, anchor wins.
- Cursors are opaque protocol values even though the current encoding is reversible. The encoding is not a compatibility requirement; correct continuation behavior and structured invalid-cursor errors are.
- The current cursor does not bind its filters and does not range-check the decoded number. Reusing it with different filters can skip results, and very large numeric payloads can yield an empty successful page instead of `invalid_cursor`. Preserve this defect for first-release parity; correction is post-parity.

### Tree and search

- `tree({path})` returns only immediate children. Directories come first, then entries, each group in Comparison Path order. An unknown directory returns an empty list.
- Directory nodes contain `kind`, basename `name`, Comparison Path `path`, descendant counts for all four statuses, and descendant `issues`. Entry nodes contain `kind`, `name`, `path`, `id`, `status`, and issue `message` where applicable.
- Workspace search trims and lowercases a substring, searches both synthesized directories and entries, sorts by path with directories first on a tie, caps at 200, and returns `truncated` rather than a cursor.
- REST search requires a nonblank `q`, accepts status filters, and validates a limit from 1 through 200. A blank query returns an empty result before validating other parameters.
- An omitted REST search status currently becomes an empty array and therefore returns no matches, although the README implies all statuses. Preserve and test this observed default for the first release; changing it is post-parity work.
- MCP `search_changes` is a different operation: it searches only changed entries through the flat index, never directory nodes or issues.

## Content and Text Difference contract

### Presentation classification

Entry detail classifies lazily:

1. either side is a symlink -> `symlink`;
2. path extension (case-insensitive) is `.avif`, `.gif`, `.jpeg`, `.jpg`, `.png`, `.svg`, or `.webp` -> `image`;
3. either regular file exceeds 20 MiB -> `oversized`;
4. a stable read is stale -> `stale`;
5. every present side strictly decodes as supported text -> `text`;
6. otherwise -> `binary`.

The image allowlist is extension-based, not content-sniffed. Symlink classification wins over image extension. A huge allowlisted image remains `image` and is not read during detail classification.

### Text decoding and limits

- Supported encodings are strict UTF-8 (with or without BOM), UTF-16LE with BOM, and UTF-16BE with BOM. `TextDecoder` strips supported BOMs. Invalid sequences are binary; valid UTF-8 containing NUL remains text.
- Missing sides return `missing`; non-files return `binary`; changed sources return `stale` where a read is attempted.
- Limits apply independently to each side:

| Result | Bytes | Decoded lines | Longest decoded line | Behavior |
| --- | ---: | ---: | ---: | --- |
| `ready` automatic | <= 2 MiB | <= 50,000 | <= 1 MiB | returns decoded text |
| `guarded` | <= 20 MiB | <= 200,000 | <= 1 MiB | metadata until `force=true` |
| `blocked` | > 20 MiB or > 200,000 | any | > 1 MiB | never returns text |

- Empty text has zero lines. Each LF or CR adds a line; CRLF is one delimiter, so a trailing newline produces a final empty line in the count.
- The line-length threshold currently counts JavaScript UTF-16 code units, not UTF-8 bytes or Unicode scalar values. A Go implementation must deliberately preserve or explicitly correct this edge rather than inherit `len(string)` or rune iteration by accident.
- A guarded request already reads and decodes the complete file to calculate line limits, and a forced request reads it again. There is no content-read concurrency limit.

### Server-generated Text Differences

- Text Differences exist only for Added, Deleted, and Modified entries. Issues, Unchanged entries, and unknown IDs map to entry-not-found at MCP.
- Context defaults to 3 and must be an integer from 0 through 20.
- Kind changes return `unavailable/mixed_entry_kinds`. Binary text returns `non_text`; automatic-limit Guarded or Blocked content returns `source_too_large`; changed files return `stale`.
- MCP never forces Guarded content. Therefore its input ceiling is the automatic 2 MiB/50,000-line/1-MiB-line boundary, not the 20 MiB browser-confirmed boundary.
- At most two calculations are active across the workspace. Further work immediately returns `server_busy`.
- The npm `diff` package receives a 5,000 ms algorithm timeout. Timeout/noncompletion returns `complexity_limit`.
- Output is an analysis-only Unified Diff. Headers are `baseline`, `target`, or `/dev/null`, never the Comparison Path. Whitespace is retained. Context hunk generation is deterministic for the current library but not promised to be directly applicable by Git or `patch`.
- A complete patch over 256 KiB returns `output_too_large` with added/deleted line counts and actual output bytes. A successful result includes encodings, counts, context, and patch.
- If both decoded sides are textually identical but their bytes differ, a Modified entry returns `no_textual_changes/encoding_or_bom_only` and both encodings.
- No Go line-diff package has yet been approved. The replacement must bound wall time **and** memory/work, preserve the result taxonomy and headers, and produce a valid deterministic Unified Diff. It must not leak a timed-out goroutine that permanently consumes a concurrency slot.

### Blob reads

- `blob(entryId, side)` is available for any regular file, not only entries classified as images. Missing sides return 404, stale sources 409, and non-file/issue data 415 through REST.
- Ready responses contain original bytes. MIME is determined only by the image extension allowlist; everything else is `application/octet-stream`.
- Allowlisted images use `Content-Disposition: inline`; other files use `attachment`. Both `filename` and RFC 5987 `filename*` parameters are emitted. SVG adds `Content-Security-Policy: sandbox; default-src 'none'; style-src 'unsafe-inline'`.
- Blob reads have no byte limit, no concurrency limit, no Range support, and load the complete file into server memory before the response is created. An arbitrarily large `.png` can therefore be fetched repeatedly, especially under `--public`. This is an unresolved resource-exhaustion defect, not a compatibility requirement.

## REST API contract

Every REST JSON response uses `Cache-Control: no-store`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`, and the restrictive application CSP. Errors have this versioned shape:

```json
{ "version": 1, "error": { "code": "UPPER_SNAKE_CASE", "message": "..." } }
```

Snapshot-bound routes require `snapshot=<current UUID>` and return `409 SNAPSHOT_CHANGED` when absent, unknown, or replaced.

| Method and route | Inputs and successful response | Important failures |
| --- | --- | --- |
| `GET /api/state` | `{version:1, workspace, snapshot?}` | other method: 405 `METHOD_NOT_ALLOWED` |
| `GET /api/events` | SSE state stream | other method: 405 |
| `POST /api/refresh` | `202 {accepted:true}` | active: 409 `REFRESH_ACTIVE`; hostile present Origin: 403 |
| `DELETE /api/refresh` | cancel server-owned Refresh, always 204 | hostile present Origin: 403 |
| `GET /api/entries` | `status`, `kind`, `path`, `cursor`, `anchor`, `limit`, `includeUnchanged`; `{entries,nextCursor?}` | malformed filter/cursor/anchor/limit: 400 |
| `GET /api/tree` | `path` defaults empty; `{children}` | stale snapshot: 409 |
| `GET /api/search` | `q`, repeated/comma statuses, limit 1-200; `{results,truncated}` | invalid filter/limit: 400 when query is nonblank |
| `GET /api/entries/:id` | browser-shaped detail | unknown/nonpositive ID: 400 or 404 |
| `GET /api/entries/:id/content/:side` | `side=baseline|target`, `force=true` opt-in; Text Content JSON | unknown/issue: 404; stale is a 200 body status |
| `GET /api/entries/:id/blob/:side` | original byte response | missing 404, stale 409, unavailable 415 |
| any `/api/*` otherwise | none | 404 `NOT_FOUND` |
| other GET path | embedded Diff HTML shell fallback | non-GET behavior is Bun bundle-owned |

REST Entry JSON intentionally flattens domain state for React:

- normal list/tree/search entries expose `id`, `path`, `status`, display `kind`, and optional `baselineSize`/`targetSize`;
- issues expose `id`, `path`, `status:"issue"`, `kind:"issue"`, and `message`;
- detail adds `presentation` and optional `baselineLinkTarget`/`targetLinkTarget`;
- state and all REST payloads use camelCase, while MCP uses snake_case.

Go's JSON encoder, router cleaning/redirect behavior, implicit HEAD handling, and HTML escaping differ from Bun. Route tests must capture status, content type, headers, omission versus null, field casing, timestamp formatting, and percent-encoded path behavior rather than relying on framework defaults.

## SSE lifecycle

- Subscription immediately emits the current `{version:1,workspace,snapshot?}` as one unnamed `data:` event followed by a blank line.
- Every later workspace state/progress publication emits the same complete shape. There are no event names, IDs, `retry`, comments, or heartbeat.
- Bun disables the request timeout for this route. Browser `EventSource` owns reconnect behavior.
- Stream cancellation unsubscribes the workspace observer. Server stop force-closes the connection.
- No Origin/Host validation, client limit, write deadline, slow-consumer policy, or explicit backpressure bound exists. A Go server needs flushing and an intentional SSE timeout/backpressure policy while retaining immediate full-state events.

## MCP Streamable HTTP contract

`/mcp` is stateless Streamable HTTP with JSON responses. Each request creates a fresh TypeScript transport and MCP server with `sessionIdGenerator: undefined` and JSON response mode. Clients receive no MCP session ID. The server identifies as `ycy-directory-diff` version `1.0.0` and publishes no resources, prompts, subscriptions, or MCP progress stream.

Tool order and names currently are:

1. `get_comparison`
2. `refresh_comparison`
3. `list_changes`
4. `list_issues`
5. `search_changes`
6. `get_text_diff`

| Tool | Input contract | Structured output |
| --- | --- | --- |
| `get_comparison` | empty | phase, optional error, optional snapshot with `snapshot_id`, roots, timestamp, counts, issues |
| `refresh_comparison` | empty | `accepted`, `already_running`; asynchronous and non-idempotent annotation |
| `list_changes` | snapshot; optional nonempty status/kind arrays, path, cursor; limit default 100/max 500 | changed Entry States and optional `next_cursor` |
| `list_issues` | snapshot; optional path/cursor; limit default 100/max 500 | issues containing path/message and optional cursor |
| `search_changes` | snapshot, trimmed nonblank query; optional status/kind; limit default 20/max 100 | changed Entry States plus `truncated` |
| `get_text_diff` | snapshot, positive safe integer Entry ID, context default 3/range 0-20 | snake-case ready/no-change/unavailable Text Difference |

- Every successful tool result includes a short text summary plus `structuredContent`. Change states are `{kind:"file",size}` or `{kind:"symlink",link_target}`.
- Read tools require the current snapshot. Replacement produces an MCP tool error with structured `snapshot_changed`; malformed list cursors produce `invalid_cursor`; unknown, Issue, or Unchanged Text Difference IDs produce `entry_not_found`.
- Tool errors set `isError:true`, provide `structuredContent.error.{code,message}`, and include text `<code>: <message>`. Zod/SDK owns input type/range validation.
- `get_comparison` continues to expose the previous snapshot while state is refreshing, canceled, or error. Clients poll after `refresh_comparison` and replace references only after successful publication.
- The accepted Go dependency is conditionally the official MCP Go SDK in stateless JSON mode, wrapped by ycy's own security and response-header adapter. SDK defaults, tool schema generation, and protocol error shapes are not presumed compatible; initialization and raw HTTP tests remain mandatory.

### MCP Origin and binding checks

- A request with no `Origin` is accepted for agent clients.
- A request with `Origin` must parse canonically, equal its own normalized origin string, equal the request URL origin derived from Host, and use an allowed hostname.
- `localhost` and the exact binding address are allowed. With `0.0.0.0`, any literal IPv4 or IPv6 hostname is allowed. No permissive CORS header is emitted.
- Cross-origin or disallowed-host requests receive HTTP 403 with a JSON-RPC error (`code: -32000`, message `MCP requests must be same-origin`) plus the common security headers.
- Hostname validation currently runs only when `Origin` is present. Agent requests without Origin remain deliberately possible, but public binding still has no authentication or TLS.

## Active React application contract

The frontend moves to the single active `web/` pnpm/Vite/React workspace; it is not archived with the TypeScript server. The following core flow must remain:

- Fetch `/api/state`, then maintain full workspace state through `/api/events`.
- Start or cancel Refresh through the two `/api/refresh` methods.
- Show all five status totals and let each status be independently filtered.
- Lazily request immediate directory children, synthesize a visible expanded tree, and virtualize rows at an estimated 30 px with overscan.
- Debounce server search by 120 ms, cancel superseded requests, send selected statuses, cap at 200, and show truncation.
- Keep multiple removable tabs, disambiguate duplicate basenames with paths, provide close/close-others/close-right/close-all actions, but mount only the active Diff renderer.
- Fetch and cache Entry detail and both text sides under snapshot-and-entry keys. A 32 MiB browser LRU evicts least-recently-used records and refuses a single record over budget.
- Clear the content cache, open tabs, and active selection whenever the snapshot ID changes.
- Use `@pierre/diffs` in browser workers for syntax highlighting, hunk calculation, split/unified display, wrapping, and line virtualization. Worker pool size is `min(hardwareConcurrency || 2, 4)`.
- Show dedicated issue, stale, binary, symlink, image, oversized, guarded-confirmation, and encoding/BOM-only states. Images use snapshot-bound Blob URLs.
- Under 900 px, move navigation to a Sheet and force Unified Diff. Above it, restore the stored desktop split/unified choice and a resizable 280-420 px sidebar.
- Persist only theme, desktop diff style, wrapping, ignore-whitespace rendering, and desktop panel layout. Ignore-whitespace affects browser hunks only, never server status/counts.
- Default theme is light and default desktop diff style is split. The server API does not own these presentation defaults.

The current frontend has no component or browser integration tests; only the content-cache unit has coverage. Fetch failures are largely collapsed into stale/empty presentation, malformed local-storage JSON can fail initial rendering, and Refresh promise failures are not surfaced consistently. These are UI robustness gaps, not backend response contracts.

### Vite/Go asset boundary

- Current Bun-specific HTMLBundle and `with { type: 'file' }` worker import cannot survive. Vite must own the `@pierre/diffs` worker edge and emit it under the common hashed asset tree.
- The active shell becomes `web/diff/index.html`; Vite emits `web/dist/diff/index.html` plus shared hashed `/assets/*`, all embedded once in the Go binary.
- Production HTML remains `no-store`; content-addressed assets remain one-year immutable. All responses retain `nosniff` and `no-referrer`; the HTML response retains the current restrictive CSP and no remote CDN/font/analytics dependency.
- The current broad fallback serves the Diff shell for arbitrary non-API paths. It can also turn missing asset-like paths or `/mcp/*` into HTML. Vite research already requires reserved `/api/*`, `/mcp/*`, and `/assets/*` namespaces to return real errors; whether other browser paths retain broad fallback is still a compatibility decision for the embed prototype.
- Vite-emitted filenames are implementation details. Shell route, resolved worker/asset URLs, MIME, caching, CSP, and absence of another application's shell are contracts.

## Security and resource boundary

The server exposes resolved root paths, comparison metadata, decoded source, original file bytes, Refresh/cancel mutations, and MCP analysis. It has no upload/edit/delete/merge route, but `--public` is not strictly read-only because callers can consume resources and replace/cancel snapshots.

Confirmed legacy risks are:

1. `--public` exposes the complete browser, REST, Blob, Refresh, and MCP surface on every IPv4 interface without authentication or TLS. The warning is documentation only.
2. REST and SSE do not validate Host or Origin. Refresh rejects a mismatching **present** Origin, but accepts no Origin and can accept an attacker-controlled same-origin Host under DNS rebinding. MCP has stronger browser checks but accepts Origin-less clients by design.
3. Blob and confirmed content reads are complete in-memory reads with no request concurrency budget. Blob has no size ceiling or Range path.
4. Target `.gitignore` can follow a rule-file symlink outside the workspace and can race its own later snapshot fingerprint.
5. Cursors are forgeable, filter-independent, and insufficiently range-validated.
6. SSE has no heartbeat, connection/slow-consumer limit, or backpressure policy.
7. Broad HTML fallback can mask reserved or missing resources.

Preserve these observed behaviors deliberately for the first Go release and keep their correction outside the port. They are not authorization to redesign authentication, limits, cursors, ignore-file handling, or routes before parity.

Security invariants that are already unambiguous and must remain are:

- default loopback binding and explicit `--public` opt-in;
- no permissive CORS;
- snapshot ID plus opaque Entry ID plus side enum for content, never caller paths;
- no file mutation/upload/arbitrary-path API;
- fail-closed fingerprints and fixed roots;
- `nosniff`, `no-referrer`, restrictive shell/API CSP, and sandboxed SVG;
- React text nodes/structured JSON for untrusted filenames and messages;
- no remote frontend dependency at runtime.

## Bun and TypeScript replacement matrix

| Current dependency/behavior | Go/Vite replacement boundary | Compatibility hazard |
| --- | --- | --- |
| Commander registration and port collector | CLI adapter selected by `Prove the Go CLI compatibility approach` | option interspersal, repeatable values, parser stderr/exit, port 0 |
| Node `realpath`, `lstat`, `open`, device/inode/ctime | command-owned filesystem/platform adapter | Windows file IDs/reparse points, no-follow race, timestamp precision, invalid filename bytes |
| `Bun.Glob.match` | `doublestar/v4` behind a Diff glob adapter | leading `!`, invalid patterns, root `**`, trailing slash, escaping, separators, dot paths |
| npm `ignore` | go-git gitignore candidate, still unapproved | hierarchy, anchoring, negation, escaped markers, spaces, Unicode/case/separators |
| JavaScript default sort and `toLowerCase` | explicit Comparison Path collation/search | UTF-16 versus UTF-8 ordering and Unicode case mapping can reorder IDs/results |
| `TextDecoder(..., fatal:true)` | explicit strict UTF-8/UTF-16 decoder | BOM handling, odd UTF-16, NUL, line-length units |
| npm `diff` asynchronous timeout | bounded Go line-diff engine, not selected | different hunks/headers/newline handling, timeout without memory/cancellation bound |
| Buffer base64url and byte length | standard Go encoders | cursor validation and UTF-8 output-byte counting |
| `crypto.randomUUID` and `Date.toISOString` | opaque UUID plus explicit UTC millisecond format | clients retain strings; Go RFC3339Nano output differs |
| `Bun.serve` route table and ReadableStream SSE | `net/http` command adapter | ServeMux cleaning, HEAD, flushing, timeouts, shutdown, header/status defaults |
| Bun HTMLBundle/file headers | Vite MPA plus shared `embed.FS` asset provider | fixed shell, common assets, missing asset fallback, source/production drift |
| TypeScript MCP SDK/Zod | official Go MCP SDK plus ycy validation/security wrapper | stateless mode, schemas/defaults, structured errors, Origin/Host behavior |
| Node network interface enumeration | `net.Interfaces` URL presenter | IPv4 filtering, address flags, ordering, port formatting |
| process signal handlers and direct exit | root cancellation/lifecycle owner | server drain, active Refresh cancellation, exit 0 |
| React 19, `@pierre/diffs`, TanStack Virtual, panels, Lucide | retained in active pnpm/Vite frontend | Bun worker import, API casing, CSS/assets, production CSP |

## Required Go and frontend tests

1. **CLI/lifecycle:** exact command/option surface; repeated exclusions; port grammar including 0/65535; missing/equal/non-directory/nested roots; bind conflict; printed local/public URLs; no browser launch; initial failure/retry; SIGINT/SIGTERM during every phase; exit 0 and no lingering server work.
2. **Glob adapter:** inline Bun-derived vectors for root/nested `**`, directory with/without slash, leading `!`/`!!`, escaped markers/metacharacters, braces/classes, invalid/empty patterns, dot names, case, duplicates, `/` and Windows input separators.
3. **Gitignore:** Target/Baseline asymmetry; root/nested hierarchy; anchoring; `**`; directory-only rules; parent exclusion and legal re-inclusion; escaped `!`/`#`; trailing spaces; CRLF/BOM/invalid UTF-8; Unicode/case; missing/directory/unreadable/symlink/mutating rule files; blocked subtree and merged issue messages on native Unix and Windows.
4. **Traversal/status:** all statuses; byte-versus-time/mode; different sizes; kind changes; broken/cyclic/outside symlinks; no traversal; hard exclusions on every OS; empty directories; case-only names where supported; nested roots; unsupported and unreadable entries; arbitrary/non-UTF-8 filename policy.
5. **Stability/publication:** file mutation before/during/after comparison; retry once; disappearance/replacement; one-sided mutation; root replacement/bind identity; stale content; cancellation/failure with and without a previous snapshot; observer state/progress throttling; immutable ownership.
6. **Ordering/queries:** exact path collation including supplementary Unicode; contiguous snapshot-local IDs; default/status/kind/path filters; kind changes; page 1/max/clamp; anchor-at-entry; cursor continuation, tamper/range/filter binding; old snapshot rejection; lazy tree aggregates/order; empty/unknown tree; REST and MCP search defaults/order/truncation.
7. **Text/content:** strict UTF-8/BOM and UTF-16LE/BE vectors; invalid/odd encodings, NUL, CR/LF/CRLF, empty/trailing lines, supplementary characters; every byte/line/single-line boundary on each side; force behavior; missing/symlink/binary/stale; observed legacy concurrent content behavior.
8. **Text Difference:** Added/Deleted/Modified/Unchanged/Issue/kind-change/binary/encoding-only; `/dev/null` headers; context 0/3/20 and invalid values; whitespace/newline/no-final-newline cases; deterministic valid patch; source/output/complexity bounds; two active jobs; busy result; request cancellation and no leaked worker/goroutine.
9. **Blob:** every allowlisted extension/case and mismatched content; ordinary attachment; filename quoting/Unicode/control characters; missing/stale/unavailable; SVG CSP; no external bytes after replacement; selected size/concurrency/Range behavior.
10. **REST raw HTTP:** every method/route/query combination; snapshot missing/old; repeated/comma/empty filters; limit/anchor/ID overflow; camelCase JSON and omitted fields; error versions/codes/status; content types/disposition/cache/CSP; Host/Origin/no-Origin/DNS-rebinding matrix for loopback and public binding; reserved and shell fallback paths.
11. **SSE:** immediate event, every phase and retained prior snapshot, framing and JSON, flush before later work, disconnect unsubscribe, EventSource reconnect, server shutdown, selected heartbeat/write-timeout/slow-consumer behavior, and bounded subscriber resources.
12. **MCP raw protocol and client:** initialization/server info; exact six tools, annotations, input/output schemas/defaults/ranges; text plus structured results; all tool result variants/errors; pagination; simultaneous refresh; polling; two stateless clients/no session ID; malformed JSON-RPC/content types/methods; response headers; complete Origin/Host/public/no-Origin matrix.
13. **Frontend/Vite:** production build resolves the Pierre worker and all local assets; three-entry isolation; fixed Diff shell and hashed common assets; built-binary cache/MIME/CSP/missing-route probes; state/SSE/refresh; lazy tree/search/filter; tabs/cache invalidation; all presentation states; long/Unicode paths; mobile/desktop threshold; themes/modes/wrapping; local-storage corruption; console/network errors; only one mounted renderer.
14. **Native/resource gates:** execute filesystem, symlink, identity, signal, HTTP, and MCP smoke suites on macOS, Linux, and Windows amd64/arm64 where behavior is platform-owned. Exercise the 100,000-entry/~20-GiB design target and adversarial concurrent content/SSE/MCP requests without memory proportional to total file bytes.

## Migration boundaries and readiness

1. Diff is a **high-risk** migration because its observable surface spans filesystem identity, two protocols, a long-running server, and an active worker-based React application. It should follow the shared web/HTTP foundation and lower-risk local commands.
2. Preserve a deep Comparison Workspace boundary: HTTP, MCP, and React must query immutable snapshot operations and must never construct filesystem paths. The exact Go package graph remains for `Choose the Go module seams and project layout`.
3. Keep explicit glob and gitignore adapters separate because their grammars and compatibility targets differ. Neither candidate library API should leak into the workspace interface.
4. Keep the MCP security wrapper outside the SDK and the web asset provider outside comparison truth. Keep browser diff presentation in React; do not precompute browser patches in Go.
5. No active code may import or dispatch into `legacy/bun/`, and no Go test may start the Bun service. Inline tests derived while reading legacy are sufficient.
6. Diff has no persistent-data migration. Snapshot replacement intentionally invalidates transient IDs and browser caches.
7. Implementation may begin after `Prove the Vite MPA to Go embed path` validates the Pierre worker/asset route. Public, route, cursor, ignore-file, and resource behavior follows this legacy inventory for the first release.
8. Completion requires the raw REST/SSE/MCP contract suite, native filesystem safety tests, frontend production-browser checks, and a built CGO-free standalone binary. A workspace unit suite alone cannot declare the command migrated.

This inventory resolves local fact finding. It does not select the line-diff or gitignore implementation, choose final module packages, approve post-parity hardening, or implement the migration.
