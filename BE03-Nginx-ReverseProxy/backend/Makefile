.PHONY: all build run test swagger clean

BINARY_NAME=store_api

all: build

build: swagger
	go build -o $(BINARY_NAME) main.go

run: swagger
	go run main.go

swagger:
	@which swag >/dev/null 2>&1 && swag init || (test -x $(shell go env GOPATH)/bin/swag && $(shell go env GOPATH)/bin/swag init) || echo "Using pre-generated swagger docs"

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
