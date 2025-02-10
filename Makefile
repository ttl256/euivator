.DEFAULT_GOAL := build

GOCMD := go

EUIVATOR_CLI_BIN := euivator
VERSION := $(shell git describe --tags --abbrev=0 --always)

GOLANGCI_CFG := .golangci.yml

# QA

.PHONY: test
test:
	$(GOCMD) test -race ./...

.PHONY: audit
audit: test
	$(GOCMD) mod tidy -diff
	$(GOCMD) mod verify
	golangci-lint run --config $(CURDIR)/$(GOLANGCI_CFG)

.PHONY: bench
bench:
	$(GOCMD) test -benchmem -bench=. ./...

# DEV

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: build
build:
	$(GOCMD) build -ldflags="-X github.com/ttl256/$(EUIVATOR_CLI_BIN)/cmd.version=$(VERSION)" -o ./bin/$(EUIVATOR_CLI_BIN)

.PHONY: build_prod
build_prod:
	$(GOCMD) build -trimpath -ldflags="-s -w -X github.com/ttl256/$(EUIVATOR_CLI_BIN)/cmd.version=$(VERSION)" -o ./bin/$(EUIVATOR_CLI_BIN)
	upx -qq --best ./bin/$(EUIVATOR_CLI_BIN)

# OPS

.PHONY: no-dirty
no-dirty:
	@test -z "$(shell git status --porcelain)"

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: push
push: confirm audit build no-dirty
