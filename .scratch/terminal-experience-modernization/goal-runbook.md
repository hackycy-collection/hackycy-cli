# Terminal Experience Modernization Goal Runbook

## Source Baseline

| Path | SHA-256 |
| --- | --- |
| `.scratch/terminal-experience-modernization/implementation-plan.md` | `4b8a817f5c45e9932415f5407024c7bb00d93c8fb5ad2249fdf096825338d1f5` |
| `.scratch/terminal-experience-modernization/inventories/03-current-command-terminal-surfaces.md` | `30c1758c8b72077737aa85c5682a865cde78654fceadf1939b8e1f6d51098ae7` |
| `.scratch/terminal-experience-modernization/issues/01-research-latest-charm-v2-stack.md` | `3466e6493a6c77c1fb05b2563d330747239625a29571ada628299b0d6b1ada17` |
| `.scratch/terminal-experience-modernization/issues/02-research-exemplary-cli-terminal-experiences.md` | `27f768379595480928c5ab09c96188cfe6c4cdc73f0366fa3b83477a03981319` |
| `.scratch/terminal-experience-modernization/issues/03-inventory-every-command-terminal-surface.md` | `f9172a9c8485522e0d7f29586f183c76f664032e878f31c3b038932b140bba57` |
| `.scratch/terminal-experience-modernization/issues/04-choose-charm-v2-and-logging-boundaries.md` | `c8a1d5ee1eeea729230e1b3b4dc6cf505dbe5c3076d6226d619599e83de27136` |
| `.scratch/terminal-experience-modernization/issues/05-prototype-vivid-terminal-visual-system.md` | `adfdf81af89332ba66f04bb71f6c7555152afb182fcdf9529e623032908a92d2` |
| `.scratch/terminal-experience-modernization/issues/06-choose-shared-terminal-experience-contract.md` | `34166d3fff0e6015fa97028fe3d175fbc80c858fb71498c1d0c472416189fc71` |
| `.scratch/terminal-experience-modernization/issues/07-choose-root-help-and-error-presentation.md` | `e28081d4337a6986e078762dc94b1f3faeb45d2687e03e7ee74fc62cd91e0356` |
| `.scratch/terminal-experience-modernization/issues/08-choose-export-env-presentation.md` | `20596a71b236eaabdeecda48321b87ad7900628b8456bb6a9b43e54d248d9f4e` |
| `.scratch/terminal-experience-modernization/issues/09-choose-config-fork-list-presentation.md` | `516d45d5a80fb0a1a733ccbaee3fd95fa6a0b21916441a181ed20294b79ac088` |
| `.scratch/terminal-experience-modernization/issues/10-choose-config-fork-add-presentation.md` | `7f5039d12cc2377f366ab2ee551a023eafa1126db1876cdd690ac6813bf06620` |
| `.scratch/terminal-experience-modernization/issues/11-choose-config-fork-remove-presentation.md` | `d02eb7411d8562f9702d9de6446966f0846652bc03679cced381608635298903` |
| `.scratch/terminal-experience-modernization/issues/12-choose-config-cm-list-presentation.md` | `5b17fc8d17043ddf9908af4c38322a2d3d742d31f7fe528559461731bc10811a` |
| `.scratch/terminal-experience-modernization/issues/13-choose-config-cm-add-presentation.md` | `b85343d999c75325a58f1d0e481bea9eac5688caf855559c0308a84dbb2c4395` |
| `.scratch/terminal-experience-modernization/issues/14-choose-config-cm-use-presentation.md` | `9c92c7d7c612b1d15fce2170f88ebfc1f4be3e4954f6e4db31523c51fd088fdb` |
| `.scratch/terminal-experience-modernization/issues/15-choose-config-cm-set-presentation.md` | `fadd58c91e2da00a734cf0a6ca7a23689f2aac544e5d6ff2fcfdad0bba6ac346` |
| `.scratch/terminal-experience-modernization/issues/16-choose-config-cm-remove-presentation.md` | `8c6f25fbdc961e5b79e8cf667144edb48dd78fee07f81bcedf5e014faf5053d6` |
| `.scratch/terminal-experience-modernization/issues/17-choose-config-cm-test-presentation.md` | `1ea070f25683553f651f317d87c62a1cec36b51461e0df6c9456f933ecf2cf5f` |
| `.scratch/terminal-experience-modernization/issues/18-choose-diff-lifecycle-log-presentation.md` | `f931611f45ae8b88592a3c12473c7ba6bac5093cf4d24a95356b6041c9a84e07` |
| `.scratch/terminal-experience-modernization/issues/19-choose-fs-lifecycle-log-presentation.md` | `6860f2f9c22909a533297af0a46e25c00e74d9c776c1c9fbd9609f1332ff6a8e` |
| `.scratch/terminal-experience-modernization/issues/20-choose-git-heat-presentation.md` | `b66cac29c9724362ca5e089fcc4b3fd8d795044d148d9ea1ca5a99e8fe36597b` |
| `.scratch/terminal-experience-modernization/issues/21-choose-git-pulse-presentation.md` | `74ad403f2e1d55d35849b7b5db9e697cac7d59cab99da8759a689b18de611db0` |
| `.scratch/terminal-experience-modernization/issues/22-choose-git-fork-presentation.md` | `b10ceb31876d8b9bdd4b4207a4d2e155b8f1e91a54892b0434265ec3043faa5d` |
| `.scratch/terminal-experience-modernization/issues/23-choose-git-cm-presentation.md` | `a2e9f8a81a2020dfea451f39c75a65f2fab1a63d8602d5687bd7166cc4333ee6` |
| `.scratch/terminal-experience-modernization/issues/24-choose-rm-presentation.md` | `1b656baf42bfc31de8c99878d923d90a5a21618732dd5f0e0821e5107c86caab` |
| `.scratch/terminal-experience-modernization/issues/25-choose-run-handoff-presentation.md` | `59f38d0c098ae525e3c8260edff7aa2414f91625536f8749ee02942530dc1e1a` |
| `.scratch/terminal-experience-modernization/issues/26-choose-tunnel-connect-lifecycle-log-presentation.md` | `e69927f9fc012de0e4ec96e7a606821d69ee2cf188d6c50449a2ce70a095ab8a` |
| `.scratch/terminal-experience-modernization/issues/27-choose-tunnel-server-lifecycle-log-presentation.md` | `89279bc77d21c8ceb78a234fbb438b0576358ac70c253b87b846b4ba90c63cf5` |
| `.scratch/terminal-experience-modernization/issues/28-choose-zip-presentation.md` | `e5ec30e94d8abc9050d43ab968f77cf4f01567f631c4d6664d25a58c0e45b560` |
| `.scratch/terminal-experience-modernization/issues/29-choose-upgrade-presentation.md` | `867722b759822bad286e9ab988e13160ae15cc5131beb78480a001da1ec57e4c` |
| `.scratch/terminal-experience-modernization/issues/30-approve-rollout-and-acceptance-plan.md` | `73b9ee80a04f158f8eaeecb8b23e8d7839f2f68b145568ec8e96fb95cd76203d` |
| `.scratch/terminal-experience-modernization/map.md` | `65b036d823e06f51b097769a46548d9a5f83938ce13e6f089ae0b09503cccd9d` |
| `.scratch/terminal-experience-modernization/research/01-latest-charm-v2-stack.md` | `51fc0ad8d29e43cbc1529988a4f297c64010053e942fa22bac36633fba73cc92` |
| `.scratch/terminal-experience-modernization/research/02-exemplary-cli-terminal-experiences.md` | `d6d921dfdec393ccb20a62134adb52d286955cc7bb21506562960ba65474fd3b` |
| `CLAUDE.md` | `35a0f66a6531e61b35dc7288884da35eeef79ad1039f5ffdfd29295cee8d566e` |
| `CONTEXT.md` | `a3400217b0e5dfaf39dbb729c2a0fb6ae38d6fd3e1eab24d38e4a76bb55c6da0` |
| `docs/platform-boundaries.md` | `be531d09b0d6bc619d54a66d6c3fb6105208f241806dfd26b70631c18f87e473` |
| `docs/project-layout.md` | `26252b0db008b5964ae2877f7cbbf1472873f14d3a6165bc123a7f925a73c5f9` |
| `internal/terminal/prototype-vivid/README.md` | `d9089026734c466ab8c1c8cf341cc621fb8348a5344571082c685c3c020a7d5e` |
| `internal/terminal/prototype-vivid/main.go` | `a2455b2620d39e2bd71fe29dda08d69017159279e03e9786e7086c425162e65d` |
| `internal/terminal/prototype-vivid/view.go` | `9347ec260a6586dccfecda7d45dacd5c2b45a489aeed5316a9650b24e1384863` |

