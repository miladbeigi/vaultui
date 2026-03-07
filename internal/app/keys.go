package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"

	"github.com/miladbeigi/vaultui/internal/config"
)

// KeyMap defines the global keybindings for the application.
type KeyMap struct {
	Quit      key.Binding
	ForceQuit key.Binding
	Help      key.Binding
	Command   key.Binding
	Filter    key.Binding

	// Navigation
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Top      key.Binding
	Bottom   key.Binding
	PageDown key.Binding
	PageUp   key.Binding

	// Quick-jump (1-6)
	Jump1 key.Binding
	Jump2 key.Binding
	Jump3 key.Binding
	Jump4 key.Binding
	Jump5 key.Binding
	Jump6 key.Binding
}

var keys = defaultKeyMap()

func defaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),

		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("⏎", "open"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "left"),
			key.WithHelp("esc/←", "back"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "page up"),
		),

		Jump1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "secret engines")),
		Jump2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "auth methods")),
		Jump3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "policies")),
		Jump4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "leases")),
		Jump5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "identity")),
		Jump6: key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "sys/config")),
	}
}

// ApplyKeybindings overrides default key assignments from config.
func ApplyKeybindings(kb config.Keybindings) {
	overrideBinding(&keys.Up, kb.Up)
	overrideBinding(&keys.Down, kb.Down)
	overrideBinding(&keys.Top, kb.Top)
	overrideBinding(&keys.Bottom, kb.Bottom)
	overrideBinding(&keys.PageDown, kb.PageDown)
	overrideBinding(&keys.PageUp, kb.PageUp)
	overrideBinding(&keys.Enter, kb.Enter)
	overrideBinding(&keys.Back, kb.Back)
	overrideBinding(&keys.Quit, kb.Quit)
	overrideBinding(&keys.Command, kb.Command)
}

func overrideBinding(b *key.Binding, keyStr string) {
	if keyStr == "" {
		return
	}
	parts := strings.Split(keyStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	b.SetKeys(parts...)
}
