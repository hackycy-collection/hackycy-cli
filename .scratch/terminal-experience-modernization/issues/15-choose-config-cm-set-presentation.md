# Choose the config cm set terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact validation and persistence Work Phases, sensitive-key redaction,
Command Result, and unsupported/missing/invalid failure treatment should
`config cm set` use while preserving its noninteractive syntax and value
semantics?

## Answer

`config cm set <profile> <key> <value>` remains a noninteractive, scriptable
atomic mutation. It does not prompt, confirm, list profiles, or read stdin.
The presentation models the existing `SetCMProfileValue` transaction as one
real Work Phase because profile validation, key dispatch, value parsing or
normalization, credential encryption, and publication all occur inside the
same locked store operation. It does not add preflight reads, duplicate
parsers, or artificial validation/persistence phases.

### Command and presentation

- The syntax remains exactly `set <profile> <key> <value>` with
  `cobra.ExactArgs(3)`. The command passes all three argument values unchanged
  to the writer, preserving profile identity, key case, whitespace, and value
  semantics.
- Rich uses eyebrow `YCY / config cm set`, title `Update CM profile`, and
  subtitle `Change one stored provider setting`. Safe profile and key
  projections appear only in a bounded detail area so long or hostile
  arguments cannot resize or corrupt the fixed title and rail.
- The Live View shows one spinner with Active detail
  `Validating setting and saving profile`. Before writer success it shows only
  safe profile and key, never value. It does not show existing profile data,
  old values, configuration paths, timing, or percentages.

### Atomic Work Phase

The immutable catalog contains exactly:

| ID | Name |
| --- | --- |
| `update-cm-profile` | `Update CM profile` |

The Experience opens before store construction and the phase becomes Active
immediately. It covers store construction and exactly one
`SetCMProfileValue(profile, key, value)` call, then becomes Completed or
Failed. Rich always renders the spinner and Plain always writes
`Updating CM profile...` to stderr, even for a fast operation; neither sleeps
to prolong it. Automation suppresses loading, Transcript, styling, and
terminal controls.

The command performs no prevalidation. Appconfig retains the complete current
contract and ordering: first verify the profile, then dispatch the exact key,
then validate/normalize its value, and publish only on success. The supported
keys remain `baseURL`, `model`, `apiKey`, `temperature`, `timeoutMs`, and
`maxOutputTokens`.

Value semantics remain unchanged: Base URLs are trimmed and lose trailing
slashes without URL validation; models are trimmed and may become empty; API
keys are encrypted from the exact input and may be empty; temperature keeps
its finite JavaScript-Number-compatible parsing and `[0,2]` range; timeout and
maximum tokens keep their integer-prefix parsing and respective `>=1000` and
`>=32` constraints. Empty/BOM/hexadecimal/binary/octal/fractional-prefix
compatibility is not tightened. Setting the same effective value still calls
the writer, publishes, and succeeds.

### Safe request projection

Profile and key use bounded, single-line, control-free projections in Rich,
Transcript, and diagnostic context; unsafe values fall back to `Profile` and
`Setting`. These projections never replace the exact writer identity.

No separate Milestone is emitted because the shared Milestone contract would
also add a Plain line. Instead, a successful phase's final detail carries the
safe request:

- `baseURL`: `Base URL: <safe URL>`, with userinfo/query/fragment removed;
  unsafe values use `Base URL configured`, and a requested empty effective
  value uses `Base URL: <empty>`.
- `model`: `Model: <safe model>`; unsafe values use `Model configured`, and a
  requested empty effective value uses `Model: <empty>`.
- `apiKey`: always `API key: [redacted]`, including an empty input. No length,
  prefix, checksum, strength, plaintext, or ciphertext is shown.
- Numeric keys: `Requested value: <safe raw input>`. The command does not
  parse or normalize it for presentation; an empty request is
  `Requested value: <empty>`.

The value appears only after the writer succeeds. Every failed request omits
the value from Live View final detail and Transcript, including missing
profile, unsupported key, invalid numeric input, encryption failure, and
storage failure. This also protects a secret accidentally supplied under a
non-sensitive key.

The existing positional `apiKey` value may still be visible to shell history
or process-argument inspection. This ticket preserves that established
noninteractive interface and guarantees only that ycy's own Live View,
Transcript, results, and diagnostics never echo it. Hidden credential input is
a separate future security/compatibility decision; this command adds no
warning or prompt.

### Transcript and streams

Successful Rich replay contains the completed `Update CM profile` phase with
safe `Profile`, `Setting`, and key-classified value detail, followed by
`Succeeded`. Failure contains the Failed phase with safe profile/key and
`Unable to update CM profile`, followed by `Failed`. Raw errors, failed values,
stored fields, API keys, ciphertext, paths, spinner frames, timing, and the
complete Command Result never enter the ledger. Rich restores the primary
screen before replay.

Plain emits only `Updating CM profile...` to stderr before the existing stdout
result; it emits no completion line or Transcript replay. Automation emits no
extra UI. The success result remains exactly `Profile <name> updated` in
semantic content and never adds key or value. The shared output boundary strips
terminal controls from hostile names without changing profile lookup.

### Results and failures

The one-shot Finish mapping is:

| Branch | Finish outcome | Command Result |
| --- | --- | --- |
| Any supported-key writer success, including same-value updates | `Succeeded` | `Profile <name> updated` |
| Store, missing profile, unsupported key, invalid value, encryption, or persistence failure | `Failed` | no stdout document |

Missing-profile checks retain precedence over key/value errors and preserve
`CM profile not found: <name>`. Unsupported-key and numeric validation errors
retain their exact current text. Failures publish no partial configuration,
submit no success-looking stdout, and return the original error for root's
single redacted diagnostic; the phase never copies that raw error.

There is no `Cancelled` Command Result. The atomic writer is not interruptible,
so context cancellation does not create a false cancellation or a no-op
cancel callback. The terminal drains the operation and uses the single
writer's real success or failure; it never retries. If persistence succeeds
but `Finish` cannot write stdout, phase and Transcript remain Succeeded and the
output error is returned without rollback, a second Finish, or another writer
call. `Close` is cleanup-only. Rich may fall back to Plain only before
AltScreen or mutation begins; later renderer failures restore the terminal
without repeating the operation.

### Evidence

Acceptance covers all six keys and their legacy normalization/parser cases;
empty values, BOM, hexadecimal/binary/octal input and integer prefixes;
missing/unsupported/invalid/encryption/storage no-publication paths; API-key
absence from Rich, Plain, Automation, Transcript, diagnostics, and persisted
plaintext; Rich fast loading, safe final details, phase states and screen
restoration; Plain stderr/stdout ordering; Automation silence; exact and
hostile profile/key/value inputs; same-value and concurrent updates; context,
renderer and stdout failures; at-most-once writer/Finish; and unchanged
arguments, help, errors, exit codes, and success result.
