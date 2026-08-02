FROM oven/bun:1.3.14 AS build

WORKDIR /src
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --ignore-scripts
COPY . .
ARG TARGETARCH
RUN case "$TARGETARCH" in \
      amd64) BUN_ARCH=x64 ;; \
      arm64) BUN_ARCH=arm64 ;; \
      *) echo "Unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac \
    && bun scripts/build.ts --target "bun-linux-${BUN_ARCH}" --outfile /out/ycy \
    && XDG_STATE_HOME=/out/tunnel-state bun scripts/install-tunnel-frp.ts "$BUN_ARCH"

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /usr/share/licenses/frp \
    && cp /usr/share/common-licenses/Apache-2.0 /usr/share/licenses/frp/LICENSE

COPY --from=build /out/ycy /usr/local/bin/ycy
COPY --from=build /out/tunnel-state/ycy/frp /opt/ycy/frp

ENV YCY_TUNNEL_DOCKER=1

ENTRYPOINT ["/usr/local/bin/ycy"]
