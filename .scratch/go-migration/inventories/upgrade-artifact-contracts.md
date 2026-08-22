# Upgrade and release-artifact compatibility inventory

> **Roadmap transition-scope amendment (2026-08-22):** this inventory remains the factual source for artifact, installer, and Go-era Upgrade behavior, but the first Bun-to-Go switch uses `scripts/install.sh`, `scripts/install.ps1`, or manual binary replacement. Bun `ycy upgrade`, Legacy Update State, and the mixed-runtime cutover tests below are historical evidence rather than first-release gates. Public Go Upgrade is required for Go-to-later-Go only.

This inventory records the release, installation, version, and self-update contracts that must survive the move from the Bun standalone executable to a CGO-free Go executable. It also records the build contents that make each published file standalone. It is a migration-specification input, not approval to reproduce the legacy updater's authorization, recovery, resource, or exit-status defects.

## First-release scope

The first Go release is parity-first: public upgrade behavior for Go-to-Go replacement, installer/manual handoff, native replacement flow, and artifact contracts remain the baseline even where this inventory identifies defects. Bun-written updater state is historical evidence only. Updater hardening and a new recovery model are post-parity work. Only a demonstrated Go executable, target-platform, Go-owned state, installer, or artifact incompatibility may create a narrow exception during `upgrade` migration.

## Contract classification

The migration uses three compatibility levels:

1. **Exact public and Go-era compatibility:** command and flag names; plain `--version` output; GitHub repository, latest-release lookup, tag convention, artifact names, `SHA256SUMS` naming and entry identity; platform mapping; install paths; checksum-before-execution; installer/manual first-Go replacement; and Go-to-later-Go update behavior.
2. **Intent compatibility:** human-facing progress layout and wording, exact temporary UUIDs, and incidental retry or polling timings may vary where no caller consumes them; the first release otherwise preserves the observed update and recovery outcomes.
3. **Post-parity defect:** unauthenticated hidden replacement, concurrent updater races, permanently blocking or malformed state, non-durable state transitions, unchecked paths and symlinks, PID-reuse ambiguity, unbounded downloads, weak release metadata validation, Windows ARM64 installer misselection, and operational failures that return exit status 0 remain later hardening findings rather than first-release redesign requirements.

The first Go release has no Bun-updater rolling constraint. Its filename, checksum publication, executable format, and plain version output must work with `scripts/install.sh`, `scripts/install.ps1`, or manual binary replacement. Public `ycy upgrade` acceptance starts only from a running Go artifact and installs a later Go artifact.

## Verification baseline and evidence

The local upgrade baseline was run on 2026-08-22 with Bun 1.3.14:

```text
bun test src/commands/upgrade

11 pass
1 skip
0 fail
34 expect() calls
1 file
```

The skipped test compiles and exercises a real detached updater only on Windows and only when `YCY_RUN_COMPILED_UPDATER_TEST=1`. The baseline does not exercise `upgrade.ts`, an installer, a published release, a macOS or Linux self-replacement, Windows ARM64, or any of the six final Go artifacts.

Historical evidence establishes the Windows bridge:

- tag `v0.0.46` still replaces the running executable synchronously;
- commit `879156f` introduced the copied detached updater and the README warning;
- `v0.0.47` is the first tag containing that updater;
- the documented support rule is therefore exact: Windows users on `v0.0.46` or earlier run the installer once before using `ycy upgrade`.

The current source evidence is `src/commands/upgrade/upgrade.ts`, `src/commands/upgrade/updater.ts`, `src/cli.ts`, `scripts/install.sh`, `scripts/install.ps1`, `scripts/build.ts`, `scripts/prepare-seven-zip.ts`, `src/commands/fs/archive-manifest.ts`, `src/commands/fs/archive-runtime.ts`, `package.json`, and `.github/workflows/release.yml`. The workflow is read only as evidence; editing CI or release automation remains out of scope.

## Terminology

Use these names consistently in later decisions and tests:

