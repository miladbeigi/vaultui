package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type dbLoadedMsg struct {
	conns       []vault.DBConnection
	roles       []vault.DBRole
	staticRoles []vault.DBStaticRole
	err         error
}

// DatabaseView displays connections, roles, and static roles for a database engine.
type DatabaseView struct {
	client      *vault.Client
	mount       string
	table       *components.Table
	conns       []vault.DBConnection
	roles       []vault.DBRole
	staticRoles []vault.DBStaticRole
	err         error
	loading     bool
	tab         int // 0 = connections, 1 = roles, 2 = static roles
}

var _ ui.View = (*DatabaseView)(nil)

var dbConnColumns = []components.Column{
	{Title: "NAME", MinWidth: 24},
	{Title: "PLUGIN", MinWidth: 30},
	{Title: "ALLOWED ROLES", MinWidth: 20, FlexFill: true},
}

var dbRoleColumns = []components.Column{
	{Title: "NAME", MinWidth: 24},
	{Title: "DB NAME", MinWidth: 24},
	{Title: "DEFAULT TTL", MinWidth: 14},
	{Title: "MAX TTL", MinWidth: 14, FlexFill: true},
}

var dbStaticRoleColumns = []components.Column{
	{Title: "NAME", MinWidth: 24},
	{Title: "DB NAME", MinWidth: 24},
	{Title: "USERNAME", MinWidth: 20},
	{Title: "ROTATION PERIOD", MinWidth: 18, FlexFill: true},
}

// NewDatabaseView creates a new database engine browser.
func NewDatabaseView(client *vault.Client, mount string) *DatabaseView {
	return &DatabaseView{
		client:  client,
		mount:   mount,
		table:   components.NewTable(dbConnColumns),
		loading: true,
	}
}

func (v *DatabaseView) Init() tea.Cmd {
	return v.fetchData
}

func (v *DatabaseView) fetchData() tea.Msg {
	conns, err := v.client.ListDBConnections(v.mount)
	if err != nil {
		return dbLoadedMsg{err: err}
	}
	roles, _ := v.client.ListDBRoles(v.mount)
	staticRoles, _ := v.client.ListDBStaticRoles(v.mount)
	return dbLoadedMsg{conns: conns, roles: roles, staticRoles: staticRoles}
}

func (v *DatabaseView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case dbLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.conns = msg.conns
		v.roles = msg.roles
		v.staticRoles = msg.staticRoles
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
			v.tab = (v.tab + 1) % 3
			v.rebuildTable()
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

func (v *DatabaseView) rebuildTable() {
	switch v.tab {
	case 0:
		v.table = components.NewTable(dbConnColumns)
		v.table.SetRows(v.buildConnRows())
	case 1:
		v.table = components.NewTable(dbRoleColumns)
		v.table.SetRows(v.buildRoleRows())
	case 2:
		v.table = components.NewTable(dbStaticRoleColumns)
		v.table.SetRows(v.buildStaticRoleRows())
	}
}

func (v *DatabaseView) handleRawOpen(format components.RawFormat) tea.Cmd {
	switch v.tab {
	case 0:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.conns) {
			return nil
		}
		next := NewDBConnectionDetailView(v.client, v.mount, v.conns[idx].Name)
		next.SetInitialRawFormat(format)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	case 1:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.roles) {
			return nil
		}
		next := NewDBRoleDetailView(v.client, v.mount, v.roles[idx].Name)
		next.SetInitialRawFormat(format)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	case 2:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.staticRoles) {
			return nil
		}
		next := NewDBStaticRoleDetailView(v.client, v.mount, v.staticRoles[idx].Name)
		next.SetInitialRawFormat(format)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	}
	return nil
}

func (v *DatabaseView) handleEnter() tea.Cmd {
	switch v.tab {
	case 0:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.conns) {
			return nil
		}
		next := NewDBConnectionDetailView(v.client, v.mount, v.conns[idx].Name)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	case 1:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.roles) {
			return nil
		}
		next := NewDBRoleDetailView(v.client, v.mount, v.roles[idx].Name)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	case 2:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.staticRoles) {
			return nil
		}
		next := NewDBStaticRoleDetailView(v.client, v.mount, v.staticRoles[idx].Name)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	}
	return nil
}

func (v *DatabaseView) buildConnRows() []components.Row {
	rows := make([]components.Row, len(v.conns))
	for i, c := range v.conns {
		rows[i] = components.Row{c.Name, c.PluginName, strings.Join(c.AllowedRoles, ", ")}
	}
	return rows
}

func (v *DatabaseView) buildRoleRows() []components.Row {
	rows := make([]components.Row, len(v.roles))
	for i, r := range v.roles {
		rows[i] = components.Row{r.Name, r.DBName, r.DefaultTTL, r.MaxTTL}
	}
	return rows
}

func (v *DatabaseView) buildStaticRoleRows() []components.Row {
	rows := make([]components.Row, len(v.staticRoles))
	for i, r := range v.staticRoles {
		rows[i] = components.Row{r.Name, r.DBName, r.Username, r.RotationPeriod}
	}
	return rows
}

const dbTitleHeight = 2

func (v *DatabaseView) View(width, height int) string {
	v.table.SetSize(width, height-dbTitleHeight)

	tabNames := []string{"Connections", "Roles", "Static Roles"}
	tabs := ""
	for i, name := range tabNames {
		if i == v.tab {
			tabs += styles.SecondaryStyle.Render("["+name+"]") + "  "
		} else {
			tabs += styles.SubtleStyle.Render(" "+name+" ") + "  "
		}
	}
	title := lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(
		styles.ViewTitleStyle.Render("Database: "+v.mount) + "  " + tabs)

	if v.loading {
		body := lipgloss.Place(width, height-dbTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading database data..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-dbTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	empty := (v.tab == 0 && len(v.conns) == 0) ||
		(v.tab == 1 && len(v.roles) == 0) ||
		(v.tab == 2 && len(v.staticRoles) == 0)
	if empty {
		msgs := []string{"No connections found", "No roles found", "No static roles found"}
		body := lipgloss.Place(width, height-dbTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render(msgs[v.tab]))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *DatabaseView) Title() string {
	return "Database: " + v.mount
}

func (v *DatabaseView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "tab", Desc: "switch tab"},
		{Key: "⏎", Desc: "view"},
		{Key: "J/y", Desc: "raw view"},
		{Key: "esc", Desc: "back"},
	}
}
