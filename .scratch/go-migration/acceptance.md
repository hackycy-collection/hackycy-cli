# Go migration Acceptance Ledger

First Go release: `v0.1.0`
Migration build identity: `0.0.0-dev`
Roadmap: [Approve the command-by-command migration roadmap](issues/16-approve-command-migration-roadmap.md)
Status: Foundation Gate integrated; later Units and release evidence remain pending

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
| `export env` | pending | Core inventory | - | - | All selected gates | - |
| `appconfig` foundation | pending | Core inventory | - | - | All selected gates | - |
| `config fork list` | pending | Core inventory | - | - | All selected gates | - |
| `config fork add` | pending | Core inventory | - | - | All selected gates | - |
| `config fork remove` | pending | Core inventory | - | - | All selected gates | - |
| `config cm list` | pending | Core inventory | - | - | All selected gates | - |
| `config cm add` | pending | Core inventory | - | - | All selected gates | - |
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
