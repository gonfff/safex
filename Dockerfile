FROM --platform=$BUILDPLATFORM rust:1.91 AS rust-builder

WORKDIR /src
RUN cargo install wasm-pack
COPY rust ./rust
WORKDIR /src/rust

RUN wasm-pack build --target web --out-dir dist/wasm --release --features wasm

ARG TARGETPLATFORM
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
    rustup target add aarch64-unknown-linux-gnu && \
    apt-get update && apt-get install -y gcc-aarch64-linux-gnu && \
    CC=aarch64-linux-gnu-gcc \
    CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc \
    cargo build --release --features server --target aarch64-unknown-linux-gnu; \
    else \
    cargo build --release --features server; \
    fi

RUN mkdir -p /output && \
    if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
    cp target/aarch64-unknown-linux-gnu/release/libsafex_rust.so /output/ && \
    cp target/aarch64-unknown-linux-gnu/release/libsafex_rust.d /output/; \
    else \
    cp target/release/libsafex_rust.so /output/ && \
    cp target/release/libsafex_rust.d /output/; \
    fi

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
FROM golang:1.25 AS app

WORKDIR /app
COPY app/go.mod app/go.sum ./app/
RUN cd app && go mod download
COPY app ./app
COPY --from=web /src/frontend/dist/htmx.min.js app/web/static/vendor/
COPY --from=web /src/frontend/dist/output.css app/web/static/css/
COPY --from=rust-builder /src/rust/dist/wasm/ app/web/static/wasm/

COPY --from=rust-builder /output/libsafex_rust.so app/lib/
COPY --from=rust-builder /output/libsafex_rust.d app/lib/

RUN cd app && CGO_ENABLED=1 go build -o /out/safex ./cmd/api

# --------------------------------
FROM alpine:latest

RUN apk add --no-cache gcompat libgcc libstdc++

WORKDIR /app
COPY --from=app /out/safex ./safex
COPY --from=rust-builder /output/libsafex_rust.so ./
ENV LD_LIBRARY_PATH=/app
EXPOSE 8000
ENTRYPOINT ["/app/safex"]
