# GitHub CLI-style ycy project structure refactor

Label: wayfinder:map
Status: resolved

## Destination

Produce an execution-ready structural refactor specification for ycy that
reorganizes the active Go CLI and repository support areas in the style of
GitHub CLI: a thin `cmd/ycy` composition root, command packages grouped by
domain under a `pkg/cmd`-shaped tree, shared implementation Modules under
`internal`, and clear package-local versus top-level acceptance evidence.
The route must preserve all command behavior, public CLI syntax, output,
exit semantics, build behavior, and existing web/generated-artifact rules.

## Notes

- Domain: Go CLI repository architecture, command-tree organization, package
  ownership, test topology, build support, and migration mechanics.
- The user approved the GitHub CLI structural model, including its `pkg/cmd`
  command layout. This adopts the layout semantics; it does not create a
  separate promise of stable Go API compatibility.
- `cmd/ycy` remains the only active binary composition root. Command packages
  should be organized by command domain and leaf, with short, discoverable
  filenames rather than the current long command-prefixed filenames.
- Shared behavior belongs in an evidence-backed `internal` Module owned by
  the narrowest responsible context. Do not invent generic `utils`,
  `services`, or pass-through layers.
- `legacy/bun` remains a read-only behavioral reference. `web`, `mock`,
  `tools`, `build`, and generated outputs retain their current build and
  ignore rules; this effort only clarifies their repository roles unless a
  later decision proves a mechanical move necessary.
- Tests remain colocated where the GitHub CLI pattern makes ownership clear,
  with cross-package, black-box, PTY, and platform acceptance evidence in a
  top-level acceptance-style area.
- Consult `research` for upstream repository facts, `domain-modeling` when
  terms or ownership boundaries change, and `codebase-design` for deep
  Module seams. This map plans structure only; implementation begins after
  the map is resolved.

## Decisions so far

<!-- Closed child tickets are indexed here by name. -->

- [Research the GitHub CLI project structure](issues/01-research-github-cli-project-structure.md): upstream evidence confirms a thin `cmd/gh`, noun/verb command packages under `pkg/cmd`, private shared Modules under `internal`, colocated leaf tests/fixtures, and separate acceptance/build/script/docs areas; the report is in [`github-cli-structure.md`](research/github-cli-structure.md).
- [Research the current ycy tree and ownership](issues/02-research-ycy-current-tree-and-ownership.md): the active tree has 30 Go packages, a 77-file flat `cmd/ycy`, a per-leaf `internal/cliapp` handler graph, clean command-domain imports, and eleven path-sensitive migration risks including an existing architecture-test ban on `pkg`; the inventory is in [`02-ycy-current-tree-and-ownership.md`](research/02-ycy-current-tree-and-ownership.md).
- [Choose command package visibility and the composition root](issues/03-choose-command-package-visibility-and-composition-root.md): adopt `cmd/ycy -> internal/ycycmd -> pkg/cmd/root -> pkg/cmd/<domain>/<leaf>`, use Factory/Options/`runF`, delete `internal/cliapp`, and enforce one-way parent/child plus `internal` dependency rules without promising a Go SDK.
- [Choose the domain command tree and file naming rules](issues/04-choose-domain-command-tree-and-file-names.md): mirror every existing CLI token in `pkg/cmd`, move command-specific implementation/Adapters/tests into leaf directories, use short responsibility filenames, and keep `cm`, `fs`, `rm`, and all other command vocabulary unchanged.
- [Choose shared Module seams and dependency direction](issues/05-choose-shared-module-seams-and-dependency-direction.md): preserve named deep `internal` Modules, add an owner-local `internal/gitprocess`, use a bounded capability Factory, keep Web as a root package, colocate platform files with owners, and enforce one-way imports.
- [Choose the repository support-area layout](issues/06-choose-repository-support-area-layout.md): add top-level `acceptance/` and `docs/project-layout.md`, retain `scripts/`, `web`, `mock`, `legacy/bun`, `tools`, `build`, `public`, and all generated-output rules with explicit ownership.
- [Choose the test, fixture, and acceptance topology](issues/07-choose-test-fixture-and-acceptance-topology.md): colocate package and Module evidence, move cross-process/black-box/PTY/native checks to tagged top-level `acceptance/`, keep Web implementation tests in `web/`, and scope fixtures by the seam they exercise.
- [Prototype the target tree and migration slices](issues/08-prototype-target-tree-and-migration-slices.md): accept the concrete final tree, exact bounded Factory, four worker/runtime Modules, root-first no-shim transition, file/test destinations, and seven gated migration slices in the linked prototype.
- [Approve structural migration gates and no-behavior evidence](issues/09-approve-structural-migration-gates.md): freeze the complete CLI surface, gate every leaf and Slice, stop on any behavioral diff, remove all transition scaffolding, and require recorded Darwin/Linux/Windows native evidence without adding CI.

## Not yet specified

<!-- No remaining fog. All structural decisions required for execution are resolved. -->

## Out of scope

- Changing command names, flags, arguments, prompts, output, exit codes,
  business rules, side effects, or error semantics.
- Refactoring business behavior, command semantics, or external contracts
  under the pretext of moving files; the approved internal command Interface
  refactor to Factory/Options/`runF` is in scope, while dependency upgrades
  and feature work remain separate efforts.
- Rebuilding the `web` applications, changing their package layout, or
  changing `web/dist`, 7-Zip payload, `build`, or other generated-artifact
  policies.
- Re-enabling or deleting `legacy/bun`, changing mock service behavior, or
  redesigning deployment/release workflows.
- Introducing a new top-level command, dashboard, plugin system, or external
  Go SDK.
