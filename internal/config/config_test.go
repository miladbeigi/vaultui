package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonexistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Contexts) != 0 {
		t.Errorf("expected empty contexts, got %d", len(cfg.Contexts))
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `current_context: dev
contexts:
  - name: dev
    address: http://localhost:8200
    token: dev-token
    namespace: dev-ns
  - name: prod
    address: https://vault.example.com
    auth:
      method: userpass
      username: admin
settings:
  theme: dark
  clipboard_timeout: 30
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CurrentContext != "dev" {
		t.Errorf("expected current_context 'dev', got %q", cfg.CurrentContext)
	}
	if len(cfg.Contexts) != 2 {
		t.Fatalf("expected 2 contexts, got %d", len(cfg.Contexts))
	}
	if cfg.Contexts[0].Address != "http://localhost:8200" {
		t.Errorf("unexpected address: %s", cfg.Contexts[0].Address)
	}
	if cfg.Contexts[1].Auth.Method != "userpass" {
		t.Errorf("unexpected auth method: %s", cfg.Contexts[1].Auth.Method)
	}
	if cfg.Settings.ClipboardTimeout != 30 {
		t.Errorf("expected clipboard_timeout 30, got %d", cfg.Settings.ClipboardTimeout)
	}
}

func TestSave_And_Reload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	cfg := &Config{
		CurrentContext: "prod",
		Contexts: []Context{
			{Name: "prod", Address: "https://vault.example.com"},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.CurrentContext != "prod" {
		t.Errorf("expected current_context 'prod', got %q", loaded.CurrentContext)
	}
	if len(loaded.Contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(loaded.Contexts))
	}
}

func TestGetContext(t *testing.T) {
	cfg := &Config{
		Contexts: []Context{
			{Name: "dev", Address: "http://localhost:8200"},
			{Name: "prod", Address: "https://vault.example.com"},
		},
	}

	ctx := cfg.GetContext("prod")
	if ctx == nil {
		t.Fatal("expected to find 'prod' context")
	}
	if ctx.Address != "https://vault.example.com" {
		t.Errorf("unexpected address: %s", ctx.Address)
	}

	if cfg.GetContext("nonexistent") != nil {
		t.Error("expected nil for nonexistent context")
	}
}

func TestCurrentCtx(t *testing.T) {
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: []Context{
			{Name: "dev", Address: "http://localhost:8200"},
		},
	}

	ctx := cfg.CurrentCtx()
	if ctx == nil {
		t.Fatal("expected current context")
	}
	if ctx.Name != "dev" {
		t.Errorf("expected 'dev', got %q", ctx.Name)
	}

	cfg.CurrentContext = ""
	if cfg.CurrentCtx() != nil {
		t.Error("expected nil when current_context is empty")
	}
}

func TestContextNames(t *testing.T) {
	cfg := &Config{
		Contexts: []Context{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
	}

	names := cfg.ContextNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("unexpected names: %v", names)
	}
}
