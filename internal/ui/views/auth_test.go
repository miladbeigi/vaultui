package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/milad/vaultui/internal/vault"
)

func TestAuthMethodsView_Title(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	if v.Title() != "Auth Methods" {
		t.Errorf("expected title 'Auth Methods', got %q", v.Title())
	}
}

func TestAuthMethodsView_Init_ReturnsCmd(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestAuthMethodsView_View_Loading(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestAuthMethodsView_Update_LoadedMsg(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))

	methods := []vault.MountEntry{
		{Path: "token/", Type: "token", Description: "Token-based auth"},
		{Path: "approle/", Type: "approle", Description: "AppRole auth"},
	}

	updated, cmd := v.Update(authLoadedMsg{methods: methods})
	av := updated.(*AuthMethodsView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if av.loading {
		t.Error("expected loading to be false after data arrives")
	}
	if len(av.methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(av.methods))
	}
}

func TestAuthMethodsView_Update_LoadedError(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))

	updated, _ := v.Update(authLoadedMsg{err: errTest})
	av := updated.(*AuthMethodsView)

	if av.loading {
		t.Error("expected loading to be false")
	}
	if av.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestAuthMethodsView_View_WithData(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	v.loading = false
	v.methods = []vault.MountEntry{
		{Path: "token/", Type: "token", Description: "Token-based auth"},
	}
	v.table.SetRows(v.buildRows())

	view := v.View(80, 20)
	if !strings.Contains(view, "token/") {
		t.Error("expected view to contain 'token/'")
	}
	if !strings.Contains(view, "Token-based") {
		t.Error("expected view to contain description")
	}
}

func TestAuthMethodsView_View_Error(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	v.loading = false
	v.err = errTest

	view := v.View(80, 20)
	if !strings.Contains(view, "test error") {
		t.Error("expected view to show error message")
	}
}

func TestAuthMethodsView_View_Empty(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	v.loading = false
	v.methods = []vault.MountEntry{}

	view := v.View(80, 20)
	if !strings.Contains(view, "No auth methods") {
		t.Error("expected empty state message")
	}
}

func TestAuthMethodsView_Navigation(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	v.loading = false
	v.methods = []vault.MountEntry{
		{Path: "token/", Type: "token"},
		{Path: "approle/", Type: "approle"},
		{Path: "oidc/", Type: "oidc"},
	}
	v.table.SetRows(v.buildRows())

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.table.Cursor() != 1 {
		t.Errorf("expected cursor 1 after j, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.table.Cursor() != 0 {
		t.Errorf("expected cursor 0 after k, got %d", v.table.Cursor())
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

func TestAuthMethodsView_KeyHints(t *testing.T) {
	v := NewAuthMethodsView(newTestClient(t))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}