## State Rules

- `implementation-plan.md` is the only source of Gate contracts; this runbook
  records state and evidence only.
- Execute only the unique `active` Gate in the Goal Ledger at one time.
- Append each slice, change, verification result, risk, and next action to the
  corresponding Gate Progress Log.
- `passed` requires explicit evidence for every Exit condition in the plan;
  ordinary implementation or verification failure keeps the Gate `active`.
- `blocked` is reserved for a plan-declared Stop condition and must include the
  blocker and recovery condition.
- Pending manual acceptance is a one-time Goal termination handoff recorded in
  Progress Log; it is not `blocked`, and the Gate remains `active`.
- When a Gate passes, activate only its direct successor and end that Goal. The
  final Gate passing records effort completion and ends the Goal.

## Current Decision State

- 2026-09-02: the current Rich finite-command contract is B / `OPS CONSOLE`:
  command/status bar, safe metadata row, aligned `STATE / PHASE / DETAIL`
  table, active form/result region, amber primary, cyan accent, symbol-paired
  states, Huh bottom focus rule, no persistent left rail, and narrow
  single-column degradation.
- Variant A / Signal Rail remains only as historical comparison material. The
  existing A PTY acceptance entries below are preserved as historical evidence
  and are superseded for current visual-contract purposes; G0/G1/G2 statuses
  remain unchanged.
- Service Commands retain neutral, line-oriented Lifecycle Logs and are not
  converted to full-screen Ops Console views.
- G3 command slices remain gated on a fresh B visual re-acceptance. No command
  slice may be treated as B-accepted until that entry has an explicit result.

## Goal Ledger

