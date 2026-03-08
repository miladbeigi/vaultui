# VaultUI

A **k9s-inspired terminal UI** for [HashiCorp Vault](https://www.vaultproject.io/).
Browse secrets, auth methods, policies, and more — without leaving your terminal.

<p align="center">
  <img src="docs/images/dashboard.png" alt="VaultUI Dashboard" width="800">
</p>

## Features

- **Dashboard** — seal status, version, HA node count, storage backend, resource counts, quick nav
- **Secret Engines browser** — list all KV, PKI, Transit, Identity, and other mounts
- **KV v2 path browser** — drill into directories, read secrets, copy values to clipboard
- **KV v2 version history** — browse version history, view specific versions, diff between versions
- **Policies browser** — list and view ACL policy bodies with syntax highlighting
- **Auth Methods browser** — list all enabled auth methods with type, accessor, description
- **PKI engine browser** — browse certificates and roles, view PEM details with copy
- **Transit engine browser** — browse encryption keys and view key properties
- **Identity browser** — browse entities and groups with tab switching and detail views
- **Context switching** — manage multiple Vault connections via `~/.vaultui.yaml`
- **Multiple auth methods** — token, userpass, and AppRole authentication
- **Command palette** — `:secrets`, `:auth`, `:policies`, `:pki`, `:transit`, `:identity`, `:ctx`
- **Configurable keybindings** — override defaults via config file
- **Clipboard with auto-clear** — copied secrets are cleared after 30 seconds
- **Responsive layout** — compact header for narrow terminals, minimum size guard
- **Error overlay** — contextual troubleshooting hints for common Vault errors
- **Headless mode** — `vaultui get` subcommand for JSON output in scripts
- **Vim-style navigation** — `j`/`k`, `g`/`G`, `Ctrl+D`/`Ctrl+U`, `Enter`, `Esc`
- **Stack-based routing** — every view preserves scroll position and state

## Quick Start

### Prerequisites

- Go 1.25+
- A running Vault instance (or use the bundled dev setup below)

### Install

```bash
go install github.com/miladbeigi/vaultui@latest
```

Or build from source:

```bash
git clone https://github.com/miladbeigi/vaultui.git
cd vaultui
go build -o vaultui .
```

### Run

```bash
# Uses VAULT_ADDR and VAULT_TOKEN from environment
vaultui

# Or pass flags explicitly
vaultui --vault-addr https://vault.example.com --token s.xxxxx

# With a namespace (Enterprise)
vaultui --namespace admin

# With userpass auth
vaultui --auth-method userpass --username admin --password secret

# With AppRole auth
vaultui --auth-method approle --role-id xxx --secret-id yyy
```

### Local Dev Environment

Spin up a local Vault with seeded test data:

```bash
docker compose up -d
VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=root vaultui
```

The seed data includes KV v2 secrets (with multiple versions), policies, userpass/AppRole auth, PKI certs, Transit keys, and Identity entities/groups.

### Headless / Scripting Mode

```bash
# Get Vault health as JSON
vaultui get health

# List secret engines
vaultui get engines

# Read a secret
vaultui get secret secret/apps/myapp/config | jq .db_host
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Open / drill in |
| `Esc` / `←` | Go back |
| `g` / `Home` | Jump to top |
| `G` / `End` | Jump to bottom |
| `Ctrl+D` | Page down |
| `Ctrl+U` | Page up |
| `1` | Secret Engines |
| `2` | Auth Methods |
| `3` | Policies |
| `4` | Identity |
| `5` | PKI |
| `6` | Transit |
| `Tab` | Switch tab (where applicable) |
| `:` | Command palette |
| `c` | Copy selected value |
| `C` | Copy secret as JSON |
| `v` | Version history (KV v2 detail) |
| `d` | Diff versions (version history) |
| `q` | Quit |

## Command Palette

Press `:` to open, then type a command:

| Command | Action |
|---------|--------|
| `:secrets` | Go to Secret Engines |
| `:auth` | Go to Auth Methods |
| `:policies` | Go to Policies |
| `:identity` | Go to Identity |
| `:pki` | Go to PKI engine |
| `:transit` | Go to Transit engine |
| `:ctx` | Switch Vault context |
| `:dash` | Go to Dashboard |
| `:q` / `:quit` | Quit |

## Configuration

VaultUI reads configuration from multiple sources (in priority order):

1. CLI flags (`--vault-addr`, `--token`, `--namespace`, `--auth-method`, etc.)
2. Environment variables (`VAULT_ADDR`, `VAULT_TOKEN`, `VAULT_NAMESPACE`)
3. Config file (`~/.vaultui.yaml`)
4. Token file (`~/.vault-token`)

### Multi-context config (`~/.vaultui.yaml`)

```yaml
current_context: dev

contexts:
  - name: dev
    address: http://127.0.0.1:8200
    token: root

  - name: staging
    address: https://vault.staging.example.com
    auth:
      method: userpass
      username: admin
      password: secret

  - name: prod
    address: https://vault.example.com
    auth:
      method: approle
      role_id: "xxx"
      secret_id: "yyy"

settings:
  clipboard_timeout: 30
  keybindings:
    quit: "q"
    back: "esc,left"
```

Switch contexts inside the TUI with `:ctx` or by editing the config file.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Language | [Go](https://go.dev) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Vault Client | [vault/api](https://pkg.go.dev/github.com/hashicorp/vault/api) |
| CLI | [Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper) |

## Project Structure

```
├── cmd/                        # CLI entrypoint (Cobra)
│   ├── root.go                 # Main command, flags, config loading
│   └── get.go                  # Headless `vaultui get` subcommand
├── internal/
│   ├── app/                    # Bubble Tea model, router, keybindings
│   ├── clipboard/              # Cross-platform clipboard with auto-clear
│   ├── config/                 # YAML config loader (~/.vaultui.yaml)
│   ├── ui/
│   │   ├── components/         # Reusable table, breadcrumb components
│   │   ├── styles/             # Lipgloss color palette and styled components
│   │   └── views/              # All TUI views (dashboard, engines, secrets, etc.)
│   └── vault/                  # Vault API client, caching, auth, engine methods
├── scripts/
│   └── seed.sh                 # Test data for local Vault dev environment
├── docs/
│   ├── DESIGN.md               # Detailed design document and roadmap
│   └── development.md          # Development guide
├── Makefile                    # Local CI: make ci, make test, make lint, etc.
├── Dockerfile                  # Multi-stage build from local source
└── docker-compose.yml          # Local Vault dev environment with seed data
```

## Development

See [docs/development.md](docs/development.md) for the full development guide.

Quick start:

```bash
docker compose up -d                    # Start Vault with seed data
VAULT_ADDR=http://127.0.0.1:8200 \
VAULT_TOKEN=root go run .               # Run from source

make ci                                 # Run all CI checks locally
make test                               # Run tests only
```

## License

MIT
