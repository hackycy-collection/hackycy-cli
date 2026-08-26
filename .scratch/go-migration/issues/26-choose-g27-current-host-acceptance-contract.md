# Choose the current-host G27 acceptance contract

Type: wayfinder
Status: resolved
Blocked by: G27

## Question

The full G27 six-target native matrix cannot be executed in one task, while
the current migration evidence is complete on native Windows x64. Which
bounded acceptance contract allows this Goal to finish without claiming
evidence for the other five targets?

## Evidence

- Native `windows/amd64` is the sole selected target for this local acceptance
  record; its candidate evidence covers the complete Windows G27 suite.
- The approved WD-24 and WD-25 Windows adaptations are implemented, and the
  complete native `internal/commands/tunnel` suite plus
  `go vet ./internal/commands/tunnel` pass on this host.
- macOS x64/arm64, Linux x64/arm64, and Windows arm64 native suites are
  deliberately deferred rather than combined into this task.
- A clean native Windows checkout already recorded exact `make bootstrap`,
  `make check`, `make build`, and six-artifact candidate evidence.

## Decision

On 2026-08-26 the user approved **Windows amd64-only local acceptance** for
this G27 Goal. This supersedes the earlier primary-host-set wording. The
selected acceptance target set is `{windows/amd64}`.

1. G27 native execution Exit condition 3 is scoped only to the selected
   `windows/amd64` artifact for this acceptance variant.
2. The native portion of Exit condition 4 is scoped to the selected target
   set; the other five target rows remain explicitly `pending` and are not
   claimed as `release-accepted`.
3. The six-artifact static contracts, current-host repository/artifact checks,
   checksum and payload inspection, generated-output hygiene, and scope
   restrictions remain applicable. This decision does not make an unavailable
   command or missing evidence pass.
4. This is a local Windows amd64 acceptance record, not a claim that the full
   six-target public release is ready. A later Goal may resume the deferred
   target evidence without changing product behavior.
5. No test is skipped, weakened, or relabeled as passed; unselected targets
   are deferred by this explicit acceptance boundary.

This decision does not authorize workflow, tag, release, CI, Docker,
deployment, protocol, data, API, or product behavior changes.