| Gate | Status | Depends on | Plan contract | Unlock evidence |
| --- | --- | --- | --- | --- |
| G0: Baseline and dependency lock | passed | none | `implementation-plan.md` -> `G0: Baseline and dependency lock` | no predecessor |
| G1: Shared semantic terminal foundation | passed | G0 | `implementation-plan.md` -> `G1: Shared semantic terminal foundation` | G0 passed |
| G2: Logging and root boundary | passed | G1 | `implementation-plan.md` -> `G2: Logging and root boundary` | G1 passed |
| G3: Low-risk finite command slices | active | G2 + B visual re-acceptance | `implementation-plan.md` -> `G3: Low-risk finite command slices` | G2 passed; B Ops Console visual re-acceptance pending |
| G4: External-read and service-startup slices | planned | G3 | `implementation-plan.md` -> `G4: External-read and service-startup slices` | G3 pending |
| G5: Mutating, destructive, and archive slices | planned | G4 | `implementation-plan.md` -> `G5: Mutating, destructive, and archive slices` | G4 pending |
| G6: Git mutation and process handoff | planned | G5 | `implementation-plan.md` -> `G6: Git mutation and process handoff` | G5 pending |
| G7: Service lifecycles and detached updater | planned | G6 | `implementation-plan.md` -> `G7: Service lifecycles and detached updater` | G6 pending |
| G8: Release candidate and final acceptance | planned | G7 | `implementation-plan.md` -> `G8: Release candidate and final acceptance` | G7 pending |

## Progress Log

### Decision revision: B Ops Console

- 2026-09-02: revised the current Rich visual contract from variant A /
  Signal Rail to variant B / `OPS CONSOLE`. Updated issue 05, the decision
  map, the explicitly dependent command issues, the implementation plan, and
  this runbook. Variant A remains historical comparison material; variant C /
  Focus Flow remains an available prototype comparison. Service Commands stay
  neutral, line-oriented Lifecycle Logs and do not become full-screen Consoles.
- 2026-09-02: recorded the new prototype source baseline for
  `internal/terminal/prototype-vivid/README.md` (`d9089026734c466ab8c1c8cf341cc621fb8348a5344571082c685c3c020a7d5e`)
  and `main.go`
  (`a2455b2620d39e2bd71fe29dda08d69017159279e03e9786e7086c425162e65d`).
  The compact B renderer also has a source baseline for `view.go`
  (`9347ec260a6586dccfecda7d45dacd5c2b45a489aeed5316a9650b24e1384863`).
  The rebuilt tracked binary hash is
  `b968872ad830eec403cf66090695d8a9d480d335a733a13569741720ba6518a3`;
  it is progress evidence only, not a source baseline.
- 2026-09-02: B PTY re-acceptance entry opened from the repository root:
  `make prototype-terminal` (default success), plus
  `make prototype-terminal PROTOTYPE_ARGS='--variant=console --outcome=failure'`,
  `make prototype-terminal PROTOTYPE_ARGS='--variant=console --outcome=cancel'`,
  `NO_COLOR=1 make prototype-terminal`, `env -u NO_COLOR TERM=xterm-256color
  make prototype-terminal`, and a 55x15 PTY resize. Automated
  smoke evidence observed the B command/status bar, safe metadata row,
  `STATE / PHASE / DETAIL` table, amber/cyan-capable hierarchy, symbol-paired
  active/done/failed/cancelled states, bottom Huh focus rule, no persistent
  rail, narrow single-column degradation, AltScreen restoration before
  Transcript replay, transcript ordering, and `[redacted]` token output.
  The color-capable run (`env -u NO_COLOR TERM=xterm-256color make
  prototype-terminal`) emitted the expected amber/cyan ANSI palette; the
  `NO_COLOR=1` run retained the same symbols and labels without color.
  Success, failure, and configured cancellation journeys completed with exit
  code 0; failure/cancellation emitted no stdout result, while success emitted
  `profile github-work configured` only after the Transcript.
  F2/F3 switched between B and A; `--variant=a|signal` and
  `--variant=c|focus` remained usable. No B visual acceptance is claimed yet:
  the explicit human response is still required before G3 command slices can
  proceed. The historical A handoff and acceptance entries below remain
  unchanged.
- 2026-09-02: verification for the B revision passed. The prototype module
  `go test ./...`, prototype `go vet ./...`, targeted
  `go test -count=1 ./internal/terminal ./internal/terminaltest`, repository
  `make check`, and `git diff --check` all passed. The source-baseline table
  below was rechecked against all 42 listed SHA-256 values with zero
  mismatches. No command arguments, results, exits, APIs, interaction
  semantics, side effects, or redaction behavior were changed.
- 2026-09-02: tightened the B narrow-terminal renderer so the single-column
  degradation keeps the `STATE / PHASE / DETAIL` heading, every form/phase
  state row, and the active content region visible in order. Rebuilt the
  tracked prototype binary and rechecked the default `make prototype-terminal`
  success journey plus console failure/cancellation, `NO_COLOR=1`, and 55x15
  PTY smoke paths. The compact renderer source is recorded as `view.go`
  (`9347ec260a6586dccfecda7d45dacd5c2b45a489aeed5316a9650b24e1384863`); the
  rebuilt binary is `b968872ad830eec403cf66090695d8a9d480d335a733a13569741720ba6518a3`.
  Automated smoke passed with preserved symbols, redaction, stream/result
  behavior, and exit code 0. This remains mechanism evidence only; explicit
  human B visual acceptance is still pending and G3 remains gated.

