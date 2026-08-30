# G9 Linux Clean Workspace Preparation

- Commit: `e916cb1ad5b7b5ef5a5b949647e0145208adb167`.
- Host: Docker Debian 12, Linux arm64; Go `1.26.7`.
- Workspace: fresh `git clone --no-local --no-hardlinks /source /workspace` in disposable volume `g9-linux-workspace-e916cb1`.
- Runtime inputs: copied the ignored `web/dist` (428 files) and `internal/sevenzipruntime/payload` (14 files) from the read-only source mount.
- Web dependency command: `pnpm --store-dir /pnpm-store --dir /workspace/web install --frozen-lockfile`, using dedicated volume `g9-linux-pnpm-store-e916cb1`.
- Go cache preparation: `GOMODCACHE=/g9-go-cache/mod GOCACHE=/g9-go-cache/build GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go mod download`, using dedicated volume `g9-linux-go-cache-e916cb1`.
- Clean-state checks: `git -C /workspace status --short` was empty and `git -C /workspace diff --check` exited zero.
- Raw output: `g9-linux-workspace-preparation.log`.

Result: PASS. The workspace is eligible for Linux G9 commands; this is preparation evidence only.
