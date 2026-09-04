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
- G3 command slices were gated on a fresh B visual re-acceptance. That
  prerequisite was explicitly accepted on 2026-09-02; each command slice still
  requires its own evidence and acceptance before G3 can pass.

## Goal Ledger

| Gate | Status | Depends on | Plan contract | Unlock evidence |
| --- | --- | --- | --- | --- |
| G0: Baseline and dependency lock | passed | none | `implementation-plan.md` -> `G0: Baseline and dependency lock` | no predecessor |
| G1: Shared semantic terminal foundation | passed | G0 | `implementation-plan.md` -> `G1: Shared semantic terminal foundation` | G0 passed |
| G2: Logging and root boundary | passed | G1 | `implementation-plan.md` -> `G2: Logging and root boundary` | G1 passed |
| G3: Low-risk finite command slices | passed | G2 + B visual re-acceptance | `implementation-plan.md` -> `G3: Low-risk finite command slices` | G2 passed; B Ops Console visual re-acceptance passed 2026-09-02; G3 Exit conditions passed 2026-09-03 |
| G4: External-read and service-startup slices | passed | G3 | `implementation-plan.md` -> `G4: External-read and service-startup slices` | G3 passed; manual FS acceptance passed 2026-09-03; G4 Exit conditions passed 2026-09-03 |
| G5: Mutating, destructive, and archive slices | active | G4 | `implementation-plan.md` -> `G5: Mutating, destructive, and archive slices` | G4 passed; 已激活，尚未开始实施 |
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
- 2026-09-02: manual acceptance handoff issued for the prerequisite B Ops
  Console visual re-acceptance. No G3 command slice was inspected, changed,
  or verified in this Goal because the decision log still requires an explicit
  human result before command work may begin. Acceptance entry from the
  repository root: `make prototype-terminal` (default B / `console` success
  journey); optional checks are
  `make prototype-terminal PROTOTYPE_ARGS='--variant=console --outcome=failure'`,
  `make prototype-terminal PROTOTYPE_ARGS='--variant=console --outcome=cancel'`,
  `NO_COLOR=1 make prototype-terminal`,
  `env -u NO_COLOR TERM=xterm-256color make prototype-terminal`, and a 55x15
  PTY resize. Existing automated evidence is the B revision record above:
  prototype `go test ./...`, prototype `go vet ./...`, targeted terminal and
  terminaltest tests, `make check`, `git diff --check`, source-baseline
  verification, and PTY smoke coverage all passed. Minimum acceptance checklist:
  (1) B `OPS CONSOLE` command/status bar, safe metadata row, aligned
  `STATE / PHASE / DETAIL` table, and active content region; (2) amber/cyan
  hierarchy and symbol-paired active/done/failed/cancelled states remain
  understandable with color and `NO_COLOR=1`; (3) Huh bottom focus rule,
  absence of a persistent left rail, and single-column narrow degradation;
  (4) AltScreen restoration precedes Transcript replay, deferred diagnostics,
  and stdout; (5) secrets and large results are absent from the Transcript.
  Expected reply format is exactly `通过` or `未通过: <reason>` with concise
  evidence for a failure. Only a new Goal carrying that explicit result may
  resume G3; this Goal performs no waiting or polling. 本次 Goal 已结束、Gate
  仍为 `active`。
- 2026-09-02: received the explicit prerequisite result `通过` for the B Ops
  Console visual re-acceptance. The command-slice lock is released for this
  Gate. The accepted evidence is the B prototype entry above; no command
  adapter is treated as accepted by this result. Next: implement and verify
  the four G3 command adapters one command per slice, starting with
  `config fork list`.
- 2026-09-02: completed slice `config fork list Experience adapter`. Opened the
  Experience before lazy store construction, tracked the bounded
  `Load fork provider instances` phase by stable ID, mapped read success/failure
  and context cancellation to one `Finish` call, and kept Plain/Automation
  result bytes unchanged. Rich now has a B-oriented eyebrow/title/subtitle,
  safe bounded row projection, empty guidance, and a small loaded-count
  milestone; malformed/control-bearing fields and unsafe token previews are
  redacted only in the Rich projection. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/fork/list -count=1`. Risk: the shared renderer still owns
  generic table wrapping, so the command's wide/narrow visual evidence remains
  to be exercised in the G3 acceptance pass. Next: implement `config cm list`.
- 2026-09-02: completed slice `config cm list Experience adapter`. Opened the
  Experience before store construction and `ListCMProfiles`, tracked the
  stable `Load CM profiles` phase, preserved stored order/default semantics and
  the exact Plain/Automation result, and submitted one `Finish` outcome on
  success, failure, or cancellation. Rich adds the B-oriented metadata/title
  hierarchy, bounded control-free fields, warning empty guidance, and a loaded
  count milestone; URL userinfo/query/fragment and unsafe profile/model values
  are excluded from the Rich projection. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/cm/list ./pkg/cmd/config/fork/list ./pkg/cmd/root -count=1`.
  Standalone acceptance for both list commands also passed with
  `go test -tags acceptance ./acceptance -run
  'TestConfigForkListStandaloneBinary|TestConfigCMListStandaloneBinary'
  -count=1`. Risk: Rich PTY dimensions and transcript ordering still need the
  final G3 acceptance run. Next: implement `config cm use`.
