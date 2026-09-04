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
| G0: Production B Ops Console foundation | passed | none | `implementation-plan.md` -> `G0: Production B Ops Console foundation` | exit conditions 1-5 verified 2026-09-04 |
| G1: Root and low-risk finite recovery | passed | G0 | `implementation-plan.md` -> `G1: Root and low-risk finite recovery` | exit conditions 1-4 verified 2026-09-04 |
| G2: External-read and service-boundary recovery | active | G1 | `implementation-plan.md` -> `G2: External-read and service-boundary recovery` | G1 passed 2026-09-04 |
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

- 2026-09-04: completed slice `descriptor/state model`. Added the semantic
  `ConsoleDescriptor`/`ConsoleMetadata` seam, bounded safe-field validation,
  and `Runtime.OpenConsole`; legacy `Open` now receives a B-compatible default
  descriptor so no command package changed. The descriptor is captured before
  Rich controller construction and exposed through the `Experience` interface
  for later command-owned adapters. Modified: `internal/terminal/{semantic.go,
  experience.go,rich.go,console.go,console_test.go}` and
  `internal/terminaltest/semantic_recording.go`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  ./internal/terminaltest` passed. Risk: the production Rich root still uses
  the generic composition until the wide-shell slice replaces it. Next action:
  implement and test the B wide shell only.

- 2026-09-04: completed slice `wide shell`. At `70x20` and above, the Rich
  root now renders the B command/status bar, bounded metadata row, divider,
  aligned `STATE / PHASE / DETAIL` table, and a single active region below the
  table. The table owns complete phase history; the active region now contains
  only the current work summary while preserving Esc's second-press
  cancellation prompt and cancellation state. Modified:
  `internal/terminal/{rich.go,console_test.go}`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  ./internal/terminaltest` passed after correcting a missing Lip Gloss import,
  removing duplicate generic phase rows, and restoring the existing Esc
  confirmation prompt. Risk: sub-`70x20` terminals still use the old compact
  projection and can omit reached phase rows. Next action: implement and test
  the B compact shell only.

- 2026-09-04: completed slice `compact shell`. Below `70x20`, the Rich root
  now uses a single-column B projection with the command/status bar, bounded
  metadata, `STATE / PHASE / DETAIL` heading, every catalog row in order, and
  the active region underneath. The active work summary retains the operation
  label, current phase detail, and Esc cancellation prompts. Modified:
  `internal/terminal/rich.go`, `internal/terminal/console_test.go`, and the
  narrow-model expectations in `internal/terminal/tracked_model_test.go`.
  Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  ./internal/terminaltest` passed. Risk: the B palette and Huh bottom-focus
  theme are still pending; compact output currently uses the existing generic
  semantic role colors. Next action: replace the shared Rich palette and Huh
  theme with the B contract only.

- 2026-09-04: completed slice `B palette and Huh theme`. Replaced generic
  Rich roles with the accepted B amber `#FFB454`, cyan `#4CC9F0`, green,
  yellow, red, text, and muted colors. Added the sole production Huh theme:
  it has the B bottom focus rule, no left rail, paired select/multi-select
  symbols, and capability-gated emphasis/background styling so `NO_COLOR=1`
  emits no ANSI bytes. Modified: `internal/terminal/{presentation.go,
  rich_form.go,theme.go,console_test.go}`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  ./internal/terminaltest` passed after correcting no-color bold ANSI output.
  Risk: reached form steps are still represented by one generic table row;
  they need their own semantic catalog projection. Next action: implement and
  test form-step projection only.

- 2026-09-04: completed slice `form-step projection`. Each successfully
  reached Rich interaction now adds one ordered Console table row from its safe
  `TranscriptLabel` (or safe message fallback); text, select, multi-select,
  confirm, and secret interaction kinds expose only their generic safe step
  detail. Completion and cancellation update the corresponding row without
  exposing answers or secret values, while the active Huh form remains below
  the table. Modified: `internal/terminal/{rich.go,console_test.go}`.
  Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./internal/terminal ./internal/terminaltest` passed. Risk: Work Phase
  catalog rows still need an explicit lasting table projection after their
  active region is replaced. Next action: implement and test phase-table
  projection only.

