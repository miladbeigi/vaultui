package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/milad/vaultui/internal/ui"
	"github.com/milad/vaultui/internal/ui/components"
	"github.com/milad/vaultui/internal/ui/styles"
	"github.com/milad/vaultui/internal/vault"
)

type secretReadMsg struct {
	data *vault.SecretData
	err  error
}

// SecretDetailView displays the key-value pairs of a single secret in a table.
type SecretDetailView struct {
	client  *vault.Client
	mount   string
	path    string
	kvV2    bool
	table   *components.Table
	secret  *vault.SecretData
	err     error
	loading bool
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
	data, err := v.client.ReadSecret(v.mount, v.path, v.kvV2)
	return secretReadMsg{data: data, err: err}
}

func (v *SecretDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case secretReadMsg:
		v.loading = false
		v.err = msg.err
		v.secret = msg.data
		v.table.SetRows(v.buildRows())
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
		}
	}

	return v, nil
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

	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, v.table.View())
}

func (v *SecretDetailView) renderBreadcrumb(width int) string {
	sep := styles.BreadcrumbStyle.Render(" ▸ ")
	parts := []string{styles.SubtleStyle.Render(v.mount)}

	segments := strings.Split(strings.TrimSuffix(v.path, "/"), "/")
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		if i == len(segments)-1 {
			parts = append(parts, styles.BreadcrumbActiveStyle.Render(seg))
		} else {
			parts = append(parts, styles.SubtleStyle.Render(seg+"/"))
		}
	}

	crumb := strings.Join(parts, sep)
	return lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(crumb)
}

func (v *SecretDetailView) Title() string {
	return v.mount + v.path
}

func (v *SecretDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "esc", Desc: "back"},
	}
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
