# Development Guide

## Prerequisites

- **Go 1.25+** — [install instructions](https://go.dev/doc/install)
- A terminal emulator with true-color support (iTerm2, Ghostty, Alacritty, kitty, etc.)
- (Optional) A running HashiCorp Vault instance for live testing

## Getting Started

Clone the repo and install dependencies:

```sh
git clone <repo-url> && cd vaultui
go mod download
```

## Build

Build the binary into the current directory:

```sh
go build -o vaultui .
```

The resulting `vaultui` binary is a standalone executable.

## Run

Run directly without building a binary:

```sh
go run .
```

### CLI Flags

| Flag | Description | Env Var |
|---|---|---|
| `--vault-addr` | Vault server address | `VAULT_ADDR` |
| `--token` | Vault authentication token | `VAULT_TOKEN` |
| `--namespace` | Vault namespace | `VAULT_NAMESPACE` |
| `--config` | Path to config file (default `~/.vaultui.yaml`) | — |

Examples:

```sh
# Using flags
go run . --vault-addr=http://127.0.0.1:8200 --token=hvs.xxxxx

# Using environment variables
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=hvs.xxxxx
go run .

# Using a custom config file
go run . --config=./my-config.yaml
```

## Test

Run all tests:

```sh
go test ./...
```

Run tests with verbose output:

```sh
go test -v ./...
```

Run tests for a specific package:

```sh
go test -v ./internal/app/...
go test -v ./internal/vault/...
```

Run tests with race detection:

```sh
go test -race ./...
```

Generate a coverage report:

```sh
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Lint / Vet

Run the built-in Go static analysis:

```sh
go vet ./...
```

If you have [golangci-lint](https://golangci-lint.run/) installed:

```sh
golangci-lint run
```

## Project Structure

```
vaultui/
├── main.go                 # Entrypoint
├── cmd/
│   └── root.go             # Cobra CLI, flag parsing, config loading
├── internal/
│   ├── app/                # Top-level Bubble Tea model, keybindings
│   ├── ui/
│   │   ├── styles/         # Lipgloss theme and style definitions
│   │   ├── components/     # Reusable UI components (table, breadcrumb, statusbar, etc.)
│   │   └── views/          # Screen views (dashboard, engines, path browser, etc.)
│   └── vault/              # Vault API client wrapper, caching, operations
├── config/                 # Viper config loader
├── pkg/
│   ├── clipboard/          # Cross-platform clipboard support
│   └── format/             # Duration, time, JSON formatters
└── docs/                   # Project documentation
```

## Local Vault for Testing

Spin up a dev-mode Vault server for local testing:

```sh
vault server -dev -dev-root-token-id=root
```

Then in another terminal:

```sh
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
go run .
```
