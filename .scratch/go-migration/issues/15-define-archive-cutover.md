# Define the legacy archive and migration cutover choreography

Type: grilling
Status: resolved
Blocked by: 10, 11, 13, 14

## Question

Choose the exact sequence for freezing and moving the Bun implementation to `legacy/bun/`, extracting the active React code into `web/`, establishing the Go composition root and modules, replacing root package-manager metadata and hooks, keeping intermediate commits locally buildable, preventing runtime dispatch into legacy, and withholding release until the final compatibility gate. Define which files move, which are rewritten, which generated artifacts remain absent, and how each step proves it did not lose reference material.

## Comments

- 2026-08-22, grilling round 1: selected a complete frozen `legacy/bun/` copy of the current frontend and backend for code reference only; rejected parallel foundations in favor of moving Bun immediately and filling in Go incrementally; approved unchanged archival disablement of the current GitHub workflows, Dockerfile, Docker ignore file, and deployment definitions; rejected additional release-lock machinery, relying on no migration-time tags/releases and a manually triggered replacement workflow only after migration completes.
- 2026-08-22, grilling round 2: selected one buildable atomic cutover commit that both archives Bun immediately and establishes the minimal Go, Vite, Make, pnpm, and Lefthook foundations before command-by-command work; selected the exact move/retain/rewrite ledger; selected four explicit legacy-to-active frontend copy mappings while keeping Ink and Bun workers reference-only; selected a source-commit and Git blob/mode manifest plus one-time Bun baseline and byte-for-byte archive verification, with no legacy execution in the active suite.
- 2026-08-22, grilling round 3: selected an active CLI containing only commands already migrated to Go, with no placeholders or legacy fallback; selected exact-known-hook replacement with Lefthook before the atomic cutover commit and a hard stop for any custom or unknown hook; selected ignored, untracked generated Web output, dependencies, binaries, maps, caches, downloaded payloads, and release artifacts; selected mechanically enforced isolation that forbids active imports, execution, symlinks, Make targets, packaging, generation, and runtime dispatch involving `legacy/bun/`, while retaining only the already-approved `ycy run` behavior that may invoke Bun in a user's project.
- 2026-08-22, final confirmation: approved the complete archive ledger, atomic cutover sequence, evidence and hook procedure, active/legacy isolation, generated-output policy, incremental command exposure, and release hold without further changes.

## Answer

Use one **Cutover Commit** as the first Go-era commit. It moves the complete tracked Bun implementation into a **Frozen Archive** at `legacy/bun/` and, in the same commit, creates the smallest active Go/Vite/pnpm/Make/Lefthook tree that passes its own build and checks. The **Active Tree** is every repository path outside `legacy/bun/`. The Frozen Archive is human-readable implementation evidence only: neither the product nor the active development lifecycle may depend on it.

### 1. Capture the Bun baseline before moving anything

Start from source commit `78358c0201b71891e36603d6abb8d7c87d54ad57`, root package version `0.0.69`, and Bun `1.3.14`. Do not tag or publish from that point onward. Before changing paths:

1. Run the existing Bun test, typecheck, lint, and build probes that are available at the frozen commit and record each command, runtime version, pass/fail/skip result, and any required platform limitation without repairing failures or changing behavior.
2. Generate `legacy/bun/MANIFEST.tsv` from the source Git tree. Each row records the original path, destination archive path, Git mode, and blob object ID for every archived tracked file. Generated outputs and installed dependencies are not inputs to this manifest.
3. Add `legacy/bun/ARCHIVE.md` recording the source commit, package and Bun versions, baseline results, archive/copy ledger, manifest schema, and reproducible verification commands. This evidence file and `MANIFEST.tsv` are cutover metadata, not claimed as files from the frozen source.
4. Preserve the old root README byte-for-byte as `legacy/bun/README.md`; the new active root README is a rewrite.

A baseline failure is recorded parity evidence, not permission to redesign the command and not a reason to omit its source. The cutover is blocked only if the source identity, manifest, or complete archive cannot be captured.

### 2. Apply the exact archive, retain, and rewrite ledger

Move these tracked paths with history into the same relative locations below `legacy/bun/`:

```text
src/**
scripts/build.ts
scripts/evaluate-git-cm.ts
scripts/generate-frp-manifest.ts
scripts/install-tunnel-frp.ts
scripts/prepare-seven-zip.ts
package.json
bun.lock
bunfig.toml
eslint.config.js
tsconfig.json
types.d.ts
.github/workflows/**
Dockerfile
.dockerignore
deploy/**
```

The archived workflows, Dockerfile, Docker ignore file, and deployment definitions change location only; their content remains byte-for-byte unchanged. Because GitHub, Docker, and deployment tooling no longer find them at their active root paths, obsolete Bun automation is disabled immediately without pretending it has already been redesigned for Go.

Keep these paths active rather than moving them:

```text
scripts/install.sh
scripts/install.ps1
LICENSE
CLAUDE.md
public/**
mock/**
.github/logo.ico
.github/logo.png
```

Rewrite or create the active root README and `CONTRIBUTING.md`, `.gitignore`, `.vscode` settings, root Go module files, `cmd/ycy` composition root, only the shared/internal packages needed by the foundation, root Make lifecycle, Lefthook policy and pinned tool module, `hookctl`, and the one `web/` pnpm/Vite plus Go `webassets` package. Follow the already-selected lazy module rule: do not create empty command packages, placeholder handlers, generic layers, or speculative Interfaces.

### 3. Extract the active React applications from the frozen copy

After `src/**` is present in the Frozen Archive, copy source through these four explicit mappings:

