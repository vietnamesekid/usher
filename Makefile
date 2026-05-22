BINARY  = usher
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build install test test-unit lint fmt clean cross-compile release-dry validate-registry

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

install:
	go install -ldflags="$(LDFLAGS)" .

test:
	go test ./... -v -race

test-unit:
	go test ./internal/... -v -race

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY) usher-linux-* usher-darwin-* usher-windows-*

cross-compile:
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o usher-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o usher-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o usher-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o usher-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o usher-windows-amd64.exe .

release-dry:
	goreleaser release --snapshot --clean

validate-registry:
	@echo "Validating registry JSON files..."
	@for f in internal/registry/mcp/*.json; do \
		python3 -c "import json,sys; json.load(open('$$f'))" && echo "  ok: $$f" || (echo "  FAIL: $$f"; exit 1); \
	done
	@echo "All registry files valid."
