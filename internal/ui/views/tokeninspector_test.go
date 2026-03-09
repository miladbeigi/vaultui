package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestTokenInspectorView_Title(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	if v.Title() != "Token Inspector" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestTokenInspectorView_Init(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestTokenInspectorView_View_Loading(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	view := v.View(80, 20)
	if !strings.Contains(view, "Inspecting token") {
		t.Error("expected loading message")
	}
}

func TestTokenInspectorView_Update_Loaded(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	details := &vault.TokenDetails{
		Accessor:    "abc123",
		DisplayName: "token-root",
		TokenType:   "service",
		Policies:    []string{"root"},
		Renewable:   true,
		TTL:         3600 * time.Second,
		CreationTTL: 7200 * time.Second,
	}

	updated, _ := v.Update(tokenInspectMsg{details: details})
	tv := updated.(*TokenInspectorView)

	if tv.loading {
		t.Error("expected loading to be false")
	}
	if tv.details == nil {
		t.Fatal("expected details to be set")
	}
	if tv.details.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", tv.details.Accessor)
	}
}

func TestTokenInspectorView_Update_Error(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	updated, _ := v.Update(tokenInspectMsg{err: errTest})
	tv := updated.(*TokenInspectorView)

	if tv.loading {
		t.Error("expected loading to be false")
	}
	if tv.err == nil {
		t.Error("expected error to be set")
	}
}

func TestTokenInspectorView_View_Error(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	v.Update(tokenInspectMsg{err: errTest}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Error") {
		t.Error("expected error in view output")
	}
}

func TestTokenInspectorView_View_Loaded(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	details := &vault.TokenDetails{
		Accessor:    "acc-xyz",
		DisplayName: "token-test",
		TokenType:   "service",
		Policies:    []string{"default", "admin"},
		Renewable:   false,
		TTL:         3600 * time.Second,
		CreationTTL: 7200 * time.Second,
		EntityID:    "entity-123",
		Path:        "auth/token/create",
	}

	v.Update(tokenInspectMsg{details: details}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Token Inspector") {
		t.Error("expected title in view output")
	}
}

func TestTokenInspectorView_KeyHints(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected non-empty hints")
	}
	found := false
	for _, h := range hints {
		if h.Key == "r" {
			found = true
		}
	}
	if !found {
		t.Error("expected refresh hint")
	}
}

func TestTokenInspectorView_Refresh(t *testing.T) {
	v := NewTokenInspectorView(newTestClient(t))
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	tv := updated.(*TokenInspectorView)

	if !tv.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

func TestFormatDurationHuman(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "∞ (no expiry)"},
		{30 * time.Second, "30s"},
		{5*time.Minute + 10*time.Second, "5m 10s"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
		{50 * time.Hour, "2d 2h"},
	}

	for _, tt := range tests {
		got := formatDurationHuman(tt.d)
		if got != tt.want {
			t.Errorf("formatDurationHuman(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
