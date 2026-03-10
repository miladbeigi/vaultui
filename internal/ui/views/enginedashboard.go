package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type engineConfigMsg struct {
	config *vault.EngineConfig
	err    error
}

// EngineDashboardView displays detailed mount configuration for a secret engine.
type EngineDashboardView struct {
	client  *vault.Client
	path    string
	config  *vault.EngineConfig
	table   *components.Table
	err     error
	loading bool
}

var _ ui.View = (*EngineDashboardView)(nil)

var engineDashColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 28},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

func NewEngineDashboardView(client *vault.Client, path string) *EngineDashboardView {
	return &EngineDashboardView{
		client:  client,
		path:    path,
		table:   components.NewTable(engineDashColumns),
		loading: true,
	}
}

func (v *EngineDashboardView) Init() tea.Cmd {
	return v.fetchConfig
}

func (v *EngineDashboardView) fetchConfig() tea.Msg {
	config, err := v.client.ReadEngineConfig(v.path)
	return engineConfigMsg{config: config, err: err}
}

func (v *EngineDashboardView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case engineConfigMsg:
		v.loading = false
		v.err = msg.err
		v.config = msg.config
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
			v.client.InvalidateCache("sys/mounts/" + v.path)
			return v, v.fetchConfig
		}
	}

	return v, nil
}

const engineDashTitleHeight = 2

func (v *EngineDashboardView) View(width, height int) string {
	v.table.SetSize(width, height-engineDashTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Engine: " + v.path)

	if v.loading {
		body := lipgloss.Place(width, height-engineDashTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading engine config..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-engineDashTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *EngineDashboardView) Title() string {
	return "Engine: " + v.path
}

func (v *EngineDashboardView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *EngineDashboardView) buildRows() []components.Row {
	if v.config == nil {
		return nil
	}
	c := v.config

	rows := []components.Row{
		{"Path", c.Path},
		{"Type", c.Type},
		{"Description", engineValOrDash(c.Description)},
		{"UUID", c.UUID},
		{"Accessor", c.Accessor},
		{"", ""},
		{"Default Lease TTL", formatEngineTTL(c.DefaultLeaseTTL)},
		{"Max Lease TTL", formatEngineTTL(c.MaxLeaseTTL)},
		{"Force No Cache", fmt.Sprintf("%v", c.ForceNoCache)},
		{"", ""},
		{"Local", fmt.Sprintf("%v", c.Local)},
		{"Seal Wrap", fmt.Sprintf("%v", c.SealWrap)},
		{"External Entropy", fmt.Sprintf("%v", c.ExternalEntropyAccess)},
	}

	if c.RunningVersion != "" {
		rows = append(rows, components.Row{"Running Version", c.RunningVersion})
	}
	if c.PluginVersion != "" {
		rows = append(rows, components.Row{"Plugin Version", c.PluginVersion})
	}
	if c.ListingVisibility != "" {
		rows = append(rows, components.Row{"Listing Visibility", c.ListingVisibility})
	}
	if c.TokenType != "" {
		rows = append(rows, components.Row{"Token Type", c.TokenType})
	}

	if len(c.Options) > 0 {
		rows = append(rows, components.Row{"", ""})
		for k, val := range c.Options {
			rows = append(rows, components.Row{"option:" + k, val})
		}
	}

	if len(c.AuditNonHMACRequestKeys) > 0 {
		rows = append(rows, components.Row{"Audit Non-HMAC Req Keys", strings.Join(c.AuditNonHMACRequestKeys, ", ")})
	}
	if len(c.AuditNonHMACResponseKeys) > 0 {
		rows = append(rows, components.Row{"Audit Non-HMAC Resp Keys", strings.Join(c.AuditNonHMACResponseKeys, ", ")})
	}
	if len(c.PassthroughRequestHeaders) > 0 {
		rows = append(rows, components.Row{"Passthrough Headers", strings.Join(c.PassthroughRequestHeaders, ", ")})
	}

	return rows
}

func engineValOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func formatEngineTTL(d time.Duration) string {
	if d <= 0 {
		return "system default"
	}
	return formatDurationHuman(d)
}
