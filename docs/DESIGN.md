# VaultUI — A k9s-inspired TUI for HashiCorp Vault

## Vision

A keyboard-driven terminal UI for browsing, inspecting, and managing HashiCorp Vault. The UX takes heavy inspiration from [k9s](https://k9scli.io/): fast navigation, vim-like keybindings, a command palette, breadcrumb context, and real-time data views — all without leaving the terminal.

---

## Technology Stack

| Layer            | Choice                                                        | Rationale                                                                                       |
| ---------------- | ------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Language         | **Go**                                                        | Same ecosystem as Vault; excellent Vault SDK; single binary distribution                        |
| TUI Framework    | **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**  | Elm-architecture, composable, well-maintained, growing ecosystem                                |
| Styling          | **[Lipgloss](https://github.com/charmbracelet/lipgloss)**     | Declarative terminal styling; pairs naturally with Bubble Tea                                   |
| Table/List       | **[Bubbles](https://github.com/charmbracelet/bubbles)**       | Ready-made table, list, text-input, viewport, spinner components                                |
| Vault Client     | **[vault/api](https://pkg.go.dev/github.com/hashicorp/vault/api)** | Official Go client; handles auth, retries, TLS                                             |
| Config           | **[Viper](https://github.com/spf13/viper)**                   | Read `VAULT_ADDR`, `VAULT_TOKEN`, config files, and CLI flags uniformly                        |
| CLI Entrypoint   | **[Cobra](https://github.com/spf13/cobra)**                   | Standard Go CLI framework; handles `--vault-addr`, `--token`, `--namespace` flags               |

---

## Core Concepts (mapped from k9s)

| k9s Concept          | VaultUI Equivalent            | Description                                                                 |
| -------------------- | ----------------------------- | --------------------------------------------------------------------------- |
| Cluster context      | **Vault connection**          | `VAULT_ADDR` + auth token + namespace                                       |
| Namespace            | **Vault namespace**           | Enterprise namespaces; CE users see `root` only                             |
| Resource type        | **Resource kind**             | Secrets engines, auth methods, policies, leases, identity, sys              |
| Resource list        | **Resource browser**          | Table of items for the selected resource kind                               |
| Describe / YAML view | **Detail pane**               | JSON or table view of a secret, policy body, mount config, etc.             |
| Port-forward / shell | **Actions**                   | Copy secret, wrap/unwrap, renew lease, enable/disable engine                |
| `:` command bar      | **Command palette**           | Quick-jump: `:secrets`, `:policies`, `:auth`, `:leases`, `:sys`            |
| `/` filter           | **Fuzzy filter**              | Filter rows in any table view                                               |
| Pulse view           | **Dashboard**                 | Health, seal status, version, HA mode, mount counts                         |

---

## Screens & Views

### 1. Dashboard (home)

```
┌──────────────────────────────────────────────────────────────────┐
│  VaultUI  ◆  https://vault.example.com  ◆  ns: root            │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Status     : unsealed ✔        HA Mode  : active              │
│   Version    : 1.15.4            Cluster  : vault-cluster-xyz   │
│   Seal Type  : shamir            Storage  : raft                │
│                                                                  │
│   Secret Engines : 12            Auth Methods : 4               │
│   Policies       : 23            Active Leases: 158             │
│                                                                  │
│  ┌─ Quick Nav ──────────────────────────────────────────────┐   │
│  │  [1] Secret Engines   [2] Auth Methods   [3] Policies    │   │
│  │  [4] Leases           [5] Identity       [6] Sys/Config  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  : command  / filter  ? help  q quit                             │
└──────────────────────────────────────────────────────────────────┘
```

**Data sources:** `GET /v1/sys/health`, `GET /v1/sys/mounts`, `GET /v1/sys/auth`, `GET /v1/sys/policies/acl`

---

### 2. Secret Engines Browser

```
┌──────────────────────────────────────────────────────────────────┐
│  Secret Engines  ◆  ns: root                   Filter: ________ │
├──────────────────────────────────────────────────────────────────┤
│  PATH            TYPE        VERSION   DESCRIPTION               │
│  ────────────    ────────    ───────   ───────────               │
│▸ secret/         kv          v2        Key/Value store           │
│  pki/            pki         -         PKI certificates          │
│  transit/        transit     -         Encryption as a service   │
│  database/       database    -         Dynamic DB creds          │
│  aws/            aws         -         AWS dynamic creds         │
│  ssh/            ssh         -         SSH certs                 │
│  cubbyhole/      cubbyhole   -         Per-token private store   │
│  sys/            system      -         System backend            │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ↑↓ navigate  ⏎ browse  d describe  e enable  D disable  / filt │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `GET /v1/sys/mounts`

**Actions:**
- `Enter` → drill into the engine's path (goes to Path Browser)
- `d` → show mount configuration detail pane
- `e` → enable a new secret engine (form)
- `D` → disable engine (with confirmation)

---

### 3. Path Browser (KV & generic)

This is the heart of the TUI — a filesystem-like browser for secrets.

```
┌──────────────────────────────────────────────────────────────────┐
│  secret/ ▸ apps/ ▸ myapp/               Filter: ____________    │
├──────────────────────────────────────────────────────────────────┤
│  NAME              TYPE        UPDATED                           │
│  ────              ────        ───────                           │
│  📁 production/    dir         -                                 │
│  📁 staging/       dir         -                                 │
│  📄 config         secret      2025-12-01 14:32 UTC             │
│  📄 database       secret      2025-11-28 09:15 UTC             │
│  📄 api-keys       secret      2025-10-03 22:41 UTC             │
│                                                                  │
│                                                                  │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ⏎ open  ← back  n new  x delete  y copy-path  / filter        │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `LIST /v1/secret/metadata/apps/myapp/`

**Behavior:**
- Directories (keys ending with `/`) → drill deeper
- Leaf secrets → open Secret Detail View
- Breadcrumb trail at top shows full path and is clickable via number keys

---

### 4. Secret Detail View

```
┌──────────────────────────────────────────────────────────────────┐
│  secret/apps/myapp/config  (v3)        ◆  KV v2                 │
├──────────────────────────────────────────────────────────────────┤
│  KEY                VALUE                                        │
│  ───                ─────                                        │
│  db_host           db.internal.example.com                       │
│  db_port           5432                                          │
│  db_name           myapp_production                              │
│  db_user           myapp_svc                                     │
│  db_password        ●●●●●●●●●●●●  [r to reveal]                │
│  api_endpoint      https://api.example.com/v2                    │
│  log_level         info                                          │
│                                                                  │
│ ─── Metadata ──────────────────────────────────────────────────  │
│  Created   : 2025-10-01 08:00 UTC                                │
│  Updated   : 2025-12-01 14:32 UTC                                │
│  Version   : 3 (max: 10)                                        │
│  Delete ver: not destroyed                                       │
├──────────────────────────────────────────────────────────────────┤
│  r reveal  c copy-value  C copy-json  v versions  e edit  ← bk  │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `GET /v1/secret/data/apps/myapp/config`

**Key features:**
- Values are **masked by default** — press `r` to reveal a field, `R` to reveal all
- `c` copies the selected field value to clipboard
- `C` copies the entire secret as JSON
- `v` opens version history (KV v2)
- `e` opens an editor overlay to modify the secret

---

### 5. Secret Version History (KV v2)

```
┌──────────────────────────────────────────────────────────────────┐
│  Versions: secret/apps/myapp/config                              │
├──────────────────────────────────────────────────────────────────┤
│  VER   CREATED                  DESTROYED   DELETED              │
│  ───   ───────                  ─────────   ───────              │
│▸  3    2025-12-01 14:32 UTC     no          no                   │
│   2    2025-11-15 10:00 UTC     no          no                   │
│   1    2025-10-01 08:00 UTC     no          yes (soft)           │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ⏎ view  f diff  u undelete  X destroy  ← back                  │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `GET /v1/secret/metadata/apps/myapp/config`

---

### 6. Diff View (between secret versions)

```
┌──────────────────────────────────────────────────────────────────┐
│  Diff: secret/apps/myapp/config  v2 → v3                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│    db_host          db.internal.example.com    (unchanged)       │
│    db_port          5432                       (unchanged)       │
│  - log_level        debug                                        │
│  + log_level        info                                         │
│  + api_endpoint     https://api.example.com/v2 (added)           │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ← back                                                          │
└──────────────────────────────────────────────────────────────────┘
```

---

### 7. Auth Methods Browser

```
┌──────────────────────────────────────────────────────────────────┐
│  Auth Methods  ◆  ns: root                  Filter: ________     │
├──────────────────────────────────────────────────────────────────┤
│  PATH             TYPE          DESCRIPTION                      │
│  ────             ────          ───────────                      │
│▸ token/           token         Token-based auth                 │
│  approle/         approle       AppRole auth                     │
│  oidc/            oidc          OIDC/JWT auth                    │
│  userpass/        userpass      Username/Password auth            │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ⏎ details  d describe  e enable  D disable  / filter            │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `GET /v1/sys/auth`

---

### 8. Policies Browser

```
┌──────────────────────────────────────────────────────────────────┐
│  Policies  ◆  ns: root                      Filter: ________    │
├──────────────────────────────────────────────────────────────────┤
│  NAME                TYPE                                        │
│  ────                ────                                        │
│▸ default             acl                                         │
│  admin               acl                                         │
│  app-readonly        acl                                         │
│  ci-deploy           acl                                         │
│  pki-admin           acl                                         │
│  root                root                                        │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ⏎ view  e edit  n new  x delete  / filter                      │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `GET /v1/sys/policies/acl`

**Enter → Policy Detail:**

```
┌──────────────────────────────────────────────────────────────────┐
│  Policy: app-readonly                                            │
├──────────────────────────────────────────────────────────────────┤
│  path "secret/data/apps/*" {                                     │
│    capabilities = ["read", "list"]                               │
│  }                                                               │
│                                                                  │
│  path "secret/metadata/apps/*" {                                 │
│    capabilities = ["read", "list"]                               │
│  }                                                               │
│                                                                  │
│  path "sys/mounts" {                                             │
│    capabilities = ["read"]                                       │
│  }                                                               │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  e edit  C copy  ← back                                          │
└──────────────────────────────────────────────────────────────────┘
```

---

### 9. Leases Browser

```
┌──────────────────────────────────────────────────────────────────┐
│  Active Leases  ◆  ns: root                Filter: ________     │
├──────────────────────────────────────────────────────────────────┤
│  LEASE ID (short)        PREFIX             TTL       RENEWABLE  │
│  ───────────────         ──────             ───       ─────────  │
│▸ aG3x...f9               database/creds    1h23m     yes        │
│  kL9q...2b               aws/sts           42m       no         │
│  pP7w...cc               database/creds    3h01m     yes        │
│  nN2r...a1               pki/issue         719h      no         │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│  ⏎ details  R renew  X revoke  / filter                          │
└──────────────────────────────────────────────────────────────────┘
```

**Data source:** `LIST /v1/sys/leases/lookup/`, `PUT /v1/sys/leases/lookup`

---

## Navigation & Keybindings

### Global Keys

| Key       | Action                        |
| --------- | ----------------------------- |
| `:`       | Open command palette          |
| `/`       | Open fuzzy filter             |
| `?`       | Show help overlay             |
| `Esc`     | Close overlay / go back       |
| `q`       | Quit (with confirmation)      |
| `Ctrl+C`  | Force quit                    |
| `1-6`     | Quick-jump to top-level views |
| `Tab`     | Cycle focus between panes     |

### Navigation Keys

| Key          | Action                          |
| ------------ | ------------------------------- |
| `j` / `↓`   | Move down                       |
| `k` / `↑`   | Move up                         |
| `Enter`      | Open / drill in                 |
| `Esc` / `←`  | Go back / up one level          |
| `g` / `Home` | Jump to top                     |
| `G` / `End`  | Jump to bottom                  |
| `Ctrl+D`     | Page down                       |
| `Ctrl+U`     | Page up                         |

### Command Palette Commands

| Command       | Target View           |
| ------------- | --------------------- |
| `:secrets`    | Secret Engines        |
| `:auth`       | Auth Methods          |
| `:policies`   | Policies              |
| `:leases`     | Leases                |
| `:identity`   | Identity (entities)   |
| `:sys`        | System config         |
| `:dash`       | Dashboard             |
| `:ns <name>`  | Switch namespace      |
| `:q`          | Quit                  |

---

## Architecture

### High-Level Component Diagram

```
┌──────────────────────────────────────────────────────┐
│                    CLI (Cobra)                        │
│  Flags: --vault-addr, --token, --namespace, --config │
└──────────────────┬───────────────────────────────────┘
                   │
┌──────────────────▼───────────────────────────────────┐
│                  App Model (Bubble Tea)               │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │
│  │ Navigator │  │ Command  │  │   View Router    │   │
│  │ (stack)   │  │ Palette  │  │                  │   │
│  └──────────┘  └──────────┘  └────────┬─────────┘   │
│                                        │             │
│       ┌────────────────────────────────┼──────┐      │
│       │            │           │       │      │      │
│  ┌────▼───┐  ┌─────▼──┐  ┌────▼──┐  ┌─▼──┐  │      │
│  │Dashboard│  │Secrets │  │ Auth  │  │Pol.│  │...   │
│  │  View   │  │Browser │  │Browser│  │View│  │      │
│  └────────┘  └────────┘  └───────┘  └────┘  │      │
│                                               │      │
└───────────────────────────────────────────────┼──────┘
                                                │
┌───────────────────────────────────────────────▼──────┐
│                  Vault Client Layer                   │
│                                                      │
│  ┌────────────┐  ┌───────────┐  ┌──────────────┐    │
│  │  API Cache  │  │  Watcher  │  │  Auth Renew  │    │
│  │  (TTL-based)│  │  (polling)│  │  (background)│    │
│  └────────────┘  └───────────┘  └──────────────┘    │
│                                                      │
│  ┌──────────────────────────────────────────────┐    │
│  │         vault/api.Client (official SDK)       │    │
│  └──────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────┘
```

### Module Layout

```
vaultui/
├── main.go                     # Entrypoint
├── go.mod
├── go.sum
│
├── cmd/
│   └── root.go                 # Cobra root command, flag parsing
│
├── internal/
│   ├── app/
│   │   ├── app.go              # Top-level Bubble Tea model
│   │   ├── router.go           # View routing / navigation stack
│   │   └── keys.go             # Global keybinding definitions
│   │
│   ├── ui/
│   │   ├── styles/
│   │   │   └── theme.go        # Lipgloss style definitions, color palette
│   │   ├── components/
│   │   │   ├── table.go        # Reusable table component
│   │   │   ├── breadcrumb.go   # Path breadcrumb bar
│   │   │   ├── statusbar.go    # Bottom status/help bar
│   │   │   ├── header.go       # Top bar (connection info)
│   │   │   ├── modal.go        # Confirmation dialog / overlay
│   │   │   ├── commandbar.go   # Command palette (`:` trigger)
│   │   │   ├── filterbar.go    # Fuzzy filter (`/` trigger)
│   │   │   └── form.go         # Generic key-value editor
│   │   │
│   │   └── views/
│   │       ├── dashboard.go    # Dashboard / home screen
│   │       ├── engines.go      # Secret engines list
│   │       ├── pathbrowser.go  # Hierarchical path browser
│   │       ├── secretdetail.go # Secret key-value detail view
│   │       ├── versions.go     # KV v2 version history
│   │       ├── diff.go         # Version diff view
│   │       ├── auth.go         # Auth methods browser
│   │       ├── policies.go     # Policies list + detail
│   │       ├── policydetail.go # Policy HCL viewer
│   │       └── leases.go       # Lease browser
│   │
│   └── vault/
│       ├── client.go           # Vault API client wrapper
│       ├── cache.go            # TTL-based response cache
│       ├── secrets.go          # Secret engine operations
│       ├── auth.go             # Auth method operations
│       ├── policies.go         # Policy CRUD
│       ├── leases.go           # Lease operations
│       ├── sys.go              # sys/ endpoints (health, mounts, etc.)
│       └── watcher.go          # Background polling for live updates
│
├── config/
│   └── config.go               # Viper config loading (~/.vaultui.yaml)
│
└── pkg/
    ├── clipboard/
    │   └── clipboard.go        # Cross-platform clipboard support
    └── format/
        └── format.go           # Duration, time, JSON formatters
```

---

## Navigation Model (Stack-based)

The navigator maintains a **stack of views**, similar to a browser's history:

```
Push: Dashboard → Engines → PathBrowser("secret/") → PathBrowser("secret/apps/") → SecretDetail
Pop:  SecretDetail → PathBrowser("secret/apps/") → PathBrowser("secret/") → Engines → Dashboard
```

Each view on the stack preserves:
- Scroll position / cursor index
- Active filter text
- Any ephemeral UI state (revealed secrets, expanded sections)

This means pressing `Esc`/`←` always returns to the **exact state** you left.

---

## Vault Client Layer

### Authentication Flow

```
1. Check VAULT_TOKEN env var
2. Check ~/.vault-token file
3. Check --token CLI flag
4. If none found → show auth method picker:
   - Token (paste)
   - Userpass (form)
   - OIDC (browser redirect)
   - AppRole (role_id + secret_id form)
```

### Caching Strategy

| Resource                | Cache TTL | Rationale                              |
| ----------------------- | --------- | -------------------------------------- |
| `sys/health`            | 10s       | Changes rarely, show near-realtime     |
| `sys/mounts`            | 60s       | Mount changes are infrequent           |
| `sys/auth`              | 60s       | Same as mounts                         |
| `sys/policies/acl`      | 30s       | May change more often in CI envs       |
| `secret/metadata/...`   | 30s       | Path listings; moderate change rate    |
| `secret/data/...`       | 0 (none)  | Always fetch fresh secret data         |
| `sys/leases/lookup/...` | 10s       | Leases are time-sensitive              |

All caches are invalidated on any **write operation** to the same path.

### Error Handling

| Scenario                | Behavior                                                |
| ----------------------- | ------------------------------------------------------- |
| 403 Forbidden           | Show inline error + which policy capability is missing  |
| 404 Not Found           | Remove from list, show toast notification               |
| 429 Rate Limited        | Exponential backoff, show "rate limited" indicator      |
| 5xx Server Error        | Retry 3x with backoff, then show error overlay          |
| Network unreachable     | Show "disconnected" banner, retry in background         |
| Token expired           | Prompt re-authentication                                |

---

## Security Considerations

1. **Secret masking**: All secret values are masked (`●●●●`) by default. Reveal is per-field, per-session only.
2. **Clipboard auto-clear**: Copied secrets are cleared from clipboard after 30s (configurable).
3. **No disk persistence of secrets**: Secrets are never written to disk, config files, or logs.
4. **Token handling**: Token is held only in memory; never logged or displayed in UI.
5. **Audit trail awareness**: All operations go through the standard Vault API, so Vault's audit log captures everything.
6. **Confirm destructive actions**: Delete, destroy, disable, and revoke all require confirmation dialogs.

---

## Configuration File (`~/.vaultui.yaml`)

```yaml
# Default Vault connection
vault:
  address: https://vault.example.com
  namespace: ""               # default namespace

# UI preferences
ui:
  theme: dark                 # dark | light
  show_icons: true            # folder/file icons in path browser
  mask_secrets: true          # mask values by default
  clipboard_clear_seconds: 30 # auto-clear clipboard (0 = disabled)
  refresh_interval: 30        # background refresh interval in seconds

# Contexts (like kubeconfig contexts)
contexts:
  production:
    address: https://vault.prod.example.com
    namespace: production
  staging:
    address: https://vault.staging.example.com
    namespace: staging
  dev:
    address: http://127.0.0.1:8200
    namespace: ""

# Keybinding overrides (optional)
keys:
  reveal_secret: "r"
  copy_value: "c"
  copy_json: "C"
```

---

## Phased Implementation Plan

### Phase 1 — Foundation (MVP)
> **Goal:** Connect to Vault and browse KV secrets.

- [x] Project scaffolding (Go modules, Cobra CLI, Bubble Tea shell)
- [x] Vault client wrapper with token auth
- [x] Connection header bar (address, namespace, seal status)
- [x] Bottom status bar with contextual keybinding hints
- [x] Secret Engines list view
- [x] Path Browser for KV v2 (list directories + secrets)
- [x] Secret Detail View (basic key-value table, copy value, copy JSON)
- [x] Navigation stack (push/pop)
- [x] Command palette (`:secrets`, `:auth`, `:dash`, `:q`)
- [x] Fix layout: sticky header/status bar with fixed body region (three-part layout)
- [x] Fix table: columns should fill available width, proper frame/borders

### Phase 2 — Broader Vault Coverage
> **Goal:** Cover the main Vault resource types.

- [x] Dashboard view (health, seal type, storage, HA nodes, resource counts, quick nav)
- [x] Auth Methods browser (table view, key 2, `:auth` command)
- [x] Policies browser + HCL detail viewer (table view, key 3, `:policies`, Enter for HCL detail with copy)
- [x] KV v2 version history view (metadata API, version table, view specific versions)
- [x] Version diff view (side-by-side comparison of two secret versions)
- [x] Breadcrumb navigation component (reusable across path-based views)

### Phase 3 — Power Features
> **Goal:** Quality-of-life features for daily Vault work.

- [x] Multiple auth methods (userpass, AppRole via CLI flags and config)
- [x] Context switching (multi-Vault, multi-namespace via `:ctx` command)
- [x] Config file support (`~/.vaultui.yaml` with contexts, settings, keybindings)
- [x] Background token renewal (auto-renew at 2/3 TTL)
- [x] Response caching with smart invalidation (TTL-based, prefix invalidation)
- [x] Clipboard auto-clear timer (30s default after copying secrets)

### Phase 4 — Polish & Advanced
> **Goal:** Production-grade TUI.

- [x] Light/dark theme support (`:theme` toggle, config-driven)
- [x] Configurable keybindings (via `settings.keybindings` in config)
- [x] PKI engine browser (certs, roles, cert PEM detail with copy)
- [x] Transit engine operations (key browser, key detail with properties)
- [x] Identity entities and groups browser (tab switching, detail view)
- [x] Mouse support (scroll wheel navigation, click-to-select table rows)
- [x] Responsive layout (compact header for narrow terminals, minimum size guard)
- [x] Error overlay with troubleshooting hints (contextual advice per error type)
- [x] `--output json` flag / `vaultui get` subcommand for headless scripting

### X — Deferred
> Items deprioritised from their original phase. Will revisit when needed.

- [ ] Help overlay (`?` keybinding to show all keybindings)
- [ ] Quit confirmation (`q` smart quit — confirm or only from root view)
- [ ] Fuzzy filter on table views (`/` to filter rows in real-time)
- [ ] Command palette auto-complete / fuzzy match
- [ ] Command palette centered modal or status bar replacement layout
- [ ] Leases browser (read-only, needs dynamic secret engines for test data)
- [ ] Lease renew/revoke (write action)
- [ ] Create/edit secrets (form overlay, write action)
- [ ] Create/edit policies (embedded editor, write action)
- [ ] Enable/disable engines and auth methods (write action)

---

## Inspirations & References

- **[k9s](https://github.com/derailed/k9s)** — The gold standard for Kubernetes TUIs; primary UX inspiration
- **[lazygit](https://github.com/jesseduffield/lazygit)** — Excellent panel-based TUI; inspiration for modal confirmations
- **[lazydocker](https://github.com/jesseduffield/lazydocker)** — Resource monitoring in a TUI
- **[Bubble Tea examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)** — Reference implementations
- **[Vault API docs](https://developer.hashicorp.com/vault/api-docs)** — Canonical API reference