- 2026-09-04: completed slice `phase-table projection`. A Track now projects
  its immutable catalog in order into the one Console status table; updates
  replace only that row's state/detail, and final rows persist after a later
  Notice or Form replaces the active region. The old generic final-phase
  Notice projection was removed, so completed phases cannot appear twice or
  displace the B hierarchy. Modified:
  `internal/terminal/{rich.go,tracked.go,console_test.go}`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./internal/terminal ./internal/terminaltest` passed. Risk: durable Rich
  documents still use generic role sequencing rather than the B-derived
  hierarchy required for root/help in its later Gate. Next action: implement
  and test direct durable B hierarchy only, without changing document content
  or terminal mode behavior.

- 2026-09-04: completed slice `direct durable B hierarchy`. Direct `WriteRich`
  documents now derive their roles from the production B palette, using cyan
  for document section/active hierarchy while retaining amber for the live
  Console bar and state marker. This is a styling-only terminal projection:
  document text, Plain output, stdout ownership, and non-AltScreen behavior
  remain unchanged. Modified:
  `internal/terminal/{presentation.go,presentation_test.go}`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  ./internal/terminal ./internal/terminaltest` passed. Risk: B structural
  model/PTY evidence still needs to demonstrate colored and `NO_COLOR` console
  frames, compact retention, and the existing restoration/replay lifecycle.
  Next action: add and run normalized model and PTY evidence only.

- 2026-09-04: completed slice `legacy introduction compatibility projection`.
  The shared Rich root now recognizes only the bounded, static three-block
  introduction shape used by commands that still use the default Console
  descriptor, and projects its safe command identity into the B bar and its
  title/description into the single metadata row. This keeps established
  Rich PTY introductions observable when a later B frame replaces the first
  frame, without changing command wording, Transcript contents, stdout
  Results, or the latest-Notice active-context rule. Modified:
  `internal/terminal/rich.go`. Verification: targeted `rm` Rich PTY success and
  cancellation journeys passed in both color and `NO_COLOR=1`; shared terminal
  tests and the prototype tests also passed. Risk: the full repository suite
  remains the final compatibility check. Next action: run the G0 repository
  verification sequence.

- 2026-09-04: completed slice `normalized model and PTY evidence`, including
  the final shared-renderer compatibility repair. Model evidence covers the
  descriptor/bar/metadata/divider/table/active-region hierarchy, B state
  symbols and palette, no-color control-free rendering, compact `70x20`
  degradation, ordered form and Track rows, no persistent rail, and bounded
  latest-Notice context. PTY evidence covers colored and `NO_COLOR=1` Rich
  journeys, resize retention, AltScreen restoration, cancellation, renderer
  lease/recovery, Transcript replay ordering, redirected stdout, and child
  handoff. The repository run initially exposed legacy `rm` Rich PTY
  expectations for the static introduction; the shared renderer now projects
  only that bounded safe three-block introduction into the B bar/metadata row,
  preserving command packages, wording, Results, Transcript semantics, and
  stdout ownership. Modified: `internal/terminal/rich.go` plus the existing
  G0 evidence files. Directed verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/terminal
  ./internal/terminaltest -count=1`,
  `cd internal/terminal/prototype-vivid && GOTOOLCHAIN=go1.26.7 GOWORK=off
  go test ./...`, and targeted `pkg/cmd/rm` Rich PTY success/cancellation
  tests in color and `NO_COLOR=1`. Repository verification passed in order:
  the shared terminal test command, `make command-surface`, `make check`, and
  `git diff --check`; `make check` completed all Go, architecture, Web, and
  command-package checks. Exit-condition evidence: (1) semantic descriptor,
  safe metadata, B bar/divider/wide table/active region: model tests and
  `internal/terminal` tests; (2) ordered form/Track rows, paired symbols,
  palette, bottom focus, and no rail: console/theme/model/PTY tests; (3)
  compact heading, rows, details, and active region below `70x20`: compact
  model and resize PTY tests; (4) Rich/Plain/Automation/redirected,
  Transcript, lease, cancellation, recovery, and exactly-once behavior:
  shared runtime and repository tests; (5) repository checks: all commands
  above passed. Risk: G1 is now active but intentionally has no
  implementation evidence in this Goal. Next action: end this Goal after the
  atomic Gate transition.

### G1: Root and low-risk finite recovery

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.
- 2026-09-04: 已激活，尚未开始实施。

- 2026-09-04: completed slice `root/help durable B hierarchy`. The discovery
  adapter now assigns the B active hierarchy role to the durable `Usage`,
  `Commands`, `Flags`, and `Examples` headings while leaving every command,
  flag, example, order, and Plain/Automation byte intact. The root remains a
  direct stdout document with no Console start, AltScreen, or Transcript.
  Modified: `pkg/cmd/root/{terminal_discovery.go,terminal_discovery_test.go}`.
  Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  -count=1 ./pkg/cmd/root` passed; the focused test verifies exact Plain
  content, colored B roles, and control-free no-color output alongside the
  existing version, completion, parser, and write-failure coverage. Risk:
  command-specific Rich PTY evidence for the remaining finite adapters is
  still absent. Next action: migrate only `config fork list` to an explicit
  safe Console descriptor and add its structural evidence.

