# Prove the Vite MPA to Go embed path

Type: prototype
Status: resolved
Blocked by: 03, 06, 07, 08

## Question

Build a disposable thin slice with three Vite HTML entries, at least one shared hashed asset, a worker/file asset, the legacy-compatible caching and route behavior, and a small Go `embed.FS` server that mounts each application at its intended route. Demonstrate clean `make build`, development behavior, missing/stale asset failures, and a test strategy that does not commit `web/dist`. Present only the build and serving interface needed for the port; do not introduce new Diff, FS, or Tunnel security policy.

## Comments

- 2026-08-22, prototype round 1 (awaiting human selection):

  The probe stayed outside the repository because this planning session may modify planning files only. It used Bun 1.3.14, the current compiled Bun CLI, pnpm 11.13.0, Vite 8.2.2, a Go 1.26 language target, `CGO_ENABLED=0`, three physical HTML inputs, a shared JS/CSS module, a module worker, and an imported file asset. The integrated Vite/Go probe ran under the host's Go 1.27.0 compiler; a separate `GOTOOLCHAIN=go1.26.7` follow-up compiled the relevant `embed.FS`/`fs.Sub`/`http.FileServerFS` combination for all six targets, while [Research the pure-Go toolchain and dependency baseline](01-research-pure-go-toolchain.md) remains the evidence for the complete pinned dependency graph. No production file was changed and no compatibility exception was needed.

  **Compiled Bun artifact baseline.** The compiled artifact, rather than source-mode lazy `Bun.HTMLBundle` initialization, is authoritative for first-release tests. All generated non-HTML assets answer `GET` and `HEAD` with `public, max-age=31536000, immutable`, `no-referrer`, `nosniff`, and their real MIME type. All shells answer with the command's existing CSP, `no-store`, `no-referrer`, and `nosniff`. Focused probes exposed method and fallback behavior that the Go adapters must preserve rather than normalize:

  | Command | Shell and missing-asset behavior proven from the compiled Bun artifact |
  | --- | --- |
  | Diff | Exact `/api/*` remains JSON 404 and exact `/mcp` remains the MCP handler. Every other probed route uses the broad shell fallback, including `/mcp/missing`, `/assets/missing.js`, arbitrary deep links, and a non-GET request for a real generated asset. `GET`, `HEAD`, `POST`, and `OPTIONS` fallback probes all returned HTML 200. |
  | FS | `/`, `/browse`, and `/browse/*` return the shell for `GET` and Bun's implicit `HEAD`; `POST` and `OPTIONS` return the existing JSON 405. A real generated asset serves only for `GET`/`HEAD`; a missing `/assets/*`, another path, or a non-GET asset request reaches the existing JSON fallback. |
  | Tunnel Server | `/`, `/clients`, `/clients/*`, `/server`, and `/accounts` return the shell for `GET` and Bun's implicit `HEAD`. `POST`/`OPTIONS` on those paths, missing `/assets/*`, and other paths reach the existing JSON 404. A real generated asset serves only for `GET`/`HEAD`. |

  Source mode lazily creates `Bun.HTMLBundle.files`, so a direct source-run probe did not initially apply the configured bundle headers. The compiled artifact did. This is a Bun development implementation detail, not permission to weaken the shipped Go artifact headers. Command migration tests must keep the compiled-artifact matrix above, including Bun's implicit `HEAD` and Diff's broad non-GET fallback, instead of inheriting `net/http` defaults.

  **Vite and embed result.** One Vite MPA with top-level `input`, `base: '/'`, `appType: 'mpa'`, `publicDir: false`, `assetsDir: 'assets'`, and `assetsInlineLimit: 0` emitted the fixed shells `dist/fs/index.html`, `dist/diff/index.html`, and `dist/tunnel-server/index.html`. All three referenced the same hashed JS and CSS assets; the worker and imported file were separately hashed. A structured HTML verifier failed on a removed file asset and on a shell changed to reference a nonexistent old hash, and passed again after a clean Vite rebuild. `web/dist` remained ignored and absent from the temporary tracker's file set.

  The smallest serving Interface proven by the probe is conceptually:

  ```go
  webassets.Load(app) (*Site, error)
  site.ServeAsset(http.ResponseWriter, *http.Request) bool
  site.ServeShell(http.ResponseWriter, *http.Request, csp string)
  ```

  `webassets` owns one validated output tree, fixed shell selection, exact generated-asset existence, MIME serving, immutable asset headers, and common shell headers. It does not own API/MCP/file/WebSocket namespaces or command-specific shell predicates. Each command's HTTP adapter calls its real protocol handlers first, then `ServeAsset`, then applies its inventory-derived shell/method fallback, so Diff, FS, and Tunnel do not collapse into a falsely uniform router. This is a concrete module, not a public provider Interface or speculative adapter seam; tests may use a private filesystem constructor if needed.

  **Build, development, and test result.** With no `web/dist`, untagged `CGO_ENABLED=0 go test ./...` passed through a typed `ErrWebAssetsUnavailable` stub, while `go build -tags embedded_web` failed immediately with `pattern dist: no matching files found`. `make build` then performed frozen pnpm install, Vite build, asset-graph verification, and `CGO_ENABLED=0 go build -tags embedded_web`; `make test` used the same web prerequisite before tagged Go integration tests. The integrated probe cross-built for all six required targets with the host Go 1.27.0 compiler, and the exact Go 1.26.7 follow-up independently cross-built the standard-library embed/serve path for those targets. Formal acceptance must run the complete tagged suite and cross-build with `GOTOOLCHAIN=go1.26.7`; it may not infer pinned-toolchain success from either partial result alone. Three app-selected commands of the exact form `pnpm --dir web run dev --mode <fs|diff|tunnel-server>` used strict, distinct ports, HMR shell rewrites, and explicit backend proxies; reserved `/api`, `/files`, `/thumbnails`, and `/mcp` requests reached the Go backend, module/asset misses stayed 404, and a second process on the selected port failed instead of drifting.

  The execution test layers should therefore be: clean-checkout untagged Go unit tests against the typed stub; structured Vite output verification; tagged handler tests reproducing the compiled Bun route/method/header matrix; production-browser tests for all three retained React applications and their real workers; six-target cross-build checks; and final native artifact tests. The active suite never invokes legacy Bun and does not keep a separate golden fixture corpus; each command derives its matrix from `legacy/bun/` when migrated.

  **Selection requested.** Recommend the paired build-constraint design: `embedded_web` selects the real adjacent `go:embed dist` implementation, while the untagged implementation returns `ErrWebAssetsUnavailable` before a web command starts. Only `make build` produces a releasable binary. This preserves the already-selected independent `check-go` (`go test ./...`) and `check-web` lifecycle while keeping Vite first in every product build. The alternative is one unconditional embed implementation, under which every clean-checkout `go build` and `go test` must first run the frontend build; it avoids a backend-only developer binary but couples all Go checks to pnpm and ordering. Committing `web/dist`, generating committed Go literals, and shipping external assets remain rejected by the map's existing constraints.

