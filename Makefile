BINARY := vaulty
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint vet clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/vaulty

test:
	go test ./... -count=1

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
