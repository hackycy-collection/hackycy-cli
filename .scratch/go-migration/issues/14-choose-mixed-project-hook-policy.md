# Choose the mixed-project Git hook policy

Type: grilling
Status: resolved
Blocked by: 02

## Question

Based on the mixed-project quality-gate research, choose the tracked hook mechanism, installation and stale-hook cleanup procedure, staged-file behavior, fast Go and pnpm/Vite checks, separation from full local verification, Windows support, dependency-not-installed behavior, bypass semantics, and contributor documentation. The selected hook must cover both stacks and contain no Bun or legacy frontend hook invocation.

## Comments

- 2026-08-22, grilling round 1: selected root-owned Lefthook v2.1.11 pinned in an isolated Go tool module; selected a parallel, non-mutating, staged-path-aware `Fast Gate`; selected `make check` as the aggregate `Complete Gate` covering Go, web, lock consistency, and active-tree Bun exclusion.
- 2026-08-22, grilling round 2: selected exact-match-only removal of the known Bun/simple-git-hooks hook with a hard stop on custom hooks or unknown hook paths; selected explicit and separate bootstrap/install/doctor/uninstall commands; selected offline fail-closed hook execution with stack-specific prerequisites; retained documented Git and Lefthook bypasses; and made native Git for Windows behavior part of migration acceptance.
- 2026-08-22, grilling round 3: retained Antfu ESLint as the sole web lint/format engine; selected staged-path selection with non-mutating worktree-content checks for partially staged files; limited Bun allowances to frozen legacy, documented migration material, and `ycy run` compatibility code/tests; and required the bootstrapped Complete Gate to run offline without modifying tracked files or the Git index.
- 2026-08-22, grilling round 4: selected a purpose-built, standard-library-only Go `hookctl` behind the Make hook lifecycle targets, excluded from release artifacts and tested in disposable repositories; selected root `CONTRIBUTING.md` as the canonical contributor policy with a README pointer and concise `make help` discovery.
- 2026-08-22, grilling round 5: confirmed both the Go `hookctl` implementation and the root `CONTRIBUTING.md` documentation contract.
- 2026-08-22, grilling round 6: selected one Grafana-style repository-root Lefthook structure and rejected parallel task-runner, hook-manager, or Windows-only command surfaces. Root Make targets remain the single contributor lifecycle interface, so GNU Make is an explicit Windows development prerequisite; installed hooks invoke the pinned root Lefthook policy directly rather than invoking Make.
- 2026-08-22, final confirmation: approved the complete policy without further changes.
- 2026-08-22, build-order amendment: [Prove the Vite MPA to Go embed path](11-prove-vite-go-embed-path.md) selected unconditional embedding. This supersedes only the earlier assumption that `check-go` can compile independently of generated Web assets; the Fast Gate and all other hook-policy decisions remain unchanged.

## Answer

Adopt one Grafana-style, repository-root Lefthook structure. Root `lefthook.yml` is the only active Git quality-gate definition; Lefthook v2.1.11 is pinned in the isolated `tools/lefthook/go.mod`/`go.sum` tool module and exposed to generated hooks through root `lefthook.rc`. Do not add Husky, lint-staged, simple-git-hooks, a second task runner, a second hook manager, a Windows-specific command surface, or a repository `pre-push` gate. `web/package.json` owns frontend commands only and has no hook-install lifecycle.

The only active hook is a parallel, check-only `pre-commit` **Fast Gate**:

1. Every commit runs `git diff --cached --check`.
2. Existing staged active `.go` paths, excluding `legacy/bun/**`, are checked with `gofmt` and fail with the offending paths plus `make fmt` guidance.
3. Existing staged JavaScript/TypeScript React and configuration paths under `web/` are checked by the package-local Antfu ESLint with zero warnings. Antfu ESLint remains the sole web lint/format engine; no Prettier is added.
4. Lefthook owns staged-path filtering, deletion filtering, quoting, command-line chunking, the `web/` working root, and parallel scheduling. A partially staged path selects the job, but the job checks the current worktree file. This is deliberately fast feedback rather than proof of the exact index blob.

The Fast Gate never runs `gofmt -w`, ESLint `--fix`, `stage_fixed`, `git add`, a stash, or any other index/worktree mutation. It never runs Go vet/tests, TypeScript project checking, Vitest, Vite build, dependency installation, module mutation, or network access. Go-only commits do not require Node, pnpm, or `web/node_modules`; a web change with missing prerequisites fails closed with exact `make bootstrap` remediation rather than skipping or installing anything. The hook-side Lefthook wrapper uses the pinned Go tool with `GOWORK=off` and `GOPROXY=off`.

