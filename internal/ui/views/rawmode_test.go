package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestSecretDetail_RawModeToggle(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "test", true)
	v.Update(secretReadMsg{data: &vault.SecretData{ //nolint:errcheck // test setup
		Data: map[string]string{"key1": "val1", "key2": "val2"},
		Keys: []string{"key1", "key2"},
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode to be true after J")
	}

	view := v.View(100, 24)
	if !strings.Contains(view, "[JSON]") {
		t.Error("expected [JSON] indicator in raw mode view")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode still true after switching to YAML")
	}
	view = v.View(100, 24)
	if !strings.Contains(view, "[YAML]") {
		t.Error("expected [YAML] indicator after switching")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyEsc}) //nolint:errcheck // test setup
	if v.rawMode {
		t.Error("expected rawMode to be false after Esc")
	}
}

func TestSecretDetail_RawModeToggleOff(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "test", false)
	v.Update(secretReadMsg{data: &vault.SecretData{ //nolint:errcheck // test setup
		Data: map[string]string{"a": "1"},
		Keys: []string{"a"},
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Fatal("expected rawMode on")
	}
	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if v.rawMode {
		t.Error("expected rawMode off after pressing J again")
	}
}

func TestTransitDetail_RawMode(t *testing.T) {
	v := NewTransitKeyDetailView(newTestClient(t), "transit/", "my-key")
	v.Update(transitKeyLoadedMsg{detail: &vault.TransitKeyDetail{ //nolint:errcheck // test setup
		Name: "my-key", Type: "aes256-gcm96", LatestVersion: 3,
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after y")
	}
	view := v.View(100, 24)
	if !strings.Contains(view, "[YAML]") {
		t.Error("expected [YAML] in view")
	}

	hints := v.KeyHints()
	foundCopy := false
	for _, h := range hints {
		if h.Key == "c" {
			foundCopy = true
		}
	}
	if !foundCopy {
		t.Error("expected copy hint in raw mode")
	}
}

func TestTokenInspector_RawMode(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	v.Update(tokenInspectMsg{details: &vault.TokenDetails{ //nolint:errcheck // test setup
		Accessor: "abc", DisplayName: "root", TokenType: "service",
		Policies: []string{"root"}, TTL: 3600 * time.Second,
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after J")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyEsc}) //nolint:errcheck // test setup
	if v.rawMode {
		t.Error("expected rawMode off after Esc")
	}
}

func TestEngineDashboard_RawMode(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	v.Update(engineConfigMsg{config: &vault.EngineConfig{ //nolint:errcheck // test setup
		Path: "secret/", Type: "kv", UUID: "abc-123",
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after J")
	}
	view := v.View(100, 24)
	if !strings.Contains(view, "[JSON]") {
		t.Error("expected [JSON] in view")
	}
}

func TestDBConnDetail_RawMode(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	v.Update(dbConnDetailLoadedMsg{detail: &vault.DBConnectionDetail{ //nolint:errcheck // test setup
		Name: "testdb", PluginName: "postgresql-database-plugin",
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after y")
	}
	view := v.View(100, 24)
	if !strings.Contains(view, "[YAML]") {
		t.Error("expected [YAML] in view")
	}
}

func TestDBRoleDetail_RawMode(t *testing.T) {
	v := NewDBRoleDetailView(newTestClient(t), "database/", "readonly")
	v.Update(dbRoleDetailLoadedMsg{detail: &vault.DBRoleDetail{ //nolint:errcheck // test setup
		Name: "readonly", DBName: "testdb", DefaultTTL: "1h",
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after J")
	}
}

func TestDBStaticRoleDetail_RawMode(t *testing.T) {
	v := NewDBStaticRoleDetailView(newTestClient(t), "database/", "monitoring")
	v.Update(dbStaticRoleDetailLoadedMsg{detail: &vault.DBStaticRoleDetail{ //nolint:errcheck // test setup
		Name: "monitoring", DBName: "testdb", Username: "monitor",
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after y")
	}
}

func TestAWSRoleDetail_RawMode(t *testing.T) {
	v := NewAWSRoleDetailView(newTestClient(t), "aws/", "deploy")
	v.Update(awsRoleDetailLoadedMsg{detail: &vault.AWSRoleDetail{ //nolint:errcheck // test setup
		Name: "deploy", CredentialTypes: []string{"iam_user"},
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after J")
	}
}

func TestAWSLeaseDetail_RawMode(t *testing.T) {
	lease := vault.AWSLease{
		LeaseID: "aws/creds/deploy/abc123", TTL: 30 * time.Minute,
		IssueTime: "2026-01-01T00:00:00Z", Renewable: true,
	}
	v := NewAWSLeaseDetailView(lease)

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after y")
	}
	view := v.View(100, 24)
	if !strings.Contains(view, "[YAML]") {
		t.Error("expected [YAML] in view")
	}
}

func TestIdentityDetail_RawMode(t *testing.T) {
	v := NewIdentityDetailView(newTestClient(t), true, "id-1", "test-entity")
	v.Update(identityDetailLoadedMsg{entity: &vault.IdentityEntity{ //nolint:errcheck // test setup
		Name: "test-entity", ID: "id-1", Policies: []string{"default"},
	}})

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if !v.rawMode {
		t.Error("expected rawMode after J")
	}
}

func TestRawMode_NoDataNoToggle(t *testing.T) {
	v := NewTransitKeyDetailView(newTestClient(t), "transit/", "my-key")

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}) //nolint:errcheck // test setup
	if v.rawMode {
		t.Error("should not enter rawMode when no data loaded")
	}
}
