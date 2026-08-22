# Git command compatibility inventory

Inventory date: 2026-08-22
Legacy baseline: `78358c0201b71891e36603d6abb8d7c87d54ad57`
Scope: `git heat`, `git pulse`, `git fork`, and `git cm`.

## First-release scope

This inventory records observed facts for a parity-first port. The first Go release reproduces the actual Bun behavior, including behavior labeled unsafe or defective below. Safety and product-policy recommendations are post-parity backlog and do not block command migration. Only a focused implementation or test proving a Bun-to-Go incompatibility may open a narrow compatibility exception for the affected Git leaf.

## Contract classification

This inventory uses the same boundary as the core-command inventory:

- **Compatibility contract**: command and option names, successful effects, repository scope, Git/provider requests, persistent formats, selection and confirmation meaning, exit status, and security boundaries that users rely on.
- **Presentation freedom**: exact ANSI colors, spinner frames, table widths, title clearing, tree glyphs, relative-time wording, and incidental whitespace may change. The semantic fields, choices, and mutation confirmations may not.
- **Legacy defect**: unsafe, internally contradictory, configuration-dependent, or nonfunctional behavior. Preserve its observable first-release behavior deliberately and keep the correction separate from the port.

These commands operate on arbitrary user repositories. Tests must use disposable repositories and local HTTP servers; active Go tests must not execute archived Bun code, contact real providers, push to real remotes, or mutate the developer's repository.

## Existing test and diagnostic baseline

The only Git-command tests are under `src/commands/git/cm`. The inspected baseline passed:

```text
27 pass, 0 fail
src/commands/git/cm/{index,run,changes,engine,evidence,model}.test.ts
```

Four related shared CM configuration/client tests also passed. They cover default/overridden timeouts, provider output limits, and timeout reporting, but not the complete command flow.

There are no tests for `git heat`, `git pulse`, `git fork`, either provider adapter, URL parsing, TAR parsing, destination publication, command-level CM staging/commit/push, Git hooks, cancellation, or Windows behavior.

Disposable-repository diagnostics confirmed four important gaps:

1. `git heat` reports paths containing tabs, newlines, or non-ASCII characters as raw Git quoted/octal strings rather than their real names.
2. `git pulse --days -1`, `--days 0`, and `--days 3oops` are all accepted and can exit 0.
3. CM evidence followed an untracked symlink and included the contents of a file outside the repository. It also included literal token values from `_authToken=...` in `.npmrc` and `apiKey` in `credentials.json`.
4. The fork extractor's `path.join(dest, "../../outside")` escapes the destination. Its overwrite confirmation inherits Clack's default `true`.

## Shared Git execution contract

- The active project may remove Bun, but these four commands still require the user's external `git` executable. This is analogous to `ycy run` invoking a selected package manager: it is product behavior, not a build dependency.
- Git command behavior is affected by the installed Git version, repository configuration, attributes, hooks, credential helpers, environment variables, safe-directory policy, filesystem encoding, and current repository state. The Go port must execute Git with explicit arguments and parse stable plumbing formats rather than replace mutating behavior with a Go Git library by assumption.
- In particular, `git commit` must continue to run the target repository's normal commit hooks and signing/configuration behavior; `git push` must continue to run its normal push hooks and credential flow. A library that writes objects or refs directly would not be compatible.
- Arguments are currently passed as argv arrays, not through a shell. This prevents ordinary shell expansion, but Git pathspec and option parsing still apply. A `--` separator does not make a Git pathspec literal; filenames containing pathspec magic remain a confirmed test gap.
- Most subprocess stdout/stderr is captured and converted into a single error message. Exact Git wording is not stable, but failure must remain nonzero and secrets must not appear in diagnostics.
- None of these commands has a machine-readable output mode. Human layout may be reimplemented with the selected Go prompt/table stack, but noninteractive behavior, stdout/stderr ownership, cancellation, and exit codes require explicit tests.
- The current modules call `process.exit` internally. Go command modules must instead return classified results/errors to the composition root so deferred cleanup, child cancellation, and tests work.

