.PHONY: build clean test release release-local

# Build variables
BINARY_NAME=go-togif
VERSION=$(shell git describe --tags --always --dirty)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

build:
	@echo "Building..."
	@go build -o $(GOBIN)/$(BINARY_NAME) .

clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)
	@rm -rf dist/
	@go clean

test:
	@echo "Testing..."
	@go test -v ./...

release:
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN is not set. Please export it first with:"; \
		echo "export GITHUB_TOKEN=<your_token_here> make release"; \
		exit 1; \
	fi
	@echo "Cleaning previous release artifacts..."
	@rm -rf dist/
	@echo "Creating release..."
	@goreleaser release

release-local:
	@echo "Cleaning previous release artifacts..."
	@rm -rf dist/
	@echo "Creating local release snapshot..."
	@goreleaser release --snapshot --clean

release-snapshot:
	@echo "Creating release snapshot..."
	@goreleaser release --snapshot --rm-dist

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/goreleaser/goreleaser@latest

.DEFAULT_GOAL := build