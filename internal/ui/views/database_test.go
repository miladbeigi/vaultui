package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestDatabaseView_Title(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	if v.Title() != "Database: database/" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestDatabaseView_Init(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestDatabaseView_View_Loading(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading database data") {
		t.Error("expected loading message")
	}
}

func TestDatabaseView_Update_Loaded(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	msg := dbLoadedMsg{
		conns: []vault.DBConnection{
			{Name: "testdb", PluginName: "postgresql-database-plugin", AllowedRoles: []string{"readonly"}},
		},
		roles: []vault.DBRole{
			{Name: "readonly", DBName: "testdb", DefaultTTL: "1h", MaxTTL: "24h"},
		},
		staticRoles: []vault.DBStaticRole{
			{Name: "monitoring", DBName: "testdb", Username: "monitor", RotationPeriod: "24h"},
		},
	}

	updated, _ := v.Update(msg)
	dv := updated.(*DatabaseView)

	if dv.loading {
		t.Error("expected loading to be false")
	}
	if len(dv.conns) != 1 {
		t.Errorf("expected 1 connection, got %d", len(dv.conns))
	}
	if len(dv.roles) != 1 {
		t.Errorf("expected 1 role, got %d", len(dv.roles))
	}
	if len(dv.staticRoles) != 1 {
		t.Errorf("expected 1 static role, got %d", len(dv.staticRoles))
	}
}

func TestDatabaseView_Update_Error(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	updated, _ := v.Update(dbLoadedMsg{err: errTest})
	dv := updated.(*DatabaseView)

	if dv.loading {
		t.Error("expected loading to be false")
	}
	if dv.err == nil {
		t.Error("expected error to be set")
	}
}

func TestDatabaseView_View_Error(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	v.Update(dbLoadedMsg{err: errTest}) //nolint:errcheck // test setup
	view := v.View(80, 20)
	if !strings.Contains(view, "Error") {
		t.Error("expected error in view output")
	}
}

func TestDatabaseView_View_Loaded(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	v.Update(dbLoadedMsg{ //nolint:errcheck // test setup
		conns: []vault.DBConnection{{Name: "mydb", PluginName: "pg"}},
	})
	view := v.View(100, 24)
	if !strings.Contains(view, "Database: database/") {
		t.Error("expected title in view output")
	}
}

func TestDatabaseView_TabSwitch(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	v.Update(dbLoadedMsg{ //nolint:errcheck // test setup
		conns:       []vault.DBConnection{{Name: "c1"}},
		roles:       []vault.DBRole{{Name: "r1"}},
		staticRoles: []vault.DBStaticRole{{Name: "sr1"}},
	})

	if v.tab != 0 {
		t.Error("expected initial tab to be 0")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	if v.tab != 1 {
		t.Errorf("expected tab 1 after first tab press, got %d", v.tab)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	if v.tab != 2 {
		t.Errorf("expected tab 2 after second tab press, got %d", v.tab)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	if v.tab != 0 {
		t.Errorf("expected tab 0 after third tab press (wrap), got %d", v.tab)
	}
}

func TestDatabaseView_KeyHints(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected non-empty hints")
	}
	found := false
	for _, h := range hints {
		if h.Key == "tab" {
			found = true
		}
	}
	if !found {
		t.Error("expected tab hint")
	}
}

func TestDatabaseView_EmptyTabs(t *testing.T) {
	v := NewDatabaseView(newTestClient(t), "database/")
	v.Update(dbLoadedMsg{}) //nolint:errcheck // test setup

	view := v.View(80, 20)
	if !strings.Contains(view, "No connections found") {
		t.Error("expected empty connections message")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	view = v.View(80, 20)
	if !strings.Contains(view, "No roles found") {
		t.Error("expected empty roles message")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab}) //nolint:errcheck // test setup
	view = v.View(80, 20)
	if !strings.Contains(view, "No static roles found") {
		t.Error("expected empty static roles message")
	}
}
