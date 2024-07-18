SHELL := /bin/bash

GO ?= go
GO_CMD := CGO_ENABLED=0 $(GO)

.PHONY: all
all: test

include .bingo/Variables.mk

.PHONY: test
test:
	$(GO_CMD) test -cover ./...

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

.PHONY: tools
tools: $(BINGO) $(GOLANGCI_LINT)
	@echo Tools built.
