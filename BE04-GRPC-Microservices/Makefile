.PHONY: all build proto up down run stop local-up local-down test clean

DOCKER_COMPOSE ?= docker compose

all: build

proto:
	@echo "==> Compiling protobuf definitions..."
	@mkdir -p auth_service/proto store_service/proto
	PATH="$(shell go env GOPATH)/bin:$$PATH" protoc \
		--proto_path=proto \
		--go_out=auth_service/proto --go_opt=paths=source_relative \
		--go-grpc_out=auth_service/proto --go-grpc_opt=paths=source_relative \
		proto/auth.proto
	@cp auth_service/proto/*.go store_service/proto/
	@echo "==> Protobuf code generation complete."

build: proto
	@echo "==> Building auth_service..."
	@cd auth_service && go build -o server main.go
	@echo "==> Building store_service..."
	@cd store_service && go build -o server main.go
	@echo "==> Build finished successfully."

up:
	@echo "==> Starting services via Docker Compose..."
	$(DOCKER_COMPOSE) up -d --build
	@echo "==> Waiting for services to become healthy..."
	@sleep 5
	$(DOCKER_COMPOSE) ps

run: up

down:
	@echo "==> Stopping Docker Compose services..."
	$(DOCKER_COMPOSE) down -v

stop: down

local-up: build
	@echo "==> Starting services natively..."
	./scripts/run_local.sh

local-down:
	@echo "==> Stopping native services..."
	./scripts/stop_local.sh

test:
	@echo "==> Running BE4 E2E Test Suite..."
	./scripts/test_e2e.sh

clean: local-down
	@echo "==> Cleaning built artifacts..."
	@rm -f auth_service/server store_service/server
	@rm -rf /tmp/be4_pids /tmp/be4_auth.log /tmp/be4_store.log
	@echo "==> Clean complete."
