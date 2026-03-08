package views

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestAuditView_Title(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	if v.Title() != "Audit & Logs" {
		t.Errorf("unexpected title: %s", v.Title())
	}
}

func TestAuditView_Init(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestAuditView_View_NoLogPath(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false
	v.logErr = fmt.Errorf("no audit log file configured")
	view := v.View(80, 20)
	if !strings.Contains(view, "no audit log") {
		t.Error("expected error about missing log file")
	}
}

func TestAuditView_View_ConnectedWaiting(t *testing.T) {
	v := NewAuditView(newTestClient(t), "/tmp/audit.log")
	v.loading = false
	v.connected = true
	view := v.View(80, 20)
	if !strings.Contains(view, "Connected") {
		t.Error("expected connected message in log stream tab")
	}
}

func TestAuditView_View_WithEntries(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false
	v.entries = []vault.AuditEntry{
		{Type: "request", Operation: "read", Path: "secret/data/foo"},
		{Type: "response", Operation: "read", Path: "secret/data/foo"},
		{Type: "request", Operation: "list", Path: "sys/mounts"},
	}
	v.scroll = len(v.entries)

	view := v.View(80, 20)
	if !strings.Contains(view, "3 entries") {
		t.Error("expected entry count in status line")
	}
	if !strings.Contains(view, "LIVE") {
		t.Error("expected LIVE status")
	}
}

func TestAuditView_View_Paused(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false
	v.paused = true
	v.entries = []vault.AuditEntry{
		{Type: "request", Operation: "read", Path: "test"},
	}
	v.scroll = 1

	view := v.View(80, 20)
	if !strings.Contains(view, "PAUSED") {
		t.Error("expected PAUSED status")
	}
}

func TestAuditView_TogglePause(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false

	if v.paused {
		t.Error("expected not paused initially")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if !v.paused {
		t.Error("expected paused after pressing p")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if v.paused {
		t.Error("expected unpaused after pressing p again")
	}
}

func TestAuditView_ClearEntries(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false
	v.entries = []vault.AuditEntry{
		{Type: "request", Operation: "read", Path: "test"},
	}
	v.scroll = 1

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if len(v.entries) != 0 {
		t.Error("expected entries cleared after pressing c")
	}
	if v.scroll != 0 {
		t.Error("expected scroll reset to 0")
	}
}

func TestAuditView_TabSwitch(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false

	if v.tab != 0 {
		t.Error("expected audit log tab initially")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.tab != 1 {
		t.Error("expected devices tab after tab press")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.tab != 0 {
		t.Error("expected audit log tab after second tab press")
	}
}

func TestAuditView_DevicesLoaded(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	devices := []vault.AuditDevice{
		{Path: "file/", Type: "file", Description: "File audit log"},
	}

	updated, _ := v.Update(auditDevicesMsg{devices: devices})
	av := updated.(*AuditView)

	if av.loading {
		t.Error("expected loading to be false")
	}
	if len(av.devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(av.devices))
	}
}

func TestAuditView_DevicesError(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.tab = 1

	updated, _ := v.Update(auditDevicesMsg{err: fmt.Errorf("permission denied")})
	av := updated.(*AuditView)

	view := av.View(80, 20)
	if !strings.Contains(view, "permission denied") {
		t.Error("expected error message in devices view")
	}
}

func TestAuditView_DevicesEmpty(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.tab = 1
	v.loading = false

	view := v.View(80, 20)
	if !strings.Contains(view, "No audit devices") {
		t.Error("expected empty state message")
	}
}

func TestAuditView_KeyHints_LogTab(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.tab = 0
	hints := v.KeyHints()
	found := false
	for _, h := range hints {
		if h.Key == "p" {
			found = true
		}
	}
	if !found {
		t.Error("expected pause hint in log tab")
	}
}

func TestAuditView_KeyHints_DevicesTab(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.tab = 1
	hints := v.KeyHints()
	for _, h := range hints {
		if h.Key == "p" {
			t.Error("pause hint should not be in devices tab")
		}
	}
}

func TestAuditView_AppendEntry_MaxCap(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	for i := 0; i < maxAuditEntries+50; i++ {
		v.appendEntry(vault.AuditEntry{Type: "request", Operation: "read"})
	}
	if len(v.entries) != maxAuditEntries {
		t.Errorf("expected %d entries, got %d", maxAuditEntries, len(v.entries))
	}
}

func TestAuditView_AppendEntry_Paused(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.paused = true
	v.appendEntry(vault.AuditEntry{Type: "request"})
	if len(v.entries) != 0 {
		t.Error("expected no entries when paused")
	}
}

func TestAuditView_StreamError(t *testing.T) {
	v := NewAuditView(newTestClient(t), "")
	v.loading = false

	updated, _ := v.Update(auditStreamErrMsg{err: fmt.Errorf("file not found")})
	av := updated.(*AuditView)

	view := av.View(80, 20)
	if !strings.Contains(view, "file not found") {
		t.Error("expected error in audit log view")
	}
}
