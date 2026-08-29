# Choose the test, fixture, and acceptance topology

Type: grilling
Status: resolved
Blocked by: 01, 02, 03, 04, 05, 06

## Comments

- Claimed for the test, fixture, and acceptance topology decision on 2026-08-29.

## Resolution

Adopt a four-level evidence topology:

1. Package-local tests stay beside the implementation they exercise. Leaf
   command tests move with their `pkg/cmd/<domain>/<leaf>` package, root and
   command-tree tests move with `pkg/cmd/root`, and shared Adapter/Module
   tests move with the owning `internal` package. These tests may exercise
   the package's public Interface and its internal seams, but do not build a
   standalone binary unless that is the behavior under test.
2. Top-level `acceptance/` owns tests that cross package seams or process
   boundaries: black-box command journeys, standalone-binary checks,
   cross-package integration, CLI-level PTY interaction, signal handling,
   and native-platform acceptance. The package is excluded from ordinary
   `go test ./...` by a uniform `//go:build acceptance` constraint.
3. Web implementation tests stay in `web/`: TypeScript/Vitest tests cover
   browser-side behavior, while Go tests cover embedding, static routes, and
   resource graphs. A browser-driven journey that launches the real ycy or
   HTTP service belongs under `acceptance/web/`; it uses the existing
   `tools/web-browser-harness` as a tool, not as the owner of assertions.
4. Test-only infrastructure remains in its owner package. In particular,
   `internal/terminaltest` remains the PTY and terminal Adapter package for
   package tests; CLI-level PTY evidence is kept in `acceptance/`.

The existing `cmd/ycy` tests are moved according to the seam they cross, with
no duplicate long-term copies: command-local assertions follow the leaf or
root package; process and Adapter assertions follow their shared Module;
`*_integration_test.go`, standalone-binary tests, and CLI signal/PTY tests
move to `acceptance/`. A temporary path-only compatibility location is
allowed only within one migration slice and must be removed before that
slice's gate passes.

Fixture ownership follows the same rule. A fixture used by one command or
Module lives in that package's `testdata/`. Cross-command scenarios that
need a real repository, configuration, HTTP service, or binary live in
`acceptance/testdata/<scenario>/`. Tests continue to create disposable state
with `t.TempDir()`; a top-level fixture is introduced only when at least two
acceptance scenarios genuinely share it. Fixtures are not copied merely to
preserve an old path.

Acceptance execution is explicit. All Go files in `acceptance/` carry the
`acceptance` build tag; Unix/Windows differences use normal Go filename
suffixes. Ordinary `go test ./...`, `go vet ./...`, and `make check` retain
their package-level scope. The implementation adds explicit Make targets (at
minimum `acceptance`, with focused standalone, PTY, and native targets as
useful) that pass `-tags acceptance` and select subsets with `-run` rather
than inventing an uncontrolled collection of required tag combinations.

There is no current generated-document pipeline to move. If one is added,
the generator and deterministic tests belong in `tools/<generator>/`; its
outputs stay in the existing generated `build/` or release boundaries and
are not package fixtures. Any CLI-help/document parity check is an explicit
tool or acceptance check. This decision does not introduce generation now.

Platform-specific tests remain next to the narrowest Module they verify
(for example, `process_unix_test.go` beside its process implementation).
Only tests whose subject is the assembled binary or a native user journey
belong under `acceptance/`, with platform suffixes and an explicit command.

These rules preserve current assertions, machine-readable output, and test
semantics; structural moves may require only import-path, build-target, or
working-directory updates.

## Question

Decide how tests and fixtures should be arranged after the structural move.
Define the boundary between package-local tests, command-level black-box
tests, cross-package integration tests, PTY/native-platform evidence,
standalone-binary checks, web-browser harnesses, and generated documentation
tests. Decide when a fixture follows a command package and when it belongs in
a shared top-level testdata area. Preserve the current test commands and
machine-readable evidence unless a path-only build update is required.

The answer must yield stable ownership rules so moving a file does not orphan
its evidence or duplicate fixtures.
