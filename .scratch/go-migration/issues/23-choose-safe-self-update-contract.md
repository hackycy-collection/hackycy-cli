# Choose a safe and recoverable self-update contract

Type: grilling
Status: resolved
Blocked by: 09, 10

## Question

Before the Go `upgrade` leaf and the Bun-to-Go cutover are approved, which corrected contract replaces the legacy updater's unsafe, under-bounded, and non-recoverable behavior while preserving the six artifact names, installer bridge, checksum/version checks, rollback intent, and direct consumption of Bun-written update state?

Choose the authenticated and isolated internal updater entry; one-transaction cross-process lock; target, staged, backup, updater, symlink/reparse, ownership, permission, and path-identity rules; versioned and durably published state machine; direct-read handling of legacy `pending`, `succeeded`, `succeeded_with_cleanup_warning`, `failed`, and malformed states; abandoned-owner detection and installer recovery; crash-point roll-forward/rollback; updater-copy lifecycle; clean Artifact Self-check and one-time result reporting; semantic-version/release-metadata validation; streaming download, redirect, byte, time, and disk bounds; and exact success/nonzero failure exits.

Also choose the platform replacement architecture and bounded process/file retry behavior for Unix, macOS quarantine, Windows locked executables, MOTW, antivirus sharing violations, ACLs, and native AMD64/ARM64 operation. Preserve the Windows `v0.0.46`-or-earlier installer-once bridge and prove current Bun to first Go plus Go-to-Go replacement without making Bun part of the active test suite.

Coordinate the internal CLI entry with [Prove the Go CLI compatibility approach](10-prove-go-cli-compatibility.md), state adapters/backups/locking mechanisms with [Choose persistent-data compatibility mechanisms](12-choose-data-compatibility-mechanisms.md), verified 7-Zip runtime state with [Choose safe and bounded contracts for the FS service](20-choose-safe-fs-service-contracts.md), and build/cutover ordering with [Define the legacy archive and migration cutover choreography](15-define-archive-cutover.md). Record the exact state transitions, recovery table, observable CLI behavior, limits, platform rules, compatibility window, and adversarial/native tests without modifying CI, Docker, deployment, or release workflows.

## Comments

- 2026-08-22, roadmap scope amendment: Bun-to-first-Go Upgrade and Legacy Update State are removed from first-release acceptance; initial installation uses an installer or manual replacement, and public Upgrade begins at Go-to-Go.

## Answer

Closed as out of scope in its updater-redesign form. The first Go release preserves the public `upgrade` surface for Go-to-later-Go, six artifact names, checksum chain, active installer behavior, and target-native Go replacement behavior from the upgrade inventory. The first Bun-to-Go switch uses an installer or manual replacement; Go does not consume Bun Legacy Update State. A new authenticated transaction, state machine, recovery model, resource policy, and platform hardening are deferred. If an in-scope Go-to-Go replacement behavior cannot be implemented on a required target, that failed scenario stops integration for a narrow Wayfinder decision.
