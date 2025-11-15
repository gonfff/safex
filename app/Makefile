run:
	go run ./cmd/api

build:
	go build -o bin/app ./cmd/api

test-cov:
	go test -cover ./...

test-cov-report:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out