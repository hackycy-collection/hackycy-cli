# Go Migration Implementation Plan

## Source Decisions

- Route and evidence ownership: [migration map](map.md), [Acceptance Ledger](acceptance.md), and [approved serial roadmap](issues/16-approve-command-migration-roadmap.md).
- Toolchain and build research: [pure-Go baseline](issues/01-research-pure-go-toolchain.md), [mixed-project quality gates](issues/02-research-mixed-project-quality-gates.md), and [Vite MPA embedding research](issues/03-research-vite-go-embedding.md).
- Command inventories and their resolved index issues: [core inventory](inventories/core-command-contracts.md) / [issue 04](issues/04-inventory-core-command-contracts.md), [Git inventory](inventories/git-command-contracts.md) / [issue 05](issues/05-inventory-git-command-contracts.md), [Diff inventory](inventories/diff-contracts.md) / [issue 06](issues/06-inventory-diff-contracts.md), [FS inventory](inventories/fs-contracts.md) / [issue 07](issues/07-inventory-fs-contracts.md), [Tunnel inventory](inventories/tunnel-contracts.md) / [issue 08](issues/08-inventory-tunnel-contracts.md), and [Upgrade/artifact inventory](inventories/upgrade-artifact-contracts.md) / [issue 09](issues/09-inventory-upgrade-artifact-contracts.md).
- Proven implementation boundaries: [CLI compatibility decision](issues/10-prove-go-cli-compatibility.md) with its [probe guide](prototypes/go-cli-compat/README.md) and [results](prototypes/go-cli-compat/RESULTS.md); [unconditional Vite embedding decision](issues/11-prove-vite-go-embed-path.md) with its [interactive evidence](prototypes/vite-go-embed-preview.html).
- Architecture and cutover: [data-compatibility scope](issues/12-choose-data-compatibility-mechanisms.md), [Go module seams](issues/13-choose-go-module-seams.md), [hook policy](issues/14-choose-mixed-project-hook-policy.md), and [archive/cutover choreography](issues/15-define-archive-cutover.md).
- First-release policy and explicit deferrals: [parity and exception policy](issues/17-choose-corrected-core-command-contracts.md), [Git hardening disposition](issues/18-choose-safe-git-command-contracts.md), [Diff hardening disposition](issues/19-choose-safe-diff-service-contracts.md), [FS hardening disposition](issues/20-choose-safe-fs-service-contracts.md), [Tunnel hardening disposition](issues/22-choose-safe-rolling-tunnel-contracts.md), and [Upgrade hardening disposition](issues/23-choose-safe-self-update-contract.md).
- FS thumbnail decision and evidence: [resolved research issue](issues/21-research-cgo-free-fs-thumbnails.md) and [full compatibility report](research/21-cgo-free-fs-thumbnails.md).
- Repository execution context: [CLAUDE.md](../../CLAUDE.md).

## Outcome

The repository produces a behavior-compatible, standard, CGO-free Go `ycy` CLI from one Go composition root and one retained pnpm/Vite/React package. Every standalone binary contains the three verified Web applications and the target-specific runtime material fixed by the decisions; the local release-candidate path emits the six public `v0.1.0` artifacts and their checksum evidence without depending on Bun or dispatching into the frozen implementation.

## Non-Negotiable Rules

