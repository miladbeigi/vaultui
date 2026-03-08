.PHONY: ci fmt vet lint test build tidy clean

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
	go build -o vaultui .

tidy:
	@echo "==> Checking module tidiness..."
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod/go.sum not tidy — commit the changes" && exit 1)

clean:
	rm -f vaultui
