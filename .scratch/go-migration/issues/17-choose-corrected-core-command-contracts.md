# Choose the first-release parity and compatibility-exception policy

Type: grilling
Status: resolved

## Question

What is the first-release compatibility policy for porting observable Bun behavior to Go, and what concrete evidence is required before an implementation mismatch may interrupt command-by-command migration for a human decision?

## Comments

- 2026-08-22, grilling round 1:
  - Q33A: adding an existing Fork alias or CM profile requires an immediate default-no replacement confirmation. Cancel or no leaves state unchanged and exits `0`; yes performs one atomic replacement; a concurrent change detected before publication leaves state unchanged and exits `1`.
  - Q34A: a Fork host is an authority only: hostname or IP, with an optional port, and no scheme, path, query, fragment, or user information. A CM base URL must be an absolute HTTP(S) URL with no credentials, query, or fragment; a path such as `/v1` is valid.
  - Q35A: validate only the effective CM numeric environment value selected by precedence. Temperature must be finite and in `[0,2]`; timeout must be an integer `>=1000` milliseconds; max output tokens must be an integer `>=32`. An invalid selected value fails closed with exit `1`; shadowed values need not be validated and there is no silent fallback.
  - Q36A: `rm` permanently denies filesystem, volume, and UNC roots; the user's home directory; the current working directory; the repository root; and every ancestor of those locations. Resolve intermediate symlinks for this comparison, while a final symlink denotes only the link itself. There is no escape hatch, and `--force` cannot bypass the deny policy.
  - Q37A: `zip --with-dir` accepts exactly one portable directory component. Unsafe input is rejected rather than sanitized.
  - Q38A: an existing ZIP triggers a default-no prompt before collection or compression. Yes writes a candidate beside the destination and atomically publishes it. Invalid input, no matches, collection, compression, candidate-write, or publication failure exits `1`; cancel/no and reveal failure exit `0`.
  - Q39A: preserve only the legacy archive metadata contract: regular-file entry paths and bytes plus one archive-creation timestamp shared by entries. Do not preserve source modification times, permissions, explicit directory entries, symlinks, ownership, ACLs, or extended attributes.
- 2026-08-22, grilling round 2:
  - Q40A: confirmed CM replacement changes only `baseURL`, `model`, and `apiKey`; it retains `temperature`, `timeoutMs`, `maxOutputTokens`, and the default-profile selection. Confirmed Fork replacement replaces the complete alias record.
  - Q41A: every explicitly missing operand or operand whose identity cannot be determined with `lstat` contributes to final exit `1`; valid operands may still be confirmed and deleted. All-invalid exits `1`, and `--force` does not suppress operand errors. An existing operand error still yields `1` if deletion of the remaining valid operands is declined; a pure cancellation with no operand error exits `0`.
  - Q42A: any protected target aborts the complete `rm` request before confirmation or mutation. Safe operands are canonicalized with final-symlink identity preserved, deduplicated, and descendants already covered by a selected real directory are collapsed. The command displays only the deterministically ordered deletion roots.
  - Q43A: after confirmation, `rm` attempts every independent deletion root. Successful deletion remains committed; any failed or partially deleted root is reported as failed and makes the final status `1`. The final deterministic summary distinguishes succeeded, failed, and preflight-missing operands; no rollback is claimed.
  - Q44A: smart directory actions match only real directories and the lockfile action only regular files; final symlinks and wrong-type names are not matches, and scans never traverse symlinks. Any unexpected read or type-inspection error makes the scan incomplete and exits `1` before selection or mutation. Policy-ignored directories are not errors; a complete scan with no matches exits `0`; matches are sorted by relative path.
  - Q45A: smart recursive depth defaults to `5` and accepts `0..64`. It is an edge count from cwd at depth `0`, so direct children are depth `1` and a target exactly at the bound is included; depth `0` selects no descendants. The value is range-validated even when the later selected action is nonrecursive.

- 2026-08-22, scope correction: the preceding Q33-Q45 selections are retained only as post-parity hardening notes and are not requirements for the first Go release. Q46-Q51 were proposed after the route had already drifted and are void; they require no answer. The ZIP metadata observation in Q39 remains useful parity evidence, not authorization to change behavior.
- 2026-08-22, transition-scope amendment from [Approve the command-by-command migration roadmap](16-approve-command-migration-roadmap.md): guaranteed Bun-written direct read is narrowed to `config.json`; the first Go install requires an installer or manual replacement rather than Bun `ycy upgrade`; and Tunnel interoperability is Go-to-Go v3 only. These explicit owner-selected exclusions do not require a failed parity probe and supersede the broader data/protocol transition language below without authorizing unrelated behavior redesign.

## Answer

The first Go release is an observable-behavior port of the frozen Bun implementation. Tests written from `legacy/bun/` define the starting contract command by command, including parser quirks, accepted unsafe edges, prompts, defaults, exit statuses, side effects, data formats, protocols, and known defects. The migration must not silently tighten validation, add confirmations or limits, change cancellation/failure meaning, make mutations transactional, or otherwise harden product behavior merely because the inventory found a defect.

A compatibility exception may be raised only after a focused Go implementation or probe demonstrates that the legacy behavior cannot be reproduced because of Bun/Node runtime semantics, the `CGO_ENABLED=0` constraint, a required operating-system or architecture difference, the behavior of an available maintained Go dependency, or direct readability of an existing data/protocol format. The exception ticket must name the affected command and failed parity case, include reproducible evidence, explain user-visible impact, and present only the smallest viable alternatives. Unproven concerns remain inventory notes, not questions or blockers.

Only Bun-written `~/.ycy-cli/config.json` remains direct-read by default. If that direct read is proven impossible, the only permitted fallback is an automatic, backed-up, failure-detectable compatibility plan chosen through Wayfinder before integration. Bun-written FS/Tunnel runtime state, Legacy Update State, Bun-to-first-Go Upgrade, and mixed Bun/Go Tunnel peers are excluded from the guarantee: Go starts normally, logs diagnosable failures, and internal operators manage any non-config residue. Public Upgrade remains Go-to-Go and Tunnel retains v3 for Go peers. Known hardening opportunities remain separate and do not gate the first Go artifacts.
