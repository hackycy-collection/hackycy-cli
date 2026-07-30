# Static File Server Command

This directory contains `ycy serve`: a root-confined filesystem workspace, a same-origin HTTP adapter, and an embedded React file browser.

## Purpose

```text
ycy serve [options] [directory]

Options:
  -p, --port <number>       Port to serve on (default: 1204)
  -a, --address <string>    Address to bind to (default: 0.0.0.0)
  -u, --upload              Enable file uploads
```

The directory defaults to the current working directory. The default binding is available to the local network; use `--address 127.0.0.1` for local-only access. Upload mode has no authentication and must be used only with a trusted directory and network.

## Architecture

```text
CLI registration (index.ts)
  -> process composition (run.ts)
     -> ServeWorkspace (workspace.ts)
     -> ServeHttpServer (server.ts)
        -> JSON and file HTTP adapter
        -> embedded React application (web/)
```

- `workspace.ts` owns root confinement, symlink policy, metadata, text decoding, and atomic upload naming. Callers pass only POSIX relative paths.
- `server.ts` maps workspace results to HTTP, validates methods and mutation origins, implements cache and Range semantics, and serves the embedded HTML bundle.
- `web/` owns History navigation, sorting, virtualization, preview state, theme, and the three-worker upload queue. It never constructs absolute filesystem paths.
- Shared Radix/Tailwind primitives live under `src/shared/web` and are consumed by both `serve` and `diff`.

## HTTP Interface

| Route | Behavior |
| --- | --- |
| `GET /`, `GET /browse/*` | Embedded React application and browser history fallback. |
| `GET /api/directory?path=` | Current directory metadata and entries. |
| `GET /api/text?path=` | UTF-8 or BOM-marked UTF-16 text up to 2 MiB. |
| `POST /api/upload?path=` | One multipart file per request when `--upload` is enabled. |
| `GET\|HEAD /files/*` | Original file bytes; `?download=1` forces attachment. |

The former direct file URL shape is intentionally not retained. A served path such as `docs/readme.txt` is available at `/files/docs/readme.txt`; `/browse/docs` is the browser route.

Errors use `{ version: 1, error: { code, message } }`. Directory and text responses are not cacheable. Original files support ETag, Last-Modified, HEAD, and one byte range. Only `/files/*` enables wildcard CORS.

## Filesystem And Upload Invariants

- The root is resolved once at startup. Absolute paths, backslashes, dot segments, malformed URL encoding, and paths whose real target escapes the root are rejected.
- Internal symlinks may be followed. Escaping, unreadable, or unsupported entries are listed as unavailable without exposing their targets.
- Text preview checks the 2 MiB size limit before reading and treats invalid supported encodings as binary.
- Each upload is capped at 1 GiB, written to a temporary file in the destination directory, then published atomically with a hard link.
- Existing names are never overwritten. Collisions receive `name (1).ext` through `name (9999).ext`.
- Upload requests require an exact same Origin and a hostname permitted by the active binding.

## Web Application Invariants

- Directory URLs use `/browse/<encoded-path>` and browser History; preview selection uses the `preview` query parameter.
- List and grid views virtualize rows with `@tanstack/react-virtual`. Directories remain ahead of files for every sort.
- Images, audio, video, and PDF use browser-native presentation. Text is rendered as escaped content; HTML and XML are never executed in the preview.
- Theme, view mode, sort key, and direction may be stored locally. Paths and file contents are not persisted.
- Multi-file upload runs at most three requests concurrently and refreshes the current listing when the queue settles.

## Verification

```bash
bun test src/commands/serve
bun test src/commands/diff
bun run typecheck
bun run lint --no-cache
bun scripts/build.ts --outfile .tmp/ycy
```

Changes to shared Web primitives must run both command suites. Changes to the HTML entrypoint must work from source and from the compiled standalone executable.
