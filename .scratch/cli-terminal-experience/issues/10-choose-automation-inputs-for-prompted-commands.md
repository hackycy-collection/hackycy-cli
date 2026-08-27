# Define automation fallback for prompted commands

Type: grilling
Status: resolved
Blocked by: 03

## Question

The completed command inventory shows that several public commands currently
depend on terminal prompts. This effort must not add flags, positional
arguments, secret-injection surfaces, implicit defaults, or changed command
workflows to make those prompts work in an Automation Session.

For each affected command, decide whether its existing invocation already has
enough declared input to proceed, or must fail before a side effect because it
would otherwise prompt. Cover `config fork add/remove`, `config cm
add/remove`, `export env`, `rm`, `run`, `git pulse`, `git fork`, `git cm`,
`zip`, and remembered `tunnel connect`. Define the terminal-only fallback for
secrets, destructive confirmation, selection/multiple selection, cancellation,
error wording, and incomplete existing input. stdin remains unavailable as an
implicit answer channel.

## Comments

- 2026-08-26: Scope corrected by the user. The prior question wording
  incorrectly invited new Automation Input grammar. Terminal modernization
  preserves current command grammar and business workflows; it only defines
  how existing prompt-dependent paths fail when prompting is unavailable.

## Answer

Approved on 2026-08-26. This is a terminal fallback decision, not a new
Automation Input grammar. In an Automation Session, an existing command path
continues only when its current invocation, configuration, and established
defaults already determine every required value without an Interaction
Request. A Prompt-Dependent Path fails before it can mutate state, write
configuration, publish an archive, or launch a child process.

The failure is one command-specific, plain `error:` line on stderr and exit
code `1`. It writes no Command Result to stdout and never renders Help, ANSI,
progress, a Transient View, or `Cancelled`. ycy neither reads stdin as an
answer nor opens a controlling terminal, infers an answer, or creates a new
flag, positional value, secret source, or shortcut. Existing more-specific
validation errors remain intact, especially Tunnel ambiguity errors. Signal
cancellation keeps its existing command-owned exit behavior.

### Existing-input matrix

| Command path | Automation behavior |
| --- | --- |
| `config fork add` | Always fails: the multi-field and secret form is prompt-dependent. |
| `config fork remove` | An empty configuration keeps its existing no-selection result; a path requiring instance selection or confirmation fails. |
| `config cm add` | Always fails: the multi-field and secret form is prompt-dependent. |
| `config cm remove <profile>` | A valid removal path fails because its existing confirmation remains required. Existing validation runs first. |
| `export env` | An already unambiguous environment proceeds. A selector path fails. The existing `--env` behavior is unchanged. |
| `rm <paths> --force` | Proceeds under its existing force semantics. A confirmation path and smart-cleanup action or target selection path fail. |
| `run` | Always fails because it selects both a script and a package manager before launching the child process. |
| `git pulse` | An existing `--days` value skips date selection. Multiple-author selection fails before report generation; zero or one author proceeds. |
| `git fork` | A new or empty destination proceeds. A nonempty destination fails before the existing overwrite-removal path. |
| `git cm`, including `--dry-run` | Generation-only paths proceed. File selection and commit or push confirmation paths fail. |
| `zip` | Always fails because its existing archive planning is a multi-step prompt flow. |
| `tunnel connect` | Preserve existing CLI, environment, token-file, remembered, and default resolution. A uniquely resolved connection proceeds; ambiguous remembered-connection selection retains the existing specific ambiguity error and does not prompt. |

Error wording may name only an existing way to avoid the interaction, such as
an already-supported `--env` or `--force` invocation. It must not advertise a
new Automation-only interface. The migration plan must add focused tests that
prove each failing path has no side effect and that each already-unambiguous
path preserves its current output and exit behavior.
