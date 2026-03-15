package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type dbConnDetailLoadedMsg struct {
	detail *vault.DBConnectionDetail
	err    error
}

// DBConnectionDetailView shows the full configuration of a database connection.
type DBConnectionDetailView struct {
	client  *vault.Client
	mount   string
	name    string
	detail  *vault.DBConnectionDetail
	table   *components.Table
	err     error
	loading bool
}

var _ ui.View = (*DBConnectionDetailView)(nil)

var dbConnDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewDBConnectionDetailView creates a detail view for a database connection.
func NewDBConnectionDetailView(client *vault.Client, mount, name string) *DBConnectionDetailView {
	return &DBConnectionDetailView{
		client:  client,
		mount:   mount,
		name:    name,
		table:   components.NewTable(dbConnDetailColumns),
		loading: true,
	}
}

func (v *DBConnectionDetailView) Init() tea.Cmd {
	return v.fetchData
}

func (v *DBConnectionDetailView) fetchData() tea.Msg {
	detail, err := v.client.ReadDBConnection(v.mount, v.name)
	return dbConnDetailLoadedMsg{detail: detail, err: err}
}

func (v *DBConnectionDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case dbConnDetailLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.detail = msg.detail
		v.table.SetRows(v.buildRows())
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			v.table.MoveDown()
		case "k", "up":
			v.table.MoveUp()
		case "g", "home":
			v.table.GoToTop()
		case "G", "end":
			v.table.GoToBottom()
		case "r":
			v.loading = true
			return v, v.fetchData
		}
	}

	return v, nil
}

const dbConnDetailTitleHeight = 2

func (v *DBConnectionDetailView) View(width, height int) string {
	v.table.SetSize(width, height-dbConnDetailTitleHeight)

	titleLine := styles.ViewTitleStyle.Width(width).Render("Connection: " + v.name)

	if v.loading {
		body := lipgloss.Place(width, height-dbConnDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading connection details..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-dbConnDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *DBConnectionDetailView) Title() string {
	return "Connection: " + v.name
}

func (v *DBConnectionDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *DBConnectionDetailView) buildRows() []components.Row {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	rows := []components.Row{
		{"Name", d.Name},
		{"Plugin", d.PluginName},
		{"Connection URL", dbValOrDash(d.ConnectionURL)},
		{"Allowed Roles", dbValOrDash(strings.Join(d.AllowedRoles, ", "))},
		{"Verify Connection", fmt.Sprintf("%v", d.VerifyConnection)},
		{"Password Policy", dbValOrDash(d.PasswordPolicy)},
	}
	if len(d.RootRotationStatements) > 0 {
		rows = append(rows, components.Row{"Root Rotation Stmts", strings.Join(d.RootRotationStatements, "; ")})
	}
	return rows
}

func dbValOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
