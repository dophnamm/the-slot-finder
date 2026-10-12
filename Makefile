# One-stop commands for The Slot Finder. Run `make` to list them.

SERVER_DIR := server
CLIENT_DIR := client
SERVER_ENV := $(SERVER_DIR)/cmd/server/.env
CLIENT_ENV := $(CLIENT_DIR)/.env
COMPOSE    := docker compose -f $(SERVER_DIR)/docker-compose.yml --env-file $(SERVER_ENV)
INFRA      := postgres redis keycloak otel-collector jaeger

.DEFAULT_GOAL := help
.PHONY: help setup env up down logs ps infra dev server client migrate seed keycloak-theme lint test build check reset-data

help: ## List the commands
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

setup: env ## First run: create .env files and install Go and npm dependencies
	cd $(SERVER_DIR) && go mod download
	cd $(CLIENT_DIR) && npm install

env: $(SERVER_ENV) $(CLIENT_ENV) ## Create missing .env files from the examples (existing ones are kept)

$(SERVER_ENV):
	cp $(SERVER_DIR)/cmd/server/.env.example $@

$(CLIENT_ENV):
	cp $(CLIENT_DIR)/.env.example $@

up: env ## Start everything in Docker: infra, migrations and the API server
	$(COMPOSE) up -d --build
	@echo ""
	@echo "  API       http://localhost:3000/api/v1/"
	@echo "  Swagger   http://localhost:3000/swagger/index.html"
	@echo "  Keycloak  http://localhost:8080/auth"
	@echo "  Jaeger    http://localhost:16686"
	@echo "  Client    make client  ->  http://localhost:5173"

down: ## Stop all containers (data is kept)
	$(COMPOSE) down

logs: ## Follow the API server logs
	$(COMPOSE) logs -f server

ps: ## Show container status
	$(COMPOSE) ps

infra: env ## Start only the infrastructure, to run the server on the host
	$(COMPOSE) up -d $(INFRA)

dev: infra ## Infra in Docker, then the server and the client on the host
	$(MAKE) -j2 server client

server: env ## Run the API server on the host (port 3000)
	cd $(SERVER_DIR) && $(MAKE) up

client: env ## Run the client dev server (port 5173, proxies /api to port 3000)
	cd $(CLIENT_DIR) && npm run dev

migrate: env ## Apply database migrations
	$(COMPOSE) run --rm migrate

seed: env ## Load the demo data (people, programs, a staffed baseline, scenarios); safe to run again
	cd $(SERVER_DIR) && $(MAKE) seed

keycloak-theme: env ## Apply the sign-in theme to the running Keycloak (a new realm gets it on import)
	$(COMPOSE) up -d --wait keycloak
	$(COMPOSE) exec keycloak bash -c '\
		/opt/keycloak/bin/kcadm.sh config credentials --server "http://localhost:8080$$KC_HTTP_RELATIVE_PATH" \
			--realm master --user "$$KC_BOOTSTRAP_ADMIN_USERNAME" --password "$$KC_BOOTSTRAP_ADMIN_PASSWORD" && \
		/opt/keycloak/bin/kcadm.sh update "realms/$$KEYCLOAK_REALM" -s "loginTheme=$$KEYCLOAK_LOGIN_THEME"'

lint: ## Lint the server and the client
	cd $(SERVER_DIR) && $(MAKE) lint
	cd $(CLIENT_DIR) && npm run lint

test: ## Run the server and client tests
	cd $(SERVER_DIR) && go test ./...
	cd $(CLIENT_DIR) && npm test

build: ## Build the server and the client
	cd $(SERVER_DIR) && go build ./...
	cd $(CLIENT_DIR) && npm run build

check: lint test build ## Lint, test and build everything

reset-data: ## Stop containers and DELETE all local data (Postgres, Redis, Keycloak volumes)
	$(COMPOSE) down -v
