SHELL := /bin/zsh

COMPOSE := docker compose
COMPOSE_ENV := --env-file .env
GO_CACHE := $(CURDIR)/.gocache

.PHONY: help env up down build logs ps restart sync query test fmt backend-test compose-config frontend-install frontend-build frontend-dev backend-run swagger clean

help:
	@echo "Targets:"
	@echo "  make env             # copy .env.example to .env if missing"
	@echo "  make up              # build and start frontend/backend"
	@echo "  make down            # stop containers"
	@echo "  make build           # rebuild containers"
	@echo "  make logs            # follow compose logs"
	@echo "  make ps              # show compose services"
	@echo "  make restart         # restart compose services"
	@echo "  make sync            # trigger Notion sync"
	@echo "  make query q='... '  # ask backend a question"
	@echo "  make test            # run backend tests"
	@echo "  make fmt             # format Go code"
	@echo "  make backend-run     # run backend locally"
	@echo "  make swagger         # regenerate Swagger docs"
	@echo "  make frontend-install# install frontend deps with yarn"
	@echo "  make frontend-build  # build frontend locally"
	@echo "  make frontend-dev    # run frontend dev server locally"
	@echo "  make compose-config  # render compose config"
	@echo "  make clean           # remove local build artifacts"

env:
	@test -f .env || cp .env.example .env

up: env
	$(COMPOSE) $(COMPOSE_ENV) up --build -d

down:
	$(COMPOSE) $(COMPOSE_ENV) down

build: env
	$(COMPOSE) $(COMPOSE_ENV) build

logs:
	$(COMPOSE) $(COMPOSE_ENV) logs -f

ps:
	$(COMPOSE) $(COMPOSE_ENV) ps

restart:
	$(COMPOSE) $(COMPOSE_ENV) restart

sync:
	curl -X POST http://localhost:$${BACKEND_PORT:-8080}/api/v1/sync

query:
	@test -n "$(q)" || (echo "usage: make query q='지난주 업무 요약해줘'" && exit 1)
	curl -X POST http://localhost:$${BACKEND_PORT:-8080}/api/v1/query \
		-H 'Content-Type: application/json' \
		-d '{"question":"$(q)"}'

test:
	mkdir -p "$(GO_CACHE)"
	cd backend && GOCACHE="$(GO_CACHE)" go test ./...

backend-test: test

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

backend-run: env
	mkdir -p "$(GO_CACHE)"
	cd backend && set -a && source ../.env && set +a && GOCACHE="$(GO_CACHE)" go run ./cmd/server

swagger:
	cd backend && $$(go env GOPATH)/bin/swag init -g main.go -d cmd/server,internal -o docs

frontend-install:
	cd frontend && yarn install

frontend-build:
	cd frontend && yarn build

frontend-dev:
	cd frontend && yarn dev

compose-config: env
	$(COMPOSE) $(COMPOSE_ENV) config

clean:
	rm -rf "$(GO_CACHE)" backend/.gocache frontend/dist frontend/node_modules frontend/*.tsbuildinfo frontend/vite.config.js frontend/vite.config.d.ts
