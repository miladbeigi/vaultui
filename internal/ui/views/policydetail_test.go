package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

const testHCL = `path "secret/data/*" {
  capabilities = ["read", "list"]
}

path "secret/metadata/*" {
  capabilities = ["list"]
}`

func TestPolicyDetailView_Title(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	if v.Title() != "Policy: readonly" {
		t.Errorf("expected title 'Policy: readonly', got %q", v.Title())
	}
}

func TestPolicyDetailView_Init_ReturnsCmd(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	cmd := v.Init()
	if cmd == nil {
		t.Error("expected Init to return a fetch command")
	}
}

func TestPolicyDetailView_View_Loading(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	view := v.View(80, 20)
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading message")
	}
}

func TestPolicyDetailView_Update_Loaded(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")

	updated, cmd := v.Update(policyLoadedMsg{body: testHCL})
	pv := updated.(*PolicyDetailView)

	if cmd != nil {
		t.Error("expected no command after load")
	}
	if pv.loading {
		t.Error("expected loading to be false")
	}
	if pv.body != testHCL {
		t.Error("expected body to be set")
	}
}

func TestPolicyDetailView_Update_Error(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")

	updated, _ := v.Update(policyLoadedMsg{err: errTest})
	pv := updated.(*PolicyDetailView)

	if pv.err == nil {
		t.Error("expected error to be stored")
	}
}

func TestPolicyDetailView_View_WithData(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	v.loading = false
	v.body = testHCL

	view := v.View(80, 20)
	if !strings.Contains(view, "secret/data") {
		t.Error("expected view to contain HCL path")
	}
	if !strings.Contains(view, "capabilities") {
		t.Error("expected view to contain capabilities")
	}
	if !strings.Contains(view, "readonly") {
		t.Error("expected view to contain policy name")
	}
}

func TestPolicyDetailView_View_Empty(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "empty")
	v.loading = false
	v.body = ""

	view := v.View(80, 20)
	if !strings.Contains(view, "Empty policy") {
		t.Error("expected empty policy message")
	}
}

func TestPolicyDetailView_View_Error(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	v.loading = false
	v.err = errTest

	view := v.View(80, 20)
	if !strings.Contains(view, "test error") {
		t.Error("expected error message")
	}
}

func TestPolicyDetailView_Scroll(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	v.loading = false
	v.body = testHCL

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if v.scroll != 1 {
		t.Errorf("expected scroll 1 after j, got %d", v.scroll)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if v.scroll != 0 {
		t.Errorf("expected scroll 0 after k, got %d", v.scroll)
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if v.scroll == 0 {
		t.Error("expected scroll > 0 after G")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if v.scroll != 0 {
		t.Errorf("expected scroll 0 after g, got %d", v.scroll)
	}
}

func TestPolicyDetailView_StatusClear(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	v.statusMsg = "Copied"

	updated, _ := v.Update(policyStatusClearMsg{})
	pv := updated.(*PolicyDetailView)

	if pv.statusMsg != "" {
		t.Error("expected status to be cleared")
	}
}

func TestPolicyDetailView_KeyHints(t *testing.T) {
	v := NewPolicyDetailView(newTestClient(t), "readonly")
	hints := v.KeyHints()
	if len(hints) == 0 {
		t.Error("expected key hints to be non-empty")
	}
}
