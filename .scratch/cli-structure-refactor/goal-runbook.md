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
| G0: Stabilize pre-migration baseline | active | none | `implementation-plan.md` -> `G0: Stabilize pre-migration baseline` | no predecessor |
| G1: Freeze surface and establish acceptance architecture | planned | G0 | `implementation-plan.md` -> `G1: Freeze surface and establish acceptance architecture` | G0 pending |
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

### G1: Freeze surface and establish acceptance architecture

- 2026-08-29: initialized as `planned`; no implementation evidence.

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