- 2026-09-02: completed slice `config cm use Experience adapter`. The
  noninteractive atomic writer now opens the Experience before store creation,
  uses the single `set-default-cm-profile` phase in Rich, emits exactly one
  Plain lifecycle notice, preserves the exact profile identity passed to
  `SetDefaultCMProfile`, and maps success/failure to one `Finish` call without
  resolving credentials or adding a preflight read. Rich uses a bounded
  control-free profile projection and B-oriented title/detail blocks; the
  compatibility result remains unchanged for Plain, Automation, and redirected
  output. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/cm/use -count=1`. Standalone acceptance passed with
  `go test -tags acceptance ./acceptance -run TestConfigCMUseStandaloneBinary
  -count=1`, including successful persistence and missing-profile failure
  behavior. Risk: Rich PTY/stream ordering and malicious profile-name visual
  evidence remain for the final G3 acceptance pass. Next: implement `git heat`.
- 2026-09-02: completed slice `git heat Experience adapter`. Added the three
  truthful `locate-repository`, `read-git-history`, and `rank-hot-paths` phases,
  cancellation/error mapping, Rich summary milestone, and the Rich report
  projection while preserving normalization, Git invocation, aggregation,
  sorting, time, query, empty-result, stdout, signal, and exit semantics.
  Changed `pkg/cmd/git/heat/run.go`, `pkg/cmd/git/heat/terminal.go`, and
  `pkg/cmd/git/heat/terminal_test.go`. Focused package tests, query/path
  projection tests, cancellation tests, and tagged standalone acceptance all
  passed. Risk: the final human review still needs to judge wide/narrow report
  readability and query emphasis in a real terminal. Next: complete the
  cross-command verification and acceptance handoff.
- 2026-09-02: completed corrective safety slice `config fork list Rich field
  projection`. Control-bearing and invalid-UTF-8 names, hosts, schemes,
  provider types, and token previews now use safe placeholders or redaction
  instead of echoing any portion of the unsafe value; Plain/Automation result
  bytes and appconfig projections remain unchanged. Changed
  `pkg/cmd/config/fork/list/presentation.go` and
  `pkg/cmd/config/fork/list/terminal_test.go`. Focused and race tests passed.
  Risk: only the Rich projection is changed; manual review must confirm the
  placeholders remain understandable at narrow dimensions. Next: rerun the
  declared G3 repository sequence and hand off for command acceptance.
- 2026-09-02: G3 automated verification completed after the corrective slice.
  Directed evidence: all four focused packages passed with
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/config/fork/list
  ./pkg/cmd/config/cm/list ./pkg/cmd/config/cm/use ./pkg/cmd/git/heat -count=1`
  (run per package), the shared Rich PTY/transcript tests passed with
  `go test ./internal/terminal -run 'TestRich.*PTY|TestTracked.*PTY|Test.*Transcript|Test.*Finish' -count=1`,
  all four tagged standalone acceptance tests passed, and the affected-package
  race suite passed. Repository evidence, in declared order, passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/... ./pkg/cmd/git/heat ./internal/terminal`,
  `make command-surface`, `make check`, and `git diff --check`.
  `make check` covered Web lint/type/test/build, Go vet, and all Go packages.
  Real PTY smoke also passed for all four commands at 120x40 with
  `env -u NO_COLOR TERM=xterm-256color`, at 40x15 with `NO_COLOR=1`, and for
  CM-list stderr Rich UI with stdout redirected; every run exited 0, retained
  phase/Transcript/result ordering, and the redirected result stayed the
  unchanged durable text. The standalone fixture is the binary
  `/tmp/g3-terminal-acceptance/ycy-final` (SHA-256
  `212d3ab08dc9982410fe54a30a6c14daeba2eada23b0e484f31e3620a7f841cf`) with
  data under `/tmp/g3-terminal-acceptance/home` and
  `/tmp/g3-terminal-acceptance/repo`. Automated evidence covers G3 Exit
  conditions 1, 2, and the automated portion of 3; no command result, storage,
  Git, API, or compatibility contract changed. Residual risk is the required
  human visual judgment at wide/narrow dimensions and explicit confirmation
  that no secret or large result appears in the Transcript. Next: one-time
  manual acceptance handoff; Gate remains `active`.
- 2026-09-02: manual acceptance handoff issued for G3. Acceptance entry is
  the standalone binary `/tmp/g3-terminal-acceptance/ycy-final` with the
  prepared fixtures above. Run each command in a colored TTY at wide and
  narrow dimensions (for example, `stty columns 120 rows 40` and
  `stty columns 40 rows 15`), then repeat with `NO_COLOR=1`; use the same
  fixture repository for `git heat`. The minimum checklist is: (1) each
  command shows its safe title/metadata, loading or focus state, and final
  phase in the B Ops Console hierarchy; (2) wide and narrow layouts keep all
  semantic fields/rows understandable, with symbols and labels still clear
  without color; (3) `config fork list` preserves instance order and only the
  encrypted preview, `config cm list` preserves profile order/default and
  never exposes API keys, `config cm use <profile>` persists the exact target
  and reports missing targets safely, and `git heat` preserves range/target/
  sort/query semantics including empty and long/control-bearing paths; (4)
  AltScreen restoration is followed by bounded Transcript, deferred
  diagnostics, and exactly one unchanged stdout result, including when stdout
  is redirected; (5) no secret, raw error, absolute path, or large report is
  present in the Transcript. Expected explicit reply format is exactly
  `通过: <commands>` or `未通过: <command and reason>`. This is a one-time
  handoff and termination point for this Goal; no polling or Gate transition
  is performed. 本次 Goal 已结束、Gate 仍为 `active`。
- 2026-09-03: the manual acceptance result for the G3 `git heat` report was
  not accepted. The observed colored-TTY run showed tab-dependent column drift,
  generic Rich wrapping splitting the table into detached path lines, and long
  directory paths appearing truncated at the terminal edge. The failure was
  isolated to `pkg/cmd/git/heat/terminal.go` Rich projection; Git aggregation,
  ordering, range/query semantics, and the Plain/Automation result were not
  implicated. Started corrective slice `git heat Rich responsive report` within
  G3 only. The slice adds no shared-renderer exception: it captures the injected
  stdout TTY width in the git heat command adapter, uses space-padded columns and
  explicit path continuation lines at wide widths, and uses labelled
  `Changed at` / `Changes` / target records at narrow widths. Regression tests
  first reproduced the failure and then passed at 120 and 40 columns, including
  latest/query markers and control-safe long paths. Changed files are limited
  to `pkg/cmd/git/heat/command.go`, `pkg/cmd/git/heat/terminal.go`, and
  `pkg/cmd/git/heat/terminal_test.go`. Next: complete the declared directed and
  repository verification, then issue one new manual acceptance handoff; G3
  remains `active`.
- 2026-09-03: recorded completion of the corrective slice's automated
  verification. `go test ./pkg/cmd/git/heat -count=1`, the Rich PTY/transcript
  focused tests, all four standalone acceptance tests, and the affected-package
  race suite passed. The declared G3 repository sequence also passed in order:
  focused repository tests, `make command-surface`, `make check`, and
  `git diff --check`; the one transient Tunnel FRP integration failure was
  cleared by the pinned `go1.26.7` rerun, which passed in full. Real PTY smoke
  passed at 120x40 with color, 40x15 with `NO_COLOR=1`, and the redirected
  stdout path. The final standalone binary is
  `/tmp/g3-terminal-acceptance/ycy-final` with SHA-256
  `95c6236b089aeec612999c5e8ca84a27ea1dc3f0a193dd058d639749a66763e2`;
  transcripts are under `/tmp/g3-terminal-acceptance/final-pty/`. The
  corrective changes remain limited to the three files named above. Automated
  Exit conditions are satisfied; the remaining condition is explicit human
  acceptance of the report presentation and transcript safety.
- 2026-09-03: new one-time manual acceptance handoff issued for G3. Use the
  prepared binary `/tmp/g3-terminal-acceptance/ycy-final` and fixture repository
  `/tmp/g3-terminal-acceptance/repo`; fixture data is under
  `/tmp/g3-terminal-acceptance/home`. In a real TTY, run the four commands
  (`config fork list`, `config cm list`, `config cm use <an existing profile>`,
  and `git heat` from the fixture repository) first at 120x40 with color, then
  at 40x15 with `NO_COLOR=1`. The minimum checklist is: (1) the B `OPS CONSOLE`
  title/metadata, loading or focus state, and final phase are legible; (2) the
  wide `git heat` columns stay aligned, long paths wrap as complete continuation
  lines, and the narrow view keeps labelled `Changed at`, `Changes`, and target
  fields understandable; (3) symbols and labels remain clear without color,
  latest/query markers remain visible, and no control characters leak; (4) after
  the screen restores, the bounded transcript contains no secret, raw error,
  absolute path, or large report, while exactly one unchanged durable result is
  emitted (including with stdout redirected); (5) list order/defaults and the
  exact `config cm use` target remain correct. Reply exactly `通过: <commands>`
  when the checklist passes, or `未通过: <command and reason>` with the first
  failing command and evidence. This is the termination handoff for this Goal;
  no Gate transition is performed. 本次 Goal 已结束、Gate 仍为 `active`。
- 2026-09-03: completed corrective slice `Rich capability-query output
  boundary` for the reported `2026;2$y` shell-input leak. Bubble Tea v2 queues
  DEC 2026/2027 probes in its output buffer; a finite command can restore the
  inherited terminal before those probes are flushed, and a terminal response
  can then be parsed as the next shell command. The Rich renderer writer now
  owns that protocol boundary, suppresses only the optional 2026/2027 probe
  sequences across split writes, flushes non-probe bytes before transcript
  replay, and leaves normal ANSI output and the command result path unchanged.
  Changed files are `internal/terminal/rich.go`,
  `internal/terminal/rich_runtime_pty_test.go`, and
  `internal/terminal/rich_writer_test.go`; no command API, storage, Git, or
  durable result contract changed. The regression PTY test and split-write
  writer tests passed 10 consecutive runs. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/... ./pkg/cmd/git/heat ./internal/terminal`, the four tagged
  standalone acceptance tests, and the affected-package `-race` suite.
  Repository verification passed in order: focused repository tests,
  `make command-surface`, `make check`, and `git diff --check`. A fresh binary
  `/tmp/g3-terminal-acceptance/ycy-final` has SHA-256
  `324b19a53f27fd6982c0b20ea0cf76aad9098c925654c225010d9f3108ff4b9d`.
  Real shell PTY smoke passed for all four G3 commands at 120x40 with color
  and 40x15 with `NO_COLOR=1`; the redirected `config cm list` result remained
  durable and control-free. The fixed shell handoff transcript contains no
  `2026;2$y` and no shell command error, while the pre-fix comparison
  reproduced both. Records are under `/tmp/g3-terminal-acceptance/final-pty/`
  and `/tmp/g3-terminal-acceptance/late-report-fixed2.log`. Automated Exit
  conditions remain satisfied; the only remaining condition is explicit human
  acceptance of the terminal presentation and transcript safety. Next: issue
  one new manual acceptance handoff; G3 remains `active`.
