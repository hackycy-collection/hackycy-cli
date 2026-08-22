# CGO-free FS thumbnail compatibility path

Date: 2026-08-23
Frozen ycy source: [`78358c0201b71891e36603d6abb8d7c87d54ad57`](https://github.com/hackycy/hackycy-cli/tree/78358c0201b71891e36603d6abb8d7c87d54ad57)
Decision ticket: [Research a CGO-free FS thumbnail compatibility path](../issues/21-research-cgo-free-fs-thumbnails.md)

## Decision

Use this exact first-release graph inside the `fs` thumbnail Module:

| Purpose | Selection | Pin |
| --- | --- | --- |
| AVIF/AV1 decode | `github.com/gen2brain/gav1d/avif` | module `github.com/gen2brain/gav1d v0.2.5`, commit `7aa50e4e898ebdb4794ae3df88ce8ad6cbb74608`, sum `h1:Zg5/DE1JBdf8kxZqIvNo2ZEizF9DVF9jdPTAKJQZbOY=` |
| WebP/VP8 decode and WebP encode | `github.com/gen2brain/vpx/webp` | module `github.com/gen2brain/vpx v0.2.1`, commit `6e928b99b498beeec69e0eb467518c8bd0691816`, sum `h1:UD6tibcs/zlaIOKNFXeVyZuHBEaW9GYfxKSxvAN9cc8=` |
| Cover resize | `golang.org/x/image/draw` | module `golang.org/x/image v0.45.0`, commit `3ebddc7c54bd879f8d84d11db82892726f5192fd`, sum `h1:FMb1nTbH5H9vF55SriQHgFw5GnNL9Jg6L25BwXKzhB0=` |
| JPEG, GIF, PNG | Go standard library | pinned Go toolchain `go1.26.7` |
| JPEG orientation | owner-local narrow EXIF/TIFF parser and orientation transform | no added module |

This path is compiled unconditionally into every ycy binary. It has no codec payload to extract, no system-codec lookup, no helper executable, no CGO, and no build tag. The modules require at most Go 1.26.4 and fit the already selected Go 1.26.7 toolchain. Their amd64/arm64 assembly is ordinary Go assembler selected with runtime CPU checks; `-tags noasm` exists but is not required for `CGO_ENABLED=0` or the release artifacts.

Preserve the existing two-worker process boundary by starting the current ycy executable in an unadvertised internal thumbnail-worker mode. This is self-execution of the one shipped file, not a second embedded or materialized runtime. A five-second timeout can then kill, reap, and replace the stalled process, which an in-process goroutine cannot guarantee.

No compatibility exception is required. Exact encoded bytes and pixels remain intent compatibility under the FS inventory; the selected path reproduces every capability and the observed orientation, animation, dimensions, crop, alpha, and static-output behavior with an acceptable visual result.

## Frozen Bun behavior

The legacy worker pins [`wasm-image-optimization@2.0.9`](https://www.npmjs.com/package/wasm-image-optimization/v/2.0.9), npm integrity `sha512-g8Pj+vk+Sijex1MISslN0LBUnDdoqmYmaVho6cekwfe/6RQt4U9P5cmkvsHVUlj/ZDgeAjaOvoiTOwJOg3uBag==`, from git commit [`a028669ff947321f95d367c866a32f95235f670d`](https://github.com/node-libraries/wasm-image-optimization/tree/a028669ff947321f95d367c866a32f95235f670d). The ycy call is fixed at:

```text
width: 160
height: 160
fit: cover
format: webp
quality: 72
animation: false
```

The pinned optimizer source:

- reads `SkCodec::getOrigin()` and applies `SkPixmapUtils::Orient` for formats where Skia reports an origin;
- decodes the codec's frames and, because `animation:false`, sends frame zero to the static WebP encoder;
- implements cover as the maximum width/height scale, centered on a 160x160 canvas;
- uses `SkFilterMode::kLinear` while resizing; and
- writes lossy WebP at the supplied quality.

The focused Bun probe established the format-specific behavior that the Go port must reproduce, rather than assuming every metadata container behaves alike:

| Input | Observed orientation | Observed animation/output |
| --- | --- | --- |
| AVIF | `irot`/`imir` not applied by the pinned optimizer path | animated input recognized; output is static frame zero |
| GIF | no orientation metadata | animated input uses the first rendered frame; output is static |
| JPEG | EXIF orientation applied | static output |
| PNG | `eXIf` orientation ignored | static output with alpha retained where representable |
| WebP | EXIF orientation applied | animated input uses the first composited frame; output is static |

The Go integration therefore must not add a generic metadata normalizer. Use `gav1d` with AVIF auto-rotation disabled, `vpx/webp` with WebP auto-rotation enabled, the owner-local JPEG EXIF transform, and ordinary PNG decoding without `eXIf` processing.

## Selected conversion contract

Dispatch by the already validated MIME/extension allowlist and preserve the existing pre-dispatch 64 MiB and 50,000,000-pixel checks. The codec operation is:

1. Decode AVIF sequences with `avif.DecodeAll(..., avif.Options{AutoRotate:false})` and select image zero.
2. Decode WebP with `webp.Options{AutoRotate:true}` and select the first composited frame for animation.
3. Decode GIF animation on its logical canvas and select its first rendered frame.
4. Decode JPEG with the standard library, parse only APP1 `Exif\x00\x00` TIFF Orientation values 1 through 8, and apply that transform.
5. Decode PNG with the standard library and do not interpret `eXIf`.
6. Compute the maximum of `160/sourceWidth` and `160/sourceHeight`, center it on a transparent 160x160 NRGBA canvas, and transform with `draw.ApproxBiLinear`.
7. Encode one frame using `webp.EncodeOptions{Quality:72, Method:4}`.

`draw.BiLinear` was not selected merely because its name resembles Skia's filter. A direct A/B over the same corpus made the wider downsampling kernel both slower and less similar to the Bun output: the worst PSNR fell from 31.08 dB to 25.90 dB and the 15-file host run rose from about 0.23 seconds to 0.69 seconds. `ApproxBiLinear` uses a four-neighbor blend while downscaling, which matched this Skia path more closely.

## Compatibility probe

The temporary Go probe used Go 1.26.7 with `CGO_ENABLED=0`, the exact pins above, the same center-cover transform and encoder options, and Bun running the frozen worker call as the reference. It covered 15 inputs:

| Case | Capability | MAE | PSNR dB |
| --- | --- | ---: | ---: |
| alpha PNG | alpha and nonsquare cover | 0.6804 | 46.98 |
| animated AVIF | sequence and frame zero | 1.6956 | 38.58 |
| large GIF | GIF decode and downscale | 3.5263 | 33.00 |
| animated WebP | composited frame zero | 1.7216 | 38.36 |
| CMYK JPEG | CMYK conversion | 2.5599 | 35.79 |
| generated cover PNG | exact center-cover geometry | 1.2043 | 42.61 |
| lossy WebP | VP8 decode | 4.9638 | 31.08 |
| oriented AVIF | preserve ignored `irot`/`imir` behavior | 1.5690 | 39.96 |
| oriented JPEG | apply EXIF orientation | 3.0913 | 34.46 |
| oriented PNG | preserve ignored `eXIf` behavior | 1.0428 | 43.81 |
| oriented WebP | apply EXIF orientation | 4.9840 | 31.19 |
| still AVIF | ordinary 8-bit AVIF | 2.2141 | 36.97 |
| still JPEG | ordinary JPEG | 3.1156 | 33.58 |
| generated animated GIF | positive animation detection and frame zero | 1.4824 | 37.81 |
| 10-bit AVIF | high-bit-depth decode | 2.5058 | 35.89 |

All 15 outputs decoded as one-frame 160x160 WebP. Visual inspection of the lowest-PSNR photographic, orientation, and animation cases showed the same crop, orientation, frame, and acceptable appearance as Bun. Exact byte equality is neither achieved nor required: two different lossy WebP encoders and their color conversions cannot be expected to emit identical bytes, and the FS inventory already classifies exact thumbnail pixels/bytes as intent compatibility.

The reproducible core fixtures come from the pinned candidates' testdata (`avif v0.6.0` `test.avifs`, `test10.avif`, and `test_rot.avif`; `vpx v0.2.1` `anim.webp`, `exif.webp`, and `test.webp`; and a CMYK JPEG), plus deterministic generated PNG/GIF geometry and metadata cases. The probe recorded SHA-256 for every input and compared decoded RGBA output, not encoded-file hashes. The implementation Slice must turn these capabilities into repository-owned tests derived from `legacy/bun/`, as required by the migration roadmap; the temporary research program itself is not production code or an active test dependency.

The exact research input manifest was:

| Input | Origin | SHA-256 |
| --- | --- | --- |
| `alpha.png` | generated alpha/cover case | `fddd374526616254786229b4d2d8bb8449c3234b60c49eface2206587bf3798c` |
| `animated.avif` | `github.com/gen2brain/avif v0.6.0/testdata/test.avifs` | `9812609737050a7a5cb4745cceb81b1ab459629329d57544cb7c2638e4b6dfe2` |
| `animated.gif` | supplemental animated GIF | `f2db72ec5656632259602b23c6c20a412d5452ee0f99d4dcbec8d8836c485251` |
| `animated.webp` | `github.com/gen2brain/vpx v0.2.1/webp/testdata/anim.webp` | `3302424d6a337987a85797595b7533c9a8176e4b4ff2c101c3821c447f6f653b` |
| `cmyk.jpg` | `github.com/gen2brain/jpegli v0.4.2/testdata/cmyk.jpg` | `7cbedec0dda13d6893bb23203b1aab272ba024db1e6ea156b1c594b5915a434c` |
| `cover.png` | generated 480x240 center-cover case | `3045271098cf271be98b769ac3c09478b7c0ac6b7b2db2ab8bf4d79d098d4af2` |
| `lossy.webp` | `github.com/gen2brain/vpx v0.2.1/webp/testdata/test.webp` | `4e867b2be218ac30f715e61169f2f1b2615c39a5387bb365aa863b8bc1b8a60d` |
| `oriented.avif` | `github.com/gen2brain/avif v0.6.0/testdata/test_rot.avif` | `cb0aa72a7c2feb467e9f659db5ffbcf60884eaac065ac3d5ac969556d7c66ff0` |
| `oriented.jpg` | supplemental EXIF JPEG | `9b344e9f0c869d8637ea22e672df9451d8d3cc1d2d0b291af3b284e538e5f124` |
| `oriented.png` | generated PNG `eXIf` Orientation 6 case | `3ed1f34f9c9946eaa864744d9c87cb8485feaf5ca1a1f2bb017bc4027b7936d8` |
| `oriented.webp` | `github.com/gen2brain/vpx v0.2.1/webp/testdata/exif.webp` | `a79a811d092f678921c6d3f28e52d72914e423ebf9979ffee6bf88db9a27e037` |
| `plain.avif` | supplemental still AVIF | `d947e4d2725e90a9c3df42360acb56db6ff0b6079af51263998c00fafba0c3de` |
| `plain.jpg` | supplemental still JPEG | `df7a970e4b0f69fd7bf227ed671c68634c907d5cafcdb01fe4f256e0029288c7` |
| `synthetic-animated.gif` | generated 80x40 two-frame GIF | `a6f2f6c34b50b7e9ea365e73a5c490c8c872a881f055c3dd9fb4dc060a8b26f6` |
| `tenbit.avif` | `github.com/gen2brain/avif v0.6.0/testdata/test10.avif` | `fd9ecae5c70d0929286e2b2296a05cd206d7154537f4a9295d64ec33d436cb4a` |

Supplemental photographic samples are evidence, not planned repository fixtures. The implementation suite should use repository-owned or dependency testdata with equivalent semantics, record its own hashes, and consult `legacy/bun/` for expected behavior rather than preserving a second legacy golden corpus.

### Six target builds

The complete pure-Go probe, including its self-exec worker mode, built with `go1.26.7`, `CGO_ENABLED=0`, `-trimpath`, and stripped symbols:

| Target | Format/CPU inspection | Bytes |
| --- | --- | ---: |
| `darwin/amd64` | Mach-O x86_64 | 5,170,128 |
| `darwin/arm64` | Mach-O arm64 | 4,570,498 |
| `linux/amd64` | statically linked ELF x86-64 | 5,136,546 |
| `linux/arm64` | statically linked ELF AArch64 | 4,587,682 |
| `windows/amd64` | PE32+ x86-64 | 5,302,784 |
| `windows/arm64` | PE32+ AArch64 | 4,596,736 |

`go version -m` on the host artifact reported Go 1.26.7, the three selected modules, `CGO_ENABLED=0`, and the requested target. macOS linked only normal OS system libraries; the Linux artifacts were static. Cross-compilation proves portability of the selected source graph, while the roadmap still requires native thumbnail tests from the final ycy artifact on every target before Release Accepted.

### Timeout and replacement probe

The same probe binary implemented an internal length-framed stdin/stdout worker. One child handled two consecutive requests with the same PID, a deliberate stalled request hit its deadline and was killed/reaped, and the next request succeeded from a different replacement PID:

```text
persistent=[pid=61537 payload=FIRST] [pid=61537 payload=SECOND]
timeout="worker timed out"
replacement=[pid=61539 payload=AFTER]
```

The self-exec variant added roughly 0.27 MiB to the host probe and no second file. It also cross-compiled for all six targets. This is sufficient to retain a real five-second termination boundary without introducing a bundled helper.

## Worker integration contract

The Go FS Module preserves the existing pool contract as follows:

- Intercept an unadvertised internal worker argument before Cobra or normal application composition. Start it only through the parent process; do not expose a product command.
- Resolve the current executable with `os.Executable()` and keep two child processes alive. Each slot accepts one task at a time; the service retains the existing 128-task queue, coalescing, cache, errors, and shutdown behavior.
- Frame request ID, format, and input length/bytes on stdin; frame request ID, success/error, and output length/bytes on stdout. Continuously drain a separate stderr stream into non-secret diagnostics so a full pipe cannot stall a worker.
- Start the five-second timer at dispatch, including pipe transfer. On timeout, return the existing 504 conversion-timeout error, call `Process.Kill`, call `Wait` to reap the process, then start a replacement in the same slot and continue dispatch.
- Treat a normal codec rejection as the existing 422 conversion failure. Treat an unexpected worker exit as the existing worker failure path; if all slots fail, reject queued work as today.
- On service close, reject queued/in-flight tasks as today, kill both workers, close pipes, and `Wait` for both. The codec worker starts no descendants, so process groups and Windows Job Objects are unnecessary for this boundary.

`Process.Kill` terminates the process on all required Go targets. Only the matching native acceptance runs can prove final executable replacement, stderr, shutdown, antivirus, and process behavior; those remain artifact gates rather than new design questions.

## Rejected candidates

| Candidate | Evidence and disposition |
| --- | --- |
| `github.com/gen2brain/jpegn v0.6.0` | Its API and maintenance looked suitable, but it panicked with an index-out-of-range in `convertToCMYK` on the representative 512x512 CMYK JPEG. Reject it. Standard-library JPEG plus the narrow owner-local EXIF parser passed the same case and adds no module. |
| `draw.BiLinear` | Passed functionally but was materially farther from Bun on the same corpus and slower. Use the evidenced `ApproxBiLinear` setting. |
| `github.com/gen2brain/avif v0.6.0` plus `github.com/gen2brain/webp v0.6.4` | Valid CGO-free fallback only with mandatory `-tags nodynamic`; without it both first inspect host shared libraries. The wazero probe passed the corpus and six builds but was about 13.88 MiB and about 1.00 second for the batch on the host, versus 4.57 MiB and about 0.23 second for the selected path. The `wasm2go` AVIF variant was about 35.26 MiB and 0.40 second. Keep no fallback code unless a later concrete candidate regression requires a new decision. |
| Direct `wasm-image-optimization@2.0.9` reuse | The npm package is about 24.3 MB unpacked and includes a 6,935,484-byte Emscripten WASM module plus generated JS imports. The WASM is not standalone; the package CLI itself requires Node. Rehosting it would embed another runtime/glue system rather than produce a Go-native one-file implementation. |
| Sharp/libvips helper | Current Sharp packages require Node and target-specific native addons/libvips. Upstream libvips releases source, not the six ready standalone helpers required here. This adds a second runtime/materialization path despite the selected in-binary graph passing. |
| `bep/imagemeta` or another generic metadata parser | Not needed. The probe proves PNG `eXIf` and AVIF transforms are ignored by the frozen Bun behavior; enabling them would redesign observable output. JPEG needs only a small TIFF Orientation 1..8 reader, while selected WebP already provides the required behavior. |

## Licensing and release evidence

The selected source licenses are compatible with the repository's MIT distribution when their notice conditions are followed:

- `gav1d`: BSD-2-Clause `COPYING` plus the Alliance for Open Media Patent License 1.0 in `PATENTS`;
- `vpx`: BSD-3-Clause `COPYING`, the webpkit MIT notice in `COPYING.webpkit`, and the WebM patent grant in `PATENTS`;
- `x/image`: BSD-3-Clause `LICENSE` and the Go patent grant in `PATENTS`; and
- Go standard library: Go's BSD `LICENSE` and patent grant in `PATENTS`.

Check in one generated or curated third-party-notices source during implementation and reproduce these exact notices/patent texts in release documentation or other written material shipped with the binary distribution. This legal material is not a runtime payload and does not prevent the executable from working as the only installed file. The Final Artifact Gate must verify the pinned module sums and the presence of the corresponding release notices.

## Implementation acceptance

The FS thumbnail Slice is complete only when tests derived from the frozen Bun reference prove:

1. all five input formats, alpha, CMYK JPEG, 10-bit AVIF, ordinary stills, and malformed inputs;
2. JPEG/WebP orientation applied, PNG/AVIF orientation ignored, and animation reduced to the correct static first frame;
3. exact 160x160 center-cover geometry, static WebP MIME/decode, quality setting, hard orientation/frame/crop assertions, and on the representative legacy-derived corpus both `MAE <= 6` and `PSNR >= 30 dB` rather than byte equality;
4. the existing 64 MiB, 50M-pixel, two-worker, 128-queue, five-second, coalescing, LRU, HTTP, ETag, status, and shutdown contracts;
5. persistent self-exec reuse, worker crash, deliberate timeout, kill plus `Wait`, different-process replacement, all-worker failure, and no child/process leak;
6. Go 1.26.7 `CGO_ENABLED=0` builds for all six fixed artifact targets and final native runs on the matching matrix; and
7. module sums, absence of system codec/helper lookup, one-file runtime operation, and release third-party notices.

## Primary sources

- [Frozen ycy thumbnail worker](https://github.com/hackycy/hackycy-cli/blob/78358c0201b71891e36603d6abb8d7c87d54ad57/src/commands/fs/thumbnail-worker.ts) and [worker pool](https://github.com/hackycy/hackycy-cli/blob/78358c0201b71891e36603d6abb8d7c87d54ad57/src/commands/fs/thumbnail-service.ts)
- [Pinned optimizer C++ source](https://github.com/node-libraries/wasm-image-optimization/blob/a028669ff947321f95d367c866a32f95235f670d/src/cpp/api/image_converter_api.cpp)
- [`gav1d v0.2.5` AVIF API](https://github.com/gen2brain/gav1d/blob/7aa50e4e898ebdb4794ae3df88ce8ad6cbb74608/avif/avif.go), [CI](https://github.com/gen2brain/gav1d/blob/v0.2.5/.github/workflows/test.yml), and [license files](https://github.com/gen2brain/gav1d/tree/v0.2.5)
- [`vpx v0.2.1` WebP API](https://github.com/gen2brain/vpx/blob/6e928b99b498beeec69e0eb467518c8bd0691816/webp/webp.go), [animation decode](https://github.com/gen2brain/vpx/blob/6e928b99b498beeec69e0eb467518c8bd0691816/webp/anim.go), [CI](https://github.com/gen2brain/vpx/blob/v0.2.1/.github/workflows/test.yml), and [license files](https://github.com/gen2brain/vpx/tree/v0.2.1)
- [`x/image/draw` interpolation definitions](https://cs.opensource.google/go/x/image/+/v0.45.0:draw/scale.go) and [`x/image` license files](https://cs.opensource.google/go/x/image/+/v0.45.0:)
- [Go `os.Executable`](https://pkg.go.dev/os#Executable), [`Process.Kill`](https://pkg.go.dev/os#Process.Kill), and [`Cmd.Wait`](https://pkg.go.dev/os/exec#Cmd.Wait)
- [Go build constraint and cgo documentation](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [Emscripten compiler output and non-standalone WASM/JS relationship](https://emscripten.org/docs/compiling/WebAssembly.html#compiler-output)
- [Sharp installation and platform requirements](https://sharp.pixelplumbing.com/install/) and [libvips releases](https://github.com/libvips/libvips/releases/tag/v8.18.5)
