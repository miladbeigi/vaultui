package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type tokenInspectMsg struct {
	details *vault.TokenDetails
	err     error
}

// TokenInspectorView displays comprehensive details about the current token.
type TokenInspectorView struct {
	client  *vault.Client
	details *vault.TokenDetails
	table   *components.Table
	err     error
	loading bool
}

var _ ui.View = (*TokenInspectorView)(nil)

var tokenInspectorColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 20},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

func NewTokenInspectorView(client *vault.Client) *TokenInspectorView {
	return &TokenInspectorView{
		client:  client,
		table:   components.NewTable(tokenInspectorColumns),
		loading: true,
	}
}

func (v *TokenInspectorView) Init() tea.Cmd {
	return v.fetchToken
}

func (v *TokenInspectorView) fetchToken() tea.Msg {
	details, err := v.client.InspectToken()
	return tokenInspectMsg{details: details, err: err}
}

func (v *TokenInspectorView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tokenInspectMsg:
		v.loading = false
		v.err = msg.err
		v.details = msg.details
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
			return v, v.fetchToken
		}
	}

	return v, nil
}

const tokenInspectorTitleHeight = 2

func (v *TokenInspectorView) View(width, height int) string {
	v.table.SetSize(width, height-tokenInspectorTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Token Inspector")

	if v.loading {
		body := lipgloss.Place(width, height-tokenInspectorTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Inspecting token..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-tokenInspectorTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *TokenInspectorView) Title() string {
	return "Token Inspector"
}

func (v *TokenInspectorView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *TokenInspectorView) buildRows() []components.Row {
	if v.details == nil {
		return nil
	}
	d := v.details

	rows := []components.Row{
		{"Type", tokenTypeDisplay(d.TokenType)},
		{"Display Name", valueOrDash(d.DisplayName)},
		{"Accessor", valueOrDash(d.Accessor)},
		{"Policies", valueOrDash(d.PoliciesString())},
		{"", ""},
		{"Renewable", fmt.Sprintf("%v", d.Renewable)},
		{"Orphan", fmt.Sprintf("%v", d.Orphan)},
		{"Num Uses", numUsesDisplay(d.NumUses)},
		{"", ""},
		{"TTL Remaining", formatDurationHuman(d.TTL)},
		{"Creation TTL", formatDurationHuman(d.CreationTTL)},
		{"Max TTL", formatDurationHuman(d.MaxTTL)},
		{"", ""},
	}

	if !d.CreationAt.IsZero() {
		rows = append(rows, components.Row{"Created", d.CreationAt.Format(time.RFC3339)})
	}
	if !d.ExpireAt.IsZero() {
		rows = append(rows, components.Row{"Expires", d.ExpireAt.Format(time.RFC3339)})
	} else {
		rows = append(rows, components.Row{"Expires", "never"})
	}

	rows = append(rows, components.Row{"", ""})

	if d.EntityID != "" {
		rows = append(rows, components.Row{"Entity ID", d.EntityID})
	}
	if d.Path != "" {
		rows = append(rows, components.Row{"Auth Path", d.Path})
	}

	for k, val := range d.Meta {
		rows = append(rows, components.Row{"meta:" + k, val})
	}

	return rows
}

func tokenTypeDisplay(t string) string {
	switch t {
	case "service":
		return "service"
	case "batch":
		return "batch"
	case "":
		return "service (default)"
	default:
		return t
	}
}

func numUsesDisplay(n int) string {
	if n == 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", n)
}

func valueOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func formatDurationHuman(d time.Duration) string {
	if d <= 0 {
		return "∞ (no expiry)"
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	switch {
	case hours > 24:
		days := hours / 24
		return fmt.Sprintf("%dd %dh", days, hours%24)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	case mins > 0:
		return fmt.Sprintf("%dm %ds", mins, secs)
	default:
		return fmt.Sprintf("%ds", secs)
	}
}
