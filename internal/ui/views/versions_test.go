package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func testVersions() []vault.VersionEntry {
	return []vault.VersionEntry{
		{Version: 3, CreatedTime: time.Now(), Destroyed: false},
		{Version: 2, CreatedTime: time.Now().Add(-1 * time.Hour), Destroyed: false},
		{Version: 1, CreatedTime: time.Now().Add(-2 * time.Hour), Destroyed: false},
	}
}

func TestVersionsView_Title(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")
	if !strings.Contains(v.Title(), "versions") {
		t.Errorf("expected title to contain 'versions', got %q", v.Title())
	}
}

func TestVersionsView_Init_ReturnsCmd(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestVersionsView_View_Loading(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestVersionsView_Update_Loaded(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")

	updated, cmd := v.Update(versionsLoadedMsg{versions: testVersions()})
	vv := updated.(*VersionsView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if vv.loading {
		t.Error("expected loading to be false")
	}
	if len(vv.versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(vv.versions))
	}
}

func TestVersionsView_Update_Error(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")

	updated, _ := v.Update(versionsLoadedMsg{err: errTest})
	vv := updated.(*VersionsView)

	if vv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestVersionsView_View_WithData(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")
	v.loading = false
	v.versions = testVersions()
	v.table.SetRows(v.buildRows())

	view := v.View(80, 20)
	if !strings.Contains(view, "v3") {
		t.Error("expected view to contain 'v3'")
	}
	if !strings.Contains(view, "v1") {
		t.Error("expected view to contain 'v1'")
	}
}

func TestVersionsView_Navigation(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")
	v.loading = false
	v.versions = testVersions()
	v.table.SetRows(v.buildRows())

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.table.Cursor() != 1 {
		t.Errorf("expected cursor 1 after j, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.table.Cursor() != 0 {
		t.Errorf("expected cursor 0 after k, got %d", v.table.Cursor())
	}
}

func TestVersionsView_KeyHints(t *testing.T) {
	v := NewVersionsView(newTestClient(t), "secret/", "apps/myapp/config")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
	found := false
	for _, h := range hints {
		if h.Key == "d" {
			found = true
		}
	}
	if !found {
		t.Error("expected diff key hint")
	}
}

func TestVersionsView_BuildRows_Status(t *testing.T) {
	versions := []vault.VersionEntry{
		{Version: 3, CreatedTime: time.Now()},
		{Version: 2, CreatedTime: time.Now(), Destroyed: true},
		{Version: 1, CreatedTime: time.Now(), DeletionTime: "2026-01-01T00:00:00Z"},
	}
	v := NewVersionsView(newTestClient(t), "secret/", "test")
	v.versions = versions
	rows := v.buildRows()

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}
