package views

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestPathBrowserView_Title(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "apps/myapp/", true)
	if v.Title() != "secret/apps/myapp/" {
		t.Errorf("expected title 'secret/apps/myapp/', got %q", v.Title())
	}
}

func TestPathBrowserView_Title_Root(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	if v.Title() != "secret/" {
		t.Errorf("expected title 'secret/', got %q", v.Title())
	}
}

func TestPathBrowserView_Init_ReturnsCmd(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestPathBrowserView_View_Loading(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
	if !strings.Contains(view, "secret/") {
		t.Error("expected breadcrumb with mount path")
	}
}

func TestPathBrowserView_Update_Loaded(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)

	entries := []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "config", IsDir: false},
	}

	updated, cmd := v.Update(pathListMsg{entries: entries})
	pv := updated.(*PathBrowserView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if pv.loading {
		t.Error("expected loading to be false")
	}
	if len(pv.entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(pv.entries))
	}
}

func TestPathBrowserView_Update_LoadError(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)

	updated, _ := v.Update(pathListMsg{err: fmt.Errorf("forbidden")})
	pv := updated.(*PathBrowserView)

	if pv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestPathBrowserView_View_WithData(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "apps/", true)
	v.loading = false
	v.entries = []vault.PathEntry{
		{Name: "production/", IsDir: true},
		{Name: "config", IsDir: false},
	}
	v.applyFilter()

	view := v.View(80, 20)
	if !strings.Contains(view, "production/") {
		t.Error("expected view to contain directory name")
	}
	if !strings.Contains(view, "config") {
		t.Error("expected view to contain secret name")
	}
	if !strings.Contains(view, "dir") {
		t.Error("expected view to show 'dir' type")
	}
	if !strings.Contains(view, "secret") {
		t.Error("expected view to show 'secret' type")
	}
}

func TestPathBrowserView_View_Error(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.loading = false
	v.err = fmt.Errorf("permission denied")

	view := v.View(80, 20)
	if !strings.Contains(view, "permission denied") {
		t.Error("expected error message in view")
	}
}

func TestPathBrowserView_View_Empty(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.loading = false
	v.entries = []vault.PathEntry{}

	view := v.View(80, 20)
	if !strings.Contains(view, "Empty") {
		t.Error("expected empty state message")
	}
}

func TestPathBrowserView_Breadcrumb(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "apps/myapp/", true)
	v.loading = false
	v.entries = []vault.PathEntry{}

	view := v.View(80, 20)
	if !strings.Contains(view, "secret/") {
		t.Error("expected breadcrumb to contain mount")
	}
	if !strings.Contains(view, "apps") {
		t.Error("expected breadcrumb to contain first segment")
	}
	if !strings.Contains(view, "myapp") {
		t.Error("expected breadcrumb to contain second segment")
	}
}

func TestPathBrowserView_Navigation(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.loading = false
	v.entries = []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "infra/", IsDir: true},
		{Name: "config", IsDir: false},
	}
	v.applyFilter()

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.table.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", v.table.Cursor())
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.table.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", v.table.Cursor())
	}
}

func TestPathBrowserView_Enter_Directory(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.loading = false
	v.entries = []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "config", IsDir: false},
	}
	v.applyFilter()

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command when entering a directory")
	}

	msg := cmd()
	pushMsg, ok := msg.(ui.PushViewMsg)
	if !ok {
		t.Fatalf("expected PushViewMsg, got %T", msg)
	}

	nextView, ok := pushMsg.View.(*PathBrowserView)
	if !ok {
		t.Fatalf("expected *PathBrowserView, got %T", pushMsg.View)
	}
	if nextView.path != "apps/" {
		t.Errorf("expected path 'apps/', got %q", nextView.path)
	}
	if nextView.mount != "secret/" {
		t.Errorf("expected mount 'secret/', got %q", nextView.mount)
	}
}

func TestPathBrowserView_Enter_LeafSecret(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "apps/", true)
	v.loading = false
	v.entries = []vault.PathEntry{
		{Name: "config", IsDir: false},
	}
	v.applyFilter()

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command when entering a leaf secret")
	}

	msg := cmd()
	pushMsg, ok := msg.(ui.PushViewMsg)
	if !ok {
		t.Fatalf("expected PushViewMsg, got %T", msg)
	}

	detailView, ok := pushMsg.View.(*SecretDetailView)
	if !ok {
		t.Fatalf("expected *SecretDetailView, got %T", pushMsg.View)
	}
	if detailView.path != "apps/config" {
		t.Errorf("expected path 'apps/config', got %q", detailView.path)
	}
}

