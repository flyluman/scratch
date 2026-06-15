.PHONY: all build run test lint vet fmt tidy migrate swagger docker-build docker-up docker-down clean

all: fmt vet lint tidy build test

build:
	go build -ldflags="-s -w" -o bin/scratch ./cmd/server

run:
	go run ./cmd/server

tidy:
	go mod tidy

test:
	go test ./internal/... -count=1 -short

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

swagger:
	PATH="$$(go env GOPATH)/bin:$$PATH" swag init -g cmd/server/main.go --output docs

migrate:
	go run ./cmd/server -mode=migrate

# Rename module path across all files
# Usage: make rename MODULE=github.com/yourname/yourproject
rename:
	@if [ -z "$(MODULE)" ]; then echo "Usage: make rename MODULE=github.com/yourname/yourproject"; exit 1; fi
	@echo "Renaming module to $(MODULE)..."
	find . -type f -name '*.go' -not -path './.git/*' -exec sed -i '' 's|github.com/flyluman/scratch|$(MODULE)|g' {} +
	sed -i '' 's|module github.com/flyluman/scratch|module $(MODULE)|' go.mod
	@echo "Done. Run 'go mod tidy' to update go.sum."

docker-build:
	docker build -t scratch:latest -f infra/docker/Dockerfile .

docker-up:
	docker compose -f infra/local/docker-compose.yml up -d

docker-down:
	docker compose -f infra/local/docker-compose.yml down

clean:
	rm -rf bin/ coverage.out coverage.html docs/
