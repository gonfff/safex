FROM --platform=$BUILDPLATFORM rust:1.91 AS rust-builder

WORKDIR /src
RUN cargo install wasm-pack
COPY rust ./rust
WORKDIR /src/rust

# Build WASM (архитектурно независимый)
RUN wasm-pack build --target web --out-dir dist/wasm --release --features wasm

# Установка cross-compilation tools для целевой архитектуры
ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
    "linux/amd64") echo "x86_64-unknown-linux-gnu" > /tmp/target ;; \
    "linux/arm64") echo "aarch64-unknown-linux-gnu" > /tmp/target ;; \
    *) echo "Unsupported platform: $TARGETPLATFORM" && exit 1 ;; \
    esac

RUN export TARGET=$(cat /tmp/target) && \
    rustup target add $TARGET && \
    if [ "$TARGET" = "aarch64-unknown-linux-gnu" ]; then \
    apt-get update && apt-get install -y gcc-aarch64-linux-gnu && \
    export CC_aarch64_unknown_linux_gnu=aarch64-linux-gnu-gcc && \
    export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc; \
    fi && \
    cargo build --release --features server --target $TARGET

# Копируем собранную библиотеку в правильное место
RUN export TARGET=$(cat /tmp/target) && \
    mkdir -p /output && \
    cp target/$TARGET/release/libsafex_rust.so /output/ && \
    cp target/$TARGET/release/libsafex_rust.d /output/

# --------------------------------
FROM node:25-alpine AS web

WORKDIR /src
COPY frontend/package*.json frontend/
RUN cd frontend && npm install
COPY frontend ./frontend
COPY app/web/ ./app/web/
RUN mkdir -p ./app/web/static
RUN cd frontend && npm run build

# --------------------------------
FROM --platform=$BUILDPLATFORM golang:1.25 AS app

ARG TARGETOS TARGETARCH

WORKDIR /app
COPY app/go.mod app/go.sum ./app/
RUN cd app && go mod download
COPY app ./app
COPY --from=web /src/frontend/dist/htmx.min.js app/web/static/vendor/
COPY --from=web /src/frontend/dist/output.css app/web/static/css/

COPY --from=rust-builder /output/libsafex_rust.so app/lib/
COPY --from=rust-builder /output/libsafex_rust.d app/lib/

RUN cd app && CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/safex ./cmd/api

# --------------------------------
FROM alpine:latest

RUN apk add --no-cache gcompat libgcc libstdc++

WORKDIR /app
COPY --from=app /out/safex ./safex
COPY --from=rust-builder /output/libsafex_rust.so ./
ENV LD_LIBRARY_PATH=/app
EXPOSE 8000
ENTRYPOINT ["/app/safex"]