- **Release Artifact:** one downloadable `ycy-<os>-<arch>[.exe]` standalone executable.
- **Artifact Set:** the six Release Artifacts plus the `SHA256SUMS` Checksum Manifest for one tagged release.
- **Release Identity:** the repository tag with one leading `v`, and the plain semantic version embedded in every matching executable.
- **Install Target:** the stable filesystem path at which an installer or updater publishes `ycy`.
- **Staged Candidate:** a fully downloaded, hashed, executable, version-checked Release Artifact not yet published as the Install Target.
- **Updater Copy:** a temporary copy of the currently running executable that waits for its parent and owns replacement after the parent exits.
- **Update Transaction:** the one-at-a-time operation connecting an Install Target, Staged Candidate, backup, expected digest/version, Updater Copy, and status.
- **Legacy Update State:** the unversioned JSON document written beside the Bun Install Target by releases starting with `v0.0.47`.
- **Artifact Self-check:** executing a candidate or installed binary with `--version` and matching the expected Release Identity.
- **Runtime Payload:** target-specific files carried inside a Release Artifact and materialized only when a command needs them, currently the web graph and 7-Zip runtime; FRP is not a Runtime Payload.

These terms describe the release boundary. Exact Go package ownership remains for `Choose the Go module seams and project layout`.

## Public command and version surface

The exact public leaves are:

```text
ycy upgrade
ycy -V
ycy --version
```

`upgrade` has no options or positional arguments. It uses the current platform and architecture and always targets the latest non-prerelease/non-draft release returned by GitHub's `releases/latest` endpoint. There is no release channel, version selector, downgrade command, dry run, proxy flag, alternate repository, or machine-readable progress mode.

The current Commander version output is the plain package version followed by a newline on `stdout`, for example:

```text
0.0.69
```

It exits 0 and is machine consumed by both installers and by staged/installed artifact self-checks. The installers and Bun updater also accept output beginning with `ycy/<version>` for compatibility with older output variants, but the active CLI's public contract is the plain version. The Go CLI must retain plain output for both `-V` and `--version`; it must not prepend update status, a product name, build metadata, color, or other text during an Artifact Self-check.

The release build must inject the tag's version without its single leading `v`. Every artifact in an Artifact Set must report the same value. Local development may use a clearly non-release identity selected later, but no artifact may be published with an implicit zero value, stale package version, dirty marker, or mismatched tag.

## GitHub release and integrity contract

The fixed upstream is:

```text
repository: hackycy/hackycy-cli
latest API: https://api.github.com/repos/hackycy/hackycy-cli/releases/latest
download:   https://github.com/hackycy/hackycy-cli/releases/download/v<VERSION>/<ARTIFACT>
checksums:  https://github.com/hackycy/hackycy-cli/releases/download/v<VERSION>/SHA256SUMS
```

Current release resolution is:

1. Fetch the latest-release API with `Accept: application/vnd.github.v3+json`.
2. Remove one leading `v` from `tag_name`.
3. Compare the embedded current version and latest version with Bun semantic-version ordering.
4. Treat current equal to or greater than latest as already current; never downgrade.
5. Map the running platform to one exact artifact name.
6. Prefer that asset's GitHub `digest` after removing `sha256:`.
7. If the digest is absent, fetch `SHA256SUMS` and select the exact artifact basename.
8. Download the artifact, require a non-empty body, calculate SHA-256, and require equality before writing/executing it.

The release continues to publish a file named exactly `SHA256SUMS`. Each line is a 64-hex-character SHA-256 digest, two spaces, and an artifact basename. Existing consumers also accept a single whitespace separator and an optional GNU binary marker `*`; a Go parser must tolerate those valid forms while requiring one exact basename match. Digests are compared case-insensitively after validation and normalization.

The existing release always uses tags shaped `v<semantic-version>` and reconstructs download URLs in that form. Go must preserve that repository convention. Missing, non-string, invalid, or unexpectedly shaped tags; malformed asset digests; duplicate/conflicting checksum entries; redirects outside the allowed release policy; truncated bodies; and content over selected limits must fail closed. The exact HTTP deadlines, byte limits, redirect policy, and diagnostic surface belong to `Choose a safe and recoverable self-update contract`.

No Release Artifact is trusted merely because it came from HTTPS or GitHub metadata. The required chain is release identity, exact artifact selection, SHA-256, candidate self-check, atomic publication with backup, installed SHA-256, and installed self-check.

## Six-artifact matrix

Artifact names and the standalone distribution model are hard constraints:

