# ycy terminal visual-language prototype

This is a throwaway visual prototype for
`Prototype the terminal visual language`. It is intentionally isolated from
the production `ycy` module and does not define implementation behavior.

Run it from this directory:

```sh
go run .
```

Use Left/Right (or `h`/`l`) to switch between the three visual directions and
`q` to leave the gallery. `go run . --static --variant signal` prints a
non-interactive snapshot. `go run . --form selection`, `--form confirm`, and
`--form secret` open small Huh demonstrations for the selected direction.

The Automation Session blocks are deliberately plain. Their command-specific
input spelling remains outside this prototype and will be decided separately.