## `git heat`

### CLI and range selection

- Leaf: `git heat`, with no operand.
- Options are `-n, --limit <number>`, `-d, --days <number>`, `-t, --type <type>`, `-s, --sort <sort>`, `-r, --relative-time`, and `-q, --query <text>`.
- `--type` accepts exact `files`, `directories`, or alias `dirs`; CLI default is `directories`.
- `--sort` accepts exact `path` or `count`; CLI default is `path`.
- `--limit` and `--days` are mutually exclusive. With neither, limit defaults to 20 commits. Both values must be positive integers after the shared permissive `parseInt` parser, so `3oops` currently becomes 3 while zero and negatives fail.
- Query text is trimmed. An empty query is ignored; a nonempty query highlights every case-insensitive, non-overlapping substring match but does not filter rows.

### Git subprocess and parsing

The command runs these operations sequentially:

```text
git rev-parse --show-toplevel
git -C <root> log -n <limit> --name-status --pretty=format:__HACKYCY_HEAT_COMMIT__%H%x1f%ct%x1f%ci
```

For a day range, `-n <limit>` is replaced with `--since=<days> days ago`.

- Repository discovery uses the process cwd and supports normal linked worktrees because Git resolves them. Bare repositories and unborn repositories fail at the log/root stage.
- Every log record increments `commitCount`, even if it has no supported file-status row.
- Supported first status letters are `M`, `A`, `D`, `R`, and `C`. Other states are ignored. Renames/copies count only the destination path.
- Copy detection is not explicitly enabled, so `C` output depends on Git configuration/heuristics. Merge diff behavior likewise follows Git defaults.
- For files, every status occurrence adds one to total and its kind count. The most recent epoch observed for a path supplies `Changed at`.
- Directory rows aggregate only each file's immediate parent, not every ancestor. Root files aggregate under `.`.
- Path sorting places nested files before root files and `.` after named directories; otherwise it uses JavaScript `localeCompare`. Count sorting is descending total then locale path.
- `%ci` is sliced to its first 19 characters, discarding the numeric timezone. Relative display uses epoch seconds and the current clock.

The parser is line/tab based and does not request `-z` or disable Git path quoting. Leading/trailing whitespace is trimmed and C-style/octal quoted names are never decoded. This is a confirmed compatibility defect for tabs, newlines, backslashes, quotes, and non-ASCII paths; Go must use a NUL-delimited format and slash-aware Git-path handling.

### Terminal and exit behavior

- With results, the command clears stdout, prints the ycy title, then renders a static Ink report that auto-unmounts after about 100 ms.
- Semantic output is repository name, selected range, optional commit count, file/directory count, rank, changed time, presence of M/A/D/R/C, and path. Exact colors, checkmarks, wrapping, earliest/latest marks, legend, and column widths are presentation.
- `--relative-time` changes only the changed-time presentation. Query changes only highlighting.
- No matching files prints an informational message and exits 0 without the report.
- Mutually exclusive ranges, invalid positive ranges, repository discovery failure, and `git log` failure exit 1. There is no interactive cancellation path.
- No child timeout or explicit signal/cancellation coordination exists.

### Required Go tests

1. Option/default/alias matrix, legacy permissive integer parsing, accepted flag combinations, empty query, and query matching.
2. Limit and fixed-clock day arguments; normal, unborn, bare, nested-directory, and linked-worktree invocation.
3. All supported statuses, unsupported statuses, rename/copy source/destination policy, merges, root files, ancestor behavior, ties, and range commit counts.
4. NUL-safe names containing spaces, tabs, newlines, quotes, backslashes, leading dashes, pathspec magic, and Unicode on Unix and Windows.
5. Stable semantic report fields at wide/narrow widths, relative time around timezone/DST boundaries, `NO_COLOR`, redirected stdout, and empty results.
6. Missing Git, Git failure, cancellation/signal propagation, and no leaked child process.

