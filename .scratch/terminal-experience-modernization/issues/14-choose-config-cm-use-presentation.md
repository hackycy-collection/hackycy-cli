# Choose the config cm use terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact lookup and persistence Work Phases, selected-profile presentation,
Command Result, and missing-profile/failure treatment should `config cm use`
use while preserving its noninteractive syntax and exit behavior?

## Answer

`config cm use <profile>` remains a noninteractive, scriptable mutation. It
does not prompt, list profiles, confirm, or read stdin. The command presents
the existing atomic `SetDefaultCMProfile` operation as one real Work Phase,
because profile lookup and default persistence occur inside the same locked
store transaction. It does not add a preflight read or split unobservable
transaction internals into artificial phases.

### Command and presentation

- Rich uses the eyebrow `YCY / config cm use`, title
  `Set default CM profile`, and generic subtitle `Choose the stored profile
  used for commit message generation`. The requested profile appears only in
  a bounded detail area, not concatenated into a fixed-size title.
- The Live View shows the safe requested profile and one spinner with active
  detail `Checking profile and saving selection`. It does not display the
  profile catalog, previous default, Base URL, model, API key, configuration
  path, timing, or a percentage.
- The command keeps its exact `use <profile>` syntax and `cobra.ExactArgs(1)`
  behavior. It never introduces Huh Select/Confirm interactions. Automation
  and redirected execution remain supported.

### Atomic Work Phase

The immutable catalog contains exactly one phase:

| ID | Name |
| --- | --- |
| `set-default-cm-profile` | `Set default CM profile` |

The Experience opens before store construction and the phase becomes Active
immediately. It spans store creation and the single
`SetDefaultCMProfile(profile)` call, then becomes Completed or Failed. Rich
always renders the spinner and Plain always emits
`Setting default CM profile...` to stderr, even when the operation completes
quickly; neither mode sleeps to prolong it. Automation suppresses phase UI,
Transcript, color, and control sequences.

The writer is called at most once with the exact argument bytes. The command
does not trim, normalize, pre-resolve, or replace the profile identity. A
target that is already the default still goes through the writer and produces
the normal success result. The command does not call `ResolveCMProfile` and
therefore never reads or decrypts API keys for presentation.

### Safety and Transcript

Live View, phase detail, Transcript, and diagnostic context use a bounded,
single-line, control-free profile projection. If the requested identity cannot
be displayed safely, they use `Requested profile` instead. The original
identity remains the writer key and the semantic value in the compatibility
Command Result; the shared output layer strips terminal controls without
changing profile lookup behavior.

No standalone `Requested profile` Milestone is created. Under the shared
contract, a Milestone would also emit a Plain line and duplicate its lifecycle
output. Instead, the phase's terminal detail carries `Profile: <safe name>`.
The successful Rich Transcript is:

1. `Set default CM profile` completed, detail `Profile: <safe name>`
2. `Succeeded`

Failure records the same safe profile detail on the Failed phase, followed by
`Failed`. Store and writer error text, paths, profile metadata, credentials,
the complete Command Result, spinner frames, and timing never enter the
ledger. Rich restores the primary screen before replaying this Transcript.
Plain has already emitted its lifecycle line and does not replay it;
Automation has no Transcript.

### Results and failures

The one-shot Finish mapping is:

| Branch | Finish outcome | Command Result |
| --- | --- | --- |
| Writer succeeds, including an already-default target | `Succeeded` | `Default CM profile set to <name>` |
| Store construction, missing profile, or persistence failure | `Failed` | no stdout document |

Success preserves the existing stdout wording. Plain writes only
`Setting default CM profile...` to stderr before that result; it adds no
completed diagnostic. Automation emits only the existing stdout result.
Missing profiles preserve the original `CM profile not found: <name>` error,
exit behavior, and no-publish guarantee. All business failures mark the phase
Failed with generic safe detail `Unable to set default CM profile`, keep
stdout empty, and return the original error for root's single redacted
diagnostic.

There is no `Cancelled` Command Result. The store transaction has no truthful
mid-operation cancellation boundary, so the terminal does not claim one or
provide a no-op cancellation callback. It drains status until the one writer
call returns; a real writer success or failure determines the business
outcome even if the context is cancelled. The writer is never retried.

If persistence succeeds but `Finish` cannot write stdout, the phase and
Transcript remain Succeeded and the output error is returned. The command does
not retry or roll back the mutation, submit a second result, or relabel the
business operation as Failed. `Finish` is called at most once and `Close` is
cleanup-only. A Rich startup failure may fall back to Plain only before
AltScreen or mutation begins; later renderer failures restore the terminal
without repeating the writer.

### Evidence

Acceptance covers Rich PTY title/detail, fast spinner and terminal phase
states, primary-screen restoration, safe Transcript ordering, Plain stderr/
stdout ordering, Automation silence beyond its existing result/error, exact
profile identity including whitespace, already-default success, missing
profile no-publish behavior, serialized concurrent configuration updates,
API-key non-access, malicious/long profile names, store/writer/context/
renderer/stdout failures, at-most-once writer and Finish calls, and unchanged
arguments, help, exit codes, and result wording.
