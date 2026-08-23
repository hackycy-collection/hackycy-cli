# Go migration Acceptance Ledger

First Go release: `v0.1.0`
Migration build identity: `0.0.0-dev`
Roadmap: [Approve the command-by-command migration roadmap](issues/16-approve-command-migration-roadmap.md)
Status: Foundation Gate, export env, and appconfig foundation integrated; later Units and release evidence remain pending

## Recording rules

- Allowed Unit states are `pending`, `integrated`, and `release-accepted`.
- A public command remains unregistered while its Unit is `pending`. Preparation Slices may be committed only when their applicable checks pass and the command remains externally unreachable.
- `integrated` requires focused inventory-derived parity tests, the Complete Gate, `make build`, one current-host standalone artifact smoke test, and all six `CGO_ENABLED=0` cross-builds. Record the Integration Commit and evidence summary.
- `release-accepted` additionally requires every applicable native OS/architecture, browser, protocol, persistence, process, installer, updater, embedded-payload, and Artifact Set gate.
- Record concise commands/results, date, host OS/architecture, candidate artifact SHA-256, outstanding native work, and any Wayfinder compatibility-decision link. Link durable logs when useful; do not paste large raw logs.
- Never record credentials, tokens, cookies, user data, private endpoints, external Bun executables, or other secrets. The active suite never executes `legacy/bun/`.
- A failed in-scope parity probe stops the serial integration lane at that Unit until a linked Wayfinder decision is complete. Do not mark a skipped or waived test as passed.

## Sources

- [Global, configuration, and local command inventory](inventories/core-command-contracts.md)
- [Git command inventory](inventories/git-command-contracts.md)
- [Diff inventory](inventories/diff-contracts.md)
- [FS inventory](inventories/fs-contracts.md)
- [Tunnel inventory](inventories/tunnel-contracts.md)
- [Upgrade and release-artifact inventory](inventories/upgrade-artifact-contracts.md)
- [Legacy archive and cutover choreography](issues/15-define-archive-cutover.md)
- [Go module seams and project layout](issues/13-choose-go-module-seams.md)
- [Vite-to-Go unconditional embed path](issues/11-prove-vite-go-embed-path.md)
- [Mixed-project Git hook policy](issues/14-choose-mixed-project-hook-policy.md)
- [CGO-free FS thumbnail compatibility path](issues/21-research-cgo-free-fs-thumbnails.md)

## Unit ledger

