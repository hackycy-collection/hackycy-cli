# Inventory command experience and output contracts

Type: task
Status: resolved

## Question

Before choosing migration slices, create a complete inventory of every public
ycy command that currently prompts, renders user-facing progress, streams a
child process, writes a command result, emits diagnostics, or exposes a
long-running lifecycle. For each, record its current adapters, stdout/stderr
behavior, whether output is machine-consumed or exact-output tested, terminal
requirements, cancellation behavior, and its likely interaction archetype:
plain result, Huh form, Huh confirmation, static styled report, or Bubble Tea
dynamic view.

The task is complete when the inventory identifies the compatibility evidence
that any later design or migration plan must retain. It should be linked from
this ticket rather than pasted into the map.

## Comments

- Claimed for repository evidence collection on 2026-08-26.

## Answer

The complete active-Go inventory is recorded in
[`03-command-experience-output-contracts.md`](../inventory/03-command-experience-output-contracts.md).

It covers every public command leaf and the root/help/completion surface,
including current adapters, stream ownership, interaction and cancellation,
exact-output and standalone evidence, and likely presentation archetypes. The
main migration constraints are mixed stdout ownership, direct child-process
passthrough for `run`, stdout readiness lines for `diff`/`fs`, stderr Tunnel
logs, Upgrade's deliberate mixed-output/exit exceptions, and the visible-input
fallback for configuration secrets.

The inventory made one previously foggy decision precise: whether individual
prompted commands receive narrow Automation Session inputs or fail explicitly.
That work was graduated into [Choose automation inputs for prompted
commands](10-choose-automation-inputs-for-prompted-commands.md).
