# CLI Structure Refactor Goal Runbook

## Source Baseline

| Path | SHA-256 |
| --- | --- |
| `.scratch/cli-structure-refactor/implementation-plan.md` | `7b419363d958d94fd06037bc2c2ecd45765b82bc2d208d9f3ae72d8e831561d3` |
| `.scratch/cli-structure-refactor/issues/01-research-github-cli-project-structure.md` | `13dd0dfa348e1cf8fa2fbd23f5c491650727eef4a3da9f3501af8f1e37fbe5dc` |
| `.scratch/cli-structure-refactor/issues/02-research-ycy-current-tree-and-ownership.md` | `c9ede864686f76aeed80efcc9ddc577cb72c650aa0abab5b6d13b2748fccfd99` |
| `.scratch/cli-structure-refactor/issues/03-choose-command-package-visibility-and-composition-root.md` | `c6992465f8298374fbe955f6cfd5fd25158cb551f33eb2c53449664bb0eb681d` |
| `.scratch/cli-structure-refactor/issues/04-choose-domain-command-tree-and-file-names.md` | `5802669121362475a09cb767a46c05a8d48e7ebde36b56a3c7ff652810a10be7` |
| `.scratch/cli-structure-refactor/issues/05-choose-shared-module-seams-and-dependency-direction.md` | `c7a4e8e19a2d42bfa90ea5d838b21cbbe1afdf125089bed250d99e333f0ffb68` |
| `.scratch/cli-structure-refactor/issues/06-choose-repository-support-area-layout.md` | `474fe4be6c75559c5e3e1493e1eb6ea0ababfe7e4b2d9484aad0550ab17c96db` |
| `.scratch/cli-structure-refactor/issues/07-choose-test-fixture-and-acceptance-topology.md` | `d8141f7b0e4824a67bf81380f251e5d2ceedd4e060bd4e8568842f200e6dabdf` |
| `.scratch/cli-structure-refactor/issues/08-prototype-target-tree-and-migration-slices.md` | `d8b512a896912123fe9019e4c0e0830d33002748e76780ac83bc6b25821f56d7` |
| `.scratch/cli-structure-refactor/issues/09-approve-structural-migration-gates.md` | `600447f7e1352eba527c1fc123b12eb9a865bd8378d13340f1ab9a9bfe45fb90` |
| `.scratch/cli-structure-refactor/map.md` | `d76a9ca55126c5fa47ca74561f023788a61f513287e3294c6992bdbcb05c7e9a` |
| `.scratch/cli-structure-refactor/prototypes/08-target-tree-and-migration-slices.md` | `e082655c37bed1a6219dcc8a5729559d7097fee5907b4ecbd6fb4203c5a8b9af` |
| `.scratch/cli-structure-refactor/research/02-ycy-current-tree-and-ownership.md` | `1f26f50ebbf81c349a2b106360c7edf20aa62eb74b7c431eb082ef99542b03c0` |
| `.scratch/cli-structure-refactor/research/github-cli-structure.md` | `c7c606d8f94572a79cb6083d0d76f59a47c4a4b09ad221a3fcd213c28e6b13a2` |
| `CLAUDE.md` | `35a0f66a6531e61b35dc7288884da35eeef79ad1039f5ffdfd29295cee8d566e` |

## State Rules

- `implementation-plan.md` 是 Gate 合同的唯一来源；本账本只记录状态和证据。
- 一次只执行 Goal Ledger 中唯一 `active` 的 Gate。
- 每轮向对应 Gate 的 Progress Log 追加 slice、修改、验证结果、风险和下一动作。
- `passed` 需要计划中每条 Exit condition 的明确证据；普通实现或验证失败保持 `active`。
- `blocked` 只用于计划声明的 Stop condition，并记录阻塞与恢复条件。
- 人工验收待确认是一次 Goal 终止交接，必须记录在 Progress Log；它不是 `blocked`，当前 Gate 必须保持 `active`，不得通过重复日志表示等待。
- 当前 Gate 通过后只激活直接后继并结束本次 Goal；直接后继虽为 `active`，但必须由新的 Goal 执行。最后一个 Gate 通过后记录 effort 完成并结束本次 Goal。