### G0: Baseline and dependency lock

- 2026-09-01: initialized as `active`; implementation has not started.
- 2026-09-01: completed slice `baseline capture` (root/help and all 11
  inventory command-family entries). Added `baselines/README.md`,
  `baselines/manifest.json`, and the command-surface/baseline test evidence
  under `evidence/`. The frozen command surface remains 35 paths with manifest
  SHA-256 `885373a338983bb3dd71810f9ec668bf7273587c0a17de5e1ecddf86a6f11a6c`.
  Directed evidence records streams, exits/signals, side effects, state files,
  permissions, process boundaries, and redaction; the representative package
  test matrix passed. Risk: probe output is semantic and tokenized, so it does
  not assert raw dynamic bytes; the cited package and acceptance tests remain
  the executable contract. Next: replace the v1 Charm requirements/imports
  with the approved v2 module boundary as an independent dependency slice.
- 2026-09-01: completed slice `dependency/import lock`. Replaced the active
  Charm v1 requirements and `internal/terminal` imports with the approved five
  v2 modules, migrated the necessary Bubble Tea v2 view/key/terminal-mode and
  pure Lip Gloss style APIs, and added architecture checks for the exact module
  graph and direct Charm/Log imports from `pkg/cmd/**`. Updated the long-list
  PTY assertion to interpret Bubble Tea v2's cursor-position differential
  update rather than a contiguous stale-frame byte string; no command runner,
  durable result, or command-package behavior changed. Directed verification
  passed: `go mod verify`, exact `go list -m all` output, architecture test,
  terminal package test, and the PTY regression twice; details are in
  `evidence/g0-dependency-import-directed.log`. Risk: the PTY check verifies
  semantic terminal state from v2's expected differential output rather than a
  raw full-frame snapshot. Next: run G0 Repository verification in its declared
  order and record final Exit-condition evidence.
- 2026-09-01: all G0 Exit conditions passed; Gate marked `passed`.
  1. Baselines exist for root/help and every inventory family in
     `baselines/README.md` and `baselines/manifest.json`; the frozen surface
     remains 35 paths at SHA-256
     `885373a338983bb3dd71810f9ec668bf7273587c0a17de5e1ecddf86a6f11a6c`.
  2. The resolved graph contains exactly the approved direct Charm v2 modules:
     Bubbles `v2.2.1`, Bubble Tea `v2.0.9`, Huh `v2.0.3`, Lip Gloss `v2.0.6`,
     and Log `v2.0.0`; `go mod verify` passed and no legacy product module was
     found. See `evidence/g0-dependency-import-directed.log`.
  3. `go test ./internal/architecture` passed, including the direct
     Charm/Log-import rejection for `pkg/cmd/**`.
  4. Final repository checks passed in declared order: `make command-surface`,
     `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
     ./internal/architecture ./internal/logging ./internal/terminal`, and
     `git diff --check`; supplemental `gofmt -d` and `go vet` also passed.
     See `evidence/g0-final-repository-verification.log`.
  Changed files are limited to the G0 module lock, `internal/terminal` v2 API
  adaptation/tests, architecture boundary test, and G0 baseline/evidence/runbook
  artifacts. Residual risk: the long-list PTY test intentionally models v2
  differential terminal updates semantically rather than comparing an invalid
  contiguous full-frame byte sequence. Next: G1 is activated but is outside
  this Goal and has not been inspected or implemented.

### G1: Shared semantic terminal foundation

- 2026-09-01: initialized as `planned`; no implementation evidence.
- 2026-09-01: activated as `active`; 尚未开始实施。
- 2026-09-01: completed slice `outcome ordering`. Added the public
  `Finish(FinishOutcome, *PresentationDocument)` seam with the three approved
  outcomes, atomic finished-state commitment, invalid-outcome rejection, and
  optional one-time stdout emission. `Close` emits no synthetic result and the
  pre-existing `Result` entry remains available to un-migrated command
  adapters; the semantic recorder now captures Finish requests. Focused
  verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^TestExperienceFinishCommitsOneOutcomeAndNeverSynthesizesAnother$'`.
  Package verification also passed for `./internal/terminal
  ./internal/terminaltest`. Risk: the compatibility `Result` path still has
  legacy multi-write semantics until command migration; it is intentionally
  outside this slice. Next: implement explicit `Milestone` routing and its
  Rich/Plain/Automation behavior.
- 2026-09-01: completed slice `milestones`. Added the semantic
  `Milestone(PresentationDocument)` operation and recorder support. Plain
  Interactive emits one control-free diagnostic line, Automation drops the
  checkpoint without touching its streams, and Rich routes it through the
  existing single Bubble Tea root; finished or closed runs reject it. Focused
  verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^TestExperienceMilestoneRoutesByInteractionCapability$'`; the full
  terminal and terminaltest package tests also passed. Risk: Rich milestones
  currently share the Live View notice storage and are not yet replayed as a
  bounded semantic ledger. Next: implement phase definitions, stable IDs, and
  transition/drain protocol validation.
