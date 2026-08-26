# Decide Windows native acceptance contracts

Type: wayfinder
Status: resolved
Blocked by: G27

## Question

The G27 native Windows x64 clean-checkout suite exposed target-specific failures
across standalone test artifacts, fresh-Go Tunnel storage, File Session
publication, Go-to-Go Upgrade replacement, and Hookctl. Which smallest Windows
contract permits the required native evidence without changing the first-release
CLI, data, release, CI, Docker, or deployment scope?

## Evidence

- Standalone test builders used extensionless executable paths, while Windows
  artifact and process invocation semantics require `.exe` names.
- A drive-letter path serialized as `file://C:\\...` is not a SQLite file URI;
  the pinned Tunnel driver accepts the Windows drive path only in its
  driver-compatible no-host form, `file:C:/...` (with URI-escaped path
  characters).
- POSIX mode bits do not express a Windows DACL. Existing File Session and
  Upgrade assertions therefore cannot prove the required owner-only access on
  Windows.
- A process holding a file without delete sharing prevents Windows replacement.
  The established Upgrade retry budget is 100 attempts at 50 milliseconds, but
  File Session publication did not use it.
- Hookctl resolves the production Lefthook binary as `lefthook.exe` on Windows,
  while its disposable-worktree fixture created an extensionless placeholder.

## Answer

WD-24 was approved on 2026-08-25. It authorizes only the following Windows
native-target adaptations for G27 acceptance:

1. Standalone test artifacts and updater copies use the target executable name
   with `.exe`; product artifact basenames remain the fixed six-name matrix.
   Tests compare filesystem paths with `filepath` semantics and do not encode
   Unix separators.
2. Fresh-Go Tunnel SQLite opens use the pinned-driver-compatible `file:C:/...`
   URI on Windows, with path characters URI-escaped and no URI host. Non-Windows
   URI behavior remains unchanged.
3. On Windows, POSIX `0o600` and `0o755` bits are not security evidence.
   File Session records and Go Upgrade state, staged, backup, updater, and
   installed files must be validated through native DACLs that grant access only
   to the current user, LocalSystem, and the built-in Administrators group.
   Unix mode assertions remain unchanged.
4. File Session publication and Go-to-Go Upgrade replacement retry only sharing
   violation or permission-denied operations within the existing bounded
   five-second budget. On exhaustion, File Session retains the last valid
   record and Upgrade retains or rolls back to the previous executable, records
   failure, and reports a nonzero result. No test may skip a lock case or claim
   an atomic replacement that did not occur.
5. Hookctl resolves the pinned Windows Lefthook as `lefthook.exe`; Git hook
   files remain Git-managed `pre-commit` files. Disposable worktree tests use
   Windows-native paths and do not treat Unix executable mode or a shell
   shebang as Windows evidence.

This decision does not authorize a general ACL/security redesign, changes to
observable command behavior, a compatibility adapter for Bun-written
non-config state, or release automation.

## Amendment

On 2026-08-26 the user approved option A: Windows uses the driver-compatible
`file:C:/...` form. This supersedes the original `file:///C:/...` wording in
item 2 after the pinned driver probe showed that the latter is interpreted as
`/C:/...` and cannot open a native Windows drive path. The amendment is limited
to the SQLite URI spelling; all other WD-24 boundaries remain unchanged.