## Goal Ledger

| Gate | Status | Depends on | Plan contract | Unlock evidence |
| --- | --- | --- | --- | --- |
| G0: Stabilize pre-migration baseline | passed | none | `implementation-plan.md` -> `G0: Stabilize pre-migration baseline` | no predecessor |
| G1: Freeze surface and establish acceptance architecture | active | G0 | `implementation-plan.md` -> `G1: Freeze surface and establish acceptance architecture` | G0 E1-E4 recorded |
| G2: Lift root and introduce the bounded Factory | planned | G1 | `implementation-plan.md` -> `G2: Lift root and introduce the bounded Factory` | G1 pending |
| G3: Migrate simple command vertical slices | planned | G2 | `implementation-plan.md` -> `G3: Migrate simple command vertical slices` | G2 pending |
| G4: Migrate nested Config groups | planned | G3 | `implementation-plan.md` -> `G4: Migrate nested Config groups` | G3 pending |
| G5: Migrate Git leaves and shared process runtime | planned | G4 | `implementation-plan.md` -> `G5: Migrate Git leaves and shared process runtime` | G4 pending |
| G6: Migrate long-running and worker-backed commands | planned | G5 | `implementation-plan.md` -> `G6: Migrate long-running and worker-backed commands` | G5 pending |
| G7: Finalize the process composition root | planned | G6 | `implementation-plan.md` -> `G7: Finalize the process composition root` | G6 pending |
| G8: Publish and audit the final structure | planned | G7 | `implementation-plan.md` -> `G8: Publish and audit the final structure` | G7 pending |
| G9: Collect native and release evidence | planned | G8 | `implementation-plan.md` -> `G9: Collect native and release evidence` | G8 pending |

## Progress Log

### G0: Stabilize pre-migration baseline

