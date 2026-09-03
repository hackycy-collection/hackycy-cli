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
| G4: External-read and service-startup slices | active | G3 | `implementation-plan.md` -> `G4: External-read and service-startup slices` | G3 passed |
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

### G5: Mutating, destructive, and archive slices

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G6: Git mutation and process handoff

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G7: Service lifecycles and detached updater

- 2026-09-01: initialized as `planned`; no implementation evidence.

### G8: Release candidate and final acceptance

- 2026-09-01: initialized as `planned`; no implementation evidence.
