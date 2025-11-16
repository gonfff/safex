
install-env:
	cd rust && make build && cd ..
	cd web && npm install && cd ..


install-deps:
	@echo "Copying wasm files..."
	mkdir -p app/web/static/wasm
	cp rust/dist/wasm/* app/web/static/wasm/

	@echo "Copying library files..."
	mkdir -p app/lib
	cp rust/target/release/libsafex_rust.rlib app/lib/
	cp rust/target/release/libsafex_rust.d app/lib/
	cp rust/target/release/libsafex_rust.dylib app/lib/

	@echo "Copying frontend assets..."
	mkdir -p app/web/static/vendor
	mkdir -p app/web/static/css
	cp frontend/dist/htmx.min.js app/web/static/vendor/
	cp frontend/dist/output.css app/web/static/css/

build:
	@echo "Build rust"
	cd rust && make build && cd ..
	@echo "Build frontend"
	cd frontend && npm run build && cd ..

	@echo "Installing dependencies..."
	make install-deps

	@echo "Build app"
	cd app && make build && cd ..

run:
	./app/bin/app
