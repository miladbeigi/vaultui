# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.0] - 2026-03-10

### Added

- Command palette auto-complete with Tab/Shift-Tab cycling and fuzzy matching
- Secret jump-to-path via `:go <path>` command for direct navigation to any secret
- Inline completion hints shown as user types in command palette
- Space key support in command palette input
- CI, release, Go Report Card, and license badges in README
- MIT license file

## [0.2.0] - 2026-03-10

### Added

- Token inspector view (`:token` command) showing policies, TTL, accessor, entity ID, auth path, token type, and metadata with refresh support
- Secret engine dashboard (`d` key in Secret Engines view) showing per-engine mount config, default/max lease TTL, UUID, accessor, seal wrap, and tuning parameters
- Headless `vaultui get token` for scripting token inspection
- UI validation step in development workflow (user confirms changes before commit)

## [0.1.0] - 2026-02-16

### Added

- Dashboard with seal status, version, HA node count, storage backend, and resource counts
- Secret Engines browser with routing to KV, PKI, Transit, and Identity engines
- KV v2 path browser with directory navigation and secret detail view
- KV v2 version history browser with version-to-version diff
- Policies browser with ACL policy body viewer
- Auth Methods browser listing all enabled methods
- PKI engine browser for certificates and roles with PEM detail and copy
- Transit engine browser for encryption keys with key properties view
- Identity browser for entities and groups with tab switching and detail views
- Context switching for multiple Vault connections via `~/.vaultui.yaml`
- Multiple authentication methods: token, userpass, AppRole
- Command palette with `:secrets`, `:auth`, `:policies`, `:pki`, `:transit`, `:identity`, `:ctx`, `:dash`
- Jump shortcuts: `1`-`6` for quick navigation to major views
- Configurable keybindings via config file
- Clipboard copy with 30-second auto-clear for secrets
- Error overlay with contextual troubleshooting hints
- Headless mode via `vaultui get` subcommand for JSON output
- Responsive layout with compact header and minimum terminal size guard
- Stack-based router preserving scroll position and view state
- Background token renewal
- Thread-safe TTL-based API response cache
- Reusable breadcrumb navigation component
- `vaultui version` subcommand with build metadata
- GoReleaser-based release pipeline with cross-platform binaries and Docker images
- Local CI via `Makefile` (`make ci`)
- GitHub Actions CI (test, lint, vet, build, tidy)
- Docker Compose local development environment with comprehensive seed data
