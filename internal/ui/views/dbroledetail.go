package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type dbRoleDetailLoadedMsg struct {
	detail *vault.DBRoleDetail
	err    error
}

// DBRoleDetailView shows the full configuration of a dynamic database role.
type DBRoleDetailView struct {
	client  *vault.Client
	mount   string
	name    string
	detail  *vault.DBRoleDetail
	table   *components.Table
	err     error
	loading bool
}

var _ ui.View = (*DBRoleDetailView)(nil)

var dbRoleDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewDBRoleDetailView creates a detail view for a dynamic database role.
func NewDBRoleDetailView(client *vault.Client, mount, name string) *DBRoleDetailView {
	return &DBRoleDetailView{
		client:  client,
		mount:   mount,
		name:    name,
		table:   components.NewTable(dbRoleDetailColumns),
		loading: true,
	}
}

func (v *DBRoleDetailView) Init() tea.Cmd {
	return v.fetchData
}

func (v *DBRoleDetailView) fetchData() tea.Msg {
	detail, err := v.client.ReadDBRole(v.mount, v.name)
	return dbRoleDetailLoadedMsg{detail: detail, err: err}
}

func (v *DBRoleDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case dbRoleDetailLoadedMsg:
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

const dbRoleDetailTitleHeight = 2

func (v *DBRoleDetailView) View(width, height int) string {
	v.table.SetSize(width, height-dbRoleDetailTitleHeight)

	titleLine := styles.ViewTitleStyle.Width(width).Render("Role: " + v.name)

	if v.loading {
		body := lipgloss.Place(width, height-dbRoleDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading role details..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-dbRoleDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *DBRoleDetailView) Title() string {
	return "Role: " + v.name
}

func (v *DBRoleDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *DBRoleDetailView) buildRows() []components.Row {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	rows := []components.Row{
		{"Name", d.Name},
		{"DB Name", d.DBName},
		{"Default TTL", dbValOrDash(d.DefaultTTL)},
		{"Max TTL", dbValOrDash(d.MaxTTL)},
		{"Role Type", dbValOrDash(d.RoleType)},
	}
	if len(d.CreationStatements) > 0 {
		rows = append(rows, components.Row{"Creation Statements", strings.Join(d.CreationStatements, "\n")})
	}
	if len(d.RevocationStatements) > 0 {
		rows = append(rows, components.Row{"Revocation Statements", strings.Join(d.RevocationStatements, "\n")})
	}
	if len(d.RollbackStatements) > 0 {
		rows = append(rows, components.Row{"Rollback Statements", strings.Join(d.RollbackStatements, "\n")})
	}
	if len(d.RenewStatements) > 0 {
		rows = append(rows, components.Row{"Renew Statements", strings.Join(d.RenewStatements, "\n")})
	}
	return rows
}
