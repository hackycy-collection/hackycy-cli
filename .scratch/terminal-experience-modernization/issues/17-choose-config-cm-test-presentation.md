# Choose the config cm test terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact profile-resolution and provider-request Work Phases, safe provider
context, response Command Result, timeout/HTTP/decoding failure presentation,
and redacted transcript should `config cm test` use without exposing API keys
or changing request and exit semantics?

## Answer

`config cm test` remains a non-mutating, parameter-driven connection check. It
keeps `config cm test` and `config cm test <profile>`, the existing profile
selection precedence, OpenAI-compatible request body, timeout calculation,
response parsing, error wrapping, stdout documents, and exit codes. Rich uses
one Bubble Tea root and Huh/Signal Rail primitives through the shared Terminal
Experience; Plain and Automation keep their existing stream and scripting
boundaries.

### Entry and Work Phases

- The command accepts zero or one positional profile argument exactly as today.
  With no argument it calls `ResolveCMProfile` using the established selection
  precedence. It never adds profile selection, confirmation, or stdin input.
- Rich Experience starts at the beginning of `RunE`, before store creation or
  profile resolution. Root argument/help/parser failures remain on their
  existing stderr path. Plain and Automation do not enter AltScreen.
- The ordered phase catalog is fixed and every real phase is shown even when it
  completes immediately:

  1. `resolve-cm-test-profile` / `Resolve CM test profile`, covering store
     creation/read and `ResolveCMProfile`.
  2. `test-cm-provider` / `Test CM provider`, covering request construction,
     HTTP exchange, response read, JSON decode, and empty-response detection.

  The second phase is created only after the first succeeds. Each phase moves
  through `Pending -> Active -> Succeeded` or `Failed`; cancellation records
  `Cancelled` at the phase where it occurs. No artificial delay or invented
  percentage is used.
- Stable semantic details are safe fields, not renderer-specific strings. The
  resolve phase may include `Profile: <safe name>` and
  `Source: explicit/default/environment`; the provider phase may include
  `Provider: <safe profile>`, `Base URL: <safe URL>`, and `Model: <safe model>`.
  Dynamic spinner frames are never part of the contract.

Rich uses eyebrow `YCY / config cm test`, title `Test commit message provider`,
and subtitle `Verify the resolved profile can answer a connection check`. Its
active provider detail can say `waiting for response`, but it never displays an
API key, Authorization header, request body, full error, or sensitive
environment variable.

### Request and safe projections

The provider request is unchanged: POST to `<BaseURL>/chat/completions` with
the existing model, temperature, max-token, DeepSeek v4 thinking behavior, and
resolved `TimeoutMS`. The fixed messages remain system `Return exactly: ok` and
user `Connection test.`. A quick request still publishes Active and terminal
phase states.

All user-facing projections first trim, remove ANSI/control characters, and
apply a fixed length bound. Profile and model values are safe, single-line
projections. Base URL is reduced to scheme/host/path and drops userinfo,
query, and fragment; unsafe or overlong values fall back to a generic
`Configured provider` label. Known API keys are replaced with `[REDACTED]`.
No general secret scanner is introduced, and raw response bodies never enter
diagnostic output.

### Success and usage

The durable stdout Command Result is unchanged and is submitted exactly once:

```text
Commit message provider test
Response:
ok
Done
```

Rich shows a bounded, cleaned response in its result area, retaining natural
line breaks and marking an over-limit value as truncated. It may also show a
compact usage summary. Usage fields are displayed only when finite, non-negative
integers are available; provider `total` wins, and total is derived only when
both prompt and completion counts are valid. Missing, partial, NaN, Inf, or
negative values are omitted. Usage never changes stdout.

### Failure, cancellation, and outcome

- Resolver failure marks `Resolve CM test profile` failed, creates no provider
  phase, and makes no network request. The command returns the original resolver
  error; stdout remains empty. Rich/Transcript expose only a safe stable
  subcategory such as `store`, `selection`, or `decrypt`.
- Provider failure keeps the existing dual-stream contract. stdout contains
  only:

  ```text
  Commit message provider test
  Provider request failed
  Provider: <safe profile>
  Base URL: <safe URL>
  Model: <safe model>
  ```

  It does not append `Done` or a partial response. Root diagnostic stderr gets
  the existing provider error after API-key redaction.
- Provider failures use stable internal categories `http-status`, `timeout`,
  `read`, `decode`, and `empty-response`. The user-facing form remains
  `Provider request failed` plus a safe category; HTTP errors include only the
  status code and a cleaned short status text, never response body or headers.
- If context is cancelled before resolution finishes, the outcome is
  `Cancelled` and no request starts. Once a request is sent, a complete response
  or explicit provider error wins over a later cancellation; cancellation wins
  only while no result exists. The implementation never retries or emits a
  second outcome.
- Finish is called once: freeze phases/transcript, restore the primary screen,
  replay the semantic Transcript, then submit the one Command Result. Renderer
  errors cannot replace the real resolver/provider outcome; after AltScreen has
  started, restore the terminal and finish with line-oriented fallback. A
  stdout write failure is returned without rollback or a second request.

### Transcript and other modes

After Rich leaves AltScreen, stderr replays a compact semantic Transcript in
order: safe context title, each meaningful phase final, then the final outcome.
A success records the two phase completions, `Response received`, and an
optional bounded length/usage summary, followed by `Succeeded`; it never copies
the response body. Resolver/provider failures record only the safe failed phase
and `Failed`. Cancellation records the actual cancellation phase and
`Cancelled`. Keystrokes, invalid secret input, spinner frames, raw errors,
credentials, and complete results are excluded.

Plain always emits loading and phase lines to stderr, for example
`Resolving CM test profile...` and `Testing CM provider...`, followed by their
terminal markers. It then writes the unchanged success or failure Command
Result once to stdout; it does not replay the Rich Transcript. Redirected
stdout receives no loading or control sequences.

Automation runs the same command and request without prompts, styling,
Transcript, or terminal controls. Its stdout/stderr, errors, side effects
(none beyond the network check), and exit codes remain exactly compatible with
the current automation contract.

### Evidence

Acceptance uses a fake resolver/store and `httptest` provider with PTY and
redirection coverage. It verifies explicit/default profile resolution, phase
catalog/order and fast-request visibility, safe profile/URL/model projections,
response trimming/control removal/truncation, usage variants, HTTP/timeout/read/
decode/empty failures, resolver failure with zero requests, cancellation races,
Rich primary-screen restoration and Transcript order, Plain stream ordering,
Automation silence, stdout/stderr write failures, redaction, and unchanged
result/error/exit contracts. Resolver, HTTP exchange, decode, Finish,
Transcript replay, and Command Result submission are each asserted
at-most-once.
