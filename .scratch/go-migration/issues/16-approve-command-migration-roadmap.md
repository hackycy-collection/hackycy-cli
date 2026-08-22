# Approve the command-by-command migration roadmap

Type: grilling
Status: resolved
Blocked by: 04, 05, 06, 07, 08, 09, 15, 17

## Question

Assemble and approve the execution-ready order and prerequisites for foundations, `export env`, configuration, `rm`, `run`, `git heat`, `git pulse`, `zip`, `git fork`, `git cm`, shared web infrastructure, `diff`, `fs`, `tunnel server`, `tunnel connect`, `upgrade`, and final artifact verification. Each leaf uses its inventory and `legacy/bun/` as the behavior baseline, adds focused Go tests before or with the port, and completes when the legacy observable contract passes under the first-release artifact constraints. Dependencies and internal ownership are selected when that leaf needs them. A failed parity test or focused implementation probe may open one narrow compatibility-exception ticket; known defects and speculative hardening never block the sequence.

## Comments

- 2026-08-22, grilling round 1: selected one canonical serial integration lane with only research, focused probes, and unmerged test preparation allowed in parallel; distinguished an `Integrated` command, which may be registered only after focused parity tests, the Complete Gate, a host standalone smoke test, and all-six cross-builds, from `Release Accepted`, which additionally requires every applicable native/platform/browser/protocol/persistence/artifact gate; fixed the Foundation Gate at the archived/isolated Bun cutover plus global Go CLI behavior, unconditional three-entry Web embedding, Make/pnpm/Lefthook/architecture checks, Release Identity injection, fixed artifact naming, and the CGO-free cross-build skeleton, while deferring `appconfig`, 7-Zip/FRP payloads, business commands, placeholders, and release automation; selected one reviewable passing integration commit per Migration Unit, permitting only smaller behavior-neutral passing preparation commits and forbidding committed red tests, half-registered commands, or legacy fallback.
- 2026-08-22, grilling round 2: selected the canonical lane `Foundation Gate -> export env -> Configuration Block -> rm -> run -> git heat -> git pulse -> zip -> git fork -> git cm -> Web Readiness Gate -> diff -> FS Foundation plus fs -> Tunnel Foundation -> tunnel server -> tunnel connect -> upgrade -> Final Artifact Gate`; selected leaf-by-leaf configuration integration after an `appconfig` compatibility foundation, in read-before-write/risk order of `config fork list/add/remove` then `config cm list/add/use/set/remove/test`; redefined the already-founded shared Web work as a Web Readiness Gate rather than a speculative common service Module; selected milestone-native verification with an Acceptance Ledger and a final aggregate rerun instead of treating cross-compilation as native proof; selected parallel CGO-free thumbnail research immediately after the Foundation Gate, nonblocking before FS but mandatory before its thumbnail phase and `fs` integration.
- 2026-08-22, grilling round 3 (partial): selected passing Preparation Slices for large Modules while their command remains unregistered and externally unreachable, followed by one end-to-end Integration Commit; selected `.scratch/go-migration/acceptance.md` as the secret-free Acceptance Ledger containing unit state, inventory, implementation commit, test-result summary, host/target facts, artifact digest, outstanding native gates, and compatibility-decision links; selected a hard stop at any proven incompatibility, with a Wayfinder decision completed before the affected Integration Commit or any changed compatibility plan, while independent research/probes/unmerged preparation may continue. Rejected a Bun-to-first-Go Upgrade interoperability procedure on the basis that installation and Upgrade replace the executable; clarification remains open on whether this removes only the cross-runtime self-update bridge or also the separately specified Tunnel mixed-peer and persisted-state compatibility gates.
- 2026-08-22, grilling round 3 scope clarification: selected installer or manual binary replacement as the only supported Bun-to-first-Go cutover, with no Bun `ycy upgrade` rolling bridge; narrowed guaranteed Bun-written durable-data compatibility to `~/.ycy-cli/config.json` only; removed Bun-client/Go-server and Go-client/Bun-server Tunnel interoperability from first-release acceptance because the CLI is used by a coordinated internal audience. These choices supersede the earlier broad legacy-state, Bun-to-Go self-update, and mixed-peer gates; the treatment of encountered non-config legacy state, the retained Go-era Upgrade surface, and the Go-only Tunnel protocol baseline remain to be fixed before the roadmap can close.
- 2026-08-22, grilling round 3 completion: selected ordinary Go startup with useful diagnostic logging and operator-managed recovery for any encountered Bun-written state outside `config.json`, with no compatibility detection, migration, deletion, fixture, exception, or release gate; retained public `ycy upgrade` for Go-to-later-Go plus installer/manual installation of Go, while excluding Bun-to-first-Go Upgrade and Legacy Update State; retained protocol v3 behavior and exact FRP/platform fields for Go client to Go server only, with no mixed-runtime artifact or rollout test.
- 2026-08-22, grilling round 4: selected passing Preparation Slices for Diff in the order Comparison Workspace/filesystem/glob/gitignore/snapshot, query/text/difference/blob, REST/SSE, MCP, CLI/production browser, then standalone/native integration; for FS in the order fresh-Go file sessions/root-confined workspace, read-only HTTP, auth/session/edit/operations, uploads, remote download, 7-Zip, research-gated thumbnails, tasks/SSE/React, then standalone/native integration; for Tunnel in the order config/fresh-Go state/protocol-v3/FRP/locks/process foundation, server persistence/domain, server HTTP/SSE/React/frps/agent gateway and scripted-Go-client integration, then connect resolution/cache/reconciliation/frpc and real Go-to-Go/FRP integration; and for Upgrade in the order release/artifact resolution, candidate verification, Go-owned updater state/hidden entry, native replacement/rollback/cleanup, installer fixtures, local Go-to-Go integration, then final six-target rerun. No large command is registered before its final Integration Commit, and none of these sequences adds a Bun-state or mixed-runtime gate.
- 2026-08-22, grilling round 5: retained only the already-evidenced `appconfig`, `filesession`, and `webassets` shared Modules up front, with every other capability owned by its first consumer and extracted at a later real caller only when invariants match and the Interface removes meaningful complexity; selected the clean-checkout Final Artifact Gate covering the complete Acceptance Ledger, one verified Vite graph, Complete Gate, exactly six CGO-free files, format/CPU/Release Identity/runtime-dependency inspection, embedded Web and target 7-Zip/license inspection, FRP-manifest-without-binaries proof, `SHA256SUMS`, all native CLI/browser/FS/Go-only-Tunnel/installer/Go-to-Go-Upgrade gates, and an untracked-output check, followed only then by a manual-dispatch Go release workflow; selected `v0.1.0` as the first Go release identity, with migration builds reporting explicit `0.0.0-dev` and every release candidate reporting plain `0.1.0`.
- 2026-08-22, final confirmation: approved the complete serial roadmap, state and commit model, leaf and Preparation Slice ordering, narrowed transition compatibility, evidence ledger, shared-module extraction rule, Final Artifact Gate, manual release-workflow timing, and first Go release identity without further changes.

