.PHONY: build clean test release

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
	@go clean

test:
	@echo "Testing..."
	@go test -v ./...

release:
	@echo "Releasing..."
	@goreleaser release --rm-dist

release-snapshot:
	@echo "Creating release snapshot..."
	@goreleaser release --snapshot --rm-dist

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go install github.com/goreleaser/goreleaser@latest

.DEFAULT_GOAL := build