Risk: **medium**. The aggregation logic is small, but the current Git output format is not path-safe and sorting/time behavior is platform-sensitive.

## `git pulse [directory]`

### Discovery and inputs

- Optional directory defaults to cwd and is resolved absolutely. The only option is `--days <number>`.
- A missing path, non-directory, or missing/unusable `git --version` exits 1 before scanning.
- Discovery is a sequential depth-first stack walk with no depth limit. Read errors are silently skipped, and the walker yields every 100 visited directories to keep the UI responsive.
- A repository is recognized only when a direct child named `.git` is a directory. Linked worktrees and submodules with a `.git` file are ignored; bare repositories are ignored. A directory containing an invalid/empty `.git` directory is initially counted as a repository.
- Discovery continues below a found repository, so nested repositories are also returned.
- Symlinked directories are not traversed. Exact excluded directory basenames are `node_modules`, `vendor`, `dist`, `.cache`, `Library`, `.Trash`, `bower_components`, `__pycache__`, `.venv`, and `venv`.
- Discovery order follows unspecified filesystem enumeration plus LIFO traversal. Final presentation later sorts repository paths.

### Date and Git log contract

- Without `--days`, scanning happens first and the user then selects Today=1, Yesterday=2, Last 3 days=3, Last 7 days=7, or Last 30 days=30.
- The boundary is local start-of-day minus `days - 1`, formatted `YYYY-MM-DD HH:mm:ss`. Value 2 therefore means since the beginning of yesterday and includes today.
- With `--days`, the shared permissive integer parser is used and there is no positive-value validation. Zero, negatives, and numeric prefixes are accepted; this is a defect.
- Up to five repository logs run concurrently:

```text
git -C <repo> log --since=<local boundary> --date=format:%Y-%m-%d %H:%M:%S --pretty=format:%an%x1f%ad%x1f%s
```

- Only history reachable from the repository's current `HEAD` is read; `--all` is not used. Records contain author name, formatted author date, and subject. Emails, hashes, branches, bodies, and changed files are not collected.
- Unit-separator parsing preserves additional separators in the subject but assumes author/date do not contain them.
- Per-repository spawn, log, and parse failures are silently dropped. Stderr is piped but not consumed, which can block an unusually noisy child. There is no timeout.
- Concurrent completion makes collection order nondeterministic. Presentation groups by lexically sorted absolute repository path and sorts commits inside each repository by descending formatted date.

### TTY, filtering, and exit behavior

- The command clears the screen, prints the title, workspace, scan progress, and fetch progress.
- No repositories or no successful commits is a successful no-result outcome (exit 0), including when every discovered repository failed.
- With zero/one distinct author, all commits are rendered without an author prompt. With two or three, all authors start selected. With more than three, none starts selected; at least one is required. Authors are sorted with `localeCompare`.
- Day-selection or author-selection cancellation reports cancellation and exits 0.
- Final semantic output is total commits/repositories and, per repository, its path plus each commit's date, author, and subject. Exact tree glyphs, icons, colors, wrapping, and spacing are presentation.
- There is no explicit non-TTY branch. `--days` avoids only the date prompt; multiple authors can still require input.
- Scanning, Git children, and prompts have no shared cancellation context. Per-repository errors are neither surfaced nor reflected in exit status.
- Author names and subjects are rendered without an explicit terminal-control-character policy.

### Required Go tests

1. Missing/file roots, nested repositories, excluded basenames, permission failures, symlink loops, invalid `.git`, linked worktrees, submodules, and bare repositories.
2. Legacy permissive days parsing; fixed-clock inclusive boundaries for all presets; timezone, locale, and DST transitions.
3. Exact Git argv, current-HEAD scope, empty/unborn/detached repositories, separator characters, Unicode, control characters, and failed/hung repositories.
4. Five-child concurrency bound, deterministic final ordering, progress accounting, cancellation during scan/fetch, child termination, and bounded memory.
5. Zero/one/two/three/many-author selection defaults, required selection, cancellation, redirected stdin/stdout, `NO_COLOR`, and narrow terminals.
6. The selected partial-failure policy: distinguish a true no-commit result from repositories that could not be inspected, with tested exit/report semantics.

