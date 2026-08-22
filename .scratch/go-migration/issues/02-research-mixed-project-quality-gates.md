# Research quality gates for a Go and Vite mixed project

Type: research
Status: resolved

## Question

How do mature, actively maintained open-source repositories containing both Go and pnpm/Node frontend code install and scope Git pre-commit checks? Use first-party repository configuration and official tool documentation. Compare tracked native hooks, hook managers, staged-file filtering, fast pre-commit versus full local checks, Go formatting/vetting/testing, frontend lint/typecheck/test coverage, Windows behavior, dependency bootstrapping, bypass semantics, and stale-hook cleanup. Recommend a policy for ycy that checks both stacks without leaving Bun or obsolete hook-manager state.

## Comments

- Research context: branch `research/mixed-project-quality-gates`; report `.scratch/go-migration/research/mixed-project-quality-gates.md`.

## Answer

Primary-source research is complete at commit `2570d2fcc862b2caabbff74a79fbe7c9ea3d02fd` on branch `research/mixed-project-quality-gates`. The recommended direction is a repository-root Lefthook configuration pinned in an isolated Go tool module, with explicit bootstrap/install/doctor/uninstall commands and no pnpm lifecycle ownership. Pre-commit should be parallel, offline, non-mutating, and staged-file-aware: always run `git diff --cached --check`, check staged Go formatting, and lint/format-check staged active `web/` files. Full `CGO_ENABLED=0` Go format/vet/test and frontend lint/typecheck/Vitest/Vite build belong in `make check`.

Migration must remove only an exact known `simple-git-hooks` generated `bun run lint` hook, preserve and stop on unknown custom hooks or `core.hooksPath`, resolve linked-worktree Git directories correctly, and verify Windows filenames and command-length behavior. Go-only commits must not require pnpm; no hook may install dependencies or access the network. Ticket `Choose the mixed-project Git hook policy` remains the human decision point for adopting or adjusting this recommendation.