| Runtime platform | Go target | Public artifact | Executable suffix | Target 7-Zip payload |
| --- | --- | --- | --- | --- |
| macOS x64 | `GOOS=darwin GOARCH=amd64` | `ycy-macos-x64` | none | macOS `7zz`, `License.txt` |
| macOS arm64 | `GOOS=darwin GOARCH=arm64` | `ycy-macos-arm64` | none | macOS `7zz`, `License.txt` |
| Linux x64 | `GOOS=linux GOARCH=amd64` | `ycy-linux-x64` | none | Linux x64 `7zz`, `License.txt` |
| Linux arm64 | `GOOS=linux GOARCH=arm64` | `ycy-linux-arm64` | none | Linux arm64 `7zz`, `License.txt` |
| Windows x64 | `GOOS=windows GOARCH=amd64` | `ycy-windows-x64.exe` | `.exe` | Windows x64 `7z.exe`, `7z.dll`, `License.txt` |
| Windows arm64 | `GOOS=windows GOARCH=arm64` | `ycy-windows-arm64.exe` | `.exe` | Windows arm64 `7z.exe`, `7z.dll`, `License.txt` |

The public vocabulary remains `x64`; Go's internal architecture vocabulary is `amd64`. Raw `runtime.GOARCH` must never leak into artifact names, GitHub URLs, the Checksum Manifest, installer selection, or legacy wire protocols.

Every Go build in the matrix uses `CGO_ENABLED=0`. Cross-compilation proves only that a file can be emitted; it does not approve runtime behavior. All six filenames, object formats, CPU architectures, embedded payloads, and native smoke tests are final gates.

The published unit is one executable per target. Users do not install Go, Bun, Node, pnpm, npm packages, a frontend directory, a source map, a sibling worker, or a sibling 7-Zip file. The executable may materialize its verified target Runtime Payload in ycy-owned state at runtime. FRP remains a separately verified runtime download governed by its pinned manifest and is not added to the ycy executable.

## Standalone build contents

The current Bun build compiles `src/cli.ts`, the FS thumbnail worker, and target-selected 7-Zip payload files into one output. It enables minification and creates an external source map, but the release workflow uploads only the six executable files and `SHA256SUMS`; source maps are not a public artifact contract.

The Go replacement must preserve the effective contents while changing ownership:

1. `make build` builds the active pnpm/Vite React MPA first and verifies its production graph.
2. Go embeds the complete Vite output once: fixed `fs`, `diff`, and `tunnel-server` HTML shells plus one shared hashed asset tree.
3. Go embeds the target-selected 7-Zip 26.02 payload and license.
4. Go embeds or generates the pinned FRP manifest data, but not FRP executable bytes.
5. Go compiles `cmd/ycy` with the selected Release Identity and `CGO_ENABLED=0` into the standalone executable.
6. The artifact assembly command emits all six fixed names and generates `SHA256SUMS` from the final bytes.

`web/dist` remains ignored and absent from source control. A clean checkout therefore needs the frontend-first build path selected by `Prove the Vite MPA to Go embed path`; a raw Go command that needs an embed directory cannot silently use stale assets.

The 7-Zip inputs remain pinned by target:

| Target family | Upstream 7-Zip 26.02 asset | Embedded files |
| --- | --- | --- |
| macOS x64/arm64 | `7z2602-mac.tar.xz` | `7zz`, `License.txt` |
| Linux x64 | `7z2602-linux-x64.tar.xz` | `7zz`, `License.txt` |
| Linux arm64 | `7z2602-linux-arm64.tar.xz` | `7zz`, `License.txt` |
| Windows x64 | `7z2602-x64.exe` | `7z.exe`, `7z.dll`, `License.txt` |
| Windows arm64 | `7z2602-arm64.exe` | `7z.exe`, `7z.dll`, `License.txt` |

The upstream archive digest, every extracted payload digest, executable bit/Windows colocation, license presence, and runtime publication must be verified by the final build and native artifact tests. At runtime, the current Bun implementation trusts an already materialized file without rehashing it; preserve that first-release behavior and defer runtime-state revalidation hardening.

## Installer compatibility

The stable install locations are exact user-visible contracts:

| Platform | Install Target | PATH behavior |
| --- | --- | --- |
| macOS/Linux | `$HOME/.ycy-cli/bin/ycy` | add `$HOME/.ycy-cli/bin` to the selected shell profile when absent |
| Windows | `%USERPROFILE%\.ycy-cli\bin\ycy.exe` | add the directory to the user PATH when absent |

Both current installers:

- query the same latest GitHub release;
- select the platform artifact and expected release digest/Checksum Manifest entry;
- download to `<target>.tmp.<pid>`;
- reject an empty or hash-mismatched candidate;
- remove quarantine/MOTW where supported;
- move an existing target to `<target>.backup.<pid>`;
- publish the candidate, then recheck its SHA-256 and `--version`;
- roll back the prior target when publication or self-check fails;
- remove the backup only after success; and
- fail nonzero on installation errors.

The Unix installer maps Darwin/Linux and `x86_64|amd64`/`arm64|aarch64` to all four Unix artifacts. It requires `curl` plus `shasum` or `sha256sum`. On macOS it attempts to remove `com.apple.quarantine`; failure is currently ignored.

The PowerShell installer currently hard-codes `ycy-windows-x64.exe`. It never inspects whether the host is Windows ARM64 even though `ycy-windows-arm64.exe` is published. This is a compatibility defect: the active installer must map native AMD64 and ARM64 hosts to their matching artifacts, reject unsupported architectures explicitly, and test both. Existing x64 installs and the stable Install Target remain unchanged.

Installer parsing and downloads are insufficiently bounded, and the Unix script parses JSON with text tools. Preserve their observable first-release behavior; later hardening may change internals without changing the public install command or paths. The cutover decision must say whether the active scripts are retained or replaced, but neither active installer may require Bun, Node, pnpm, or Go on the user's machine.

A completed or stale Bun Legacy Update State beside the Install Target is not detected, consumed, cleaned, fixture-tested, or treated as an installer gate. The installer replaces the target through its normal path; ordinary diagnostics and internal operator recovery cover unrelated residue.

## Public `upgrade` behavior

The current successful path is:

```text
resolve latest release
  -> compare versions
  -> select target artifact and digest
  -> download and hash bytes
  -> write same-directory Staged Candidate
  -> chmod/xattr and candidate --version
  -> copy current executable to temporary Updater Copy
  -> write pending Go-owned Update State
  -> spawn Updater Copy detached and unreferenced
  -> parent reports scheduled update and exits
```

Using a same-directory Staged Candidate is important: final publication is a same-filesystem rename rather than a cross-device copy. The Updater Copy lives under the OS temporary directory and includes `.exe` on Windows. A UUID separates staged, backup, updater, and state-temporary files.

Go-owned state uses `<target>.go-update-state.json` and same-directory temporary files with the same prefix. Go Upgrade and its internal apply mode read only that namespace. They never open Bun's historical `<target>.update-state.json`, so first installation cannot accidentally become a cross-runtime state migration.

The command currently buffers the complete artifact in memory and has no explicit connection, response, total, or self-check timeout and no response size cap. It also trusts weakly validated release JSON and asset digest strings. Go may stream into a staged candidate while hashing when that does not introduce new limits or change the artifact/digest and failure behavior; resource and metadata hardening is post-parity.

The current human output uses Clack spinners and colored progress. Wording, color, and spinner layout are not machine contracts. The following outcomes are behavioral contracts:

- already current is a successful no-op;
- unsupported OS/architecture is a failure;
- API/rate-limit, metadata, checksum, download, write, candidate self-check, scheduling, replacement, installed self-check, and rollback failures retain the frozen command's visible output and exact observed exit mapping;
- a successfully scheduled detached transaction may return before replacement finishes, but the next invocation can determine its state;
- no failure may leave an unverified new executable presented as the Install Target.

The Bun implementation returns exit status 0 from several HTTP, missing-checksum, empty-download, and checksum-mismatch abort branches; only thrown exceptions force exit 1. Preserve and test those zero exits for first-release parity; correcting them is later work.

## Legacy Update State contract (historical only)

> The shapes below document the Bun reference. The first Go implementation must not parse, detect, migrate, delete, or fixture them. Any update state used by Go is created and consumed only by Go-to-Go flows.

For an Install Target `<target>`, Bun uses:

