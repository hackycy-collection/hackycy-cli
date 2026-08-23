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
| G0: Foundation Gate | passed | none | `implementation-plan.md` -> `G0: Foundation Gate` | no predecessor |
| G1: export env | passed | G0 | `implementation-plan.md` -> `G1: export env` | G0 passed |
| G2: appconfig foundation | passed | G1 | `implementation-plan.md` -> `G2: appconfig foundation` | all G2 Exit conditions accepted |
| G3: config fork list | active | G2 | `implementation-plan.md` -> `G3: config fork list` | G2 passed |
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
- 2026-08-23: slice `frozen baseline capture` completed. Modified tracked files: none. Evidence: source commit `78358c0201b71891e36603d6abb8d7c87d54ad57` exists; current tree differs from it only by the tracked migration-planning material; Bun `1.3.14`; `bun test` passed (357 pass, 1 Windows-only skip, 0 fail); `bun run typecheck`, `bun run lint`, and `bun run build` passed. Hook preflight found no `core.hooksPath`; the existing 222-byte `pre-commit` hash is the approved `7bc48fcc880a58ab4f92dbe45343a82eea1b2539c86e5c05dc6713d39bdf5d95`. Risk: none of the G0 stop conditions is present. Next: generate the frozen archive manifest and metadata, then verify its source mapping before foundation implementation.
- 2026-08-23: slice `frozen archive capture` completed. Modified files: moved the decision-ledger Bun sources, build scripts, package metadata, workflows, Docker/deploy definitions, and original README to `legacy/bun/`; added `legacy/bun/MANIFEST.tsv` and `legacy/bun/ARCHIVE.md`; copied the four approved frontend source trees to `web/` before adaptation. Verification: the manifest has 242 source rows; manifest-to-index comparison passed with exact source blob IDs and modes; all four archive-to-active frontend directory comparisons passed byte-for-byte before adaptation. Risk: active Web sources are copied but not yet Vite-adapted, and the active Go foundation does not exist yet; no stop condition is present. Next: adapt only `web/` into the one Vite MPA and prove its generated asset graph.
- 2026-08-23: slice `Vite MPA foundation` completed. Modified files: added `web/` with the three copied React applications, Vite MPA inputs/configuration, pnpm lock/workspace metadata, TypeScript, Antfu ESLint, Vitest, Monaco worker entrypoints, and `verify-assets.mjs`. The verifier follows manifest-reachable bundles to Vite-emitted `new Worker` outputs so it validates every generated worker instead of mistaking Vite's omitted worker manifest entries for orphaned files. Verification: `pnpm --dir web run typecheck` passed; `pnpm --dir web run lint` passed with zero warnings; `pnpm --dir web run test` passed (7 files, 27 tests); `pnpm --dir web run build` passed and verified 3 shells with 424 reachable generated assets. Risk: Vite emits the standard large-chunk advisory for Monaco workers; no required verifier or contract failed. Next: add the minimal Go CLI, unconditional embed, layout/architecture checks, and host build lifecycle without registering a business command.
- 2026-08-23: slice `Go CLI and unconditional embed foundation` completed. Modified files: added root `go.mod`/`go.sum` pinned to Go 1.26.7 and Cobra; added explicit `internal/logging` runtime; added `internal/cliapp` with fresh Cobra roots, global `--log-level`, help/version/error/panic outcomes; added `web` Go `webassets` package with unconditional `//go:embed dist`; added `cmd/ycy` as the sole `os.Exit` owner. Verification: `GOTOOLCHAIN=go1.26.7 CGO_ENABLED=0 go vet ./...` and `go test ./...` passed; a `CGO_ENABLED=0` standalone build passed; its `--version` printed `0.0.0-dev` and `--help` exposed only global foundation behavior after startup validated all three embedded shells. The obsolete untracked root `node_modules` dependency tree was moved recoverably to the local Trash, leaving the active Go package pattern clean. Risk: no business leaf is registered by design; command-specific route behavior remains deferred to owning Gates. Next: add pinned Lefthook and safely testable hook lifecycle tooling.
- 2026-08-23: slice `pinned hook lifecycle` completed. Modified files: added isolated `tools/lefthook` Go tool module pinned to Lefthook v2.1.11, root `lefthook.yml`/`lefthook.rc`, and standard-library `tools/hookctl`. Verification: `GOTOOLCHAIN=go1.26.7 CGO_ENABLED=0 go test ./tools/hookctl` passed with disposable Git repositories covering approved 222-byte legacy replacement, idempotence, custom-hook byte preservation, local/global `core.hooksPath` refusal, linked worktrees, and non-restoring uninstall; the real pinned binary validated the policy; a separate disposable repository completed install, doctor, and uninstall using the generated hook wrapper. Risk: the current repository's legacy hook remains untouched until the final cutover lifecycle is run after all foundation checks are present. Next: add root Make lifecycle, active/legacy isolation and architecture checks, then rewrite root metadata and contributor documentation.
- 2026-08-23: slice `bootstrap reproducibility repair` completed. Modified files: added direct `@radix-ui/react-dropdown-menu` 2.1.24 and `uuid` 14.0.1 declarations to `web/package.json` and regenerated `web/pnpm-lock.yaml`. Evidence: after the retired root dependency tree was removed, `make bootstrap` exposed the two imports as undeclared; frozen pnpm installation then completed and Web typecheck, lint, and 27 tests passed. Risk: none; these are existing active-source imports and the lockfile is now self-contained. Next: verify the Make lifecycle and active-tree isolation end to end.
- 2026-08-23: slice `root lifecycle and active-tree isolation` completed. Modified files: added root `Makefile`, standard-library `tools/check-no-bun`, `cmd/ycy` architecture test, new `.gitignore`, Node/Go editor settings, root README, and CONTRIBUTING guidance. Verification: `make bootstrap` selected Go 1.26.7 and pnpm 11.13.0; `make hooks-install` replaced only the preflight-approved hook and passed its doctor; `make hooks-doctor`, `make check`, and `make build` passed. The Complete Gate passed Go vet/tests, tool-module and pnpm lock verification, active-tree isolation, Web lint/typecheck/Vitest, and Vite verification of 3 shells/424 reachable assets. Risk: Vite's Monaco chunk-size advisory remains informational; no contract check is weakened. Next: cross-build all six CGO-free targets, smoke the host artifact, and verify the staged tree from a disposable clean checkout.
- 2026-08-23: slice `generated-output hygiene` completed. Modified files: root `.gitignore`. Verification: root `.eslintcache` is now ignored by the active policy, absent from the eligible untracked set, and `git diff --check` passed. Risk: none; this retains a generated legacy ESLint cache outside the Cutover Commit. Next: stage only the Foundation archive and active foundation, then reproduce the required lifecycle from a disposable clean checkout of that staged tree.
- 2026-08-23: slice `archive verification recipe repair` completed. Modified files: `legacy/bun/ARCHIVE.md`. Verification: the documented staged-index comparison now excludes its two declared cutover-metadata files, so it compares exactly the 242 manifest rows rather than producing false metadata rows. Risk: none; archive content and manifest schema are unchanged. Next: re-run source/index identity verification, then create the disposable staged-tree checkout.
- 2026-08-23: slice `staged clean-checkout integration verification` completed. Modified files: none; verification used disposable local checkouts only. Evidence: the full-history staged-tree checkout verified all 242 manifest source/index path, mode, and blob mappings against `78358c0201b71891e36603d6abb8d7c87d54ad57`; the four frontend copy mappings and pre-adaptation byte comparisons remain recorded in `legacy/bun/ARCHIVE.md` and the earlier archive slice. In that checkout, `make bootstrap`, `make hooks-install`, `make hooks-doctor`, `make check`, `make build`, and `make cross-build` passed. The Complete Gate passed 27 Vitest tests, Go tests/vet, lock verification, active-tree isolation, and the Vite graph check for 3 shells and 424 reachable assets. The built host artifact reported `0.0.0-dev`, validated all embedded shells at startup, exposed only global CLI behavior, and passed no-argument/unknown-command exit checks. All six fixed artifacts were nonempty and inspected as Mach-O x64/arm64, static ELF x64/arm64, and PE x64/arm64; host metadata reported Go 1.26.7 with `CGO_ENABLED=0`, and `otool -L` showed only system libraries. Generated dependencies, Web output, maps, tool binaries, and build artifacts remained untracked. Risk: Vite's Monaco large-chunk advisory is informational; no required check failed. Next: record the integrated Foundation Unit and atomically transition only the Gate ledger.
- 2026-08-23: slice `Cutover Commit preflight` did not complete. Modified files: none. The actual Lefthook pre-commit passed `gofmt` and whitespace, then failed `web-eslint` because its staged-file glob passed intentionally ignored `web/verify-assets.mjs` to ESLint and `--max-warnings=0` rejected the resulting warning. No Cutover Commit was created; G0 remains `active` and G1 remains `planned`. Risk: the hook policy must match the Web lint configuration without excluding the verifier from meaningful lint coverage. Next: lint the verifier with only its required top-level-await exception, then re-run the hook and full applicable verification.
- 2026-08-23: slice `hook lint coverage repair` completed. Modified files: `web/eslint.config.js`. The generated-asset verifier is now linted by the standard Web lint command; only its required top-level-await rule is disabled. Verification: `pnpm --dir web run lint` passed; a new full-history staged-tree checkout again passed the 242-row archive identity check, `make bootstrap`, hook install/doctor, `make check`, `make build`, `make cross-build`, host smoke, and the actual Lefthook pre-commit with all three commands successful. Risk: none; lint coverage increased and no product behavior changed. Next: record final Foundation evidence and create the one Cutover Commit.
- 2026-08-23: final G0 acceptance completed. Exit 1: frozen archive, 242-row manifest/source identity, and four recorded frontend copy mappings have reproducible evidence. Exit 2: CLI, layout, unconditional three-shell embedding, hook lifecycle, architecture, and legacy isolation passed with no business placeholder. Exit 3: all Directed and Repository checks, host standalone smoke, and six cross-builds passed from the disposable staged-tree checkout. Exit 4: the Foundation Unit is `integrated` in `acceptance.md` with the Cutover Commit (this commit) and durable evidence. G0 is `passed`; G1 is activated but not started. Risk: none. Next: end this Goal; G1 requires a new Goal.

