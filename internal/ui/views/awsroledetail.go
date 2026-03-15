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

type awsRoleDetailLoadedMsg struct {
	detail *vault.AWSRoleDetail
	err    error
}

// AWSRoleDetailView shows the full configuration of an AWS role.
type AWSRoleDetailView struct {
	client           *vault.Client
	mount            string
	name             string
	detail           *vault.AWSRoleDetail
	table            *components.Table
	rawView          *components.RawView
	err              error
	loading          bool
	rawMode          bool
	pendingRawFormat *components.RawFormat
}

var _ ui.View = (*AWSRoleDetailView)(nil)

var awsRoleDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewAWSRoleDetailView creates a detail view for an AWS role.
func NewAWSRoleDetailView(client *vault.Client, mount, name string) *AWSRoleDetailView {
	return &AWSRoleDetailView{
		client:  client,
		mount:   mount,
		name:    name,
		table:   components.NewTable(awsRoleDetailColumns),
		loading: true,
	}
}

func (v *AWSRoleDetailView) SetInitialRawFormat(format components.RawFormat) {
	v.pendingRawFormat = &format
}

func (v *AWSRoleDetailView) Init() tea.Cmd {
	return v.fetchData
}

func (v *AWSRoleDetailView) fetchData() tea.Msg {
	detail, err := v.client.ReadAWSRole(v.mount, v.name)
	return awsRoleDetailLoadedMsg{detail: detail, err: err}
}

func (v *AWSRoleDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case awsRoleDetailLoadedMsg:
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

const awsRoleDetailTitleHeight = 2

func (v *AWSRoleDetailView) View(width, height int) string {
	v.table.SetSize(width, height-awsRoleDetailTitleHeight)

	titleLine := styles.ViewTitleStyle.Width(width).Render("AWS Role: " + v.name)

	if v.loading {
		body := lipgloss.Place(width, height-awsRoleDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading role details..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-awsRoleDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.rawMode && v.rawView != nil {
		v.rawView.SetSize(width, height-awsRoleDetailTitleHeight)
		rawTitle := titleLine + "  " + styles.SecondaryStyle.Render("["+v.rawView.FormatLabel()+"]")
		return lipgloss.JoinVertical(lipgloss.Left, rawTitle, v.rawView.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *AWSRoleDetailView) Title() string {
	return "AWS Role: " + v.name
}

func (v *AWSRoleDetailView) KeyHints() []ui.KeyHint {
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

func (v *AWSRoleDetailView) toggleRaw(format components.RawFormat) {
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

func (v *AWSRoleDetailView) buildData() map[string]interface{} {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	data := map[string]interface{}{
		"Name":             d.Name,
		"Credential Types": awsValOrDash(strings.Join(d.CredentialTypes, ", ")),
		"Default STS TTL":  awsValOrDash(d.DefaultSTSTTL),
		"Max STS TTL":      awsValOrDash(d.MaxSTSTTL),
	}
	if len(d.RoleARNs) > 0 {
		data["Role ARNs"] = strings.Join(d.RoleARNs, ", ")
	}
	if len(d.PolicyARNs) > 0 {
		data["Policy ARNs"] = strings.Join(d.PolicyARNs, ", ")
	}
	if d.PolicyDocument != "" {
		data["Policy Document"] = d.PolicyDocument
	}
	if len(d.IAMGroups) > 0 {
		data["IAM Groups"] = strings.Join(d.IAMGroups, ", ")
	}
	if d.UserPath != "" {
		data["User Path"] = d.UserPath
	}
	return data
}

func (v *AWSRoleDetailView) buildRows() []components.Row {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	rows := []components.Row{
		{"Name", d.Name},
		{"Credential Types", awsValOrDash(strings.Join(d.CredentialTypes, ", "))},
	}
	if len(d.RoleARNs) > 0 {
		rows = append(rows, components.Row{"Role ARNs", strings.Join(d.RoleARNs, "\n")})
	}
	if len(d.PolicyARNs) > 0 {
		rows = append(rows, components.Row{"Policy ARNs", strings.Join(d.PolicyARNs, "\n")})
	}
	if d.PolicyDocument != "" {
		rows = append(rows, components.Row{"Policy Document", d.PolicyDocument})
	}
	if len(d.IAMGroups) > 0 {
		rows = append(rows, components.Row{"IAM Groups", strings.Join(d.IAMGroups, ", ")})
	}
	rows = append(rows,
		components.Row{"Default STS TTL", awsValOrDash(d.DefaultSTSTTL)},
		components.Row{"Max STS TTL", awsValOrDash(d.MaxSTSTTL)},
	)
	if d.UserPath != "" {
		rows = append(rows, components.Row{"User Path", d.UserPath})
	}
	return rows
}
