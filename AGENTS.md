# AGENTS.md

## Project: VaultUI

A k9s-inspired TUI for HashiCorp Vault, built with Go, BubbleTea, and Cobra.

---

## Build & CI

Always use `make` targets — **never** use `go install` directly.

| Command | Purpose |
|---|---|
| `make build` | Build binary with `-ldflags` injecting `Version`/`Commit`/`Date` |
| `make ci` | Run the full pipeline: fmt, vet, lint, test, build, tidy |
| `make fmt` | Check formatting |
| `make vet` | Run `go vet ./...` |
| `make lint` | Run golangci-lint |
| `make test` | Run all tests |
| `make tidy` | Run `go mod tidy` |
| `make clean` | Remove the binary |

> **`go install github.com/miladbeigi/vaultui@latest`** builds without `-ldflags`, so the version always shows `dev`. Use `make build` to get a properly versioned binary.

## Version System

`internal/version/version.go` exports `Version`, `Commit`, `Date` as package-level vars injected via `-ldflags`:

```makefile
LDFLAGS = -s -w \
          -X github.com/miladbeigi/vaultui/internal/version.Version=$(VERSION) \
          -X github.com/miladbeigi/vaultui/internal/version.Commit=$(COMMIT) \
          -X github.com/miladbeigi/vaultui/internal/version.Date=$(DATE)
```

- `Version` is set by `git describe --tags --always --dirty`
- `Commit` is `git rev-parse --short HEAD`
- `Date` is `date -u +%Y-%m-%dT%H:%M:%SZ`

`Banner()` draws a unicode box with these values. `String()` embeds `Banner()` and adds runtime info.

## Common Pitfalls

1. **`fmt.Sprintf` placeholder count must match arg count** — always count `%s`/`%d` placeholders and verify against the argument list. Vet catches this but only *after* commit if skipped locally.
2. **Box width** — `contentWidth` in `version.go` must be wide enough for the longest possible `Version` string (e.g. `v0.8.0-1-g67b20f2-dirty`). Set `contentWidth = 52` as a safe minimum.
3. **`make ci` before committing** — run the full pipeline locally before every commit to avoid CI feedback loops.

## Code Structure

```
cmd/
  root.go          # Cobra root command
  version.go       # `vaultui version` subcommand
internal/
  version/         # Version banner logic
  app/             # BubbleTea app model
  config/          # YAML config loading
  vault/           # Vault API client
  ui/              # TUI views, components, styles
```
