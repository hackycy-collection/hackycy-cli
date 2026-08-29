# Choose the repository support-area layout

Type: grilling
Status: resolved
Blocked by: 01, 02, 03

## Comments

- Claimed for the repository support-area decision on 2026-08-29.

## Answer

Adopt the following support-area layout and roles:

```text
cmd/ycy/          thin binary entry only
pkg/cmd/          command packages
pkg/cmdutil/      bounded command Factory and shared command utilities
internal/         private shared Modules and process composition
acceptance/       top-level black-box, PTY, standalone, and native acceptance
docs/             maintainer and project-structure documentation
web/              Vite sources plus Go embedding/static-route package
mock/             independent mock applications and their tests
legacy/bun/       frozen read-only Bun behavior reference
tools/            one directory per build/check/harness tool
scripts/          user-facing install scripts; keep the plural name
build/            build metadata and generated local/cross-platform outputs
public/           versioned static source assets
```

Create a real top-level `acceptance/` package for cross-package, black-box,
PTY, standalone-binary, and native-platform evidence. Its tests use explicit
build tags or commands and never run as an implicit side effect of ordinary
`go test ./...`; the exact test migration and fixture ownership is decided by
the later test-topology ticket.

Keep `scripts/` rather than renaming it to GitHub CLI's `script/`: the current
installer paths and documentation are user-facing contracts. Keep each
`tools/<name>` directory independent, including the separate
`tools/lefthook/go.mod`; tools may use explicitly allowed support packages but
do not import command leaf packages. The Makefile remains the single task
entry point.

Add `docs/project-layout.md` and link it from the existing development/readme
documentation. It becomes the canonical explanation of the entry chain,
package ownership, dependency direction, test levels, frozen/reference areas,
and generated outputs.

Preserve current output boundaries: `web/dist`, `build`, `release`, 7-Zip
payloads, Lefthook binaries, `.tmp`, and `.cache` remain generated/ignored;
`public` remains versioned source. Do not add parallel `output`, `artifacts`,
or `vendor` roots, and do not place business code in `build`.

`legacy/bun` and `mock/nginx-proxy-manager` stay in place and independent.
The former remains a frozen behavior reference with no active imports; the
latter remains a mock application. Existing `check-no-bun`, architecture
checks, and Makefile isolation rules continue to cover these boundaries.

## Question

Decide the target roles and locations of non-command areas in the style of
GitHub CLI: `web`, `mock`, `legacy/bun`, `tools`, `scripts`, `build`, public
assets, generated outputs, and any future `acceptance` or `test` tree. Keep
the current web/build/ignore rules and legacy read-only policy intact while
making the repository root easy to scan. Clarify which directories are
source, test infrastructure, frozen reference, generated output, or release
support, and whether any directory only needs documentation rather than a
physical move.

The answer must avoid turning this effort into a web, deployment, or release
workflow redesign.