### G1: export env

- 2026-08-23: initialized as `planned`; no implementation evidence.
- 2026-08-23: activated as `active`; not started.
- 2026-08-23: slice `direct env-file discovery` completed. Modified files: `internal/commands/exportenv/discovery.go`, `internal/commands/exportenv/discovery_test.go`. Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/commands/exportenv` passed with inventory-derived empty/base-only/environment-only, exact-exclusion, direct-file, nested-path, and lexical-order cases; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: discovery is intentionally limited to direct files and does not read contents; the existing Vite large-chunk advisory remains informational. Next: add the tested dotenv parser and merge behavior without registering the command.
- 2026-08-23: slice `dotenv parsing and merge` completed. Modified files: `internal/commands/exportenv/parser.go`, `internal/commands/exportenv/parser_test.go`. Verification: the focused Go suite passed an inventory-derived `dotenv@17.4.2` grammar vector covering CRLF normalization, comments, single/double/backtick quotes, literal dollar references, `export`, Unicode, empty values, double-quote escapes, quoted multiline content, ignored malformed lines, and later-content override; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: parser behavior is deliberately command-owned and only processes supplied text; selection and file I/O remain unimplemented. Next: add deterministic selection and cancellation semantics without registering the command.
- 2026-08-23: slice `environment selection` completed. Modified files: `internal/commands/exportenv/selection.go`, `internal/commands/exportenv/selection_test.go`. Verification: the focused Go suite passed explicit-environment, merge ordering, exact missing-environment error, base-only automatic, environment-only prompt, sorted legacy labels, merge-hidden-base, and cancellation cases; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: the selector is a narrow command-owned Adapter seam and has no terminal implementation yet; no file content or output side effect occurs in this slice. Next: add ordered file reading and deterministic JSON publication without registering the command.
- 2026-08-23: slice `selected file reading` completed. Modified files: `internal/commands/exportenv/reader.go`, `internal/commands/exportenv/reader_test.go`. Verification: the focused Go suite passed exact absolute direct-file paths, selected-order reads, text preservation, and underlying read-error propagation; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: reading has no parsing, publication, or terminal behavior, and its only Adapter is command-owned. Next: encode parsed values as deterministic JSON without output side effects.
- 2026-08-23: slice `deterministic JSON encoding` completed. Modified files: `internal/commands/exportenv/json.go`, `internal/commands/exportenv/json_test.go`. Verification: the focused Go suite passed `JSON.stringify(..., null, 2)`-compatible key ordering, two-space formatting, no trailing newline, literal `<>&`, escaped newlines, U+2028 preservation, and literal-backslash safety; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: encoding has no output side effect and the only remaining publication behaviors are target resolution, file writing, and terminal presentation. Next: add output-file publication without registering the command.
- 2026-08-23: slice `output-file publication` completed. Modified files: `internal/commands/exportenv/writer.go`, `internal/commands/exportenv/writer_test.go`. Verification: the focused Go suite passed cwd-relative target resolution, absolute target preservation, byte publication, and underlying writer-error propagation; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: concrete overwrite/missing-parent/permission behavior remains for the final command-owned OS Adapter test; success presentation remains separate. Next: add stdout/file-success presentation without registering the command.
- 2026-08-23: slice `result presentation` completed. Modified files: `internal/commands/exportenv/presentation.go`, `internal/commands/exportenv/presentation_test.go`. Verification: the focused Go suite passed stdout success ordering, raw output-file success message, non-duplicated JSON, and exact cancellation message delegation without process exit; `make check` and `make build` passed, including 27 Web tests and Vite verification of 3 shells/424 reachable assets. Risk: the concrete terminal Adapter and full command composition are intentionally deferred; no CLI route is visible yet. Next: integrate the complete Module with its typed CLI binder and whole-binary tests.
- 2026-08-23: final G1 acceptance completed. Exit 1: focused inventory-derived discovery, parser/merge, selection/cancellation, filesystem, JSON, output, and typed-binder evidence passed in `go test ./internal/commands/exportenv ./internal/cliapp ./cmd/ycy`. Exit 2: the standalone `build/ycy` exposed `export env [dir]` with the required flags and retained plain `0.0.0-dev`; named merge/stdout output, cwd-relative file output, missing-file/missing-environment exit-1 errors, and EOF cancellation exit 0 all passed without legacy dispatch. Exit 3: `make check`, `make build`, current-host smoke, and `make cross-build` passed; the host candidate SHA-256 is recorded in `acceptance.md`, and six nonempty Mach-O/ELF/PE artifacts were inspected. The `export env` Unit is `integrated` in `acceptance.md` with this Integration Commit and outstanding native evidence. G1 is `passed`; G2 is activated but not started. Risk: Vite's existing Monaco large-chunk advisory remains informational. Next: end this Goal; a new Goal owns G2.

### G2: appconfig foundation

- 2026-08-23: initialized as `planned`; no implementation evidence.
- 2026-08-23: activated as `active`; not started.
- 2026-08-23: slice `schema/read normalization` completed. Modified files: added `internal/appconfig/store.go`, `internal/appconfig/schema.go`, and `internal/appconfig/schema_test.go`. Verification: `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test ./internal/appconfig`, `make check`, and `make build` passed; the Complete Gate retained 27 Web tests and verified 3 Vite shells with 424 reachable generated assets. Risk: only private normalized document state exists; no crypto, locking, publication, semantic caller surface, or public CLI route has been added. Next: add deterministic PBKDF2/AES-GCM compatibility vectors without introducing a write path.
- 2026-08-23: slice `deterministic crypto vectors` completed. Modified files: added `internal/appconfig/crypto.go` and `internal/appconfig/crypto_test.go`. Verification: focused tests proved PBKDF2-HMAC-SHA256 at 100,000 iterations and AES-256-GCM with a 16-byte IV match the fixed legacy key and `base64(iv):base64(tag):base64(ciphertext)` vector, including Unicode plaintext; malformed serialization rejects. `make check` and `make build` passed. Risk: crypto remains internal and unbound to machine identity or persistence; no public CLI route has been added. Next: add platform machine-ID parsing and config-path precedence separately.
- 2026-08-23: slice `machine identity and path precedence` completed. Modified files: updated `internal/appconfig/store.go`; added `internal/appconfig/machineid.go`, platform-specific `machineid_darwin.go`, `machineid_linux.go`, `machineid_windows.go`, and `machineid_test.go`. Verification: focused tests proved unconditional `USERPROFILE`, then `HOME`, then user-home path precedence; macOS/Linux/Windows machine-ID parsing; hostname/username fallback; and a nonempty native macOS `ioreg` adapter result without recording the identifier. `make check` and `make build` passed. Risk: identity is private and no lock, write, semantic caller surface, or public CLI route exists yet. Next: add directory-lock ownership, timeout, stale recovery, and owner-safe release.
- 2026-08-23: slice `lock ownership and recovery` completed. Modified files: updated `internal/appconfig/store.go`; added `internal/appconfig/lock.go`, `process_unix.go`, `process_windows.go`, and `lock_test.go`. Verification: focused tests proved directory-create serialization between independent callers, live-owner timeout, ownerless publication grace, stale/dead-owner rename-then-remove recovery, malformed-owner rejection, UUID-safe release, and current-process liveness. `make check` and `make build` passed. Risk: no configuration contents are written yet, so candidate cleanup/replacement and semantic concurrency remain for later slices; no public CLI route has been added. Next: add atomic config publication, mode requests, and failure cleanup.
- 2026-08-23: slice `atomic publication and cleanup` completed. Modified files: updated `internal/appconfig/store.go` and `internal/appconfig/schema.go`; added `internal/appconfig/publish.go`, `publish_test.go`, `replace_unix.go`, and `replace_windows.go`. Verification: focused tests proved same-directory candidate publication, trailing newline, cleanup after successful and injected replacement failure, target preservation on failure, private Unix mode requests, lock cleanup, and next-write unknown-field discard. Cache-only `go build ./internal/appconfig` passed for Darwin, Linux, and Windows; `make check` and `make build` passed. Risk: no semantic Fork/CM/Tunnel caller surface exists yet, so concurrent field updates remain unproven; no public CLI route has been added. Next: add the sole semantic operations and their concurrent-update evidence.
- 2026-08-23: slice `semantic Fork/CM/Tunnel operations` completed. Modified files: updated `internal/appconfig/store.go` and `schema.go`; added `ordered.go`, `operations.go`, `fork.go`, `cm.go`, `tunnel.go`, and `operations_test.go`. Verification: uncached focused tests proved direct decryption of fixed Bun-written current/legacy Fork, CM, and Tunnel ciphertext; insertion-order preservation; safe list projections; encrypted mutation; CM defaults/environment precedence; concurrent independent Fork/CM updates; Tunnel ID, sort, cap, and corrupt-entry isolation. `make check` and `make build` passed. Risk: only G2's architecture/no-public-command proof and Gate-close common evidence remain; no public CLI route has been added. Next: run the Gate-close architecture, standalone smoke, and six-target verification before recording the Integration Unit.
- 2026-08-23: final G2 acceptance completed. Exit 1: current and legacy config documents, including fixed Bun-written encrypted Fork, CM, and Tunnel values, directly normalize and decrypt through `appconfig`; schema, PBKDF2/AES-GCM, ordering, and safe projection tests passed. Exit 2: directory locking, stale recovery, publication grace, owner-safe release, concurrent independent semantic updates, candidate cleanup, private mode requests, and macOS native machine-ID/lock/replacement tests passed; Linux/Windows production paths compile and their native execution evidence remains recorded as a later release-acceptance obligation. Exit 3: architecture enforcement reserves `config.json`/`.config.lock` to `internal/appconfig`; `--help` omits `config` and standalone `ycy config` returns the expected unknown-command exit-1 outcome. Exit 4: `go test -count=1 ./internal/appconfig ./internal/cliapp ./cmd/ycy`, `make check`, `make build`, host smoke, and `make cross-build` passed; host SHA-256 and six Mach-O/ELF/PE inspections are recorded in `acceptance.md`. G2 is `passed`; G3 is activated but not started. Risk: Vite's existing Monaco large-chunk advisory remains informational. Next: end this Goal; G3 requires a new Goal.

### G3: config fork list

- 2026-08-23: initialized as `planned`; no implementation evidence.
- 2026-08-23: activated as `active`; not started.

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