Risk: **medium**. Filesystem discovery is simple, but repository identity, silent partial failure, date boundaries, and cancellation are currently underspecified.

## `git fork <repo> [dest]`

The command name is historical: it downloads a working tree without Git history; it does not create a provider-side fork.

### Repository input and configuration

- Required `repo` and optional `dest`; there are no options.
- Supported input intentions are full HTTP(S) URL, `host/owner/repo`, configured `alias:owner/repo`, or `owner/repo` defaulting to `https://github.com`. A trailing `.git` is removed. A `#ref` suffix selects a branch/tag/ref.
- SSH/scp syntax such as `git@github.com:owner/repo` is not supported and is interpreted as an alias.
- Owners may contain slashes, which supports GitLab subgroups. The current parser accepts an empty owner or repository (`/repo`, `owner/`), repeated/extra path components, arbitrary URL schemes, and an empty fragment. These are defects, not required inputs.
- Alias lookup decrypts the configured token. Otherwise an exact host match uses the first matching configured instance, its provider type/token, and its configured scheme, even when the input URL specified another scheme.
- Without config, hostnames containing `github` or `gitlab` are auto-detected; an unknown custom host fails with configuration guidance.
- This depends on the shared config schema, PBKDF2/AES-GCM compatibility, locking, and machine identity recorded in the core inventory. Plaintext tokens must never be persisted or printed.

### Destination mutation

- Destination defaults to the parsed repository name and is resolved against cwd.
- The command calls `readdir`. A nonempty destination asks `Directory "..." is not empty. Overwrite?`; because no initial value is supplied, the installed Clack version defaults confirmation to **yes**.
- Cancellation/no exits 0 without deletion. Confirmation recursively and forcefully deletes the complete destination **before** default-branch lookup or download.
- Empty existing directories are reused without a prompt. Every `readdir` error, including permission or wrong-type errors, is treated as if the destination did not exist.
- There are no guards against cwd, repository root, home, filesystem roots, or paths outside the current project. Because empty repository names are accepted, malformed input can make cwd the default destination.
- There is no temporary sibling, backup, rollback, or atomic publication. A network/extraction/clone failure can destroy the original destination or leave a partial tree.

### Provider and archive behavior

- With no explicit ref, the provider repository API is called to obtain `default_branch`. Failure is nonfatal: the command warns and later performs a clone using the remote default branch.
- GitHub.com uses `https://api.github.com/repos/<owner>/<repo>` and `/tarball/<ref>` with `Accept: application/vnd.github.v3+json`; configured tokens use `Authorization: Bearer`.
- Other GitHub bases use `<base>/api/v3`. This makes configured `http://github.com` behave like Enterprise rather than public GitHub.
- GitLab uses `/api/v4/projects/<percent-encoded owner/repo>` and its archive endpoint with encoded `sha`; configured tokens use `PRIVATE-TOKEN`.
- Archive fetch follows redirects. HTTP 401/403 gets an authentication hint; every non-OK, network, decode, or extraction failure falls through to clone.
- The complete compressed response and complete decompressed TAR are held in memory. Gzip decompression is synchronous.
- The custom TAR parser handles basic ustar names and GNU long-name records but does not validate checksum/bounds and does not implement PAX metadata. It returns files/directories and ignores links and other entries.
- Extraction strips the first path component, joins the remainder to the destination, and launches an unbounded `Promise.all` of directory/file writes. It does not validate containment. `../` escapes the destination; Windows separators add another escape surface.
- File modes, executable bits, mtimes, ownership, links, sparse data, and empty-directory metadata are not preserved. Parallel conflicting entries can race.

