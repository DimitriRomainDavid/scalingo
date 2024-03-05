.PHONY: all wire lint test build

all: wire lint test build

wire:
	@echo ">> Generating wire_gen.go"
	@$(GOPATH)/bin/wire ./cmd/app

test:
	@echo ">> Running tests"
	@go test -v ./...

build:
	@echo ">> Building binary"
	@go build -o ./cmd/app/app ./cmd/app

lint:
	@echo ">> Running golangci-lint"
	@$(GOPATH)/bin/golangci-lint run

.DEFAULT_GOAL := all