- 2026-09-04: completed slice `config fork list descriptor/context`. The
  command now opens the frozen G0 Console seam with its own bounded context:
  `YCY / config fork list`, `provider inventory`, and the static scope `git
  fork configuration`. It does not place instance names, hosts, token previews,
  paths, or configuration data in the B bar/metadata row. Read ordering,
  loading phase, result documents, Plain/Automation behavior, and Transcript
  projection are unchanged. Modified: `pkg/cmd/config/fork/list/{command.go,
  terminal_test.go}`. Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off
  CGO_ENABLED=0 go test -count=1 ./pkg/cmd/config/fork/list` passed. Risk:
  this command still needs fresh wide/compact color/no-color Rich PTY evidence.
  Next action: add only that PTY/stream evidence.

- 2026-09-04: completed slice `config fork list Rich PTY and stream evidence`.
  Added controlled PTY coverage at `120x40` and `40x15`, with and without
  color. The evidence asserts the command-owned B bar and safe metadata,
  ordered `STATE / PHASE / DETAIL` structure, active/loading and completed
  phase projection, compact width-bounded context, AltScreen restoration,
  Transcript-before-result ordering, Transcript redaction, complete durable
  Result fields, and control-free `NO_COLOR=1` output. Modified:
  `pkg/cmd/config/fork/list/terminal_pty_test.go`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./pkg/cmd/config/fork/list` passed. Risk: the remaining G1 adapters still
  need their own descriptor/context, phase or form, and structural PTY
  evidence. Next action: migrate only `config cm list`.

- 2026-09-04: completed slice `config cm list descriptor/context and default
  summary`. The command now opens its own bounded Console descriptor before
  store construction: `YCY / config cm list`, target `profile inventory`, and
  static scope `commit message configuration`. Rich successful runs retain a
  safe `Default profile: <name>` milestone only for a listed, display-safe
  default; no profile rows, Base URLs, models, or API keys enter the
  Interaction Transcript. Modified:
  `pkg/cmd/config/cm/list/{command.go,presentation.go,terminal_test.go}`.
  Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test
  -count=1 ./pkg/cmd/config/cm/list` passed. Risk: command-specific Rich PTY
  and stream evidence is now complete for this adapter; remaining G1
  adapters still need migration. Next action: migrate only `config cm use`.

- 2026-09-04: completed slice `config cm list Rich PTY and stream evidence`.
  Added controlled `120x40`/`40x15` color and `NO_COLOR=1` PTY journeys that
  assert B context, phase table order, completed state, primary-screen
  restoration, Transcript phase/summary/default/outcome ordering and
  redaction, and complete post-AltScreen Result tables. Plain and Automation
  assertions preserve exact result bytes, loading-diagnostic ownership,
  control-free streams, and no partial output on read failure. Modified:
  `pkg/cmd/config/cm/list/terminal_pty_test.go`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./pkg/cmd/config/cm/list` passed. Risk: `config cm use` and `git heat`
  remain unqualified in G1. Next action: migrate only `config cm use`.

