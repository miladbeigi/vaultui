package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type transitLoadedMsg struct {
	keys []vault.TransitKey
	err  error
}

// TransitView displays keys from a Transit engine.
type TransitView struct {
	client  *vault.Client
	mount   string
	table   *components.Table
	keys    []vault.TransitKey
	err     error
	loading bool
}

var _ ui.View = (*TransitView)(nil)

var transitColumns = []components.Column{
	{Title: "KEY NAME", MinWidth: 24, FlexFill: true},
}

func NewTransitView(client *vault.Client, mount string) *TransitView {
	return &TransitView{
		client:  client,
		mount:   mount,
		table:   components.NewTable(transitColumns),
		loading: true,
	}
}

func (v *TransitView) Init() tea.Cmd {
	return v.fetchKeys
}

func (v *TransitView) fetchKeys() tea.Msg {
	keys, err := v.client.ListTransitKeys(v.mount)
	return transitLoadedMsg{keys: keys, err: err}
}

func (v *TransitView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case transitLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.keys = msg.keys
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
		case key.Matches(msg, navKeys.Enter):
			cmd := v.handleEnter()
			return v, cmd
		case msg.String() == "J":
			cmd := v.handleRawOpen(components.FormatJSON)
			return v, cmd
		case msg.String() == "y":
			cmd := v.handleRawOpen(components.FormatYAML)
			return v, cmd
		}
	}

	return v, nil
}

const transitTitleHeight = 2

func (v *TransitView) View(width, height int) string {
	v.table.SetSize(width, height-transitTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Transit: " + v.mount)

	if v.loading {
		body := lipgloss.Place(width, height-transitTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading transit keys..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-transitTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if len(v.keys) == 0 {
		body := lipgloss.Place(width, height-transitTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No transit keys found"))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *TransitView) Title() string {
	return "Transit: " + v.mount
}

func (v *TransitView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "view key"},
		{Key: "J/y", Desc: "raw view"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *TransitView) handleRawOpen(format components.RawFormat) tea.Cmd {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.keys) {
		return nil
	}
	next := NewTransitKeyDetailView(v.client, v.mount, v.keys[idx].Name)
	next.SetInitialRawFormat(format)
	return func() tea.Msg { return ui.PushViewMsg{View: next} }
}

func (v *TransitView) handleEnter() tea.Cmd {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.keys) {
		return nil
	}
	next := NewTransitKeyDetailView(v.client, v.mount, v.keys[idx].Name)
	return func() tea.Msg { return ui.PushViewMsg{View: next} }
}

func (v *TransitView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.keys))
	for i, k := range v.keys {
		rows[i] = components.Row{k.Name}
	}
	return rows
}