### Clone fallback and observable differences

Fallback argv is:

```text
git clone --depth=1 --single-branch [--branch <ref>] <credential-bearing URL> <absolute destination>
```

- After success, `<dest>/.git` is recursively removed. Other Git metadata such as `.gitmodules` remains.
- GitHub puts the raw token in URL userinfo; GitLab puts `oauth2:<token>` there. Reserved token characters are not encoded. The credential is visible in the child argv/process table and may be echoed by Git errors. This violates the required secret boundary.
- Archive and clone success are not equivalent: executable modes differ, archive vs Git LFS content may differ, Git filters/credential helpers may run only for clone, and neither path initializes submodules. Which content contract is authoritative must be decided.
- A partially extracted destination usually makes fallback clone fail because it is nonempty. No cleanup/rollback follows final failure.
- Clone failure exits 1. Successful archive or clone exits 0 and reports the original destination string. Default-branch/API and archive failures are nonfatal only when fallback succeeds.
- HTTP-configured instances send credentials without transport encryption. There is no timeout, size/file-count limit, checksum, signal-aware cancellation, or Windows-specific destination validation.

### Required Go tests

1. A table for every accepted input form, aliases/host precedence, nested owners, `.git`, refs with slashes/encoding, case/ports, malformed/empty parts, unsupported schemes, and SSH syntax.
2. Existing encrypted config decryption, unknown aliases/hosts, duplicate hosts, scheme precedence, and non-disclosure on every failure.
3. Local fake GitHub/GitHub Enterprise/GitLab APIs covering URLs, headers, redirects, default branches, auth errors, malformed JSON, missing fields, timeouts, and fallback decisions.
4. Streaming gzip/TAR fixtures for ustar/GNU/PAX names, modes, directories, links, long paths, malformed/truncated headers, duplicate/conflicting entries, traversal, absolute/Windows paths, file-count/size limits, and cancellation.
5. A dangerous-destination deny table, default-negative overwrite decision, empty/file/unreadable destinations, sibling staging, atomic replacement/backup, rollback, cross-device behavior, and Unix/Windows permissions.
6. Fake-Git clone argv and exit behavior, ref/no-ref, cleanup, partial archive fallback, no credentials in argv/logs, credential-helper behavior, hooks/filters/LFS/submodule policy, and archive/clone content equivalence.
7. Native Windows tests for reserved names, long paths, separators, file replacement, and cleanup; native Unix tests for modes and symlinks.

Risk: **critical**. The current command can delete broad targets before network success, write archive files outside the destination, and expose credentials through process arguments.

## `git cm`

### CLI mode matrix

Leaf options are:

```text
--profile <name>
--timeout-ms <milliseconds>
-l, --lang <en|zh>                 default en
-S, --staged
-s, --stage
-a, --stage-all
-p, --push [remote]
-c, --stage-push [remote]
-d, --dry-run
-b, --body
```

- `--timeout-ms` uses strict `Number` parsing and requires a safe integer at least 1000. This is stricter than the shared integer parser.
- Language is validated only after repository inspection, optional staging, and profile resolution; invalid values can therefore leave index mutations before failure.
- Default and plain `--dry-run` generate from all uncommitted changes and do not commit.
- `--staged` generates from the index and, unless combined with `--dry-run`, asks to create a commit.
- `--stage` selects files, rewrites index state, generates from the resulting index, and asks to commit.
- `--stage-all` runs `git add -A`, generates from the index, and asks to commit.
- `--stage-push [remote]` is `--stage` plus commit and push. `--push [remote]` is accepted only with a commit-producing stage/staged mode. Omitted remote means `origin`.
- `--body` permits a validated multiline message. Without it, model output must be one line.
- Stage/stage-push conflicts with stage-all and dry-run; any push conflicts with dry-run; push alone fails. Other redundant/contradictory combinations are accepted.
- Notably, `--stage-all --dry-run` does **not** stage all: it silently generates from whatever is already staged. `--stage --staged` is accepted. When both push forms are present, `stagePush || push` can ignore the requested `--push` remote. Preserve these defects for first-release parity; correction is post-parity work.

