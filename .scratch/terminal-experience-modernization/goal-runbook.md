# Terminal Experience Modernization Recovery Goal Runbook

## Source Baseline

| Path | SHA-256 |
| --- | --- |
| `.scratch/terminal-experience-modernization/b-ops-console-audit.md` | `156dd963f9b6c0e5c52bae26abfd8cf14df4ec4c814a304294bd1a96fedf039d` |
| `.scratch/terminal-experience-modernization/baselines/README.md` | `73df9a2bcc6014801d9269cbc24c742a7c83721f1a8dddcdacbf47b101654d9f` |
| `.scratch/terminal-experience-modernization/baselines/manifest.json` | `f38bbc6cb9c87be8166b123ee53fc9c3cf018efab076c22927b4036e776b48f5` |
| `.scratch/terminal-experience-modernization/implementation-plan.md` | `e51ef38ffb89ec2ff041f939fe43b061298e40b9a0e4f019e1b8f865c63a3aad` |
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
| `Makefile` | `a0e779709707bc6fb03438e7ab03cf1f44ff3dd978a875f61b415da9861b368c` |
| `docs/platform-boundaries.md` | `be531d09b0d6bc619d54a66d6c3fb6105208f241806dfd26b70631c18f87e473` |
| `docs/project-layout.md` | `26252b0db008b5964ae2877f7cbbf1472873f14d3a6165bc123a7f925a73c5f9` |
| `internal/terminal/prototype-vivid/README.md` | `d9089026734c466ab8c1c8cf341cc621fb8348a5344571082c685c3c020a7d5e` |
| `internal/terminal/prototype-vivid/go.mod` | `24a685071f55866af716995af1d3ef9454a5dd98593a87b4ce848414f017183d` |
| `internal/terminal/prototype-vivid/go.sum` | `e59854f7db40ac0d364fe94ed6a622fbf453fe82f94d6da8977c5ad5b70d8398` |
| `internal/terminal/prototype-vivid/main.go` | `a2455b2620d39e2bd71fe29dda08d69017159279e03e9786e7086c425162e65d` |
| `internal/terminal/prototype-vivid/model.go` | `9d7c1de8bba09cc7aaa4488e68c09fedf22272cfd80e9c09ac17563fa055650a` |
| `internal/terminal/prototype-vivid/theme.go` | `3294b46f4f2448c928bd4d9590df4c53cab3f76e3c1ad33850c269abab13ec84` |
| `internal/terminal/prototype-vivid/view.go` | `9347ec260a6586dccfecda7d45dacd5c2b45a489aeed5316a9650b24e1384863` |

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
| G0: Production B Ops Console foundation | active | none | `implementation-plan.md` -> `G0: Production B Ops Console foundation` | no predecessor |
| G1: Root and low-risk finite recovery | planned | G0 | `implementation-plan.md` -> `G1: Root and low-risk finite recovery` | G0 pending |
| G2: External-read and service-boundary recovery | planned | G1 | `implementation-plan.md` -> `G2: External-read and service-boundary recovery` | G1 pending |
| G3: Mutation and archive recovery | planned | G2 | `implementation-plan.md` -> `G3: Mutation and archive recovery` | G2 pending |
| G4: Git mutation and process handoff | planned | G3 | `implementation-plan.md` -> `G4: Git mutation and process handoff` | G3 pending |
| G5: Service lifecycles and detached updater | planned | G4 | `implementation-plan.md` -> `G5: Service lifecycles and detached updater` | G4 pending |
| G6: Release verification and full visual review | planned | G5 | `implementation-plan.md` -> `G6: Release verification and full visual review` | G5 pending |

## Progress Log

### G0: Production B Ops Console foundation

- 2026-09-04: initialized as `active`; recovery implementation has not
  started. The prior G0-G8 control plane and its former manual handoff are
  archived under `superseded/2026-09-04-pre-b-ops-console-reset/` and are not
  imported because they do not establish the accepted B production shell.
  Control-plane compilation reviewed the current renderer and B prototype,
  replaced the plan/runbook, and preserved all existing command changes.
  Verification: every map/issue decision is resolved; the baseline above locks
  the source set. Risk: production Rich rendering remains structurally generic.
  Next action: execute the descriptor/state-model slice defined by G0.

### G1: Root and low-risk finite recovery

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G2: External-read and service-boundary recovery

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G3: Mutation and archive recovery

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G4: Git mutation and process handoff

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G5: Service lifecycles and detached updater

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G6: Release verification and full visual review

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.
