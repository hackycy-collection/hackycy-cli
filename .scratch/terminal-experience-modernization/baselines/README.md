# G0 Behavior Baselines

This directory records the pre-presentation-change contract for the terminal
experience modernization. The baseline was captured from source commit
`b4a9f34d50242cb0fac251839f99d7d76d8bdfc6` on 2026-09-01, before any G0
implementation change.

`manifest.json` is the authoritative index. It contains one entry for the
root/help surface and every registered leaf command, grouped by command
family. Each entry records the observed stdin, stdout, stderr, exit/signal,
side-effect, state-file, process-boundary, and redaction contract, together
with the focused tests that provide executable evidence. Paths, credentials,
provider responses, and other run-specific values are represented by tokens
such as `<temp-root>` and `<secret>`; no captured artifact contains a secret.

The frozen command surface is retained separately under
`acceptance/testdata/command-surface/`. Its manifest and normalized help and
completion files are comparison artifacts, not generated output to be updated
as part of a presentation change.

## Capture method

- Command grammar and help: `make command-surface` and the frozen surface set.
- Stream/result behavior: package terminal/root-outcome tests and the
  representative standalone acceptance tests named in `manifest.json`.
- Signals, side effects, state files, permissions, and child ownership:
  command-specific process/signal/acceptance tests named in the corresponding
  entry. A finite command records signal behavior as not applicable when it
  has no foreground process lifecycle.
- Redaction: package and acceptance tests that assert encrypted previews,
  provider-key omission, safe errors, and logger redaction.

The probe summary is intentionally semantic rather than a raw transcript so
that temporary paths, timestamps, child output, and credentials cannot become
part of the baseline evidence.
