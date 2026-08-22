# FS command compatibility inventory

> **Roadmap transition-scope amendment (2026-08-22):** this inventory remains the factual source for normal FS behavior, but Bun-written FS session carryover is not a first-release compatibility gate. Go starts normally, logs diagnosable failures, and internal operators manage any old session residue. Only `config.json` has a Bun-written direct-read guarantee elsewhere in the roadmap.

This inventory records the observable `ycy fs` contract before the Bun CLI and TypeScript server are frozen under `legacy/bun/`. It covers `src/commands/fs`, the active React application that will move to `web/`, `src/shared/file-session`, and the 7-Zip preparation path. It is a migration specification input, not approval to preserve confirmed defects.

## First-release scope

The first Go release is nevertheless parity-first: actual Bun behavior remains the implementation and test baseline even where this inventory labels it defective. Safety, resource-limit, and product-policy corrections are post-parity backlog. Only a demonstrated in-scope Go capability, platform, dependency, or fresh-Go protocol mismatch may interrupt the FS port for a narrow compatibility decision; Bun-written FS state is not such a gate.

## Contract classification

The migration applies three levels of compatibility:

1. **Exact or protocol compatibility:** command and option names; documented defaults and precedence; route/method/query/header grammar; versioned JSON fields and omission rules; status/error taxonomy; session and cookie continuity; Workspace Path meaning; Range/cache behavior; task state machines; collision rules; archive support; and browser routes consumed by the retained frontend.
2. **Intent compatibility:** terminal colors/layout, URL note formatting, localized file collation, React spacing, wording that is not machine-consumed, and the exact pixels/bytes of a thumbnail may change while retaining the same capability and limits.
3. **Post-parity defect:** startup failures that exit 0, the ignored chunk-size flag, same-origin active HTML that can exercise management APIs, DNS-unaware remote-download checks, path race windows, unbounded request/SSE/download work, and unverifiable extracted 7-Zip state remain findings for later hardening, not first-release redesign requirements.

The first-release compatibility rule is to preserve existing accepted inputs, failures, and workflows. Any later safety correction is separate work after parity.

## Verification baseline

The local baseline was run on 2026-08-22 with Bun 1.3.14:

```text
bun test src/commands/fs src/shared/file-session
133 pass
0 fail
625 expect() calls
18 test files
20.31 s
```

The suite exercised real 7-Zip 26.02 extraction through the build-time preparation cache, HTTP servers on port 0, the Bun HTML bundle, image conversion workers, persistent sessions, filesystem operations, and unit-level React state. It did not build the standalone executable.

Additional read-only probes confirmed:

- a missing browse root, invalid account specification, and caught startup/configuration failures print a Clack cancellation on stdout and exit **0**;
- Commander exposes `--upload-chunk-size 12` as `uploadChunkSize`, while `resolveFsChunkedUploadOptions` reads `uploadChunkSizeMiB`; the CLI flag is therefore ignored, although `YCY_FS_UPLOAD_CHUNK_SIZE_MIB` works;
- the README refers to `bun run test:fs:performance`, but no such script or test exists.

These probes are evidence of legacy behavior. The two CLI defects require correction rather than Go emulation.

## Domain contract

Use these terms consistently in the Go specification and tests:

- **Browse Root:** the once-resolved directory exposed by one `fs` process.
- **Workspace Path:** a slash-separated, root-relative protocol path. It is never an OS path and never begins with `/`.
- **Entry Path:** a Workspace Path naming the final directory entry. It may name an internal symbolic link.
- **Management Mode:** the `--manage` capability gate for text writes, uploads, remote downloads, extraction, create, rename, copy, move, and permanent delete.
- **Login Mode:** the presence of at least one `--account`; all data/file/thumbnail routes then require the same persistent cookie session.
- **Caller Owner:** the authenticated raw session token in process memory, or the single literal `anonymous` owner without Login Mode, used to isolate chunked-upload sessions.
- **Staging Entry:** a hidden destination-local `.upload-*`, `.download-*`, `.edit-*`, or `.extract-*` entry that is not returned by directory listings.
- **Terminal Task:** a download or extraction task in `done`, `error`, or `cancelled` state.

The current ownership chain is:

```text
CLI registration
  -> process composition and signals
     -> root-confined workspace
     -> optional persistent authentication
     -> HTTP adapter
        -> direct and chunked upload
        -> remote-download queue
        -> single-worker extraction queue and 7-Zip
        -> thumbnail queue/cache
        -> embedded React file browser
```

The Go package graph is deliberately not selected here. HTTP, React, and task queues must continue to pass Workspace Paths into one root-confined filesystem boundary; none may assemble absolute paths.

## CLI and process lifecycle

### Command surface

The exact leaf is `ycy fs [options] [directory]`. There is no `serve` alias. `directory` defaults to the process working directory.

| Option or environment | Legacy contract |
| --- | --- |
| `-p, --port <number>` | `parseInt` collector, default `1204`; port 0 is accepted by the server and the actual assigned port is printed |
| `-a, --address <string>` | default `0.0.0.0`; passed directly to Bun as hostname |
| `-m, --manage` | default false; required for every mutation, download, and extraction route |
| `--safe-html` | default false; when true, HTML/XHTML originals become sandboxed attachments |
| `--account <username:password>` | repeatable, default empty; first colon separates username and password |
| `--session-dir <path>` | CLI value, then `YCY_FS_SESSION_DIR`, then the platform/root-derived default |
| `--session-idle-days <days>` | CLI value, then `YCY_FS_SESSION_IDLE_DAYS`, then 7; positive safe integer days |
| `--chunked-upload` | CLI value, then `YCY_FS_CHUNKED_UPLOAD`; env accepts `0`, `1`, `true`, or `false`, case-insensitively |
| `--upload-chunk-size <MiB>` | advertised integer 4-16, default 8; currently ignored due to the property-name defect |
| `YCY_FS_UPLOAD_CHUNK_SIZE_MIB` | working 4-16 integer chunk-size override when no working CLI value exists |

The shared permissive integer collector accepts numeric prefixes and negatives. Strict full-string/range parsing is already owned by `Prove the Go CLI compatibility approach`; FS must contribute port 0/65535, idle-day overflow, and 4/8/16 MiB boundary vectors.

### Account grammar

