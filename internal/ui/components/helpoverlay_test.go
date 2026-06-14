package components

import (
	"strings"
	"testing"

	"github.com/miladbeigi/vaultui/internal/ui"
)

func TestHelpOverlay_View_ShowsTitle(t *testing.T) {
	overlay := HelpOverlay{
		Sections: []HelpSection{
			{Title: "General", Hints: []ui.KeyHint{{Key: "?", Desc: "help"}}},
		},
	}

	out := overlay.View(80, 30)
	if !strings.Contains(out, "Keyboard Shortcuts") {
		t.Error("expected title in help overlay output")
	}
}

func TestHelpOverlay_View_ShowsSections(t *testing.T) {
	overlay := HelpOverlay{
		Sections: []HelpSection{
			{
				Title: "General",
				Hints: []ui.KeyHint{
					{Key: "?", Desc: "help"},
					{Key: ":", Desc: "command"},
				},
			},
			{
				Title: "This View",
				Hints: []ui.KeyHint{{Key: "c", Desc: "copy value"}},
			},
		},
	}

	out := overlay.View(80, 40)
	if !strings.Contains(out, "General") {
		t.Error("expected General section title")
	}
	if !strings.Contains(out, "This View") {
		t.Error("expected This View section title")
	}
	if !strings.Contains(out, "help") {
		t.Error("expected help hint description")
	}
	if !strings.Contains(out, "copy value") {
		t.Error("expected contextual hint description")
	}
	if !strings.Contains(out, "─") {
		t.Error("expected section dividers in help overlay")
	}
}

func TestHelpOverlay_View_ShowsCloseHint(t *testing.T) {
	overlay := HelpOverlay{Sections: []HelpSection{{Title: "General", Hints: []ui.KeyHint{{Key: "q", Desc: "quit"}}}}}

	out := overlay.View(60, 20)
	if !strings.Contains(out, "esc") || !strings.Contains(out, "?") {
		t.Error("expected close hint mentioning esc and ?")
	}
}

func TestHelpOverlay_View_NarrowWidth(t *testing.T) {
	overlay := HelpOverlay{
		Sections: []HelpSection{
			{Title: "General", Hints: []ui.KeyHint{{Key: "q", Desc: "quit"}}},
		},
	}

	out := overlay.View(40, 20)
	if out == "" {
		t.Error("expected non-empty output at narrow width")
	}
}

func TestHelpOverlay_View_SkipsEmptySections(t *testing.T) {
	overlay := HelpOverlay{
		Sections: []HelpSection{
			{Title: "Empty", Hints: nil},
			{Title: "General", Hints: []ui.KeyHint{{Key: "q", Desc: "quit"}}},
		},
	}

	out := overlay.View(80, 30)
	if strings.Contains(out, "Empty") {
		t.Error("expected empty sections to be omitted")
	}
}
