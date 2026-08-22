# Inventory fs compatibility contracts

Type: task
Status: resolved

## Question

Inventory the complete observable contract of `fs`: CLI parsing, HTTP and SSE routes, authentication and session persistence, Range and MIME behavior, file-management safety, uploads and retryable chunks, editing conflicts, downloads, preview and thumbnail processing, worker/concurrency behavior, archive inspection/extraction, embedded 7-Zip acquisition and verification, browser/static caching, shutdown, limits, platform-specific paths and permissions, and existing tests. Identify existing formats and Bun-specific semantics that a pure-Go implementation must reproduce or migrate automatically.

## Comments

- 2026-08-22, scope correction: FS defects no longer require corrected contracts before the port. Only the evidenced possibility that Sharp's thumbnail capability lacks a maintained CGO-free Go equivalent remains a focused compatibility research item.
- 2026-08-22, transition-scope amendment: Bun-written FS session carryover is no longer a first-release compatibility or acceptance requirement. The direct-read finding remains factual evidence, but Go uses its normal FS startup and diagnostic path and internal operators handle any old session residue.

## Answer

The [FS command compatibility inventory](../inventories/fs-contracts.md) records the CLI, lifecycle, root-confined filesystem, HTTP/Range/MIME, persistent authentication, operations, upload/download/extraction queues, SSE, thumbnail, embedded 7-Zip, retained React, Vite, platform, and test contracts needed by the Go migration.

The verified legacy baseline is **133 passing / 0 failing tests** across 18 files. Existing FS session state needs no one-way migration: the 32-byte `.session-key`, SHA-256 token filenames, and version 1 JSON records are directly readable from Go, including the base64url HMAC-SHA256 credential revision.

That direct-read capability is not required by the final roadmap. Fresh-Go FS authentication/session behavior remains in scope, but no Bun-session fixture, migration, cleanup, or carryover release gate is required.

The ignored `--upload-chunk-size` option, startup failures that exit 0, same-origin active HTML, DNS-unaware remote downloads, pathname race windows, effectively unbounded JSON/download/SSE work, insufficient login/session/runtime-state protections, and native Windows gaps remain observed parity behavior and later hardening candidates. [Research a CGO-free FS thumbnail compatibility path](21-research-cgo-free-fs-thumbnails.md) subsequently resolved the only open runtime-capability question with a six-target pure-Go path and no compatibility exception.
