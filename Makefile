.PHONY: test lint build run tidy ci fmt docker-build docker-up docker-down docker-logs

COMPOSE = docker compose -f docker/docker-compose.yaml

GO              ?= go
GOLANGCI_LINT   ?= golangci-lint

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

docker-build:
	$(COMPOSE) build
docker-up:
	$(COMPOSE) up -d --build
docker-down:
	$(COMPOSE) down
docker-logs:
	$(COMPOSE) logs -f bot
