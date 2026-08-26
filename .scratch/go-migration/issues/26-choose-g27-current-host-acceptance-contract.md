# Choose the current-host G27 acceptance contract

Type: wayfinder
Status: resolved
Blocked by: G27

## Question

The full G27 six-target native matrix cannot be executed in the current
environment, while the ongoing migration is primarily exercised on native
Windows x64 and macOS arm64 machines. Which bounded acceptance contract allows
this Goal to finish without claiming evidence for unavailable targets?

## Evidence

- The selected primary target set is native `windows/amd64` and
  `darwin/arm64`.
- The approved WD-24 and WD-25 Windows adaptations are implemented, and the
  complete native `internal/commands/tunnel` suite plus
  `go vet ./internal/commands/tunnel` pass on this host.
- Matching macOS x64, Linux x64, and Windows arm64 environments are not
  available in this execution context; the remaining target matrix cannot be
  completed here.
- The repository's exact `make bootstrap`, `make check`, and `make build`
  commands still require a `make` executable, which this host does not expose.

## Decision

On 2026-08-26 the user approved **primary-host-set acceptance** for this G27
Goal. The selected acceptance target set is `{windows/amd64, darwin/arm64}`.

1. G27 native execution Exit condition 3 is scoped to the selected
   `windows/amd64` and `darwin/arm64` artifacts for this acceptance variant.
2. The native portion of Exit condition 4 is scoped to the selected target
   set; the other four target rows remain explicitly `pending` and are not
   claimed as `release-accepted`.
3. The six-artifact static contracts, current-host repository/artifact checks,
   checksum and payload inspection, generated-output hygiene, and scope
   restrictions remain applicable. This decision does not make an unavailable
   command or missing evidence pass.
4. This is a local primary-host-set acceptance record, not a claim that the
   full six-target public release is ready. A later Goal may resume G27 with
   the remaining target evidence without changing product behavior.
5. No test is skipped, weakened, or relabeled as passed; unselected targets
   are deferred by this explicit acceptance boundary.

This decision does not authorize workflow, tag, release, CI, Docker,
deployment, protocol, data, API, or product behavior changes.