| Migration Unit | State | Owning source | Integration commit and evidence | Native/artifact evidence | Outstanding gates | Compatibility decision |
| --- | --- | --- | --- | --- | --- | --- |
| Foundation Gate | integrated | Cutover, CLI, embed, layout, hooks | Cutover Commit (this commit): 2026-08-23 clean staged checkout passed `make bootstrap`, hook install/doctor, `make check`, `make build`, host CLI smoke, `make cross-build`, and the actual Lefthook pre-commit. | macOS arm64 host; Go 1.26.7 with `CGO_ENABLED=0`; all three embedded shells validated at startup; fixed Mach-O x64/arm64, static ELF x64/arm64, and PE x64/arm64 artifacts inspected; host links only system libraries. | Release-accepted Artifact Gate and later applicable native evidence | - |
| `export env` | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test ./internal/commands/exportenv ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; host binary help/version, named merge/stdout, cwd-relative output, missing-file, missing-environment, and EOF-cancellation probes passed. | macOS arm64 host (Darwin 25.5.0); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `26f92bea2312947337ef5eaad938e62f4423d45284f293792f62c2bb7403dac4`; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts. | Release-accepted Artifact Gate and later applicable native evidence | - |
| `appconfig` foundation | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test -count=1 ./internal/appconfig ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; direct current/legacy encrypted config, ordering, semantic concurrency, publication failure, architecture ownership, and no-command evidence passed. | macOS arm64 host (Darwin 25.5.0); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `0ec04fdcee79418cfd1c8cf887c9fbb46c147fb68aef885b238728c21e92f437`; native macOS machine-ID/lock/permission/replacement tests passed; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts with matching target metadata. | Release-accepted Artifact Gate and native Linux/Windows machine-ID, locking, permission, and replacement execution evidence | [data scope](issues/12-choose-data-compatibility-mechanisms.md), [module ownership](issues/13-choose-go-module-seams.md), [parity policy](issues/17-choose-corrected-core-command-contracts.md) |
| `config fork list` | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test ./internal/commands/config/fork ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; a standalone binary test covered ordered safe field projection, defaulted scheme, ciphertext preview, plaintext/full-ciphertext non-disclosure, and the absence of `add`/`remove` placeholders. | macOS arm64 host (Darwin); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `a44070e5198de7f465373be09bd17fc58779de42c4711bd821212dc493901057`; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts with matching target CPU inspection. | Release-accepted Artifact Gate and later applicable native evidence | [module ownership](issues/13-choose-go-module-seams.md), [parity policy](issues/17-choose-corrected-core-command-contracts.md) |
| `config fork add` | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test -count=1 ./internal/commands/config/fork ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; standalone tests cover prompt order/validation, cancellation, silent overwrite, encrypted persisted shape, host normalization, save failure, typed binding, and absent `remove`. | macOS arm64 host (macOS 26.5.1, Darwin 25.5.0); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `8b65f5aee6a754709df99e28a26adc4b4b6b464c1994d11da37def134c17ff1a`; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts with matching target CPU inspection. | Release-accepted Artifact Gate and later applicable native evidence | [data scope](issues/12-choose-data-compatibility-mechanisms.md), [module ownership](issues/13-choose-go-module-seams.md), [parity policy](issues/17-choose-corrected-core-command-contracts.md) |
| `config fork remove` | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test -count=1 ./internal/commands/config/fork ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; standalone tests cover ordered selection, empty, cancellation, declined/confirmed removal, safe output, real appconfig mutation, and the complete Fork leaf set without placeholders or legacy dispatch. | macOS arm64 host (Darwin 25.5.0); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `ce7699c9bf1912725891a4678c98d8700611f988545e598589584b68b4a3113e`; isolated empty-config host smoke left `config.json` absent; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts with matching target CPU inspection. | Release-accepted Artifact Gate and later applicable native evidence | [data scope](issues/12-choose-data-compatibility-mechanisms.md), [module ownership](issues/13-choose-go-module-seams.md), [parity policy](issues/17-choose-corrected-core-command-contracts.md) |
| `config cm list` | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test -count=1 ./internal/commands/config/cm ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; module and standalone tests cover empty and populated lists, stored insertion order, default marking, field equivalence, API-key non-disclosure, and only the real CM list route. | macOS arm64 host (Darwin 25.5.0); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `f0c069062d40d8effb813123879a0aef62493bf984df58fb1bd1d387d7c30eaa`; empty-config host smoke printed `0.0.0-dev`, returned the add hint without creating `config.json`, and `config cm --help` listed only `list`; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts with matching target CPU inspection. | Release-accepted Artifact Gate and later applicable native evidence | [data scope](issues/12-choose-data-compatibility-mechanisms.md), [module ownership](issues/13-choose-go-module-seams.md), [parity policy](issues/17-choose-corrected-core-command-contracts.md) |
| `config cm add` | integrated | Core inventory | Integration Commit (this commit): 2026-08-23 focused `go test -count=1 ./internal/commands/config/cm ./internal/cliapp ./cmd/ycy`, `make check`, and `make build` passed; module and standalone tests cover prompt order/validation, cancellation, normalized encrypted persistence, first-default behavior, silent overwrite, save failure, typed binding, and only the real CM list/add routes. | macOS arm64 (macOS 26.5.1, Darwin 25.5.0); Go 1.26.7 with `CGO_ENABLED=0`; host `build/ycy` SHA-256 `02098459ab58b5a7a635198ebc72b10f2e5092d651d95b92051468cb26563231`; isolated host smoke added and listed an encrypted CM profile without API-key disclosure; `make cross-build` produced six nonempty Mach-O/ELF/PE artifacts with matching target CPU inspection. | Release-accepted Artifact Gate and later applicable native evidence | [data scope](issues/12-choose-data-compatibility-mechanisms.md), [module ownership](issues/13-choose-go-module-seams.md), [parity policy](issues/17-choose-corrected-core-command-contracts.md) |
| `config cm use` | pending | Core inventory | - | - | All selected gates | - |
| `config cm set` | pending | Core inventory | - | - | All selected gates | - |
| `config cm remove` | pending | Core inventory | - | - | All selected gates | - |
| `config cm test` | pending | Core inventory | - | - | All selected gates | - |
| `rm` | pending | Core inventory | - | - | All selected gates | - |
| `run` | pending | Core inventory | - | - | All selected gates | - |
| `git heat` | pending | Git inventory | - | - | All selected gates | - |
| `git pulse` | pending | Git inventory | - | - | All selected gates | - |
| `zip` | pending | Core inventory | - | - | All selected gates | - |
| `git fork` | pending | Git inventory | - | - | All selected gates | - |
| `git cm` | pending | Git inventory | - | - | All selected gates | - |
| Web Readiness Gate | pending | Embed, Diff/FS/Tunnel inventories | - | - | All selected gates | - |
| `diff` | pending | Diff inventory | - | - | All selected gates | - |
| FS Foundation | pending | FS inventory | - | - | All selected gates | - |
| `fs` | pending | FS inventory | - | - | Selected pure-Go thumbnail integration and all selected gates | - |
| Tunnel Foundation | pending | Tunnel inventory | - | - | All selected gates | - |
| `tunnel server` | pending | Tunnel inventory | - | - | All selected gates | - |
| `tunnel connect` | pending | Tunnel inventory | - | - | All selected gates | - |
| `upgrade` | pending | Upgrade inventory | - | - | All selected Go-to-Go gates | - |
| Final Artifact Gate | pending | Upgrade/artifact inventory and roadmap | - | - | Complete six-target matrix | - |

