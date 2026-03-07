package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/miladbeigi/vaultui/internal/ui"
)

// stubView is a minimal ui.View implementation for testing.
type stubView struct {
	title string
}

func (v *stubView) Init() tea.Cmd                       { return nil }
func (v *stubView) Update(_ tea.Msg) (ui.View, tea.Cmd) { return v, nil }
func (v *stubView) View(_, _ int) string                { return v.title }
func (v *stubView) Title() string                       { return v.title }
func (v *stubView) KeyHints() []ui.KeyHint              { return nil }

func TestRouter_Empty(t *testing.T) {
	r := NewRouter()

	if r.Depth() != 0 {
		t.Errorf("expected depth 0, got %d", r.Depth())
	}
	if r.Current() != nil {
		t.Error("expected nil current view on empty router")
	}
}

func TestRouter_Push(t *testing.T) {
	r := NewRouter()
	v1 := &stubView{title: "view1"}
	v2 := &stubView{title: "view2"}

	r.Push(v1)
	if r.Depth() != 1 {
		t.Errorf("expected depth 1, got %d", r.Depth())
	}
	if r.Current().Title() != "view1" {
		t.Errorf("expected current 'view1', got %q", r.Current().Title())
	}

	r.Push(v2)
	if r.Depth() != 2 {
		t.Errorf("expected depth 2, got %d", r.Depth())
	}
	if r.Current().Title() != "view2" {
		t.Errorf("expected current 'view2', got %q", r.Current().Title())
	}
}

func TestRouter_Pop(t *testing.T) {
	r := NewRouter()
	v1 := &stubView{title: "view1"}
	v2 := &stubView{title: "view2"}

	r.Push(v1)
	r.Push(v2)

	ok := r.Pop()
	if !ok {
		t.Error("expected Pop to succeed")
	}
	if r.Depth() != 1 {
		t.Errorf("expected depth 1 after pop, got %d", r.Depth())
	}
	if r.Current().Title() != "view1" {
		t.Errorf("expected current 'view1' after pop, got %q", r.Current().Title())
	}
}

func TestRouter_Pop_RootProtection(t *testing.T) {
	r := NewRouter()
	r.Push(&stubView{title: "root"})

	ok := r.Pop()
	if ok {
		t.Error("expected Pop to fail on root view")
	}
	if r.Depth() != 1 {
		t.Errorf("expected depth to remain 1, got %d", r.Depth())
	}
}

func TestRouter_Pop_Empty(t *testing.T) {
	r := NewRouter()

	ok := r.Pop()
	if ok {
		t.Error("expected Pop to fail on empty stack")
	}
}

func TestRouter_Replace(t *testing.T) {
	r := NewRouter()
	v1 := &stubView{title: "view1"}
	v2 := &stubView{title: "view2"}

	r.Push(v1)
	r.Replace(v2)

	if r.Depth() != 1 {
		t.Errorf("expected depth 1 after replace, got %d", r.Depth())
	}
	if r.Current().Title() != "view2" {
		t.Errorf("expected current 'view2' after replace, got %q", r.Current().Title())
	}
}

func TestRouter_Replace_Empty(t *testing.T) {
	r := NewRouter()
	v := &stubView{title: "view1"}

	r.Replace(v)

	if r.Depth() != 1 {
		t.Errorf("expected depth 1 after replace on empty, got %d", r.Depth())
	}
	if r.Current().Title() != "view1" {
		t.Errorf("expected current 'view1', got %q", r.Current().Title())
	}
}

func TestRouter_PreservesState(t *testing.T) {
	r := NewRouter()
	v1 := &stubView{title: "view1"}
	v2 := &stubView{title: "view2"}
	v3 := &stubView{title: "view3"}

	r.Push(v1)
	r.Push(v2)
	r.Push(v3)

	if r.Depth() != 3 {
		t.Errorf("expected depth 3, got %d", r.Depth())
	}

	r.Pop()
	if r.Current().Title() != "view2" {
		t.Errorf("expected 'view2' after first pop, got %q", r.Current().Title())
	}

	r.Pop()
	if r.Current().Title() != "view1" {
		t.Errorf("expected 'view1' after second pop, got %q", r.Current().Title())
	}
}
