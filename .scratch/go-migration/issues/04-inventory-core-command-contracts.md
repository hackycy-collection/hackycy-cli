# Inventory global, configuration, and local command contracts

Type: task
Status: resolved

## Question

Produce a per-leaf-command compatibility inventory from the frozen implementation for global CLI behavior; `export env`; every `config fork` and `config cm` subcommand; `rm`; `run`; and `zip`. For each, record inputs, option/default/parser behavior, TTY branches, output and exit semantics, filesystem or persistent-state effects, operating-system differences, current Bun/Node/library dependencies, known defects that must not accidentally become requirements, and the Go tests needed to establish contract compatibility. Treat `ycy run` invoking Bun in a user project as retained external command behavior.

## Comments

- 2026-08-22, scope correction: defects remain observed parity facts for the first Go release. They no longer require corrected contracts before migration; only a demonstrated Bun-to-Go implementation mismatch may open a narrow exception.

## Answer

The completed compatibility inventory is [Global, configuration, and local command compatibility inventory](../inventories/core-command-contracts.md). It records the global CLI contract and all 13 scoped leaves, including inputs, parser/default behavior, interactive branches, output and exit semantics, persistent and filesystem effects, platform concerns, Bun/Node replacement points, risk, and the Go tests required for migration.

The inventory confirms that `export env` is the best first vertical slice and that shared configuration storage is a high-risk prerequisite for both `config` and later Git commands because Go must directly read the existing PBKDF2-SHA256/AES-256-GCM secrets, schema variants, and locked file. Unreachable `run` passthrough, silent config replacement, unsafe `rm` targets and success-on-partial-failure, and ZIP path/error/resource hazards remain documented legacy behavior for first-release parity and post-parity hardening, not up-front migration blockers.