- 2026-09-01: completed slice `Huh interaction adapter`. Rich Text, Secret,
  Select, MultiSelect, and Confirm all remain Huh v2 child models under the
  one Bubble Tea root. Select and MultiSelect now use the same native
  two-step filtering flow (Enter or Escape commits the filter; Enter submits),
  and the wrapper delegates Escape to Huh whenever Huh reports that filtering
  is active. Added focused child-model regressions for filter state and
  submission, then exercised the 200-option PTY flow across navigation,
  resize, filtering, selection, and final result. Verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^(TestRichHuhFormLetsHuhEndListFiltering|TestRichHuhFormSequenceEndsListAndSubmits|TestRichRuntimeLongListsStayVisibleAcrossNavigationAndResize)$'`.
  Risk: PTY assertions intentionally model Huh v2's differential redraws by
  semantic transition rather than contiguous frame bytes. Next: prove Rich
  lease teardown ordering with transcript replay, deferred diagnostics, and
  stdout result tests.
- 2026-09-01: completed slice `Rich lease/replay ordering`. Added a focused
  PTY Finish path that proves the sole AltScreen session restores before the
  semantic checkpoint and outcome replay, then releases deferred diagnostics
  in FIFO order, and only then emits the durable stdout result. Verification
  passed: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./internal/terminal -run '^TestRichRuntimeFinishRestoresThenReplaysBeforeDiagnosticsAndResult$'`.
  Risk: the PTY test observes ordering from the merged inherited terminal
  stream; redirected-stdout separation remains covered by the existing Rich
  redirected-output test. Next: audit Automation transcript suppression and
  phase/transcript contract edges with focused tests.
- 2026-09-01: completed slice `Automation transcript suppression`. Automation
  now drops milestones and final tracked phases before they reach the ledger,
  and `Finish` neither appends nor freezes a Transcript in that mode; command
  stdout results remain available. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^(TestAutomationRunDoesNotRecordInteractionTranscript|TestExperienceMilestoneRoutesByInteractionCapability|TestExperienceTrackValidatesPhaseProtocolAndDrainsUpdates|TestExperienceFinishCommitsOneOutcomeAndNeverSynthesizesAnother)$'`.
  Risk: the test uses the runtime ledger directly to prove non-emission state,
  while stream control-free degradation is verified separately. Next: force
  Plain, Automation, and preflight Rich fallback result rendering through the
  control-free projection.
- 2026-09-01: completed slice `single Huh form stack and capability
  degradation`. Removed the obsolete private `richListForm` implementation now
  that Huh v2 Select/MultiSelect satisfy the largest-list PTY evidence. Durable
  results are control-free for Plain and Automation modes (and after a Rich
  preflight fallback), even if callers provide terminal/color stream flags.
  Added focused tests for both modes; `internal/terminal` and
  `internal/terminaltest` package verification passed. Risk: Rich retains its
  normal styled result only while Rich remains enabled and stdout is a terminal;
  command-owned result wording is unchanged. Next: complete phase protocol and
  Transcript edge-case coverage before the repository-wide checks.
- 2026-09-01: completed slice `Transcript outcome retention`. The bounded
  ledger now reserves capacity for one truncation marker and the final outcome,
  keeps event sequence numbers contiguous, and retains the outcome after later
  events overflow the configured event or byte budget. Focused Transcript
  truncation verification passed, including the final outcome projection.
  Risk: deliberately tiny caller-supplied byte budgets may be unable to fit the
  reserved marker and outcome; no unbounded fallback is introduced. Next: audit
  phase protocol and interaction edge cases.
- 2026-09-01: completed slice `MultiSelect selection order`. The Huh v2
  MultiSelect adapter tracks semantic selection order across deselection and
  re-selection while preserving Huh's original option values. Focused
  re-selection verification passed. Risk: order tracking is only applied to
  Rich Huh MultiSelect answers; Plain parsing retains its established order.
  Next: audit phase protocol, control-free interaction projections, and Rich
  renderer recovery boundaries.
- 2026-09-01: completed slice `phase catalog/update contract`. The phase
  protocol now rejects conflicting compatibility catalog fields and conflicting
  update ID aliases while accepting identical aliases and canonicalizing a
  `PhaseID`-only update. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^TestPhaseProtocol'`. Risk: legacy name-based operations remain
  supported for un-migrated adapters; catalog-backed operations use stable IDs.
  Next: harden interaction default shapes and control-free prompt projections.
- 2026-09-01: completed slice `interaction default-shape validation`. Select,
  MultiSelect, Text, Secret, and Confirm defaults now reject incompatible answer
  fields and duplicate multi-selections before any input or diagnostic I/O.
  Focused malformed-request verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^TestInteractionValidationRejectsMalformedRequestsBeforeIO$'`. Risk:
  command-owned `ParsePlain` requests may still omit option catalogs for their
  Plain grammar; Rich construction must continue to reject an empty Huh list.
  Next: enforce control-free prompt and option projections.