- 2026-08-29: initialized as `active`; implementation has not started.
- 2026-08-29: Slice `baseline-record` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Directed verification before tests: rollback commit `1b0a4f8b931eec472083b4bdea66265adb929ba4`; `git status --short` was empty; host `Darwin arm64`; `go version go1.27.0 darwin/arm64`. Risk: the required test commands pin `GOTOOLCHAIN=go1.26.7`, which differs from the host default and must remain available without tracked-source changes. Next: run the first of three ordinary-suite repetitions and stop on any failure or worktree/commit change.
- 2026-08-29: Slice `ordinary-suite-1-of-3` stopped. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 1 before affected packages could compile: `web/assets.go` could not embed missing `web/dist`, and `internal/commands/fs/sevenzipruntime/payload_darwin_arm64.go` could not embed missing `payload/darwin-arm64/7zz`. This is a G0 Stop condition; the remaining two ordinary-suite runs and the focused 20-run Tunnel case were not run. Post-failure evidence: commit remained `1b0a4f8b931eec472083b4bdea66265adb929ba4`; `git status --short` contains only this runbook's evidence edits; `git diff --check` passed; both required paths were absent. Risk: a retry cannot satisfy G0 after this failure. Recovery input: a separate, explicitly authorized prerequisite-preparation task or clean environment must provide the Web distribution and Darwin arm64 7-Zip payload without tracked-source changes, then a new G0 Goal must restart all baseline evidence from the beginning.
- 2026-08-29: Recovery prerequisite preparation completed under explicit user authorization. `make PNPM='npx --yes pnpm@11.13.0' bootstrap` succeeded after bypassing the host's incompatible `pnpm 11.6.0` without changing the global tool; `make PNPM='npx --yes pnpm@11.13.0' check-web` passed lint, type checking, 33 tests, Vite build, and generated-asset verification; `make prepare-7zip` succeeded. Verification found 428 files under `web/dist` and a Darwin arm64 `payload/darwin-arm64/7zz` executable. `web/node_modules`, `web/dist`, the payload, `tools/lefthook/bin`, and `.tmp/7zip/downloads` are ignored; `git status --short` and `git diff --name-only` still show only this runbook. Risk: this restores prerequisites but does not erase the failed ordinary-suite evidence. Next: begin a newly authorized G0 baseline run from its directed-record slice; G0 remains `blocked` until that restart is explicitly initiated.
- 2026-08-29: User explicitly authorized a G0 restart after prerequisite preparation. G0 is reactivated; the prior failed ordinary-suite invocation remains invalid evidence and will not be counted or retried as part of the new repetition set. Scope remains G0 only. Next: record a fresh rollback commit/worktree/environment snapshot before running ordinary-suite repetition 1 of 3.
- 2026-08-29: Slice `baseline-record-restart` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Fresh directed verification before the new repetition set: rollback commit `1b0a4f8b931eec472083b4bdea66265adb929ba4`; `git status --short` and `git diff --name-only` showed only this runbook; host `Darwin arm64`; host `go version go1.27.0 darwin/arm64`; pinned toolchain `go version go1.26.7 darwin/arm64`. Generated Web and 7-Zip prerequisites remain ignored. Risk: any new test failure invalidates this restarted set. Next: run ordinary-suite repetition 1 of 3.
- 2026-08-29: Slice `ordinary-suite-1-of-3-restart` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 0; all listed packages passed, including `cmd/ycy` (97.063s) and `internal/commands/tunnel` (73.810s). Post-run directed verification: commit remained `1b0a4f8b931eec472083b4bdea66265adb929ba4`; tracked diff remains only this runbook; `git diff --check` passed. Risk: two independent ordinary-suite repetitions and the focused 20-run case remain. Next: run ordinary-suite repetition 2 of 3.
- 2026-08-29: Slice `ordinary-suite-2-of-3-restart` stopped. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 1 because `internal/commands/tunnel` failed `TestFRPSupervisorRejectsAnActivationExitAndKeepsConfigurationFailuresStopped`: `frp_supervisor_unix_test.go:103: condition did not become true within 1s`. This is the known Tunnel supervisor timeout described by the G0 decision and a Stop condition. The first restarted ordinary-suite repetition remains recorded but the repetition set is invalid; ordinary-suite repetition 3 and the focused 20-run case were not run. Post-failure directed verification: commit remained `1b0a4f8b931eec472083b4bdea66265adb929ba4`; tracked diff remains only this runbook; `git diff --check` passed; no production migration occurred. Risk: G0 cannot proceed on a retry or flaky-test allowance. Recovery input: stabilize this test or its underlying supervisor behavior in a separate task and commit, then start a new G0 run with all baseline evidence from the beginning.
- 2026-08-29: User-requested diagnosis (non-evidence; G0 remains `blocked`) completed. The failed assertion waits only 1s for a fixture shell to write its startup marker after `configured.Start`; `Start` treats the child as activated after its 20ms liveness window without guaranteeing the shell has executed that marker write. A diagnostic `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=20 ./internal/commands/tunnel -run '^TestFRPSupervisorRejectsAnActivationExitAndKeepsConfigurationFailuresStopped$'` exited 0 in 8.549s, confirming that the named test is not deterministically failing when isolated. Risk: this supports a load-sensitive test deadline rather than proving there is no runtime hang. Recommended separate stabilization task: retain the marker assertion and replace its 1s wait with the existing bounded `frpSupervisorFixtureStartupTimeout` (5s), add state/PID diagnostics on expiry, then prove the change with repeated focused and full-suite runs. If the 5s bound fails, diagnose the child process state before changing production behavior. Next: await the separate stabilization result before a fresh G0 restart.
- 2026-08-29: Separate pre-migration stabilization completed in local commit `cd0e7f179939a1235d92ec5f9f43c2d2f0ec0072` (`test(tunnel): stabilize supervisor fixture startup wait`). It changes only `internal/commands/tunnel/frp_supervisor_unix_test.go`: retains the startup-marker and configuration-failure assertions while using the existing bounded 5s fixture startup timeout. Verification before commit: the named test passed `-count=100` in 38.487s, and `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 0, including Tunnel in 73.751s; formatting and diff hygiene passed. G0 is reactivated against this new rollback commit; no production implementation changed. Next: record a fresh directed baseline snapshot, then run ordinary-suite repetition 1 of 3.
- 2026-08-29: Slice `baseline-record-after-stabilization` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Directed verification before the new repetition set: rollback commit `cd0e7f179939a1235d92ec5f9f43c2d2f0ec0072`; `git status --short` and `git diff --name-only` showed only this runbook; host `Darwin arm64`; host `go version go1.27.0 darwin/arm64`; pinned toolchain `go version go1.26.7 darwin/arm64`. Generated Web and 7-Zip prerequisites remain ignored. Risk: any ordinary-suite failure invalidates this new set. Next: run ordinary-suite repetition 1 of 3.
- 2026-08-29: Slice `ordinary-suite-1-of-3-after-stabilization` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 0; all listed packages passed, including `cmd/ycy` (79.679s) and `internal/commands/tunnel` (61.025s). Post-run directed verification: commit remained `cd0e7f179939a1235d92ec5f9f43c2d2f0ec0072`; tracked diff remains only this runbook; `git diff --check` passed. Risk: two independent ordinary-suite repetitions and the focused 20-run case remain. Next: run ordinary-suite repetition 2 of 3.
- 2026-08-29: Slice `ordinary-suite-2-of-3-after-stabilization` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 0; all listed packages passed, including `cmd/ycy` (96.074s) and `internal/commands/tunnel` (72.945s). Post-run directed verification: commit remained `cd0e7f179939a1235d92ec5f9f43c2d2f0ec0072`; tracked diff remains only this runbook; `git diff --check` passed. Risk: one independent ordinary-suite repetition and the focused 20-run case remain. Next: run ordinary-suite repetition 3 of 3.
- 2026-08-29: Slice `ordinary-suite-3-of-3-after-stabilization` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` exited 0; all listed packages passed, including `cmd/ycy` (79.607s) and `internal/commands/tunnel` (60.566s). Post-run directed verification: commit remained `cd0e7f179939a1235d92ec5f9f43c2d2f0ec0072`; tracked diff remains only this runbook; `git diff --check` passed. Risk: only the focused 20-run Tunnel case and final diff/no-production-migration audit remain. Next: run the named Tunnel supervisor test with `-count=20`.
- 2026-08-29: Slice `tunnel-supervisor-20-run-after-stabilization` completed. Modified `.scratch/cli-structure-refactor/goal-runbook.md` only. Repository verification `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=20 ./internal/commands/tunnel -run '^TestFRPSupervisorRejectsAnActivationExitAndKeepsConfigurationFailuresStopped$'` exited 0 in 8.074s with no skipped or retried failure. Final G0 acceptance: E1 satisfied by the `cd0e7f179939a1235d92ec5f9f43c2d2f0ec0072` rollback commit, recorded worktree, Darwin/arm64, and Go 1.27.0 / pinned Go 1.26.7 facts; E2 satisfied by three distinct zero-exit ordinary-suite runs after stabilization (`cmd/ycy`: 79.679s, 96.074s, 79.607s; Tunnel: 61.025s, 72.945s, 60.566s); E3 satisfied by the zero-exit 20-run named Tunnel result; E4 satisfied by final `git status --short` / `git diff --name-only` showing only this runbook, `git diff --check` passing, and the only stabilization commit changing `internal/commands/tunnel/frp_supervisor_unix_test.go` rather than production implementation. G0 is `passed`; its direct successor G1 is activated but not started. Risk: none remaining for G0. Next: end this Goal without reading, implementing, or verifying G1.

### G1: Freeze surface and establish acceptance architecture

- 2026-08-29: initialized as `planned`; no implementation evidence.
- 2026-08-29: activated after G0 E1-E4 were recorded; implementation has not started.

### G2: Lift root and introduce the bounded Factory

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G3: Migrate simple command vertical slices

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G4: Migrate nested Config groups

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G5: Migrate Git leaves and shared process runtime

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G6: Migrate long-running and worker-backed commands

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G7: Finalize the process composition root

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G8: Publish and audit the final structure

- 2026-08-29: initialized as `planned`; no implementation evidence.

### G9: Collect native and release evidence

- 2026-08-29: initialized as `planned`; no implementation evidence.
