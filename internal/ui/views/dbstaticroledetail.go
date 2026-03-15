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

type dbStaticRoleDetailLoadedMsg struct {
	detail *vault.DBStaticRoleDetail
	err    error
}

// DBStaticRoleDetailView shows the full configuration of a static database role.
type DBStaticRoleDetailView struct {
	client           *vault.Client
	mount            string
	name             string
	detail           *vault.DBStaticRoleDetail
	table            *components.Table
	rawView          *components.RawView
	err              error
	loading          bool
	rawMode          bool
	pendingRawFormat *components.RawFormat
}

var _ ui.View = (*DBStaticRoleDetailView)(nil)

var dbStaticRoleDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewDBStaticRoleDetailView creates a detail view for a static database role.
func NewDBStaticRoleDetailView(client *vault.Client, mount, name string) *DBStaticRoleDetailView {
	return &DBStaticRoleDetailView{
		client:  client,
		mount:   mount,
		name:    name,
		table:   components.NewTable(dbStaticRoleDetailColumns),
		loading: true,
	}
}

func (v *DBStaticRoleDetailView) SetInitialRawFormat(format components.RawFormat) {
	v.pendingRawFormat = &format
}

func (v *DBStaticRoleDetailView) Init() tea.Cmd {
	return v.fetchData
}

func (v *DBStaticRoleDetailView) fetchData() tea.Msg {
	detail, err := v.client.ReadDBStaticRole(v.mount, v.name)
	return dbStaticRoleDetailLoadedMsg{detail: detail, err: err}
}

func (v *DBStaticRoleDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case dbStaticRoleDetailLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.detail = msg.detail
		v.table.SetRows(v.buildRows())
		if v.pendingRawFormat != nil {
			v.toggleRaw(*v.pendingRawFormat)
			v.pendingRawFormat = nil
		}
		return v, nil

	case tea.KeyMsg:
		if v.rawMode {
			switch msg.String() {
			case "j", "down":
				v.rawView.ScrollDown()
			case "k", "up":
				v.rawView.ScrollUp()
			case "g", "home":
				v.rawView.GoToTop()
			case "G", "end":
				v.rawView.GoToBottom()
			case "ctrl+d":
				v.rawView.PageDown()
			case "ctrl+u":
				v.rawView.PageUp()
			case "c":
				if err := v.rawView.CopyContent(); err != nil {
					v.rawView.Status = "✗ " + err.Error()
				} else {
					v.rawView.Status = "✓ Copied " + v.rawView.FormatLabel() + " to clipboard"
				}
			case "J":
				v.toggleRaw(components.FormatJSON)
			case "y":
				v.toggleRaw(components.FormatYAML)
			case "r":
				v.rawMode = false
				v.loading = true
				return v, v.fetchData
			case "esc":
				v.rawMode = false
				return v, nil
			}
			return v, nil
		}

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
		case "J":
			v.toggleRaw(components.FormatJSON)
		case "y":
			v.toggleRaw(components.FormatYAML)
		}
	}

	return v, nil
}

const dbStaticRoleDetailTitleHeight = 2

func (v *DBStaticRoleDetailView) View(width, height int) string {
	v.table.SetSize(width, height-dbStaticRoleDetailTitleHeight)

	titleLine := styles.ViewTitleStyle.Width(width).Render("Static Role: " + v.name)

	if v.loading {
		body := lipgloss.Place(width, height-dbStaticRoleDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading static role details..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-dbStaticRoleDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.rawMode && v.rawView != nil {
		v.rawView.SetSize(width, height-dbStaticRoleDetailTitleHeight)
		rawTitle := titleLine + "  " + styles.SecondaryStyle.Render("["+v.rawView.FormatLabel()+"]")
		return lipgloss.JoinVertical(lipgloss.Left, rawTitle, v.rawView.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *DBStaticRoleDetailView) Title() string {
	return "Static Role: " + v.name
}

func (v *DBStaticRoleDetailView) KeyHints() []ui.KeyHint {
	if v.rawMode {
		return []ui.KeyHint{
			{Key: "↑↓", Desc: "scroll"},
			{Key: "c", Desc: "copy"},
			{Key: "J/y", Desc: "json/yaml"},
			{Key: "r", Desc: "refresh"},
			{Key: "esc", Desc: "table view"},
		}
	}
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "J/y", Desc: "json/yaml"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *DBStaticRoleDetailView) buildRows() []components.Row {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	rows := []components.Row{
		{"Name", d.Name},
		{"DB Name", d.DBName},
		{"Username", d.Username},
		{"Rotation Period", dbValOrDash(d.RotationPeriod)},
		{"Last Vault Rotation", dbValOrDash(d.LastVaultRotation)},
	}
	if len(d.RotationStatements) > 0 {
		rows = append(rows, components.Row{"Rotation Statements", strings.Join(d.RotationStatements, "\n")})
	}
	return rows
}

func (v *DBStaticRoleDetailView) buildData() map[string]interface{} {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	return map[string]interface{}{
		"Name":               d.Name,
		"DBName":             d.DBName,
		"Username":           d.Username,
		"RotationPeriod":     dbValOrDash(d.RotationPeriod),
		"LastVaultRotation":  dbValOrDash(d.LastVaultRotation),
		"RotationStatements": strings.Join(d.RotationStatements, "\n"),
	}
}

func (v *DBStaticRoleDetailView) toggleRaw(format components.RawFormat) {
	if v.rawMode && v.rawView.Format() == format {
		v.rawMode = false
		return
	}
	data := v.buildData()
	if data == nil {
		return
	}
	if v.rawView == nil {
		v.rawView = components.NewRawView(data, format)
	} else {
		v.rawView.SetData(data)
		v.rawView.SetFormat(format)
	}
	v.rawView.Status = ""
	v.rawMode = true
}