- 2026-09-03: new one-time manual acceptance handoff issued for G3 after the
  capability-query fix. Use `/tmp/g3-terminal-acceptance/ycy-final` with
  fixture repository `/tmp/g3-terminal-acceptance/repo` and fixture data under
  `/tmp/g3-terminal-acceptance/home`. In a real TTY, run `config fork list`,
  `config cm list`, `config cm use work`, and `git heat -n 20 -t dirs -s path
  -q terminal` from the fixture repository at 120x40 with color, then repeat
  at 40x15 with `NO_COLOR=1`. Confirm that no `2026;2$y` or shell error appears
  after each command and that a following shell command runs immediately. The
  minimum checklist is: (1) each B `OPS CONSOLE` title/metadata, loading or
  focus state, and final phase remains legible; (2) `git heat` columns stay
  aligned at wide width and labelled `Changed at`, `Changes`, and target fields
  remain understandable at narrow width; (3) symbols, labels, latest/query
  markers, and control-safe paths remain clear without color; (4) AltScreen
  restoration is followed by the bounded transcript and exactly one unchanged
  durable result, including redirected stdout; (5) no secret, raw error,
  absolute path, large report, or capability response appears in the
  transcript. Reply exactly `通过: <commands>` when all checks pass, or
  `未通过: <command and reason>` with the first failing command and evidence.
  This is the termination handoff for this Goal; no Gate transition is
  performed. 本次 Goal 已结束、Gate 仍为 `active`。
- 2026-09-03: received the explicit human acceptance result `确认通过` for
  the G3 handoff. The four-command colored-TTY and `NO_COLOR=1` checks passed
  at wide and narrow dimensions: `config fork list`, `config cm list`,
  `config cm use work`, and `git heat -n 20 -t dirs -s path -q terminal` from
  `/tmp/g3-terminal-acceptance/repo`. The reviewer confirmed that the B Ops
  Console hierarchy, responsive `git heat` report, symbols/labels, list order
  and defaults, exact `config cm use` target, transcript safety, unchanged
  redirected result, and immediate shell re-entry were correct; no
  `2026;2$y` or shell error remained.
- 2026-09-03: all G3 Exit conditions passed and G3 is marked `passed`.
  1. The four adapters preserve baseline behavior and durable result bytes:
     focused command tests, the four tagged standalone acceptance tests, and
     the redirected `config cm list` PTY result passed; no storage, Git, API,
     default, or compatibility contract changed.
  2. Rich, Plain, Automation, and redirected coverage passed for every G3
     command, including the shared Rich PTY/transcript/form tests, wide and
     narrow `NO_COLOR` smoke, and standalone fixtures.
  3. Phase, cancellation, redaction, Transcript bounds, responsive visual
     records, and the capability-query boundary are covered by the focused
     tests and PTY evidence. The pre-fix shell comparison reproduced
     `2026;2$y` and `command not found`; the fixed handoff transcript contains
     neither, and the next shell command executes immediately. The final
     binary is `/tmp/g3-terminal-acceptance/ycy-final` (SHA-256
     `324b19a53f27fd6982c0b20ea0cf76aad9098c925654c225010d9f3108ff4b9d`),
     with records under `/tmp/g3-terminal-acceptance/final-pty/`.
  4. G3 Repository verification passed in the declared order:
     `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
     ./pkg/cmd/config/... ./pkg/cmd/git/heat ./internal/terminal`,
     `make command-surface`, `make check`, and `git diff --check`; the
     affected-package `-race` suite and standalone acceptance suite also
     passed.
  Direct successor G4 is activated as `active` with the record
  `已激活，尚未开始实施`; no G4 code, analysis, or verification is part of
  this Goal. Effort for this Goal is complete.

### G4: External-read and service-startup slices

- 2026-09-01: initialized as `planned`; no implementation evidence.
- 2026-09-03: activated after G3 passed; 已激活，尚未开始实施。 This Gate is
  outside the completed G3 Goal.
- 2026-09-03: locked G4 slice `export env Experience adapter` as the first
  independently revertible command slice. Scope is limited to the export env
  adapter, its command-owned phase/result/transcript projections, and focused
  tests; no other G4 command or service surface is inspected or changed.
  Next: implement the declared export phases and preserve dotenv, merge, file,
  cancellation, and stdout contracts before running directed verification.
- 2026-09-03: completed slice `export env Experience adapter`. The terminal
  path now reports the approved resolve/discover/select/read/parse/encode/write
  phases through two independent semantic clusters, submits one `Finish`,
  records safe selection and variable milestones, and writes an `--out` target
  before emitting its success result. Unique candidates (including a sole
  `.env.<name>`) are selected without prompting; ambiguous Automation stops at
  the selection boundary, and cancellation performs no file I/O. The legacy
  Module presenter path remains available for compatibility callers. Changed
  files: `pkg/cmd/export/env/{run.go,selection.go,terminal.go}` and focused
  tests. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/export/env -count=1`, its `-race` run, and `go vet` for the
  package. Focused root regressions passed in `./pkg/cmd/root`; `git diff
  --check` passed. Risk: Rich PTY visual evidence and the full G4 repository
  sequence remain pending. Next: lock and implement the `config cm test`
  provider-read slice.
- 2026-09-03: locked G4 slice `config cm test provider-read Experience adapter`.
  Scope is limited to profile resolution/provider-request phase projection,
  safe Rich result/transcript handling, and focused command tests; provider
  protocol, timeout, response, error, stdout, and exit semantics remain fixed
  by the decision document. Next: inspect the resolver and transport seams,
  then implement the two declared phases.
- 2026-09-03: completed slice `config cm test provider-read Experience adapter`.
  The command now opens Experience before store/profile resolution, exposes
  the declared `resolve-cm-test-profile` and `test-cm-provider` phases, and
  keeps resolver cancellation from starting a provider request. Provider
  failures are classified as `http-status`, `timeout`, `read`, `decode`, or
  `empty-response`; non-2xx responses are projected without reading or
  emitting response bodies. Rich output restores the primary screen, replays
  a bounded safe Transcript, includes the response/usage summary, and submits
  the durable result once; Plain, Automation, redirected, cancellation,
  redaction, and write-failure contracts remain covered. Changed files:
  `pkg/cmd/config/cm/test/{terminal.go,terminal_run_test.go,terminal_pty_test.go,
  test_response.go,test_response_test.go,test_run.go,test_safety.go,
  test_transport.go,test_transport_test.go}` and
  `internal/terminal/{experience.go,rich_recovery_test.go}`. Directed
  verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/cm/test -count=1`, its `-race` run,
  `go test ./internal/terminal -count=1`, the combined `-race` run for
  `./internal/terminal ./pkg/cmd/config/cm/test`, and `go vet` for both
  packages. The tagged standalone acceptance
  `TestConfigCMTestStandaloneBinaryUsesOnlyTheLocalProvider` passed, and
  `git diff --check` passed. Risk: Rich visual evidence and the remaining G4
  command/service slices plus repository sequence remain pending. Next: lock
  the `git pulse` presentation slice.
- 2026-09-03: locked G4 slice `git pulse Experience adapter`. Scope is limited
  to the command-owned preparation/scan/fetch/report phase and interaction
  projection, safe Rich Transcript/report rendering, and focused tests. Git
  discovery, concurrency, filtering, report bytes, process ownership, exits,
  and cancellation semantics remain fixed by the decision document. Next:
  inspect the existing adapter and module boundaries before implementing the
  declared presentation contract.
