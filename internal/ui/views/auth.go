package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/milad/vaultui/internal/ui"
	"github.com/milad/vaultui/internal/ui/components"
	"github.com/milad/vaultui/internal/ui/styles"
	"github.com/milad/vaultui/internal/vault"
)

type authLoadedMsg struct {
	methods []vault.MountEntry
	err     error
}

// AuthMethodsView displays the list of enabled auth methods.
type AuthMethodsView struct {
	client  *vault.Client
	table   *components.Table
	methods []vault.MountEntry
	err     error
	loading bool
}

var _ ui.View = (*AuthMethodsView)(nil)

var authColumns = []components.Column{
	{Title: "PATH", MinWidth: 20},
	{Title: "TYPE", MinWidth: 14},
	{Title: "DESCRIPTION", MinWidth: 30, FlexFill: true},
}

func NewAuthMethodsView(client *vault.Client) *AuthMethodsView {
	return &AuthMethodsView{
		client:  client,
		table:   components.NewTable(authColumns),
		loading: true,
	}
}

func (v *AuthMethodsView) Init() tea.Cmd {
	return v.fetchMethods
}

func (v *AuthMethodsView) fetchMethods() tea.Msg {
	methods, err := v.client.ListAuthMethods()
	return authLoadedMsg{methods: methods, err: err}
}

func (v *AuthMethodsView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case authLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.methods = msg.methods
		v.table.SetRows(v.buildRows())
		return v, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, navKeys.Up):
			v.table.MoveUp()
		case key.Matches(msg, navKeys.Down):
			v.table.MoveDown()
		case key.Matches(msg, navKeys.Top):
			v.table.GoToTop()
		case key.Matches(msg, navKeys.Bottom):
			v.table.GoToBottom()
		case key.Matches(msg, navKeys.PageDown):
			v.table.PageDown()
		case key.Matches(msg, navKeys.PageUp):
			v.table.PageUp()
		}
	}

	return v, nil
}

const authTitleHeight = 2

func (v *AuthMethodsView) View(width, height int) string {
	v.table.SetSize(width, height-authTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Auth Methods")

	if v.loading {
		body := lipgloss.Place(width, height-authTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading auth methods..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-authTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if len(v.methods) == 0 {
		body := lipgloss.Place(width, height-authTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No auth methods found"))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *AuthMethodsView) Title() string {
	return "Auth Methods"
}

func (v *AuthMethodsView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func (v *AuthMethodsView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.methods))
	for i, m := range v.methods {
		rows[i] = components.Row{m.Path, m.Type, m.Description}
	}
	return rows
}