## Answer

Execute the migration through one canonical serial integration lane. Research, focused probes, and unmerged test preparation may run in parallel, but the Active Tree integrates one **Migration Unit** at a time in the order below. Every committed state remains buildable and passes all checks applicable to the code then present.

```text
Foundation Gate
-> export env
-> appconfig compatibility foundation
-> config fork list
-> config fork add
-> config fork remove
-> config cm list
-> config cm add
-> config cm use
-> config cm set
-> config cm remove
-> config cm test
-> rm
-> run
-> git heat
-> git pulse
-> zip
-> git fork
-> git cm
-> Web Readiness Gate
-> diff
-> FS Foundation
-> fs
-> Tunnel Foundation
-> tunnel server
-> tunnel connect
-> upgrade
-> Final Artifact Gate
```

The already resolved [Research a CGO-free FS thumbnail compatibility path](21-research-cgo-free-fs-thumbnails.md) remains scheduled as parallel evidence immediately after the Foundation Gate in the execution history. It does not block earlier commands or early FS Preparation Slices; its pinned pure-Go engine and self-exec worker contract are now fixed prerequisites for the thumbnail Slice and for `fs` to become Integrated.

### State and commit model

Use exactly these acceptance states in the linked [Go migration Acceptance Ledger](../acceptance.md):

