# Research the Vite MPA to Go embedding contract

Type: research
Status: resolved

## Question

Using official Vite, pnpm, and Go documentation plus primary-source examples, what production and development build contracts can support one pnpm/Vite package with `fs`, `diff`, and `tunnel-server` HTML entries, hashed shared assets, workers, Monaco assets, Tailwind, and a CGO-free Go binary that serves each application with the current cache and URL behavior? Compare manifest-driven embedding, fixed output roots, generated Go assets, and other viable approaches. Determine what `make build`, tests, and clean checkouts must guarantee without committing `web/dist`.

## Comments

- Research context: branch `research/vite-go-embedding`; report `.scratch/go-migration/research/vite-go-embedding.md`.

## Answer

Primary-source research is complete at commit `88ee6f5710bc14b5c21f85223b1b2a70c9c1e204` on branch `research/vite-go-embedding`. Use one pnpm/Vite multi-page package with physical inputs `web/fs/index.html`, `web/diff/index.html`, and `web/tunnel-server/index.html`; emit stable shells plus a shared content-hashed `/assets/` tree, embed the complete output once, and serve fixed shells separately from assets. Preserve HTML `no-store`, immutable hashed assets, current security headers and command-specific deep-link routes, with reserved API/file/MCP/WebSocket namespaces taking precedence over shell fallback.

Keep complete Vite-generated HTML rather than adding manifest-driven Go rendering. Replace Bun-specific Pierre worker imports and hard-coded Monaco asset filenames with Vite-owned worker edges, and use the official React and Tailwind Vite plugins. The production order is frozen pnpm install, Vite build plus asset-graph verification, then `CGO_ENABLED=0` Go build. Because ignored `web/dist` cannot satisfy unconditional `go:embed` on a clean checkout, ticket `Prove the Vite MPA to Go embed path` must validate the recommended `embedded_web` build-tag implementation plus a typed unavailable-assets stub, and compare it with a mandatory frontend-first raw Go workflow before selection.
