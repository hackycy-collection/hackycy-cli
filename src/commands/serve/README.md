# Static File Server Command

This directory contains `ycy serve`: a root-confined filesystem workspace, a same-origin HTTP adapter, and an embedded React file browser.

## Purpose

```text
ycy serve [options] [directory]

Options:
  -p, --port <number>       Port to serve on (default: 1204)
  -a, --address <string>    Address to bind to (default: 0.0.0.0)
  -m, --manage              Enable uploads, remote downloads, and filesystem management
      --account <user:pass> Require login with an account (repeatable)
```

The directory defaults to the current working directory. The default binding is available to the local network; use `--address 127.0.0.1` for local-only access. Management mode allows upload, copy, move, rename, and permanent deletion. Without `--account`, use it only with a trusted directory and network.

Passing one or more accounts enables login mode:

```bash
ycy serve --account 'alice:password123' --account 'bob:another-password' ./shared
```

The first colon separates the username from the password, so the password may contain additional colons. Usernames contain 1-64 ASCII letters, numbers, dots, underscores, or hyphens and are matched without case sensitivity. Passwords contain 8-256 characters. Duplicate usernames stop startup. Account specifications are process arguments and may be visible in shell history and process inspection tools.

## Architecture

```text
CLI registration (index.ts)
  -> process composition (run.ts)
     -> ServeWorkspace (workspace.ts)
     -> ServeAuthentication (authentication.ts, when accounts are configured)
     -> ServeHttpServer (server.ts)
        -> JSON and file HTTP adapter
        -> RemoteDownloadManager with a bounded process-local queue
        -> ThumbnailService and two persistent conversion workers
        -> embedded React application (web/)
```

- `workspace.ts` owns root confinement, symlink policy, metadata, text decoding, and atomic upload naming. Callers pass only POSIX relative paths.
- `authentication.ts` owns account parsing, password hashing, bounded process-local sessions, expiration, and revocation notifications.
- `server.ts` maps workspace results to HTTP, validates methods and mutation origins, implements cache and Range semantics, and serves the embedded HTML bundle.
- `download-service.ts` validates remote HTTP(S) targets, blocks literal private and reserved IP addresses, follows validated redirects, and owns the bounded process-local download queue.
- `thumbnail-service.ts` owns input limits, request coalescing, the bounded worker queue, and the session-only LRU. `thumbnail-worker.ts` performs WASM decoding and WebP conversion off the HTTP thread.
- `web/` owns directory History navigation, sorting, virtualization, preview state, theme, one syntax-highlighting worker, and the three-worker upload queue. It never constructs absolute filesystem paths.
- Shared Radix/Tailwind primitives live under `src/shared/web` and are consumed by both `serve` and `diff`.

## HTTP Interface

| Route | Behavior |
| --- | --- |
| `GET /`, `GET /browse/*` | Embedded React application and browser history fallback. |
| `GET\|POST\|DELETE /api/session` | Inspect the current login state, authenticate, or end the current session. |
| `GET /api/directory?path=` | Current directory metadata and entries. |
| `GET /api/text?path=` | UTF-8 or BOM-marked UTF-16 text up to 2 MiB. |
| `POST /api/upload?path=` | One multipart file per request when `--manage` is enabled. |
| `POST /api/operations` | Validated create-directory, rename, copy, move, and permanent-delete commands in management mode. |
| `GET\|POST\|DELETE /api/downloads` | List, create, or clear terminal remote-download tasks in management mode. |
| `GET /api/downloads/events` | Server-sent task snapshots for remote-download progress. |
| `POST /api/downloads/:id/cancel` | Cancel one queued or active remote download. |
| `POST /api/downloads/:id/retry` | Retry one failed or cancelled remote download as a new task. |
| `GET\|HEAD /files/*` | Original file bytes; `?download=1` forces attachment. |
| `GET\|HEAD /thumbnails/*` | 160×160 WebP thumbnail for JPEG, PNG, WebP, AVIF, or GIF input. |

The former direct file URL shape is intentionally not retained. A served path such as `docs/readme.txt` is available at `/files/docs/readme.txt`; `/browse/docs` is the browser route.

Errors use `{ version: 1, error: { code, message } }`. Directory and text responses are not cacheable. Original files support ETag, Last-Modified, HEAD, and one byte range. Thumbnail responses support ETag, Last-Modified, and conditional 304 responses. Only `/files/*` enables wildcard CORS, and only when login mode is disabled.

## Authentication Invariants

- Login mode is enabled by the presence of at least one `--account`. Without accounts, the HTTP behavior remains unauthenticated and `/api/session` reports that authentication is disabled.
- The application shell and its compiled assets remain public so the browser can render the login form. Every other `/api/*`, `/files/*`, and `/thumbnails/*` request requires a valid session.
- Passwords are hashed with Argon2id during startup. Failed logins use the same verification path whether or not the username exists.
- Sessions use random process-local tokens in an `HttpOnly`, `SameSite=Strict` cookie. They expire after 12 hours, do not survive a restart, and are bounded to eight per account and 128 for the server.
- Logging out, session expiry, or bounded-session eviction revokes the token and closes its active remote-download event stream.
- All configured accounts have the same read and management permissions. There is no registration, role, password-change, or persistent account interface.
- Authentication does not add TLS. Passwords and session cookies travel over the HTTP connection exposed by `serve`.

