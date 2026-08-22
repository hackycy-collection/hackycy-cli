# Choose safe and rolling-compatible contracts for Tunnel

Type: grilling
Status: resolved
Blocked by: 08

## Question

Before either Go Tunnel leaf is approved, which corrected contracts replace the legacy control plane's unsafe, unbounded, or rollout-hostile behavior while preserving ordinary deployments and protocol-v3 peers?

Decide the mixed-version support window and future protocol negotiation for legacy Bun clients, Go clients, legacy Bun servers, and Go servers; Node-to-Go platform/architecture wire mapping; HTTP versus HTTPS control origins; Host, Origin, DNS-rebinding, trusted-proxy, forwarded-host, advertised-FRP-host, cookie `Secure`, and Origin-less automation policy; login rate/concurrency behavior; JSON/import/custom-page body limits; account/client/tunnel/route and desired-snapshot limits; SSE/WebSocket subscriber, heartbeat, hello-deadline, payload, backpressure, and slow-client behavior; cancellable reconnect and bounded shutdown semantics; and exact errors/close codes for every correction.

Also select the fail-closed behavior for missing, incomplete, corrupt, older, newer, or unknown SQLite schemas; the protection properties required for SQLite/WAL, session, client-state, generated FRP configuration, remembered credentials, and backups on Unix and Windows; FRP download/install locking, byte/time/decompression/publication bounds; and native Windows process-tree ownership, graceful stop, forced termination, PID/lock checks, atomic replacement, and ACL acceptance gates.

Coordinate concrete adapters, backup/rollback, one-way migration, and lock implementation with `Choose persistent-data compatibility mechanisms`; coordinate shell routing/cache/CSP mechanics with `Prove the Vite MPA to Go embed path`. Do not duplicate those mechanism decisions here. Record exact CLI/HTTP/WebSocket behavior, limits, security assumptions, observability, native-platform requirements, rolling-upgrade order, and adversarial/compatibility tests without requiring manual credential or state recreation.

## Comments

- 2026-08-22, roadmap scope amendment: mixed Bun/Go peers and non-config Bun state carryover are no longer first-release requirements; Go-only protocol v3 remains.

## Answer

Closed as out of scope in its broad safety-redesign form. The first Go release retains protocol v3, FRP behavior, and wire platform mappings for Go client to Go server. Remembered Tunnel connections inside `config.json` remain compatible; Bun-written sessions, SQLite, client cache, generated runtime state, and mixed Bun/Go peers have no carryover or rollout guarantee. Future protocol negotiation, trust policy, resource limits, schema hardening, credential permissions, and process-management improvements do not block parity. A narrow Wayfinder decision is required only when an in-scope Go-to-Go, `config.json`, FRP, or native-target probe fails.