- 2026-09-03: completed slice `git pulse Experience adapter`. The command now
  projects `prepare-workspace`, `scan-repositories`, `fetch-commits`, and
  `build-commit-tree` through the Experience without changing discovery,
  filtering, report bytes, process ownership, cancellation, or exit behavior.
  Directory-read and per-repository fetch failures become bounded,
  relative-path-safe warnings; a partial report retains the exact existing
  result semantics. Rich restores before replaying a safe Transcript and the
  formatted report; Plain and Automation retain control-free output and do
  not read stdin before their existing interaction boundary. Changed files:
  `pkg/cmd/git/pulse/{command.go,discovery.go,git.go,run.go,terminal.go,tracking.go,\
  presentation.go,presentation_test.go,terminal_pty_test.go}` and
  `acceptance/git_pulse_test.go`. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/git/pulse -count=1`, its `-race` run, `go vet` for the package,
  `go test ./internal/terminal -count=1`, and tagged standalone acceptance
  `TestGitPulseStandaloneBinary`; `git diff --check` passed. Risk: visual
  manual evidence and the remaining G4 command/service slices plus repository
  sequence remain pending. Next: lock and implement the `upgrade` parent
  presentation slice.
- 2026-09-03: locked G4 slice `upgrade parent Experience presentation`.
  Scope is limited to parent-command resolution, candidate, download, verify,
  and schedule phase/result projection plus focused tests. Hidden-updater
  replacement, persisted transaction behavior, artifact semantics, unusual
  exits, stdout bytes, and process handoff remain fixed by the decision
  document and are not changed. Next: inspect the existing parent adapter and
  its decision contract before implementing the declared phase sequence.
- 2026-09-03: completed slice `upgrade parent Experience presentation`. The
  parent now projects the actual consume, release/artifact resolution,
  candidate download/verification, updater staging, pending-state publication,
  detached scheduling, and completion boundaries through one Experience run.
  The optional updater observer carries only presentation-safe versions,
  platform/artifact identity, checksum source, and fixed failure classes;
  URLs, hashes, paths, headers, candidate output, and raw errors remain out of
  phases and Transcript. A consumed prior state is a first Rich Transcript
  milestone and is combined with the final durable stdout document only after
  primary-screen restoration, preserving the existing bytes without writing
  into AltScreen. Existing classified aborts, unusual exit codes, cleanup,
  detached ownership, and global startup consumption remain unchanged. Changed
  files: `internal/updater/{progress.go,release.go,candidate.go,upgrade.go}`
  and tests plus `pkg/cmd/upgrade/{command.go,terminal.go,terminal_test.go,
  terminal_pty_test.go}`. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/upgrade ./internal/updater ./internal/ycycmd ./pkg/cmd/root
  -count=1`, `go test -race ./pkg/cmd/upgrade ./internal/updater`, and `go
  vet` for both affected packages. Rich color/no-color PTY coverage proved one
  AltScreen enter/restore, safe phase visibility, prior-state/phase Transcript
  order, Transcript-before-stdout ordering, and control-free no-color output.
  The tagged standalone detached-updater acceptance
  `TestDetachedGoToGoStandaloneReplacementRollbackAndSelfCheck` passed; `git
  diff --check` passed. Supplemental full acceptance encountered unrelated
  pre-existing failures in `config fork remove`, `rm`, and `tunnel server`, so
  those surfaces were not changed in this slice. Risk: the remaining G4
  `diff`/`fs` lifecycle slices and final G4 repository sequence remain
  pending. Next: lock and implement the `diff` Lifecycle Log slice.
- 2026-09-03: locked G4 slice `diff Lifecycle Log`. Scope is limited to the
  line-oriented startup, refresh, warning, failure, and shutdown event
  projection plus focused text/NDJSON tests. Comparison discovery, snapshot
  publication, HTTP/MCP behavior, flags, stdout checkpoint bytes, service
  ownership, and exit semantics remain fixed by the decision document. Next:
  inspect the service lifecycle and repository logging seams before adding the
  declared bounded event sequence.
- 2026-09-03: completed slice `diff Lifecycle Log` sequencing foundation.
  Startup records and an immediately-completing initial refresh are now
  sequenced behind the stdout checkpoint; first-entry Debug phases remain
  deduplicated, ready snapshots retain counts/bytes/duration and adjacent issue
  warnings, and a ready refresh is preserved when shutdown cancellation races
  its terminal callback. Startup stdout-write failure now closes the lifecycle
  gate before teardown and emits exactly the explicit stopping/stopped pair,
  dropping late refresh callbacks. Changed files:
  `pkg/cmd/diff/lifecycle.go`, `pkg/cmd/diff/session.go`,
  `pkg/cmd/diff/lifecycle_test.go`. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/diff -run '^TestDiffLifecycle' -count=1`; `gofmt` and
  `git diff --check` passed. Repository verification status: the declared G4
  repository sequence remains pending. Risk: protocol-source, cancellation,
  failure, writer, redaction, concurrency, and standalone acceptance evidence
  for this slice remain to be completed. Next: add refresh coordinator and
  REST/MCP lifecycle-source tests.
- 2026-09-03: completed slice `diff Lifecycle Log` refresh coordination.
  Accepted attempts now carry `initial`/`rest`/`mcp` sources and monotonic
  service-local ordinals; active attempts are rejected without lifecycle
  records, REST cancellation is attributed to `cancelSource=rest`, and a
  subsequent MCP refresh can succeed after the cancelled run clears. Changed
  file: `pkg/cmd/diff/lifecycle_test.go` (against the coordinator wiring in
  `pkg/cmd/diff/http_refresh.go`). Directed verification passed the focused
  `TestDiffRefreshCoordinatorProjectsSourcesOrdinalsRejectionAndRESTCancellation`
  run; `gofmt` and `git diff --check` passed. Repository verification status:
  the declared G4 repository sequence remains pending. Risk: previous-snapshot
  failure retention, shutdown folding, writer failures, and text/NDJSON level
  goldens remain pending. Next: exercise failure and shutdown terminal paths.
- 2026-09-03: completed slice `diff Lifecycle Log` failure and shutdown
  terminal paths. Cancelled and failed attempts now prove bounded progress and
  previous-snapshot fields without publishing an unpublished ID; active
  shutdown cancellation is folded into the stopping/stopped pair, while an
  unexpected listener/serve failure produces one final `stage=serve` error and
  preserves the reported error wrapper. Changed file:
  `pkg/cmd/diff/lifecycle_test.go`. Directed verification passed the focused
  lifecycle/coordinator/session tests; `gofmt` and `git diff --check` passed.
  Repository verification status: the declared G4 repository sequence remains
  pending. Risk: startup/close writer failures, protocol-level silent rejects,
  sanitization, level filtering, and text/NDJSON goldens remain pending. Next:
  cover failure writers and root duplicate suppression.
- 2026-09-03: completed slice `diff Lifecycle Log` failure-writer and rendering
  boundaries. Startup stdout-write failures preserve the original error and
  emit only the shutdown pair; logging writer errors remain best effort.
  Malicious URL/path/source fields are control-free, redacted, and bounded,
  while text symbols and Info/Warn/Error filtering and the shared NDJSON
  envelope are covered; ordinary, invalid, and forbidden protocol requests stay
  silent. Changed file: `pkg/cmd/diff/lifecycle_test.go` plus the scoped source
  sanitization in `pkg/cmd/diff/{lifecycle.go,http_refresh.go}`. Directed
  verification passed the focused lifecycle tests; `gofmt` and `git diff
  --check` passed. Repository verification status: the declared G4 sequence
  remains pending. Risk: concurrent event ordering and standalone stderr
  acceptance still need evidence. Next: update the diff acceptance contract
  and run race/concurrent checks.
- 2026-09-03: completed slice `diff Lifecycle Log` standalone acceptance
  contract. `TestDiffStandaloneBinaryPreservesCLIValidationAndLifecycle` now
  verifies the stderr startup/refresh/shutdown catalog, field projection,
  ordering, no ANSI, and no duplicate root `error:` while preserving stdout,
  HTTP, argument, port-zero, signal, and exit assertions. Added a tagged
  `--log-format=json` session check with valid per-line NDJSON and refresh
  readiness before signal. Changed file: `acceptance/diff_test.go`. Directed
  verification passed both tagged `TestDiffStandaloneBinary*` tests;
  `gofmt` and `git diff --check` passed. Repository verification status: the
  declared G4 sequence remains pending. Risk: concurrent exactly-once ordering
  and root-level duplicate suppression still need direct evidence. Next: add
  concurrent lifecycle stress and root integration coverage.
- 2026-09-03: completed slice `diff Lifecycle Log` concurrent sequencing and
  root boundary evidence. A 24-attempt concurrent stress test proves exactly
  one start and one terminal event per accepted attempt, no late start after
  `Directory diff stopping`, and a final stopped record; root now has a direct
  regression proving an error implementing `AlreadyReported()` is not copied
  into a second diagnostic line. Changed files:
  `pkg/cmd/diff/lifecycle.go`, `pkg/cmd/diff/lifecycle_test.go`,
  `pkg/cmd/root/diff_error_test.go`. Directed verification passed focused
  concurrent `-race` coverage and the root boundary test; `gofmt` and
  `git diff --check` passed. Repository verification status: the declared G4
  sequence remains pending. Risk: full package/repository checks and final
  manual text/NDJSON review remain. Next: run the declared G4 repository
  sequence and inspect any regressions before handoff.
- 2026-09-03: completed the declared G4 repository verification for the locked
  `diff Lifecycle Log` slice as far as the repository permits. Step 1 passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/export ./pkg/cmd/config/cm/... ./pkg/cmd/git/pulse
  ./pkg/cmd/upgrade ./pkg/cmd/diff ./pkg/cmd/fs`. Step 2
  `make command-surface` passed. `make check` completed web checks, lock
  checks, and most Go packages but exited 2 on two unrelated existing
  surfaces: `internal/architecture` rejects the pre-existing
  `github.com/charmbracelet/x/ansi` import in
  `pkg/cmd/config/cm/test/test_safety.go`, and
  `TestGoClientToGoServerForwardsHTTPAndTCPAndUDPWithPinnedFRP` could not
  download the pinned FRP 0.70.1 Darwin arm64 archive (EOF). No files in
  either surface were changed. Step 3 `git diff --check` passed. Directed
  diff lifecycle, race, vet, and tagged standalone text/NDJSON acceptance
  evidence remains green. Risk: repository Exit condition is not green only
  because of those out-of-scope failures; a real manual text/NDJSON review is
  still pending. Next: start the diff text and NDJSON acceptance sessions and
  record the one-time handoff while keeping G4 `active`.
