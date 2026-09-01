# Choose the Charm v2 and logging boundaries

Type: grilling
Status: resolved
Blocked by: 01

## Question

Which latest Charm v2 modules should own forms, lists, styling, progress, and
logs; should `charm.land/log/v2` replace or sit behind the current logging
Runtime; and what repository interfaces preserve redaction, injected streams,
level/format configuration, NDJSON stability, renderer leases, and command
ownership without creating parallel UI stacks?

## Answer

Pin Bubble Tea `v2.0.9`, Bubbles `v2.2.1`, Huh `v2.0.3`, Lip Gloss
`v2.0.6`, and `charm.land/log/v2` `v2.0.0`. The delivered module graph must
contain only the Charm v2 module paths: do not retain v1 compatibility
wrappers, parallel form implementations, runtime feature flags, or direct
Charm imports in command packages.

`internal/terminal` owns the terminal Experience and is the only Module that
may compose Bubble Tea, Bubbles, Huh, Lip Gloss, and color-profile handling.
Its semantic Interface remains the command-facing seam. Commands continue to
own their information hierarchy, wording, Interaction Requests, Work Phases,
and Command Results without knowing which Charm implementation renders them.
An architecture test must enforce these import rules.

Inside that Module:

- Bubble Tea owns the one root event loop and terminal modes for an
  `ExperienceRun`. A run enters and leaves AltScreen at most once.
- Huh owns all text, secret, select, multi-select, and confirm forms as child
  models of that root. Do not invoke standalone `Form.Run`.
- Replace `richListForm` with Huh v2 Select and MultiSelect. Preserve
  `InteractionOption.Label`, `Description`, and `Value` separately; the Huh
  Adapter projects label and description into visible, searchable content but
  returns the original value. Transcript rendering consumes the semantic
  fields rather than parsing Huh output. Retain a private custom field only if
  the visual prototype and largest-list PTY evidence prove Huh insufficient.
- Bubbles supplies spinner, progress, viewport, and similar rendering
  primitives where a command-specific view needs them. It is not a second
  form or orchestration layer.
- Lip Gloss owns pure styles. It must not use a package-global writer or
  perform implicit output routing.
- `internal/terminal` owns semantic state, the Interaction Transcript ledger,
  injected streams, renderer leases, and capability policy.

Keep the repository-owned `StreamCapability{Terminal, Color}` model rather
than exposing Charm's color-profile type through the semantic Interface. The
terminal implementation receives explicit streams and an environment snapshot,
constructs a private color-profile writer for stdout and stderr, and gives
Bubble Tea the same environment/profile. `Color=false`, Plain Interactive, and
Automation force control-free output; otherwise Charm performs normal ANSI,
256-color, or TrueColor downsampling. This is the first-version compatibility
strategy instead of a repository-maintained terminal matrix.

A Rich startup failure may transparently fall back to Plain Interactive only
before AltScreen begins and before an answer or other semantic state is
committed. Once the Rich session has started, a renderer failure restores the
terminal, replays the safe partial transcript, and returns the original error;
it must not repeat questions or work. Service Commands do not hold a Live View
while running: commands without a form never start Bubble Tea, while a command
such as `tunnel connect` completes its selection, restores the terminal,
replays the selection transcript, and releases the renderer before beginning
its Lifecycle Log.

The Rich completion sequence is fixed:

1. Freeze the semantic transcript ledger.
2. Stop Bubble Tea and wait for the primary screen and terminal modes to be
   restored.
3. While the renderer lease still owns stderr, render the Interaction
   Transcript through it.
4. Release the lease and flush deferred Diagnostic Records in FIFO order.
5. Emit the complete Command Result once on stdout, when one exists.
6. Only then hand the terminal to a child process such as `run`.

`internal/logging` remains the deep logging Module. Callers continue to use
`*logging.Runtime` and value `logging.Logger`; do not add a public interface
for a single implementation. Runtime continues to own filtering, injected
clock sampling, scope and context merging, recursive redaction, dynamic
level/format configuration, output serialization, writer selection, and the
exact JSON/NDJSON projection. Existing loggers must continue to observe later
Runtime configuration changes.

Log v2 is a private, text-only Lifecycle Log Adapter behind that Interface.
It receives a complete normalized and redacted record and may only project it
to human-readable text. It must not sample time, refilter levels, cache scoped
Log child loggers, select a format, write JSON/logfmt, call `Fatal`, use the
package-global logger, or receive unredacted values. It first renders one
complete record into a private buffer; Runtime then performs one serialized
write to the lease-aware diagnostic writer so records remain atomic and
deferred records retain FIFO order. Commands never import Log v2 directly.

Human-readable text logs may change their exact bytes to gain vivid but
professional symbols, color, hierarchy, and spacing. They must remain one
record per line, retain timestamp, level, scope, message, and structured
context, and remain ANSI-free when redirected or color is disabled. The
existing JSON/NDJSON schema, timestamp behavior, record boundaries, redaction,
and configuration behavior remain exact compatibility contracts. Supporting
evidence and Log v2 limitations are recorded in
[`Latest compatible Charm v2 stack`](../research/01-latest-charm-v2-stack.md).
