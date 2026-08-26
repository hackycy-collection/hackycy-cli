# Chart the Bun CLI to Go release-artifact migration

Label: wayfinder:map
Status: resolved

## Destination

Produce an execution-ready, parity-first route that turns ycy from a Bun-hosted CLI into a standard, CGO-free Go CLI while retaining the current Bun implementation's observable command behavior, active React frontends, `config.json`, Go-only protocol behavior, and six standalone release artifacts. The route fixes the foundational Go structure, Vite multi-page build, root Git hooks, legacy cutover, command order, and acceptance process; Bun-written non-config runtime state, Bun-to-first-Go self-upgrade, and mixed Bun/Go Tunnel deployment are explicit transition exclusions rather than redesign invitations.

## Notes

- Domain: CLI runtime migration, compatibility engineering, Go project architecture, embedded multi-page web applications, local developer quality gates, and cross-platform release artifacts.
- Consult `domain-modeling` when terminology or compatibility contracts change, `codebase-design` when choosing module interfaces and seams, `research` for external facts, and `prototype` for cheap compatibility spikes.
- `legacy/bun/` will contain the frozen Bun CLI, TypeScript servers, old build scripts, and old tests as read-only implementation reference. Active code must not import from or dispatch into legacy.
- The active frontend is one `web/` package using pnpm, Vite, and React. It retains separate `fs`, `diff`, and `tunnel-server` HTML entries and shared React code. Bun is absent from the active repository toolchain and runtime.
- The canonical full build is `make build`: Vite builds and verifies the web assets first, then Go unconditionally embeds them and emits the standalone binary. There is no build-tag/stub variant, so clean-checkout Go compilation that reaches the active web package requires generated `web/dist`; the output remains ignored and uncommitted.
- The Go release path is a hard `CGO_ENABLED=0` target. `cmd/ycy` is the composition root; internal code is organized as deep modules. Do not add generic `utils`, `services`, `interfaces`, or speculative `pkg` layers.
- The first Go release is a Bun behavior port after a supported installation, not a hardening release. Preserve command names, options, defaults, parser quirks, prompts, exit behavior, machine-readable output, fresh-Go persistent behavior, Go-to-Go protocols, side effects, and core interaction flows, including known defects where they are observable. Only `config.json` has a Bun-written direct-read guarantee; the explicit transition exclusions below supersede broader cross-runtime compatibility findings. Human-facing layout, color, wording, and diagnostic logging may change only where they are not machine-consumed.
- Go-idiomatic internal structure, streaming, ownership, and dependency choices may differ from Bun when focused tests prove that the observable contract remains unchanged. Parity targets behavior and artifacts, not a line-by-line TypeScript translation.
- A **compatibility exception** exists only when concrete implementation evidence shows that in-scope Bun/Node semantics cannot be reproduced under Go, `CGO_ENABLED=0`, a required target OS/architecture, an available maintained dependency, or the existing `config.json` format. Stop the serial integration lane and complete one narrow Wayfinder decision before changing the affected integration plan. Bun-written non-config state, Bun-to-first-Go Upgrade, and mixed Bun/Go Tunnel peers are already out of scope and do not open exceptions. Do not ask speculative parameter-safety, hardening, or product-policy questions.
- Known security, resource-bound, atomicity, validation, and error-status defects remain documented in the inventories as post-parity backlog. They do not block the first Go release unless they are themselves the source of a proven Bun-to-Go incompatibility.
- Only `~/.ycy-cli/config.json`, including Fork, CM, and remembered Tunnel connections, must remain directly readable. Bun-written FS sessions, Tunnel sessions/SQLite/client cache/generated runtime state, and Legacy Update State receive no detection, migration, deletion, fixture, or release gate. Go follows its normal startup path, emits useful diagnostics on failure, and leaves recovery to the internal operator; only an evidenced `config.json` mismatch may require an automatic, backed-up, failure-detectable compatibility decision.
- Preserve the existing macOS, Linux, and Windows x64/arm64 artifact names and standalone distribution model. Preserve embedded web and 7-Zip behavior and the current verified FRP acquisition model.
- `ycy run` continues to recognize and invoke Bun in user projects when Bun lockfiles select it; this is command behavior, not a ycy project dependency. Its first-release argv behavior follows the legacy implementation, including currently rejected passthrough forms, unless a parity probe proves that behavior cannot be reproduced.
- The first Bun-to-Go installation uses `scripts/install.sh`, `scripts/install.ps1`, or manual binary replacement. The public Go `ycy upgrade` command is retained only for Go-to-later-Go replacement and does not consume Bun Legacy Update State.
- Tunnel retains protocol v3, its wire platform mapping, and pinned FRP fields for Go client to Go server behavior. Bun-client/Go-server and Go-client/Bun-server coexistence are not supported or tested.
- Migration tests are written by consulting `legacy/bun/` command-by-command. Do not create or maintain a separate golden/black-box fixture corpus and do not run legacy Bun as part of the active Go test suite.
- Git pre-commit checks use the selected single repository-root, pinned, Grafana-style Lefthook policy covering both Go and the pnpm/Vite frontend; no Bun hook remains active.
- The accepted ordering principle is: establish foundations, prove one low-risk vertical command, then follow dependency and risk order through configuration, local commands, Git commands, shared web infrastructure, `diff`, `fs`, `tunnel`, `upgrade`, and the final artifact gate.
- The canonical serial integration lane and its `pending -> integrated -> release-accepted` evidence are fixed by [Approve the command-by-command migration roadmap](issues/16-approve-command-migration-roadmap.md) and tracked in the [Go migration Acceptance Ledger](acceptance.md). A failed in-scope compatibility probe stops the lane for a Wayfinder decision; later Units are not skipped into the Active Tree.
- Migration binaries report plain `0.0.0-dev`. The first Go release candidate and all six artifacts report plain `0.1.0`; the first Go tag is `v0.1.0`.
- Each command is migrated from its inventory and frozen legacy reference. Implement the parity tests first, port the behavior, and open a compatibility-exception ticket only after a test or focused probe demonstrates an actual Node/Bun-to-Go mismatch.
- The atomic cutover may relocate the existing CI workflows, Dockerfile, Docker ignore file, and deployment definitions byte-for-byte into `legacy/bun/` so obsolete Bun automation is inactive. Redesigning, replacing, or reactivating them remains outside this map; after migration, the release workflow is rebuilt for manual dispatch only.

