package vault

import (
	"fmt"
	"path"
)

// AuthMethod represents a supported authentication method.
type AuthMethod string

const (
	AuthToken    AuthMethod = "token"
	AuthUserpass AuthMethod = "userpass"
	AuthAppRole  AuthMethod = "approle"
)

// AuthConfig holds parameters for non-token auth methods.
type AuthConfig struct {
	Method    AuthMethod
	MountPath string

	// Userpass fields
	Username string
	Password string

	// AppRole fields
	RoleID   string
	SecretID string
}

// Authenticate performs authentication using the configured method and
// sets the resulting token on the client. For token auth this is a no-op
// since the token is already set via ClientConfig.
func (c *Client) Authenticate(cfg AuthConfig) error {
	switch cfg.Method {
	case AuthToken, "":
		return nil
	case AuthUserpass:
		return c.authUserpass(cfg)
	case AuthAppRole:
		return c.authAppRole(cfg)
	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.Method)
	}
}

func (c *Client) authUserpass(cfg AuthConfig) error {
	if cfg.Username == "" {
		return fmt.Errorf("userpass auth requires --username")
	}
	if cfg.Password == "" {
		return fmt.Errorf("userpass auth requires --password")
	}

	mount := cfg.MountPath
	if mount == "" {
		mount = "userpass"
	}

	loginPath := path.Join("auth", mount, "login", cfg.Username)
	secret, err := c.raw.Logical().Write(loginPath, map[string]interface{}{
		"password": cfg.Password,
	})
	if err != nil {
		return fmt.Errorf("userpass login: %w", err)
	}
	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("userpass login returned no auth data")
	}

	c.raw.SetToken(secret.Auth.ClientToken)
	return nil
}

func (c *Client) authAppRole(cfg AuthConfig) error {
	if cfg.RoleID == "" {
		return fmt.Errorf("approle auth requires --role-id")
	}

	mount := cfg.MountPath
	if mount == "" {
		mount = "approle"
	}

	loginPath := path.Join("auth", mount, "login")
	data := map[string]interface{}{
		"role_id": cfg.RoleID,
	}
	if cfg.SecretID != "" {
		data["secret_id"] = cfg.SecretID
	}

	secret, err := c.raw.Logical().Write(loginPath, data)
	if err != nil {
		return fmt.Errorf("approle login: %w", err)
	}
	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("approle login returned no auth data")
	}

	c.raw.SetToken(secret.Auth.ClientToken)
	return nil
}
