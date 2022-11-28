GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
BINARY_NAME=nacre-server

.PHONY: all test build run

all: test build

test:
	$(GOTEST) -v --race ./...

build:
	mkdir -p out/bin
	$(GO) build -o out/bin/$(BINARY_NAME) ./cmd/server

run:
	$(GO) run cmd/server/main.go