# Choose the remaining Windows Tunnel/FRP native acceptance contract

Type: wayfinder
Status: resolved
Blocked by: G27

## Question

The native Windows amd64 G27 suite still fails five Tunnel tests after the
approved WD-24 SQLite URI adaptation. Which smallest Windows contract permits
the required evidence without changing the public Tunnel API, protocol-v3
messages, Go-only data boundary, release scope, or the no-skip/no-waiver gate?

## Evidence

- With `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0`, the focused URI tests
  and `go vet ./internal/commands/tunnel` pass on native `windows/amd64`.
- Repeated `TestRunClientDispatchesRestartFRPC` runs report
  `runtime start/restart = 1/30` through `1/572`, rather than `1/1`. The test
  fixture waits for any `process_state=running` frame before sending `revoke`,
  while the client also emits a running frame during the initial apply. This
  is a control-frame/test-lifecycle ordering mismatch; no protocol change has
  been authorized.
- `TestClientAppliedStateRoundTripsAtomicallyAndIgnoresInvalidCache` writes
  `last-applied.json` with `Chmod(0600)` but native Windows `FileInfo.Mode()`
  reports `0666`. POSIX mode bits are not Windows access evidence.
- `TestManagedFRPSWritesAndRemovesCustom404Page` and
  `TestManagedFRPSPublishesConfiguration` make the same invalid POSIX-mode
  assertion for Windows-generated FRP files.
- `TestManagedFRPSReadsCustom404Page` treats a Windows path-not-found result
  for a file whose parent is itself a regular file as ordinary absence;
  Unix reports the blocked-parent condition as `CONFIGURATION_FAILED`.
- WD-24 item 3 names File Session and Upgrade files only. It does not authorize
  a general ACL redesign or Tunnel client/FRP file policy.

## Decision boundary

Any accepted choice must keep Unix mode checks and behavior unchanged, retain
the existing public protocol/API and error codes, preserve the exact G27
native-suite vectors, and avoid skipped or weakened assertions. It may add
only owner-local Windows adapters and test-fixture synchronization needed to
express the same contract natively.

## Options

### A. Extend the bounded native file contract (proposed)

Amend WD-24 for the following Tunnel-owned Windows files only:

1. `last-applied.json`, active/candidate `frpc` configuration, `frps.toml`,
   and `404.html` receive a protected DACL granting access only to the current
   user, LocalSystem, and built-in Administrators; Windows tests verify that
   DACL, while Unix tests retain `0600`/`0700` mode evidence.
2. Missing custom-page targets remain ordinary absence, but a non-directory
   parent or other access/path failure remains `CONFIGURATION_FAILED` on every
   target through a narrow path-error adapter.
3. The restart vector gets a test-only acknowledgement/transition barrier so
   it proves exactly one `restart_frpc` dispatch; production control frames,
   reconnect policy, and FRP lifecycle remain unchanged.
4. No generic ACL framework, protocol revision, data migration, or release/CI
   change is included.

### B. Preserve WD-24 literally and waive the five Windows vectors

Leave Tunnel client/FRP files and the fixture unchanged, mark these tests as
non-blocking for Windows, and accept static or Unix evidence instead. This is
not compatible with G27: its contract forbids skipped/waived native items and
requires matching target evidence.

### C. Introduce a repository-wide Windows security/process abstraction

Move all state, generated files, process supervision, and path errors behind a
shared Windows policy layer. This could cover the failures, but it expands
past the Artifact Set scope, changes multiple prior Unit ownership boundaries,
and is explicitly excluded by WD-24.

## Answer

WD-25 A was approved by the user on 2026-08-26. It authorizes only the bounded
adaptations in option A:

1. On Windows, the listed Tunnel client and FRP files use a protected DACL
   granting access only to the current user, LocalSystem, and built-in
   Administrators; Unix mode evidence remains unchanged.
2. Custom-page absence is distinct from a blocked parent or access/path error,
   which remains `CONFIGURATION_FAILED` on every target.
3. The restart vector may add a test-only acknowledgement/transition barrier;
   production protocol frames, reconnect policy, and FRP lifecycle remain
   unchanged.
4. No generic ACL framework, protocol revision, data migration, release/CI,
   Docker, deployment, or test waiver is authorized.