Root Make targets form the single contributor lifecycle Interface:

- `make bootstrap` verifies the committed Go, Node, and pnpm requirements, downloads product/tool Go modules, warms the pinned Lefthook tool, and performs `pnpm --dir web install --frozen-lockfile`. It may access the network but does not install hooks implicitly.
- `make hooks-install` invokes the repository-owned, standard-library-only Go `hookctl`, safely removes only a recognized legacy hook, installs pinned Lefthook offline, then runs the doctor.
- `make hooks-doctor` performs read-only reporting of Git roots and hook paths, manager identity, pinned tool/runtime prerequisites, dependency readiness, and active Bun/obsolete-manager residue.
- `make hooks-uninstall` removes only Lefthook-managed hooks and verifies that no old Bun hook was resurrected. It does not restore legacy hooks or modify unknown/global hook policy.
- `make fmt` is the explicit mutating Go/ESLint-fix path. Contributors run it and choose what to stage themselves.
- `make check` is the offline, non-mutating **Complete Gate** after bootstrap; `make build` remains the Vite-first, Go-embed product build.

`make check` aggregates stable `check-go`, `check-web`, `check-locks`, and `check-no-bun` targets, with the Vite production build and structured asset-graph verification completing before any Go package that unconditionally embeds `web/dist` is compiled. `check-go` verifies all active Go formatting and then runs `CGO_ENABLED=0 go vet ./...` plus `CGO_ENABLED=0 go test ./...` against that verified output; it is not an independently runnable clean-checkout compile gate. `check-web` runs package-local ESLint, `tsc --noEmit`, `vitest run`, the Vite MPA production build, and asset-graph verification. `check-locks` verifies the product and isolated-tool Go modules without rewriting them and validates the pnpm frozen lock contract offline. No Complete Gate target may change tracked files or the Git index; ignored `web/dist` is generated build output, and missing caches fail with bootstrap guidance.

`check-no-bun` uses a narrow, reviewed allowance rather than banning the text `bun`: it permits the frozen `legacy/bun/**` tree, migration documentation, and the production/tests/fixtures needed for `ycy run` to recognize and execute Bun in a user's project. It rejects Bun dependencies or invocations in active root/web manifests, Lefthook-generated content, Make targets, frontend sources, build/release scripts, and every other active toolchain path. It also rejects simple-git-hooks, lint-staged, obsolete Bun lockfiles, and pnpm lifecycle ownership of hooks outside legacy/reference material.

The purpose-specific Go `hookctl` is migration/lifecycle tooling, not another quality-gate runner and not part of any ycy release artifact. It resolves the repository root and common Git directory through Git, inspects effective local/global `core.hooksPath` sources, and handles linked worktrees. The current 222-byte simple-git-hooks `pre-commit` (`SHA-256 7bc48fcc880a58ab4f92dbe45343a82eea1b2539c86e5c05dc6713d39bdf5d95`) is an approved legacy template. Only an exact approved template/hash may be removed automatically. Extra lines, an unknown manager, a custom hook, or an unknown hook path are preserved byte-for-byte and cause a guided hard stop. Only an exact known stale ycy-owned local `core.hooksPath` may be unset; arbitrary local and all global hook policies are never silently changed. Reinstallation is idempotent.

Native Windows support means Git for Windows plus the committed Go/Node/pnpm versions and GNU Make as an explicit contributor prerequisite. The hook itself invokes the pinned Lefthook policy directly, does not invoke Make, requires neither PowerShell nor WSL, and relies on Lefthook rather than shell loops for filenames and argument limits. Acceptance covers macOS, Linux, and Windows with spaces, Unicode, renames, deletions, large staged sets, LF behavior, custom hook paths, and linked worktrees.

Git's `git commit --no-verify` and one-operation `LEFTHOOK=0` bypasses remain documented. A bypass is not a passed gate; later CI may call the same Complete Gate, but CI wiring remains out of scope now. Root `CONTRIBUTING.md` is the canonical documentation for prerequisites, bootstrap, hook lifecycle, Fast/Complete Gate semantics, partial staging, bypass, custom-hook recovery, and Windows setup; root README links to it and `make help` provides concise discovery.

Implementation acceptance uses disposable Git repositories/worktrees only. It must prove clean install/reinstall/uninstall, exact legacy removal, byte-preserving refusal of modified/custom hooks, local/global hook-path policy, no Bun-generated hook content, path filtering/chunking, partial-stage non-mutation, offline missing-dependency failures, bypass behavior, every Complete Gate failure independently, and native execution on all three development operating systems. No test may mutate the developer's real Git hooks.
