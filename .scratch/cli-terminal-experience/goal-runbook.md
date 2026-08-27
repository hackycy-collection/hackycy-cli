# CLI Terminal Experience Goal Runbook

## Source Baseline

| Path | SHA-256 |
| --- | --- |
| `.scratch/cli-terminal-experience/implementation-plan.md` | `3956cae169224c2e788dfec8cb6b4b014edafe50d3626167f15ca859cc24c4f8` |
| `.scratch/cli-terminal-experience/inventory/03-command-experience-output-contracts.md` | `81af47a95c4ec166aaa5bd6811fbb1fd0ca93572e0741c1d0c6be29855bd0edb` |
| `.scratch/cli-terminal-experience/issues/01-research-charmbracelet-terminal-runtime.md` | `5736a5752dc2e501fe9f73acdf9596e701f1c5d838f722ab3f17e27eb4f95486` |
| `.scratch/cli-terminal-experience/issues/02-define-terminal-capability-and-automation-contract.md` | `c3185b0aeac6b234434ca7eb29b7f2ebed78c218e35c6877224844fbdad239c5` |
| `.scratch/cli-terminal-experience/issues/03-inventory-command-experience-and-output-contracts.md` | `a0750641c2af9ac73804670751d538a769efa2e058dcbf047e89643bc54df741` |
| `.scratch/cli-terminal-experience/issues/04-prototype-the-terminal-visual-language.md` | `9adb84e17df9933a1f868266d3905fcb808825fb704b74d775880003acc26296` |
| `.scratch/cli-terminal-experience/issues/05-choose-the-terminal-experience-module-seam.md` | `bf9846c232ca9503d3a0dff1be29477356ec59bf7fd91c635e7e59d1734d5268` |
| `.scratch/cli-terminal-experience/issues/06-define-output-and-diagnostic-contract.md` | `db7d135c90b78faea10ed57679cccdc2ee623d1e95a818f2cff765f40b5e9aa2` |
| `.scratch/cli-terminal-experience/issues/07-choose-the-long-running-operation-model.md` | `ace2687573252e279c343ba245abb4cc4e007186ac344b4577fb8a16bc85c677` |
| `.scratch/cli-terminal-experience/issues/08-define-help-error-and-completion-experience.md` | `fec1816b30288c8ff503a69770183e3f684c35d644913d4dee64b3feccf45afb` |
| `.scratch/cli-terminal-experience/issues/09-approve-the-cli-experience-migration-and-acceptance-plan.md` | `5ed0a69cbf1b2c3438d1c25e5e6f6424d65b139841a127a9bcde4ee5b5d03d75` |
| `.scratch/cli-terminal-experience/issues/10-choose-automation-inputs-for-prompted-commands.md` | `2a650235be44ae248d6528ec44bfc7a6445b6d6cd5219b4a9eb5066cd84db538` |
| `.scratch/cli-terminal-experience/map.md` | `91b90b00fe5d3907a0d7c64bd43b027715626b81210615762523df47c15ea26e` |
| `.scratch/cli-terminal-experience/prototype/bun-baseline/README.md` | `f200c29877408e19bd22ca96d08b41a6bbab10de02c2226dd6f395b70513bfba` |
| `.scratch/cli-terminal-experience/research/01-charmbracelet-terminal-runtime.md` | `59a8e2e90eac1c6bf1c0801c42a78eeb13f19d5cdd8a17a66628ddd1aa7f7c6e` |
| `.scratch/go-migration/acceptance.md` | `61f90f051d982dfafe7b94aab7ac3e7b065c9190db88ddb0304b62a910b86667` |
| `.scratch/go-migration/map.md` | `d81d65e6fb5a31fa20291efbda5a79d6f389856f3d2ae56897d5f72b58802986` |
| `CONTEXT.md` | `2fd5522e5f4fbaf77271ef0b60411e472f35b3e9e1eac2918803c5cc15621cf3` |

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
| G0: Characterize Protected Behavior | `active` | none | `implementation-plan.md` -> `G0: Characterize Protected Behavior` | no predecessor |
| G1: Establish Terminal Foundation | `planned` | G0 | `implementation-plan.md` -> `G1: Establish Terminal Foundation` | G0 pending |
| G2: Wire Root and Diagnostics | `planned` | G1 | `implementation-plan.md` -> `G2: Wire Root and Diagnostics` | G1 pending |
| G3: Migrate Static and Continuous Journeys | `planned` | G2 | `implementation-plan.md` -> `G3: Migrate Static and Continuous Journeys` | G2 pending |
| G4: Migrate Ordinary Form Journeys | `planned` | G3 | `implementation-plan.md` -> `G4: Migrate Ordinary Form Journeys` | G3 pending |
| G5: Migrate Tracked Git Journeys | `planned` | G4 | `implementation-plan.md` -> `G5: Migrate Tracked Git Journeys` | G4 pending |
| G6: Audit Completion and Accept | `planned` | G5 | `implementation-plan.md` -> `G6: Audit Completion and Accept` | G5 pending |

## Progress Log

### G0: Characterize Protected Behavior

- 2026-08-27: initialized as `active`; implementation has not started.

### G1: Establish Terminal Foundation

- 2026-08-27: initialized as `planned`; no implementation evidence.

### G2: Wire Root and Diagnostics

- 2026-08-27: initialized as `planned`; no implementation evidence.

### G3: Migrate Static and Continuous Journeys

- 2026-08-27: initialized as `planned`; no implementation evidence.

### G4: Migrate Ordinary Form Journeys

- 2026-08-27: initialized as `planned`; no implementation evidence.

### G5: Migrate Tracked Git Journeys

- 2026-08-27: initialized as `planned`; no implementation evidence.

### G6: Audit Completion and Accept

- 2026-08-27: initialized as `planned`; no implementation evidence.
