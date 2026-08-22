# Choose safe and bounded contracts for the Diff service

Type: grilling
Status: resolved
Blocked by: 06

## Question

Before the Go Diff server is approved, which corrected contracts replace the legacy service's unsafe or under-specified behavior? Decide the `--public` trust/authentication and warning model; Host, Origin, DNS-rebinding, and no-Origin policy across REST, SSE, Refresh, and MCP; Blob and confirmed-content size/concurrency/Range limits; SSE connection/backpressure/heartbeat behavior; filter-bound and range-checked cursors; fail-closed Target `.gitignore` symlink/mutation and invalid-path handling; reserved-route versus HTML fallback behavior; and the default behavior of REST search without status filters. Preserve loopback usability and Origin-less local MCP clients without copying unauthenticated LAN exposure or unbounded memory work by accident. Record exact HTTP/MCP errors, limits, and tests for the selected policy.

## Comments

## Answer

Closed as out of scope for the parity-first Go release. Diff retains the CLI, REST, SSE, MCP, filesystem, cache, and React-facing behavior recorded in its inventory. New trust, Host/Origin, resource-bound, cursor, symlink, search-default, and fallback policies are post-parity hardening rather than migration prerequisites. Only a demonstrated Go implementation incompatibility may create a narrower decision.
