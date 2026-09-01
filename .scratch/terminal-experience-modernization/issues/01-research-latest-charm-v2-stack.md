# Research the latest compatible Charm v2 stack

Type: research
Status: resolved

## Question

What are the latest stable, mutually compatible versions and relevant APIs of
Bubble Tea v2, Bubbles v2, Huh v2, Lip Gloss v2, and `charm.land/log/v2`; what
migration constraints do their module paths, rendering lifecycle, input/output
injection, themes, accessibility, structured logging, and licensing impose on
this repository's terminal Experience and Diagnostic Record contracts?

## Answer

Pin Bubble Tea `v2.0.9`, Bubbles `v2.2.1`, Huh `v2.0.3`, Lip Gloss `v2.0.6`,
and `charm.land/log/v2` `v2.0.0`. Bubble Tea v2 makes alternate-screen state a
`tea.View` property, Huh v2 supplies the required controls and theme surface,
and Lip Gloss v2 removes writer-bound Renderers. Log v2 may be used only behind
the current logging facade: it has no redaction hook or custom formatter and
its JSON schema is incompatible with ycy's stable NDJSON contract. Full API,
migration, licensing, and evidence details are in
[`01-latest-charm-v2-stack.md`](../research/01-latest-charm-v2-stack.md).
