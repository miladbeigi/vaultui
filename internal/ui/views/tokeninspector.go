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
	client           *vault.Client
	details          *vault.TokenDetails
	table            *components.Table
	rawView          *components.RawView
	rawMode          bool
	pendingRawFormat *components.RawFormat
	err              error
	loading          bool
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

func (v *TokenInspectorView) SetInitialRawFormat(format components.RawFormat) {
	v.pendingRawFormat = &format
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
				return v, v.fetchToken
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
			return v, v.fetchToken
		case "J":
			v.toggleRaw(components.FormatJSON)
		case "y":
			v.toggleRaw(components.FormatYAML)
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

	if v.rawMode && v.rawView != nil {
		v.rawView.SetSize(width, height-tokenInspectorTitleHeight)
		rawTitle := title + "  " + styles.SecondaryStyle.Render("["+v.rawView.FormatLabel()+"]")
		return lipgloss.JoinVertical(lipgloss.Left, rawTitle, v.rawView.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *TokenInspectorView) Title() string {
	return "Token Inspector"
}

func (v *TokenInspectorView) KeyHints() []ui.KeyHint {
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

func (v *TokenInspectorView) toggleRaw(format components.RawFormat) {
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

func (v *TokenInspectorView) buildData() map[string]interface{} {
	if v.details == nil {
		return nil
	}
	d := v.details
	data := map[string]interface{}{
		"Type":          tokenTypeDisplay(d.TokenType),
		"Display Name":  valueOrDash(d.DisplayName),
		"Accessor":      valueOrDash(d.Accessor),
		"Policies":      valueOrDash(d.PoliciesString()),
		"Renewable":     d.Renewable,
		"Orphan":        d.Orphan,
		"Num Uses":      numUsesDisplay(d.NumUses),
		"TTL Remaining": formatDurationHuman(d.TTL),
		"Creation TTL":  formatDurationHuman(d.CreationTTL),
		"Max TTL":       formatDurationHuman(d.MaxTTL),
	}
	if !d.CreationAt.IsZero() {
		data["Created"] = d.CreationAt.Format(time.RFC3339)
	}
	if !d.ExpireAt.IsZero() {
		data["Expires"] = d.ExpireAt.Format(time.RFC3339)
	} else {
		data["Expires"] = "never"
	}
	if d.EntityID != "" {
		data["Entity ID"] = d.EntityID
	}
	if d.Path != "" {
		data["Auth Path"] = d.Path
	}
	if len(d.Meta) > 0 {
		data["Meta"] = d.Meta
	}
	return data
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
