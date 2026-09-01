# Vivid Terminal Prototype

> THROWAWAY PROTOTYPE. This is visual-decision evidence, not production code.

This PTY prototype compares three structurally different treatments of the
same configure-profile journey:

- `signal`: a persistent left-hand workflow rail;
- `console`: a dense operator-oriented status table;
- `focus`: one centered decision or active phase at a time.

Run it from the repository root:

```sh
make prototype-terminal
```

The default is the Signal Rail success journey. The bottom switcher uses F2
and F3 for the previous/next variant, F4 for success/failure/cancellation, and
F5 to restart. The same state can be launched directly:

```sh
make prototype-terminal PROTOTYPE_ARGS='--variant=console --outcome=failure'
make prototype-terminal PROTOTYPE_ARGS='--variant=focus --outcome=cancel'
```

All values are in memory. The example credential is synthetic and every
transcript renders it as `[redacted]`.