Optional flag values are Commander-specific. Cobra/pflag does not consume the following bare token identically, so `Prove the Go CLI compatibility approach` must test long/short, omitted, `=remote`, bare remote, and next-flag forms.

### Repository inspection and mutation

Repository root and change discovery use:

```text
git rev-parse --show-toplevel
git status --porcelain=v1 -z --untracked-files=all
git ls-files --stage -z
```

- Status is NUL-delimited and sorted by path. Renames/copies preserve destination plus original path.
- Entries whose index mode is `160000` are filtered as submodules. The command therefore does not actually model the complete index promised by the README.
- `--stage` initially selects every displayed file. Unselected tracked paths are reset with `git restore --staged -- <paths>`; selected existing paths use `git add -A -- <paths>`; missing paths use `git update-index --remove -- <paths>`.
- These paths are passed as Git pathspecs, not explicitly literal pathspecs. A real filename containing pathspec magic can select or reset a different set.
- Existing partial staging is overwritten for selected paths. Cancellation before selection mutates nothing; cancellation, provider failure, invalid output, or commit refusal after staging leaves the new index state in place.
- `--stage-all` stages all paths, including submodule changes later omitted from evidence. Pre-existing staged submodule changes also remain. If another modeled file exists, `git commit` can commit the omitted submodule change under a message that never represented it.
- No unresolved-conflict, intent-to-add, sparse checkout, submodule, or unborn-HEAD policy is declared.

### Snapshot and concurrency contract

For staged scope the command captures cached patch/numstat. For all-uncommitted scope it also captures worktree patch/numstat and every untracked file:

```text
git diff --cached --no-ext-diff --find-renames --unified=0
git diff --cached --numstat -z --find-renames
git diff --no-ext-diff --find-renames --unified=0
git diff --numstat -z --find-renames
git cat-file --batch
```

- Tracked patch parsing is textual and does not decode Git quoted paths; `cat-file --batch` specs are newline-delimited. Status/numstat are mostly path-safe, but tabs/newlines/quotes/pathspec magic can make patch, manifest, staging, and status views disagree.
- Up to eight untracked files are hashed/read concurrently. Tracked Git capture operations also run concurrently; no repository-wide snapshot lock exists.
- A SHA-256 snapshot ID covers scope, classified status, cached/worktree patch text, and untracked content hashes. Before commit the same scope is recaptured; a changed ID refuses the commit.
- For staged scope, later worktree-only edits intentionally do not invalidate the index snapshot. A change to the staged index does.
- There remains a time-of-check/time-of-use window between recapture and `git commit`, and no transaction spans staging, model I/O, confirmation, hooks, commit, and push.

### Evidence and security boundary

- Every modeled file receives a role: source, test, docs, config, dependency, generated, binary, sensitive, or unknown.
- Exact `.env`/`.env.*`, `id_rsa`, `id_ed25519`, and `.pem/.key/.p12/.pfx` paths are redacted. Matching is partly case-sensitive.
- Known binary extensions, Bun/npm/pnpm/Yarn lockfiles, `dist`, `build`, `coverage`, minified JS/maps, files over 200,000 bytes, and patches over 200,000 characters are metadata-only.
- Other text hunks are converted into structured directory context and selected semantic facts. The local estimate targets about 3,000 tokens and hard-stops at 4,000; facts are included whole or omitted whole.
- Protected content is excluded from directory/fact text, but contributes to local snapshot integrity. The provider sees summary counts, not protected paths/content.
- `package.json` receives structured dependency/script/version facts. Other configuration/source lines may be sent literally.
- The path list is deny-based rather than a complete secret boundary. Diagnostics proved literal secrets from `.npmrc` and `credentials.json` are sent. There is no content-value redaction for token/password/authorization assignments like the shared logger has.
- More critically, untracked paths are read through `Bun.file` after `lstat` without rejecting symlinks or nonregular files. A symlink can cause repository-external content to be hashed and sent to the provider. Named pipes/devices also lack an explicit policy.
- Repository-controlled filenames and content are model input. The system prompt says to use facts and ignore evidence instructions, but there is no terminal-control-character sanitation for Git metadata or returned model text.

