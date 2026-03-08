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

type identityLoadedMsg struct {
	entities []vault.IdentityEntity
	groups   []vault.IdentityGroup
	err      error
}

// IdentityView displays entities and groups from the Identity engine.
type IdentityView struct {
	client   *vault.Client
	table    *components.Table
	entities []vault.IdentityEntity
	groups   []vault.IdentityGroup
	err      error
	loading  bool
	tab      int // 0 = entities, 1 = groups
}

var _ ui.View = (*IdentityView)(nil)

var identityEntityColumns = []components.Column{
	{Title: "NAME", MinWidth: 24},
	{Title: "ID", MinWidth: 40, FlexFill: true},
}

var identityGroupColumns = []components.Column{
	{Title: "NAME", MinWidth: 24},
	{Title: "ID", MinWidth: 40},
	{Title: "TYPE", MinWidth: 12, FlexFill: true},
}

// NewIdentityView creates a new identity browser view.
func NewIdentityView(client *vault.Client) *IdentityView {
	return &IdentityView{
		client:  client,
		table:   components.NewTable(identityEntityColumns),
		loading: true,
	}
}

func (v *IdentityView) Init() tea.Cmd {
	return v.fetchData
}

func (v *IdentityView) fetchData() tea.Msg {
	entities, err := v.client.ListIdentityEntities("")
	if err != nil {
		return identityLoadedMsg{err: err}
	}
	groups, _ := v.client.ListIdentityGroups()
	return identityLoadedMsg{entities: entities, groups: groups}
}

func (v *IdentityView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case identityLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.entities = msg.entities
		v.groups = msg.groups
		v.rebuildTable()
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
		case msg.String() == "tab":
			v.tab = (v.tab + 1) % 2
			v.rebuildTable()
		case key.Matches(msg, navKeys.Enter):
			cmd := v.handleEnter()
			return v, cmd
		}
	}

	return v, nil
}

func (v *IdentityView) rebuildTable() {
	if v.tab == 0 {
		v.table = components.NewTable(identityEntityColumns)
		v.table.SetRows(v.buildEntityRows())
	} else {
		v.table = components.NewTable(identityGroupColumns)
		v.table.SetRows(v.buildGroupRows())
	}
}

func (v *IdentityView) handleEnter() tea.Cmd {
	if v.tab == 0 {
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.entities) {
			return nil
		}
		next := NewIdentityDetailView(v.client, true, v.entities[idx].ID, v.entities[idx].Name)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	}
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.groups) {
		return nil
	}
	next := NewIdentityDetailView(v.client, false, v.groups[idx].ID, v.groups[idx].Name)
	return func() tea.Msg { return ui.PushViewMsg{View: next} }
}

func (v *IdentityView) buildEntityRows() []components.Row {
	rows := make([]components.Row, len(v.entities))
	for i, e := range v.entities {
		rows[i] = components.Row{e.Name, e.ID}
	}
	return rows
}

func (v *IdentityView) buildGroupRows() []components.Row {
	rows := make([]components.Row, len(v.groups))
	for i, g := range v.groups {
		rows[i] = components.Row{g.Name, g.ID, g.Type}
	}
	return rows
}

const identityTitleHeight = 2

func (v *IdentityView) View(width, height int) string {
	v.table.SetSize(width, height-identityTitleHeight)

	tabNames := []string{"Entities", "Groups"}
	tabs := ""
	for i, name := range tabNames {
		if i == v.tab {
			tabs += styles.SecondaryStyle.Render("["+name+"]") + "  "
		} else {
			tabs += styles.SubtleStyle.Render(" "+name+" ") + "  "
		}
	}
	title := lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(
		styles.ViewTitleStyle.Render("Identity") + "  " + tabs)

	if v.loading {
		body := lipgloss.Place(width, height-identityTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading identity data..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-identityTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	emptyMsg := "No entities found"
	if v.tab == 1 {
		emptyMsg = "No groups found"
	}
	if (v.tab == 0 && len(v.entities) == 0) || (v.tab == 1 && len(v.groups) == 0) {
		body := lipgloss.Place(width, height-identityTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render(emptyMsg))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *IdentityView) Title() string {
	return "Identity"
}

func (v *IdentityView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "tab", Desc: "switch tab"},
		{Key: "⏎", Desc: "view"},
		{Key: "esc", Desc: "back"},
	}
}
