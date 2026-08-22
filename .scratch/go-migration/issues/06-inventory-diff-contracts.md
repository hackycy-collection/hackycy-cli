# Inventory diff compatibility contracts

Type: task
Status: resolved

## Question

Inventory the complete observable contract of `diff`: CLI parsing, directory traversal, glob and gitignore semantics, immutable snapshots and refresh behavior, REST routes, SSE lifecycle, browser launch and shutdown, static asset caching, MCP transport and tool schemas, error mapping, concurrency, large-file limits, security assumptions, and existing tests. Identify every Bun or TypeScript-specific behavior whose Go replacement could alter clients or the active React application.

## Comments

- 2026-08-22, scope correction: Diff policy and safety defects are retained as observed first-release behavior and post-parity backlog. They no longer block Vite embedding, module layout, or command migration.

## Answer

The completed inventory is [Diff command compatibility inventory](../inventories/diff-contracts.md). It records the CLI and lifecycle contract, directional scope and exact exclusion layers, fail-closed filesystem/snapshot behavior, query and content limits, REST/SSE/MCP wire contracts, active React flows, Bun/TypeScript replacement hazards, current test baseline, and the Go/frontend/native tests required before `diff` can be declared migrated.

The inventory confirms that Diff is a high-risk migration which must follow the shared Vite/Go web path and retain one deep immutable Comparison Workspace behind HTTP and MCP adapters. Unauthenticated `--public` exposure, incomplete Host/Origin protection, unbounded Blob/content and SSE work, filter-independent cursors, Target `.gitignore` symlink/mutation races, ambiguous REST search defaults, and broad HTML fallback remain parity facts and later hardening candidates rather than first-release policy decisions.
