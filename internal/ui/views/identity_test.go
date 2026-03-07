package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/vault"
)

func TestIdentityView_Title(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	if v.Title() != "Identity" {
		t.Errorf("expected title 'Identity', got %q", v.Title())
	}
}

func TestIdentityView_Init(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestIdentityView_View_Loading(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestIdentityView_Update_Loaded(t *testing.T) {
	v := NewIdentityView(newTestClient(t))

	entities := []vault.IdentityEntity{
		{ID: "id1", Name: "entity1", Policies: []string{"p1"}},
	}
	groups := []vault.IdentityGroup{
		{ID: "g1", Name: "group1", Type: "internal", Policies: []string{}},
	}

	updated, cmd := v.Update(identityLoadedMsg{entities: entities, groups: groups})
	iv := updated.(*IdentityView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if iv.loading {
		t.Error("expected loading to be false after data arrives")
	}
	if len(iv.entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(iv.entities))
	}
	if len(iv.groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(iv.groups))
	}
}

func TestIdentityView_Update_LoadedError(t *testing.T) {
	v := NewIdentityView(newTestClient(t))

	updated, _ := v.Update(identityLoadedMsg{err: errTest})
	iv := updated.(*IdentityView)

	if iv.loading {
		t.Error("expected loading to be false")
	}
	if iv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestIdentityView_View_WithData(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	v.loading = false
	v.entities = []vault.IdentityEntity{
		{ID: "id1", Name: "entity1", Policies: []string{}},
	}
	v.groups = []vault.IdentityGroup{}
	v.rebuildTable()

	view := v.View(80, 20)
	if !strings.Contains(view, "entity1") {
		t.Error("expected view to contain 'entity1'")
	}
	if !strings.Contains(view, "id1") {
		t.Error("expected view to contain 'id1'")
	}
}

func TestIdentityView_View_Error(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	v.loading = false
	v.err = errTest

	view := v.View(80, 20)
	if !strings.Contains(view, "test error") {
		t.Error("expected view to show error message")
	}
}

func TestIdentityView_View_Empty(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	v.loading = false
	v.entities = []vault.IdentityEntity{}
	v.groups = []vault.IdentityGroup{}
	v.rebuildTable()

	view := v.View(80, 20)
	if !strings.Contains(view, "No entities found") {
		t.Error("expected empty state message for entities tab")
	}

	v.tab = 1
	v.rebuildTable()
	view = v.View(80, 20)
	if !strings.Contains(view, "No groups found") {
		t.Error("expected empty state message for groups tab")
	}
}

func TestIdentityView_TabSwitch(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	v.loading = false
	v.entities = []vault.IdentityEntity{{ID: "e1", Name: "ent1"}}
	v.groups = []vault.IdentityGroup{{ID: "g1", Name: "grp1", Type: "internal"}}
	v.rebuildTable()

	if v.tab != 0 {
		t.Error("expected initial tab to be 0 (entities)")
	}

	// Tab key: use KeyTab type
	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.tab != 1 {
		t.Errorf("expected tab 1 after tab key, got %d", v.tab)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.tab != 0 {
		t.Errorf("expected tab 0 after second tab, got %d", v.tab)
	}
}

func TestIdentityView_KeyHints(t *testing.T) {
	v := NewIdentityView(newTestClient(t))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}

func TestIdentityDetailView_Title(t *testing.T) {
	v := NewIdentityDetailView(newTestClient(t), true, "id1", "entity1")
	if v.Title() != "Entity: entity1" {
		t.Errorf("expected title 'Entity: entity1', got %q", v.Title())
	}

	v2 := NewIdentityDetailView(newTestClient(t), false, "g1", "group1")
	if v2.Title() != "Group: group1" {
		t.Errorf("expected title 'Group: group1', got %q", v2.Title())
	}
}

func TestIdentityDetailView_Init(t *testing.T) {
	v := NewIdentityDetailView(newTestClient(t), true, "id1", "entity1")
	if v.Init() == nil {
		t.Error("expected Init to return a command")
	}
}

func TestIdentityDetailView_View_Loading(t *testing.T) {
	v := NewIdentityDetailView(newTestClient(t), true, "id1", "entity1")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestIdentityDetailView_Update_Loaded(t *testing.T) {
	v := NewIdentityDetailView(newTestClient(t), true, "id1", "entity1")
	entity := &vault.IdentityEntity{ID: "id1", Name: "entity1", Policies: []string{"p1", "p2"}}

	updated, _ := v.Update(identityDetailLoadedMsg{entity: entity})
	idv := updated.(*IdentityDetailView)

	if idv.loading {
		t.Error("expected loading to be false")
	}
	if idv.entity == nil || idv.entity.Name != "entity1" {
		t.Error("expected entity to be set")
	}
}

func TestIdentityDetailView_KeyHints(t *testing.T) {
	v := NewIdentityDetailView(newTestClient(t), true, "id1", "entity1")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}
