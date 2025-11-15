test:
	wasm-pack test --node

test-ci:
	wasm-pack test --node --release

build-wasm:
	wasm-pack build --target web --out-dir dist/wasm --release --features wasm

build-server:
	cargo build --release --features server

build: build-wasm build-server