```text
state:   <target>.update-state.json
staged:  <target>.new.<transaction UUID>
backup:  <target>.backup.<transaction UUID>
updater: <os-temp>/ycy-updater-<transaction UUID>[.exe]
temp:    <state>.<transaction UUID>.tmp
```

The unversioned JSON object contains:

```json
{
  "transactionId": "...",
  "parentPid": 123,
  "targetPath": "...",
  "stagedPath": "...",
  "backupPath": "...",
  "expectedHash": "...",
  "expectedVersion": "...",
  "statePath": "...",
  "updaterPath": "...",
  "createdAt": "...",
  "status": "pending|succeeded|succeeded_with_cleanup_warning|failed",
  "message": "optional text"
}
```

An earlier rolling-cutover plan would have required Go to recognize and consume this shape. That plan is void: Go must not act on any of these Bun-written fields or files. The Go-to-Go implementation may consult the legacy flow for observable public scheduling/replacement behavior, but it creates its own state only after a Go process initiates Upgrade. A versioned schema, new locking, durable publication, and abandoned-state recovery remain post-parity redesign topics.

## Detached replacement and rollback

The legacy Updater Copy currently:

1. Parses name/value pairs after the hidden `--internal-apply-update` marker.
2. If a readable state has the same transaction ID, replaces parsed values with that stored object.
3. Polls the parent PID every 50 ms for at most 30 seconds.
4. Requires a Staged Candidate and refuses an existing transaction backup.
5. Renames the old target to backup when present.
6. Renames staged to target.
7. On Unix, applies mode `0755`; on macOS, removes quarantine.
8. Recalculates installed SHA-256 and executes installed `--version` with `YCY_INTERNAL_UPDATE_VERIFY=1`.
9. On failure, removes the new target and renames backup to target.
10. On success, removes backup; a cleanup failure records a warning rather than undoing a verified installation.
11. Atomically renames a temporary JSON state to `succeeded`, `succeeded_with_cleanup_warning`, or `failed`.

Rename/unlink operations retry `EACCES`, `EBUSY`, and `EPERM` up to 100 times with 50 ms sleeps. This primarily accommodates Windows executable/file-lock release. Reproduce these observable retries where the target OS permits; any necessary platform deviation requires a focused native probe and narrow compatibility exception.

Unix permits unlinking the running Updater Copy, so it attempts self-cleanup. Windows keeps the executing copy until the next normal CLI invocation consumes completed state. macOS accepts `xattr -d com.apple.quarantine` exit code 0 or 1 because absence of the attribute returns 1; other failures abort and roll back. Windows installer `Unblock-File` and macOS quarantine handling remain native artifact gates.

Rollback is required whenever the old target was moved and the candidate fails publication, hashing, execution, or version matching. If rollback itself fails, both causes must remain diagnosable and recovery material must not be silently deleted. Removing a verified old backup may produce a cleanup warning but must not turn the verified new target back into a failure.

## Startup consumption behavior

> This table is Bun reference evidence only. Go does not inspect `<process.execPath>.update-state.json` as Bun Legacy Update State during the first installation or ordinary startup.

Every normal Bun CLI invocation examines `<process.execPath>.update-state.json` before Commander parsing unless `YCY_INTERNAL_UPDATE_VERIFY=1`:

| Legacy state | Current Bun behavior | First-Go disposition |
| --- | --- | --- |
| none | continue normally | normal Go startup |
| `pending` | print retry message to `stderr`, exit 1, retain state | ignore as unsupported Bun state |
| `succeeded` | print once, remove state/updater, continue command | ignore as unsupported Bun state |
| `succeeded_with_cleanup_warning` | print once, remove what it can, continue | ignore as unsupported Bun state |
| `failed` | print rollback result once, remove state/updater, continue | ignore as unsupported Bun state |
| malformed/unrecognized | ignore, leave file indefinitely, continue | no detection or cleanup |

The Bun result line is emitted before the requested command, including `--version`, and can break machine consumers and installers. It is not carried across the first installation. Equivalent output from a Go-created update result follows the public Go-to-Go parity tests; redesigning result reporting remains post-parity.

