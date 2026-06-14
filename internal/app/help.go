package app

import (
	"github.com/charmbracelet/bubbles/key"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
)

func globalHelpSections() []components.HelpSection {
	return []components.HelpSection{
		{
			Title: "General",
			Hints: bindingsToHints(keys.Quit, keys.ForceQuit, keys.Help, keys.Command),
		},
		{
			Title: "Navigation",
			Hints: bindingsToHints(
				keys.Up, keys.Down, keys.Enter, keys.Back,
				keys.Top, keys.Bottom, keys.PageDown, keys.PageUp,
			),
		},
		{
			Title: "Quick Jump",
			Hints: bindingsToHints(
				keys.Jump1, keys.Jump2, keys.Jump3, keys.Jump4,
				keys.Jump5, keys.Jump6, keys.Jump7, keys.Jump8,
			),
		},
	}
}

func (m Model) buildHelpOverlay() components.HelpOverlay {
	sections := append([]components.HelpSection{}, globalHelpSections()...)

	if current := m.router.Current(); current != nil {
		hints := current.KeyHints()
		if len(hints) > 0 {
			sections = append(sections, components.HelpSection{
				Title: "This View",
				Hints: hints,
			})
		}
	}

	return components.HelpOverlay{Sections: sections}
}

func bindingsToHints(bindings ...key.Binding) []ui.KeyHint {
	hints := make([]ui.KeyHint, 0, len(bindings))
	for _, b := range bindings {
		h := b.Help()
		hints = append(hints, ui.KeyHint{Key: h.Key, Desc: h.Desc})
	}
	return hints
}
