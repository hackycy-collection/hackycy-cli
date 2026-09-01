# Choose the zip terminal experience

Type: grilling
Status: resolved
Blocked by: 03, 05, 06

## Question

What exact planning choice, glob multiselect, output-name form, archive Work
Phases, planning notes, redacted Interaction Transcript, cancellation/failure
treatment, and final archive Command Result should `zip` use without changing
selection defaults, glob semantics, or filesystem behavior?

## Answer

`zip` is a finite archive command. It uses a vivid Rich Live View in AltScreen
for its bounded planning interactions, then presents truthful archive phases
and one final Command Result. Existing discovery heuristics, selection order,
glob semantics, filename sanitization, archive bytes, filesystem effects,
reveal behavior, ResultKind mapping, and exit contracts remain unchanged.

### Archive Planning

The command keeps its existing conditional order:

1. select a workspace package when multiple packages are discovered;
2. select a source directory;
3. select file patterns;
4. edit the output basename; and
5. show the completed plan before archive work begins.

Rich uses one Bubble Tea root with Huh v2 Select, MultiSelect, and Input
controls. Package and source options preserve current order, recommendation,
project type, confidence, safe relative labels, and bounded hints. Plain keeps
the numbered prompt, default first option, invalid-selection retry, and
control-free output. Automation never reads stdin: it returns the existing
`zip requires an interactive terminal` error before scanning, collection,
compression, reveal, or file writes.

The planning notes show workspace/package counts, project kind, candidate
count, recommendation confidence, and the final safe plan. `git remote` is
used only to derive the existing default name; lookup failure silently falls
back to package name or directory basename. No remote URL, hidden traversal,
absolute path, or raw discovery error is displayed. Safe labels use relative
paths, visible control-character escapes, and bounded wrapping; the original
absolute source path remains the only value passed to archive work.

### Pattern and output interactions

The fixed glob choices and their order remain `**/*`, `**/*.html`, `**/*.js`,
`**/*.css`, `assets/**/*`, and `static/**/*`. Rich renders them as a searchable
Huh v2 MultiSelect with `**/*` selected by default. Plain preserves numbered
comma/space input, `all`, `none`, and invalid retry. Selecting All or submitting
an empty set normalizes to the unique default `**/*`; custom selections retain
their order. Hidden files/directories, VCS directories, symlinks, non-regular
files, and the output archive itself retain their existing collector
exclusions.

The output basename uses Huh v2 Input with the current remote/package/directory
default. Empty input accepts that default, then `SanitizeFileName` is applied
and one `.zip` suffix is appended. `--with-dir` is passed unchanged to the
builder; the plan only presents a safe summary of whether a top-level prefix
is enabled. `--without-open` only disables the final platform revealer and
does not add confirmation.

### Work Phases

Every reached phase enters Active and a terminal state even when it completes
immediately; no artificial delay, file-level spinner, percentage, or ETA is
invented. The command-owned phase catalog is:

- `discover-workspace`: workspace/package/project candidate discovery;
- `select-source`: source selection milestone;
- `select-patterns`: glob selection milestone;
- `prepare-archive`: output name, archive destination, and `with-dir` plan;
- `collect-files`: filesystem walk, glob expansion, and regular-file filtering,
  finishing with `collectedCount`;
- `compress-files`: complete in-memory ZIP construction, finishing with
  `includedCount`;
- `write-archive`: one-time direct publication of the ZIP file; and
- `reveal-archive`: optional host reveal after publication.

The builder remains a complete in-memory operation, so progress is represented
by phase state and final counts rather than fabricated percentages. The output
zip is never claimed to exist until `WriteZipFile` succeeds.

### Outcomes, cancellation, and failures

Cancellation in any planning interaction yields `ResultCancelled`, records the
safe cancellation location, and performs no later collection, compression,
write, or reveal. Context cancellation or a signal observed before archive
work stops later phases without rewriting the outcome as a user form decline;
existing signal/exit behavior remains authoritative.

Directory-not-found, path-not-directory, no-files, no-valid-files,
collection-failed, compression-failed, and write-failed retain their existing
`ResultKind` and normal/failed status semantics. Rich and Transcript use only
bounded safe paths, counts, and failure classes (`collection`, `compression`,
`write`, `directory`, `path`, `no-files`, or `no-valid-files`). Raw filesystem
errors stay on the existing diagnostic path and are not copied into the
Transcript. Empty collection or post-filter compression never creates or
claims an incomplete archive.

Successful publication returns `ResultCompleted` with collected/included
counts, safe output basename, and `RevealFailed` only when an enabled host
reveal fails. A reveal failure is a warning, not an archive failure. With
`Open=false`, no revealer call is made. Stdout emits the existing `Done!` or
other result exactly once; no large file list or duplicate Transcript is sent
to stdout.

### Transcript and stream projection

Rich freezes the semantic ledger, restores the primary screen, and replays a
compact redacted Transcript to stderr containing completed choices, safe plan,
phase final states, counts, cancellation/failure location, and final outcome.
It never replays keystrokes, invalid answers, search activity, animation
frames, complete file lists, absolute paths, or archive bytes. Plain emits
line-oriented status and the existing durable result without a second replay.
Automation emits no phases, Transcript, styles, or terminal controls.

Text presentation uses vivid status symbols and color only on capable TTYs;
redirected and Plain output remove ANSI while retaining readable state words
and symbols. Long values wrap or use bounded labels. `with-dir`, source paths,
package names, and output names are projected independently from the original
mutation values so display safety cannot alter archive semantics.

### Redaction and acceptance evidence

No planning form, note, Transcript, diagnostic projection, or result summary
may expose git remote URLs, absolute paths, hidden traversal paths, raw control
characters, complete file lists, archive bytes, or raw filesystem errors.

Acceptance covers workspace/package/source recommendation and ordering;
project detection and remote-name fallback; all four Huh v2 forms; Plain
numbered/`all`/`none` grammar and invalid retries; cancellation at every
planning step; Automation's no-stdin/no-side-effect boundary; glob
normalization and hidden/VCS/symlink/non-regular/output exclusions;
`with-dir` and filename sanitization; collection/compression/write/reveal
phase order and counts; fast completion without fabricated progress;
every ResultKind and reveal-warning mapping; Context/SIGINT behavior; Rich
AltScreen teardown and bounded Transcript; Text/Plain ANSI degradation;
exactly-once stdout result; archive bytes/entries/permissions/path;
cross-platform reveal commands; redaction of paths, remotes, controls, and
raw errors; and idempotent Close/Result behavior.