These are documented security defects and remain observable first-release behavior. Reproduce the provider-data boundary and current redaction behavior in isolated tests; a no-follow containment and expanded secret policy is post-parity hardening.

### Provider and message contract

- Profile resolution uses explicit `--profile`, CM environment variables, stored default/first profile, and the existing encrypted shared config contract. `--timeout-ms` wins over environment/stored/default timeout.
- Exactly one OpenAI-compatible `POST <baseURL>/chat/completions` request is made with bearer token, model, temperature, system/evidence messages, and hard-coded `max_tokens: 4096`. DeepSeek V4 behavior uses the shared client rule.
- There is no retry. HTTP, JSON, empty response, and timeout errors exit 1 and print profile name/base URL/model but not the API key. Provider error bodies and invalid model output are printed and can contain echoed evidence.
- Output language is English or Chinese. The accepted subject grammar is exact type from `feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert`, a nonempty parenthesized scope, `: `, and nonempty description.
- Subject length and most control characters are not bounded. Without `--body`, any newline fails. With body, only Markdown fences in later lines are explicitly rejected.
- A single outer quote/fence wrapper is stripped. Invalid output is not retried.
- Human output includes message, selected profile/model, provider token usage if present, local token estimate, cluster/fact coverage, and a compaction notice. Exact note styling is presentation; secret non-disclosure and the semantic diagnostics are contracts.

### Commit, hooks, push, and exit behavior

- Commit confirmation inherits Clack default **yes**. Cancel/no exits 0 and keeps any earlier index mutations.
- Immediately before commit, the staged snapshot is recaptured. Commit argv is `git commit -m <subject> [-m <body>]`; it commits the complete current index and runs normal Git hooks/signing/configuration.
- Commit failure exits 1 and leaves Git's resulting state. Success without push exits 0.
- Push determines `git branch --show-current`, rejects detached HEAD, then runs `git push -u <remote-or-origin> <branch>`. Push failure exits 1 after the commit already exists; no rollback is attempted.
- A remote beginning `-` is accepted by Commander (for example `--push=-f`) and is placed where Git can parse it as an option. Preserve this argv behavior for the first release; literal/option-safe rejection is post-parity.
- No changes is a successful informational return. Selection cancellation/empty selection and commit refusal are exit 0. Invalid option combinations, repository/profile/provider/model/snapshot/commit/push failures are exit 1.
- There is no explicit non-TTY policy for selection/confirmation and no command-owned cancellation context for Git or HTTP work.

### Required Go tests

1. Full legacy flag-combination truth table, Commander-compatible optional remote forms, legacy language/timeout parsing, and observed `--stage-all --dry-run` semantics.
2. Disposable-repository status/snapshot tests for staged, unstaged, partial, untracked, delete, rename/copy, binary, large, symlink, submodule, conflict, intent-to-add, unborn, sparse, worktree, detached, and bare states.
3. Literal filenames containing whitespace, tabs, newlines, Unicode, leading dashes, colons, glob/pathspec magic, quotes, and backslashes across status, patch, manifest, staging, hashing, and commit.
4. A deny/allow matrix for sensitive paths (including case), credentials/configs, generated/binary formats, tracked/untracked symlinks, FIFOs/devices, repository escape, and assignment-level secret redaction. Assert exact provider request bodies never contain sentinel secrets.
5. Deterministic snapshot/evidence/hash behavior, 3,000/4,000 budgets, cluster fairness, manifest facts, protected counts, provider request count/body/usage, DeepSeek rule, timeout, and every model-output validation error.
6. Stage selection and partial-index behavior, literal pathspecs, cancel/error index policy, submodule visibility, stale-snapshot races, hook success/failure/mutation, signing, commit subject/body, and complete committed-tree agreement with modeled scope.
7. Push default/custom remote, upstream setting, detached HEAD, option-like remote rejection, credential/helper prompts, hook behavior, push failure after commit, and explicit result reporting.
8. Non-TTY and redirected I/O, `NO_COLOR`, narrow terminals, control-character sanitation, Ctrl-C during Git/provider/prompt work, no orphan processes, and native Windows/Unix behavior.

