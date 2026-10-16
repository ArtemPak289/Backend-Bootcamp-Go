.PHONY: all build up down run stop test local-up local-down certs hosts clean

DOCKER_COMPOSE ?= docker-compose

all: build

build: certs
	@echo "==> Building services..."
	$(DOCKER_COMPOSE) -f docker-compose.yml build

up: build
	@echo "==> Starting all services via Docker Compose..."
	$(DOCKER_COMPOSE) -f docker-compose.yml up -d
	@echo "==> Services started. Waiting for healthchecks..."
	@sleep 5
	$(DOCKER_COMPOSE) -f docker-compose.yml ps

run: up

down:
	@echo "==> Stopping all services via Docker Compose..."
	$(DOCKER_COMPOSE) -f docker-compose.yml down

stop: down

local-up: certs
	@echo "==> Starting native services locally..."
	./scripts/run_local.sh

local-down:
	@echo "==> Stopping native services locally..."
	./scripts/stop_local.sh

test:
	@echo "==> Running automated test suite..."
	./scripts/test_nginx.sh

certs:
	@if [ ! -f ssl/server.crt ] || [ ! -f ssl/server.key ]; then \
		echo "==> Generating SSL certificates..."; \
		./ssl/generate_certs.sh; \
	fi

hosts:
	@echo "==> Setting up local DNS in /etc/hosts..."
	./scripts/setup_hosts.sh

clean: down local-down
	@echo "==> Cleaning artifacts..."
	rm -f backend/server
	rm -rf /tmp/nginx_cache /tmp/store_pids
	rm -f /tmp/backend_*.log /tmp/backend_*.pid /tmp/nginx_*.log
	@echo "==> Clean complete."