Installer temporary files shaped `<target>.tmp.<pid>` are cleaned under the active installer/Go behavior when that PID no longer exists. Bun legacy-state temporary cleanup is historical only and is not implemented by Go. Recovery for Go-owned backup, staged, and state files remains limited to the observed Go-to-Go replacement contract.

## Confirmed defects retained for post-parity hardening

The following findings are not first-release blockers. The port preserves their observable behavior; the final column records later hardening candidates:

| Finding | Risk | Post-parity candidate |
| --- | --- | --- |
| Hidden updater accepts caller-supplied target/staged/backup in one directory without requiring matching pending state | direct invocation can turn ycy into an arbitrary same-directory executable replacement primitive | authenticate the transaction from trusted state and bind every path/identity to the current ycy target |
| Hidden marker is found anywhere in argv | ordinary parsing is bypassed by a user-reachable token | preserve the marker scan outside Cobra and test every observed position; authentication hardening is post-parity |
| No transaction lock | two `upgrade` calls can both pass the pending check, overwrite state, and race target/backup renames | one cross-process lock per Install Target with owner/recovery semantics |
| State is unversioned, loosely validated, and not fsynced | crash, truncation, forged fields, or future schema drift can strand or misdirect replacement | versioned validation and durable Go-owned file/directory publication |
| Any valid `pending` state blocks forever | a crashed/missing updater permanently disables the CLI, including `--version` | live-owner detection, bounded takeover/recovery, installer recovery, and explicit diagnostics |
| Malformed main state is ignored forever | corruption is invisible and cannot self-heal | preserve evidence through backup/quarantine and continue or fail according to selected safe state |
| PID liveness proves only PID reuse, not parent identity | updater can wait on an unrelated process or misclassify abandonment | bind transaction identity beyond a bare PID where each OS permits and retain bounded fallback |
| Paths, symlinks, permissions, and target identity are under-validated | replacement may escape the intended executable or follow a changed link | canonical/opened-handle or equivalent OS-specific target policy with no arbitrary path capability |
| Download and release JSON are fully buffered without deadlines/limits | network or metadata can exhaust memory or hang indefinitely | streaming hash-to-stage, bounded bodies/time/redirects, strict schema/digest/tag validation |
| Several aborts exit 0 | scripts cannot distinguish update failure from success | exact nonzero operational failures; already-current remains 0 |
| PowerShell always selects x64 | Windows ARM64 users do not receive the published native artifact | native architecture detection and both Windows artifact tests |
| Completed state can pollute `--version` | installer and updater self-checks can reject a valid artifact | clean self-check channel plus deterministic result reporting |
| Crash points are not transactionally recovered | target, backup, staged, and state may disagree after power loss | enumerate every publication point and define idempotent roll-forward/rollback |
| Native replacement proof is almost absent | cross-platform filesystem and process semantics differ materially | native macOS/Linux/Windows x64/arm64 artifact tests, with real Windows detached replacement |

The hidden entry remains an observable Go-to-Go parity target even though Go may require different internal process or platform APIs. It does not accept Bun Legacy Update State. If a required native target cannot reproduce Go replacement, the failed focused probe creates the narrow compatibility exception; broad updater redesign remains separate.

## Supported installation and update directions

The accepted transitions are deliberately narrow:

1. **Bun installation to first Go:** run `scripts/install.sh`, `scripts/install.ps1`, or manually replace the binary. No Bun `ycy upgrade` bridge or legacy-state consumption is supported.
2. **First Go to later Go:** the Go updater downloads, verifies, stages, replaces, rolls back, and reports without Bun, Node, or frontend tooling.
3. **Installer to any Go release:** both active installers select the native artifact, verify digest and plain version, and publish at the stable target with rollback.

Go tests consult `legacy/bun/` for the public behavior baseline but start Upgrade integration from a Go artifact and use only Go-created state. They never run Bun, retain tagged Bun executables, or construct Legacy Update State fixtures.

## Existing coverage and required tests

### Legacy evidence to consult

The current 11 passing tests cover:

- successful target replacement and backup removal;
- checksum failure followed by rollback;
- verified installation with backup-cleanup warning;
- stale installer temporary cleanup while preserving a live PID;
- orphaned, active, pending, and malformed state-temporary cleanup;
- cleanup after atomic state replacement fails;
- exact hidden argument construction;
- completed-state one-time consumption and updater cleanup;
- pending-state retention and startup gate;
- parent-process polling; and
- failure-state recording after rollback.