- 2026-09-01: completed slice `control-free interaction projection`. Plain
  prompts and Huh titles, descriptions, placeholders, and option labels now
  strip ANSI/C1 controls while retaining semantic option values unchanged.
  Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  -run '^(TestInteractionPromptProjectionStripsTerminalControlWithoutChangingValues|TestHuhOptionsPreserveValuesAndProjectDescriptions|TestWritePlainRendersDurableDocumentToStdoutWithoutTerminalControl)$'`.
  Risk: the shared projection replaces non-line control characters visibly,
  rather than silently removing business text. Next: exercise Rich renderer
  failure/recovery and Transcript minimum-limit edges.
- 2026-09-01: completed slice `cancellation drain safety`. Cancellation
  callbacks now start asynchronously so an unbuffered producer can publish its
  cancellation update while the runtime drains through channel closure; the
  callback remains at-most-once. Focused verification passed, including an
  unbuffered cancellation producer. Risk: a command cancellation callback must
  still eventually return; the runtime cannot cancel a callback that blocks
  independently. Next: exercise Rich renderer recovery and Transcript budget
  edges.
- 2026-09-01: completed slice `Transcript minimum limits and empty milestones`.
  Mandatory truncation/outcome events are preserved under tiny configured
  budgets, event numbering remains contiguous, events after an outcome are
  ignored, and empty/control-only milestones are no-ops. Focused Transcript and
  milestone verification passed. Risk: tiny budgets are normalized upward to
  accommodate mandatory semantic events. Next: exercise Rich renderer failure
  recovery.
- 2026-09-01: completed slice `interaction error and phase projection`.
  Plain validation feedback and Rich phase labels/details now pass through the
  shared control-free projection, while semantic answers and phase identity are
  unchanged. Focused prompt, option, validation-error, and phase-view tests
  passed. Risk: replacement characters make malformed control bytes visible for
  diagnosis without allowing terminal interpretation. Next: exercise Rich
  renderer failure recovery.
- 2026-09-01: completed slice `Rich renderer failure recovery`. A started Rich
  controller that exits with a renderer error now restores/replays the bounded
  safe transcript while its lease is held, releases deferred diagnostics, keeps
  the original error, and blocks later semantic retries instead of silently
  repeating interaction or work. Focused recovery verification passed with a
  stopped-controller fixture; existing Rich PTY teardown tests remain green.
  Risk: a renderer failure terminates that Experience run, so callers must use
  their normal error/Close path rather than retrying semantic operations.
  Next: run the complete directed G1 verification.

#### Historical A visual acceptance records (superseded by B; retained verbatim)

The following entries record the 2026-09-01 Signal Rail review and remain
historical evidence only. They do not satisfy the current B Ops Console visual
re-acceptance prerequisite.

- 2026-09-01: completed Directed verification for G1. The full terminal and
  terminaltest suites passed with `-count=1`, including Rich PTY lifecycle,
  five Huh interaction families, phase validation/draining, transcript bounds,
  lease ordering, capability degradation, and exactly-once result behavior.
  The same packages also passed `go test -race -count=1`. Risk: manual Signal
  Rail visual acceptance is still outstanding. Next: run Repository
  verification in the declared order.
- 2026-09-01: Repository verification initially exposed a compatibility
  regression in the shared control projection: converting tabs to spaces
  changed the existing `git heat` result and the root `fs` help snapshot. The
  projection was corrected to preserve tabs (and existing line layout) while
  still removing ANSI/C1 controls; focused `git heat`, root command-surface,
  and terminal package tests then passed. Risk: redirected result bytes remain
  a compatibility boundary. Next: rerun `make check` from the declared start.
- 2026-09-01: `make check` then exposed a second compatibility edge: legacy
  Plain `ParsePlain` interactions intentionally permit an empty prompt message
  because the command owns their grammar. Validation was narrowed so only
  non-parser requests require a visible message; zip cancellation and all
  previously failing package tests passed afterward. Tabs are preserved in
  durable projections, so the original `git heat` and root help contracts stay
  byte-compatible. Next: rerun `make check` to completion.
- 2026-09-01: Repository verification completed successfully in the declared
  order after the compatibility fixes. `make check` passed module verification,
  no-bun/import checks, web lint/typecheck/tests/build and asset verification,
  Go formatting, `go vet ./...`, and the complete repository test suite;
  `git diff --check` passed. G1 automated evidence is now complete: (1) all
  semantic methods and protocol errors have focused tests; (2) Rich, Plain,
  Automation, redirected streams, and Text/Secret/Select/MultiSelect/Confirm
  coverage pass; (3) transcript bounds, lease/replay teardown, cancellation
  draining, renderer recovery, and exactly-once outcome behavior pass. Risk:
  the required Signal Rail visual review remains a human decision. Next:
  start the local prototype and record the one-time manual acceptance handoff.
- 2026-09-01: completed slice `operation identity alias validation`. The
  phase protocol rejects conflicting `TrackedOperation.ID` and
  `OperationID` compatibility fields before consuming updates, with focused
  protocol coverage passing. Risk: existing adapters that omit both fields
  retain their legacy behavior. Next: rerun final Directed and Repository
  verification.
- 2026-09-01: final post-fix Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./internal/terminal ./internal/terminaltest`. Final Repository verification
  passed in declared order: `make check` completed module, web, vet, and full
  repository tests, followed by `git diff --check`. No command migration or
  command-owned wording changed in G1; the only touched production boundaries
  are shared `internal/terminal` mechanics and their fixtures. Next: launch
  the Signal Rail prototype and issue the manual acceptance handoff.
