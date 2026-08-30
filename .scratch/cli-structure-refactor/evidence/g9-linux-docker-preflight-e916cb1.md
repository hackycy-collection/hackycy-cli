# G9 Linux Docker Tool-Image Preflight

- Commit to be tested: `e916cb1ad5b7b5ef5a5b949647e0145208adb167`.
- Image: `g9-linux-evidence:e916cb1`.
- Image ID: `sha256:9a0be9b0bf4567aa682a32b7c4d83aea44b51bbdbffd34dfd70219e93c10e056`.
- Runtime: Debian GNU/Linux 12 (bookworm), Linux `aarch64` / arm64.
- Go: `go version go1.26.7 linux/arm64`.
- Node: `v24.20.0`; pnpm: `11.13.0`.
- Browser: `/usr/bin/chromium`, Chromium `151.0.7922.173`.
- Native-run prerequisites: `/etc/machine-id` is nonempty (33 bytes), `/dev/ptmx` is a character device, and `devpts` is mounted with `ptmxmode=666`.
- Supporting tools: GNU Make 4.3 and Git 2.39.5.

Result: PASS. This prepares a disposable Linux/arm64 execution environment only; it is not an ordinary-suite, tagged-acceptance, or Web-acceptance result.
