# Exemplary open-source CLI terminal experiences

Date: 2026-08-31

## Question

Which maintained open-source CLIs provide the strongest applicable patterns
for vivid forms, finite-work progress, semantic completion transcripts,
failure presentation, and long-running lifecycle logs, and which concrete
patterns should ycy adopt or reject without copying product-specific behavior?

## Method and scope

This report uses only first-party documentation and source code. The source
links are pinned to current default-branch commits so the findings remain
reproducible. All five projects had recent default-branch activity at the time
of research:

- [Gum at `4d089f9`](https://github.com/charmbracelet/gum/tree/4d089f95507708a71f64dacfe7ca513219dd5267)
- [GitHub CLI at `40b742f`](https://github.com/cli/cli/tree/40b742f76d68e6b1f472942a6368db4b5d765641)
- [Dagger at `6938a67`](https://github.com/dagger/dagger/tree/6938a67509861b0f231553cb4886b7fb6643975d)
- [Pulumi at `793f7b2`](https://github.com/pulumi/pulumi/tree/793f7b2e160db4321fb7fb6b0607461e01cb251e)
- [Docker Compose at `727fa0e`](https://github.com/docker/compose/tree/727fa0e67d952f93757fa5d86d2ff91d96f0ad1b)

The goal is not to rank the products as a whole. It is to identify small,
portable presentation decisions that fit ycy's Go and latest-Charm-v2
direction and its existing Terminal Experience vocabulary.

## Recommendation

Use a composite reference rather than copy one CLI:

| ycy concern | Strongest reference | Pattern to carry forward |
| --- | --- | --- |
| Vivid forms | Gum, with GitHub CLI as the in-product Huh v2 example | Strong focus/selection states and a typed form adapter; transient rendering is separate from the durable result |
| Finite-work progress | Dagger and Pulumi | Drive both live progress and final text from semantic state, not captured terminal frames |
| Interaction Transcript | Dagger | Stop the live renderer, switch to a final projection, write it to stderr, and keep the command's stdout result independent |
| Failure presentation | Dagger and Pulumi | Attribute one root cause, retain successful prior phases, avoid duplicate errors, preserve the original exit code |
| Lifecycle Log | Docker Compose, rendered through Charm Log v2 | Stable line prefixes/fields, meaningful lifecycle events, stdout/stderr fidelity, and an unchanged structured mode |

Dagger is the closest architectural exemplar for finite commands. Its code
literally distinguishes transient progress from "durable terminal output" and
performs a separate final render after the TUI has exited. Gum is the better
visual reference, but it is a component CLI rather than a model for a complete
product workflow. Pulumi and Compose are strongest where ycy has high-volume
progress and service logs respectively.

## Evidence and decisions

### 1. Vivid forms: Gum for visual primitives, GitHub CLI for product integration

Gum v2 gives selection controls a conspicuous default cursor, selected marker,
and bright accent colors; it also keeps header, cursor, item, and selected-item
styles independently configurable
([options](https://github.com/charmbracelet/gum/blob/4d089f95507708a71f64dacfe7ca513219dd5267/choose/options.go#L9-L34)).
The program renders its Bubble Tea interaction to stderr and emits only the
submitted value after the program completes
([choose](https://github.com/charmbracelet/gum/blob/4d089f95507708a71f64dacfe7ca513219dd5267/choose/command.go#L144-L173),
[input](https://github.com/charmbracelet/gum/blob/4d089f95507708a71f64dacfe7ca513219dd5267/input/command.go#L62-L80)).
Password input changes the echo mode instead of inventing a separate prompt
path
([password input](https://github.com/charmbracelet/gum/blob/4d089f95507708a71f64dacfe7ca513219dd5267/input/command.go#L26-L48)).

GitHub CLI is a useful current proof that Huh v2 can sit behind a mature Go CLI
prompter interface. Its implementation injects input and output, applies a
theme in one constructor, translates Huh cancellation into the CLI's existing
cancellation contract, and exposes typed select, multiselect, input, password,
confirm, and destructive-confirmation flows
([prompter construction and cancellation](https://github.com/cli/cli/blob/40b742f76d68e6b1f472942a6368db4b5d765641/internal/prompter/huh_prompter.go#L15-L39),
[field implementations](https://github.com/cli/cli/blob/40b742f76d68e6b1f472942a6368db4b5d765641/internal/prompter/huh_prompter.go#L41-L226)).

Adopt:

- Keep one ycy form adapter over Huh v2, including input/output injection,
  cancellation translation, validation, and redaction metadata.
- Give focused questions, selectors, checked items, validation failures, and
  completion states visibly different treatments. Gum's independent style
  slots are the right granularity; exact colors and glyphs belong to the visual
  prototype ticket.
- Keep the Live View on the interaction stream and emit the Command Result
  separately. The ycy-specific extension is to build and replay the semantic
  Interaction Transcript on stderr after leaving the alternate screen.
- Treat password/token fields as normal typed form fields with a secret flag.
  Their Transcript answer is a fixed redaction such as `provided`, never the
  entered text or its length.

Reject:

- Do not invoke Gum as a subprocess. Its source is a reference for composition
  and styling; ycy should use Huh/Bubble Tea/Lip Gloss v2 directly.
- Do not use Gum's submitted stdout value as ycy's transcript. A shell helper
  returns one value; ycy needs a multi-step semantic record with redaction and
  final phase states.
- Do not copy GitHub CLI's conservative Base16 appearance. It proves the Huh v2
  integration boundary, but the confirmed ycy direction is more vivid.
- Do not expose raw key presses, filter queries, rejected secret input, or help
  bars in the Transcript. They are Live View state, not completed decisions.

### 2. Finite-work progress: Dagger's state projection plus Pulumi's event model

Dagger defines a command view whose `SetFinal` method explicitly switches from
"transient progress" to "durable terminal output"
([interface](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend.go#L128-L144)).
Its final setup changes rendering mode, removes focus, and rebuilds from stored
state rather than scraping the last screen
([final-state setup](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_pretty.go#L1703-L1734)).
The Dagger renderer also distinguishes live facts from final facts: for example,
"waiting on" is meaningful only while work is active and is omitted from the
final render
([duration rendering](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend.go#L787-L820)).

Pulumi models progress as events and resource rows with explicit running,
failed, and done state, while choosing separate interactive and
non-interactive renderers over the same display data
([display model](https://github.com/pulumi/pulumi/blob/793f7b2e160db4321fb7fb6b0607461e01cb251e/pkg/backend/display/progress.go#L47-L179),
[renderer selection](https://github.com/pulumi/pulumi/blob/793f7b2e160db4321fb7fb6b0607461e01cb251e/pkg/backend/display/progress.go#L219-L264)).
It retains otherwise-hidden unchanged work as a count so a large operation
does not look stalled
([unchanged-resource rationale](https://github.com/pulumi/pulumi/blob/793f7b2e160db4321fb7fb6b0607461e01cb251e/pkg/backend/display/progress.go#L117-L124)).

Adopt:

- Represent each command's finite work as command-owned Work Phases with stable
  semantic identity and status: active, succeeded, failed, cancelled, or
  skipped. The shared layer owns rendering mechanics, not the phase list.
- Render the Live View and Interaction Transcript as two projections of the
  same state. The transcript is not a capture of the final frame.
- While active, show a spinner, phase label, and elapsed time for work without
  a measurable total. For measurable work, show real current/total data.
- On completion, replace animation-only facts with a terminal status and useful
  measurements. Preserve meaningful aggregate counts when detailed rows are
  intentionally collapsed.
- Show every real ycy Work Phase even if it completes quickly, as already
  decided by the map. Dagger initializes a `TooFastThreshold` for its own
  workload
  ([run defaults](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_pretty.go#L1153-L1162));
  that product-specific filtering rule should not override ycy's decision.

Reject:

- Do not force all commands into Dagger's trace tree or Pulumi's resource
  table. Their information architecture matches dependency graphs, not ycy's
  heterogeneous leaf commands.
- Do not turn arbitrary log lines into phases. A Work Phase is a finite state
  transition that the command owns and can conclude.
- Do not show fabricated percentages for work whose total is unknown.
- Do not use periodic heartbeats in a normal interactive finite-work Live View.
  They are a non-interactive liveness technique, not user progress. Dagger, for
  example, reserves a 30-second heartbeat for report-only consumers
  ([report heartbeat](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_pretty.go#L1202-L1247)).

### 3. Semantic completion transcripts: Dagger is the direct precedent

Dagger's `Run` contract starts the TUI, stops it, and then prints primary output
to the appropriate streams. Its final report normally goes to stderr expressly
so redirected stdout remains the command's result
([run and stream policy](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_pretty.go#L1153-L1200)).
`FinalRender` is called only after the TUI has exited, suppresses interactive
key hints, and renders the rebuilt final view
([final rendering](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_pretty.go#L2118-L2159)).

Its output replay also preserves stream identity. On failed report-mode runs it
can replay stdout while omitting stderr already represented by the rendered
failure, explicitly preventing duplicate failure output
([primary-output replay](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend.go#L1059-L1118)).

Adopt:

- Give the Terminal Experience an explicit end sequence: freeze semantic
  state, stop and restore the alternate screen, render the Interaction
  Transcript to stderr, then emit the Command Result once to stdout.
- Use a dedicated transcript projection that includes completed questions and
  redacted answers, meaningful final Work Phase states, failure/cancellation
  location, and the final outcome. It must omit cursor/focus state, key hints,
  spinner frames, and transient "waiting" text.
- Track whether a fact has already been owned by the Transcript, Command
  Result, or Diagnostic Record. This prevents accidental replay without
  relying on string comparison.
- Preserve ycy's special `run` handoff: replay its choice summary, end the
  Terminal Experience, and only then give the terminal to the child process.

Reject:

- Do not print the last alternate-screen frame verbatim. It contains transient
  layout, may be clipped to viewport height, and cannot correctly redact or
  deduplicate information.
- Do not duplicate large Command Results in the Transcript. Dagger's explicit
  report/result stream split demonstrates the needed ownership boundary.
- Do not make the transcript an unbounded debug trace. Diagnostic detail has a
  separate role and may be increased by log level without changing the
  transcript.

### 4. Failure presentation: one attributed cause, then durable context

Dagger checks whether the root cause is already visible; if so, it skips the
poorer duplicate error representation. It still replays pre-failure stdout when
appropriate and preserves a child process's original exit code rather than
flattening failures to `1`
([failure deduplication and exit codes](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_pretty.go#L2161-L2185)).
Its report policy separately decides whether to show subtests, descendant logs,
root cause, and next-step suggestions, with comments explaining that a cause
must stay attached to the failing operation and must not appear in a passing
report
([failure render policy](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_trace_policy.go#L5-L35),
[root-cause guards](https://github.com/dagger/dagger/blob/6938a67509861b0f231553cb4886b7fb6643975d/dagql/idtui/frontend_trace_policy.go#L88-L138)).

Pulumi's end processing fixes a useful semantic order: finish progress, print
resource changes, policies, diagnostics, outputs, and finally the summary
([end sequence](https://github.com/pulumi/pulumi/blob/793f7b2e160db4321fb7fb6b0607461e01cb251e/pkg/backend/display/progress.go#L668-L725)).
Diagnostics are retained during work but printed together at the end, grouped
by the resource that owns them
([diagnostic retention](https://github.com/pulumi/pulumi/blob/793f7b2e160db4321fb7fb6b0607461e01cb251e/pkg/backend/display/progress.go#L47-L67),
[grouped final diagnostics](https://github.com/pulumi/pulumi/blob/793f7b2e160db4321fb7fb6b0607461e01cb251e/pkg/backend/display/progress.go#L779-L815)).

Adopt:

- After restoring the terminal, show prior completed phases, mark the failed or
  cancelled phase in place, then print one actionable failure summary. Emit
  additional Diagnostic Records afterward, once.
- Attach an error to the Work Phase or command action that failed. Do not lead
  with an unscoped `error:` line when the user needs to know which step failed.
- Separate cause from consequence. A cleanup failure, partial result, or child
  exit code can be retained without replacing the primary cause.
- Preserve current ycy exit codes and cancellation semantics exactly.
- Offer a next action only when the command can name a concrete recovery step;
  otherwise keep the failure terse and let debug diagnostics carry detail.

Reject:

- Do not print Cobra's error, a presenter's error, a logger error, and the same
  wrapped error again. One layer owns the human summary.
- Do not hide all successful phases on failure. They explain partial side
  effects and make retries safer.
- Do not append the full diagnostic history by default when a short root cause
  and recovery action are sufficient.

### 5. Long-running lifecycle logs: Docker Compose's source-oriented lines

Docker Compose keeps service output line-oriented. It maintains a presenter per
source, assigns a stable color, recomputes aligned prefixes, prefixes every
line of a multiline message, and preserves stdout versus stderr
([log consumer](https://github.com/docker/compose/blob/727fa0e67d952f93757fa5d86d2ff91d96f0ad1b/cmd/formatter/logs.go#L34-L129),
[prefix layout](https://github.com/docker/compose/blob/727fa0e67d952f93757fa5d86d2ff91d96f0ad1b/cmd/formatter/logs.go#L138-L162)).
Lifecycle changes such as exited, restarting, and recreated are emitted as
status messages through the same consumer instead of opening a dashboard
([container events](https://github.com/docker/compose/blob/727fa0e67d952f93757fa5d86d2ff91d96f0ad1b/pkg/compose/printer.go#L25-L56)).
Users can request follow mode, timestamps, no color, no prefix, and time/tail
filters in the official command surface
([Compose logs reference](https://github.com/docker/compose/blob/727fa0e67d952f93757fa5d86d2ff91d96f0ad1b/docs/reference/compose_logs.md#L1-L24)).

Compose also chooses its finite-operation progress renderer by the terminal
capability of stderr, because that is where progress is written, and keeps
explicit TTY, plain, quiet, and JSON modes
([progress selection](https://github.com/docker/compose/blob/727fa0e67d952f93757fa5d86d2ff91d96f0ad1b/cmd/compose/compose.go#L686-L728)).
That reinforces ycy's existing distinction between presentation/diagnostics on
stderr and Command Result on stdout.

Charm's own Gum v2 log command is direct evidence that `charm.land/log/v2` can
provide the intended Go/Charm renderer: it writes to stderr, styles levels and
key/value fields independently, and selects text, JSON, or logfmt formatters
([Gum log implementation](https://github.com/charmbracelet/gum/blob/4d089f95507708a71f64dacfe7ca513219dd5267/log/command.go#L15-L102)).

Adopt:

- Keep `fs`, `diff`, `tunnel connect`, and `tunnel server` as line-oriented
  Service Commands. Use a stable layout with timestamp when useful, level,
  scope, message, and compact key/value fields.
- Emit meaningful lifecycle transitions at normal level: starting, listening
  or connected, reconnecting, stopping, stopped, and failed. Put request-level,
  heartbeat, polling, and protocol detail behind debug level.
- Use color and a strong status marker to aid scanning, but keep the message
  and level text semantically complete.
- Preserve the existing text stream ownership and the current JSON/NDJSON
  schema. `charm.land/log/v2` may render redesigned text, but adopting its JSON
  formatter is not permission to change ycy's machine contract.
- Ensure every physical line of a multiline event remains attributable to its
  scope, either by repeating a prefix or by deliberately indenting continuation
  lines under one prefixed header.

Reject:

- Do not add an alternate-screen dashboard, spinner, or completion Transcript
  to a Service Command. Its Lifecycle Log is already the durable record.
- Do not copy Compose's rainbow-per-container palette where a ycy command has
  only one or two scopes. Reserve vivid colors for level and state hierarchy.
- Do not log every request or periodic heartbeat at info level. This buries the
  lifecycle transitions the user is waiting for.
- Do not globally force timestamps or decorative prefixes onto Command Results.
  The log renderer applies only to Lifecycle Logs and Diagnostic Records.

## Cross-command contract implied by the examples

The external examples support a small shared mechanism while leaving each
command's information design independent:

1. A form response record contains prompt identity, display label, a redacted
   display answer, and completion/cancellation state.
2. A Work Phase record contains semantic identity, command-owned label, active
   detail, final status, optional real progress counts, duration, and an
   optional failure cause.
3. The Live View renders current records in the alternate screen.
4. The Interaction Transcript renders only durable projections of those same
   records to stderr after the screen is restored.
5. The Command Result is written once to stdout by the command presenter.
6. The failure presenter owns one human summary and preserves the underlying
   exit behavior; the diagnostic logger never repeats that summary by default.
7. Service Commands bypass the finite-work renderer and send typed lifecycle
   events through the line-oriented logger.

This is the reusable boundary. Phase names, ordering, result layout, failure
wording, and which details are worth retaining remain command-specific, as the
map requires.

## Prototype implications

The visual prototype should exercise the findings rather than imitate a
single product:

- Gum-level contrast for focus, selection, validation, and status.
- Huh v2 typed text, password, select, multiselect, and confirm fields behind a
  ycy adapter.
- Dagger-style independent live and final projections, with the final
  projection rendered only after alternate-screen restoration.
- Pulumi-style durable ordering of final phase states, diagnostics, result,
  and summary, adjusted to ycy's stderr/stdout contract.
- Compose-style lifecycle lines for the four Service Commands, including
  multiline messages and stop/failure transitions.

Acceptance evidence should capture both the in-progress PTY frame and the
terminal scrollback after completion. Redirected stdout must contain only the
existing Command Result; automation output must contain no transcript or
terminal control sequences.
