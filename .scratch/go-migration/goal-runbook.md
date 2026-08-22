# Go Migration Goal Runbook

## Source Baseline

| Path | SHA-256 |
| --- | --- |
| `.scratch/go-migration/acceptance.md` | `7d8964319a45b560b086ab2a7938f0d0160da7b526dbae1471fcfbfe0746dd05` |
| `.scratch/go-migration/implementation-plan.md` | `3d6b322b9d603e24a45e936b85584e5edf7b168040a5d18b63b4a08fd35efa0a` |
| `.scratch/go-migration/inventories/core-command-contracts.md` | `d8cf671f8a168a93284d15830f8f4e126dbd4e3d3c30c6f8f08d747ed245e1cd` |
| `.scratch/go-migration/inventories/diff-contracts.md` | `306553fa5bfe830638874dd28d9ac018d8b5de8935df9f894f9f4e64babb2b49` |
| `.scratch/go-migration/inventories/fs-contracts.md` | `73a743bb5baf7d74bf54ca5a54f9b363cae594ff36cf3331a74f87b2ecab9dff` |
| `.scratch/go-migration/inventories/git-command-contracts.md` | `eff204a0f92d9b40220a94a0ea6f5dd22dc5499227049e62fb3ee1b01a69ce51` |
| `.scratch/go-migration/inventories/tunnel-contracts.md` | `47643fea49117ea3282762aae22d48760ab0c97ead0a8aa1af6249803d2efeb5` |
| `.scratch/go-migration/inventories/upgrade-artifact-contracts.md` | `0e3f6229d1c59a2f05475ae134c1f13c8dcdd8f46569f684b2be77e143ca64c2` |
| `.scratch/go-migration/issues/01-research-pure-go-toolchain.md` | `b87e6aafde254b1f85a19d3d3fa35705393a542ccca8b6bcaf8352fe6c2dcca7` |
| `.scratch/go-migration/issues/02-research-mixed-project-quality-gates.md` | `4e47598d609f8a408bdc94752042d2c1db49f558f6c790931bfd8d82ed9e07de` |
| `.scratch/go-migration/issues/03-research-vite-go-embedding.md` | `3bf2a6b9e6519f83e35ff9cab8bfa71b17c5d36d11c823f2c23844266e49e636` |
| `.scratch/go-migration/issues/04-inventory-core-command-contracts.md` | `e1c2e0454f694bdf240713721921259fa8b270c673e076e5edaf7ee39f8bc3d6` |
| `.scratch/go-migration/issues/05-inventory-git-command-contracts.md` | `869ec7ab907493182b46e1cb700bff53dc411c27786830b52944a0e4cd8d582d` |
| `.scratch/go-migration/issues/06-inventory-diff-contracts.md` | `b2835461e60fac03d10038c68bd7685b0bbd73a679164af2e0b1b97bea258783` |
| `.scratch/go-migration/issues/07-inventory-fs-contracts.md` | `eeb9f912e0b36630fe9bb259c97b349ac36d74363e34bbffc718158e50bfe9ae` |
| `.scratch/go-migration/issues/08-inventory-tunnel-contracts.md` | `fd7cdd928659456b36d8aa9d3b30e9c715411a6153bac8e1393c376eea8a02f2` |
| `.scratch/go-migration/issues/09-inventory-upgrade-artifact-contracts.md` | `6935af200521f7b4a9b37e285d890057dc65312554cf08983553658a54542a39` |
| `.scratch/go-migration/issues/10-prove-go-cli-compatibility.md` | `72d265f33155808525fc844b0398cb008630bd8300084717e529bc6307f3eafc` |
| `.scratch/go-migration/issues/11-prove-vite-go-embed-path.md` | `7b8e23775eec954ecb8d9e195ee5d3cdc19ed347ed38c3214fbfea6e495fc8ca` |
| `.scratch/go-migration/issues/12-choose-data-compatibility-mechanisms.md` | `01163e82656bbc7deb3df56ce8762f9589bdb5c9e2ee3f44ddd5e68d8ee896c8` |
| `.scratch/go-migration/issues/13-choose-go-module-seams.md` | `09adc09737e2d77218dfb71f82dd9bcf7345358842df4ce151bd2135309e5277` |
| `.scratch/go-migration/issues/14-choose-mixed-project-hook-policy.md` | `6c779ab10063df3dbb409b4af5ce14e82154c3a2fcb275652aa9067745b34edd` |
| `.scratch/go-migration/issues/15-define-archive-cutover.md` | `ad41d8fd8414d21d9d2e61f33902b53a99477e158ca1835afad59c93205e4189` |
| `.scratch/go-migration/issues/16-approve-command-migration-roadmap.md` | `1552880d04783cdc33a94958829e8232d2c950ef119107183dfd9a9c0fef91a0` |
| `.scratch/go-migration/issues/17-choose-corrected-core-command-contracts.md` | `946dd26e7cc6f0e40f2151eb4387b606733e29dfa9492fe220119299c8f2b158` |
| `.scratch/go-migration/issues/18-choose-safe-git-command-contracts.md` | `98189775c1ff7a732d389b721be1115d2eed7dd43a1d28a59868b50bdb385afb` |
| `.scratch/go-migration/issues/19-choose-safe-diff-service-contracts.md` | `0dc55c3c7632eb08e1d8059fc4989383229dd0d674a5efe641d8bfe41dd833f8` |
| `.scratch/go-migration/issues/20-choose-safe-fs-service-contracts.md` | `797e427de505eafcdd7d464725b846733692e77b2dd1b41f0cb34fca669895e2` |
| `.scratch/go-migration/issues/21-research-cgo-free-fs-thumbnails.md` | `fb84c8950331c840a3b89504e047312c7588638204bf41f279d7339bf0fb7a8e` |
| `.scratch/go-migration/issues/22-choose-safe-rolling-tunnel-contracts.md` | `de560c7fd13ae730ad1d009a04df8d113170781d39c1054ff539692c1aafea2b` |
| `.scratch/go-migration/issues/23-choose-safe-self-update-contract.md` | `70588a8bb5f62bcd89b71a985b991d43580cafc5ef8294a095fc90ce9d38ff02` |
| `.scratch/go-migration/map.md` | `b4e3432a8147dd96daa1e8ff50f9b139f232bc833398fb5179f83f557653f591` |
| `.scratch/go-migration/prototypes/go-cli-compat/README.md` | `d76900c32d841c390b65de4f4894834a2aabe1b83ad80432a0b9ee3c7afb7773` |
| `.scratch/go-migration/prototypes/go-cli-compat/RESULTS.md` | `6f202c4a8e9c79a5baf7193e9a77467a5bf8f77ffde7662ef574400462af763e` |
| `.scratch/go-migration/prototypes/vite-go-embed-preview.html` | `3438c4d03d4da8709059f0f2c3d59b8fd191b7c68cfcb9a3ecf33dcafa71187d` |
| `.scratch/go-migration/research/21-cgo-free-fs-thumbnails.md` | `72cab075883631b7574b6bba3b139a39dcf6c9c988626e1cc2a37fcfda6e2cef` |
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
| G0: Foundation Gate | active | none | `implementation-plan.md` -> `G0: Foundation Gate` | no predecessor |
| G1: export env | planned | G0 | `implementation-plan.md` -> `G1: export env` | G0 pending |
| G2: appconfig foundation | planned | G1 | `implementation-plan.md` -> `G2: appconfig foundation` | G1 pending |
| G3: config fork list | planned | G2 | `implementation-plan.md` -> `G3: config fork list` | G2 pending |
| G4: config fork add | planned | G3 | `implementation-plan.md` -> `G4: config fork add` | G3 pending |
| G5: config fork remove | planned | G4 | `implementation-plan.md` -> `G5: config fork remove` | G4 pending |
| G6: config cm list | planned | G5 | `implementation-plan.md` -> `G6: config cm list` | G5 pending |
| G7: config cm add | planned | G6 | `implementation-plan.md` -> `G7: config cm add` | G6 pending |
| G8: config cm use | planned | G7 | `implementation-plan.md` -> `G8: config cm use` | G7 pending |
| G9: config cm set | planned | G8 | `implementation-plan.md` -> `G9: config cm set` | G8 pending |
| G10: config cm remove | planned | G9 | `implementation-plan.md` -> `G10: config cm remove` | G9 pending |
| G11: config cm test | planned | G10 | `implementation-plan.md` -> `G11: config cm test` | G10 pending |
| G12: rm | planned | G11 | `implementation-plan.md` -> `G12: rm` | G11 pending |
| G13: run | planned | G12 | `implementation-plan.md` -> `G13: run` | G12 pending |
| G14: git heat | planned | G13 | `implementation-plan.md` -> `G14: git heat` | G13 pending |
| G15: git pulse | planned | G14 | `implementation-plan.md` -> `G15: git pulse` | G14 pending |
| G16: zip | planned | G15 | `implementation-plan.md` -> `G16: zip` | G15 pending |
| G17: git fork | planned | G16 | `implementation-plan.md` -> `G17: git fork` | G16 pending |
| G18: git cm | planned | G17 | `implementation-plan.md` -> `G18: git cm` | G17 pending |
| G19: Web Readiness Gate | planned | G18 | `implementation-plan.md` -> `G19: Web Readiness Gate` | G18 pending |
| G20: diff | planned | G19 | `implementation-plan.md` -> `G20: diff` | G19 pending |
| G21: FS Foundation | planned | G20 | `implementation-plan.md` -> `G21: FS Foundation` | G20 pending |
| G22: fs | planned | G21 | `implementation-plan.md` -> `G22: fs` | G21 pending |
| G23: Tunnel Foundation | planned | G22 | `implementation-plan.md` -> `G23: Tunnel Foundation` | G22 pending |
| G24: tunnel server | planned | G23 | `implementation-plan.md` -> `G24: tunnel server` | G23 pending |
| G25: tunnel connect | planned | G24 | `implementation-plan.md` -> `G25: tunnel connect` | G24 pending |
| G26: upgrade | planned | G25 | `implementation-plan.md` -> `G26: upgrade` | G25 pending |
| G27: Final Artifact Gate | planned | G26 | `implementation-plan.md` -> `G27: Final Artifact Gate` | G26 pending |

## Progress Log

### G0: Foundation Gate

- 2026-08-23: initialized as `active`; implementation has not started.

### G1: export env

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G2: appconfig foundation

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G3: config fork list

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G4: config fork add

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G5: config fork remove

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G6: config cm list

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G7: config cm add

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G8: config cm use

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G9: config cm set

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G10: config cm remove

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G11: config cm test

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G12: rm

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G13: run

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G14: git heat

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G15: git pulse

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G16: zip

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G17: git fork

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G18: git cm

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G19: Web Readiness Gate

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G20: diff

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G21: FS Foundation

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G22: fs

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G23: Tunnel Foundation

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G24: tunnel server

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G25: tunnel connect

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G26: upgrade

- 2026-08-23: initialized as `planned`; no implementation evidence.

### G27: Final Artifact Gate

- 2026-08-23: initialized as `planned`; no implementation evidence.