func TestPathBrowserView_KeyHints(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
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

func TestPathBrowserView_Refresh(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.loading = false

	updated, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	pv := updated.(*PathBrowserView)

	if !pv.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Error("expected a command from refresh")
	}
}

func runeKey(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestPathBrowserView_Filter_NarrowsRows(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
		{Name: "config", IsDir: false},
	}})

	updated, _ := v.Update(runeKey("/"))
	pv := updated.(*PathBrowserView)
	if !pv.filter.Active {
		t.Error("expected '/' to activate the filter input")
	}

	updated, _ = pv.Update(runeKey("bnk"))
	pv = updated.(*PathBrowserView)

	if len(pv.filtered) != 1 || pv.filtered[0].Name != "bank/" {
		t.Fatalf("expected filtered [bank/], got %v", pv.filtered)
	}
	if pv.table.RowCount() != 1 {
		t.Errorf("expected 1 table row, got %d", pv.table.RowCount())
	}
	if sel := pv.selectedEntry(); sel == nil || sel.Name != "bank/" {
		t.Error("expected selected entry to come from the filtered set")
	}
}

func TestPathBrowserView_Filter_CommitKeepsFilter(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
	}})
	v.Update(runeKey("/"))
	v.Update(runeKey("ban"))

	updated, _ := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pv := updated.(*PathBrowserView)

	if pv.filter.Active {
		t.Error("expected Enter to close the filter input")
	}
	if !pv.HasFilter() {
		t.Error("expected the filter to stay applied after Enter")
	}
	if len(pv.filtered) != 1 {
		t.Errorf("expected 1 filtered entry, got %d", len(pv.filtered))
	}
}

func TestPathBrowserView_Filter_EscWhileTypingClears(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
	}})
	v.Update(runeKey("/"))
	v.Update(runeKey("ban"))

	updated, _ := v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	pv := updated.(*PathBrowserView)

	if pv.filter.Active {
		t.Error("expected Esc to close the filter input")
	}
	if pv.HasFilter() {
		t.Error("expected Esc to clear the filter")
	}
	if len(pv.filtered) != 2 {
		t.Errorf("expected all entries after clearing, got %d", len(pv.filtered))
	}
}

func TestPathBrowserView_ClearFilter(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
	}})
	v.Update(runeKey("/"))
	v.Update(runeKey("ban"))
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	v.ClearFilter()

	if v.HasFilter() || v.FilterActive() {
		t.Error("expected ClearFilter to drop the filter")
	}
	if len(v.filtered) != 2 {
		t.Errorf("expected all entries after ClearFilter, got %d", len(v.filtered))
	}
}

func TestPathBrowserView_Filter_NoMatches(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
	}})
	v.Update(runeKey("/"))
	v.Update(runeKey("zzz"))

	view := v.View(80, 20)
	if !strings.Contains(view, "No matches") {
		t.Error("expected no-matches message")
	}
	if !strings.Contains(view, "zzz") {
		t.Error("expected the filter input to render its text")
	}
}

func TestPathBrowserView_Filter_SurvivesReload(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)
	v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
	}})
	v.Update(runeKey("/"))
	v.Update(runeKey("ban"))
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	updated, _ := v.Update(pathListMsg{entries: []vault.PathEntry{
		{Name: "apps/", IsDir: true},
		{Name: "bank/", IsDir: true},
		{Name: "infra/", IsDir: true},
	}})
	pv := updated.(*PathBrowserView)

	if len(pv.filtered) != 1 || pv.filtered[0].Name != "bank/" {
		t.Errorf("expected filter to survive reload, got %v", pv.filtered)
	}
}

func TestPathBrowserView_KeyHints_IncludeFilter(t *testing.T) {
	v := NewPathBrowserView(newTestClient(t), "secret/", "", true)

	found := false
	for _, h := range v.KeyHints() {
		if h.Key == "/" && h.Desc == "filter" {
			found = true
		}
	}
	if !found {
		t.Error("expected '/' filter hint")
	}
}
