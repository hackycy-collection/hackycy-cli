# G9 Restart Commit and Darwin Snapshot

- Commit: `e916cb1ad5b7b5ef5a5b949647e0145208adb167`.
- Host: macOS 26.5.2 (25F84), Darwin arm64 (`uname -a`: `Darwin hackycydeMacBook-Pro.local 25.5.0 Darwin Kernel Version 25.5.0: Tue Jun  9 22:28:29 PDT 2026; root:xnu-12377.121.10~1/RELEASE_ARM64_T6030 arm64`).
- Go host: `go version go1.27.0 darwin/arm64`.
- Required Go: `go version go1.26.7 darwin/arm64`.
- Required command environment: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0`.
- Source worktree: clean outside the mutable `.scratch/cli-structure-refactor/` runbook and evidence area.
- Web generated files: 428 files under `web/dist` (ignored).
- 7-Zip payloads: 14 files under `internal/sevenzipruntime/payload` (ignored).
- Docker availability for Linux evidence: Docker Desktop server `linux/arm64 28.1.1`; tool image `g9-linux-evidence:e916cb1` (`sha256:9a0be9b0bf4567aa682a32b7c4d83aea44b51bbdbffd34dfd70219e93c10e056`).
- Native Windows execution host: not available in the current local environment.

All preceding G9 evidence named an older commit and was discarded before this snapshot. Result: PASS. This is the immutable G9 restart baseline.