Risk: **critical** for provider-data containment and committed-scope integrity; **high** for CLI/mutation compatibility.

## Confirmed defects retained for post-parity disposition

| Area | Legacy behavior | First-release disposition |
| --- | --- | --- |
| Heat paths | Line parsing exposes Git quoting instead of real arbitrary paths | Preserve observable output; use byte-safe parsing only where output remains identical |
| Heat copy/merge semantics | Output depends on Git defaults/configuration | Preserve and test representative Git configurations |
| Pulse days | Accepts zero, negatives, and numeric prefixes | Preserve the Commander parsing matrix |
| Pulse repository set | Ignores linked worktrees/bare repos and silently treats failures as no commits | Preserve and test; broader discovery is post-parity |
| Fork destination | Default-yes recursive deletion before network success, with no dangerous-target guard | Preserve in disposable tests; transactional hardening is post-parity |
| Fork extraction | TAR traversal, malformed/PAX gaps, unbounded memory/concurrency, partial publication | Preserve accepted archive behavior; safety bounds are post-parity |
| Fork credentials | Raw token in clone argv/URL and possible diagnostics; reserved characters break URL | Preserve first-release behavior; credential hardening is post-parity |
| Fork source parity | Archive and clone can produce different modes/LFS/filter content | Preserve both legacy paths and their differences |
| CM provider boundary | Follows untracked symlinks outside repo and leaks ordinary config token values | Preserve the observed request boundary; redaction hardening is post-parity |
| CM scope integrity | Submodule changes can be committed without model evidence; pathspecs are nonliteral | Preserve and test the observed staging/commit scope |
| CM side effects | Staging precedes validation/provider/confirmation and remains after cancel/failure | Preserve and test the mutation timing |
| CM option/push safety | Contradictory modes and option-like remotes are accepted | Preserve the legacy argv and Git invocation matrix |

## Go migration boundaries and readiness

1. Keep an injected external-Git runner as the authoritative adapter for repository discovery, log, status, diff, staging, commit, hooks, and push. Do not adopt `go-git` for these operations without a separate parity proof; its object/worktree APIs do not automatically reproduce CLI config, hooks, credential helpers, filters, signing, or error behavior.
2. Git output parsing must be command-owned and byte/path safe. Prefer NUL-delimited plumbing, explicit formats, literal pathspecs, and `/` for Git-internal paths; use `filepath` only at the local filesystem boundary.
3. `git heat` is the lowest-risk Git leaf after CLI foundations. `git pulse` follows using the repository identity, partial-failure, fixed-clock date, and cancellation behavior already recorded here.
4. Shared encrypted config must land before `git fork` or `git cm`.
5. `git fork` keeps acquisition and publication behind one command-owned module, but its first release reproduces the legacy destination, archive/clone, and credential behavior. Transactional and credential hardening is later work.
6. `git cm` should be last among Git leaves. Keep snapshot capture, legacy evidence policy, model transport, and mutation orchestration behind one deep command module; do not expose generic Git/provider helpers as pass-through layers.
7. Completion requires command-level Go tests, not only ported pure-function tests. The final gates must prove that the modeled CM scope equals the committed index, secrets never cross the provider boundary, fork cannot escape or destroy its destination on failure, and cancellation leaves no child or partial publication.

This inventory resolves local fact finding. It does not select final Go packages, implement any command, or authorize post-parity behavior changes.