- **Pending:** the Migration Unit is not yet integrated. Passing internal work may exist, but its public command remains absent.
- **Integrated:** focused parity tests derived from the owning inventory and `legacy/bun/`, the repository Complete Gate, `make build`, a current-host standalone artifact smoke test, and all six `CGO_ENABLED=0` cross-builds pass. A public leaf enters `cliapp` and `cmd/ycy` only in the Integration Commit that establishes this state.
- **Release Accepted:** every applicable native OS/architecture, browser, protocol, persistence, process, installer, updater, embedded-payload, and Artifact Set gate also passes with evidence in the ledger.

Small leaves should reach Integrated in one reviewable commit. A large Module may use multiple **Preparation Slices** containing real internal implementation and focused tests while its command remains unregistered and externally unreachable. Each Preparation Slice must pass the then-applicable checks; committed red tests, half-registered commands, placeholders, temporary legacy dispatch, and broken intermediate commits are forbidden. The final Integration Commit adds the whole-binary tests and registers the command.

If an in-scope parity test or focused implementation probe fails, stop the serial integration lane at that Unit. Attach the reproducible failure to a narrow Wayfinder decision and agree on the integration or compatibility change before committing it. Independent research, probes, and unmerged Preparation Slices may continue, but later Migration Units may not be integrated out of order.

### Foundation and lower-risk lane

The Foundation Gate is the buildable atomic cutover selected by [Define the legacy archive and migration cutover choreography](15-define-archive-cutover.md). It freezes and verifies `legacy/bun/`, creates the root Go module, `cmd/ycy`, `cliapp`, global help/version/logging/error behavior, the three-entry Vite application and unconditional Go embed, root Make/pnpm/Lefthook/architecture gates, Release Identity injection, fixed six-name cross-build skeleton, and active/legacy isolation. It contains no `appconfig`, 7-Zip or FRP payload implementation, business command, placeholder, or active release workflow.

Migration builds report the explicit plain version `0.0.0-dev`. The first release candidate injects plain `0.1.0`; the first Go tag is `v0.1.0`.

After Foundation, integrate the low-risk `export env` vertical slice using the exact discovery, dotenv grammar, prompt, JSON, path, output, and exit tests in the core inventory. This is the first proof that a command-owned Module can pass typed input/results through the CLI boundary and run from the standalone embedded-Web binary without legacy dispatch.

Next implement `appconfig` as a non-command Migration Unit. It must directly read and preserve the current and legacy `~/.ycy-cli/config.json` shapes, machine identity, PBKDF2-SHA256/AES-256-GCM ciphertext, path precedence, locking, atomic publication, and unrelated Fork/CM/Tunnel fields. Native macOS, Linux, and Windows machine-ID, lock, and replacement evidence remains required. Then integrate configuration leaves one by one in read-before-write and increasing-risk order:

```text
config fork list -> add -> remove
config cm list -> add -> use -> set -> remove -> test
```

The parent command group may appear when its first real leaf is registered; absent siblings are not represented by placeholders. `config cm test` initially owns its provider transport.

Then integrate `rm`, `run`, `git heat`, `git pulse`, `zip`, `git fork`, and `git cm` in that exact order. Each leaf uses every parser, prompt, mutation, subprocess, output, cancellation, failure, and known-defect vector in its core or Git inventory. Destructive cases run only in disposable roots. `run` retains its current rejected passthrough grammar and its ability to invoke Bun in a user's project. `git heat` and `git pulse` precede ZIP as the Git basics used by the local-command lane; shared encrypted configuration is already ready before `git fork` and `git cm`, and `git cm` remains the final Git leaf.

