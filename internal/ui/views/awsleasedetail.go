package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

// AWSLeaseDetailView shows the details of a single AWS lease.
type AWSLeaseDetailView struct {
	lease   vault.AWSLease
	table   *components.Table
	rawView *components.RawView
	rawMode bool
}

var _ ui.View = (*AWSLeaseDetailView)(nil)

var awsLeaseDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewAWSLeaseDetailView creates a detail view for an AWS lease.
func NewAWSLeaseDetailView(lease vault.AWSLease) *AWSLeaseDetailView {
	v := &AWSLeaseDetailView{
		lease: lease,
		table: components.NewTable(awsLeaseDetailColumns),
	}
	v.table.SetRows(v.buildRows())
	return v
}

func (v *AWSLeaseDetailView) SetInitialRawFormat(format components.RawFormat) {
	v.toggleRaw(format)
}

func (v *AWSLeaseDetailView) Init() tea.Cmd {
	return nil
}

func (v *AWSLeaseDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
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
		case "J":
			v.toggleRaw(components.FormatJSON)
		case "y":
			v.toggleRaw(components.FormatYAML)
		}
	}
	return v, nil
}

const awsLeaseDetailTitleHeight = 2

func (v *AWSLeaseDetailView) View(width, height int) string {
	v.table.SetSize(width, height-awsLeaseDetailTitleHeight)

	role, _ := splitLeaseID(v.lease.LeaseID)
	titleLine := styles.ViewTitleStyle.Width(width).Render("Lease: " + role)

	if v.rawMode && v.rawView != nil {
		v.rawView.SetSize(width, height-awsLeaseDetailTitleHeight)
		rawTitle := titleLine + "  " + styles.SecondaryStyle.Render("["+v.rawView.FormatLabel()+"]")
		return lipgloss.JoinVertical(lipgloss.Left, rawTitle, v.rawView.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *AWSLeaseDetailView) Title() string {
	role, _ := splitLeaseID(v.lease.LeaseID)
	return "Lease: " + role
}

func (v *AWSLeaseDetailView) KeyHints() []ui.KeyHint {
	if v.rawMode {
		return []ui.KeyHint{
			{Key: "↑↓", Desc: "scroll"},
			{Key: "c", Desc: "copy"},
			{Key: "J/y", Desc: "json/yaml"},
			{Key: "esc", Desc: "table view"},
		}
	}
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "J/y", Desc: "json/yaml"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *AWSLeaseDetailView) toggleRaw(format components.RawFormat) {
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

func (v *AWSLeaseDetailView) buildData() map[string]interface{} {
	l := v.lease
	role, shortID := splitLeaseID(l.LeaseID)
	return map[string]interface{}{
		"Lease ID":    l.LeaseID,
		"Short ID":    shortID,
		"Role":        role,
		"TTL":         formatLeaseTTL(l.TTL),
		"Issue Time":  awsValOrDash(l.IssueTime),
		"Expire Time": awsValOrDash(l.ExpireTime),
		"Renewable":   l.Renewable,
	}
}

func (v *AWSLeaseDetailView) buildRows() []components.Row {
	l := v.lease
	role, shortID := splitLeaseID(l.LeaseID)

	issueTime := l.IssueTime
	if len(issueTime) > 19 {
		issueTime = issueTime[:19]
	}
	expireTime := l.ExpireTime
	if len(expireTime) > 19 {
		expireTime = expireTime[:19]
	}

	return []components.Row{
		{"Lease ID", l.LeaseID},
		{"Short ID", shortID},
		{"Role", role},
		{"TTL", formatLeaseTTL(l.TTL)},
		{"Issue Time", awsValOrDash(issueTime)},
		{"Expire Time", awsValOrDash(expireTime)},
		{"Renewable", fmt.Sprintf("%v", l.Renewable)},
	}
}
