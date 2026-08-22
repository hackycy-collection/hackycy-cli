# ycy

`ycy` is a CGO-free Go command-line application. The current migration build identifies itself as `0.0.0-dev`; command leaves are added only after their focused compatibility work is complete.

## Development

```sh
make bootstrap
make hooks-install
make check
make build
```

`make build` builds the Vite applications first and embeds all three generated web shells in the standalone binary. See [CONTRIBUTING.md](CONTRIBUTING.md) for prerequisites, hook lifecycle, verification, and platform notes.
