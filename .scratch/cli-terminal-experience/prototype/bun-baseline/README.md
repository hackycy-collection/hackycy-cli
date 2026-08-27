# Bun Baseline Feasibility

This is a throwaway assessment for the terminal visual-language decision. It
uses the output and event order in `legacy/bun/` as the reference. It does not
change production code or approve a new output contract.

## Approved parity target

The approved target is visual and behavioral parity with the Bun journeys:
their command-local title behavior, prompt/work/result sequence, English
wording, semantic colors, spacing, and report structure. Huh and Lip Gloss
may use their native control chrome; ycy does not need to reproduce Clack's
exact glyphs, borders, or escape sequences.

## Baseline

The Bun implementation has four separable concerns:

1. prompt state and cancellation (`@clack/prompts`);
2. transient work state (Clack spinners and Ink's final tree);
3. durable command text (messages, summaries, and outcomes); and
4. diagnostic logging (the scoped stderr logger).

The Go migration already has a seam for each concern in three of the four
journeys. The missing piece is stage progress for `git cm`.

## Journey Mapping

| Journey | Bun reference | Existing Go seam | Feasibility |
| --- | --- | --- | --- |
| `git cm` | `legacy/bun/src/commands/git/cm/run.ts`: file multiselect, stage/generate/commit/push spinners, generated-message note, confirm, green outro | `StagePrompter`, `CommitPrompter`, `GeneratedMessage`, and `Result` in `internal/commands/git/cm` | High. Huh multiselect/confirm and Lip Gloss note are direct adapters. Add an optional phase-event/progress boundary for the four spinner phases; do not infer phases in a goroutine from Git calls. |
| `git pulse` | `legacy/bun/src/commands/git/pulse/pulse.ts` and `components/CommitTree.tsx`: title, workspace line, scan/fetch spinners, date/author selections, colored repository tree | `Prompter`, `Presenter`, `Report`, and scan/fetch callbacks in `internal/commands/git/pulse` | Very high. Huh selections and a Lip Gloss renderer can preserve the report. The existing Presenter events can feed a Bubble Tea transient view if that view is approved. |
| `config cm add` | `legacy/bun/src/commands/config/cm.ts`: cleared title, intro, four ordered questions, hidden API key, save spinner, success outro | `AddPrompter`, `AddPresenter`, `AddInput`, and `SaveAdd` in `internal/commands/config/cm` | High. Huh `Form`/`Input`/password fields map to the existing contract. Add a save-phase status adapter if the spinner is retained. The current visible-stdin password fallback must be removed for Automation Sessions. |
| `tunnel connect` | `legacy/bun/src/commands/tunnel/index.ts`: remembered-connection select with `maskTunnelToken`, then stderr lifecycle logs | `ClientConnectionSelector`, `ResolveClientConfig`, and scoped logging in `internal/commands/tunnel` | High. Huh select can replace only the selector. The current adapter checks stdin/stdout only; it must receive the approved three-stream Session classification. |
| `tunnel server` | `legacy/bun/src/commands/tunnel/server/run.ts` plus `legacy/bun/src/shared/log/index.ts`: timestamp, level, scope, redacted JSON context on stderr | `logging.Runtime` and scoped Tunnel logger calls | High. This is a formatter change, not a TUI. Current Go output is close but differs in scope brackets, debug/error styling, and `NO_COLOR` handling. |

## What "matching Bun" means

The following are reproducible without importing the old runtime:

- English wording and the order of prompts, work phases, and outcomes;
- green/cyan/yellow/red/gray semantic roles;
- the generated-message metadata and compaction warning;
- the Pulse repository grouping, sorting, connectors, dates, authors, and
  subjects;
- masked tunnel tokens and stderr-only service diagnostics; and
- clear-screen/title behavior when, and only when, the Session is Rich
  Interactive.

Huh's stock theme is not Clack's renderer. Its default checkbox/radio glyphs,
guide bars, and prompt layout are different. If byte-for-byte Clack output is a
hard requirement, each prompt would need a custom Bubble Tea/Lip Gloss model,
which largely gives up Huh's value. The practical parity target is therefore
the Bun information architecture and visual roles, with Charm-native prompt
chrome.

## Required Go Changes (later implementation tickets)

1. Classify the Session before constructing prompt or transient-view adapters;
   Automation must never let Huh or Bubble Tea open a controlling TTY.
2. Introduce a shared terminal adapter package in the composition layer. Keep
   `internal/commands` dependent only on its existing semantic interfaces.
3. Give `git cm` explicit phase events (collect/stage, generate, commit, push)
   and let the Rich adapter render them as transient status. Plain and
   Automation adapters emit deterministic lines.
4. Implement one Lip Gloss palette and width-aware report helpers, then use
   them in Pulse, Git CM, and configuration presenters. `NO_COLOR` removes
   styles only; it does not change the Session kind.
5. Replace the config password fallback with terminal-only echo-disabled input;
   missing Automation Input must fail before a write.
6. Bring the Go logger formatter to the Bun baseline: bracketed scopes,
   gray/debug and bold-red/error roles, deterministic redaction/context, and a
   `NO_COLOR`-aware color decision.

## Verification Shape

The implementation can be proven without running Bun:

- semantic adapter tests for every existing prompt and Presenter event;
- Rich PTY snapshots for the four journeys, including Ctrl-C cleanup;
- Plain Interactive snapshots with `TERM=dumb`; and
- redirected/`CI` tests proving no prompt, ANSI, cursor control, or implicit
  `/dev/tty` access, while preserving current stdout results and stderr logs.

**Verdict:** the approved Bun parity target is implementable in Go. No business
Module or Cobra command tree needs to be replaced.
