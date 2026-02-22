package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the official Vault API client with convenience methods
// and token resolution logic.
type Client struct {
	raw       *vaultapi.Client
	addr      string
	namespace string
}

// ClientConfig holds the parameters needed to create a Client.
type ClientConfig struct {
	Address   string
	Token     string
	Namespace string
}

// NewClient creates a configured Vault client. Token resolution order:
//  1. Explicit token from config/flag
//  2. VAULT_TOKEN environment variable (handled by vault/api)
//  3. ~/.vault-token file
func NewClient(cfg ClientConfig) (*Client, error) {
	apiCfg := vaultapi.DefaultConfig()
	if apiCfg == nil {
		apiCfg = &vaultapi.Config{}
	}

	if cfg.Address != "" {
		apiCfg.Address = cfg.Address
	}

	raw, err := vaultapi.NewClient(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	if cfg.Namespace != "" {
		raw.SetNamespace(cfg.Namespace)
	}

	resolveToken(raw, cfg.Token)

	return &Client{
		raw:       raw,
		addr:      raw.Address(),
		namespace: cfg.Namespace,
	}, nil
}

// resolveToken sets the client token using the first available source:
// explicit value, then VAULT_TOKEN (already set by vault/api), then ~/.vault-token file.
func resolveToken(c *vaultapi.Client, explicit string) {
	if explicit != "" {
		c.SetToken(explicit)
		return
	}

	// vault/api.NewClient already reads VAULT_TOKEN into the client,
	// so if a token is set at this point, we're done.
	if c.Token() != "" {
		return
	}

	if token := readTokenFile(); token != "" {
		c.SetToken(token)
	}
}

func readTokenFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".vault-token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Raw returns the underlying vault/api.Client for advanced usage.
func (c *Client) Raw() *vaultapi.Client {
	return c.raw
}

// Address returns the Vault server address.
func (c *Client) Address() string {
	return c.addr
}

// Namespace returns the configured Vault namespace.
func (c *Client) Namespace() string {
	return c.namespace
}

// HasToken reports whether the client has a token set.
func (c *Client) HasToken() bool {
	return c.raw.Token() != ""
}
