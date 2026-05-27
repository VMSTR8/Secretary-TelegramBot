# Makefile

GO              ?= go
GOLANGCI_LINT   ?= golangci-lint

.PHONY: test lint build run tidy ci fmt

test:
	$(GO) test -race -count=1 ./...

lint:
	$(GOLANGCI_LINT) run ./...

build:
	$(GO) build -o bin/service ./cmd/service

run:
	$(GO) run ./cmd/service

tidy:
	$(GO) mod tidy

fmt:
	$(GOLANGCI_LINT) run --fix ./...

ci: lint test
