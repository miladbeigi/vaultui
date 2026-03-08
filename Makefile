VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w \
           -X github.com/miladbeigi/vaultui/internal/version.Version=$(VERSION) \
           -X github.com/miladbeigi/vaultui/internal/version.Commit=$(COMMIT) \
           -X github.com/miladbeigi/vaultui/internal/version.Date=$(DATE)

.PHONY: ci fmt vet lint test build tidy clean release

ci: fmt vet lint test build tidy

fmt:
	@echo "==> Checking formatting..."
	@test -z "$$(gofmt -l .)" || (gofmt -l . && echo "Run 'gofmt -w .' to fix" && exit 1)

vet:
	@echo "==> Running go vet..."
	go vet ./...

lint:
	@echo "==> Running golangci-lint..."
	@which golangci-lint > /dev/null 2>&1 || (echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

test:
	@echo "==> Running tests..."
	go test ./...

build:
	@echo "==> Building binary..."
	go build -ldflags '$(LDFLAGS)' -o vaultui .

tidy:
	@echo "==> Checking module tidiness..."
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod/go.sum not tidy — commit the changes" && exit 1)

clean:
	rm -f vaultui

release:
ifndef VERSION_TAG
	$(error Usage: make release VERSION_TAG=v0.1.0)
endif
	@echo "==> Tagging $(VERSION_TAG)..."
	git tag -a $(VERSION_TAG) -m "Release $(VERSION_TAG)"
	git push origin $(VERSION_TAG)
	@echo "==> Tag $(VERSION_TAG) pushed. GitHub Actions will create the release."