- 2026-09-01: manual acceptance handoff issued for G1; awaiting one explicit
  human result, and the Gate remains `active`. Acceptance entry from the
  repository root:
  `make prototype-terminal` (default Signal Rail success journey). Optional
  scenarios are `make prototype-terminal PROTOTYPE_ARGS='--variant=signal
  --outcome=failure'`, `make prototype-terminal
  PROTOTYPE_ARGS='--variant=signal --outcome=cancel'`, and the same commands
  with `NO_COLOR=1`; resize the PTY for narrow-dimension review. The entry was
  started and smoke-exercised in a PTY: the five-step Signal Rail, masked
  credential field, active spinner/work view, AltScreen lifecycle, and
  cancellation restoration/transcript were observed successfully.
  Automated evidence is complete: Directed
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./internal/terminal ./internal/terminaltest` passed; the supplemental
  package race run passed; Repository `make check` passed module locks,
  architecture/no-bun checks, web lint/typecheck/tests/build/assets, `go vet
  ./...`, and the complete repository suite; final `git diff --check` passed.
  Minimum acceptance checklist: (1) wide and narrow terminal layouts remain
  legible; (2) colored TTY and `NO_COLOR=1` keep state meaning via symbols and
  labels; (3) focus, pending/active/completed/warning/error symbols, and loading
  remain visible; (4) AltScreen exits before the durable transcript; (5) the
  ordering is transcript, deferred diagnostics, then stdout result; (6) secret
  values and large command results never appear in the transcript. Expected
  reply format is exactly `通过` or `未通过: <reason>` (include concise
  evidence with a failure). Next action: only a new Goal/task carrying that
  explicit result may continue G1; this Goal performs no waiting or polling.
  本次 Goal 已结束、Gate 仍为 `active`。
- 2026-09-01: manual acceptance received with the explicit result `通过`.
  All G1 Exit conditions are satisfied: (1) focused tests cover shared
  semantic methods and protocol errors; (2) Rich, Plain, Automation,
  redirected behavior, and all five Huh form families passed; (3) Transcript,
  lease/teardown, cancellation, and exactly-once behavior passed; and (4) the
  recorded Signal Rail visual review was accepted. Final automated evidence:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./internal/terminal ./internal/terminaltest`, its `-race` counterpart,
  `make check`, and `git diff --check` all passed. G1 is marked `passed`.
  Its direct successor G2 is activated but has not been inspected,
  implemented, or verified in this Goal.

### G2: Logging and root boundary

- 2026-09-01: initialized as `planned`; no implementation evidence.
- 2026-09-01: activated as `active`; 尚未开始实施。
- 2026-09-01: completed slice `private Log v2 text adapter`. Added the
  text-only adapter inside `internal/logging`; Runtime still samples the
  timestamp, filters levels, merges and recursively redacts context, owns the
  JSON/NDJSON serializer, renders one complete Log v2 text record into a
  private buffer, then performs one write to the injected lease-aware
  diagnostic writer. Text retains timestamp, complete level, scope, message,
  and normalized context; JSON/NDJSON bytes remain covered by exact snapshots.
  Added an architecture check requiring every direct `charm.land/log/v2`
  import to remain under `internal/logging`. The required transitive module
  checksums were locked in `go.mod` and `go.sum`. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/logging
  ./internal/architecture -count=1`. Risk: Log v2 text bytes intentionally
  differ from the old handwritten renderer, while the machine JSON/NDJSON
  contract is unchanged. Next: prove atomic redacted writes and FIFO ordering
  through an active renderer lease.
- 2026-09-01: completed slice `redaction and atomic lease-aware writes`.
  Added an integration regression around the real
  `terminal.LeaseAwareDiagnosticWriter`: while a renderer lease is active,
  Runtime contributes no destination writes; release flushes two FIFO records,
  each as exactly one complete line and without message or context secrets.
  This proves Log v2's internal rendering stays private to the buffer and does
  not split Runtime's serialized diagnostic write. Focused directed verification
  passed: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./internal/logging -run '^TestRuntimeBuffersRedactedRecordsForLeaseAwareFIFO$'
  -count=1`. Risk: the test covers sequential Runtime calls; mutual exclusion
  remains owned by Runtime and the terminal writer's existing concurrency
  tests. Next: migrate the root discovery/help presenter so write failures
  propagate without changing durable stdout semantics.
