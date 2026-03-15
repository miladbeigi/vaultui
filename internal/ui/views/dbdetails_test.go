package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

// ── Connection detail tests ──────────────────────────────

func TestDBConnectionDetailView_Title(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	if v.Title() != "Connection: testdb" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestDBConnectionDetailView_Init(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestDBConnectionDetailView_Update_Loaded(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	detail := &vault.DBConnectionDetail{
		Name:             "testdb",
		PluginName:       "postgresql-database-plugin",
		ConnectionURL:    "postgresql://{{username}}:{{password}}@localhost:5432/testdb",
		AllowedRoles:     []string{"readonly", "readwrite"},
		VerifyConnection: true,
	}
	updated, _ := v.Update(dbConnDetailLoadedMsg{detail: detail})
	cv := updated.(*DBConnectionDetailView)

	if cv.loading {
		t.Error("expected loading to be false")
	}
	if cv.detail == nil {
		t.Fatal("expected detail to be set")
	}
	if cv.detail.PluginName != "postgresql-database-plugin" {
		t.Errorf("unexpected plugin: %s", cv.detail.PluginName)
	}
}

func TestDBConnectionDetailView_Update_Error(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	updated, _ := v.Update(dbConnDetailLoadedMsg{err: errTest})
	cv := updated.(*DBConnectionDetailView)

	if cv.err == nil {
		t.Error("expected error to be set")
	}
}

func TestDBConnectionDetailView_View_Loading(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading connection details") {
		t.Error("expected loading message")
	}
}

func TestDBConnectionDetailView_View_Error(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	v.Update(dbConnDetailLoadedMsg{err: errTest}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Error") {
		t.Error("expected error in view output")
	}
}

func TestDBConnectionDetailView_Refresh(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	cv := updated.(*DBConnectionDetailView)

	if !cv.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

func TestDBConnectionDetailView_KeyHints(t *testing.T) {
	v := NewDBConnectionDetailView(newTestClient(t), "database/", "testdb")
	hints := v.KeyHints()
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

// ── Role detail tests ────────────────────────────────────

func TestDBRoleDetailView_Title(t *testing.T) {
	v := NewDBRoleDetailView(newTestClient(t), "database/", "readonly")
	if v.Title() != "Role: readonly" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestDBRoleDetailView_Update_Loaded(t *testing.T) {
	v := NewDBRoleDetailView(newTestClient(t), "database/", "readonly")
	detail := &vault.DBRoleDetail{
		Name:               "readonly",
		DBName:             "testdb",
		DefaultTTL:         "1h",
		MaxTTL:             "24h",
		CreationStatements: []string{"CREATE ROLE ..."},
	}
	updated, _ := v.Update(dbRoleDetailLoadedMsg{detail: detail})
	rv := updated.(*DBRoleDetailView)

	if rv.loading {
		t.Error("expected loading to be false")
	}
	if rv.detail == nil {
		t.Fatal("expected detail to be set")
	}
	if rv.detail.DBName != "testdb" {
		t.Errorf("unexpected db_name: %s", rv.detail.DBName)
	}
}

func TestDBRoleDetailView_Update_Error(t *testing.T) {
	v := NewDBRoleDetailView(newTestClient(t), "database/", "readonly")
	updated, _ := v.Update(dbRoleDetailLoadedMsg{err: errTest})
	rv := updated.(*DBRoleDetailView)

	if rv.err == nil {
		t.Error("expected error to be set")
	}
}

func TestDBRoleDetailView_View_Loading(t *testing.T) {
	v := NewDBRoleDetailView(newTestClient(t), "database/", "readonly")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading role details") {
		t.Error("expected loading message")
	}
}

func TestDBRoleDetailView_Refresh(t *testing.T) {
	v := NewDBRoleDetailView(newTestClient(t), "database/", "readonly")
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	rv := updated.(*DBRoleDetailView)

	if !rv.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

// ── Static role detail tests ─────────────────────────────

func TestDBStaticRoleDetailView_Title(t *testing.T) {
	v := NewDBStaticRoleDetailView(newTestClient(t), "database/", "monitoring")
	if v.Title() != "Static Role: monitoring" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestDBStaticRoleDetailView_Update_Loaded(t *testing.T) {
	v := NewDBStaticRoleDetailView(newTestClient(t), "database/", "monitoring")
	detail := &vault.DBStaticRoleDetail{
		Name:           "monitoring",
		DBName:         "testdb",
		Username:       "monitor",
		RotationPeriod: "24h",
	}
	updated, _ := v.Update(dbStaticRoleDetailLoadedMsg{detail: detail})
	sv := updated.(*DBStaticRoleDetailView)

	if sv.loading {
		t.Error("expected loading to be false")
	}
	if sv.detail == nil {
		t.Fatal("expected detail to be set")
	}
	if sv.detail.Username != "monitor" {
		t.Errorf("unexpected username: %s", sv.detail.Username)
	}
}

func TestDBStaticRoleDetailView_Update_Error(t *testing.T) {
	v := NewDBStaticRoleDetailView(newTestClient(t), "database/", "monitoring")
	updated, _ := v.Update(dbStaticRoleDetailLoadedMsg{err: errTest})
	sv := updated.(*DBStaticRoleDetailView)

	if sv.err == nil {
		t.Error("expected error to be set")
	}
}

func TestDBStaticRoleDetailView_View_Loading(t *testing.T) {
	v := NewDBStaticRoleDetailView(newTestClient(t), "database/", "monitoring")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading static role details") {
		t.Error("expected loading message")
	}
}

func TestDBStaticRoleDetailView_Refresh(t *testing.T) {
	v := NewDBStaticRoleDetailView(newTestClient(t), "database/", "monitoring")
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	sv := updated.(*DBStaticRoleDetailView)

	if !sv.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}
