SHELL := /bin/zsh

COMPOSE := docker compose
COMPOSE_ENV := --env-file .env
GO_CACHE := $(CURDIR)/.gocache
LOCAL_DATA_DIR := $(CURDIR)/.local-data
BACKEND_PORT ?= 8080
LOCAL_HTTP_ADDR ?= :$(BACKEND_PORT)

.PHONY: help env up down build logs ps restart sync query test fmt backend-test compose-config frontend-install frontend-ensure frontend-build frontend-dev backend-run swagger clean clean-deps clean-data

help:
	@echo "Targets:"
	@echo "  make env                         # copy .env.example to .env if missing"
	@echo "  make up                          # build and start frontend/backend"
	@echo "  make down                        # stop containers"
	@echo "  make build                       # rebuild containers"
	@echo "  make logs                        # follow compose logs"
	@echo "  make ps                          # show compose services"
	@echo "  make restart                     # restart compose services"
	@echo "  make sync                        # trigger Notion sync"
	@echo "  make query q='... '              # ask backend a question"
	@echo "  make test                        # run backend tests"
	@echo "  make fmt                         # format Go code"
	@echo "  make backend-run                 # run backend locally with .local-data"
	@echo "  make backend-run BACKEND_PORT=8081 # run backend on another port"
	@echo "  make swagger                     # regenerate Swagger docs"
	@echo "  make frontend-install            # install frontend deps with yarn"
	@echo "  make frontend-ensure             # install frontend deps only when missing"
	@echo "  make frontend-build              # build frontend locally"
	@echo "  make frontend-dev                # run frontend dev server locally"
	@echo "  make compose-config              # render compose config"
	@echo "  make clean                       # remove local build artifacts"
	@echo "  make clean-deps                  # remove frontend node_modules"
	@echo "  make clean-data                  # remove local backend data"

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
	curl -X POST http://localhost:$(BACKEND_PORT)/api/v1/sync

query:
	@test -n "$(q)" || (echo "usage: make query q='지난주 업무 요약해줘'" && exit 1)
	curl -X POST http://localhost:$(BACKEND_PORT)/api/v1/query \
		-H 'Content-Type: application/json' \
		-d '{"question":"$(q)"}'

test:
	mkdir -p "$(GO_CACHE)"
	cd backend && GOCACHE="$(GO_CACHE)" go test ./...

backend-test: test

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

backend-run: env
	@if command -v lsof >/dev/null 2>&1 && lsof -nP -iTCP:$(BACKEND_PORT) -sTCP:LISTEN >/dev/null 2>&1; then \
		echo "port $(BACKEND_PORT) is already in use. Run 'make down' or 'make backend-run BACKEND_PORT=8081'."; \
		exit 1; \
	fi
	mkdir -p "$(GO_CACHE)" "$(LOCAL_DATA_DIR)"
	cd backend && set -a && source ../.env && set +a && DATA_DIR="$(LOCAL_DATA_DIR)" HTTP_ADDR="$(LOCAL_HTTP_ADDR)" GOCACHE="$(GO_CACHE)" go run ./cmd/server

swagger:
	cd backend && $$(go env GOPATH)/bin/swag init -g main.go -d cmd/server,internal -o docs

frontend-install:
	cd frontend && yarn install --frozen-lockfile

frontend-ensure:
	cd frontend && test -x node_modules/.bin/vite || yarn install --frozen-lockfile

frontend-build: frontend-ensure
	cd frontend && yarn build

frontend-dev: frontend-ensure
	cd frontend && yarn dev

compose-config: env
	$(COMPOSE) $(COMPOSE_ENV) config

clean:
	rm -rf "$(GO_CACHE)" backend/.gocache frontend/dist frontend/*.tsbuildinfo frontend/vite.config.js frontend/vite.config.d.ts

clean-deps:
	rm -rf frontend/node_modules

clean-data:
	rm -rf "$(LOCAL_DATA_DIR)"