- 2026-09-04: completed slice `config cm use descriptor and safe phase
  context`. The atomic default-selection command now opens a bounded Console
  descriptor before store construction: `YCY / config cm use`, target `profile
  selection`, and static scope `commit message configuration`. Its single
  Work Phase includes the bounded safe requested profile in the active and
  successful detail while preserving the exact profile identity passed to the
  writer and the existing result/error contract. Modified:
  `pkg/cmd/config/cm/use/{command.go,terminal_test.go}`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./pkg/cmd/config/cm/use` passed. Risk: command-specific PTY and stream
  evidence remains to be recorded. Next action: add only that evidence.

- 2026-09-04: completed slice `config cm use Rich PTY and stream evidence`.
  Added controlled `120x40`/`40x15` color and `NO_COLOR=1` PTY journeys that
  assert B context, safe profile detail, phase/table order, completed state,
  primary-screen restoration, Transcript-before-result ordering, and durable
  success output. Plain and Automation assertions preserve loading diagnostic
  ownership, exact result bytes, control-free output, and no partial output
  on a missing-profile failure. Modified:
  `pkg/cmd/config/cm/use/terminal_pty_test.go`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./pkg/cmd/config/cm/use` passed. Risk: only `git heat` remains before G1
  repository verification. Next action: migrate only `git heat`.

- 2026-09-04: completed slice `git heat descriptor/context`. The finite Git
  inspection command now opens the frozen Console seam after input
  normalization and before repository discovery, with only the bounded
  command identity, target, normalized range, sort, and time mode in its
  descriptor. Query text, repository paths, Git arguments, timings, and path
  rows remain outside Console metadata; Plain/Automation result and error
  behavior remain unchanged. Modified: `pkg/cmd/git/heat/{terminal.go,
  terminal_test.go}`. Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off
  CGO_ENABLED=0 go test -count=1 ./pkg/cmd/git/heat` passed. Risk: the
  command still needs the final fresh structural PTY and stream evidence
  review before G1 repository verification. Next action: add only the
  `git heat` Rich PTY/stream evidence.

- 2026-09-04: completed slice `git heat Rich PTY and stream evidence`.
  Added controlled `120x40` and `40x15` color/no-color journeys covering the
  B bar, safe normalized metadata, all three truthful phases, wide/compact
  rows, latest/earliest markers, primary-screen restoration, bounded
  Transcript redaction, and complete durable Result output. Plain and
  Automation fixtures assert loading ownership, exact result bytes, control
  stripping, and no partial output on repository failure. Modified:
  `pkg/cmd/git/heat/terminal_pty_test.go`. Verification:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./pkg/cmd/git/heat` passed. Risk: final G1 repository checks and human
  review remain. Next action: run the declared G1 repository verification
  sequence in order.

