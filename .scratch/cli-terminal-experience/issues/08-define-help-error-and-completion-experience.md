# Define help, error, and completion experience

Type: grilling
Status: resolved
Blocked by: 04, 05

## Question

How should the existing Cobra command tree present help, validation failures,
unknown commands, suggestions, examples, completion guidance, and recovery
actions while retaining its public command grammar and automation-safe plain
text? Decide whether global help templates, command-group summaries, error
rendering, and shell-completion messaging belong to the terminal-experience
Module or remain narrow Cobra configuration.

The decision must preserve predictable `--help` output in Automation Sessions
and avoid embedding transient UI behavior in Cobra's parser layer.

## Answer

Approved on 2026-08-26. Cobra remains the command-tree and grammar Module.
It owns Help content, direct descendants, flags, argument grammar, and the
candidate lookup used for parser-recovery guidance. `internal/terminal` never
inspects a Cobra command tree, and `internal/cliapp` never imports
Charmbracelet.

### Command discovery

`internal/cliapp` produces a stable Command Discovery Document: the command
summary, usage, immediate descendants, flags, and approved examples. A thin
Terminal Adapter passes it to `internal/terminal` for presentation:

- Rich Interactive Sessions use a compact, static, semantically colored
  document. It may improve scanability and width handling, but has no title
  clear, Huh form, Bubble Tea program, animation, or command dashboard.
- Plain Interactive and Automation Sessions retain stable, unstyled,
  Cobra-compatible text on stdout. Public command syntax, flag spellings,
  visible-command topology, Help exit codes, and discovery output ownership
  remain unchanged.
- The root and command groups keep short discovery content. Leaf commands
  whose syntax, impact, or typical flag combination is not obvious receive
  one or two hand-written, safe examples. Examples cannot contain secrets,
  personal paths, fictional Automation Inputs, or behavior not yet approved
  by the Automation Input decision.
- Existing command nesting remains the information architecture. Direct child
  commands render in stable alphabetical order, with clear short
  descriptions; Rich rendering does not invent category headings that plain
  Help lacks.

This presentation path applies equally to `--help` and the explicit Cobra
Help command. It does not reclassify Help as a Diagnostic Event or suppress it
under logging controls.

### Parser recovery and errors

User-Actionable Errors remain exactly one plain `error: <message>` line on
stderr with their established exit semantics, in every Session. They never
use ANSI styling or a Transient View.

For an unknown command or flag, `internal/cliapp` may append recovery guidance
only when it has exactly one high-confidence direct candidate. The same line
then names that candidate and gives its exact Help invocation, for example:

```text
error: unknown command 'pulsee'; did you mean 'pulse'? Run 'ycy git pulse --help' for usage.
```

When no such candidate exists, the line instead directs the user to the
nearest valid parent's `--help` command. ycy does not print a full Usage block
after an error, emit multiple suggestions, or make speculative/ambiguous
guesses. Existing command-specific validation errors retain their current
messages unless a migration slice deliberately makes them more actionable
within this one-line contract.

### Completion

`ycy completion bash|fish|powershell|zsh` remains a raw script producer: it
writes only the unstyled generated script bytes to stdout, with no title,
prefix, diagnostic, install instruction, or Rich Interactive wrapper. This
keeps `source` and `eval` consumers safe. Its Help page is a normal Command
Discovery Document and may include a shell-specific, copyable invocation
example; generation output itself never does.

### Seams and evidence

The implementation keeps Cobra configuration narrow in `internal/cliapp`:
plain discovery serialization, candidate calculation, error normalization,
and raw completion generation remain there. Terminal Adapters translate only
the discovery document to Rich presentation. Tests cross the seams through
plain-output golden/regression cases in `cliapp`, recording-Experience Adapter
tests for semantic documents, and a small PTY suite for Rich layout. The
acceptance plan must additionally prove raw completion bytes, no ANSI in
Automation Help/errors/completion, stable single-line error recovery, and
unchanged parser exit codes.