- The first `:` is the separator; later colons belong to the password.
- Username is 1-64 ASCII `A-Z`, `a-z`, digits, dot, underscore, or hyphen. Matching and duplicate detection are ASCII case-insensitive; the original configured spelling is returned to the browser.
- Password length is 5-256 JavaScript UTF-16 code units, despite the documentation saying characters.
- A missing separator, bad username, bad length, or duplicate case-folded username stops startup before the HTTP server.
- Account secrets remain process arguments and may be visible in history and process inspection. No new secret-input feature is in this migration's scope.

### Startup, output, signals, and exit meaning

- Startup prints the shared title and `File Browser` intro, then creates the workspace, authentication/session store, chunk configuration, and server in that order.
- The running note shows Local/Network URLs, resolved directory, bind, Management, Chunked uploads, HTML execution, Authentication, and optional Session storage.
- `0.0.0.0` presentation shows `localhost` plus every non-internal IPv4 network interface. Any other address prints one URL. The current presenter fails to bracket IPv6 literals; preserve that first-release output unless Go formatting cannot reproduce it.
- The command does not open a browser.
- `SIGINT` and `SIGTERM` share an idempotent stop path. Stop closes session ownership, waits for chunk operations, cancels downloads and extraction, terminates thumbnail workers, force-stops Bun HTTP, resolves `finished`, and prints `File Browser stopped.`
- There is no explicit shutdown deadline. A stalled chunk can delay stop until its five-minute write timeout.
- Workspace, account, chunk-config, and bind failures are caught, rendered as cancellation text, and return normally. This produces exit 0 and must become a nonzero Go result with the same useful cause.
- Uncaught failures still reach the root handler and exit 1.

## Workspace Path and filesystem contract

### Root and path grammar

- Startup applies absolute resolution and `realpath` once. A missing path is `NOT_FOUND`; a non-directory is `NOT_DIRECTORY`.
- Protocol inputs reject NUL, backslash, a leading slash, and a drive-letter prefix. A segment equal to `.` or `..` is forbidden. Empty segments are discarded, so `a//b` becomes `a/b`.
- Resource routes decode every nonempty URL path segment and then join with `/`. Malformed percent encoding is `400 INVALID_PATH`. A percent-decoded slash becomes a Workspace Path separator.
- Query paths come from WHATWG `URLSearchParams`; Go query decoding and router path cleaning must not be inherited without raw HTTP parity tests.
- Containment uses `realpath` plus a native path-prefix check. Internal links may be followed for reads and traversal; links escaping Browse Root are unavailable/forbidden.
- Listing an escaping, broken, unreadable, or special entry returns an `unavailable` row without exposing its link target. Direct access fails.
- The current check-then-use design reopens several names after containment checks. An ancestor can be swapped between `realpath`, `lstat`, `stat`, open, rename, copy, move, or delete. This race is not a compatibility requirement; Go needs descriptor/handle-relative containment and native reparse-point tests.
- Unix non-UTF-8 filenames are forced through Node strings and JSON and have no defined safe identity. Go must select a visible fail-closed policy rather than emit invalid JSON or let replacement-character aliases address the wrong entry.

### Listing and entry representation

`GET /api/directory?path=` returns `version:1`, `rootName`, normalized `path`, optional `parentPath`, `managementEnabled`, `maxUploadBytes:1073741824`, optional chunk capability, and `entries`.

Each entry exposes:

```text
name, path, kind, isSymlink, previewKind, extractable
optional size, modifiedAt, mimeType, syntaxLanguage,
browseUrl, fileUrl, thumbnailUrl, downloadUrl
```

- Kinds sort as directory, file, unavailable. Within a kind, JavaScript `localeCompare` uses case-insensitive comparison then a case-sensitive tie break. ICU/locale variation means the exact non-ASCII collation is not stable enough to become a new cross-platform promise.
- `modifiedAt` is UTC ISO text with JavaScript millisecond formatting. Optional fields are omitted, not `null`.
- Temporary names matching `.download-<uuid>.tmp`, `.extract-<uuid>.tmp[.outer]`, `.edit-<uuid>.tmp`, or `.upload-<uuid>.tmp` are hidden even when stale.
- Directories receive `/browse/<encoded Workspace Path>`.
- Files receive `/files/<encoded Workspace Path>` and the same URL with `?download=1`.
- Raster AVIF, GIF, JPEG, PNG, and WebP also receive `/thumbnails/<encoded Workspace Path>`; SVG never does.
- Archive recognition is filename-suffix based and case-insensitive.

### MIME and preview classification

- AVIF, GIF, JPEG/JPG, PNG, SVG, and WebP use an explicit command-owned MIME map.
- Everything else uses `Bun.file(path).type`, including parameters such as `text/plain;charset=utf-8` and `application/json;charset=utf-8`; empty detection falls back to `application/octet-stream`.
- Syntax language is derived from `@pierre/diffs` filename mapping. `.env` and `.env.*` force `dotenv`; a lowercased second lookup makes known uppercase extensions recognizable.
- Preview priority is image, syntax-known text, video, audio, PDF, structured/text MIME, then none.
- Directory listing does not sniff contents. The preview UI probes `/api/text` for files not already image/video/audio/PDF, allowing an extensionless file to become text after selection.

A deterministic Go MIME adapter must own the relevant extension table. `mime.TypeByExtension` can vary by platform and does not by itself reproduce Bun parameters or preview classification.

## HTTP adapter contract

### Common responses and access gates

API JSON uses:

```json
{ "version": 1, "error": { "code": "UPPER_SNAKE_CASE", "message": "..." } }
```