At the available native milestones, record real filesystem, process, signal, external-Git, prompt, and platform-path evidence rather than postponing all such work to the final gate.

### Shared ownership

The only pre-evidenced shared product Modules are `appconfig`, `filesession`, and `webassets`. All other behavior starts with its first consumer. Extract it at a later real caller only when both callers share the same invariants and the resulting Interface removes meaningful complexity:

- `git heat` initially owns its external-Git process Adapter. At `git pulse`, extract only genuinely shared execution/cancellation/error behavior; parsing and command semantics remain leaf-owned.
- `config cm test` initially owns provider HTTP behavior. At `git cm`, extract a narrow provider transport only if the second implementation proves the same request, timeout, response, usage, and redaction invariants.
- Diff, FS, and Tunnel retain their own HTTP, SSE, MCP, WebSocket, route-priority, lifecycle, and protocol Adapters. Do not create a generic Web server Module.

Do not add global Git, HTTP, process, adapter, fixture, or `testutil` packages. Continue enforcing the dependency and naming rules from [Choose the Go module seams and project layout](13-choose-go-module-seams.md).

### Web Readiness Gate and Diff

The Web Readiness Gate is a checkpoint, not another common production Module. Foundation already owns Vite, `webassets`, and unconditional embedding. Immediately before Diff, complete the reusable production handler/browser harness and re-prove all three physical shells, their reachable hashed assets, workers, Monaco resources, MIME types, cache/CSP/security headers, deep links, reserved routes, and development proxies. Command-specific APIs remain absent from this Gate.

Implement Diff through passing Preparation Slices in this order:

```text
Comparison Workspace + filesystem/glob/gitignore/snapshot
-> queries + text/content/difference/blob
-> REST + SSE
-> MCP
-> CLI lifecycle + production React browser
-> standalone/native Integration Commit
```

REST, MCP, and React query the immutable Comparison Workspace and never construct local filesystem paths. Diff becomes Integrated only after its raw REST/SSE/MCP suites, native filesystem/signal checks, production-browser flows, host standalone smoke test, Complete Gate, and six cross-builds pass.

### FS

Implement FS through passing Preparation Slices in this order:

```text
fresh-Go filesession + root-confined workspace
-> listing/original/text/read-only HTTP
-> authentication/session/edit/management operations
-> direct and chunked upload
-> remote download
-> target 7-Zip runtime/inspection/extraction
-> research-approved thumbnail engine
-> tasks/SSE/production React browser
-> standalone/native Integration Commit
```

The FS Foundation proves the active Go session and workspace ownership used later by Tunnel. Bun-written FS sessions receive no compatibility detection, fixture, migration, deletion, or release gate. Go follows normal startup; failures produce useful, non-secret diagnostic logs for internal operator recovery. The thumbnail Slice implements the research-selected `gav1d`/`vpx`/standard-library graph and the same-binary persistent worker boundary; it adds no WASM fallback, helper artifact, system-codec lookup, or behavior redesign.

FS becomes Integrated only after the full raw HTTP/session/filesystem/upload/download/archive/thumbnail/task/SSE/browser contract passes from a standalone artifact. Its Release Accepted gates include native filesystem/process behavior, real target 7-Zip materialization and extraction, the selected thumbnail engine, and browser workflows on the required target matrix.

### Tunnel

Build the Tunnel Foundation before either public leaf:

```text
config.json remembered-connection Adapter
+ fresh-Go sessions and SQLite/domain schema
+ protocol-v3 wire types and platform mapping
+ pinned FRP manifest, typed TOML, acquisition and verification
+ state locks and file publication
+ owner-local OS process supervision
```

Then implement server persistence/domain transactions, followed by the server HTTP/SSE/React control plane, frps supervision, and protocol-v3 agent gateway. Use scripted Go protocol clients to complete the server's standalone/native Integration Commit and register `tunnel server` without waiting for the public client leaf.