## Filesystem And Management Invariants

- The root is resolved once at startup. Absolute paths, backslashes, dot segments, malformed URL encoding, and paths whose real target escapes the root are rejected.
- Internal symlinks may be followed. Escaping, unreadable, or unsupported entries are listed as unavailable without exposing their targets.
- Text preview checks the 2 MiB size limit before reading and treats invalid supported encodings as binary.
- Thumbnail input is capped at 64 MiB and 50 million pixels. SVG is never sent to the raster converter. Failed or oversized thumbnails remain file icons in the main browser and never trigger an original-image fallback.
- Thumbnail conversion uses two persistent workers, at most 128 queued tasks, a five-second task timeout, and replacement of a timed-out worker. Concurrent requests for the same file revision share one conversion.
- Thumbnail output stays in a process-local LRU keyed by path, size, and modification time. The cache holds at most 1000 entries or 32 MiB, writes nothing to the served directory, and is discarded when the server stops.
- Each upload is capped at 1 GiB, written to a temporary file in the destination directory, then published atomically with a hard link.
- Remote downloads have no upload-size cap. Response bodies are streamed through bounded chunks into hidden destination-local temporary files, then published atomically with the same collision-safe naming rules as uploads.
- At most two remote downloads run concurrently, at most 100 wait in the queue, and terminal task records are pruned to a bounded history. Cancelling a task or stopping the server aborts the request and removes its temporary file.
- Remote download URLs permit only HTTP(S) and cannot contain credentials. Literal loopback, private, link-local, multicast, documentation, and other reserved IP addresses are rejected without resolving domain names. Every redirect is revalidated and the chain is capped at five hops.
- Remote-download state is process-local. Browser reloads recover active tasks through the list and event interfaces; process restarts do not resume tasks or retain their history.
- Existing names are never overwritten. Collisions receive `name (1).ext` through `name (9999).ext`.
- The served root cannot be renamed, moved, copied, or deleted. Directories cannot be copied or moved into themselves.
- Rename and move reject collisions. Copy and upload choose collision-safe numbered names. Recursive copy does not dereference symbolic links.
- Delete is permanent and operates on the final directory entry itself, so deleting a symbolic link never deletes its target.
- All management requests require an exact same Origin and a hostname permitted by the active binding. Batch operations accept at most 1000 paths and report partial results per item.

## Web Application Invariants

- The application resolves `/api/session` before mounting the file browser. Authentication failures from JSON, upload, or event-stream requests return it to the login form.
- Directory URLs use `/browse/<encoded-path>` and browser History. Preview selection uses the `preview` query parameter but always replaces the current URL, so opening, switching, and closing previews never create History entries.
- List and grid views virtualize fixed-size rows with `@tanstack/react-virtual`, rendering four extra list rows or one extra grid row around the viewport. Grid columns are measured before entries mount and only change at layout breakpoints. Directories remain ahead of files for every sort. Search filters only the loaded directory.
- Main-list image elements load only `thumbnailUrl`, with lazy asynchronous low-priority decoding. Original `fileUrl` bytes are reserved for preview, opening, and download.
- Selection follows desktop file-manager conventions: clicking a file selects and previews it, Ctrl/Cmd toggles, Shift selects a range, and modifiers never open a preview. Directories open on double-click or Enter.
- Cut, copy, and paste use an in-memory browser clipboard. It survives directory navigation but not a page reload and never reads or writes the system clipboard.
- Source code, configuration files, and dotenv variants use the `@pierre/diffs` Shiki-based read-only file renderer. Plain text remains escaped content; HTML and XML are never executed in the preview.
- Images retain an inline preview and use `react-photo-view` for a current-image-only full-viewport viewer with pan, zoom, rotate, and reset controls. Audio, video, and PDF keep browser-native presentation.
- Theme, view mode, sort key, and direction may be stored locally. Paths and file contents are not persisted.
- Multi-file upload runs at most three requests concurrently and refreshes the current listing when the queue settles.
- Remote-download tasks share the activity center with uploads and file operations. They show transferred bytes, known percentage, server-measured speed, cancellation, and retry controls; unknown response lengths use an indeterminate progress bar.

## Verification

```bash
bun test src/commands/serve
bun run test:serve:performance
bun test src/commands/diff
bun run typecheck
bun run lint --no-cache
bun scripts/build.ts --outfile .tmp/ycy
```

Changes to shared Web primitives must run both command suites. Changes to the HTML entrypoint must work from source and from the compiled standalone executable.
