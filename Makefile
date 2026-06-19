BINARY=karots-drop
BINDIR=bin
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build build-linux build-darwin build-windows build-arm64 test clean

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINDIR)/$(BINARY) ./cmd/$(BINARY)/

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINDIR)/$(BINARY)-linux-amd64 ./cmd/$(BINARY)/

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINDIR)/$(BINARY)-darwin-amd64 ./cmd/$(BINARY)/

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINDIR)/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)/

build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINDIR)/$(BINARY)-linux-arm64 ./cmd/$(BINARY)/

test:
	go test ./... -v -count=1

clean:
	rm -rf $(BINDIR)/
