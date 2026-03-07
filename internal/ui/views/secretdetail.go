package views

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/clipboard"
	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type secretReadMsg struct {
	data *vault.SecretData
	err  error
}

type statusClearMsg struct{}

// SecretDetailView displays the key-value pairs of a single secret in a table.
type SecretDetailView struct {
	client    *vault.Client
	mount     string
	path      string
	kvV2      bool
	version   int
	table     *components.Table
	secret    *vault.SecretData
	err       error
	loading   bool
	statusMsg string
}

var _ ui.View = (*SecretDetailView)(nil)

var detailColumns = []components.Column{
	{Title: "KEY", MinWidth: 20},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewSecretDetailView creates a detail view for a specific secret.
func NewSecretDetailView(client *vault.Client, mount, path string, kvV2 bool) *SecretDetailView {
	return &SecretDetailView{
		client:  client,
		mount:   mount,
		path:    path,
		kvV2:    kvV2,
		table:   components.NewTable(detailColumns),
		loading: true,
	}
}

func (v *SecretDetailView) Init() tea.Cmd {
	return v.fetchSecret
}

func (v *SecretDetailView) fetchSecret() tea.Msg {
	if v.version > 0 && v.kvV2 {
		data, err := v.client.ReadSecretVersion(v.mount, v.path, v.version)
		return secretReadMsg{data: data, err: err}
	}
	data, err := v.client.ReadSecret(v.mount, v.path, v.kvV2)
	return secretReadMsg{data: data, err: err}
}

var copyKeys = struct {
	CopyVal  key.Binding
	CopyJSON key.Binding
}{
	CopyVal:  key.NewBinding(key.WithKeys("c")),
	CopyJSON: key.NewBinding(key.WithKeys("C")),
}

func (v *SecretDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case secretReadMsg:
		v.loading = false
		v.err = msg.err
		v.secret = msg.data
		v.table.SetRows(v.buildRows())
		return v, nil

	case statusClearMsg:
		v.statusMsg = ""
		return v, nil

	case tea.KeyMsg:
		if v.secret == nil {
			return v, nil
		}
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
		case key.Matches(msg, copyKeys.CopyVal):
			cmd := v.copySelectedValue()
			return v, cmd
		case key.Matches(msg, copyKeys.CopyJSON):
			cmd := v.copyJSON()
			return v, cmd
		case msg.String() == "v":
			if v.kvV2 {
				next := NewVersionsView(v.client, v.mount, v.path)
				return v, func() tea.Msg { return ui.PushViewMsg{View: next} }
			}
		}
	}

	return v, nil
}

func (v *SecretDetailView) copySelectedValue() tea.Cmd {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.secret.Keys) {
		return nil
	}
	k := v.secret.Keys[idx]
	val := v.secret.Data[k]

	if err := clipboard.Write(val); err != nil {
		v.statusMsg = "✗ " + err.Error()
	} else {
		v.statusMsg = fmt.Sprintf("✓ Copied '%s' to clipboard", k)
	}
	return v.clearStatusAfter()
}

func (v *SecretDetailView) copyJSON() tea.Cmd {
	jsonBytes, err := json.MarshalIndent(v.secret.Data, "", "  ")
	if err != nil {
		v.statusMsg = "✗ " + err.Error()
		return v.clearStatusAfter()
	}

	if err := clipboard.Write(string(jsonBytes)); err != nil {
		v.statusMsg = "✗ " + err.Error()
	} else {
		v.statusMsg = fmt.Sprintf("✓ Copied %d keys as JSON", len(v.secret.Keys))
	}
	return v.clearStatusAfter()
}

func (v *SecretDetailView) clearStatusAfter() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return statusClearMsg{}
	}
}

const detailBreadcrumbHeight = 2 // breadcrumb + blank line

func (v *SecretDetailView) View(width, height int) string {
	v.table.SetSize(width, height-detailBreadcrumbHeight)

	breadcrumb := v.renderBreadcrumb(width)

	if v.loading {
		body := lipgloss.Place(width, height-detailBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading secret..."))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-detailBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	if v.secret == nil || len(v.secret.Keys) == 0 {
		body := lipgloss.Place(width, height-detailBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Empty — no data in this secret"))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, breadcrumb, v.table.View())

	if v.statusMsg != "" {
		statusLine := styles.SuccessStyle.Render(v.statusMsg)
		bodyLines := strings.Split(body, "\n")
		totalHeight := height
		if len(bodyLines) >= totalHeight {
			bodyLines[totalHeight-1] = statusLine
		} else {
			for len(bodyLines) < totalHeight-1 {
				bodyLines = append(bodyLines, "")
			}
			bodyLines = append(bodyLines, statusLine)
		}
		return strings.Join(bodyLines, "\n")
	}

	return body
}

func (v *SecretDetailView) renderBreadcrumb(width int) string {
	suffix := ""
	if v.version > 0 {
		suffix = fmt.Sprintf("v%d", v.version)
	}
	return components.Breadcrumb(v.mount, v.path, suffix, width)
}

func (v *SecretDetailView) Title() string {
	title := v.mount + v.path
	if v.version > 0 {
		title += fmt.Sprintf(" (v%d)", v.version)
	}
	return title
}

func (v *SecretDetailView) KeyHints() []ui.KeyHint {
	hints := []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "c", Desc: "copy value"},
		{Key: "C", Desc: "copy JSON"},
	}
	if v.kvV2 {
		hints = append(hints, ui.KeyHint{Key: "v", Desc: "versions"})
	}
	hints = append(hints, ui.KeyHint{Key: "esc", Desc: "back"})
	return hints
}

func (v *SecretDetailView) buildRows() []components.Row {
	if v.secret == nil {
		return nil
	}
	rows := make([]components.Row, len(v.secret.Keys))
	for i, k := range v.secret.Keys {
		rows[i] = components.Row{k, v.secret.Data[k]}
	}
	return rows
}