Next implement `tunnel connect` resolution, remembered configuration, client cache, reconciler, frpc supervision, and protocol lifecycle. Complete real Go client to Go server protocol-v3 and pinned-FRP forwarding acceptance before its Integration Commit. Retain exact v3 fields, close/authentication meaning, `darwin|linux|win32` and `x64|arm64` wire values, and pinned FRP artifact identities for Go peers.

Bun-written Tunnel sessions, SQLite, client cache, generated runtime state, Bun-client/Go-server, and Go-client/Bun-server behavior receive no compatibility fixture, migration, or release gate. Only remembered Tunnel connections inside `config.json` retain Bun-written direct-read compatibility. Normal startup diagnostics and coordinated internal operation replace a mixed-version rollout guarantee.

### Upgrade and transition scope

The first Bun-to-Go switch is supported only through `scripts/install.sh`, `scripts/install.ps1`, or manual binary replacement. Bun `ycy upgrade` to the first Go binary and Bun Legacy Update State are explicitly unsupported and untested.

Retain the public Go `ycy upgrade` leaf for Go-to-later-Go replacement. Implement it through passing Preparation Slices in this order:

```text
release identity/API/artifact/checksum resolution
-> staged candidate download/hash/version self-check
-> Go-owned `<target>.go-update-state.json` and isolated hidden entry
-> Unix/macOS/Windows replacement, rollback and cleanup
-> installer fixture tests
-> local Go-to-Go standalone/native Integration Commit
-> final six-target install/upgrade rerun
```

The public command retains its observed first-release CLI and replacement behavior except for the explicit cross-runtime exclusion. It never parses Bun Legacy Update State. Installer tests cover fresh and replacement installation of Go artifacts; Upgrade tests begin with a running Go artifact and install a later Go candidate.

Outside `config.json`, Bun-written FS/Tunnel/updater state is not a compatibility target. Do not detect, migrate, delete, or maintain legacy fixtures for it. Normal Go startup and actionable, non-secret logging apply; internal operators handle residue. This explicit transition scope does not authorize unrelated command redesign or hardening.

### Final Artifact Gate and release

The Final Artifact Gate starts only when every Migration Unit is Integrated and all earlier milestone-native results are recorded in the Acceptance Ledger. From a clean checkout with no `web/dist`, dependency tree, cache, downloaded payload, binary, or prior artifact output:

1. Run bootstrap, produce one Vite build, verify its complete three-shell/shared-asset graph, and run the offline Complete Gate.
2. Inject Release Identity `0.1.0` and emit exactly the six public artifact names with `CGO_ENABLED=0`.
3. Inspect Mach-O, ELF, and PE format, CPU, Go build metadata, Release Identity, nonzero size, and absence of Bun/Node runtime dependencies.
4. Prove each binary contains all three Web shells and reachable assets, only its target 7-Zip 26.02 runtime plus license, the compiled pure-Go thumbnail engine with no codec helper/system lookup, and the exact FRP manifest but no FRP executable bytes; reproduce the pinned thumbnail dependencies' required notices and patent grants in release documentation.
5. Generate `SHA256SUMS`; reject missing, extra, duplicate, or mismatched files and prove installer/Upgrade basename selection for all six targets.
6. On matching native macOS, Linux, and Windows x64/arm64 hosts, run standalone help/version/parser/signal, command-specific native, production-browser, FS/7-Zip/thumbnail, Go-only Tunnel/FRP, installer, and Go-to-Go Upgrade acceptance.
7. Move every applicable ledger entry to Release Accepted and confirm generated Web output, dependencies, source maps, caches, downloaded payloads, binaries, checksum files, and release staging remain untracked.

Only after this local gate passes may the repository add a new Go release workflow. It is `workflow_dispatch` only, invokes the same Artifact Gate, and does not restore or call the archived Bun workflow. No tag or release exists during migration. After the workflow is ready, create the first Go tag `v0.1.0` and perform the first release manually.
