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

type enginesLoadedMsg struct {
	engines []vault.MountEntry
	err     error
}

// EnginesView displays the list of mounted secret engines.
type EnginesView struct {
	client  *vault.Client
	table   *components.Table
	engines []vault.MountEntry
	err     error
	loading bool
}

var _ ui.View = (*EnginesView)(nil)

var engineColumns = []components.Column{
	{Title: "PATH", MinWidth: 20},
	{Title: "TYPE", MinWidth: 14},
	{Title: "VERSION", MinWidth: 10},
	{Title: "DESCRIPTION", MinWidth: 30, FlexFill: true},
}

// NewEnginesView creates a new secret engines browser.
func NewEnginesView(client *vault.Client) *EnginesView {
	return &EnginesView{
		client:  client,
		table:   components.NewTable(engineColumns),
		loading: true,
	}
}

func (v *EnginesView) Init() tea.Cmd {
	return v.fetchEngines
}

func (v *EnginesView) fetchEngines() tea.Msg {
	engines, err := v.client.ListSecretEngines()
	return enginesLoadedMsg{engines: engines, err: err}
}

func (v *EnginesView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case enginesLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.engines = msg.engines
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
		}
	}

	return v, nil
}

const enginesTitleHeight = 2 // title + blank line

func (v *EnginesView) View(width, height int) string {
	v.table.SetSize(width, height-enginesTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Secret Engines")

	if v.loading {
		body := lipgloss.Place(width, height-enginesTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading secret engines..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-enginesTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if len(v.engines) == 0 {
		body := lipgloss.Place(width, height-enginesTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No secret engines found"))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *EnginesView) Title() string {
	return "Secret Engines"
}

func (v *EnginesView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "browse"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func (v *EnginesView) handleEnter() tea.Cmd {
	engine := v.SelectedEngine()
	if engine == nil {
		return nil
	}

	var next ui.View
	switch engine.Type {
	case "pki":
		next = NewPKIView(v.client, engine.Path)
	case "transit":
		next = NewTransitView(v.client, engine.Path)
	default:
		kvV2 := engine.Version == "v2"
		next = NewPathBrowserView(v.client, engine.Path, "", kvV2)
	}
	return func() tea.Msg {
		return ui.PushViewMsg{View: next}
	}
}

// SelectedEngine returns the currently highlighted engine, or nil.
func (v *EnginesView) SelectedEngine() *vault.MountEntry {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.engines) {
		return nil
	}
	return &v.engines[idx]
}

func (v *EnginesView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.engines))
	for i, e := range v.engines {
		rows[i] = components.Row{e.Path, e.Type, e.Version, e.Description}
	}
	return rows
}

// navKeys is a local keybinding set for table navigation within views.
var navKeys = struct {
	Up       key.Binding
	Down     key.Binding
	Top      key.Binding
	Bottom   key.Binding
	PageDown key.Binding
	PageUp   key.Binding
	Enter    key.Binding
}{
	Up:       key.NewBinding(key.WithKeys("k", "up")),
	Down:     key.NewBinding(key.WithKeys("j", "down")),
	Top:      key.NewBinding(key.WithKeys("g", "home")),
	Bottom:   key.NewBinding(key.WithKeys("G", "end")),
	PageDown: key.NewBinding(key.WithKeys("ctrl+d")),
	PageUp:   key.NewBinding(key.WithKeys("ctrl+u")),
	Enter:    key.NewBinding(key.WithKeys("enter")),
}
