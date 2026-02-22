package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient_DefaultAddress(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Address() == "" {
		t.Error("expected a default address, got empty string")
	}
}

func TestNewClient_CustomAddress(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	addr := "http://10.0.0.1:8200"
	client, err := NewClient(ClientConfig{Address: addr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Address() != addr {
		t.Errorf("expected address %q, got %q", addr, client.Address())
	}
}

func TestNewClient_Namespace(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	client, err := NewClient(ClientConfig{Namespace: "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Namespace() != "prod" {
		t.Errorf("expected namespace %q, got %q", "prod", client.Namespace())
	}
}

func TestNewClient_EmptyNamespace(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Namespace() != "" {
		t.Errorf("expected empty namespace, got %q", client.Namespace())
	}
}

func TestNewClient_ExplicitToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	client, err := NewClient(ClientConfig{Token: "my-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.HasToken() {
		t.Error("expected client to have a token")
	}
	if client.Raw().Token() != "my-token" {
		t.Errorf("expected token %q, got %q", "my-token", client.Raw().Token())
	}
}

func TestNewClient_EnvToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "env-token")

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.HasToken() {
		t.Error("expected client to have a token from VAULT_TOKEN")
	}
	if client.Raw().Token() != "env-token" {
		t.Errorf("expected token %q, got %q", "env-token", client.Raw().Token())
	}
}

func TestNewClient_ExplicitTokenOverridesEnv(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "env-token")

	client, err := NewClient(ClientConfig{Token: "explicit-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Raw().Token() != "explicit-token" {
		t.Errorf("expected explicit token to win, got %q", client.Raw().Token())
	}
}

func TestNewClient_TokenFromFile(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	tokenContent := "  file-token  \n"
	if err := os.WriteFile(filepath.Join(tmpHome, ".vault-token"), []byte(tokenContent), 0o600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Raw().Token() != "file-token" {
		t.Errorf("expected file token %q, got %q", "file-token", client.Raw().Token())
	}
}

func TestNewClient_NoToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.HasToken() {
		t.Errorf("expected no token, got %q", client.Raw().Token())
	}
}

func TestNewClient_EnvAddress(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://envaddr:8200")
	t.Setenv("VAULT_TOKEN", "")

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Address() != "http://envaddr:8200" {
		t.Errorf("expected env address, got %q", client.Address())
	}
}

func TestNewClient_ExplicitAddressOverridesEnv(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://envaddr:8200")
	t.Setenv("VAULT_TOKEN", "")

	client, err := NewClient(ClientConfig{Address: "http://explicit:8200"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Address() != "http://explicit:8200" {
		t.Errorf("expected explicit address to win, got %q", client.Address())
	}
}