- The [serial lane and state model](issues/16-approve-command-migration-roadmap.md#state-and-commit-model) govern ordering. Research and unmerged preparation may proceed independently, but integration never skips a Gate.
- The [first-release parity policy](issues/17-choose-corrected-core-command-contracts.md#answer) governs observable behavior. A decision may change only after a focused, reproducible in-scope incompatibility; documented hardening findings do not authorize drift.
- Tests are derived command-by-command from the frozen source and inventories. The maintained suite never imports, generates from, dispatches into, or executes `legacy/bun/`, subject only to the documented `ycy run` behavior in a user's project.
- The [module ownership rules](issues/13-choose-go-module-seams.md#answer), [unconditional embed lifecycle](issues/11-prove-vite-go-embed-path.md#answer), and [root hook policy](issues/14-choose-mixed-project-hook-policy.md#answer) apply across every Gate.
- Every product build is frontend-first, uses the pinned Go toolchain with `CGO_ENABLED=0`, and preserves the fixed six-artifact vocabulary. Generated Web output, dependencies, caches, payload downloads, binaries, checksums, and staging output remain untracked.
- Bun-written direct-read compatibility is limited to `config.json`. Other Bun-written state, Bun-to-first-Go self-upgrade, and mixed Bun/Go Tunnel operation retain the exclusions indexed by [the map](map.md#out-of-scope).
- The Acceptance Ledger's `pending`, `integrated`, and `release-accepted` values describe migration evidence only. Goal execution state exists only in `goal-runbook.md`; neither state model substitutes for the other.
- No Gate pushes a remote branch, creates a release tag, publishes a release, or restores/replaces CI, Docker, or deployment automation. The [map's scope boundary](map.md#out-of-scope) remains authoritative.

### Common Unit Contract

- G0 through G26 unlock their successor when the matching Acceptance Ledger Unit meets the documented `integrated` rule. Native or artifact evidence obtained at a milestone is recorded immediately; evidence still outstanding for `release-accepted` remains explicit until G27.
- A public leaf stays absent through all Preparation Slices. Its focused tests, whole-binary tests, registration, and composition land together in one Integration Commit. Non-command foundations use the same one-Unit commit boundary.
- Each slice contains one behavior or one external-call cluster, begins with inventory-derived tests, and leaves the repository buildable. Large Modules follow their roadmap-listed Preparation Slice order.
- Gate-close verification always includes the Gate's focused Directed checks, `make check`, `make build`, one current-host standalone artifact smoke test, and all six `CGO_ENABLED=0` cross-builds. G0 first establishes the repository-owned cross-build entry; its recorded invocation is reused unchanged thereafter.
- Evidence is concise and secret-free. The Acceptance Ledger records Unit state, Integration Commit, host/target facts, commands/results, candidate SHA-256, native work still outstanding, and any compatibility-decision link. The Goal Progress Log references that durable evidence and records the slice result.
- A focused in-scope incompatibility, a missing required verifier, decision drift, or inability to preserve the preceding Gate stops the serial lane. The current Unit remains unintegrated until the linked narrow Wayfinder decision or missing evidence is supplied.
- Rollback is the smallest completed slice or the Unit's Integration Commit. Revert that boundary while keeping the public leaf absent; never compensate by weakening tests, adding legacy fallback, or integrating a later Unit.

## Gate Overview

| Gate | Name | Unlock condition | Outcome |
| --- | --- | --- | --- |
| G0 | Foundation Gate | Start | One verified atomic Bun archive/Go-era cutover commit |
| G1 | export env | G0 Exit all satisfied | First low-risk command vertical integrated |
| G2 | appconfig foundation | G1 Exit all satisfied | Shared `config.json` compatibility owner integrated |
| G3 | config fork list | G2 Exit all satisfied | Read-only Fork surface integrated |
| G4 | config fork add | G3 Exit all satisfied | Fork creation surface integrated |
| G5 | config fork remove | G4 Exit all satisfied | Fork removal surface integrated |
| G6 | config cm list | G5 Exit all satisfied | Read-only CM surface integrated |
| G7 | config cm add | G6 Exit all satisfied | CM creation surface integrated |
| G8 | config cm use | G7 Exit all satisfied | CM default-selection surface integrated |
| G9 | config cm set | G8 Exit all satisfied | CM field-update surface integrated |
| G10 | config cm remove | G9 Exit all satisfied | CM removal surface integrated |
| G11 | config cm test | G10 Exit all satisfied | CM provider-test surface integrated |
| G12 | rm | G11 Exit all satisfied | Legacy-compatible removal command integrated |
| G13 | run | G12 Exit all satisfied | Project package-manager runner integrated |
| G14 | git heat | G13 Exit all satisfied | First Git analysis leaf integrated |
| G15 | git pulse | G14 Exit all satisfied | Workspace Git activity leaf integrated |
| G16 | zip | G15 Exit all satisfied | Interactive archive command integrated |
| G17 | git fork | G16 Exit all satisfied | Repository acquisition leaf integrated |
| G18 | git cm | G17 Exit all satisfied | Final Git leaf integrated |
| G19 | Web Readiness Gate | G18 Exit all satisfied | Shared production Web verification checkpoint integrated |
| G20 | diff | G19 Exit all satisfied | Diff CLI, protocols, and React flow integrated |
| G21 | FS Foundation | G20 Exit all satisfied | Fresh-Go session and root workspace foundation integrated |
| G22 | fs | G21 Exit all satisfied | FS CLI, service, payloads, and React flow integrated |
| G23 | Tunnel Foundation | G22 Exit all satisfied | Go-only state, protocol, FRP, and process foundation integrated |
| G24 | tunnel server | G23 Exit all satisfied | Tunnel control plane integrated |
| G25 | tunnel connect | G24 Exit all satisfied | Go client and Go-to-Go forwarding integrated |
| G26 | upgrade | G25 Exit all satisfied | Go-to-Go self-update leaf integrated |
| G27 | Final Artifact Gate | G26 Exit all satisfied | Six-artifact local release candidate fully accepted |

## G0: Foundation Gate

### Purpose

Establish the single atomic cutover that every later Unit can extend without runtime or build dependence on the Bun implementation.

### Inputs

- [Foundation definition](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane) and [complete cutover choreography](issues/15-define-archive-cutover.md#answer).
- [CLI boundary](issues/10-prove-go-cli-compatibility.md#answer), [embed boundary](issues/11-prove-vite-go-embed-path.md#answer), [module layout](issues/13-choose-go-module-seams.md#answer), and [hook policy](issues/14-choose-mixed-project-hook-policy.md#answer).
- Frozen source commit `78358c0201b71891e36603d6abb8d7c87d54ad57` and the current repository tree, whose intervening tracked changes are planning material only.

### Objective

Create one Cutover Commit that satisfies the Foundation Gate contract and changes its Acceptance Ledger Unit to `integrated` with complete archive, build, hook, smoke, and six-target evidence.

### Scope boundary

Use only the archive/retain/rewrite/copy ledger in [cutover sections 1-5](issues/15-define-archive-cutover.md#1-capture-the-bun-baseline-before-moving-anything) and the Foundation contents in the [roadmap](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane). Anything that roadmap paragraph defers belongs to a successor Gate.

### Constraints

- The archive manifest, byte/mode identity, frontend copy provenance, nonempty Go/Vite foundation, and no-legacy dependency checks are one indivisible commit contract.
- Unknown/custom hook state is preserved and triggers the cutover's hard stop; only the selected exact legacy hook may be replaced.
- The initial CLI exposes only real foundation behavior. No business-command placeholder, legacy alias, fallback subprocess, or release workflow is allowed.

### Slice policy

Rehearse capture, archive verification, frontend extraction, foundation build, and hook lifecycle as independently verified worktree slices, but stage and record them as the single Cutover Commit required by the decision.

### Verification

#### Directed

- Execute every archive, copy-provenance, index cleanliness, isolation, hookctl disposable-repository, architecture, route/header, global CLI, host smoke, and six-target check indexed by [cutover section 6](issues/15-define-archive-cutover.md#6-prove-the-cutover-commit-before-recording-it).
- Prove migration identity `0.0.0-dev`, fixed artifact naming, unconditional three-shell embedding, and absence of Bun/Node runtime dependence from the built standalone file.

#### Repository

1. `make bootstrap`
2. `make hooks-doctor`
3. `make check`
4. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Archive and provenance evidence prove Exit 1; Directed CLI/Web/isolation checks prove Exit 2; Repository, standalone, and cross-build results prove Exit 3; the Integration Commit plus the Foundation row in [acceptance.md](acceptance.md#unit-ledger) prove Exit 4.

### Stop conditions

- The frozen source identity, complete archive manifest, or byte/mode mapping cannot be established.
- Hook inspection finds custom or unknown policy, or any required foundation verifier fails without a behavior-preserving fix.
- A focused Bun-to-Go compatibility mismatch meets the [exception threshold](issues/17-choose-corrected-core-command-contracts.md#answer).

### Rollback

Revert the whole Cutover Commit; do not retain a partial archive or partial foundation. Before removing the new hook lifecycle, use its selected uninstall path; never recreate a removed legacy hook implicitly.

### Exit conditions

1. The frozen archive, manifest, source identity, and four frontend copy mappings have reproducible proof.
2. The new tree satisfies CLI, layout, Web embedding, hook, architecture, and legacy-isolation contracts without a placeholder command.
3. All Directed and Repository checks, the current-host standalone smoke, and all six cross-builds succeed from a disposable clean checkout.
4. The Foundation Unit is `integrated` with its Cutover Commit and evidence recorded in the Acceptance Ledger.

## G1: export env

### Purpose

Prove the command Module, typed CLI binder, prompt/filesystem adapters, and standalone execution path with the roadmap's lowest-risk leaf.

### Inputs

- [Lower-risk lane](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`export env` contract and required tests](inventories/core-command-contracts.md#export-env-dir).
- Frozen `legacy/bun/` implementation as read-only test-writing reference.

### Objective

Meet the Acceptance Ledger `integrated` contract for `export env` and expose the leaf only in its Integration Commit.

### Scope boundary

Limit work to the command-owned Module, its CLI binding, prompt/filesystem ports, focused tests, and composition. Do not introduce shared configuration or another command.

### Constraints

- The [Common Unit Contract](#common-unit-contract) and inventory's parser, selection, output, filesystem, and exit behavior govern the port.
- Dotenv compatibility is implemented through a tested parser boundary, not ad hoc line splitting.

### Slice policy

Slice by discovery, dotenv parsing/merge, selection, output publication, then whole-command registration; one behavior per slice.

### Verification

#### Directed

- Run every vector in the inventory's `export env` Required Go tests after its owning slice and as one focused suite at Gate close.
- Run the Common Unit Contract standalone smoke and six-target verification at Gate close.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Focused tests prove Exit 1; registration and whole-binary probes prove Exit 2; the common verification bundle and Acceptance Ledger entry prove Exit 3.

### Stop conditions

- A focused parser, prompt, path, JSON, or exit case cannot reproduce the inventory under the migration constraints.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the smallest preparation slice or the `export env` Integration Commit; keep the leaf absent after rollback.

### Exit conditions

1. All inventory-derived `export env` behaviors have focused passing evidence.
2. The standalone CLI exposes the leaf without legacy dispatch and preserves the preexisting global surface.
3. The Unit is `integrated` with commit, build, smoke, cross-build, and outstanding-native evidence recorded.

## G2: appconfig foundation

### Purpose

Establish the sole semantic owner of the shared encrypted configuration before any configuration or dependent Git/Tunnel leaf.

### Inputs

- [Roadmap appconfig prerequisite](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [Shared persistence contract and tests](inventories/core-command-contracts.md#shared-configuration-persistence-contract).
- [Data scope](issues/12-choose-data-compatibility-mechanisms.md#answer), [module ownership](issues/13-choose-go-module-seams.md#answer), and [parity exception policy](issues/17-choose-corrected-core-command-contracts.md#answer).

### Objective

Integrate `appconfig` with direct-read and behavior-compatible update evidence for existing `config.json` while publishing no new command.

### Scope boundary

Own schema normalization, path selection, machine identity, crypto, locking, atomic publication, and semantic Fork/CM/Tunnel operations only. Provider HTTP behavior and public leaves remain outside this Gate.

### Constraints

- Callers never receive the mutable root document, keys, ciphertext internals, or lock ownership.
- Only `config.json` receives Bun-written compatibility; no other state adapter or migration may be introduced.

### Slice policy

Slice by schema/read normalization, deterministic crypto vectors, platform machine identity, lock ownership/recovery, publication, then semantic operations.

### Verification

#### Directed

- Run all Shared Configuration Required Go tests, including deterministic ciphertext, concurrent update, failure injection, and platform adapter suites.
- Record native machine-ID, lock, permission, and replacement work by target; run the Common Unit Contract standalone and cross-build checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Schema/crypto vectors prove Exit 1; concurrency/publication/platform tests prove Exit 2; architecture and no-command probes prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A same-machine/user Bun-written `config.json` cannot be read or preserved exactly enough to satisfy the decided contract.
- A required native locking/replacement behavior cannot be established, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the appconfig Integration Commit as one non-command seam; no successor may retain a private replacement config owner.

### Exit conditions

1. Current and legacy config shapes plus encrypted Fork/CM/Tunnel values have direct-read evidence.
2. Locking, concurrent semantic updates, atomic publication, cleanup, and platform behavior meet the inventory contract.
3. Architecture checks confirm `appconfig` is the sole owner and no public command was added.
4. The Unit is `integrated` with complete common evidence recorded.

## G3: config fork list

### Purpose

Integrate the first read-only consumer of appconfig and establish the real `config fork` group without placeholders.

### Inputs

- [Configuration order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config fork list` contract](inventories/core-command-contracts.md#config-fork-list) and [config leaf test boundary](inventories/core-command-contracts.md#dependencies-and-tests).

### Objective

Meet the `integrated` criteria for `config fork list` and register only that real leaf under its parent groups.

### Scope boundary

Implement listing, empty-state behavior, secret-safe projection, CLI binding, and tests. Add/remove mutations remain absent.

### Constraints

- Use appconfig's semantic read surface; do not decrypt or expose plaintext for presentation.
- Follow the Common Unit Contract and inventory's compatibility/presentation boundary.

### Slice policy

Slice by read projection, empty/nonempty rendering semantics, CLI binding, then whole-binary registration.

### Verification

#### Directed

- Run the inventory-derived empty, ordering, field-equivalence, ciphertext-preview, and plaintext non-disclosure cases.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Focused projection tests prove Exit 1; parser/whole-binary tests prove Exit 2; common evidence and the Unit row prove Exit 3.

### Stop conditions

- The leaf would require bypassing appconfig ownership or changing a machine-consumed field.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and remove the newly exposed `config fork` surface if it has no remaining real leaf.

### Exit conditions

1. Empty and populated list contracts, ordering, and secret non-disclosure have focused evidence.
2. Only `config fork list` is registered; absent siblings have no placeholder.
3. The Unit is `integrated` with common evidence recorded.

## G4: config fork add

### Purpose

Add the first Fork mutation through appconfig while preserving the frozen interaction and update behavior.

### Inputs

- [Configuration order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config fork add` contract](inventories/core-command-contracts.md#config-fork-add) and [Fork test boundary](inventories/core-command-contracts.md#dependencies-and-tests).

### Objective

Meet the `integrated` criteria for `config fork add` without changing the decided overwrite, validation, encryption, cancellation, or exit behavior.

### Scope boundary

Implement only add prompts, validation, encrypted semantic update, CLI binding, and tests. Removal and CM behavior are unchanged.

### Constraints

- Persist only through appconfig and preserve unrelated fields under concurrent updates.
- Inventory defects remain parity cases; the deferred hardening decision is not an implementation input.

### Slice policy

Slice by prompt/validation, typed mutation, cancellation/failure outcomes, then registration.

### Verification

#### Directed

- Run focused validation, prompt order/cancellation, overwrite, encryption, persisted-shape, concurrency, and failure cases from the Fork inventory.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Prompt/mutation tests prove Exit 1; config preservation tests prove Exit 2; common evidence and the Unit row prove Exit 3.

### Stop conditions

- A focused behavior cannot be reproduced without changing appconfig or the accepted first-release contract.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config fork add`; retain preceding read behavior.

### Exit conditions

1. Every add, overwrite, validation, cancellation, encryption, and failure path has focused evidence.
2. Concurrent updates preserve unrelated config fields and never expose plaintext.
3. The Unit is `integrated` with common evidence recorded.

## G5: config fork remove

### Purpose

Complete the Fork leaf set with behavior-compatible selection, confirmation, and concurrent removal.

### Inputs

- [Configuration order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config fork remove` contract](inventories/core-command-contracts.md#config-fork-remove) and [Fork test boundary](inventories/core-command-contracts.md#dependencies-and-tests).

### Objective

Meet the `integrated` criteria for `config fork remove` while retaining all prior Fork behavior.

### Scope boundary

Implement only removal selection/confirmation, semantic deletion, CLI binding, and focused tests.

### Constraints

- Empty, cancellation, negative confirmation, missing/concurrent state, and update failures retain their inventory meanings.
- Use appconfig's mutation operation; do not read-modify-write the root document in the command.

### Slice policy

Slice by selection/empty state, confirmation outcomes, semantic deletion/concurrency, then registration.

### Verification

#### Directed

- Run empty, selection, cancellation, negative/positive confirmation, concurrent update, and failure-path tests.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Focused interaction and mutation tests prove Exit 1; regression tests over all Fork leaves prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- Removal cannot preserve the accepted mutation/cancellation behavior through appconfig.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config fork remove`.

### Exit conditions

1. All removal branches and concurrent preservation behavior have focused evidence.
2. The complete Fork leaf set passes regression and standalone tests without placeholders or legacy dispatch.
3. The Unit is `integrated` with common evidence recorded.

## G6: config cm list

### Purpose

Introduce the read-only CM surface and profile projection on top of appconfig.

### Inputs

- [CM leaf order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config cm list` and shared resolution contract](inventories/core-command-contracts.md#config-cm-leaves).

### Objective

Meet the `integrated` criteria for `config cm list` and expose no unimplemented CM sibling.

### Scope boundary

Implement only list projection, default marking, empty state, CLI binding, and tests; provider resolution and mutations remain for later Gates.

### Constraints

- API keys remain inside appconfig and never appear in output or diagnostics.
- Preserve insertion/default semantics while treating layout details as presentation freedom.

### Slice policy

Slice by projection/default semantics, empty/nonempty presentation, CLI binding, then registration.

### Verification

#### Directed

- Run focused empty, insertion-order, default-marker, field-equivalence, and secret non-disclosure tests.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Projection and non-disclosure tests prove Exit 1; whole-binary registration tests prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- The leaf would require exposing storage or crypto details outside appconfig.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and remove the newly exposed `config cm` surface if it has no remaining real leaf.

### Exit conditions

1. Empty/populated CM list and default projection behavior have focused evidence without key disclosure.
2. Only the real list leaf is exposed under the CM group.
3. The Unit is `integrated` with common evidence recorded.

## G7: config cm add

### Purpose

Add CM profile creation through the established appconfig owner.

### Inputs

- [CM leaf order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config cm add` and shared profile contracts](inventories/core-command-contracts.md#config-cm-add).

### Objective

Meet the `integrated` criteria for `config cm add` without changing the accepted validation, overwrite, encryption, default-selection, cancellation, or exit behavior.

### Scope boundary

Implement add prompts, normalization, encrypted semantic update, CLI binding, and focused tests only.

### Constraints

- Persist through appconfig and preserve all unrelated config data.
- First-release behavior follows the inventory; later overwrite/URL hardening remains outside this Gate.

### Slice policy

Slice by prompt/validation, normalization/encryption, default/overwrite mutation, failure outcomes, then registration.

### Verification

#### Directed

- Run focused name/base URL/model/key, cancellation, overwrite, first-default, encryption, concurrent preservation, and failure tests.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Prompt and persisted-shape tests prove Exit 1; appconfig/concurrency regression proves Exit 2; common evidence proves Exit 3.

### Stop conditions

- The accepted mutation cannot be reproduced through appconfig without policy change.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config cm add`.

### Exit conditions

1. Add, overwrite, default, validation, cancellation, encryption, and failure behavior have focused evidence.
2. Appconfig ownership and unrelated-field preservation remain intact.
3. The Unit is `integrated` with common evidence recorded.

## G8: config cm use

### Purpose

Integrate explicit CM default-profile selection.

### Inputs

- [CM leaf order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config cm use` contract](inventories/core-command-contracts.md#config-cm-use-profile).

### Objective

Meet the `integrated` criteria for `config cm use` while preserving missing-profile and successful-selection behavior.

### Scope boundary

Implement only typed profile selection, semantic default update, CLI binding, and focused tests.

### Constraints

- Profile existence and result mapping follow the inventory.
- No provider request or profile-field mutation is added.

### Slice policy

Slice by lookup/result behavior, semantic update, CLI binding, then registration.

### Verification

#### Directed

- Run existing/missing profile, default transition, persistence, unrelated-field, parser, output, and exit tests.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Focused selection tests prove Exit 1; all prior CM/Fork config regression proves Exit 2; common evidence proves Exit 3.

### Stop conditions

- Selection semantics cannot be preserved through appconfig, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config cm use`.

### Exit conditions

1. Existing and missing profile selection paths have focused evidence.
2. Prior config behavior and shared-file preservation remain verified.
3. The Unit is `integrated` with common evidence recorded.

## G9: config cm set

### Purpose

Integrate profile-field updates with their command-specific parsing and encryption behavior.

### Inputs

- [CM leaf order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config cm set` contract](inventories/core-command-contracts.md#config-cm-set-profile-key-value).
- [First-release parity policy](issues/17-choose-corrected-core-command-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `config cm set` across every supported key and accepted parser edge.

### Scope boundary

Implement only set-key dispatch, validation/parsing, API-key encryption, semantic update, CLI binding, and tests.

### Constraints

- Supported keys and legacy parser behavior come only from the inventory.
- No post-parity numeric, URL, or empty-model correction is introduced.

### Slice policy

Slice by string fields, encrypted key, numeric fields, error outcomes, then registration.

### Verification

#### Directed

- Run the full key/value table, boundary and permissive-parser vectors, missing profile/key, encryption, persistence, and exit tests.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Key/parser tests prove Exit 1; encryption/persistence and regression tests prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- A focused parser or persisted-value case cannot be reproduced under the selected CLI/appconfig boundaries.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config cm set`.

### Exit conditions

1. Every supported key, boundary, legacy parser edge, and failure has focused evidence.
2. Secret storage and all prior configuration behavior remain verified.
3. The Unit is `integrated` with common evidence recorded.

## G10: config cm remove

### Purpose

Integrate confirmed CM profile removal and default reassignment.

### Inputs

- [CM leaf order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config cm remove` contract](inventories/core-command-contracts.md#config-cm-remove-profile).

### Objective

Meet the `integrated` criteria for `config cm remove` across cancellation, missing, default, and last-profile cases.

### Scope boundary

Implement only confirmation, semantic deletion/default update, CLI binding, and focused tests.

### Constraints

- Preserve confirmation and exit meanings; do not add a new recovery or overwrite policy.
- All writes use appconfig and retain unrelated state.

### Slice policy

Slice by confirmation/cancellation, deletion/default transition, failure outcomes, then registration.

### Verification

#### Directed

- Run cancellation/no, missing, nondefault/default/last removal, persistence, concurrency, output, and exit tests.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Focused branch tests prove Exit 1; config regression proves Exit 2; common evidence proves Exit 3.

### Stop conditions

- Removal/default semantics cannot be preserved through appconfig, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config cm remove`.

### Exit conditions

1. Every removal, confirmation, cancellation, missing, and default-transition path has focused evidence.
2. The prior config surface and unrelated data remain verified.
3. The Unit is `integrated` with common evidence recorded.

## G11: config cm test

### Purpose

Complete the CM block with the first OpenAI-compatible provider call and its diagnostic boundary.

### Inputs

- [CM leaf order and ownership](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`config cm test` plus profile-resolution/provider contracts](inventories/core-command-contracts.md#config-cm-test-profile).

### Objective

Meet the `integrated` criteria for `config cm test` and establish its provider transport as the first-consumer implementation.

### Scope boundary

Implement profile resolution needed by this leaf, its command-owned provider transport, request/response mapping, CLI binding, and tests. Do not extract a shared provider Module before G18 proves matching invariants.

### Constraints

- Precedence, request fields, DeepSeek condition, timeout, usage, error summary, and secret boundaries follow the inventory.
- Tests use local/injected transports and never contact a real provider.

### Slice policy

Slice by profile resolution, request construction, success decoding, timeout/HTTP/JSON failures, diagnostics, then registration.

### Verification

#### Directed

- Run the full precedence matrix and provider URL/header/body, DeepSeek, timeout, HTTP, JSON, empty-content, usage, and secret non-disclosure cases.
- Run regression over every config leaf plus the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Resolution/request/response tests prove Exit 1; local-transport and non-disclosure tests prove Exit 2; full config regression and common evidence prove Exit 3.

### Stop conditions

- A provider or profile behavior cannot be reproduced without a product-policy change.
- Verification would require real credentials/network access, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister only `config cm test`; leave any command-owned provider implementation unreachable or revert its preparation slice.

### Exit conditions

1. Profile resolution and every provider request/result/error path have focused local evidence.
2. Secrets are absent from output, logs, fixtures, and recorded evidence.
3. The complete configuration block and common Gate-close verification succeed; the Unit is `integrated`.

## G12: rm

### Purpose

Port the destructive local removal workflow deliberately from its observed contract before adding further subprocess-heavy commands.

### Inputs

- [Lower-risk command order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`rm` inventory](inventories/core-command-contracts.md#rm-paths) and [parity disposition](issues/17-choose-corrected-core-command-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `rm` without silently applying the deferred hardening policy.

### Scope boundary

Implement explicit and smart planning, prompts, deletion execution, results, CLI binding, and focused tests only.

### Constraints

- Every destructive test runs in a validated disposable root; no test targets the repository, user home, cwd ancestor, or real external data.
- Legacy safety and status defects remain isolated parity vectors; later policy decisions are not implemented here.

### Slice policy

Slice by explicit planning, smart discovery/actions, prompt semantics, deletion/result mapping, cancellation/signals, then registration.

### Verification

#### Directed

- Run all explicit/smart mode, depth, path-kind, hidden/VCS, force, cancellation, missing, partial-failure, and accepted defect vectors in disposable roots.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Disposable-root focused tests prove Exit 1; safety-of-test-harness and process cleanup prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- A test cannot be contained in a disposable root, or the Go implementation cannot reproduce an in-scope observable case.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister `rm`; remove only staging created by this Gate's disposable tests.

### Exit conditions

1. All inventory-derived explicit/smart, prompt, deletion, failure, and cancellation paths have focused evidence.
2. Tests prove containment and leave no unintended filesystem mutation or child work.
3. The Unit is `integrated` with common evidence recorded.

## G13: run

### Purpose

Port project package-manager discovery, interactive script selection, and child outcome propagation while retaining the command's observed argv grammar.

### Inputs

- [Lower-risk command order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`run` inventory](inventories/core-command-contracts.md#run-path) and [CLI prototype result](prototypes/go-cli-compat/RESULTS.md#observed-behavior).

### Objective

Meet the `integrated` criteria for `run`, including external Bun selection in user projects and the frozen passthrough rejection matrix.

### Scope boundary

Implement project/script/package-manager discovery, prompts, child-process ownership, legacy CLI binding, result mapping, and focused tests. Do not add the future delimiter feature.

### Constraints

- Invoking Bun selected by a user's project is permitted; the ycy repository/build/runtime must still have no Bun dependency or legacy dispatch.
- The composition root remains the only exit owner; child code/signal results are typed.

### Slice policy

Slice by project/script discovery, manager ordering, prompt behavior, CLI rejection table, child launch/stdio, then exit/signal integration and registration.

### Verification

#### Directed

- Run path/package/script/lockfile, prompt/cancellation, exact argv/cwd/stdio, missing executable, child exit/signal, external Bun, and passthrough-rejection tests.
- Prove no orphan process on supported host paths; run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Discovery/parser tests prove Exit 1; child process and signal tests prove Exit 2; no-Bun architecture checks and common evidence prove Exit 3.

### Stop conditions

- Cobra binding cannot represent a focused legacy argv case, or native child ownership cannot preserve an in-scope result.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit, unregister `run`, and terminate/reap only children created by the Gate's test harness.

### Exit conditions

1. All project discovery, selection, manager-order, argv, and cancellation contracts have focused evidence.
2. Child status/signals propagate through the composition root with no orphan, and user-project Bun remains the only Bun allowance.
3. The Unit is `integrated` with common evidence recorded.

## G14: git heat

### Purpose

Integrate the lowest-risk Git leaf and establish command-owned external-Git execution.

### Inputs

- [Git lane ordering](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`git heat` contract and tests](inventories/git-command-contracts.md#git-heat).

### Objective

Meet the `integrated` criteria for `git heat` using the user's Git executable and preserving the inventory's aggregation, parser, output, and failure behavior.

### Scope boundary

Implement only heat options, Git discovery/log adapter, parsing/aggregation, report semantics, CLI binding, and tests. Keep the Git process adapter owned here until G15 proves a shared invariant.

### Constraints

- Tests use disposable repositories and never mutate the developer repository or contact a remote.
- External Git behavior remains authoritative; do not replace it with a repository library by assumption.

### Slice policy

Slice by option/range parsing, Git invocation/path parsing, aggregation/sorting/time, report/error/signal behavior, then registration.

### Verification

#### Directed

- Run the inventory's option, fixed-clock, repository-kind, status/rename/copy, arbitrary-path, report, failure, and cancellation cases.
- Run native filename/process cases available on the host and the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Git fixture tests prove Exit 1; path/process/report tests prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- A focused Git output or path case cannot preserve observable behavior with stable plumbing.
- Tests would touch a non-disposable repository, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister `git heat`; leave no test repository process running.

### Exit conditions

1. Parser, Git invocation, aggregation, sorting/time, report, and failure contracts have focused evidence.
2. Disposable/native tests cover arbitrary paths and child cleanup without real repository mutation.
3. The Unit is `integrated` with common evidence recorded.

## G15: git pulse

### Purpose

Integrate workspace repository discovery and activity reporting, extracting only genuinely shared Git execution behavior.

### Inputs

- [Git lane ordering and extraction rule](issues/16-approve-command-migration-roadmap.md#shared-ownership).
- [`git pulse` contract and tests](inventories/git-command-contracts.md#git-pulse-directory).

### Objective

Meet the `integrated` criteria for `git pulse` with deterministic tested results and the accepted partial-failure behavior.

### Scope boundary

Implement pulse traversal, date selection, bounded Git-log execution, author selection, report/result behavior, CLI binding, and tests. Extract from heat only execution/cancellation/error invariants proven identical.

### Constraints

- Repository parsing and product semantics remain leaf-owned.
- All Git repositories are disposable and all clocks/network/processes are controlled in tests.

### Slice policy

Slice by traversal/repository identity, day boundaries, Git concurrency/parsing, author selection/report, cancellation/cleanup, then registration.

### Verification

#### Directed

- Run missing/path/repository-kind, excluded/symlink, legacy day/fixed-clock, Git argv/failure/concurrency, author prompt, TTY, ordering, and cancellation tests.
- Re-run heat regression after any extraction; run the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Traversal/time/Git tests prove Exit 1; author/report/cancellation and heat regression prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- A proposed shared seam changes either leaf's invariants, or a focused repository/partial-failure case cannot be reproduced.
- Any Common Unit Contract stop condition occurs.

### Rollback

Revert the pulse Integration Commit; revert a shared extraction with it if heat cannot remain independently verified.

### Exit conditions

1. Discovery, time, Git execution, author selection, partial-failure, and report contracts have focused evidence.
2. Any shared Git execution seam is justified by both real callers and heat regression remains green.
3. The Unit is `integrated` with common evidence recorded.

## G16: zip

### Purpose

Integrate the interactive project discovery, archive planning, ZIP generation, and reveal workflow.

### Inputs

- [Lower-risk command order](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`zip` inventory](inventories/core-command-contracts.md#zip-directory) and [parity policy](issues/17-choose-corrected-core-command-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `zip` while preserving selected entries, paths, metadata loss, publication, interaction, and observed result meanings.

### Scope boundary

Implement discovery/planning, prompts, archive execution, reveal adapter, CLI binding, and focused tests. Do not introduce deferred path, atomicity, resource, or error-status hardening.

### Constraints

- Filesystem and archive tests use disposable roots and inspect archive bytes structurally.
- A structured workspace parser may replace heuristics only when accepted inputs remain covered by parity tests.

### Slice policy

Slice by workspace/project discovery, candidate scoring/naming, prompt/glob plan, archive content/publication, reveal/result behavior, then registration.

### Verification

#### Directed

- Run all workspace/project/candidate/confidence, naming, prompt, glob/path, dot/symlink, content/metadata, failure/status, reveal, and large-input vectors.
- Run the Common Unit Contract standalone and six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Planner tests prove Exit 1; structural archive/publication/result tests prove Exit 2; common evidence proves Exit 3.

### Stop conditions

- A selected Go archive implementation cannot reproduce a focused observable archive case.
- A test cannot be safely contained, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister `zip`; remove only disposable archive outputs created by tests.

### Exit conditions

1. Discovery, recommendation, naming, prompt, glob, and planner behavior have focused evidence.
2. Archive paths/bytes/metadata/publication plus all result and reveal branches have structural evidence.
3. The Unit is `integrated` with common evidence recorded.

## G17: git fork

### Purpose

Integrate repository acquisition after appconfig and basic Git execution are proven.

### Inputs

- [Git lane ordering](issues/16-approve-command-migration-roadmap.md#foundation-and-lower-risk-lane).
- [`git fork` contract and tests](inventories/git-command-contracts.md#git-fork-repo-dest).
- [Deferred Git hardening disposition](issues/18-choose-safe-git-command-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `git fork` across configured providers, archive acquisition, clone fallback, destination effects, and accepted failure behavior.

### Scope boundary

Implement input/config resolution, provider adapters, archive handling, clone fallback, destination orchestration, CLI binding, and tests. No broader Git or provider framework is introduced.

### Constraints

- All providers are local fixture servers; all Git remotes and destinations are disposable. Evidence never records credentials.
- Reproduce the first-release behavior selected by the decisions; do not silently substitute the deferred transactional or credential redesign.

### Slice policy

Slice by input/config resolution, each provider API, archive parse/extract, clone fallback, destination/result orchestration, then registration.

### Verification

#### Directed

- Run every input/config, provider, redirect/failure, TAR shape, destination, clone argv/content, credential non-recording, cancellation, and native path vector required by the inventory.
- Re-run appconfig and shared Git regressions; run the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Input/provider tests prove Exit 1; archive/clone/destination tests prove Exit 2; secret-free regression and common evidence prove Exit 3.

### Stop conditions

- A focused accepted archive, clone, credential, or destination behavior cannot be reproduced under Go/target constraints.
- Verification would contact a real provider or mutate a non-disposable path, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Integration Commit and unregister `git fork`; clean only Gate-owned disposable destinations and fixture servers.

### Exit conditions

1. Input, appconfig, provider, archive, and fallback behavior have focused local evidence.
2. Destination effects, failure/cancellation outcomes, native path cases, and credential handling are verified in disposable environments.
3. The Unit is `integrated` with common evidence recorded.

## G18: git cm

### Purpose

Complete the Git block with evidence capture, provider generation, staged mutation, commit, and optional push behavior.

### Inputs

- [Git lane ordering and provider extraction rule](issues/16-approve-command-migration-roadmap.md#shared-ownership).
- [`git cm` contract and tests](inventories/git-command-contracts.md#git-cm).
- [CLI optional-value proof](prototypes/go-cli-compat/RESULTS.md#observed-behavior) and [deferred Git hardening](issues/18-choose-safe-git-command-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `git cm` across the full legacy flag matrix, evidence boundary, provider call, index mutation, commit, hooks, and push outcomes.

### Scope boundary

Implement the deep command Module, its Git/provider/prompt/process ports, CLI binder, and tests. Extract provider transport from G11 only if both real callers prove identical transport/redaction invariants.

### Constraints

- Tests use disposable repositories, local provider servers, controlled hooks/remotes, and sentinel secrets; they never push externally.
- Preserve observed first-release provider and mutation boundaries. Deferred containment/redaction/transaction corrections remain outside this Gate.

### Slice policy

Slice by flag binder, repository/status snapshot, evidence classification/budget, provider/output validation, staging/confirmation, commit/hooks, push/result, then registration.

### Verification

#### Directed

- Run all flag combinations, arbitrary Git states/paths, evidence/hash/budget, provider request/output, staging/index, hook/commit, push, TTY, signal, and native vectors in the inventory.
- Assert exact local provider bodies against sentinel fixtures, re-run config/Git regressions, and run the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Binder/repository/evidence tests prove Exit 1; provider/secret tests prove Exit 2; mutation/commit/push tests prove Exit 3; regression and common evidence prove Exit 4.

### Stop conditions

- Cobra cannot reproduce a focused optional-value case, or a provider/index/commit behavior fails the exception threshold.
- Verification would expose a secret, invoke a real provider, or push a real remote; any Common Unit Contract stop condition also stops the Gate.

### Rollback

Revert the Integration Commit and unregister `git cm`; revert any premature shared provider extraction that cannot leave G11 independently verified.

### Exit conditions

1. The complete CLI, repository-state, evidence-selection, and output-validation contracts have focused evidence.
2. Local provider requests and diagnostics reproduce the decided boundary without leaking test or real secrets into recorded evidence.
3. Staging, cancellation, snapshot recheck, hooks, commit, push, and partial-result behavior are verified only in disposable repositories/remotes.
4. The Unit is `integrated` with common evidence recorded.

## G19: Web Readiness Gate

### Purpose

Re-prove the shared production Web asset and browser harness immediately before the first Web-owning command, without creating a generic server Module.

### Inputs

- [Web Readiness definition](issues/16-approve-command-migration-roadmap.md#web-readiness-gate-and-diff).
- [Unconditional embed and routing contract](issues/11-prove-vite-go-embed-path.md#answer).
- Web boundaries in the [Diff](inventories/diff-contracts.md#vitego-asset-boundary), [FS](inventories/fs-contracts.md#vite-boundary), and [Tunnel](inventories/tunnel-contracts.md#static-react-application) inventories.

### Objective

Meet the `integrated` criteria for the Web Readiness checkpoint with one reusable production handler/browser harness and complete three-entry asset/routing evidence.

### Scope boundary

Modify only Web build verification, `webassets`, production handler/browser test harnesses, development proxies, and entry-owned frontend adaptations needed for readiness. Command-specific APIs, MCP, file routes, WebSockets, and business behavior remain absent.

### Constraints

- This Gate is a checkpoint, not a new generic HTTP production Module.
- Each command adapter retains its own method, route priority, CSP, and fallback semantics; filenames under the hashed asset tree are not contracts.

### Slice policy

Slice by structured asset graph, common serving headers/MIME, each entry's route matrix, workers/Monaco/file assets, development proxy mode, then production browser harness.

### Verification

#### Directed

- Prove all three shells, reachable hashed assets, workers/Monaco resources, MIME/cache/CSP/security headers, deep links, reserved namespaces, and strict development proxies from production output.
- Inject missing/stale asset failures and reproduce the compiled routing matrix; run the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Asset-verifier failure/success evidence proves Exit 1; raw handler routing/header tests prove Exit 2; production-browser/development-proxy tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A required asset/worker cannot be represented in one Vite graph, or a command's compiled routing behavior cannot be preserved.
- A placeholder API is required to make the harness pass, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the checkpoint Integration Commit; retain the G0 embed foundation but remove readiness-only harness/adaptations that cannot pass independently.

### Exit conditions

1. Structured verification fails closed on missing/stale output and succeeds on one current three-entry graph.
2. Raw production handlers reproduce each command's shell/asset/method/header matrix without cross-entry fallback.
3. Production browser and development proxy harnesses exercise real workers/assets without command-specific placeholder APIs.
4. The Web Readiness Unit is `integrated` with common evidence recorded.

## G20: diff

### Purpose

Port the high-risk Comparison Workspace, HTTP/SSE/MCP adapters, CLI lifecycle, and retained React application behind one deep command Module.

### Inputs

- [Diff Preparation Slice order](issues/16-approve-command-migration-roadmap.md#web-readiness-gate-and-diff).
- [Complete Diff inventory](inventories/diff-contracts.md) and [deferred policy decision](issues/19-choose-safe-diff-service-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `diff` from focused domain/protocol/browser suites through a standalone native-host command.

### Scope boundary

Follow only the roadmap-listed Diff slices and inventory contract. No generalized Web server, new persistent state, trust redesign, resource policy, or cross-command filesystem API is added.

### Constraints

- HTTP, MCP, and React query the immutable Comparison Workspace and never construct local filesystem paths.
- Glob and gitignore grammars, MCP SDK wrapping, content stability, and browser diff rendering remain separately owned boundaries.

### Slice policy

Use the exact roadmap order: workspace/filesystem/glob/gitignore/snapshot; query/text/difference/blob; REST/SSE; MCP; CLI/production browser; standalone/native Integration Commit.

### Verification

#### Directed

- Run every suite in [Required Go and frontend tests](inventories/diff-contracts.md#required-go-and-frontend-tests) at its owning slice and the complete raw REST/SSE/MCP/browser/native-host suite at Gate close.
- Exercise production assets from the standalone artifact, then run the Common Unit Contract six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Workspace/query/content suites prove Exit 1; raw REST/SSE/MCP suites prove Exit 2; CLI/browser/native-host artifact tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A focused filesystem, protocol, SDK, line-diff, gitignore, browser-worker, or target behavior meets the compatibility-exception threshold.
- Required raw-protocol, browser, or native-host verification cannot run; any Common Unit Contract stop condition also applies.

### Rollback

Revert the current passing Preparation Slice, or revert the final Integration Commit and unregister `diff`; preserve the preceding Web Readiness checkpoint.

### Exit conditions

1. Comparison Workspace, traversal/snapshot, query, content/blob, and text-difference contracts have focused evidence.
2. Raw REST, SSE, and MCP contracts pass without exposing arbitrary filesystem paths or relying on framework defaults.
3. CLI lifecycle, signals, production React workflows/workers, and native-host standalone behavior pass.
4. The Unit is `integrated` with common evidence and applicable native work recorded.

## G21: FS Foundation

### Purpose

Establish the fresh-Go persistent session owner and root-confined workspace that the FS command requires and Tunnel may later reuse only where invariants match.

### Inputs

- [FS roadmap boundary](issues/16-approve-command-migration-roadmap.md#fs).
- [FS domain/workspace contract](inventories/fs-contracts.md#domain-contract), [Workspace Path contract](inventories/fs-contracts.md#workspace-path-and-filesystem-contract), and [fresh-Go session contract](inventories/fs-contracts.md#authentication-and-persistent-session-contract).
- [Non-config state exclusion](issues/12-choose-data-compatibility-mechanisms.md#answer).

### Objective

Meet the `integrated` criteria for FS Foundation with Go-owned session restart behavior and root-confined filesystem operations, while exposing no `fs` command.

### Scope boundary

Implement the Go-created file-session owner/lock/record lifecycle and the root-confined Workspace Path/handle boundary needed by later FS slices. HTTP routes, management workflows, 7-Zip, thumbnails, tasks, and React integration remain outside.

### Constraints

- Do not detect, read, migrate, delete, or fixture Bun-written FS session state.
- Absolute paths never cross the workspace Interface; platform-specific containment stays owner-local.

### Slice policy

Slice by Workspace Path grammar, root identity/opened-handle containment, listing/basic file identity, Go-owned session lock/key/record lifecycle, then native restart/permission tests.

### Verification

#### Directed

- Run the applicable Workspace Path, symlink/reparse/root replacement, Go-owned session resume/refresh/revoke/prune/lock, concurrency, permission, and native-host tests from the FS inventory.
- Assert that no Bun-state fixture or public command exists; run the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Workspace/path tests prove Exit 1; Go-owned session/native tests prove Exit 2; architecture/no-command tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- Opened-handle containment or Go-owned session ownership cannot be proven on the current host and no faithful target adapter can be specified.
- Work would require Bun-state carryover, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the FS Foundation Integration Commit; no later command may retain a private duplicate of the removed owners.

### Exit conditions

1. Workspace Path, root identity, containment, symlink/reparse, and stable file operations have focused evidence.
2. Fresh-Go session creation, restart, refresh, revocation, pruning, locking, and native permission behavior have focused evidence without Bun fixtures.
3. Ownership/architecture checks pass and no `fs` public leaf is registered.
4. The FS Foundation Unit is `integrated` with common evidence recorded.

## G22: fs

### Purpose

Port the complete FS CLI, HTTP/session/filesystem workflows, bundled 7-Zip path, selected thumbnail engine, task streams, and retained React application.

### Inputs

- [FS Preparation Slice order](issues/16-approve-command-migration-roadmap.md#fs).
- [Complete FS inventory](inventories/fs-contracts.md), [deferred FS policy](issues/20-choose-safe-fs-service-contracts.md#answer), and [thumbnail decision/report](research/21-cgo-free-fs-thumbnails.md#decision).
- [Upgrade/artifact payload boundary](inventories/upgrade-artifact-contracts.md#standalone-build-contents).

### Objective

Meet the `integrated` criteria for `fs` from raw HTTP and filesystem behavior through production browser, 7-Zip, thumbnail, and native-host standalone evidence.

### Scope boundary

Follow only the roadmap-listed FS slices. Bun-written state carryover and deferred HTML/trust, DNS, containment-policy, resource, exit, SSE, and shutdown redesigns remain outside.

### Constraints

- HTTP and queues pass Workspace Paths into the root-confined owner; React never receives absolute paths.
- The selected pure-Go thumbnail graph and self-exec worker are mandatory; no codec helper, WASM fallback, system lookup, or extra shipped file is allowed.
- 7-Zip remains the target-specific verified runtime payload fixed by the decisions; FRP is not part of this Gate.

### Slice policy

Use the exact roadmap order: read-only HTTP; auth/session/edit/operations; uploads; remote download; target 7-Zip; selected thumbnails; tasks/SSE/production React; standalone/native Integration Commit.

### Verification

#### Directed

- Run every suite in [Required Go and frontend tests](inventories/fs-contracts.md#required-go-and-frontend-tests) and the report's [thumbnail implementation acceptance](research/21-cgo-free-fs-thumbnails.md#implementation-acceptance) at its owning slice.
- At Gate close, run raw HTTP/session/filesystem/upload/download/archive/thumbnail/task/SSE/browser suites from the standalone artifact, then the Common Unit Contract six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Filesystem/HTTP/session/operation suites prove Exit 1; upload/download/7-Zip/thumbnail/task suites prove Exit 2; production browser/native-host artifact tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A focused filesystem, HTTP, payload, codec, worker, browser, or native-target mismatch meets the compatibility-exception threshold.
- Thumbnail pins/sums/notices or verified target 7-Zip inputs are unavailable; any Common Unit Contract stop condition also applies.

### Rollback

Revert the current passing Preparation Slice, or revert the final Integration Commit and unregister `fs`; preserve FS Foundation and remove only Gate-owned generated/runtime test material.

### Exit conditions

1. CLI, raw HTTP, authentication/session, original/text, operations, upload, download, and task contracts have focused evidence.
2. Target 7-Zip and selected thumbnail behavior, worker timeout/replacement, pins, sums, and notice inputs have focused artifact evidence.
3. SSE, production React/worker flows, lifecycle/signals, and native-host standalone behavior pass without Bun-state carryover.
4. The Unit is `integrated` with common evidence and applicable native/payload work recorded.

## G23: Tunnel Foundation

### Purpose

Establish the Go-only state, protocol-v3, pinned FRP, publication, and native process foundations shared by the two Tunnel leaves.

### Inputs

- [Tunnel Foundation definition](issues/16-approve-command-migration-roadmap.md#tunnel).
- [Complete Tunnel inventory](inventories/tunnel-contracts.md), [Go-only scope decision](issues/22-choose-safe-rolling-tunnel-contracts.md#answer), and [remembered config scope](issues/12-choose-data-compatibility-mechanisms.md#answer).

### Objective

Meet the `integrated` criteria for Tunnel Foundation without exposing either public Tunnel leaf.

### Scope boundary

Implement remembered-connection appconfig adapter, fresh-Go session/SQLite/domain primitives, protocol-v3 types/platform mapping, pinned FRP manifest/acquisition/typed TOML, locks/publication, and owner-local process supervision. Server HTTP/domain workflows and client reconciliation remain later Gates.

### Constraints

- Bun-written direct read applies only to remembered Tunnel connections in `config.json`; every other Tunnel store is Go-created.
- Only Go-to-Go protocol v3 is a first-release target. Mixed runtime peers and Bun-written non-config stores have no fixtures or gates.
- FRP executable bytes remain runtime-acquired, digest-pinned material and are never embedded in ycy.

### Slice policy

Slice by remembered config identity, Go-owned sessions/SQLite schema primitives, v3 wire/platform types, FRP manifest/TOML, acquisition/publication, locks, then each OS process supervisor.

### Verification

#### Directed

- Run applicable remembered-config, Go-owned session/SQLite restart, v3 frame/platform, manifest/archive/hash/version, TOML, lock/publication, and native process tests from [required Tunnel coverage](inventories/tunnel-contracts.md#missing-or-weak-coverage-to-add).
- Assert no mixed-runtime/non-config fixtures and no public leaf; run the Common Unit Contract standalone/six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Config/state tests prove Exit 1; protocol/FRP/TOML tests prove Exit 2; lock/process/native and architecture tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- Remembered `config.json`, Go-created SQLite, protocol v3, pinned FRP, or a required native process behavior fails a focused probe.
- Work would add mixed-runtime or Bun-state compatibility, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the Tunnel Foundation Integration Commit; no public leaf or duplicate private foundation may remain.

### Exit conditions

1. Remembered-connection direct read and fresh-Go session/SQLite restart contracts have focused evidence without other Bun-state fixtures.
2. Exact protocol-v3 types/platform mapping, FRP manifest/acquisition/verification, and typed TOML have focused evidence.
3. Locks, file publication, and owner-local native process supervision pass applicable host tests; neither Tunnel leaf is registered.
4. The Tunnel Foundation Unit is `integrated` with common evidence recorded.

## G24: tunnel server

### Purpose

Port the Tunnel control plane, durable domain transactions, browser API/SSE/React flow, frps supervision, and protocol-v3 agent gateway.

### Inputs

- [Tunnel server order](issues/16-approve-command-migration-roadmap.md#tunnel).
- [`tunnel server`, domain, browser, protocol, and supervision contracts](inventories/tunnel-contracts.md).
- [Go-only first-release scope](issues/22-choose-safe-rolling-tunnel-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `tunnel server` using scripted Go protocol clients and a standalone native-host control plane.

### Scope boundary

Implement server persistence/domain transactions, accounts/clients/tunnels/import, HTTP/SSE/React, frps control, and the v3 agent gateway. The public connect client and mixed-runtime support remain outside.

### Constraints

- SQLite transactions remain desired-state truth; the agent gateway is not a second persistence layer.
- Only Go-created server stores are tested. Browser/API behavior and first-release defects follow the inventory/deferred-policy boundary.

### Slice policy

Slice by SQLite domain transactions; accounts/sessions/ownership; clients/tunnels/import; HTTP/SSE/React; frps supervision; agent authorization/v3 gateway; scripted-client standalone/native Integration Commit.

### Verification

#### Directed

- Run the server-applicable cases in [required Tunnel coverage](inventories/tunnel-contracts.md#missing-or-weak-coverage-to-add), including Go-created DB/WAL, auth/ownership, raw API/SSE/browser, v3 scripted clients, frps, process, and native-host behavior.
- Run production Vite namespace/route tests and the Common Unit Contract six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Domain/account/session tests prove Exit 1; raw HTTP/SSE/browser tests prove Exit 2; agent/frps/scripted-client/native tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A focused Go-created SQLite, browser API, protocol-v3, FRP, or native supervisor behavior meets the exception threshold.
- Verification would require a Bun peer or Bun-written non-config state, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the current passing Preparation Slice, or revert the Integration Commit and unregister `tunnel server`; preserve Tunnel Foundation.

### Exit conditions

1. Go-created persistence/domain/account/ownership/revision transactions have focused restart and rollback evidence.
2. Raw HTTP, session, SSE, retained React, route/header, and frps-control contracts pass from production assets.
3. The v3 agent gateway passes scripted Go-client, lifecycle, FRP, and applicable native-host standalone tests without mixed-runtime evidence.
4. The Unit is `integrated` with common evidence recorded.

## G25: tunnel connect

### Purpose

Complete Tunnel with remembered resolution, instance ownership, authenticated v3 client, reconciler, frpc supervision, and real Go-to-Go forwarding.

### Inputs

- [Tunnel connect order](issues/16-approve-command-migration-roadmap.md#tunnel).
- [`tunnel connect`, remembered identity, reconciliation, protocol, and FRP contracts](inventories/tunnel-contracts.md).
- [Go-only first-release scope](issues/22-choose-safe-rolling-tunnel-contracts.md#answer).

### Objective

Meet the `integrated` criteria for `tunnel connect`, including real Go client-to-Go server protocol-v3 and pinned-FRP forwarding evidence.

### Scope boundary

Implement resolution/selection/remembering, Go-created instance cache/lock, WebSocket lifecycle, desired-state reconciliation, frpc supervision, CLI binding, and tests. No mixed-runtime peer or Bun instance-state path is added.

### Constraints

- Authentication always precedes cold activation; cached state never authorizes startup.
- Wire fields, platform vocabulary, artifact pins, revision/rollback meaning, and reconnect behavior follow the inventory.

### Slice policy

Slice by field-aware resolution/remembering, instance identity/lock/cache, probe/WebSocket welcome, reconciler transaction/rollback, frpc supervision/reconnect, then real Go-to-Go/FRP Integration Commit.

### Verification

#### Directed

- Run client-applicable config/selection/TTY, instance, v3 frame/artifact, reconciliation crash-point, signal/reconnect, FRP process, and native-host cases from the Tunnel required coverage.
- Run real local Go server/client HTTP/TCP/UDP forwarding with pinned FRP, server regression, and the Common Unit Contract six-target checks.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Resolution/instance tests prove Exit 1; protocol/reconciliation/process tests prove Exit 2; real Go-to-Go forwarding and native-host tests prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A focused remembered-config, protocol-v3, reconciliation, pinned-FRP, or native process case meets the exception threshold.
- Verification would require mixed Bun/Go peers or Bun-written instance state, or a Common Unit Contract stop condition occurs.

### Rollback

Revert the current passing Preparation Slice, or revert the Integration Commit and unregister `tunnel connect`; preserve the server and Tunnel Foundation.

### Exit conditions

1. Resolution, remembered-pair persistence, instance identity/lock/cache, and TTY/error behavior have focused evidence.
2. Authentication, v3 welcome/snapshot, reconciliation verify/publish/rollback, reconnect, and frpc supervision have focused evidence.
3. Real Go client-to-Go server HTTP/TCP/UDP forwarding and applicable native-host standalone behavior pass with exact pinned FRP fields.
4. The Unit is `integrated` with common evidence recorded.

## G26: upgrade

### Purpose

Port the public Go-to-Go Upgrade surface, Go-owned update transaction files, hidden apply path, native replacement, installer fixtures, and rollback behavior.

### Inputs

- [Upgrade slice order and transition scope](issues/16-approve-command-migration-roadmap.md#upgrade-and-transition-scope).
- [Upgrade/artifact inventory](inventories/upgrade-artifact-contracts.md), [CLI hidden-entry boundary](issues/10-prove-go-cli-compatibility.md#answer), and [deferred updater redesign](issues/23-choose-safe-self-update-contract.md#answer).

### Objective

Meet the `integrated` criteria for `upgrade` starting from a Go artifact and replacing it with a later verified Go artifact on the current native host.

### Scope boundary

Implement release/artifact/checksum resolution, candidate verification, Go-owned state/hidden entry, OS-specific replacement/rollback/cleanup, installer fixtures, CLI binding, and tests. Never inspect Bun Legacy Update State or support Bun `ycy upgrade` to Go.

### Constraints

- Public version output remains plain and isolated from status reporting/self-checks.
- Network tests use local transports/fixture servers; filesystem/process tests use temporary user/install roots and never touch the developer installation.
- First-Go installation is installer/manual only; the Upgrade leaf is Go-to-later-Go.

### Slice policy

Use the exact roadmap order: release/artifact resolution; candidate verification; Go-owned state/hidden entry; Unix/macOS/Windows replacement/rollback/cleanup; installer fixtures; local Go-to-Go standalone/native Integration Commit.

### Verification

#### Directed

- Run [required Go unit/integration](inventories/upgrade-artifact-contracts.md#required-go-unit-and-integration-tests) and [installer tests](inventories/upgrade-artifact-contracts.md#required-installer-tests) with local fixtures and Go-created state only.
- Run current-host detached Go-to-Go replacement/rollback/self-check and the Common Unit Contract six-target checks; record outstanding target-native replacement gates.

#### Repository

1. `make check`
2. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Release/checksum/candidate tests prove Exit 1; Go-owned transaction and native replacement tests prove Exit 2; installer fixtures and Bun-state non-access assertions prove Exit 3; common evidence proves Exit 4.

### Stop conditions

- A focused Go-to-Go replacement, installer, artifact mapping, checksum, or required native-target behavior meets the exception threshold.
- Verification would touch a real installation/contact GitHub or read Bun Legacy Update State; any Common Unit Contract stop condition also applies.

### Rollback

Revert the current passing Preparation Slice, or revert the Integration Commit and unregister `upgrade`; test transactions restore their temporary prior target through the verified rollback path.

### Exit conditions

1. Public CLI/version, release/artifact/checksum, download, candidate hash/version, and observed result mappings have focused local evidence.
2. Go-owned state/hidden entry and current-host replacement, rollback, cleanup, and self-check behavior pass without opening the Bun state namespace.
3. Installer fixture tests cover first-Go and Go replacement at temporary stable paths without requiring Bun/Node/Go on the simulated user host.
4. The Unit is `integrated` with common evidence and outstanding native-target work recorded.

## G27: Final Artifact Gate

### Purpose

Aggregate every Unit and native milestone into one clean-checkout, six-target local release-candidate proof.

### Inputs

- [Final Artifact Gate roadmap](issues/16-approve-command-migration-roadmap.md#final-artifact-gate-and-release) and [Acceptance checklist](acceptance.md#final-artifact-gate).
- [Artifact Set tests](inventories/upgrade-artifact-contracts.md#required-artifact-set-tests), [native artifact tests](inventories/upgrade-artifact-contracts.md#required-native-artifact-tests), and all prior Unit evidence in [acceptance.md](acceptance.md).

### Objective

From a clean checkout, produce and fully verify exactly six `0.1.0` standalone artifacts plus `SHA256SUMS`, move every applicable Unit/target to `release-accepted`, and record the migration as locally release-ready.

### Scope boundary

This Gate may add/finalize the repository-owned local Artifact Set verifier, release-candidate assembly inputs, checksum generation, artifact inspection, and third-party release notices required by the decisions. It does not add a release workflow, tag, GitHub release, Docker/deployment replacement, or perform publication.

### Constraints

- Start with no generated Web output, installed dependencies, caches, downloaded payloads, binaries, checksums, or prior staging; use the exact Acceptance checklist without waiver.
- Build one verified three-shell Web graph, inject plain `0.1.0`, use `CGO_ENABLED=0`, and emit only the six fixed public basenames.
- Native execution evidence must come from matching target hosts/artifacts. Cross-compilation or cross-format inspection never substitutes for it.

### Slice policy

Slice by clean bootstrap/Web graph, complete repository checks, six-target assembly, format/metadata inspection, embedded payload inspection, checksum/parser verification, then per-target native CLI/browser/FS/Tunnel/installer/Upgrade evidence and final ledger reconciliation.

### Verification

#### Directed

- Execute the complete [Acceptance checklist](acceptance.md#final-artifact-gate), [Artifact Set tests](inventories/upgrade-artifact-contracts.md#required-artifact-set-tests), and [native artifact tests](inventories/upgrade-artifact-contracts.md#required-native-artifact-tests) with no skipped or waived item.
- Recompute every SHA-256, inspect Mach-O/ELF/PE and Go metadata, inspect embedded Web/7-Zip/FRP/thumbnail/notices, and prove generated/acquired output remains untracked.

#### Repository

1. `make bootstrap`
2. `make check`
3. `make build`

#### Manual acceptance

- 无。

### Evidence rule

Clean build/repository results prove Exit 1; artifact/metadata/payload/checksum inspection proves Exit 2; matching native-host suites prove Exit 3; fully reconciled Unit and target rows in the Acceptance Ledger prove Exit 4; scope inspection proves Exit 5.

### Stop conditions

- Any Unit is not `integrated`, any applicable native/artifact evidence is missing, any checksum/payload/identity check fails, or any generated output is tracked.
- A target cannot execute its required native suite, a test is skipped/waived, or an in-scope incompatibility requires a Wayfinder decision.
- Completing the Gate would require workflow/tag/release, CI, Docker, deployment, or another map-excluded change.

### Rollback

Discard only the failed local release-candidate staging/output and revert the smallest Artifact Gate slice. Preserve all prior integrated Units and their evidence; never alter a public contract or weaken a gate to salvage candidate bytes.

### Exit conditions

1. Clean bootstrap, one current verified Web graph, the Complete Gate, and the canonical product build succeed without tracked-output mutation.
2. Exactly six fixed `0.1.0` CGO-free artifacts and a correct exact-entry `SHA256SUMS` pass format, CPU, Go metadata, runtime-dependency, embedded Web, target 7-Zip, FRP-manifest, thumbnail, and notice inspection.
3. Every required matching native CLI/browser/FS/7-Zip/thumbnail/Go-only-Tunnel/FRP/installer/Go-to-Go-Upgrade suite succeeds for its target artifact.
4. Every applicable Unit and native target is `release-accepted`, with candidate SHA-256 and durable evidence recorded in the Acceptance Ledger.
5. Generated/acquired outputs remain untracked, and no workflow, tag, release, Docker, or deployment change was introduced by this effort.

## Definition Of Done

- G0 through G27 each satisfy every numbered Exit condition in strict order, and the Goal Runbook contains the corresponding evidence transitions.
- The Acceptance Ledger contains an Integration Commit and focused/build/smoke/cross-build evidence for every Migration Unit, plus `release-accepted` native/artifact evidence wherever applicable.
- The first-release behavior, direct-read `config.json` boundary, Go-only Tunnel v3 boundary, Go-to-Go Upgrade boundary, module ownership, unconditional Web embed, and six-artifact contracts remain traceable to the Source Decisions.
- The final local Artifact Set and `SHA256SUMS` satisfy G27, while all generated/dependency/cache/payload/binary/staging material remains outside Git.
- The repository is locally ready for separately scoped manual-dispatch release automation and a future `v0.1.0` release; neither is performed by this plan.

## Explicitly Out Of Scope

- Everything indexed by the [map's Out of scope section](map.md#out-of-scope), including actual publication and CI/Docker/deployment replacement.
- The corrected safety/resource/trust/transaction policies explicitly deferred by [issues 18](issues/18-choose-safe-git-command-contracts.md#answer), [19](issues/19-choose-safe-diff-service-contracts.md#answer), [20](issues/20-choose-safe-fs-service-contracts.md#answer), [22](issues/22-choose-safe-rolling-tunnel-contracts.md#answer), and [23](issues/23-choose-safe-self-update-contract.md#answer).
- Compatibility, migration, detection, deletion, fixtures, or recovery for Bun-written state outside `config.json`; Bun-to-first-Go Upgrade; and mixed Bun/Go Tunnel peers, as fixed by [issue 12](issues/12-choose-data-compatibility-mechanisms.md#answer) and the [roadmap transition scope](issues/16-approve-command-migration-roadmap.md#upgrade-and-transition-scope).
- New CLI features, post-parity behavior corrections, speculative abstractions, a second frontend/product module, a generic server/process/provider layer, or a maintained legacy oracle/golden corpus.
