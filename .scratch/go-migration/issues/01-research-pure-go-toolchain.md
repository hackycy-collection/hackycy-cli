# Research the pure-Go toolchain and dependency baseline

Type: research
Status: resolved

## Question

Using current official Go documentation and primary sources for candidate libraries, which maintained Go version and dependency candidates can satisfy `CGO_ENABLED=0` across macOS, Linux, and Windows on x64 and arm64? Cover the CLI parser, SQLite with existing-file compatibility, Argon2id and existing password hashes, AES-GCM/PBKDF2, WebSocket custom headers, TOML, glob/gitignore behavior, ZIP/archive handling, MIME detection, terminal prompts/tables, MCP server support, and cross-platform process/signal handling. Record maintenance status, licensing, portability limits, and the decisions that later prototypes must validate.

## Comments

- Research context: branch `research/pure-go-toolchain`; report `.scratch/go-migration/research/01-pure-go-toolchain.md`.

## Answer

Primary-source research is complete at commit `69da548c3286433bd6a38443d56a2df608d220b1` on branch `research/pure-go-toolchain`; the full report is `.scratch/go-migration/research/01-pure-go-toolchain.md` on that branch.

Adopt `go 1.26.0` with `toolchain go1.26.7` and enforce `CGO_ENABLED=0`. A probe importing the complete candidate dependency graph compiled for all six required `darwin`, `linux`, and `windows` amd64/arm64 targets. The starting baseline selects Cobra behind a Commander-compatibility adapter, `ncruces/go-sqlite3` as the first SQLite candidate, standard-library AES-GCM/PBKDF2, `x/crypto/argon2`, coder/websocket, go-toml/v2, doublestar, the official MCP Go SDK, Huh/Lip Gloss, and `x/sys`; go-git gitignore, flock, and modernc SQLite remain conditional rather than default dependencies.

Cross-compilation is not runtime approval. Existing SQLite/WAL behavior, encrypted configuration, Argon2 PHC strings, Gitignore semantics, WebSocket/MCP protocols, FRP lifecycle, prompts, Windows process trees, and all six real artifacts retain explicit native or compatibility gates. The report also confirms the root Go module with `cmd/ycy`, private deep modules under `internal/`, one active `web/`, frozen `legacy/bun/`, and no speculative `pkg` or generic pass-through layers.
