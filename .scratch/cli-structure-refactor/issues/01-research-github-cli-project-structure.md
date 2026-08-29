# Research the GitHub CLI project structure

Type: research
Status: resolved

## Question

Using the official `cli/cli` repository at
https://github.com/cli/cli as the primary source, document the structural
conventions that are relevant to ycy: the role of `cmd/gh`, the split between
`pkg/cmd`, `internal`, `api`, `context`, `acceptance`, `script`, `test`, and
`utils`, command noun/verb directory patterns, shared-package ownership,
test/fixture placement, generated documentation, and build/release support.

Distinguish intentional architecture from repository-specific scale or
GitHub-only concerns. Record concrete paths and source links in
`.scratch/cli-structure-refactor/research/github-cli-structure.md`.
Do not modify production code.

## Comments

- Research completed against `cli/cli` trunk commit
  `9b323de8005a9988f398ce547697bf43b944e505` using first-party source and
  maintenance documentation only.

## Answer

The report is recorded in
[`github-cli-structure.md`](../research/github-cli-structure.md). GitHub CLI
keeps `cmd/gh` thin, places command packages under `pkg/cmd/<noun>/<verb>`,
uses `internal` for project-private Modules, and separates `acceptance`,
`script`, `build`, `docs`, and generated support from command code. Leaf
packages keep constructor/options/run logic, tests, and fixtures together;
cross-command sharing uses explicit `shared` packages and injected Factory
dependencies. Acceptance uses a build tag and top-level testdata so ordinary
unit tests stay offline. The report maps these conventions to ycy without
changing behavior or requiring a new Go API promise.
