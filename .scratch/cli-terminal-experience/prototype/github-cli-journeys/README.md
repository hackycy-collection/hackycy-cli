# ycy GitHub CLI-inspired journeys

This is a throwaway terminal prototype for `Prototype the terminal visual
language`. It is not production ycy code and does not change command syntax,
stream ownership, or Automation Input rules.

Run a representative command journey directly in a terminal:

```sh
go run . --journey git-cm
go run . --journey git-pulse
go run . --journey tunnel-connect
go run . --journey tunnel-server
go run . --journey config-cm-add
go run . --journey automation
```

Use `--width 60` to inspect the narrow layout. `NO_COLOR=1` removes color but
keeps wording and hierarchy.

The following commands run real Huh prompt demonstrations using the same
visual language:

```sh
go run . --form git-cm
go run . --form git-pulse
go run . --form tunnel-connect
go run . --form config-cm-add
```

The static renderings deliberately resemble the small, command-local terminal
surfaces used by GitHub CLI: no global banner, no full-screen dashboard, and
color only where it communicates state.
