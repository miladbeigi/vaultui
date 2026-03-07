package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// HomeView is the landing screen shown when VaultUI starts.
type HomeView struct{}

// Compile-time check that HomeView implements ui.View.
var _ ui.View = (*HomeView)(nil)

// NewHomeView creates a new home view.
func NewHomeView() *HomeView {
	return &HomeView{}
}

func (v *HomeView) Init() tea.Cmd {
	return nil
}

func (v *HomeView) Update(_ tea.Msg) (ui.View, tea.Cmd) {
	return v, nil
}

func (v *HomeView) View(width, height int) string {
	msg := styles.SubtleStyle.Render("Welcome to VaultUI\n\nPress : for commands, 1 for secret engines, q to quit")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
}

func (v *HomeView) Title() string {
	return "Home"
}

func (v *HomeView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: ":", Desc: "command"},
		{Key: "1", Desc: "secret engines"},
		{Key: "q", Desc: "quit"},
	}
}
