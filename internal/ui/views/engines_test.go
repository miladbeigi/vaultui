package views

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func newTestClient(t *testing.T) *vault.Client {
	t.Helper()
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	c, err := vault.NewClient(vault.ClientConfig{
		Address: "http://127.0.0.1:8200",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return c
}

func TestEnginesView_Title(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	if v.Title() != "Secret Engines" {
		t.Errorf("expected title 'Secret Engines', got %q", v.Title())
	}
}

func TestEnginesView_Init_ReturnsCmd(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestEnginesView_View_Loading(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestEnginesView_Update_LoadedMsg(t *testing.T) {
	v := NewEnginesView(newTestClient(t))

	engines := []vault.MountEntry{
		{Path: "secret/", Type: "kv", Version: "v2", Description: "Key/Value store"},
		{Path: "pki/", Type: "pki", Description: "PKI certificates"},
	}

	updated, cmd := v.Update(enginesLoadedMsg{engines: engines})
	ev := updated.(*EnginesView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if ev.loading {
		t.Error("expected loading to be false after data arrives")
	}
	if len(ev.engines) != 2 {
		t.Errorf("expected 2 engines, got %d", len(ev.engines))
	}
}

func TestEnginesView_Update_LoadedError(t *testing.T) {
	v := NewEnginesView(newTestClient(t))

	updated, _ := v.Update(enginesLoadedMsg{err: errTest})
	ev := updated.(*EnginesView)

	if ev.loading {
		t.Error("expected loading to be false")
	}
	if ev.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestEnginesView_View_WithData(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	v.loading = false
	v.engines = []vault.MountEntry{
		{Path: "secret/", Type: "kv", Version: "v2", Description: "Key/Value store"},
	}
	v.table.SetRows(v.buildRows())

	view := v.View(80, 20)
	if !strings.Contains(view, "secret/") {
		t.Error("expected view to contain 'secret/'")
	}
	if !strings.Contains(view, "kv") {
		t.Error("expected view to contain 'kv'")
	}
}

func TestEnginesView_View_Error(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	v.loading = false
	v.err = errTest

	view := v.View(80, 20)
	if !strings.Contains(view, "test error") {
		t.Error("expected view to show error message")
	}
}

func TestEnginesView_View_Empty(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	v.loading = false
	v.engines = []vault.MountEntry{}

	view := v.View(80, 20)
	if !strings.Contains(view, "No secret engines") {
		t.Error("expected empty state message")
	}
}

func TestEnginesView_Navigation(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	v.loading = false
	v.engines = []vault.MountEntry{
		{Path: "secret/", Type: "kv"},
		{Path: "pki/", Type: "pki"},
		{Path: "transit/", Type: "transit"},
	}
	v.table.SetRows(v.buildRows())

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.table.Cursor() != 1 {
		t.Errorf("expected cursor 1 after j, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.table.Cursor() != 2 {
		t.Errorf("expected cursor 2 after second j, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.table.Cursor() != 1 {
		t.Errorf("expected cursor 1 after k, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if v.table.Cursor() != 2 {
		t.Errorf("expected cursor 2 after G, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if v.table.Cursor() != 0 {
		t.Errorf("expected cursor 0 after g, got %d", v.table.Cursor())
	}
}

func TestEnginesView_SelectedEngine(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	v.loading = false
	v.engines = []vault.MountEntry{
		{Path: "secret/", Type: "kv"},
		{Path: "pki/", Type: "pki"},
	}
	v.table.SetRows(v.buildRows())

	sel := v.SelectedEngine()
	if sel == nil || sel.Path != "secret/" {
		t.Error("expected first engine to be selected")
	}

	v.table.MoveDown()
	sel = v.SelectedEngine()
	if sel == nil || sel.Path != "pki/" {
		t.Error("expected second engine after move down")
	}
}

func TestEnginesView_KeyHints(t *testing.T) {
	v := NewEnginesView(newTestClient(t))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}

var errTest = fmt.Errorf("test error")