## Research prerequisite

| Research | Status | Evidence | Blocks |
| --- | --- | --- | --- |
| [Research a CGO-free FS thumbnail compatibility path](issues/21-research-cgo-free-fs-thumbnails.md) | resolved | [15-case, six-target report](research/21-cgo-free-fs-thumbnails.md) | Cleared; its selected engine and self-exec worker contract are mandatory for the FS thumbnail Slice |

## Native artifact matrix

| Target | Required artifact | State | Host/date | Candidate SHA-256 | Evidence | Outstanding gates |
| --- | --- | --- | --- | --- | --- | --- |
| macOS x64 | `ycy-macos-x64` | pending | - | - | - | All selected gates |
| macOS arm64 | `ycy-macos-arm64` | pending | - | - | - | All selected gates |
| Linux x64 | `ycy-linux-x64` | pending | - | - | - | All selected gates |
| Linux arm64 | `ycy-linux-arm64` | pending | - | - | - | All selected gates |
| Windows x64 | `ycy-windows-x64.exe` | pending | - | - | - | All selected gates |
| Windows arm64 | `ycy-windows-arm64.exe` | pending | - | - | - | All selected gates |

## Final Artifact Gate

- [ ] Every Migration Unit is `integrated` and every applicable milestone-native result is recorded.
- [ ] The candidate starts from a clean checkout without `web/dist`, dependencies, caches, downloaded payloads, binaries, or prior artifact output.
- [ ] Bootstrap succeeds; one Vite production graph is built and structurally verified; the offline Complete Gate passes.
- [ ] Exactly six `CGO_ENABLED=0` artifacts are emitted with the fixed public basenames and plain Release Identity `0.1.0`.
- [ ] Mach-O/ELF/PE formats, CPUs, Go build metadata, nonzero sizes, and lack of Bun/Node runtime dependencies are verified.
- [ ] Every artifact contains all three Web shells and reachable assets, only its target 7-Zip 26.02 runtime/license, and the FRP manifest without FRP executable bytes.
- [ ] The selected thumbnail modules and sums are exact, no codec helper/system lookup is present, and release third-party documentation reproduces their required notices and patent grants.
- [ ] `SHA256SUMS` contains exactly one verified entry for every artifact and is accepted by installer and Upgrade parsers.
- [ ] Native standalone CLI, production-browser, FS/7-Zip/thumbnail, Go-only Tunnel/FRP, installer, and Go-to-Go Upgrade gates pass on all applicable targets.
- [ ] Every applicable Unit and target is `release-accepted`.
- [ ] Generated Web output, dependencies, source maps, caches, downloads, binaries, checksums, and release staging are untracked.
- [ ] The later Go release workflow is manual-dispatch only and invokes this same gate before the first `v0.1.0` release.

## Evidence entry template

```text
Unit:
State:
Inventory/version:
Implementation commit:
Date:
Host OS/architecture:
Candidate artifact and SHA-256:
Focused test commands/results:
Complete Gate/build/cross-build results:
Native/browser/protocol/payload results:
Outstanding gates:
Compatibility decision:
Notes (no secrets):
```