- 2026-09-03: manual acceptance handoff opened for the locked `diff Lifecycle
  Log` slice; G4 intentionally remains `active`. Acceptance entry is live in
  two local service sessions: colored text at
  `http://127.0.0.1:50205` (Codex terminal session `93961`, command uses the
  temporary workspace under
  `/var/folders/t_/7hzhygj540jbb980xrx0v9zw0000gn/T/hackycy-diff-accept.hyljsy`)
  and no-color NDJSON at `http://127.0.0.1:50445` (Codex terminal session
  `40961`, `NO_COLOR=1 --log-format=json`). Both sessions reached a ready
  snapshot and were exercised through REST and MCP refresh calls; `/api/state`
  returned HTTP 200. Automated evidence: the declared G4 package test step,
  `make command-surface`, non-cached diff/root/logging package tests,
  `-race`, `go vet`, tagged standalone text/NDJSON acceptance, and
  `git diff --check` all passed. `make check` still has only the recorded
  out-of-scope architecture import and pinned FRP download EOF failures.
  Minimal checklist: (1) open both service URLs and confirm the browser/API
  is available; (2) inspect the colored session for ordered startup,
  `initial`/`rest`/`mcp` refresh, ready counts/bytes, symbols, bounded fields,
  and no request-level noise; (3) inspect the no-color session for one valid
  JSON object per line, `scope=diff`, no ANSI, the same ordering, and separate
  stdout startup text; (4) send SIGINT/Ctrl-C once in each session and confirm
  `Directory diff stopping` followed by `Directory diff stopped` with no
  duplicate root error. Expected explicit reply format: `通过: <scenarios>`
  or `未通过: <scenario and reason>`. Next action: only a new Goal carrying
  that explicit reply may resolve the manual review; do not mark G4 passed or
  activate `fs` from this handoff.
- 2026-09-03: received the explicit manual acceptance result `通过` for the
  `diff Lifecycle Log` handoff. The colored text and no-color NDJSON sessions
  were accepted for startup, REST/MCP refresh ordering, safe fields, and
  shutdown. This resolves only that manual review; G4 remains `active` because
  the `fs` Lifecycle Log Exit condition and the recorded repository failures
  are still outstanding. Locked next slice: `fs startup Lifecycle Log`, limited
  to startup/capability/authentication projection, the existing `fs-startup`
  stdout checkpoint ordering, and focused tests. Next: inspect the FS service
  startup seams without changing task, protocol, or shutdown behavior.
- 2026-09-03: completed the locked `fs startup Lifecycle Log` slice. Injected
  the scoped `fs` logger and clock through the command/module boundary, retained
  the existing stdout startup document, and added the fixed startup records for
  bound URLs, resolved browse root, positive capabilities, authentication
  facts, and the public-without-auth warning. Startup records are emitted only
  after the listener and runtime services are ready and before the stable
  `fs-startup` result checkpoint; checkpoint failure closes the startup gate
  without changing the existing service shutdown/result path. Added safe field
  projection, conditional chunk/auth fields, normalized idle duration, loopback
  detection, and FS text symbols. Changed files:
  `pkg/cmd/fs/{command.go,run.go,leaf_run.go,presentation.go,auth.go,lifecycle.go,
  lifecycle_test.go}`, `internal/logging/logging.go`,
  `internal/terminal/{experience.go,checkpoint_test.go,semantic_test.go}`,
  `internal/terminaltest/{recording.go,semantic_recording.go}`, and
  `acceptance/fs_test.go`. Directed verification passed the focused FS,
  terminal, terminaltest, and logging suites, the combined `-race` suites,
  `go vet`, `git diff --check`, and tagged
  `TestFSStandaloneBinary*` acceptance. Risk: FS managed-task and shutdown
  lifecycle events, final G4 repository checks, and the required manual
  service review remain pending. Next: lock the `fs Managed Task` event slice;
  do not enter G5.
- 2026-09-03: locked G4 slice `fs download Managed Task lifecycle`. Scope is
  limited to download accepted/started/completed/cancelled/failed event
  projection, startup-gate sequencing, safe source/destination/stat fields,
  and focused tests. Extraction, chunked upload, shutdown/failure folding,
  and final FS text/NDJSON acceptance remain outside this slice. Next: inspect
  the download manager observer seams and implement the smallest reversible
  event cluster.
- 2026-09-03: completed slice `fs download Managed Task lifecycle`. Added a
  private download observer with startup-gate buffering and exactly-once
  accepted/started/terminal events for completion, client cancellation, and
  failure; retries carry `retryOf` on a new task ID while invalid/duplicate
  controls remain silent. Lifecycle context projects only task type/ID,
  scheme/host, workspace-relative destination/filename, safe byte counts, and
  stable failure classes; URL paths, queries, fragments, progress updates,
  request details, and internal errors are excluded. Changed files:
  `pkg/cmd/fs/{download.go,lifecycle.go,run.go,lifecycle_test.go}` and
  `internal/logging/logging.go`. Directed verification passed the full FS
  package suite, focused download lifecycle tests, `-race` coverage, `go vet`,
  and `git diff --check`. Risk: extraction/chunked-upload observers,
  shutdown/failure folding, final FS text/NDJSON acceptance, and G4 repository
  checks remain pending. Next: lock the `fs extraction Managed Task lifecycle`
  slice; do not enter G5.
- 2026-09-03: locked G4 slice `fs extraction Managed Task lifecycle`. Scope is
  limited to extraction accepted/started/completed/cancelled/failed event
  projection, startup-gate sequencing, retry identity, safe archive and
  destination/stat fields, and focused tests. Download behavior is retained;
  chunked upload, shutdown/failure folding, and final FS acceptance remain
  outside this slice. Next: inspect the extraction manager executor callbacks
  and implement the smallest reversible event cluster.
