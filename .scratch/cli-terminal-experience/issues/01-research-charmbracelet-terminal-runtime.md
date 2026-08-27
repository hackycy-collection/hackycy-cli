# Research the Charmbracelet terminal runtime contract

Type: research
Status: resolved

## Question

Using first-party Charmbracelet and Go sources, determine the maintained,
CGO-free package graph that ycy can use with Go 1.26.7 for Huh, Bubble Tea,
Lip Gloss, and any required terminal/color dependency. Establish the supported
runtime behavior relevant to this effort: input/output routing, ordinary and
alternate-screen rendering, cursor cleanup, cancellation and signal handling,
non-TTY behavior, `NO_COLOR`/`TERM=dumb` behavior, width detection, Windows
support, and test facilities.

The answer must recommend exact package versions and a minimal integration
constraint set for ycy. It must distinguish facts guaranteed by an upstream
public interface from behavior that ycy must own itself. Capture the report at
`.scratch/cli-terminal-experience/research/01-charmbracelet-terminal-runtime.md`
with source links; do not modify production code.

## Comments

- Claimed for primary-source research during map charting.

## Answer

Primary-source research is complete in
[`01-charmbracelet-terminal-runtime.md`](../research/01-charmbracelet-terminal-runtime.md).

Pin Huh `v1.0.0`, Bubble Tea `v1.3.10`, and Lip Gloss `v1.1.0`; keep the
existing `golang.org/x/term v0.45.0` as ycy's capability API. Add Bubbles
`v1.0.0` only when a complex view imports one of its components. The selected
graph built with Go `1.26.7` and `CGO_ENABLED=0` for Darwin, Linux/amd64, and
Windows/amd64. No direct termenv, go-isatty, Charm x/term, or Charm log
dependency is needed initially.

The central integration decision is that ycy, not Charmbracelet, owns terminal
mode selection. Bubble Tea can open a controlling TTY when stdin is redirected
and can emit redraw escapes to non-TTY output; Automation Sessions therefore
must use a separate deterministic plain path. The report also records required
handling for Huh context cancellation, terminal cleanup, `NO_COLOR`/
`TERM=dumb`, width, Windows, logging during a live TUI, and test seams.
