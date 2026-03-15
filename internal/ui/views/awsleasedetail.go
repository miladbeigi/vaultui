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
	lease vault.AWSLease
	table *components.Table
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

func (v *AWSLeaseDetailView) Init() tea.Cmd {
	return nil
}

func (v *AWSLeaseDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "j", "down":
			v.table.MoveDown()
		case "k", "up":
			v.table.MoveUp()
		case "g", "home":
			v.table.GoToTop()
		case "G", "end":
			v.table.GoToBottom()
		}
	}
	return v, nil
}

const awsLeaseDetailTitleHeight = 2

func (v *AWSLeaseDetailView) View(width, height int) string {
	v.table.SetSize(width, height-awsLeaseDetailTitleHeight)

	role, _ := splitLeaseID(v.lease.LeaseID)
	titleLine := styles.ViewTitleStyle.Width(width).Render("Lease: " + role)

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *AWSLeaseDetailView) Title() string {
	role, _ := splitLeaseID(v.lease.LeaseID)
	return "Lease: " + role
}

func (v *AWSLeaseDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "esc", Desc: "back"},
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