JSON responses carry `Cache-Control: no-store`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`, `Referrer-Policy: no-referrer`, and `X-Content-Type-Options: nosniff`.

- With Login Mode, `/api`, `/api/*`, `/files`, `/files/*`, `/thumbnails`, and `/thumbnails/*` require a resumable session. `/`, `/browse`, `/browse/*`, and emitted application assets remain public so the login form can load.
- Every mutation requires a present `Origin` exactly equal to the request URL origin and a Host hostname allowed by the configured bind.
- For `0.0.0.0`, allowed mutation hosts are `localhost` or any literal IPv4. For `::`, they are `localhost` or any IP literal. Loopback accepts its literal or `localhost`; another explicit bind accepts only its exact hostname text.
- Read routes do not validate Host or Origin. Login Mode is authentication, not TLS.
- Go must preserve explicit method errors. Standard `ServeMux` path redirects, implicit HEAD, multipart limits, and `http.ServeContent` multi-range behavior are not presumed compatible.

### Route matrix

| Method and route | Successful contract | Important failures/gates |
| --- | --- | --- |
| `GET /`, `/browse`, `/browse/*` | public React shell | other listed methods: 405 `METHOD_NOT_ALLOWED` |
| `GET /api/session` | disabled state, or authenticated/unauthenticated account state | stale cookie is cleared |
| `POST /api/session` | authenticate and set cookie | Origin, JSON media type/body/schema, 401 credentials |
| `DELETE /api/session` | revoke and clear cookie, 204 | Origin |
| `GET /api/directory?path=` | versioned listing/capabilities | path/workspace errors |
| `GET /api/text?path=` | ready/too_large/binary union | path/workspace errors |
| `PUT /api/text?path=` | conditional text save | Manage, Origin, media type, If-Match, 10 MiB, revision |
| `POST /api/upload?path=` | direct multipart publication | Manage, Origin, multipart, 1 GiB |
| `POST /api/uploads` | `201` chunk session | Manage, feature, Origin, JSON <=64 KiB, size >20 MiB |
| `GET /api/uploads/:uuid` | current confirmed offset/status | owner/session/UUID |
| `PUT /api/uploads/:uuid` | append one exact ordered chunk | Origin, octet stream, Content-Range/length, busy/timeout |
| `POST /api/uploads/:uuid/complete` | publish or replay completed result | Origin, complete/busy state |
| `DELETE /api/uploads/:uuid` | cancel, 204 | Origin, complete/busy state |
| `POST /api/operations` | per-item operation result | Manage, Origin, strict JSON command |
| `GET/POST/DELETE /api/downloads` | list, enqueue `202`, clear terminal `204` | Manage; mutations require Origin; `terminal=1` to clear |
| `GET /api/downloads/events` | immediate and later full task-list SSE | Manage |
| `POST /api/downloads/:id/cancel|retry` | task, retry `202` | Manage, Origin, state/id |
| `GET/POST/DELETE /api/extractions` | list, enqueue `202`, clear terminal `204` | same pattern; 1-100 paths |
| `GET /api/extractions/events` | immediate and later full task-list SSE | Manage |
| `POST /api/extractions/:id/cancel|retry` | task, retry `202` | Manage, Origin, state/id |
| `GET/HEAD/OPTIONS /files/*` | original bytes, range/cache/download/CORS | file/path; other method 405 |
| `GET/HEAD /thumbnails/*` | 160x160 WebP | format/size/pixels/queue/timeout; other method 405 |
| any other `/api/*` | none | 404 `NOT_FOUND` |

Operation/download/extraction/login JSON bodies rely on Bun's global request ceiling of 1 GiB plus 1 MiB; only chunk-session creation has a 64 KiB local bound. This is a confirmed resource defect.

## Authentication and persistent session contract

### Password and cookie behavior

- Startup hashes each configured plaintext password with Argon2id, memory cost 65,536 KiB and time cost 3. These Argon hashes are process-local and are **not persisted**.
- Unknown usernames verify against the first account's hash, using the same expensive path as a wrong password.
- Successful login issues 32 random bytes encoded as unpadded base64url (43 characters).
- Cookie name is `ycy_fs_session`; attributes are `HttpOnly; SameSite=Strict; Path=/; Max-Age=<idle seconds>`. There is no `Secure` attribute because this server is HTTP-only.
- Every successful authenticated protected request resumes the session, synchronously rewrites and fsyncs its record, extends idle expiry, and returns a refreshed cookie.
- Default idle lifetime is seven days with no absolute lifetime. The React app also calls `/api/session` every 24 hours while open.
- Logout, expiry, LRU eviction, a changed configured password, or missing account revokes the session. An observer closes the associated download or extraction SSE stream on revocation.
- Limits are eight sessions per account and 128 total; least recently accessed records are evicted first.
- There is no login rate/concurrency limit, role model, registration, password change, or persistent account database. Every account has the same read and management capability selected for the process.

### Existing on-disk format

The default directory is:

```text
<platform state root>/ycy/fs/sessions/<sha256(path.resolve(input directory))>
```

Platform state root is `%LOCALAPPDATA%` or `%USERPROFILE%/AppData/Local` on Windows, `$HOME/Library/Application Support` on macOS, and `$XDG_STATE_HOME` or `$HOME/.local/state` elsewhere. The hash uses the lexical absolute CLI directory before the workspace's `realpath`, not the resolved Browse Root.

The directory contains:

| Entry | Exact format |
| --- | --- |
| `.session-key` | 32 raw random bytes; mode 0600 where meaningful |
| `.session.lock` | one JSON line containing `{ "id": <UUID>, "pid": <positive integer> }`; O_EXCL ownership |
| `<tokenHash>.json` | one compact JSON object plus newline; filename is lowercase hex SHA-256 of the raw cookie token |
| `*.tmp-<UUID>` | atomic-write candidate; removed at next open |

Session JSON version 1 contains, in this order in current writes, `version`, `tokenHash`, `subject`, `revision`, `createdAt`, `lastAccessAt`, and `expiresAt`. Timestamps are valid ISO strings. The credential revision is unpadded base64url HMAC-SHA256 with `.session-key` over:

```text
lowercase_username + NUL + plaintext_password
```

Only the token hash, lowercase subject, HMAC revision, and timestamps persist. Raw tokens, plaintext passwords, and Argon hashes do not.

Open creates/chmods the directory to 0700, acquires the lock, validates the key length, removes interrupted candidates, prunes malformed/expired/misnamed records, chmods records to 0600, applies limits, and schedules expiry. A live PID blocks a second owner; an unreadable/stale owner is unlinked and retried once. PID reuse and file-only locking are residual limitations.

Go can technically read this format with standard SHA-256, HMAC, base64url, and JSON. That fact does not create a first-release carryover contract: do not detect, read, migrate, delete, or fixture a Bun-written FS session directory. Test resume, refresh, revocation, pruning, lock ownership, and native permissions only for state created by Go; ordinary startup diagnostics cover failures and internal operators handle residue.

## Original files, caching, and Range

- Original MIME comes from the listing rules. `inline` applies to `text/*`, `image/*`, `video/*`, `audio/*`, PDF, JSON, XML, JavaScript, XHTML, and JSON-LD unless `download=1` or safe HTML forces attachment.
- `Content-Disposition` includes both a percent-encoded quoted `filename` and RFC 5987-style `filename*=UTF-8''...`.
- Responses include weak `ETag: W/"<size>-<mtime milliseconds>"`, `Last-Modified`, `Accept-Ranges: bytes`, exact `Content-Length`, `Cache-Control: no-cache`, `nosniff`, and `no-referrer`.
- `If-None-Match` takes precedence over `If-Modified-Since`; exact listed weak tag or `*` returns 304.
- Exactly one `bytes=start-end`, `bytes=start-`, or `bytes=-suffix` range is supported. End clamps to size. Empty/invalid/unsatisfiable/multiple ranges return 416 and `Content-Range: bytes */<size>`.
- A valid range returns 206 for GET or HEAD. `If-Range` rejects weak tags; an exact strong-looking tag or sufficiently recent HTTP date permits range. The server's own ETag is weak, so replaying it in `If-Range` deliberately yields the full 200 response.
- An `If-Range` mismatch ignores Range and returns 200. Conditional 304 is checked before Range.
- Requesting a directory under `/files/*` redirects 302 to `/` or `/browse/<path>`.
- Without Login Mode only, original files and OPTIONS expose `Access-Control-Allow-Origin: *`; OPTIONS advertises GET/HEAD/OPTIONS and selected conditional headers. No other route enables wildcard CORS.

By default HTML and XHTML are inline with no CSP and execute as same-origin documents. SVG/XML receive a sandbox CSP. `--safe-html` turns HTML/XHTML into attachments and gives them the sandbox CSP. Because a default inline HTML document shares the API origin and authenticated cookie, it can call management endpoints when viewed. This is a critical legacy security defect and is not silently approved.

The weak validator is derived before body consumption, and the current path is not held by a stable opened handle. Replacement between `stat` and streaming can return bytes inconsistent with headers or outside the checked object. The Go file response must bind metadata and bytes to one safely opened handle.

## Text preview and conditional editing

### Read contract

- Maximum source size is exactly 10 MiB. Size is checked before and after reading. Larger files return `{status:"too_large",size,maxBytes}` without content.
- Supported text is strict UTF-8 with optional BOM, UTF-16LE with BOM, or UTF-16BE with BOM. Invalid sequences are binary. Valid UTF-8 containing NUL remains text.
- Ready returns decoded text, encoding, byte size, and lowercase hex SHA-256 revision of the original bytes.
- Binary returns only status and byte size.
- Content capability, not filename, controls this endpoint.

### Save contract

- Save is Management-only and requires exact same Origin, `Content-Type: text/plain` with no parameters other than optional `charset=utf-8`, and `If-Match` containing the raw 64-hex revision value.
- The request body must be strict UTF-8 and at most 10 MiB. Missing precondition is 428; stale source is 412; unsupported source is 409; oversized input/output is 413.
- Final symbolic links are readable but never editable. The server repeats root, regular-file, encoding, size, and revision checks independently of React.
- Saves serialize per resolved path inside one process.
- Output preserves source encoding and BOM, original mode bits, dominant line ending (CRLF or CR only when strictly more common; LF wins ties), and whether the source ended in any newline.
- Draft line endings are normalized. If the source ended in a newline, all draft trailing newlines collapse to exactly one; otherwise all draft trailing newlines are removed.
- A destination-local `.edit-<uuid>.tmp` is exclusively created, written, chmodded, fsynced, checked against the source again, then renamed over the file. Ownership, xattrs, ACLs, and the old inode are not explicitly preserved.
- Success returns revision, byte size, `modifiedAt`, and encoding. JSON serializes `Date` as UTC milliseconds.

The UTF-16 encoder iterates JavaScript UTF-16 code units. Go must test supplementary and lone-surrogate drafts rather than accidentally changing byte output through rune iteration.

The React editor appears only for a ready, non-symlink file in Management Mode. It retains draft/conflict state in memory, supports Ctrl/Cmd+S and dirty-navigation guards, and on 412 offers only Reload remote or Download draft. It has no force overwrite, save-as, new-file, clipboard-copy, or persisted draft flow.

## Management operations and publication

### Operation grammar

`POST /api/operations` accepts one strict tagged JSON command:

```text
create-directory { parentPath, name }
rename           { path, newName }
copy             { paths[1..1000], destinationPath }
move             { paths[1..1000], destinationPath }
delete           { paths[1..1000] }
```

Paths/names are capped at 4096 characters by the HTTP schema. A name cannot be blank after trim, `.`, `..`, contain slash/backslash, or NUL. Operation names preserve otherwise-valid leading/trailing spaces; upload/download filename validators trim them.

- Browse Root cannot be renamed, copied, moved, or deleted.
- Create, rename, and move reject collisions. Rename of the same name therefore fails at the workspace layer, although React suppresses it.
- Copy chooses `name`, then `name (1)` through `(9999)`, preserving the final extension for files. Direct/chunked upload and remote download use the same file rule.
- Extracted directories use the archive base, then `base (1)` through `(9999)`.
- Recursive copy does not dereference symbolic links. A directory cannot be copied or moved into itself.
- Move is native rename only; a cross-filesystem move fails rather than falling back to copy/delete.
- Delete is permanent. It `lstat`s and removes the final entry itself, so deleting a symlink does not delete its target.
- Batch work is sequential and nontransactional. The HTTP response remains 200 and reports an ordered `ok` or structured `error` item per input, so partial success is part of the contract.

### Direct upload

- `POST /api/upload?path=` accepts one multipart field named `file`; additional fields/files are ignored.
- Bun's server body ceiling is 1 GiB plus 1 MiB and the file object limit is exactly 1 GiB.
- Upload filename is trimmed, must be one nonempty safe name, and is written to a destination-local `.upload-<uuid>.tmp`.
- Publication hard-links staging to the first unused collision-safe name, then unlinks staging. Existing names are never overwritten.
- Failure removes staging. Abnormal termination can leave a hidden direct-upload file indefinitely; direct upload itself does not run the 24-hour stale cleanup.

Hard-link publication, rename semantics, mode behavior, and failures differ across Windows/filesystems. Go must prove native behavior and select a safe fallback without reducing no-overwrite atomicity.

## Chunked upload protocol

Chunking is advertised only when both Management Mode and the feature are enabled. The frontend selects it only for files **larger than** 20 MiB; exactly 20 MiB uses direct upload.

Creation JSON is `{directoryPath,filename,size}` where size is a positive safe integer and must exceed 20 MiB. The server returns:

```text
uploading: id, status, size, uploadedBytes, chunkSizeBytes
complete:  same fields plus result {filename,path,size}
```

- Session ID is UUID. It is process-local and owner-bound; a wrong owner is indistinguishable from missing.
- At most three uploading sessions per owner and 100 retained sessions globally. Completed records count toward the global limit for five minutes.
- Uploading sessions expire after 30 inactive minutes. GET refreshes inactivity; create and successful writes also touch it.
- Each PUT requires `application/octet-stream`, exact `Content-Range: bytes start-end/total`, matching `Content-Length`, current confirmed start, the session total, and a body no larger than configured 4-16 MiB chunk size.
- Only one write or completion runs per session. A write times out after five minutes.
- Partial transport failure does not advance `uploadedBytes`; retry starts from the server-confirmed offset and overwrites the same staging range.
- Completion requires every byte, serializes concurrent retries, publishes once, and remains idempotently readable/replayable during retention.
- Cancellation cannot interrupt an active write/completion through the DELETE endpoint; it returns busy. An idle cancellation removes staging. Cancelling a completed session is a no-op.
- Normal server stop waits tracked create/write/complete work, aborts all unpublished staging, and clears records.

Before staging, the workspace serializes capacity checks per process, groups reservations by filesystem device, subtracts all active declared sizes, and preserves 10% of current available bytes capped at 1 GiB. Staging is exclusively created beside the destination. Reservations are process-local and do not protect against another server or external disk consumption.

The browser sends sequential chunks, retries a failed chunk three times with 250/500/1000 ms backoff, reads confirmed status between retries, retries completion three times, and best-effort DELETEs on final error/cancellation. Selected browser files and sessions do not resume after page/process restart. Multi-file upload runs three files concurrently.

## Remote download contract

### URL, response, and filename behavior

- Requests contain strict `{url,directoryPath,filename?}` with URL <=8192 and path/name <=4096.
- URL must parse under WHATWG rules, use HTTP(S), and contain no username/password.
- Literal `localhost`, `*.localhost`, `*.local`, and a command-defined set of literal private/reserved IPv4/IPv6 addresses are rejected.
- Hostnames are **not resolved before validation**. Bun fetch selects the actual address, so a public-looking hostname can resolve or rebind to loopback/private services. Every redirect URL is rechecked, but only with the same literal-host logic.
- Redirect statuses 301/302/303/307/308 are followed manually, at most five hops. Missing Location, other non-2xx, or missing body fails the task.
- Header acquisition and every body-read idle period have a 60-second timeout; there is no total duration limit.
- Bodies stream to `.download-<uuid>.tmp` and publish with collision-safe hard linking. There is no maximum byte count, disk reservation, or declared/actual Content-Length consistency check.
- Optional request filename wins. Otherwise a limited UTF-8 `filename*`, then quoted/plain `filename`, then decoded final URL segment, then `download` is used. The final component is trimmed and path separators are stripped/rejected.
- A Content-Length is exposed only when no Content-Encoding is present and it is a nonnegative safe integer. Progress and average bytes/second update at most once per 250 ms plus forced state boundaries.

### Queue and task state

- At most two tasks run, at most 100 wait, and terminal history is pruned oldest-first toward 100 records. Tasks list newest first.
- State is `queued -> running -> done|error|cancelled`. Fields include UUID, normalized URL, directory, optional filename, downloaded/total bytes, percent, speed, destination, and UTC millisecond timestamps.
- Cancel removes a queued task from the queue or aborts the active fetch/stream and staging. Retry is allowed only for error/cancelled and creates a new UUID while retaining the old record.
- Clear requires `terminal=1` and removes only terminal tasks. Process restart loses all tasks/history and does not resume bytes.
- Subscription immediately publishes the complete current list; later snapshots are complete replacements.

The DNS/connection gap is a documented SSRF defect, but its observable behavior remains first-release parity. Resolved-address, redirect, proxy, and DNS-rebinding hardening belongs to post-parity work unless `net/http` proves unable to reproduce a specific accepted legacy request.

## Archive and 7-Zip contract

### Accepted names and queue behavior

Case-insensitive supported suffixes are:

```text
.7z .zip .rar .tar .gz .gzip .bz2 .bzip2 .xz .zst .zstd
.cab .arj .lzh .lha .cpio
.tar.gz .tar.bz2 .tar.bzip2 .tar.xz .tar.zst .tar.zstd
.tgz .tbz .tbz2 .txz .tzst
```

The longest suffix is removed for the destination. Empty, `.`, or `..` bases become `Extracted`.

- Extraction requests contain 1-100 Workspace Paths. Exactly one task executes; up to 100 wait and at most 100 records are retained.
- Tasks expose UUID, archivePath, status, optional progress/inspection/destination/timestamps/error. Lists are newest first.
- Cancel/retry/clear/SSE follow the remote-download state pattern. Retry creates a new task.
- Cancellation or shutdown kills the active child and removes staging. Queue and history are process-local.

### Pinned runtime and state

Runtime is official 7-Zip 26.02 for all six release pairs:

| Target | Embedded runtime |
| --- | --- |
| macOS arm64/x64 | upstream mac archive, `7zz`, `License.txt` |
| Linux arm64/x64 | matching upstream Linux archive, `7zz`, `License.txt` |
| Windows arm64/x64 | matching installer, `7z.exe`, `7z.dll`, `License.txt` |

The artifact archive SHA-256 values are pinned in `archive-manifest.ts`. The preparation script downloads from the official GitHub release, verifies the archive, extracts named files, and supplies them as Bun compile entrypoints. Preparing a Windows installer on a non-Windows host requires `7zz` or `7z` in PATH.

At runtime embedded files are materialized under `<platform state root>/ycy/7zip/26.02/`. Source mode falls back to `7zz`, then `7z` in PATH when no complete embedded set exists.

There is an important implementation/documentation mismatch:

- the build cache reuses extracted runtime files by existence, without per-file digest checks;
- the manifest pins archive digests but not extracted-file digests;
- runtime reuses any existing state file without verifying bytes, type, owner, or permissions;
- runtime writes have no protected directory/lock contract.

The accepted product intent is verified embedded 7-Zip. The Go build must pin and verify the downloaded artifact and extracted runtime manifest, publish state atomically with restrictive native permissions, and verify safely before execution. Corrupt state is regenerated or fails detectably; users never repair it manually.

### Inspection and extraction

- Inspection invokes `7z l -slt -sccUTF-8 -- <archive>` with C locale, parses entry `Path`, `Size`, and `Encrypted`, sums safe-integer unpacked bytes, counts entries, and rejects encrypted/multi-volume input.
- Capacity preserves 10% available bytes capped at 1 GiB and, where reported, 10% free inodes capped at 1024.
- Normal extraction invokes `7z x -y -sccUTF-8 -bso0 -bse2 -bsp1 -o<staging> -- <archive>` and maps progress.
- Compressed-TAR names first extract to `.outer`, require exactly one recursively found `.tar` file, inspect capacity again, then extract the inner TAR using progress spans 0-35 and 35-100.
- 7-Zip warning/command/memory/interruption and recognized stderr classes map to normalized workspace errors; raw stderr remains only as an internal cause.
- Extraction relies on pinned 7-Zip's default path sanitation and does not pass `-spf`/`-spf2`. Backslash entry behavior is intentionally delegated to 7-Zip and can differ by OS.
- After extraction, the workspace recursively accepts only directories, regular files, and non-broken relative symlinks whose lexical and real targets remain in staging. Special files, absolute/escaping/broken links fail publication.
- Staging is beside the archive, the original archive remains, and a same-parent atomic rename publishes to the first collision-safe directory.
- Hidden extraction directories older than 24 hours are removed on a later extraction; matching regular files are not removed.

Real archive tests cover 7z, ZIP, RAR, TAR, gzip, bzip2, XZ, Zstandard, Unicode/spaces/leading dash/empty archives, encrypted/damaged/multipart/capacity/path sanitation, and normalized failures. Several shell/special-file cases simply return on Windows and therefore do not establish native Windows parity.

## Thumbnail service contract

- Input route is extension/MIME allowlisted to AVIF, GIF, JPEG, PNG, and WebP. SVG is rejected before conversion.
- Maximum input is 64 MiB and 50,000,000 decoded pixels. Dimensions are parsed before worker dispatch.
- Output is a 160x160, cover-fit, nonanimated WebP at quality 72. Exact codec bytes are presentation, but WebP MIME, dimensions/fit, static output, and acceptable visual result remain capabilities.
- The frozen optimizer applies JPEG and WebP EXIF orientation, ignores PNG `eXIf` and AVIF `irot`/`imir` in the observed paths, and reduces animated AVIF/GIF/WebP to the rendered or composited first frame. These format-specific results are the first-release policy; do not add a generic metadata normalizer.
- Two persistent workers execute conversions; 128 tasks may wait; each task times out after five seconds. A timed-out worker is replaced. If every worker fails, queued work fails.
- Concurrent requests for the same `Workspace Path + size + mtime milliseconds` share one conversion.
- Process-local LRU holds at most 1000 entries and 32 MiB. It writes nothing under Browse Root and is discarded on stop.
- Thumbnail ETag is `W/"thumb-<size>-<mtime ms>-160-72"`; responses are `image/webp`, no-cache, conditional 304, Last-Modified, exact length, HEAD-capable, nosniff, and no-referrer. Range and CORS are absent.
- Failures are `THUMBNAIL_ERROR` with 404 unsupported, 413 byte/pixel limit, 422 malformed/conversion, 503 stopped/full/worker failure, or 504 timeout.

The current converter is `wasm-image-optimization` inside Bun workers. [Research a CGO-free FS thumbnail compatibility path](../issues/21-research-cgo-free-fs-thumbnails.md) selects pinned `gav1d/avif v0.2.5`, `vpx/webp v0.2.1`, `x/image/draw v0.45.0`, standard JPEG/GIF/PNG, and an owner-local JPEG EXIF transform. It compiles into the one ycy file on all six targets. Two unadvertised persistent self-exec child modes retain the hard five-second kill/reap/replacement boundary without a codec runtime payload or helper artifact.

## SSE contract

Download and extraction SSE are separate endpoints with the same framing:

```text
data: {"version":1,"tasks":[...]}

```

- Subscribe emits one complete snapshot synchronously, then complete snapshots after eligible updates.
- Events have no name, ID, retry field, comment, or heartbeat.
- Headers are API security headers plus `Cache-Control:no-cache`, `Connection:keep-alive`, `Content-Type:text/event-stream; charset=utf-8`, and `X-Accel-Buffering:no`.
- Bun disables request timeout. Client cancellation unsubscribes. Session revocation closes an authenticated stream; server stop force-closes all streams.
- React first fetches the list, then opens EventSource. Native EventSource reconnect behavior is retained. On stream error it checks `/api/session` to distinguish logout.
- There is no connection count, bounded channel, slow-consumer policy, write deadline, heartbeat, or explicit memory backpressure. Go must retain the immediate flushed snapshot behavior; adding limits is post-parity work.

## Active React application contract

The active app moves from `src/commands/fs/web` into the single pnpm/Vite `web/` workspace. Its core workflow remains:

- Resolve `/api/session` before rendering data. Show login when enabled/unauthenticated and return there after any JSON, XHR upload, or SSE-derived authentication failure.
- Directory route is `/` or `/browse/<encoded Workspace Path>`. A malformed browser encoding shows an error and falls back to root.
- Preview is `?preview=<Workspace Path>` and uses `history.replaceState`, so opening/switching/closing preview does not add history. Directory/tree navigation pushes history and clears selection/query/editor state.
- Dirty editor state guards close, preview switch, directory/tree navigation, browser back/forward, and page unload.
- List and grid keep directories before files, local substring search, name/size/modified sorting, multi-selection, Enter/double-click directory opening, single-click file preview, modifier/range selection, context actions, and fixed-row virtualization.
- In-memory copy/cut clipboard survives directory navigation but not reload and never touches the system clipboard. Partial move removes only successfully moved sources.
- Management exposes new folder, rename, copy, move, permanent delete with exact confirmation, upload/drop, remote download, extraction, retry/cancel, and a unified activity center.
- Image lists fetch only `thumbnailUrl` with lazy/async/low priority. Original bytes are reserved for preview/open/download. Images have a one-image full-screen pan/zoom/rotate viewer; video/audio/PDF use browser-native controls.
- Text preview uses escaped text or `@pierre/diffs` code rendering. HTML/XML text is never executed inside the preview. Opening the original HTML in a new tab follows server safe-HTML policy.
- Monaco editing retains language workers for JSON, CSS family, HTML/XML, TypeScript/JavaScript, and generic editor, with other language IDs mapped to supported/plaintext modes.
- Multi-upload uses three browser workers. Large chunk uploads are sequential per file with server-offset recovery and three retries.
- Completion of a visible-directory download or adjacent extraction refreshes the listing.

Only these local-storage keys persist:

```text
ycy-fs-theme
ycy-fs-view
ycy-fs-sort
ycy-fs-sort-direction
ycy-fs-navigation-width
```

Navigation width defaults 400 px and clamps 180-560. Paths, preview, clipboard, file content, drafts, upload selections, and task history are not persisted.

### Vite boundary

- Bun HTMLBundle import, Bun file-loader import attributes, direct hashed paths under `node_modules/monaco-editor/min/vs/assets`, the TypeScript thumbnail worker, and the Pierre worker import cannot survive unchanged.
- The shell becomes `web/fs/index.html`; Vite must own Pierre and Monaco worker URLs and emit them into the common hashed asset tree.
- Production HTML remains no-store. Hashed assets remain one-year public immutable. HTML retains the current self-only CSP including `worker-src`, `connect-src`, `frame-src`, media/image rules, `nosniff`, and `no-referrer`.
- Only `/`, `/browse`, and `/browse/*` receive the FS shell. `/api/*`, `/files/*`, `/thumbnails/*`, `/assets/*`, and another application's routes must not fall back to it.
- Vite filenames are not contracts. Resolvable worker URLs, route ownership, headers, CSP, no remote runtime assets, and behavior from the built CGO-free binary are contracts.

## Security and resource observations

Confirmed legacy findings retained for first-release parity and post-parity hardening are:

1. Default inline HTML/XHTML is active same-origin content. In Management Mode it can issue mutations; in Login Mode it can act with the viewer's cookie. `--safe-html` being opt-in leaves a stored-XSS/capability-confusion path.
2. Root containment is check-then-use and pathname-based. Symlink/ancestor replacement can race reads and mutations; invalid UTF-8 names have no safe identity contract.
3. Remote download rejects only literal private addresses. DNS resolution, rebinding, custom resolution, and proxy dialing can reach forbidden networks.
4. Login/operations/download/extraction JSON can consume nearly the global 1 GiB body allowance. Directory entry fan-out, remote bytes/duration/disk, and concurrent authenticated session refresh work are insufficiently bounded.
5. SSE has no subscriber, queued-byte, slow-client, heartbeat, or write-deadline policy.
6. Login has no rate/concurrency policy despite expensive Argon2 work. Session refresh fsyncs on every request; Windows ACL guarantees are unspecified; the PID file lock is vulnerable to PID reuse/stale races.
7. The extracted 7-Zip runtime is not verified at runtime and its state directory lacks a protected publication contract.
8. Startup failures exit 0, the chunk-size CLI option is ignored, IPv6 URL presentation is malformed, and shutdown can wait five minutes.
9. Several native semantics are untested on Windows: reparse containment, case-only operations, hard-link publication, replace-rename, executable termination, archive special entries, and actual secret ACLs.

Copy the observable defaults for the first Go release because the active React client and existing callers are part of the parity target. Trust/HTML, Host/Origin, DNS/dial, limits, errors, and shutdown corrections are post-parity work. Bun-written FS persistence has no direct-read or migration gate; 7-Zip artifact assembly remains coordinated with `Inventory upgrade and release-artifact compatibility contracts`.

Already-unambiguous invariants that remain are:

- no absolute-path API and no target disclosure for unavailable links;
- Management Mode and same-origin protection for mutation;
- Login Mode protects every data/original/thumbnail route while keeping only the login shell/assets public;
- no overwrite publication, permanent delete is explicit, and batch partial results are visible;
- strict conditional text saves with no force overwrite;
- bounded thumbnail source/pixels/queue/cache and bounded extraction concurrency;
- verified pinned 7-Zip intent and no manual user repair;
- restrictive API/shell headers, sandboxed active document types, and no remote frontend dependency.

## Bun and TypeScript replacement matrix

| Current dependency/behavior | Go/Vite replacement boundary | Compatibility hazard |
| --- | --- | --- |
| Commander option collectors | selected Go CLI adapter | optional directory, repeated account, permissive ints, option property bug, errors/exits |
| `Bun.serve` routes and HTMLBundle | `net/http` adapter plus embedded Vite assets | path cleaning, HEAD/OPTIONS, body ceilings, JSON/date omission, graceful stop |
| `Bun.file`, `Bun.write`, Node path/fs | command-owned opened-handle workspace | MIME, mtime precision, symlink/reparse races, invalid filename bytes, hard links |
| Bun MIME database | deterministic FS MIME/preview adapter | parameters, OS registry variation, structured text and inline policy |
| `Bun.password` Argon2id | `x/crypto/argon2` or approved wrapper | parameters/timing/concurrency; no persisted PHC string to migrate |
| shared `FileSessionManager` | shared deep persistent-session module | direct v1 JSON/key read, locks, fsync, expiry races, Windows ACLs |
| Web Streams and `Bun.write` chunking | bounded `io.Reader`/opened file operations | exact lengths, cancellation, partial retry, disk reservation, publication |
| Bun fetch/WHATWG URL | hardened `net/http` download client | IDNA/URL grammar, DNS/dial/proxy SSRF, decompression, redirects, timeouts |
| `Bun.spawn`/`Bun.which` | `os/exec` plus runtime locator | cancellation/process tree, locale, stderr bounds, Windows DLL colocation |
| Bun `embeddedFiles` | target-specific `go:embed` runtime payload | per-target contents, digests, permissions, atomic verified state |
| Bun Worker plus WASM image optimizer | pinned pure-Go codec graph plus same-binary self-exec workers | AVIF/WebP matrix, format-specific orientation/animation, WebP encode, hard timeout/replacement, output variance |
| Bun HTML/file-loader worker imports | Vite worker/asset graph | Pierre/Monaco worker URLs, CSP, shared hashed assets, entry isolation |
| React/Radix/TanStack/Pierre/Monaco/Photo View | retained under pnpm/Vite | API shape, worker bundling, route history, mobile/desktop behavior |
| JavaScript dates, numbers, undefined | explicit Go wire types | millisecond UTC, safe integer cap, omitted versus zero/null fields |

## Required Go and frontend tests

1. **CLI/lifecycle:** exact options/defaults/env precedence; working CLI chunk size; account repetition/grammar; port 0/65535/bad values; missing/non-directory root; session/bind failure; nonzero failure status; actual URL output including IPv6; SIGINT/SIGTERM with idle and active upload/download/extraction; bounded shutdown and no process/temp leak.
2. **Workspace Path:** empty/duplicate separators, dot/dotdot, drive/absolute/backslash/NUL, malformed and encoded percent/slash, plus/query behavior, Unicode, non-UTF-8 Unix policy, internal/broken/escaping/cyclic links, ancestor replacement, root replacement, and Windows reparse points using opened-handle containment.
3. **Listing/MIME:** directory/file/unavailable and symlink fields; URL encoding; omitted JSON fields; UTC milliseconds; temporary hiding; deterministic directory-first sorting; Bun-derived MIME/syntax/preview vectors; archive and thumbnail suffix cases.
4. **Original HTTP:** GET/HEAD/OPTIONS, inline/attachment filenames including Unicode/control characters, default and selected HTML policy, SVG/XML CSP, authenticated/no-auth CORS, directory redirect, weak ETag/date precedence, every single/suffix/open/invalid/multiple/empty Range and If-Range case, mutation during response, and stable opened bytes.
5. **Authentication/session:** all account boundaries including UTF-16 length; unknown/wrong timing path; cookie attributes; login/logout/refresh; 8/128 LRU; fresh-Go idle restart continuity; password-change revoke; observer close; unavailable storage 503; Go-owned corruption/expiry/temp pruning; concurrent refresh; Unix mode and native Windows ACL/lock tests. Do not add a Bun-written v1 fixture or carryover gate.
6. **Origin/trust:** reproduce the legacy allowed, absent, hostile, malformed, and DNS-rebinding Host/Origin behavior under loopback, wildcard, IPv6, and explicit LAN binds; active HTML follows the legacy default; authentication does not imply TLS.
7. **Text:** exact 10 MiB boundaries; UTF-8 BOM/no-BOM, UTF-16LE/BE BOM, invalid/odd bytes, NUL; SHA revision; missing/symlink/special/too-large; all media-type and If-Match errors; LF/CRLF/CR ties/trailing runs; supplementary/lone-surrogate drafts; mode and selected metadata; simultaneous/external conflict with no partial overwrite.
8. **Operations:** strict schema/extra fields/1000 bounds; invalid names; root immutability; same/case-only names; collision exhaustion; file/dir/symlink copy; self/descendant checks; cross-device move; permanent link delete; ordered partial success; concurrent target creation and no overwrite on every native OS.
9. **Direct/chunk upload:** multipart field/type/body limits, zero/1 GiB boundaries, collision and publication failure; exact 20 MiB switch; 4/8/16 chunks; safe integer/range/content length/type; owner isolation; one-write busy; partial retry; timeout; 3/100 limits; 30/5-minute expiry; idempotent concurrent completion; capacity/reservation races; stale cleanup; stop/page cancellation.
10. **Remote download:** WHATWG-derived accepted/rejected URL vectors; credentials/IDNA/IPv4/IPv6; every redirect; response filename forms; encoding and length mismatch; connection/body/total timeouts; DNS resolution/rebinding and proxy attempts; selected byte/disk quotas; stream cleanup/collision; 2/100/history ordering; cancel/retry/clear/restart and progress throttling.
11. **Archive/runtime:** all suffixes/base collisions; pinned six-target manifest and per-file digests/license; corrupted/missing/symlinked runtime state; source PATH fallback; inspection parsing and normalized failures; encrypted/multipart/bombs/bytes/inodes; all real formats/layered TAR; unsafe names/backslashes/links/special files; cancel/process-tree cleanup; native Windows and Unix publication.
12. **Thumbnail:** all five input formats/case, SVG rejection, malformed data, 64 MiB and 50M-pixel boundaries, JPEG/WebP orientation application plus PNG/AVIF nonapplication, animated first-frame policy, 160 cover-fit static WebP result, 2/128/5-second limits, persistent self-exec reuse, kill/`Wait`/different-PID replacement, worker recovery/all-worker failure, in-flight coalescing, 1000/32 MiB LRU, mutation/ETag/HEAD/304, shutdown, and adversarial decoder memory.
13. **Tasks/SSE:** exact state fields/UTC milliseconds/order; immediate and progress snapshots; framing/headers/flush; stale response merging in React; disconnect and session revoke; EventSource reconnect; cancel/retry/clear; selected connection/backpressure/heartbeat/write-timeout limits; stop with no subscribers or goroutines left.
14. **Frontend/Vite:** production pnpm build resolves Pierre and every Monaco worker; fixed FS shell and common hashed assets; no cross-entry fallback; built-binary cache/MIME/CSP probes; login/session expiry; directory/history/preview/dirty editor; selection/sort/search/virtualization; all preview types; partial operations; upload/download/extraction activity; long/Unicode names; local-storage corruption; 899 px mobile boundary; no console/network errors.
15. **Native/resource gates:** run filesystem, session, HTTP, 7-Zip, process, and browser smoke tests on all six artifact pairs. Exercise representative large-directory and concurrent workflows against the observed legacy behavior. Tests are written from `legacy/bun/` command-by-command; active tests never start Bun and do not create a separately maintained golden service corpus.

## Migration boundaries and readiness

1. FS is a **critical/high-risk** migration: it combines arbitrary local bytes, permanent mutation, credentials, persistent sessions, SSRF-capable networking, native archive execution, image decoding, streaming HTTP, and an active React application.
2. Preserve one deep root-confined workspace API. HTTP and queues own protocol/lifecycle; workspace owns safe handles, capacity, staging, and publication; React owns interaction and never receives absolute paths. Exact packages remain for `Choose the Go module seams and project layout`.
3. Implement the fresh-Go session owner/lock before Login Mode and reuse only matching invariants for fresh-Go Tunnel state. Do not add a Bun-state reader, transformation, fixture, or reconfiguration workflow.
4. A practical implementation order inside FS is: read-only workspace/list/file/text, behavior-compatible HTTP/auth/session handling over fresh-Go state, conditional editing and operations, direct/chunk upload, remote download, 7-Zip extraction, thumbnail engine, then the retained React application and production embed gate.
5. Process-local upload/download/extraction/thumbnail state intentionally does not migrate across the Bun-to-Go cutover. Hidden incomplete staging is cleaned under the documented age/name rules; it is never treated as user configuration.
6. [Research a CGO-free FS thumbnail compatibility path](../issues/21-research-cgo-free-fs-thumbnails.md) has resolved the known runtime-capability question with a fixed pure-Go engine and self-exec worker contract. HTML/trust, containment, remote dial, resource, SSE, login, exit, and shutdown hardening remains deferred.
7. 7-Zip remains an external bundled executable/data payload used by a CGO-free Go binary. Its exact cross-target assembly joins the release-artifact inventory; active Go code does not depend on Bun to prepare or execute it.
8. Completion requires raw HTTP/session compatibility, native filesystem and process tests, frontend production-browser verification, and the final CGO-free standalone artifact. A workspace unit suite or Vite dev server alone cannot declare `fs` migrated.

This inventory resolves local fact finding. It does not approve a thumbnail library, design the Go package graph, implement the service, run legacy Bun from the future active test suite, or authorize post-parity hardening.
