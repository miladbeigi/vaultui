package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern string
		s       string
		want    bool
	}{
		{"", "anything", true},
		{"bank", "bank/", true},
		{"bnk", "bank/", true},
		{"BANK", "bank/", true},
		{"bank", "BANK/", true},
		{"my", "bank/", false},
		{"bkna", "bank", false},
		{"bank", "kna", false},
	}

	for _, tt := range tests {
		if got := Match(tt.pattern, tt.s); got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.s, got, tt.want)
		}
	}
}

func TestFilter(t *testing.T) {
	type item struct{ name string }
	items := []item{{"apps/"}, {"bank/"}, {"config"}}
	key := func(i item) string { return i.name }

	got := Filter(items, "", key)
	if len(got) != 3 {
		t.Errorf("expected empty pattern to keep all items, got %d", len(got))
	}

	got = Filter(items, "bnk", key)
	if len(got) != 1 || got[0].name != "bank/" {
		t.Errorf("expected [bank/], got %v", got)
	}

	got = Filter(items, "zzz", key)
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestFilterBar_Update_Typing(t *testing.T) {
	f := FilterBar{Active: true}

	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'n', 'k'}})

	if f.Text != "b ank" {
		t.Errorf("expected 'b ank', got %q", f.Text)
	}
	if !f.Active {
		t.Error("expected filter to stay active while typing")
	}

	f.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if f.Text != "b an" {
		t.Errorf("expected backspace to trim the input, got %q", f.Text)
	}
}

func TestFilterBar_Update_EnterCommits(t *testing.T) {
	f := FilterBar{Active: true, Text: "bank"}

	f.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if f.Active {
		t.Error("expected Enter to close the input")
	}
	if f.Text != "bank" {
		t.Errorf("expected Enter to keep the filter applied, got %q", f.Text)
	}
}

func TestFilterBar_Update_EscClears(t *testing.T) {
	f := FilterBar{Active: true, Text: "bank"}

	f.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if f.Active {
		t.Error("expected Esc to close the input")
	}
	if f.Text != "" {
		t.Errorf("expected Esc to clear the filter, got %q", f.Text)
	}
}

func TestFilterBar_Update_CtrlUClearsLine(t *testing.T) {
	f := FilterBar{Active: true, Text: "bank"}

	f.Update(tea.KeyMsg{Type: tea.KeyCtrlU})

	if f.Text != "" {
		t.Errorf("expected ctrl+u to clear the line, got %q", f.Text)
	}
	if !f.Active {
		t.Error("expected ctrl+u to keep the input active")
	}
}

func TestFilterBar_View(t *testing.T) {
	f := FilterBar{Active: true, Text: "bank"}

	view := f.View(40)
	if !strings.Contains(view, "bank") {
		t.Error("expected view to contain the filter text")
	}
	if !strings.Contains(view, "/") {
		t.Error("expected view to contain the '/' prompt")
	}
}