## Decisions so far

<!-- Closed child tickets are indexed here by name. -->

- [Research the pure-Go toolchain and dependency baseline](issues/01-research-pure-go-toolchain.md): Go 1.26.7 and a pinned cgo-free candidate graph compile for all six targets, with runtime and format-parity gates retained for conditional dependencies.
- [Research quality gates for a Go and Vite mixed project](issues/02-research-mixed-project-quality-gates.md): primary-source evidence supports root-owned, pinned Lefthook with fast staged checks and a separate complete `make check` gate.
- [Research the Vite MPA to Go embedding contract](issues/03-research-vite-go-embedding.md): one Vite MPA should emit three fixed shells and a shared hashed asset tree that Go embeds and serves behind command-specific routes.
- [Inventory global, configuration, and local command contracts](issues/04-inventory-core-command-contracts.md): the global surface and 13 core leaves now have classified compatibility contracts, migration risks, dependency boundaries, and required Go tests.
- [Inventory Git command compatibility contracts](issues/05-inventory-git-command-contracts.md): all four Git leaves have exact Git/provider/mutation baselines, test gaps, migration order, and documented post-parity defects.
- [Inventory diff compatibility contracts](issues/06-inventory-diff-contracts.md): Diff has an end-to-end CLI, filesystem, snapshot, REST/SSE/MCP, React, and Go-replacement parity baseline with native and production-artifact checks.
- [Inventory fs compatibility contracts](issues/07-inventory-fs-contracts.md): FS has a complete fresh-Go behavior baseline and one evidenced CGO-free thumbnail capability question; Bun-written session carryover is no longer a release gate.
- [Inventory tunnel compatibility contracts](issues/08-inventory-tunnel-contracts.md): Tunnel has separate server/client baselines for Go-only protocol v3, persistence, FRP, and native behavior; mixed Bun/Go peers and non-config state carryover are excluded.
- [Inventory upgrade and release-artifact compatibility contracts](issues/09-inventory-upgrade-artifact-contracts.md): Upgrade retains the exact six-artifact standalone model, installer behavior, and Go-to-Go replacement baseline; the first Go install does not use Bun Upgrade or Legacy Update State.
- [Prove the Go CLI compatibility approach](issues/10-prove-go-cli-compatibility.md): a deep CLI App can hide Cobra behind typed command binding across all six targets; concrete argv and exit behavior remains governed by the legacy command inventories.
- [Prove the Vite MPA to Go embed path](issues/11-prove-vite-go-embed-path.md): every ycy binary unconditionally embeds one verified three-shell Vite output tree, while command adapters preserve the distinct compiled-Bun routing contracts.
- [Choose the Go module seams and project layout](issues/13-choose-go-module-seams.md): one lazily created, feature-first root Go module uses explicit composition, evidence-backed shared Modules, owner-local platform files and Adapters, and mechanically enforced dependency rules.
- [Choose the mixed-project Git hook policy](issues/14-choose-mixed-project-hook-policy.md): one root-owned pinned Lefthook Fast Gate covers staged Go/web files, while explicit Make lifecycle commands provide safe Bun-hook migration and the offline Complete Gate across macOS, Linux, and Windows.
- [Define the legacy archive and migration cutover choreography](issues/15-define-archive-cutover.md): one atomic buildable cutover freezes the complete Bun reference, extracts the active React applications, establishes the Go-era foundations and legacy isolation, and withholds release until final compatibility acceptance.
- [Approve the command-by-command migration roadmap](issues/16-approve-command-migration-roadmap.md): one serial, always-buildable lane integrates every leaf through explicit local/native gates, narrows transition compatibility to `config.json` and Go-only operation, and finishes with six `v0.1.0` artifacts before manual release automation.
- [Choose the first-release parity and compatibility-exception policy](issues/17-choose-corrected-core-command-contracts.md): the first Go release reproduces Bun behavior by default; only evidence-backed implementation mismatches may open narrow compatibility decisions, while hardening is deferred.
- [Research a CGO-free FS thumbnail compatibility path](issues/21-research-cgo-free-fs-thumbnails.md): pinned pure-Go AVIF/WebP plus standard image codecs reproduce the Bun thumbnail capability in one binary across all six builds, with self-exec workers preserving hard timeout replacement and no compatibility exception.
- [Decide Windows native acceptance contracts](issues/24-choose-windows-native-acceptance-contract.md): G27 may make the narrowly evidenced Windows executable, path, SQLite URI, DACL, sharing-retry, and Hookctl fixture adaptations required for native acceptance without changing public behavior.
- [Choose the remaining Windows Tunnel/FRP native acceptance contract](issues/25-choose-windows-tunnel-frp-acceptance-contract.md): G27 may make only the approved Tunnel-owned file DACL, path-error, and test-synchronization adaptations, without changing public protocol behavior.
- [Choose the current-host G27 acceptance contract](issues/26-choose-g27-current-host-acceptance-contract.md): the user-approved primary-host-set variant applies native Exit 3/4 evidence only to `windows/amd64` and `darwin/arm64`, keeps the other target rows pending, and does not claim full six-target release readiness.