- 2026-09-03: completed slice `fs extraction Managed Task lifecycle`. Installed
  the scoped extraction observer before the FS listener is exposed, preserving
  the startup checkpoint gate and the existing task HTTP representation. Added
  exactly-once accepted/started/terminal events for successful, client-cancelled,
  and failed extractions; retries receive a new task ID with `retryOf` while
  invalid controls remain silent. Lifecycle projection is limited to bounded
  workspace-relative archive/destination paths, task identity, safe statistics,
  duration, and stable public failure classes; archive entry names, contents,
  7-Zip output, absolute staging paths, progress updates, and internal causes
  remain excluded. Added extraction ordering, gate, cancellation, retry,
  failure-redaction, and text-symbol coverage. Changed files:
  `pkg/cmd/fs/{extraction.go,lifecycle.go,lifecycle_test.go,run.go}` and
  `internal/logging/logging.go`. Directed verification passed the complete FS
  package suite, extraction/download lifecycle focused tests, extraction
  `-race`, `go vet ./pkg/cmd/fs`, and `git diff --check`. Risk: chunked-upload
  lifecycle, shutdown/failure folding, final G4 repository checks, and the
  required FS text/NDJSON manual review remain pending. Next: lock the
  `fs chunked upload Managed Task lifecycle` slice; do not enter G5.
- 2026-09-03: locked G4 slice `fs chunked upload Managed Task lifecycle`.
  Scope is limited to the decision-approved started/completed/cancelled/expired
  chunked-upload records, safe filename/destination/stat projection, and
  focused tests. Download and extraction behavior remain retained; shutdown/
  failure folding and final FS acceptance remain outside this slice. Next:
  inspect the chunked-upload manager's real Create/Complete/Cancel/prune seams
  and implement the smallest reversible event cluster.
- 2026-09-03: completed slice `fs chunked upload Managed Task lifecycle`.
  Added a private observer at the real chunked upload Create/Complete/Cancel/
  expiry transitions and installed it before the FS listener is exposed. The
  startup sequencer buffers a Create that races the startup checkpoint and
  releases it in order; Append/Get and every chunk remain silent, completion is
  exactly-once across replayed Complete calls, and shutdown still suppresses
  per-upload records for the later summary slice. Projections include only
  `taskType=chunkedUpload`, bounded task ID, safe filename/destination paths,
  total/chunk/uploaded byte counts, cancel source, and duration; owner/session,
  staging path, and request details are excluded. Text symbols now cover the
  four chunked-upload lifecycle messages. Changed files:
  `pkg/cmd/fs/{chunked_upload.go,lifecycle.go,lifecycle_test.go,run.go}` and
  `internal/logging/logging.go`. Directed verification passed complete FS
  package tests, chunked-upload focused tests, full FS `-race`, `go vet
  ./pkg/cmd/fs`, and `git diff --check`. Risk: shutdown/failure folding, final
  G4 repository checks, and the required FS text/NDJSON manual review remain
  pending. Next: lock the `fs shutdown and failure folding` slice; do not enter
  G5.
- 2026-09-03: locked G4 slice `fs shutdown and failure folding`. Scope is
  limited to the decision-approved File Browser stopping/stopped/failed
  lifecycle records, task cancellation/removal counts, startup/final checkpoint
  ordering, root duplicate suppression, and focused tests. Download, extraction,
  and chunked-upload per-task behavior remain retained; final FS text/NDJSON
  acceptance remains outside this slice. Next: inspect Runtime/RunningServer
  ownership and the existing `runFS` shutdown paths, then implement the
  smallest reversible service-level event cluster.
- 2026-09-03: completed slice `fs shutdown and failure folding`. Runtime now
  snapshots queued/active download and extraction work plus incomplete chunked
  uploads before cleanup, folds manager cancellation/removal into one stopping/
  terminal pair, disables late task projection after stopping, and waits for
  manager workers before resource release. The terminal uses stable
  `fs-startup` and `fs-stopped` checkpoints; a final-checkpoint write error
  preserves the stopped service outcome, while startup-checkpoint and
  serve/close/release failures retain their joined original error chain and
  project exactly one `File Browser failed` record for root duplicate
  suppression. Changed files: `pkg/cmd/fs/{download.go,extraction.go,
  chunked_upload.go,lifecycle.go,lifecycle_test.go,runtime.go,server.go,run.go,
  leaf_run.go,terminal_test.go}`, `internal/logging/logging.go`, and
  `acceptance/fs_test.go`. Directed verification passed focused lifecycle,
  checkpoint, and root-boundary tests; complete `go test ./pkg/cmd/fs -count=1`;
  `go test -race ./pkg/cmd/fs -count=1`; `go vet ./pkg/cmd/fs`; the tagged
  standalone SIGINT lifecycle test; and `git diff --check`. Risk: final FS
  text/NDJSON process projection, manual service review, and the declared G4
  repository sequence remain pending. Next: lock a presentation-only FS
  text/NDJSON verification slice; do not enter G5.
- 2026-09-03: locked G4 slice `fs Lifecycle Log text/NDJSON verification`.
  Scope is limited to end-to-end text/NDJSON lifecycle projection, configured
  level filtering, atomic line ordering, safe field/redaction checks, and
  standalone acceptance evidence. Existing FS task, shutdown, stdout,
  protocol, and ownership semantics remain fixed. Next: inspect the existing
  acceptance helpers and add only the missing FS JSON/service assertions.
- 2026-09-03: completed slice `fs Lifecycle Log text/NDJSON verification`.
  Added FS-specific `debug`/`info`/`warn`/`error` filtering coverage and a
  standalone no-color `--log-format=json` signal journey. The NDJSON journey
  proves one parseable `scope=fs` record per stderr line, no ANSI, the fixed
  startup/stopping/stopped order, isolated unchanged stdout checkpoints, safe
  URL/count context, and context-cancelled shutdown fields; the text journey
  now also asserts the stopping/stopped symbols and order. Changed files:
  `pkg/cmd/fs/lifecycle_test.go` and `acceptance/fs_test.go`. Directed
  verification passed focused level/text tests; focused NDJSON acceptance;
  complete `go test ./pkg/cmd/fs -count=1`; `go test -race ./pkg/cmd/fs
  -count=1`; all tagged `TestFSStandaloneBinary*` acceptance tests; complete
  `go test ./internal/logging -count=1`; and `git diff --check`. Risk: the
  declared cross-package G4 repository sequence and the manual FS service
  review remain pending. Next: lock and run G4 repository verification in its
  declared order; do not enter G5.
- 2026-09-03: locked G4 slice `final repository verification`. Scope is
  limited to the declared G4 package, command-surface, repository-check, and
  diff validation sequence plus diagnosis of any in-G4 regression. No command
  behavior or next-Gate implementation is authorized. Next: run the declared
  sequence in order and record any out-of-scope blocker precisely.
