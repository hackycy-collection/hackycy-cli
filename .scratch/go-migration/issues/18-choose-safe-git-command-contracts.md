# Choose safe and failure-aware contracts for Git commands

Type: grilling
Status: resolved
Blocked by: 05

## Question

Which corrected contracts replace the confirmed unsafe or contradictory Git-command behaviors while retaining the intended workflows?

For `git heat` and `git pulse`, decide copy/merge counting, the repository set (normal repositories, linked worktrees, bare repositories, nested repositories, and submodules), strict day ranges, and whether partial per-repository failures warn, fail, or produce a distinct partial-success result.

For `git fork`, decide valid repository/ref grammar; dangerous-target denial; overwrite confirmation/default; transactional staging, backup, publication, rollback, and cleanup; the authoritative content/mode/LFS/submodule contract across archive and clone paths; archive entry/type/resource limits; and a credential transport that never puts tokens in argv, URLs shown to users, or errors. State the policy for explicitly configured plain HTTP instances.

For `git cm`, decide the provider-data boundary for symlinks, nonregular files, repository escape, sensitive paths, and assignment-level secret redaction; whether protected filenames may be sent as metadata; how submodules and every committed index entry remain represented; literal path handling; when staging changes become durable and what cancellation/provider/commit failures leave behind; the corrected flag-combination matrix; option-like remote rejection; snapshot/commit race handling; and terminal-control-character sanitation.

Record exact defaults, prompts, exit codes, partial-success reporting, rollback guarantees, and acceptance tests. `Prove the Go CLI compatibility approach` still owns the parser mechanics for optional values and integer/run argv compatibility; this ticket owns the product behavior selected on top of that parser.

## Comments

## Answer

Closed as out of scope for the parity-first Go release. `git heat`, `git pulse`, `git fork`, and `git cm` are ported from the Git compatibility inventory, including observable legacy quirks and defects. The proposed validation, transactional, credential-transport, redaction, race, and safety redesign remains post-parity hardening. A narrow decision may be opened during a Git leaf's implementation only if a parity test proves that Go cannot reproduce the Bun behavior under the migration's hard constraints.

2026-08-23 parity clarification for G18: `git cm --stage-all --dry-run` follows frozen `legacy/bun/src/commands/git/cm/run.ts`, which selects the `all-uncommitted` snapshot scope after suppressing `shouldStageAll` for dry runs. This deliberate first-release behavior supersedes the prior contradictory inventory wording.
