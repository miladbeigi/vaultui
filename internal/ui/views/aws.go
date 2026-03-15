package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type awsLoadedMsg struct {
	roles  []vault.AWSRole
	config *vault.AWSConfig
	lease  *vault.AWSLeaseConfig
	leases []vault.AWSLease
	err    error
}

// AWSView displays roles, config, and leases for an AWS secrets engine.
type AWSView struct {
	client  *vault.Client
	mount   string
	table   *components.Table
	roles   []vault.AWSRole
	config  *vault.AWSConfig
	lease   *vault.AWSLeaseConfig
	leases  []vault.AWSLease
	err     error
	loading bool
	tab     int // 0 = roles, 1 = config, 2 = leases
}

var _ ui.View = (*AWSView)(nil)

var awsRoleColumns = []components.Column{
	{Title: "NAME", MinWidth: 24},
	{Title: "CREDENTIAL TYPE", MinWidth: 20},
	{Title: "POLICY ARNS", MinWidth: 30, FlexFill: true},
}

var awsConfigColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

var awsLeaseColumns = []components.Column{
	{Title: "ROLE", MinWidth: 20},
	{Title: "LEASE ID", MinWidth: 20, FlexFill: true},
	{Title: "TTL", MinWidth: 10},
	{Title: "ISSUE TIME", MinWidth: 20},
}

// NewAWSView creates a new AWS engine browser.
func NewAWSView(client *vault.Client, mount string) *AWSView {
	return &AWSView{
		client:  client,
		mount:   mount,
		table:   components.NewTable(awsRoleColumns),
		loading: true,
	}
}

func (v *AWSView) Init() tea.Cmd {
	return v.fetchData
}

func (v *AWSView) fetchData() tea.Msg {
	roles, err := v.client.ListAWSRoles(v.mount)
	if err != nil {
		return awsLoadedMsg{err: err}
	}
	config, _ := v.client.ReadAWSConfig(v.mount)
	lease, _ := v.client.ReadAWSLeaseConfig(v.mount)
	leases, _ := v.client.ListAWSLeases(v.mount)
	return awsLoadedMsg{roles: roles, config: config, lease: lease, leases: leases}
}

func (v *AWSView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case awsLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.roles = msg.roles
		v.config = msg.config
		v.lease = msg.lease
		v.leases = msg.leases
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
		case msg.String() == "r":
			v.loading = true
			return v, v.fetchData
		}
	}

	return v, nil
}

func (v *AWSView) rebuildTable() {
	switch v.tab {
	case 0:
		v.table = components.NewTable(awsRoleColumns)
		v.table.SetRows(v.buildRoleRows())
	case 1:
		v.table = components.NewTable(awsConfigColumns)
		v.table.SetRows(v.buildConfigRows())
	case 2:
		v.table = components.NewTable(awsLeaseColumns)
		v.table.SetRows(v.buildLeaseRows())
	}
}

func (v *AWSView) handleEnter() tea.Cmd {
	switch v.tab {
	case 0:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.roles) {
			return nil
		}
		next := NewAWSRoleDetailView(v.client, v.mount, v.roles[idx].Name)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	case 2:
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.leases) {
			return nil
		}
		next := NewAWSLeaseDetailView(v.leases[idx])
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	}
	return nil
}

func (v *AWSView) buildRoleRows() []components.Row {
	rows := make([]components.Row, len(v.roles))
	for i, r := range v.roles {
		rows[i] = components.Row{r.Name, r.CredentialType, strings.Join(r.PolicyARNs, ", ")}
	}
	return rows
}

func (v *AWSView) buildConfigRows() []components.Row {
	if v.config == nil {
		return nil
	}
	c := v.config
	rows := []components.Row{
		{"Access Key", awsValOrDash(c.AccessKey)},
		{"Region", awsValOrDash(c.Region)},
		{"IAM Endpoint", awsValOrDash(c.IAMEndpoint)},
		{"STS Endpoint", awsValOrDash(c.STSEndpoint)},
		{"Max Retries", fmt.Sprintf("%d", c.MaxRetries)},
	}
	if v.lease != nil {
		rows = append(rows,
			components.Row{"Lease", awsValOrDash(v.lease.Lease)},
			components.Row{"Lease Max", awsValOrDash(v.lease.LeaseMax)},
		)
	}
	return rows
}

func (v *AWSView) buildLeaseRows() []components.Row {
	rows := make([]components.Row, len(v.leases))
	for i, l := range v.leases {
		ttl := formatLeaseTTL(l.TTL)
		issueTime := l.IssueTime
		if len(issueTime) > 19 {
			issueTime = issueTime[:19]
		}
		role, shortID := splitLeaseID(l.LeaseID)
		rows[i] = components.Row{role, shortID, ttl, issueTime}
	}
	return rows
}

func formatLeaseTTL(d time.Duration) string {
	if d <= 0 {
		return "expired"
	}
	return formatDurationHuman(d)
}

func splitLeaseID(id string) (role, short string) {
	parts := strings.Split(id, "/")
	if len(parts) >= 3 {
		role = parts[len(parts)-2]
		short = parts[len(parts)-1]
		return role, short
	}
	return "-", id
}

const awsTitleHeight = 2

func (v *AWSView) View(width, height int) string {
	v.table.SetSize(width, height-awsTitleHeight)

	tabNames := []string{"Roles", "Config", "Leases"}
	tabs := ""
	for i, name := range tabNames {
		if i == v.tab {
			tabs += styles.SecondaryStyle.Render("["+name+"]") + "  "
		} else {
			tabs += styles.SubtleStyle.Render(" "+name+" ") + "  "
		}
	}
	title := lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(
		styles.ViewTitleStyle.Render("AWS: "+v.mount) + "  " + tabs)

	if v.loading {
		body := lipgloss.Place(width, height-awsTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading AWS data..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-awsTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	empty := (v.tab == 0 && len(v.roles) == 0) ||
		(v.tab == 1 && v.config == nil) ||
		(v.tab == 2 && len(v.leases) == 0)
	if empty {
		msgs := []string{"No roles found", "No config found", "No active leases"}
		body := lipgloss.Place(width, height-awsTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render(msgs[v.tab]))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *AWSView) Title() string {
	return "AWS: " + v.mount
}

func (v *AWSView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "tab", Desc: "switch tab"},
		{Key: "⏎", Desc: "view"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
}

func awsValOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
