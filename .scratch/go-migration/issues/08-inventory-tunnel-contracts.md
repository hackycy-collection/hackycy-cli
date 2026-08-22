# Inventory tunnel compatibility contracts

Type: task
Status: resolved

## Question

Inventory `tunnel server` and `tunnel connect` separately, including option/environment/config precedence, interactive selection, remembered connection formats and identity, HTTP/SSE/WebSocket protocols, authentication and password hashes, SQLite schema and transaction behavior, agent reconciliation, FRP manifest/download/checksum/configuration, frps/frpc supervision and recovery, control-plane UI contracts, logging, signals, platform paths, and existing tests. Identify protocol or persistence changes that would disconnect existing clients or invalidate server data.

## Comments

- 2026-08-22, scope correction: first-release Tunnel work preserves protocol v3 and legacy behavior. Broad trust, bounds, schema, permission, and process hardening is deferred; only failed Bun/Go interoperability or target-platform probes may open narrow exceptions.
- 2026-08-22, transition-scope amendment: first-release Tunnel acceptance is Go client to Go server on protocol v3 only. Guaranteed Bun-written compatibility is limited to remembered connections inside `config.json`; mixed Bun/Go peers and carryover of sessions, SQLite, client cache, or generated runtime state are removed from the release gate.

## Answer

The complete compatibility inventory is [Tunnel command compatibility inventory](../inventories/tunnel-contracts.md). It separately specifies `tunnel server` and `tunnel connect`, including CLI/environment/remembered-value precedence, interactive ambiguity handling, encrypted multi-connection identity, platform paths and locks, SQLite v1 transactions, Argon2id accounts and persistent sessions, browser JSON/SSE/static routes, agent WebSocket protocol v3, authentication-first reconciliation and rollback, pinned FRP 0.70.1 artifacts/configuration/import, frps/frpc supervision, retained React workflows, current verification evidence, and the required Go/native test matrix.

Existing remembered connections inside `config.json` remain directly readable and retain their machine-identity/encryption tests. The other recorded file-session, client-cache, generated-TOML, Argon2, and SQLite shapes remain implementation evidence for normal Go behavior but receive no Bun-written carryover gate. The first Go release preserves protocol v3, wire platform names, and pinned FRP fields for Go client to Go server only; it does not run or support a mixed Bun/Go peer matrix.

Startup handling of unknown SQLite schemas, incomplete secret-state permissions, under-bounded network/resource work, missing WebSocket/SSE policies, reconnect timing, unlocked FRP installation, and unverified Windows process-tree behavior remain documented parity facts and later hardening candidates. First-release blockers are limited to Go-to-Go protocol-v3, `config.json`, FRP, or native-platform mismatches demonstrated while porting Tunnel.
