package views

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func testSecretData() *vault.SecretData {
	return &vault.SecretData{
		Data: map[string]string{
			"db_host":     "db.internal.example.com",
			"db_password": "s3cret!",
			"db_port":     "5432",
		},
		Keys: []string{"db_host", "db_password", "db_port"},
	}
}

func TestSecretDetailView_Title(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	want := "secret/apps/config"
	if v.Title() != want {
		t.Errorf("expected title %q, got %q", want, v.Title())
	}
}

func TestSecretDetailView_Init_ReturnsCmd(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestSecretDetailView_View_Loading(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestSecretDetailView_Update_Loaded(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)

	updated, cmd := v.Update(secretReadMsg{data: testSecretData()})
	sv := updated.(*SecretDetailView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if sv.loading {
		t.Error("expected loading to be false")
	}
	if sv.secret == nil {
		t.Fatal("expected secret to be populated")
	}
	if len(sv.secret.Keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(sv.secret.Keys))
	}
}

func TestSecretDetailView_Update_LoadError(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)

	updated, _ := v.Update(secretReadMsg{err: fmt.Errorf("forbidden")})
	sv := updated.(*SecretDetailView)

	if sv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestSecretDetailView_View_WithData(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.secret = testSecretData()
	v.table.SetRows(v.buildRows())

	view := v.View(80, 20)

	if !strings.Contains(view, "db_host") {
		t.Error("expected view to contain key 'db_host'")
	}
	if !strings.Contains(view, "db.internal.example.com") {
		t.Error("expected view to contain value")
	}
	if !strings.Contains(view, "db_password") {
		t.Error("expected view to contain key 'db_password'")
	}
}

func TestSecretDetailView_View_Error(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.err = fmt.Errorf("permission denied")

	view := v.View(80, 20)
	if !strings.Contains(view, "permission denied") {
		t.Error("expected error message in view")
	}
}

func TestSecretDetailView_View_Empty(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.secret = &vault.SecretData{Data: map[string]string{}, Keys: []string{}}

	view := v.View(80, 20)
	if !strings.Contains(view, "Empty") {
		t.Error("expected empty state message")
	}
}

func TestSecretDetailView_Navigation(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.secret = testSecretData()
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

func TestSecretDetailView_Breadcrumb(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/myapp/config", true)
	v.loading = false
	v.secret = testSecretData()
	v.table.SetRows(v.buildRows())

	view := v.View(80, 20)
	if !strings.Contains(view, "secret/") {
		t.Error("expected breadcrumb to contain mount")
	}
	if !strings.Contains(view, "apps") {
		t.Error("expected breadcrumb to contain path segment")
	}
	if !strings.Contains(view, "config") {
		t.Error("expected breadcrumb to contain secret name")
	}
}

func TestSecretDetailView_KeyHints(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}

	hintMap := make(map[string]bool)
	for _, h := range hints {
		hintMap[h.Key] = true
	}
	for _, expected := range []string{"c", "C"} {
		if !hintMap[expected] {
			t.Errorf("expected hint for key %q", expected)
		}
	}
}

func TestSecretDetailView_CopyValue_SetsStatus(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.secret = testSecretData()
	v.table.SetRows(v.buildRows())

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if v.statusMsg == "" {
		t.Error("expected statusMsg to be set after copy")
	}
	if cmd == nil {
		t.Error("expected a clear command to be returned")
	}
}

func TestSecretDetailView_CopyJSON_SetsStatus(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.secret = testSecretData()
	v.table.SetRows(v.buildRows())

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})

	if v.statusMsg == "" {
		t.Error("expected statusMsg to be set after copy JSON")
	}
	if cmd == nil {
		t.Error("expected a clear command to be returned")
	}
}

func TestSecretDetailView_StatusClearMsg(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false
	v.secret = testSecretData()
	v.statusMsg = "some status"

	v.Update(statusClearMsg{})
	if v.statusMsg != "" {
		t.Error("expected statusMsg to be cleared")
	}
}

func TestSecretDetailView_NoKeyHandling_BeforeLoad(t *testing.T) {
	v := NewSecretDetailView(newTestClient(t), "secret/", "apps/config", true)
	v.loading = false

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd != nil {
		t.Error("expected no command when no secret loaded")
	}
}