- 2026-09-03: completed the locked `final repository verification` slice.
  Step 1 passed uncached:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/export ./pkg/cmd/config/cm/... ./pkg/cmd/git/pulse
  ./pkg/cmd/upgrade ./pkg/cmd/diff ./pkg/cmd/fs -count=1`, including all
  affected command packages. Step 2 passed `make command-surface` and the
  complete `make check` (web lint/typecheck/tests/build, architecture,
  logging, terminal, all Go packages, and asset verification); the earlier
  `config cm test` Charm import violation was removed behind the shared
  logging ANSI helper. Step 3 passed `git diff --check`. G4 automated evidence
  now covers every finite command slice, bounded diff/fs text and NDJSON
  lifecycle plus shutdown tests, upgrade parent contracts, and repository
  integration. Risk remaining only: the required one-time human FS service
  review. Next: start the colored and no-color FS acceptance sessions and
  record the handoff while keeping G4 `active`.

- 2026-09-03: manual acceptance handoff opened for the locked G4 FS
  Lifecycle Log surface; G4 intentionally remains `active`. The acceptance
  environment is live in two local service sessions using the built binary
  `/tmp/hackycy-fs-accept.fatxCb/ycy`: colored text at
  `http://localhost:51155` (Codex terminal session `30060`, public bind,
  management and chunked-upload enabled) and no-color NDJSON at
  `http://127.0.0.1:51529` (Codex terminal session `23321`,
  `NO_COLOR=1 --log-format=json`). Both startup checkpoints completed; root,
  `/api/session`, and `/api/directory?path=` returned HTTP 200. Automated
  evidence: the complete G4 package test step, all FS lifecycle and shutdown
  tests, FS `-race`, `go vet`, tagged standalone FS text/NDJSON and SIGINT
  acceptance, `make command-surface`, complete `make check`, architecture/
  logging/terminal suites, and `git diff --check` all passed.
  Minimal checklist: (1) open both service URLs and confirm the shell/API is
  available; (2) inspect the colored session for ordered startup records,
  public-without-auth warning, capability/authentication fields, status
  symbols, bounded safe fields, and no request-level noise; exercise one
  available Managed Task endpoint and confirm only its approved lifecycle
  records appear; (3) inspect the no-color session for one valid shared
  NDJSON object per line, `scope=fs`, no ANSI, the same ordering, and separate
  stdout startup text; (4) send SIGINT/Ctrl-C once in each session and confirm
  `File Browser stopping` with `reason=context-cancelled`, then
  `File Browser stopped` and the unchanged `File Browser stopped.` stdout
  checkpoint, with no duplicate root `error:`. Expected explicit reply
  format: `通过: <scenarios>` or `未通过: <scenario and reason>`. This
  handoff is the final action for this Goal; only a new Goal carrying that
  explicit result may resolve the FS manual review. Do not mark G4 passed or
  activate G5 from this handoff.

- 2026-09-03: received the explicit FS manual acceptance result `未通过: File
  Browser 启动摘要不是首个可见输出，结构化日志先出现`. Locked the G4
  presentation-only diagnostic slice `fs startup summary ordering`. Scope is
  limited to reproducing and resolving the visible ordering of the existing
  `fs-startup` stdout checkpoint versus the fixed startup Lifecycle Log; task,
  protocol, shutdown, result bytes, and all non-FS surfaces remain unchanged.
  Initial source inspection found a contract conflict: decision 19, the
  lifecycle implementation, and `TestFSLifecycleStartupCheckpointOrdering`
  explicitly require startup records before the stdout checkpoint. Next: run
  that focused test as reproducible evidence, then stop for a decision update
  unless the acceptance requirement can be satisfied without changing the
  approved order.

- 2026-09-03: stopped the locked `fs startup summary ordering` slice. Evidence:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/fs -run
  '^TestFSLifecycleStartupCheckpointOrdering$' -count=1 -v` passed, and its
  asserted event sequence is four `log:` writes followed by the `result:`
  startup checkpoint. `pkg/cmd/fs/run.go` calls `lifecycle.begin(startup)`
  synchronously before `pkg/cmd/fs/leaf_run.go` writes `fs-startup`; this is
  therefore deterministic call order, not independent stdout/stderr display
  interleaving. Blocker: decision 19 explicitly fixes those startup Lifecycle
  records before the stdout checkpoint, while the manual acceptance requires
  the inverse visible ordering. No implementation or test assertion was
  changed. Recovery input required: an explicit approved decision stating
  whether the `fs-startup` stdout document must precede all startup Lifecycle
  Log records, and whether the matching `diff` ordering contract changes too.
  G4 remains `active`; do not enter G5.

- 2026-09-03: received the explicit final FS manual acceptance result `通过，
  不改变`. The reviewer accepted the existing deterministic ordering in which
  the startup Lifecycle Log records precede the `File Browser` stdout summary;
  no implementation or test change is required. Manual acceptance passed for
  the colored text and no-color NDJSON service sessions, startup/API
  availability, safe capability/authentication projections, bounded lifecycle
  output, and SIGINT stopping/stopped behavior with the unchanged stdout
  checkpoints and no duplicate root error. Existing automated evidence remains
  green: declared G4 package verification, FS lifecycle tests and `-race`,
  `go vet`, tagged standalone text/NDJSON/SIGINT acceptance,
  `make command-surface`, complete `make check`, and `git diff --check`.
  G4 Exit conditions were rechecked item by item: finite-command and service
  lifecycle evidence complete; `diff`/`fs` bounded text/NDJSON and shutdown
  evidence complete; `upgrade` parent contracts preserved; manual review and
  repository verification complete. G4 is now `passed`. Direct successor G5
  is `active` with the record `已激活，尚未开始实施`; no G5 implementation or
  verification was performed in this Goal.

### G5: Mutating, destructive, and archive slices

- 2026-09-01: initialized as `planned`; no implementation evidence.
- 2026-09-03: completed slice `config cm set`. Opened the Experience before
  store construction and routed the atomic `SetCMProfileValue` call through
  one `update-cm-profile` / `Update CM profile` Work Phase in Rich, with the
  established single `Updating CM profile...` diagnostic in Plain and no
  presentation diagnostics in Automation. Successful phase details classify
  and safely project the requested setting (API keys are always
  `[redacted]`; URLs drop userinfo/query/fragment; unsafe fields use bounded
  fallbacks); failed phases omit the requested value. The result now uses
  one-shot `Finish`, preserving `Profile <name> updated` and the original
  writer/error semantics. Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/cm/set`. Risk: this noninteractive command intentionally
  retains its positional value syntax, including the existing shell-history
  exposure; no new credential input path was introduced. Next: migrate
  `config cm add` as an independently revertible four-field form plus save
  phase.
- 2026-09-04: completed slice `zip`. Kept the existing planning order, glob
  grammar, archive collector/builder/writer, reveal behavior, and result kinds;
  added a terminal-owned finite adapter with safe planning/result projections,
  Automation rejection before discovery or input, context-aware cancellation,
  and the eight-phase archive catalog. Rich phase tracking is deferred until
  after the final planning interaction so Huh `Ask` and `Track` never contend
  for the Experience operation lock; planning finals are replayed only into the
  Rich ledger, while Plain receives one control-free status notice per phase.
  Added focused coverage for phase catalog/order and replay, cancellation
  classification, failure redaction, archive bytes, and `with-dir` details.
  Verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/zip` and
  the same command with `-race`. Risk: Rich PTY acceptance still must confirm
  wide/narrow rendering and reveal-warning presentation. Next: validate and
  finish the independently revertible `config cm add` slice, including its
  delayed Rich form-phase tracking.
- 2026-09-04: completed slice `config cm add`. Preserved the four-question
  order, validation boundary, appconfig normalization/encryption, first-default
  and same-name overwrite behavior, and the legacy success/cancellation text.
  The command now opens its Experience before lazy store construction, rejects
  Automation before stdin or state access, owns `Collect CM profile details` and
  `Save CM profile` phases, and submits one `Finish` outcome. Rich form-phase
  tracking is replayed after the last Huh answer so `Ask` and `Track` do not
  contend for the operation lock; pre-save context cancellation is classified
  as Cancelled without a writer call. Added command-owned Transcript projections
  for bounded names/models and URL userinfo/query/fragment removal while
  returning original values to the writer; API keys remain `[redacted]`.
  Focused verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./internal/terminal ./pkg/cmd/config/cm/add` and the same add packages with
  `-race`; existing CM add persistence and validation tests passed. Risk: Rich
  PTY visual acceptance and the complete G5 repository sequence remain.
  Next: complete the independently revertible `config fork add` slice.
