# Inventory every current command terminal surface

Type: task
Status: resolved

## Question

For the root surface and every leaf command, what is currently written to
stdin/stdout/stderr across success, empty, cancellation, validation failure,
operational failure, progress, redirection, and automation paths; which output
is a Live View, Command Result, Diagnostic Record, raw child stream, or
Lifecycle Log; and where are states currently absent, duplicated, lost by the
alternate screen, or covered by tests?

## Answer

Completed the root/process and all 22 leaf-command inventory in
[`03-current-command-terminal-surfaces.md`](../inventories/03-current-command-terminal-surfaces.md).
The current runtime separates Live View on stderr from Command Results on
stdout and correctly releases `run` before raw child handoff, but it has no
Interaction Transcript ledger: Rich questions, answers, notices, and phases are
lost with AltScreen. Only Git Pulse, Git Fork, and Git CM use typed phases;
Diff/FS expose stdout startup results while Tunnel alone has Lifecycle Logs;
Automation suppresses all notices and phases; and several command-specific
partial, duplicated, silent, or exit-0 failure states must be preserved and
made explicit by their later presentation tickets. Existing tickets cover the
newly concrete gaps, so no child ticket was added.
