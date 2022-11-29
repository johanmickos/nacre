GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
BINARY_NAME=nacre-server

TOOLS_BIN_DIR ?= $(shell pwd)/tmp/bin
export PATH := $(TOOLS_BIN_DIR):$(PATH)

GOLINT=$(TOOLS_BIN_DIR)/golint
GOSTATICCHECK=$(TOOLS_BIN_DIR)/staticcheck
TOOLS=$(GOLINT) $(GOSTATICCHECK)

.PHONY: all
all: test build

.PHONY: test
test:
	$(GOTEST) -v --race ./...

.PHONY: build
build:
	mkdir -p out/bin
	$(GO) build -o out/bin/$(BINARY_NAME) ./cmd/server

.PHONY: run
run:
	$(GO) run cmd/server/main.go

.PHONY: check
check: $(GOLINT) $(GOSTATICCHECK)
	golint ./...
	staticcheck ./...

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(TOOLS): $(TOOLS_BIN_DIR)
	@echo "Installing build tools from tools.go"
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(TOOLS_BIN_DIR) xargs -tI % go install -mod=readonly -modfile=go.mod %