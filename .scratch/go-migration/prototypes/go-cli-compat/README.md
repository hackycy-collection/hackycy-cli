# Throwaway Go CLI compatibility probe

This prototype answers one question: can Cobra sit entirely inside a thin CLI adapter while ycy command modules receive only a context and typed input, without losing the difficult Commander-era argv and exit contracts?

It is disposable planning evidence. It is not the start of the production Go tree, and its internal updater name is deliberately not a security design.

Run it with the pinned Go 1.26.7 toolchain:

```sh
GOTOOLCHAIN=go1.26.7 go run . --help
GOTOOLCHAIN=go1.26.7 go run . git cm --push upstream
GOTOOLCHAIN=go1.26.7 go run . diff left right -x 'one,two' -x three
GOTOOLCHAIN=go1.26.7 go run . run ./project -- --watch --port 3000
GOTOOLCHAIN=go1.26.7 go run . --log-level debug run -- --log-level trace
GOTOOLCHAIN=go1.26.7 go run . prompt
GOTOOLCHAIN=go1.26.7 go run . wait
```

The probe prints parsed typed input as JSON. `prompt` needs a terminal and accepts `1`, `2`, or `q`; `wait` runs until interrupted so signal-to-context behavior can be inspected.

Set `YCY_PROTOTYPE_RAW_COBRA=1` to bypass only the optional-value argv normalizer and observe Cobra/pflag's native behavior:

```sh
YCY_PROTOTYPE_RAW_COBRA=1 GOTOOLCHAIN=go1.26.7 go run . git cm --push upstream
```

The isolated updater demonstration is accepted only as the first argv token:

```sh
GOTOOLCHAIN=go1.26.7 go run . __ycy_internal_apply_update --transaction-id demo
```

Authentication, trusted-state binding, and replacement safety remain unresolved by this probe and belong to the dedicated self-update decision.

See [RESULTS.md](RESULTS.md) for the observed compatibility matrix, three candidate Interfaces, recommendation, and decisions still requiring human selection.