```text
legacy/bun/src/commands/diff/web/**          -> web/diff/**
legacy/bun/src/commands/fs/web/**            -> web/fs/**
legacy/bun/src/commands/tunnel/server/web/** -> web/tunnel-server/**
legacy/bun/src/shared/web/**                  -> web/shared/**
```

The archive-side files remain untouched after the copy. Adapt only the active copies to the selected Vite MPA structure, three physical HTML inputs, Vite-owned workers and shared assets, package-local Antfu ESLint, TypeScript, and Vitest configuration. Ink interfaces, Bun server workers, Bun build plugins, and other backend/runtime code stay reference-only in `legacy/bun/`; they are not copied into the Active Tree.

The initial composition root exposes only real root/global behavior supplied by the new CLI foundation. A business command enters the fixed manifest only when its Go implementation and focused parity tests are complete. Until then it is absent: there is no placeholder, "not migrated" response, hidden legacy alias, subprocess fallback, or runtime switch into Bun. The three Web applications are nevertheless built and unconditionally embedded in every intermediate ycy binary, even before all three owning commands are registered.

### 4. Keep generated and acquired material outside Git

Track reproducibility inputs such as `go.mod`, every required `go.sum`, `web/package.json`, `pnpm-lock.yaml`, Vite/TypeScript/ESLint/Vitest configuration, asset verification code, and source assets. Ignore and never add:

- `web/dist`, `web/node_modules`, root legacy `node_modules`, coverage, browser-test reports, ESLint caches, temporary directories, and tool/download caches;
- generated JavaScript/CSS/assets and source maps;
- host or six-target ycy binaries, `SHA256SUMS`, archives, package bundles, and other release staging output;
- downloaded or extracted 7-Zip and FRP payloads and any other runtime-acquisition cache.

The Frozen Archive contains tracked source and cutover evidence, not copied dependency trees, compiled Bun executables, historical `dist` directories, caches, or downloaded payloads. `make build` always creates and verifies `web/dist` before any Go compilation that reaches the unconditional `go:embed`; raw clean-checkout Go commands are not an alternative build mode.

### 5. Replace the old hook before committing the cutover

Once the new hook lifecycle exists in the cutover worktree, run the selected `hookctl` path before creating the Cutover Commit. It resolves the effective Git/common directory and hook path, then may remove only the approved 222-byte simple-git-hooks `pre-commit` whose SHA-256 is `7bc48fcc880a58ab4f92dbe45343a82eea1b2539c86e5c05dc6713d39bdf5d95`, or another exact approved legacy template recorded by the hook-policy decision. It installs the pinned root Lefthook policy offline and runs the doctor.

Any extra line, custom hook, unknown manager, unknown hook path, or unapproved local/global `core.hooksPath` state is preserved byte-for-byte and hard-stops the commit with recovery guidance. The cutover does not overwrite, merge, disable, or silently bypass custom policy. No pnpm lifecycle installs hooks, and no Bun hook remains active after a successful replacement.

### 6. Prove the Cutover Commit before recording it

Stage the archive and foundation together and create no intermediate archival or scaffolding commit. Before committing:

1. Verify every `MANIFEST.tsv` row against the source commit and its mapped `legacy/bun/` path; source blob ID and mode must match exactly, and no required row may be missing or duplicated.
2. Verify the four frontend copy mappings were sourced from the archived files and that all later Vite adaptations occurred only in `web/`.
3. Run `make bootstrap`, the hook doctor, the offline non-mutating `make check`, and `make build`; smoke-test the actually registered CLI surface and cross-build the minimal `CGO_ENABLED=0` binary for all six required targets.
4. Run the architecture and `check-no-bun` checks. Outside allowed migration documentation, they reject active imports, execution, generation, packaging, Make targets, runtime dispatch, and symlinks that reference or resolve through `legacy/bun/`. They also reject Bun/package-manager residue outside the Frozen Archive.
5. Permit only the separately specified `ycy run` code/tests/fixtures that detect a Bun lockfile in a user's project and invoke that user's Bun executable. This allowance never names, opens, or executes `legacy/bun/` and does not make Bun an active ycy build dependency.
6. Confirm with the Git index that ignored generated outputs, dependencies, binaries, maps, caches, downloaded payloads, and release artifacts are untracked. Reproduce bootstrap, checks, and build in a clean disposable checkout of the staged tree.

Only after all six checks pass may the one Cutover Commit be created. The active suite never executes the archived Bun CLI as an oracle; future parity tests encode behavior by consulting the frozen source.

### 7. Migrate incrementally without releasing

Every later command commit remains locally buildable under the same frontend-first Complete Gate. For each command, consult its inventory and `legacy/bun/`, add the focused parity tests, implement the Go behavior, and only then add its typed handler to `cliapp` and compose it in `cmd/ycy`. A failed parity test or implementation probe may open the narrow compatibility-exception decision allowed by the first-release policy; known defects and desired hardening do not alter this choreography.

No migration-time tag, GitHub release, active release workflow, Docker build, or deployment is produced, and no additional Release Lock file, CI guard, sentinel version, or tag-blocking mechanism is added. The absence of active workflows plus the maintainer's no-tag/no-release policy is the migration hold.

After every command and the final browser, protocol, persistence, native-platform, installer/update, embedded-payload, and six-artifact compatibility gates pass, create a new Go-era release workflow that supports manual dispatch only. Do not restore the archived Bun workflow. The first tag and manual release occur only after that workflow and the complete release-artifact gate are ready. Archived Docker and deployment definitions remain reference-only until a separately scoped redesign or replacement is approved.
