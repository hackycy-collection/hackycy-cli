# Inventory upgrade and release-artifact compatibility contracts

Type: task
Status: resolved

## Question

Inventory the hidden updater path, public `upgrade` command, installer-visible version output, GitHub release lookup, artifact naming, SHA-256 verification, macOS quarantine handling, Unix replacement, Windows detached replacement and retry behavior, rollback/status files, failure recovery, embedded resources, and all six platform targets. Separate contracts that the Go binary must preserve from existing defects, and define the local artifact tests required even though CI, Docker, and release workflow edits are out of scope.

## Comments

- 2026-08-22, scope correction: updater defects remain first-release parity facts and later hardening backlog. Only a demonstrated Bun-state, installer, executable-replacement, or target-platform incompatibility may interrupt `upgrade` migration.
- 2026-08-22, transition-scope amendment: the first Go installation requires the active installer scripts or manual binary replacement. Bun-to-first-Go `ycy upgrade` and Legacy Update State are no longer supported or tested; the public Go Upgrade leaf covers Go-to-later-Go only.

## Answer

The complete compatibility inventory is [Upgrade and release-artifact compatibility inventory](../inventories/upgrade-artifact-contracts.md). It specifies the public `upgrade` and plain version surface, GitHub latest-release and SHA-256 chain, exact six-artifact/GOOS/GOARCH mapping, stable installer paths, target-specific embedded web and 7-Zip contents, FRP exclusion, Bun detached transaction/state behavior, macOS and Windows replacement details, and the local unit, installer, Artifact Set, rolling-cutover, and native-platform gates required before the Go build can replace the Bun build.

The verified legacy baseline is 11 passing, 1 Windows-only conditional skip, and 0 failures. The `v0.0.47` detached-updater and older Windows bridge remain historical evidence, not first-Go acceptance gates. Existing Bun users install the first Go binary through `scripts/install.sh`, `scripts/install.ps1`, or manual replacement; Go does not consume the unversioned Bun-written update state. All six files remain standalone and `CGO_ENABLED=0`, embed the Vite MPA and target 7-Zip 26.02 runtime/license, and keep FRP as a verified runtime acquisition.

The hidden updater's caller-directed replacement, unlocked concurrency, unversioned and non-durable state, blocking or ignored malformed states, unbounded downloads, exit-0 failures, `--version` contamination, missing native crash coverage, and PowerShell x64 selection remain historical findings. The Go Upgrade leaf reproduces the public Go-to-Go command and target-native replacement behavior; only a focused Go-to-Go or required-target probe may interrupt its integration.