The optional compiled Windows x64 test covers one successful detached replacement and the next invocation's result/state/updater cleanup. It does not run by default and has no ARM64 equivalent.

There are no tests for the public GitHub/API/download/version-comparison flow, installers, Unix/macOS real self-replacement, release manifest completeness, Go/Bun interoperability, web or 7-Zip bytes inside the final executable, concurrency, crash recovery, path attacks, or hostile network metadata.

### Required Go unit and integration tests

Before `upgrade` is command-complete, active tests must cover:

1. Exact `upgrade`, `-V`, and `--version` parsing, stdout/stderr, already-current status 0, and the frozen zero/nonzero status for every operational failure.
2. Semantic versions including equal, newer current, normal upgrade, prerelease/build metadata, invalid/missing tag, one leading `v`, and no accidental downgrade.
3. Exact six-way runtime-to-artifact mapping, especially `amd64` to public `x64`, plus unsupported OS/architecture rejection.
4. GitHub API status/rate-limit behavior, strict release/asset schema, digest normalization/validation, fallback Checksum Manifest parsing, missing/duplicate/conflicting entries, and fixed repository/tag URLs.
5. Downloads under the frozen buffering/timeout policy; zero, short/truncated, redirected, cancelled, disk-full/permission, and digest-mismatch cases; and no candidate execution before digest success. New byte/time/redirect limits remain post-parity.
6. Candidate and installed self-checks for the exact plain version, empty/multiline/wrong version, nonzero exit, hang, signal, and separation from ordinary CLI output.
7. Go-owned `<target>.go-update-state.json` generated by the Go command for every status and optional field it uses, malformed/interrupted Go-owned state behavior, and one-time Go-to-Go result consumption. Prove that `<target>.update-state.json` is never opened; there is no Bun Legacy Update State fixture.
8. Exact hidden-entry isolation and validation needed by the observed Go-to-Go replacement flow, including malformed/extra arguments and mismatched Go-created transaction paths. A new authenticated transaction protocol is post-parity.
9. The observed behavior for two concurrent Go upgrades, installer/update collision, and a normal command racing final status; do not silently add a new locking/recovery policy.
10. Failure injection at stage write, chmod/quarantine, state publication, target-to-backup rename, staged-to-target rename, installed hash/self-check, backup removal, result write, and updater cleanup, asserting the frozen rollback/visible-failure outcomes rather than a new crash-recovery state machine.
11. Normal rollback success/failure, target absent, backup conflict, read-only directory, low disk, native busy-file behavior, and cleanup-warning semantics. New path-identity and attack hardening remains deferred.
12. Go-to-Go result reporting that does not contaminate candidate/installed plain-version self-check output.

Network tests use injected/local transports and fixture servers; they do not contact GitHub. State fixtures are constructed in the Go tests by consulting `legacy/bun/`; active tests do not invoke Bun.

### Required installer tests

The active Unix and PowerShell installers need local fixture-server or command-mock tests for:

1. all six platform/architecture mappings, including Windows ARM64 and unsupported targets;
2. asset digest preference and `SHA256SUMS` fallback;
3. zero/truncated/wrong-hash/wrong-version/non-executable candidates;
4. fresh install, replacement, target/self-check failure rollback, backup cleanup warning, and preserved user PATH behavior; do not inspect Bun Legacy Update State;
5. spaces and Unicode in user-profile paths, restrictive directory/file permissions or ACLs, and no dependency on Bun/Node/Go;
6. macOS quarantine and Windows MOTW handling; and
7. replacement of an existing Bun or Go binary solely as bytes at the stable Install Target, without executing it or consuming its updater state.

Tests must redirect HOME/UserProfile, install directories, release endpoints, and profile/PATH mutation into temporary controlled locations. They must never overwrite the developer's real installed ycy.

### Required Artifact Set tests

A local release-candidate command, independent of CI configuration, must:

1. start from a clean checkout state with no `web/dist`, install frozen pnpm dependencies, build Vite once, verify the three fixed shells/shared hashed graph, and then build Go;
2. emit exactly the six public artifact basenames with `CGO_ENABLED=0`, the expected Mach-O/ELF/PE format and CPU, nonzero size, embedded Release Identity, and no Bun/Node runtime dependency;
3. generate `SHA256SUMS`, recompute every entry, reject missing/extra/duplicate files, and prove the installer/upgrader parser selects each exact basename;
4. inspect Go build metadata and the build recipe so a stale version, wrong GOOS/GOARCH, cgo-enabled object, or dirty/unreproducible input is visible;
5. prove each executable carries all three Vite shells, their reachable hashed assets/workers/Monaco resources, and no stale or cross-entry shell;
6. prove each executable carries only its target 7-Zip runtime plus `License.txt`, with pinned input and extracted-file digests; Windows must carry the matching DLL beside `7z.exe` when materialized;
7. prove FRP executable bytes are absent while the exact six-target pinned FRP manifest is present; and
8. leave generated `web/dist`, artifact directories, downloaded preparation caches, and source maps out of version control.

Cross-artifact inspection can run on one host with Go's `debug/macho`, `debug/elf`, and `debug/pe`. Execution claims require native target tests.

### Required native artifact tests

On matching macOS, Linux, and Windows x64/arm64 hosts, every candidate must pass:

1. `--version`, `--help`, parser-error, signal, and exit-status smoke tests from the standalone file with no Bun/Node/pnpm available.
2. Installer fresh-install/upgrade/rollback at the stable platform path shape but under a temporary user root.
3. Go-to-Go detached self-update success, checksum/version rejection, parent timeout, busy-file retry, rollback, crash recovery, and concurrent-update exclusion.
4. First-Go installation through each active installer or manual replacement, with no Bun Upgrade invocation or Legacy Update State.
5. macOS quarantine removal and executable mode, Unix atomic rename/permission behavior, and Windows MOTW, locked-running-executable, antivirus-style sharing violations, ACLs, and updater-copy cleanup under the observed replacement contract.
6. Functional serving of all embedded Vite entries and a real 7-Zip extraction that materializes the expected target runtime/license and rejects corrupted runtime state.
7. Pinned FRP acquisition and version verification as owned by the Tunnel gate, demonstrating that FRP is obtained at runtime rather than silently omitted or embedded.

No CI, Docker, deployment, or release workflow edits are authorized by this inventory. These are local specification and release-readiness gates to be placed in the final roadmap.

## Migration boundaries and ordering implications

1. Release identity and the exact artifact matrix are foundations, not the last-minute concern of `upgrade`. Every vertical command proof eventually runs from a real standalone Go artifact.
2. The CLI prototype must prove hidden/internal entry isolation, exact version output, parser exits, and a command action that can return errors without terminating the process from deep code.
3. Implement only Go-owned updater state for Go-to-Go replacement. Do not recognize, continue, clean, or fixture Bun Legacy Update State; durable redesign, new locking, and recovery mechanics are not first-release prerequisites.
4. The Vite/embed prototype and FS 7-Zip contract must settle deterministic target payload assembly before any Artifact Set can pass.
5. Implement public release resolution, artifact selection, legacy-compatible staging, and candidate verification independently from platform replacement. Platform replacement remains behind OS-specific files.
6. Prove the observed Unix and Windows replacement/recovery behavior before enabling Go-to-Go self-update. Any target-specific mismatch becomes a narrow compatibility exception.
7. Keep all six filenames and the `SHA256SUMS` format so both installers and Go-to-Go Upgrade select the exact artifacts.
8. Migrate `upgrade` after the lower-risk commands and shared runtime foundations, then finish with the complete local Artifact Set and native gates. A cross-compiled file alone is not completion.
9. Legacy remains reference-only after archival. Active Go code and tests do not import it, execute Bun, or depend on legacy build scripts.

## Compatibility checks surfaced

The first-release implementation must prove that both installers can acquire or replace with the first Go artifact, that a running Go artifact can upgrade to a later Go artifact, and that the observed Unix/macOS/Windows replacement behavior works for all six targets. It must not invoke Bun Upgrade or consume Bun-written updater state. A failed in-scope focused probe may open one narrow compatibility-exception ticket. Authentication redesign, cross-process locking, a versioned state machine, abandoned-state recovery, new download/metadata bounds, corrected exits, and crash-hardening remain post-parity work.
