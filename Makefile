.PHONY: help build test clean

help:
	@echo "Backend-Bootcamp-Go Monorepo Targets:"
	@echo "  make build  - Compile all Go projects and microservices"
	@echo "  make test   - Run unit and integration tests across modules"
	@echo "  make clean  - Remove compiled binaries and test artifacts"

build:
	@echo "==> Building BE01-Http-Clients..."
	cd BE01-Http-Clients && go build ./...
	@echo "==> Building BE02-Store-REST-API..."
	cd BE02-Store-REST-API && go build ./...
	@echo "==> Building BE03-Nginx-ReverseProxy backend..."
	cd BE03-Nginx-ReverseProxy/backend && go build ./...
	@echo "==> Building BE04-GRPC-Microservices (auth_service)..."
	cd BE04-GRPC-Microservices/auth_service && go build ./...
	@echo "==> Building BE04-GRPC-Microservices (store_service)..."
	cd BE04-GRPC-Microservices/store_service && go build ./...
	@echo "==> All Go services compiled successfully!"

test:
	@echo "==> Testing BE01-Http-Clients..."
	cd BE01-Http-Clients && go test -v ./...
	@echo "==> Testing BE02-Store-REST-API..."
	cd BE02-Store-REST-API && go test -v ./...
	@echo "==> Testing BE04-GRPC-Microservices..."
	cd BE04-GRPC-Microservices/auth_service && go test -v ./...
	cd BE04-GRPC-Microservices/store_service && go test -v ./...

clean:
	@echo "==> Cleaning compiled artifacts..."
	rm -f BE01-Http-Clients/be1-httpauth
	rm -f BE02-Store-REST-API/appliance-store
	rm -f BE03-Nginx-ReverseProxy/backend/appliance-store
	rm -f BE04-GRPC-Microservices/auth_service/auth_service
	rm -f BE04-GRPC-Microservices/store_service/store_service
	@echo "==> Clean complete."
