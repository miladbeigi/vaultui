package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestEngineDashboardView_Title(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	if v.Title() != "Engine: secret/" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestEngineDashboardView_Init(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestEngineDashboardView_View_Loading(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading engine config") {
		t.Error("expected loading message")
	}
}

func TestEngineDashboardView_Update_Loaded(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	cfg := &vault.EngineConfig{
		Path:            "secret/",
		Type:            "kv",
		Description:     "Key/Value store",
		UUID:            "abc-123",
		Accessor:        "kv_abc",
		DefaultLeaseTTL: 30 * time.Minute,
		MaxLeaseTTL:     24 * time.Hour,
		Options:         map[string]string{"version": "2"},
	}

	updated, _ := v.Update(engineConfigMsg{config: cfg})
	ev := updated.(*EngineDashboardView)

	if ev.loading {
		t.Error("expected loading to be false")
	}
	if ev.config == nil {
		t.Fatal("expected config to be set")
	}
	if ev.config.Type != "kv" {
		t.Errorf("expected type kv, got %s", ev.config.Type)
	}
}

func TestEngineDashboardView_Update_Error(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	updated, _ := v.Update(engineConfigMsg{err: errTest})
	ev := updated.(*EngineDashboardView)

	if ev.loading {
		t.Error("expected loading to be false")
	}
	if ev.err == nil {
		t.Error("expected error to be set")
	}
}

func TestEngineDashboardView_View_Error(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	v.Update(engineConfigMsg{err: errTest}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Error") {
		t.Error("expected error in view output")
	}
}

func TestEngineDashboardView_View_Loaded(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "pki/")
	cfg := &vault.EngineConfig{
		Path:            "pki/",
		Type:            "pki",
		Description:     "PKI certificates",
		UUID:            "def-456",
		Accessor:        "pki_def",
		DefaultLeaseTTL: 0,
		MaxLeaseTTL:     87600 * time.Hour,
		RunningVersion:  "v1.15.6+builtin",
	}

	v.Update(engineConfigMsg{config: cfg}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Engine: pki/") {
		t.Error("expected title in view output")
	}
}

func TestEngineDashboardView_KeyHints(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
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

func TestEngineDashboardView_Refresh(t *testing.T) {
	v := NewEngineDashboardView(newTestClient(t), "secret/")
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	ev := updated.(*EngineDashboardView)

	if !ev.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

func TestFormatEngineTTL(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "system default"},
		{30 * time.Minute, "30m 0s"},
		{24 * time.Hour, "24h 0m"},
		{50 * time.Hour, "2d 2h"},
	}
	for _, tt := range tests {
		got := formatEngineTTL(tt.d)
		if got != tt.want {
			t.Errorf("formatEngineTTL(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
