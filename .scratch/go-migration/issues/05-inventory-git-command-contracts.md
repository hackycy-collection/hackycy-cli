# Inventory Git command compatibility contracts

Type: task
Status: resolved

## Question

Produce a separate compatibility inventory for `git heat`, `git pulse`, `git fork`, and `git cm`. For each command, record CLI and TTY behavior, Git subprocess arguments and parsing assumptions, terminal presentation contracts, provider/config dependencies, filesystem and repository mutations, concurrency and cancellation behavior, security-sensitive filtering, platform differences, current incompatibilities with Go libraries, and tests required before the command can be declared migrated.

## Comments

- 2026-08-22, scope correction: unsafe or contradictory Git behavior remains part of the first-release parity baseline unless a focused Go implementation proves it cannot be reproduced. Hardening is deferred and no longer blocks the migration route.

## Answer

The completed inventory is [Git command compatibility inventory](../inventories/git-command-contracts.md). It records the CLI/TTY behavior, exact Git and provider operations, repository and filesystem effects, concurrency/cancellation gaps, platform risks, security boundaries, current test coverage, and required Go tests for `git heat`, `git pulse`, `git fork`, and `git cm`.

The migration order is now clear: port `git heat` first, then `git pulse`; shared encrypted configuration must precede `git fork` and `git cm`; and `git cm` should be the final Git leaf. Destructive pre-download fork overwrite, TAR traversal and token-in-argv exposure, repository-external symlink reads and secret exposure, and submodule evidence gaps remain explicitly documented for parity tests and later hardening rather than receiving new contracts during the port.