## Not yet specified

- None. Any future command-specific incompatibility begins only when an implementation test or focused probe demonstrates it; the map's compatibility-exception policy governs that new evidence rather than leaving a speculative planning decision open.


## Out of scope

- Implementing or shipping the Go migration; this map ends at an execution-ready specification.
- Redesigning, replacing, or reactivating CI workflows, Dockerfiles, or deployment definitions, and performing an actual release. Relocating the current files byte-for-byte into the Frozen Archive is part of the cutover; implementing the later manual-only release workflow remains post-migration work.
- Adding new CLI features, changing existing product behavior for convenience, or refactoring unrelated code.
- Preemptive correction of legacy validation, safety, atomicity, resource-bound, error-status, or trust-policy defects before behavior parity is complete.
- [Choose persistent-data compatibility mechanisms](issues/12-choose-data-compatibility-mechanisms.md): only `config.json` has a Bun-written direct-read guarantee; migration and compatibility guarantees for every other Bun-written state format are out of scope.
- [Choose safe and failure-aware contracts for Git commands](issues/18-choose-safe-git-command-contracts.md): Git behavior hardening is deferred; first-release behavior comes from the Git compatibility inventory.
- [Choose safe and bounded contracts for the Diff service](issues/19-choose-safe-diff-service-contracts.md): Diff trust and resource-policy redesign is deferred; first-release HTTP/MCP behavior follows the legacy inventory.
- [Choose safe and bounded contracts for the FS service](issues/20-choose-safe-fs-service-contracts.md): FS safety and limit redesign is deferred; only proven Go capability mismatches may interrupt the FS port.
- [Choose safe and rolling-compatible contracts for Tunnel](issues/22-choose-safe-rolling-tunnel-contracts.md): future negotiation, trust, bounds, and persistence hardening are deferred; only Go-to-Go protocol-v3 and fresh-Go state behavior remain first-release requirements.
- [Choose a safe and recoverable self-update contract](issues/23-choose-safe-self-update-contract.md): updater redesign is deferred; the first Go install uses an installer or manual replacement, while public Upgrade retains Go-to-Go artifact and native replacement behavior only.
- Bun-to-first-Go transition through the Bun `ycy upgrade` command or consumption of Bun Legacy Update State; the first Go install uses an installer or manual replacement.
- Direct-read, migration, backup, or recovery guarantees for Bun-written FS sessions, Tunnel sessions/SQLite/client cache/generated runtime state, or any other data outside `config.json`.
- Bun-client/Go-server or Go-client/Bun-server Tunnel coexistence and rolling deployment; coordinated internal users replace both sides before relying on the Go-only v3 implementation.
