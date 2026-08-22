# Research a CGO-free FS thumbnail compatibility path

Type: research
Status: resolved
Blocked by: 01, 07

## Question

Using primary sources and reproducible probes, determine whether a maintained, license-compatible `CGO_ENABLED=0` path can reproduce the current FS thumbnail capability: AVIF, GIF, JPEG, PNG, and WebP input; the observed EXIF/orientation and animation behavior; and static 160x160 cover-fit WebP output inside all six macOS, Linux, and Windows x64/arm64 standalone artifacts. Compare pure-Go, Go-hosted WASM, and bundled-helper candidates only as needed, pin the smallest viable candidate and runtime artifacts, and verify representative legacy-derived input/output behavior plus all six target builds. Do not design new limits, concurrency policy, or thumbnail behavior. If exact parity is not feasible, record the failed cases and surface only the smallest implementation compatibility decision.

## Comments

- 2026-08-22, scope correction: retain this research only because Sharp/Bun thumbnail capability may lack an exact maintained `CGO_ENABLED=0` Go equivalent. It does not block Vite embedding, the base Go layout, cutover planning, or commands before FS; it must not expand into thumbnail product redesign or general FS hardening.
- 2026-08-23, research report: [CGO-free FS thumbnail compatibility path](../research/21-cgo-free-fs-thumbnails.md). The focused probe froze source commit `78358c0201b71891e36603d6abb8d7c87d54ad57`, compared 15 Bun/Go outputs, built all six targets with Go 1.26.7 and `CGO_ENABLED=0`, and proved persistent self-exec worker timeout/replacement.

## Answer

Use an unconditionally compiled pure-Go thumbnail graph: `github.com/gen2brain/gav1d/avif` from module `github.com/gen2brain/gav1d v0.2.5`, `github.com/gen2brain/vpx/webp` from `github.com/gen2brain/vpx v0.2.1`, `golang.org/x/image/draw v0.45.0`, and the Go standard library for JPEG, GIF, and PNG. Add only an owner-local JPEG EXIF Orientation 1..8 parser/transform. Pin the module versions, sums, and Go 1.26.7 recorded in the report; no codec build tag, runtime payload, system library, Node/Bun runtime, or helper file is permitted.

Reproduce the observed Bun behavior, not a generalized metadata policy: disable AVIF auto-rotation, enable WebP EXIF auto-rotation, apply JPEG EXIF orientation, ignore PNG `eXIf`, select the rendered/composited first frame of animations, center-cover into 160x160 with `draw.ApproxBiLinear`, and encode a static lossy WebP using quality 72 and method 4. The 15-case probe covered all five formats, alpha, CMYK JPEG, 10-bit and animated AVIF, animated GIF/WebP, every observed orientation class, and cover geometry. Every result was a valid one-frame 160x160 WebP; MAE ranged 0.6804-4.9840 and PSNR 31.08-46.98 dB against Bun, within the inventory's intent-compatibility contract. All six stripped cross-builds passed at approximately 4.57-5.30 MB.

Preserve the existing two persistent conversion workers through an unadvertised self-exec mode in the same ycy binary, entered before ordinary CLI composition. Use request-ID/length-framed stdin/stdout, drain stderr, and start the five-second timer at dispatch. On timeout, return the existing 504 behavior, `Process.Kill`, `Wait`, then launch a replacement from `os.Executable`; normal codec failures retain 422 and worker failure retains the current failure path. The probe demonstrated same-PID reuse, a stalled worker kill/reap, different-PID replacement, and all six builds without adding another shipped file.

Reject `jpegn v0.6.0` because the representative CMYK JPEG caused an index-out-of-range panic. Keep the `nodynamic` `gen2brain/avif v0.6.0` plus `gen2brain/webp v0.6.4` WASM graph only as research evidence, not fallback code: it passed but was materially larger and slower. Direct Emscripten reuse and Sharp/libvips helpers are also unnecessary. Retain the selected modules' BSD/MIT notices and patent grants in release third-party documentation. No compatibility exception or follow-on Wayfinder ticket is needed.
