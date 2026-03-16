.PHONY: build clean test fmt vet install run

BINARY=codetree
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)
	rm -rf dist/

test:
	go test -v -race ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

install:
	go install .

run:
	go run .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  clean     - Remove binary and dist/"
	@echo "  test      - Run tests"
	@echo "  fmt       - Format code"
	@echo "  vet       - Vet code"
	@echo "  lint      - Format and vet"
	@echo "  install   - Install to GOPATH/bin"
	@echo "  run       - Run directly"
	@echo "  build-all - Build for all platforms"