- 2026-08-22, human selection: chose unconditional embedding because the product objective is one binary that works out of the box, matching the compiled Bun artifact's unconditional payload. The paired `embedded_web`/stub design is rejected. The updated [interactive preview](../prototypes/vite-go-embed-preview.html) defaults to the selected lifecycle and retains the routing-parity simulator.

## Answer

Use one unconditional Go embedding implementation. Every ycy binary contains the complete, verified Vite output; there is no `embedded_web` build tag, no `ErrWebAssetsUnavailable` stub, and no backend-only ycy binary. This makes “a built ycy binary is self-contained and its Diff, FS, and Tunnel frontends are present” a compile-time/product invariant rather than a release-command convention.

Keep `web/dist` ignored. The supported clean-checkout lifecycle is frontend-first: frozen pnpm installation when dependencies are not already bootstrapped, Vite MPA build, structured asset-graph verification, then `CGO_ENABLED=0` Go compilation. `make build`, `make test`, and the Complete Gate must encode that dependency. A raw `go build ./cmd/ycy`, `go vet ./...`, or `go test ./...` before `web/dist` exists is intentionally unsupported and must fail at the `go:embed` pattern; after a current output tree has been generated and verified, those Go commands may run against that tree. No placeholder asset, committed `web/dist`, generated Go literal, external runtime asset directory, or alternate non-Web binary may weaken this guarantee.

The embedded layout remains one Vite MPA with fixed shells `dist/fs/index.html`, `dist/diff/index.html`, and `dist/tunnel-server/index.html`, plus one shared hashed `dist/assets/` tree containing Vite-owned worker and file assets. Before Go compilation, a structured verifier must prove that every shell reference resolves, required app shells exist, referenced worker/file outputs exist, stale hashes are absent, and generated assets obey the expected layout. Missing or stale output fails before a binary is emitted.

The serving module owns the validated output tree, fixed shell selection, exact asset existence and MIME handling, immutable generated-asset headers, and common no-store shell headers. Its implementation-facing surface is the proven concrete shape `webassets.Load(app) (*Site, error)`, `site.ServeAsset(...) bool`, and `site.ServeShell(...)`; command adapters continue to own API, MCP, file, WebSocket, method, and fallback routing. Acceptance tests must reproduce the compiled Bun artifact matrix recorded above, including Diff's broad HTML fallback, FS's JSON 405/404 paths, Tunnel's JSON 404 paths, implicit `HEAD`, exact MIME types, immutable asset caching, and no-store shells.

Development keeps the three explicit Vite modes with strict ports, HMR shell rewriting, and reserved-path proxies to a Go backend. Production acceptance runs the structured Vite verification, tagged-free embedded handler tests, real-browser tests for all three React applications and workers, the complete test suite, and all six `CGO_ENABLED=0` cross-builds under pinned Go 1.26.7. The probe found no Bun/Node behavior that requires a compatibility exception.
