# Development Guide

## Prerequisites

- **Go 1.25+** — [install instructions](https://go.dev/doc/install)
- **golangci-lint** — [install instructions](https://golangci-lint.run/welcome/install/)
- A terminal emulator with true-color support (iTerm2, Ghostty, Alacritty, kitty, etc.)
- (Optional) Docker & Docker Compose for the local Vault dev environment

## Getting Started

Clone the repo and install dependencies:

```sh
git clone https://github.com/miladbeigi/vaultui.git && cd vaultui
go mod download
```

## Build

```sh
make build
```

This injects version, commit, and date via `-ldflags`. You can also build manually:

```sh
go build -o vaultui .
```

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
| `--namespace` | Vault namespace (Enterprise) | `VAULT_NAMESPACE` |
| `--config` | Path to config file (default `~/.vaultui.yaml`) | — |
| `--auth-method` | Auth method: `token`, `userpass`, `approle` | — |
| `--auth-mount` | Custom mount path for auth method | — |
| `--username` | Username for userpass auth | — |
| `--password` | Password for userpass auth | — |
| `--role-id` | Role ID for AppRole auth | — |
| `--secret-id` | Secret ID for AppRole auth | — |

Examples:

```sh
# Using flags
go run . --vault-addr=http://127.0.0.1:8200 --token=root

# Using environment variables
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
go run .

# With userpass auth
go run . --vault-addr=http://127.0.0.1:8200 --auth-method=userpass --username=testuser --password=testpass

# With AppRole auth
go run . --vault-addr=http://127.0.0.1:8200 --auth-method=approle --role-id=xxx --secret-id=yyy

# Using a custom config file
go run . --config=./my-config.yaml
```

## Local CI

The project includes a `Makefile` that mirrors the GitHub Actions CI pipeline. Run all checks locally before pushing:

```sh
make ci
```

Individual targets:

| Target | Description |
|--------|-------------|
| `make fmt` | Check `gofmt` formatting |
| `make vet` | Run `go vet` static analysis |
| `make lint` | Run `golangci-lint` |
| `make test` | Run `go test ./...` |
| `make build` | Build the binary with version ldflags |
| `make tidy` | Check `go.mod` / `go.sum` tidiness |
| `make clean` | Remove the built binary |
| `make release VERSION_TAG=v0.2.0` | Tag and push a release |

## Test

Run all tests:

```sh
make test
# or
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
go test -v ./internal/ui/views/...
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

## Project Structure

```
vaultui/
├── main.go                       # Entrypoint
├── cmd/
│   ├── root.go                   # Cobra CLI, flag parsing, config loading
│   ├── get.go                    # Headless `vaultui get` subcommand
│   └── version.go                # `vaultui version` subcommand
├── internal/
│   ├── app/                      # Top-level Bubble Tea model, router, keybindings
│   ├── clipboard/                # Cross-platform clipboard with auto-clear
│   ├── config/                   # YAML config loader (~/.vaultui.yaml)
│   ├── version/                  # Build-time version info (ldflags)
│   ├── ui/
│   │   ├── styles/               # Lipgloss color palette and style definitions
│   │   ├── components/           # Reusable UI components (table, breadcrumb)
│   │   └── views/                # Screen views (dashboard, engines, secrets, etc.)
│   └── vault/                    # Vault API client, caching, auth, engine methods
├── scripts/
│   └── seed.sh                   # Test data seed script for local dev
├── docs/
│   ├── DESIGN.md                 # Design document and roadmap
│   └── development.md            # This file
├── .goreleaser.yaml              # GoReleaser cross-platform release config
├── CHANGELOG.md                  # Release changelog
├── Makefile                      # Local CI runner and build with version injection
├── Dockerfile                    # Multi-stage build from local source
└── docker-compose.yml            # Local Vault dev environment
```

## Local Vault for Testing

### Docker Compose (recommended)

The project includes a `docker-compose.yml` that starts a dev Vault server and seeds it with realistic test data:

```sh
# Start Vault and seed test data
docker compose up -d

# Run vaultui against it
VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root go run .

# Tear down when done
docker compose down
```

### Seed Data

The seed script (`scripts/seed.sh`) populates:

**KV v2 secrets** (with multiple versions for diff testing):

```
secret/
├── apps/
│   ├── myapp/
│   │   ├── config       (db credentials — 3 versions for history/diff)
│   │   ├── database     (connection string)
│   │   └── api-keys     (stripe, sendgrid)
│   └── billing/
│       └── config       (api url, timeout)
└── infra/
    ├── tls/
    │   └── wildcard     (cert + key)
    ├── aws              (access keys)
    └── ssh/
        └── deploy-key   (SSH keypair)
```

**Auth methods:**

- `userpass` — user `testuser` / password `testpass` (policies: `base-read`, `app-secrets`)
- `approle` — role `test-role` (policies: `base-read`, `infra-secrets`)
- `ldap` — enabled but unconfigured (for UI display)

**Policies:**

| Policy | Access |
|--------|--------|
| `base-read` | `sys/mounts`, `sys/auth`, `sys/policies`, `sys/health`, `secret/metadata` |
| `admin` | Full `*` access |
| `app-secrets` | Read `secret/data/apps/*`, list `secret/metadata/apps/*` |
| `infra-secrets` | Read `secret/data/infra/*`, list `secret/metadata/infra/*` |

**PKI engine:**

- Root CA ("Test Root CA")
- Role `test-role` (allowed domain: `test.example.com`)
- Issued certificate (`app1.test.example.com`)

**Transit engine:**

- Key `my-app-key` (default type)
- Key `payment-key` (aes256-gcm96)

**Identity:**

- Entities: `test-user-entity`, `admin-entity`
- Groups: `dev-team`, `ops-team`

### Testing Auth Methods

After `docker compose up -d`:

```sh
# Root token (full access)
VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root go run .

# Userpass (restricted to app secrets)
go run . --vault-addr=http://127.0.0.1:8200 \
  --auth-method=userpass --username=testuser --password=testpass

# AppRole (restricted to infra secrets)
ROLE_ID=$(docker compose exec vault vault read -field=role_id auth/approle/role/test-role/role-id)
SECRET_ID=$(docker compose exec vault vault write -f -field=secret_id auth/approle/role/test-role/secret-id)
go run . --vault-addr=http://127.0.0.1:8200 \
  --auth-method=approle --role-id=$ROLE_ID --secret-id=$SECRET_ID
```

### Manual (without Docker)

If you have the Vault CLI installed locally:

```sh
vault server -dev -dev-root-token-id=root
```

Then in another terminal:

```sh
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
go run .
```

## Headless Mode

The `vaultui get` subcommand outputs JSON for scripting:

```sh
vaultui get health                      # Vault health status
vaultui get engines                     # List secret engines
vaultui get auth                        # List auth methods
vaultui get policies                    # List policies
vaultui get secret secret/apps/myapp/config  # Read a secret
vaultui get policy admin                # Read a policy body
```

## Releasing

The project uses [GoReleaser](https://goreleaser.com/) and GitHub Actions for automated releases.

### Versioning

We follow [Semantic Versioning](https://semver.org/) (`vMAJOR.MINOR.PATCH`):

- **PATCH** — bug fixes, doc updates
- **MINOR** — new features (new engine browser, new command, etc.)
- **MAJOR** — breaking changes (config format change, removed flags, etc.)

### Creating a Release

1. Update `CHANGELOG.md` — move items from `[Unreleased]` to a new version section
2. Commit the changelog update
3. Tag and push:

```sh
make release VERSION_TAG=v0.2.0
```

This creates an annotated git tag and pushes it. GitHub Actions will automatically:

- Build cross-platform binaries (linux/darwin, amd64/arm64)
- Generate checksums
- Build and push Docker images to `ghcr.io/miladbeigi/vaultui`
- Create a GitHub Release with the binaries attached

### Version Info

The binary embeds version metadata via `-ldflags` at build time:

```sh
vaultui version
# vaultui v0.1.0 (commit: abc1234, built: 2026-02-16T12:00:00Z, darwin/arm64)
```

The `make build` target injects these automatically. GoReleaser does the same for release builds.

## Coding Conventions

- Every new Vault API method goes in `internal/vault/` and includes caching where appropriate.
- Every new TUI view implements the `ui.View` interface and goes in `internal/ui/views/`.
- Views are wired in `internal/app/app.go` (router push, commands, jump keys).
- New secret engines are routed from `engines.go` via type check in `handleEnter`.
- Unit tests sit next to the code they test (`_test.go` suffix).
- Run `make ci` before committing to catch formatting, lint, and test issues.
- Use `gofmt` formatting — no exceptions.