- 2026-09-01: completed slice `root discovery/help result propagation`.
  The root discovery adapter now completes one semantic run with
  `Finish(Succeeded)`, closes it, and joins both presentation and cleanup
  errors. The fresh Cobra tree used by each invocation captures a help-callback
  presentation failure and routes it through the existing single root-error
  path; no-argument discovery uses the same error-returning presenter. Added
  regressions for the Finish/Close operation order and for a failing stdout
  writer (one write attempt, one stderr diagnostic, code 1). Focused directed
  verification passed: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/root -run
  '^(TestTerminalDiscoveryPresenterTranslatesAndClosesOneExperienceRun|TestDiscoveryWriteFailureBecomesOneRootDiagnostic)$'
  -count=1`. Risk: Cobra owns the void help callback API, so the per-tree
  capture stays deliberately local to `App.Execute`. Next: prove version and
  completion continue to bypass discovery styling and invalid diagnostic
  controls.
- 2026-09-01: completed slice `root version and completion bypass` without
  production changes. Existing root behavior already keeps `--version` as the
  exact version plus newline and Cobra completion as raw stdout, independent of
  Rich terminal capability and invalid same-invocation diagnostic controls.
  Focused directed verification passed: `GOTOOLCHAIN=go1.26.7 GOWORK=off
  CGO_ENABLED=0 go test ./pkg/cmd/root -run
  '^(TestGlobalLogPrecedenceAndVersionBypass|TestDiagnosticConfigurationRunsBeforeEffectsAndBypassesDiscovery|TestRichTerminalDiscoveryPreservesVersionAndRawCompletion)$'
  -count=1`. Risk: raw completion compatibility is additionally guarded by
  the frozen command-surface suite, to be run in repository verification.
  Next: make parser and root error projections redacted, control-free,
  single-line diagnostics while preserving error ownership and exits.
- 2026-09-01: completed slice `parser and root-error safe projection`.
  Added `logging.RedactDiagnostic` for control-free single-line diagnostic
  text and a root-owned 1024-rune bound with `[truncated]` omission marker.
  Normal root errors and panic summaries now remove ANSI/C1 controls and CR/LF,
  redact credential-shaped text, and preserve the original error chain in
  `Outcome.Err`; debug stack behavior is unchanged. Existing parser recovery
  wording, one-line stderr ownership, exits, and panic mapping remain intact.
  Focused directed verification passed: `GOTOOLCHAIN=go1.26.7 GOWORK=off
  CGO_ENABLED=0 go test ./pkg/cmd/root -run
  '^(TestRootDiagnosticProjectionIsRedactedControlFreeAndBounded|TestParserRecoveryUsesOneActionableErrorLine|TestParserRecoveryLeavesCommandErrorsUntouched|TestPanicMappingRedactsAndAddsDebugStack)$'
  -count=1`. Risk: diagnostic presentation now deliberately collapses unusual
  whitespace; ordinary parser and command wording remains unchanged. Next: run
  the full G2 Directed logging/root/architecture suites before repository
  verification.
- 2026-09-01: completed G2 Directed verification. `GOTOOLCHAIN=go1.26.7
  GOWORK=off CGO_ENABLED=0 go test ./internal/logging ./pkg/cmd/root
  ./internal/architecture -count=1` passed. This covers Log v2 text records,
  exact JSON/NDJSON records, dynamic level/format changes, recursive redaction,
  injected/lease-aware writers, architecture privacy, and root
  help/discovery/version/completion/parser recovery including the frozen
  command-surface comparison. Risk: none beyond the intentional text-only Log
  v2 visual byte change; machine-facing JSON/NDJSON and raw root machine
  outputs remain exact. Next: run the declared G2 Repository verification in
  order.
- 2026-09-01: completed corrective slice `Log v2 lifecycle-log assertion`.
  The first `make check` after the terminal short-write slice exposed one
  related, stale text-format assertion in `pkg/cmd/tunnel/server/run_test.go`:
  it expected the retired `[tunnel.server]` decoration even though the tested
  Runtime path correctly emitted the required scope and all lifecycle messages
  as `tunnel.server:` through the approved Log v2 text adapter. The test now
  independently asserts the scope and all three lifecycle messages, preserving
  service behavior coverage without treating a plan-authorized human-text
  delimiter as a machine contract. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/tunnel/server -run
  '^TestRunServerOwnsTheForegroundUntilContextCancellation$' -count=1`.
  Risk: no production behavior changed; Repository verification must be
  restarted in its declared order after this related-test correction. Next:
  rerun all G2 Repository checks, then record per-condition evidence if they
  pass.
- 2026-09-01: all G2 Exit conditions passed; Gate marked `passed`.
  1. Log v2 remains private behind `internal/logging.Runtime`: the private
     buffered text adapter receives only complete normalized/redacted records,
     and `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
     ./internal/architecture -count=1` passed with the direct-import boundary
     enforcement.
  2. Text and exact JSON/NDJSON logging behavior, dynamic configuration,
     recursive redaction, injected writers, and lease-aware FIFO/atomic
     ordering passed in `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
     ./internal/logging -count=1`; the lease-focused regression and the
     terminal short-write regression also passed. Human text follows the
     approved Log v2 projection; machine JSON/NDJSON bytes remain exact.
  3. Root help, discovery, version, completion, parser recovery, safe
     single-line root diagnostics, error exits, and frozen command-surface
     behavior passed in `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
     ./pkg/cmd/root -count=1` and `make command-surface`.
  4. Final Repository verification passed in the declared order:
     `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
     ./internal/logging ./pkg/cmd/root ./internal/architecture`,
     `make command-surface`, `make check`, and `git diff --check`.
     `make check` passed module verification, Web lint/type/tests/build, and
     all Go packages, including `pkg/cmd/tunnel/server`.
  Manual acceptance: 无. Residual risk is limited to the intentionally changed
  human-readable Log v2 text bytes; logging scope, level, timestamp, message,
  context, redaction, record atomicity, and all machine-facing output contracts
  are covered by the recorded evidence. Next: G3 is activated but is outside
  this Goal and has not been inspected, implemented, or verified.

### G3: Low-risk finite command slices

- 2026-09-01: initialized as `planned`; no implementation evidence.
- 2026-09-01: activated as `active`; 尚未开始实施。

### G4: External-read and service-startup slices

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G5: Mutating, destructive, and archive slices

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G6: Git mutation and process handoff

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G7: Service lifecycles and detached updater

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G8: Release candidate and final acceptance

- 2026-09-01: initialized as `planned`; no implementation evidence.
