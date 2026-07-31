# Static File Server Command

This directory contains `ycy serve`: a root-confined filesystem workspace, a same-origin HTTP adapter, and an embedded React file browser.

## Purpose

```text
ycy serve [options] [directory]

Options:
  -p, --port <number>       Port to serve on (default: 1204)
  -a, --address <string>    Address to bind to (default: 0.0.0.0)
  -m, --manage              Enable uploads and filesystem management
```

The directory defaults to the current working directory. The default binding is available to the local network; use `--address 127.0.0.1` for local-only access. Management mode has no authentication and allows upload, copy, move, rename, and permanent deletion. Use it only with a trusted directory and network.

## Architecture

```text
CLI registration (index.ts)
  -> process composition (run.ts)
     -> ServeWorkspace (workspace.ts)
     -> ServeHttpServer (server.ts)
        -> JSON and file HTTP adapter
        -> ThumbnailService and two persistent conversion workers
        -> embedded React application (web/)
```

- `workspace.ts` owns root confinement, symlink policy, metadata, text decoding, and atomic upload naming. Callers pass only POSIX relative paths.
- `server.ts` maps workspace results to HTTP, validates methods and mutation origins, implements cache and Range semantics, and serves the embedded HTML bundle.
- `thumbnail-service.ts` owns input limits, request coalescing, the bounded worker queue, and the session-only LRU. `thumbnail-worker.ts` performs WASM decoding and WebP conversion off the HTTP thread.
- `web/` owns directory History navigation, sorting, virtualization, preview state, theme, one syntax-highlighting worker, and the three-worker upload queue. It never constructs absolute filesystem paths.
- Shared Radix/Tailwind primitives live under `src/shared/web` and are consumed by both `serve` and `diff`.

## HTTP Interface

| Route | Behavior |
| --- | --- |
| `GET /`, `GET /browse/*` | Embedded React application and browser history fallback. |
| `GET /api/directory?path=` | Current directory metadata and entries. |
| `GET /api/text?path=` | UTF-8 or BOM-marked UTF-16 text up to 2 MiB. |
| `POST /api/upload?path=` | One multipart file per request when `--manage` is enabled. |
| `POST /api/operations` | Validated create-directory, rename, copy, move, and permanent-delete commands in management mode. |
| `GET\|HEAD /files/*` | Original file bytes; `?download=1` forces attachment. |
| `GET\|HEAD /thumbnails/*` | 160×160 WebP thumbnail for JPEG, PNG, WebP, AVIF, or GIF input. |

The former direct file URL shape is intentionally not retained. A served path such as `docs/readme.txt` is available at `/files/docs/readme.txt`; `/browse/docs` is the browser route.

Errors use `{ version: 1, error: { code, message } }`. Directory and text responses are not cacheable. Original files support ETag, Last-Modified, HEAD, and one byte range. Thumbnail responses support ETag, Last-Modified, and conditional 304 responses. Only `/files/*` enables wildcard CORS.

## Filesystem And Management Invariants

- The root is resolved once at startup. Absolute paths, backslashes, dot segments, malformed URL encoding, and paths whose real target escapes the root are rejected.
- Internal symlinks may be followed. Escaping, unreadable, or unsupported entries are listed as unavailable without exposing their targets.
- Text preview checks the 2 MiB size limit before reading and treats invalid supported encodings as binary.
- Thumbnail input is capped at 64 MiB and 50 million pixels. SVG is never sent to the raster converter. Failed or oversized thumbnails remain file icons in the main browser and never trigger an original-image fallback.
- Thumbnail conversion uses two persistent workers, at most 128 queued tasks, a five-second task timeout, and replacement of a timed-out worker. Concurrent requests for the same file revision share one conversion.
- Thumbnail output stays in a process-local LRU keyed by path, size, and modification time. The cache holds at most 1000 entries or 32 MiB, writes nothing to the served directory, and is discarded when the server stops.
- Each upload is capped at 1 GiB, written to a temporary file in the destination directory, then published atomically with a hard link.
- Existing names are never overwritten. Collisions receive `name (1).ext` through `name (9999).ext`.
- The served root cannot be renamed, moved, copied, or deleted. Directories cannot be copied or moved into themselves.
- Rename and move reject collisions. Copy and upload choose collision-safe numbered names. Recursive copy does not dereference symbolic links.
- Delete is permanent and operates on the final directory entry itself, so deleting a symbolic link never deletes its target.
- All management requests require an exact same Origin and a hostname permitted by the active binding. Batch operations accept at most 1000 paths and report partial results per item.

## Web Application Invariants

- Directory URLs use `/browse/<encoded-path>` and browser History. Preview selection uses the `preview` query parameter but always replaces the current URL, so opening, switching, and closing previews never create History entries.
- List and grid views virtualize fixed-size rows with `@tanstack/react-virtual`, rendering four extra list rows or one extra grid row around the viewport. Grid columns are measured before entries mount and only change at layout breakpoints. Directories remain ahead of files for every sort. Search filters only the loaded directory.
- Main-list image elements load only `thumbnailUrl`, with lazy asynchronous low-priority decoding. Original `fileUrl` bytes are reserved for preview, opening, and download.
- Selection follows desktop file-manager conventions: clicking a file selects and previews it, Ctrl/Cmd toggles, Shift selects a range, and modifiers never open a preview. Directories open on double-click or Enter.
- Cut, copy, and paste use an in-memory browser clipboard. It survives directory navigation but not a page reload and never reads or writes the system clipboard.
- Source code, configuration files, and dotenv variants use the `@pierre/diffs` Shiki-based read-only file renderer. Plain text remains escaped content; HTML and XML are never executed in the preview.
- Images retain an inline preview and use `react-photo-view` for a current-image-only full-viewport viewer with pan, zoom, rotate, and reset controls. Audio, video, and PDF keep browser-native presentation.
- Theme, view mode, sort key, and direction may be stored locally. Paths and file contents are not persisted.
- Multi-file upload runs at most three requests concurrently and refreshes the current listing when the queue settles.

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
