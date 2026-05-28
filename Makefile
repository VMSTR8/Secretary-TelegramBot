.PHONY: test lint build run tidy ci fmt docker-build docker-up docker-down docker-logs traefik-up traefik-down traefik-logs

COMPOSE = docker compose --env-file .env -f docker/docker-compose.yaml

GO              ?= go
GOLANGCI_LINT   ?= golangci-lint

TRAEFIK = docker compose -f docker/traefik/docker-compose.yaml

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

traefik-up:
	docker network inspect proxy >/dev/null 2>&1 || docker network create proxy
	@touch docker/traefik/acme.json
	@chmod 600 docker/traefik/acme.json
	$(TRAEFIK) up -d

traefik-down:
	$(TRAEFIK) down

traefik-logs:
	$(TRAEFIK) logs -f traefik