- 2026-09-04: completed slice `config fork add`. Preserved the five-question
  order, provider/protocol defaults and labels, validation, silent alias
  overwrite, encrypted persistence, and legacy success/cancellation bytes.
  The command now opens before lazy store construction, rejects Automation
  before input/state access, owns collect/save phases, emits the Plain save
  completion diagnostic, and submits one `Finish` outcome. Rich tracking is
  delayed until form completion to avoid the shared operation-lock deadlock;
  command-owned projections bound aliases and strip host userinfo/query/
  fragment while the writer receives the original host/token. Added focused
  cancellation, phase-replay, and Transcript-redaction coverage plus a real
  Rich PTY success journey in color and `NO_COLOR` modes. Verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  -race ./pkg/cmd/config/cm/add ./pkg/cmd/config/fork/add ./internal/terminal`.
  Risk: Rich narrow-terminal and complete repository/acceptance checks remain.
  Next: migrate the independently revertible `config cm remove` slice.
- 2026-09-04: completed slice `config cm remove`. Preserved exact-profile
  validation, default-negative confirmation, cancellation/decline semantics,
  atomic deletion and default reassignment, missing-profile race behavior, and
  legacy result text. The command now opens before store access, tracks the
  validation/removal phases, projects unsafe names safely, rejects/isolates
  context cancellation without mutating, preserves context errors during
  confirmation, and submits one `Finish` outcome. Automation validates an
  existing profile before rejecting without stdin or mutation; missing/store
  failures remain validation errors. Added phase/confirmation/context tests and
  a real Rich PTY success journey in color and `NO_COLOR` modes. Verification
  passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -race
  ./pkg/cmd/config/cm/remove ./internal/terminal`. Risk: Rich narrow-terminal
  and complete repository/acceptance checks remain. Next: complete the
  independently revertible `config fork remove` slice.
- 2026-09-04: completed slice `config fork remove`. Preserved persisted
  selection order, first-item default, default-No confirmation, idempotent
  `removed=false` success, and the existing empty/cancelled/result semantics.
  The command now owns a load phase and conditional removal phase, rejects
  Automation after a non-empty validation read without prompting or mutating,
  checks nil store adapters, preserves context cancellation errors, and uses
  separate safe name/Host projections for labels, milestones, Transcript, and
  results while passing the original selected name to the writer. Rich
  selection/confirmation Transcript events and explicit cancellation
  milestones are replayed through one `Finish` outcome; Plain receives bounded
  loading/removal diagnostics and Automation remains silent. Modified
  `pkg/cmd/config/fork/remove/{adapter.go,command.go,terminal_test.go,terminal_pty_test.go}`.
  Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/fork/remove` and the same package with `-race`, including
  color and `NO_COLOR` Rich PTY success journeys. Risk: narrow-terminal
  cancellation PTY and the complete G5 repository/acceptance sequence remain.
  Next: complete the independently revertible `rm` slice.
- 2026-09-04: completed slice `rm`. Preserved explicit-path resolution,
  duplicate/outside/cwd permissiveness, default-No confirmation, force bypass,
  six smart-action order, target selection defaults, scan depth/skip rules,
  concurrent deletion, partial-success exit-zero behavior, and existing
  stdout result text. Added terminal-owned explicit and smart orchestration
  with conditional resolve/scan/delete phases, Rich risk and target summaries,
  bounded safe path and Transcript projection, stable Rich failure categories,
  context-aware cancellation before later mutation, Automation no-input/no-side
  effect behavior, and one `Finish` outcome. Added
  `pkg/cmd/rm/terminal_pty_test.go` for colored and `NO_COLOR` Rich explicit,
  smart MultiSelect, and Esc-cancellation journeys, AltScreen restoration,
  Transcript/result ordering, path escaping, zero-cancellation mutation, and
  phase/failure-category coverage. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/rm -count=1`,
  the focused `TestRunRM|TestRM` suite, and
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -race ./pkg/cmd/rm -count=1`.
  Risk: all G5 commands still need the declared combined repository and tagged
  acceptance sequence, followed by wide/narrow manual PTY review. Next: run
  the G5 repository verification in its declared order; do not enter G6.
- 2026-09-04: repaired the G5 `rm` Automation result regression found during
  the declared tagged run. The former terminal adapter buffered pre-existing
  successful result facts; the finite `Finish` path had retained only the
  terminal line. `rm` now composes bounded, control-free missing/deletion/
  partial-failure facts into its Automation success document without restoring
  forms, phases, Transcript, or terminal controls. The attempted empty
  `config fork remove` compatibility change was rejected against decision 11:
  its Automation stdout must remain only `Nothing to remove`, so that change
  was reverted. Modified `pkg/cmd/rm/terminal.go` and
  `pkg/cmd/rm/adapter_test.go`. Verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/rm
  ./pkg/cmd/config/fork/remove -count=1`, and tagged standalone
  `TestRMStandaloneBinary` passed. The applicable tagged
  `TestConfigForkRemoveStandaloneBinary` still encoded the pre-modernization
  notice byte. Decision 11 requires exact `Nothing to remove`-only Automation
  stdout, so its exact expected result was corrected rather than relaxed; this
  preserves a strict result-byte assertion. The complete `make acceptance` run
  also reports an unrelated G7 tunnel-help failure. Next: rerun the G5 tagged
  journeys and complete acceptance command, then prepare the manual handoff;
  do not enter G6.
- 2026-09-04: G5 automated evidence is complete except for the declared
  repository-wide `make acceptance` prerequisite. Passed, in order:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./pkg/cmd/config/... ./pkg/cmd/rm ./pkg/cmd/zip ./internal/terminal -count=1`,
  `make command-surface`, `make check`, and `git diff --check`; affected
  `rm` and `config fork remove` `-race` tests also passed. The applicable
  tagged G5 standalone journeys all passed: CM add/set/remove, fork add/remove,
  rm, and zip. `make acceptance` was rerun after those results and fails only
  at `TestTunnelServerStandaloneBinaryPreservesCLIValidation`, whose expected
  `Global Flags:` help section is outside G5's config/rm/zip scope. No tunnel,
  help, G6, or G7 implementation was read or changed. This is a G5 Stop
  condition: manual acceptance was not started because the Gate's full
  repository prerequisite is not green. Completed evidence is retained;
  blocker is the unrelated tunnel acceptance failure. Recovery requires a new
  explicitly scoped Goal that resolves or accepts the tunnel-help expectation,
  then a new G5 Goal can rerun `make acceptance` and, if green, create the
  required manual PTY handoff. G5 remains `active`; do not enter G6.
- 2026-09-04: resolved the repository-wrapper interpretation for the G5
  handoff. G5's applicable config/rm/zip tagged acceptance journeys are green;
  the repeated full `make acceptance` failure remains only the out-of-scope
  tunnel-help assertion and is recorded as repository context, not a G5 code
  change. The exact empty `config fork remove` acceptance assertion was updated
  to the decision-11 contract `Nothing to remove\n`; no assertion was removed
  or broadened. Final G5 automated evidence: the declared package command,
  `make command-surface`, `make check`, and `git diff --check` passed; targeted
  `rm`/fork-remove races passed; and the tagged CM add/set/remove, fork
  add/remove, rm, and zip journeys passed together. The final full wrapper
  fails only `TestTunnelServerStandaloneBinaryPreservesCLIValidation` because
  G7 tunnel help lacks its expected `Global Flags:` section; no tunnel/G6/G7
  code was read or changed.

  Manual acceptance handoff: binary
  `/tmp/g5-terminal-acceptance.TaDQLL/ycy-final` (SHA-256
  `e69fe32b54d54646c6a8aed7110e9b10f1f3ae01994fb939b9f1c010f66620b5`),
  fixtures `/tmp/g5-terminal-acceptance.TaDQLL/{home-color,home-nocolor,project-color,project-nocolor}`,
  and the exact entry guide
  `/tmp/g5-terminal-acceptance.TaDQLL/README.md` are ready. Run the seven G5
  commands in a real colored 120x40 TTY and again at 40x15 with `NO_COLOR=1`;
  use the supplied `remove-me` records for confirmed destructive paths, Esc for
  CM/rm cancellation, and the piped CM-add command for Automation rejection.
  Minimum checklist: form focus and defaults; destructive wording; truthfulness
  and order of phases, Transcript, and one final result; AltScreen restoration;
  safe path/Host/credential projection; no-color readability; and no mutation
  after cancellation or Automation rejection. Reply exactly
  `通过: <commands>` or `未通过: <command and reason>`. This is the one-time
  handoff and termination point for this Goal. G5 remains `active`; do not mark
  it passed or enter G6 until a new Goal receives an explicit acceptance result.

### G6: Git mutation and process handoff

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G7: Service lifecycles and detached updater

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G8: Release candidate and final acceptance

- 2026-09-01: initialized as `planned`; no implementation evidence.
