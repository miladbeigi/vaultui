# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.6.0] - 2026-03-15

### Added

- AWS secrets engine browser with tabbed view (Roles, Config, Leases)
- Role detail view showing credential type, policy ARNs, role ARNs, policy document, STS TTLs
- Config viewer showing access key, region, IAM/STS endpoints, max retries, lease config
- Leases browser with role, short ID, TTL, and issue time columns
- Lease detail view showing full lease ID, TTL, issue/expire time, renewable status
- Jump shortcut `8` and `:aws` command for quick access to AWS engine
- LocalStack container in Docker Compose for AWS engine testing
- Seed data: AWS engine config pointing at LocalStack, 3 roles (iam_user, assumed_role, federation_token)
- Headless DB scripting added to deferred items in DESIGN.md

## [0.5.0] - 2026-03-15

### Added

- Database secrets engine browser with tabbed view (Connections, Roles, Static Roles)
- Connection detail view showing plugin, connection URL, allowed roles, verify status
- Dynamic role detail view with creation/revocation statements and TTL config
- Static role detail view with username, rotation period, and last rotation timestamp
- Jump shortcut `7` and `:db` command for quick access to database engine
- PostgreSQL container in Docker Compose for database engine seed data
- Seed data: database connection, 3 dynamic roles, 1 static role
- Colored CI output (green on success, red on failure)

## [0.4.0] - 2026-03-10

### Fixed

- Secret Engines PATH column now auto-sizes to fit the longest engine path
- Reduced cyclomatic complexity of `Update()` for better Go Report Card score

## [0.3.1] - 2026-03-10

### Fixed

- Secret Engines PATH column now auto-sizes to fit the longest engine path instead of truncating with `...`

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
