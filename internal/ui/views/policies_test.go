package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPoliciesView_Title(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	if v.Title() != "Policies" {
		t.Errorf("expected title 'Policies', got %q", v.Title())
	}
}

func TestPoliciesView_Init_ReturnsCmd(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestPoliciesView_View_Loading(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestPoliciesView_Update_LoadedMsg(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))

	policies := []string{"default", "admin", "readonly", "root"}

	updated, cmd := v.Update(policiesLoadedMsg{policies: policies})
	pv := updated.(*PoliciesView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if pv.loading {
		t.Error("expected loading to be false")
	}
	if len(pv.policies) != 4 {
		t.Errorf("expected 4 policies, got %d", len(pv.policies))
	}
}

func TestPoliciesView_Update_LoadedError(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))

	updated, _ := v.Update(policiesLoadedMsg{err: errTest})
	pv := updated.(*PoliciesView)

	if pv.loading {
		t.Error("expected loading to be false")
	}
	if pv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestPoliciesView_View_WithData(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	v.loading = false
	v.policies = []string{"default", "admin", "root"}
	v.table.SetRows(v.buildRows())

	view := v.View(80, 20)
	if !strings.Contains(view, "default") {
		t.Error("expected view to contain 'default'")
	}
	if !strings.Contains(view, "admin") {
		t.Error("expected view to contain 'admin'")
	}
	if !strings.Contains(view, "root") {
		t.Error("expected view to contain 'root'")
	}
}

func TestPoliciesView_View_Error(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	v.loading = false
	v.err = errTest

	view := v.View(80, 20)
	if !strings.Contains(view, "test error") {
		t.Error("expected view to show error message")
	}
}

func TestPoliciesView_View_Empty(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	v.loading = false
	v.policies = []string{}

	view := v.View(80, 20)
	if !strings.Contains(view, "No policies") {
		t.Error("expected empty state message")
	}
}

func TestPoliciesView_Navigation(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	v.loading = false
	v.policies = []string{"default", "admin", "readonly"}
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
}

func TestPoliciesView_SelectedPolicy(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	v.loading = false
	v.policies = []string{"default", "admin"}
	v.table.SetRows(v.buildRows())

	if v.selectedPolicy() != "default" {
		t.Error("expected first policy to be selected")
	}

	v.table.MoveDown()
	if v.selectedPolicy() != "admin" {
		t.Error("expected second policy after move down")
	}
}

func TestPoliciesView_BuildRows_RootType(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	v.policies = []string{"default", "root"}
	rows := v.buildRows()

	if rows[0][1] != "acl" {
		t.Errorf("expected 'default' type to be 'acl', got %q", rows[0][1])
	}
	if rows[1][1] != "root" {
		t.Errorf("expected 'root' type to be 'root', got %q", rows[1][1])
	}
}

func TestPoliciesView_KeyHints(t *testing.T) {
	v := NewPoliciesView(newTestClient(t))
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}