- 2026-09-04: repository verification step 1 passed. Command and shared
  terminal package suite completed successfully:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./pkg/cmd/root
  ./pkg/cmd/config/fork/list ./pkg/cmd/config/cm/list ./pkg/cmd/config/cm/use
  ./pkg/cmd/git/heat ./internal/terminal`. No additional source changes were
  required. Risk: command-surface, repository check, acceptance journeys, and
  diff hygiene remain. Next action: run `make command-surface`.

- 2026-09-04: repository verification step 2a passed. `make command-surface`
  completed successfully, confirming the frozen command/help surface remains
  compatible. No additional source changes were required. Risk: full
  repository checks, tagged G1 acceptance journeys, and diff hygiene remain.
  Next action: run `make check`.

- 2026-09-04: diagnosed and repaired the first repository-check failure
  without changing G1 runtime behavior. New PTY assertions had imported the
  Charm ANSI helper directly from command packages, violating the repository
  dependency boundary; the helper now lives behind `internal/terminaltest` and
  command tests call that boundary. Modified:
  `internal/terminaltest/{streams.go,terminaltest_test.go}` and the five G1
  PTY/discovery test files. Directed re-verification passed:
  `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1
  ./pkg/cmd/root ./pkg/cmd/config/fork/list ./pkg/cmd/config/cm/list
  ./pkg/cmd/config/cm/use ./pkg/cmd/git/heat ./internal/terminal
  ./internal/terminaltest`. Risk: full `make check` must be rerun; its prior
  unrelated tunnel-server integration timeout remains under observation. Next
  action: rerun `make check`.

- 2026-09-04: repository verification step 2b passed. The rerun of
  `make check` completed across the Go and web repository suites, including
  architecture/dependency checks, all command packages, and the tunnel-server
  integration suite. The earlier Charm import violations did not recur; no
  unrelated source changes were made. Risk: tagged G1 standalone acceptance
  journeys and final diff hygiene remain. Next action: run the declared
  acceptance command.

- 2026-09-04: repository verification step 3 passed. The tagged standalone
  acceptance journeys for `config fork list`, `config cm list`, `config cm
  use`, and `git heat` completed successfully with `GOTOOLCHAIN=go1.26.7
  GOWORK=off CGO_ENABLED=0 go test -count=1 -tags=acceptance ./acceptance
  -run '^(TestConfigForkListStandaloneBinary|TestConfigCMListStandaloneBinary|TestConfigCMUseStandaloneBinary|TestGitHeatStandaloneBinary)'`.
  Risk: only final diff hygiene and manual acceptance handoff remain. Next
  action: run `git diff --check`.

- 2026-09-04: repository verification step 4 passed: `git diff --check`
  completed with no whitespace errors. Final scope review found only G1
  presentation adapters, their evidence, and the necessary internal test
  boundary; no G2+ command was changed. Automated exit evidence is complete:
  the declared G1 package suite, `make command-surface`, rerun `make check`,
  the four tagged standalone journeys, and final diff hygiene all passed.

  Manual acceptance handoff: a local binary built from this worktree is ready
  at `/tmp/hackycy-g1-acceptance.l54jPH/ycy`; its isolated fake-config home is
  `/tmp/hackycy-g1-acceptance.l54jPH/home`, and its contained two-commit Git
  fixture is `/tmp/hackycy-g1-acceptance.l54jPH/repo`. The fixture smoke run
  passed for fork list, CM list, CM use, and Git heat, without touching the
  user configuration or the repository state. In a colored `120x40` TTY, set
  `HOME=/tmp/hackycy-g1-acceptance.l54jPH/home` and `USERPROFILE=` and inspect
  `ycy --help`, `ycy config fork list`, `ycy config cm list`, `ycy config cm
  use work`, and, from the contained Git fixture, `ycy git heat -n 2 -t files
  -s path -r -q main`. Repeat those scenarios in a `40x15` TTY with
  `NO_COLOR=1`.

  Minimum acceptance checklist: root/help stays a complete durable discovery
  document without AltScreen; each finite command shows the B hierarchy with
  safe bounded metadata and complete wide/compact phase/result rows; no-color
  symbols remain readable and have no persistent rail; finite commands restore
  the primary screen before a bounded Transcript and unchanged durable Result;
  no fixture token, API key, path, or full report row leaks into Transcript.
  Reply exactly once using `通过: root/help, config fork list, config cm list,
  config cm use, git heat` or `未通过: <command and reason>`. Next action: a
  new Goal with that explicit result may continue G1; this Goal ends now with
  G1 still `active` and no successor activated.

- 2026-09-04: final G1 acceptance completed. The user explicitly replied
  `全部通过`, accepting the handoff scenarios for root/help, `config fork
  list`, `config cm list`, `config cm use`, and `git heat`. Exit condition 1
  passed: root/help retained its durable discovery, parser, version,
  completion, redirected-output, and write-failure contracts while using the
  B hierarchy, evidenced by the focused root suite and `make
  command-surface`. Exit condition 2 passed: all four finite commands have
  fresh wide/compact color/no-color B structural PTY evidence plus behavior,
  redaction, Transcript, Plain, Automation, and standalone coverage. Exit
  condition 3 passed through the explicit manual acceptance. Exit condition 4
  passed: the declared package suite, `make command-surface`, rerun `make
  check`, tagged standalone acceptance journeys, and `git diff --check` all
  passed. G1 is now `passed`; its direct successor G2 is activated but has not
  been read, analyzed, implemented, or verified in this Goal.

### G2: External-read and service-boundary recovery

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.
- 2026-09-04: 已激活，尚未开始实施。

### G3: Mutation and archive recovery

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G4: Git mutation and process handoff

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G5: Service lifecycles and detached updater

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.

### G6: Release verification and full visual review

- 2026-09-04: initialized as `planned`; no recovery implementation evidence.
