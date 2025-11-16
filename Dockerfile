FROM rust:1.91 AS rust

WORKDIR /src
RUN cargo install wasm-pack
COPY rust ./rust
WORKDIR /src/rust
RUN wasm-pack build --target web --out-dir dist/wasm --release --features wasm
RUN cargo build --release --features server

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

COPY --from=rust /src/rust/target/release/libsafex_rust.so app/lib/
COPY --from=rust /src/rust/target/release/libsafex_rust.d app/lib/

RUN cd app && CGO_ENABLED=1 GOOS=linux go build -o /out/safex ./cmd/api

# --------------------------------
FROM alpine:latest

RUN apk add --no-cache gcompat libgcc libstdc++

WORKDIR /app
COPY --from=app /out/safex ./safex
COPY --from=rust /src/rust/target/release/libsafex_rust.so ./
ENV LD_LIBRARY_PATH=/app
EXPOSE 8000
ENTRYPOINT ["/app/safex"]

