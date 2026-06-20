# Copilot instructions for VaultUI

Purpose: quick, repository-specific guidance for future Copilot sessions working on this project.

----

## Build / test / lint (commands and single-test examples)

- Local quick build: `make build` (injects version metadata via -ldflags) or `go build -o vaultui .`
- Run from source: `go run .`
- Full test suite: `make test` or `go test ./...`
- Run a single package tests (verbose): `go test -v ./internal/vault/...`
- Run a single test function in a package:
  - `go test -v ./internal/ui/views -run '^TestMyThing$'`
  - or `go test -v ./internal/ui/views -run TestMyThing`
- Race detection: `go test -race ./...`
- Coverage: `go test -coverprofile=coverage.out ./...` then `go tool cover -html=coverage.out -o coverage.html`

- Format check: `make fmt` (or `gofmt -l .` to list diffs; `gofmt -w .` to fix)
- Vet: `make vet` (runs `go vet ./...`)
- Lint: `make lint` (uses `golangci-lint run`)
  - Install linter: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Local CI mirror (runs fmt, vet, lint, test, build, tidy): `make ci`
- Module tidiness: `make tidy` (runs `go mod tidy` and fails if go.mod/go.sum changed)

CI (GitHub Actions): `go mod download`, `go test ./...`, `golangci-lint run`, `go vet`, `go build`, `go mod tidy` (see .github/workflows/ci.yml)

----

## High-level architecture (big picture)

- Language: Go. Entrypoint: `main.go` + Cobra CLI under `cmd/`.
- UI layer: `internal/app/` (Bubble Tea model, router, keybindings) + `internal/ui/`:
  - `internal/ui/views/` — screen implementations (dashboard, engines, secrets, etc.)
  - `internal/ui/components/` — reusable components (table, breadcrumb)
  - `internal/ui/styles/` — Lipgloss theme and named style variables
- Vault client layer: `internal/vault/` — Vault API wrappers, auth flows, caching, renewer, engine helpers
- Other helpers: `internal/clipboard/` (auto-clear clipboard), `internal/config/` (~/.vaultui.yaml multi-context loader), `internal/version/` (ldflags-injected metadata)
- Routing & wiring:
  - Views implement a UI interface (Init/Update/View/Title/KeyHints) and are composed by the router in `internal/app/app.go`.
  - Command palette handling and jump keys are wired in `internal/app/app.go` (`executeCommand()`) and `internal/app/keys.go`.
- Headless mode: `cmd/get.go` implements `vaultui get` that emits JSON for scripting.
- Local development: `docker-compose.yml` + `scripts/seed.sh` provide a seeded Vault for deterministic testing.

----

## Key conventions & repository-specific patterns

- Go toolchain: Require Go 1.25+. Follow `gofmt` formatting (Makefile + CI enforce it).
- Tests: unit tests live next to code (`*_test.go`). Use package-level helpers (e.g., `newTestClient(t)`) for view/API tests.
- Vault API code: put Vault methods in `internal/vault/`, return typed structs (avoid raw map[string]interface{}), and use the provided cache for list-heavy operations (`c.cache.Get/Set`). Wrap errors with `%w` for context.
- TUI views:
  - Each view implements the `ui.View` interface (Init, Update, View, Title, KeyHints).
  - Use `components.Table` for list UIs and `components.Breadcrumb` for path-based navigation.
  - Style constants live in `internal/ui/styles/theme.go` — don't inline Lipgloss colors.
  - Navigation is stack-based (Push/Pop/ResetToRoot) managed by `internal/app/router.go`.
  - Add commands by updating `executeCommand()` in `internal/app/app.go` and add jump keys in `internal/app/keys.go` if needed.
- Seed data: update `scripts/seed.sh` with test fixtures for new features; run `docker compose up -d` and verify seed via `docker compose logs seed | tail -1`.
- Pre-commit / pre-push workflow: run `make ci` locally; CI runs the same checks.
- Releases: `make release VERSION_TAG=vX.Y.Z` tags and relies on GoReleaser via GitHub Actions to publish cross builds and images.

----

## Where to look for authoritative details

- Development guide: `docs/development.md` (detailed run/build/test examples)
- Design / roadmap: `docs/DESIGN.md`
- CI: `.github/workflows/ci.yml` and `.goreleaser.yaml`
- Lint config: `.golangci.yml`
- Local dev seed: `docker-compose.yml` + `scripts/seed.sh`
- Project structure summary: top-level README
- Editor/test helpers: check `internal/*_test.go` for patterns used in tests

----

## Other assistant rules discovered and included

- Cursor rules present under `.cursor/` — they contain extra, repository-specific development lifecycle notes (setup, run, test example commands). Important items from those rules are incorporated above (go version, golangci-lint install, seed checks, `make ci` pre-commit policy).

----

If anything here should be more detailed (example single-test patterns, specific files to open for wiring new views, or adding CI matrix entries), say which area to expand.
