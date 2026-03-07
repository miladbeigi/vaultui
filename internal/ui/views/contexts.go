package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/config"
	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// SwitchContextMsg is sent when the user selects a context to switch to.
type SwitchContextMsg struct {
	Context config.Context
}

// ContextsView displays available Vault contexts for switching.
type ContextsView struct {
	cfg      *config.Config
	table    *components.Table
	contexts []config.Context
}

var _ ui.View = (*ContextsView)(nil)

var ctxColumns = []components.Column{
	{Title: "NAME", MinWidth: 16},
	{Title: "ADDRESS", MinWidth: 30, FlexFill: true},
	{Title: "NAMESPACE", MinWidth: 14},
	{Title: "AUTH", MinWidth: 10},
}

func NewContextsView(cfg *config.Config) *ContextsView {
	v := &ContextsView{
		cfg:      cfg,
		table:    components.NewTable(ctxColumns),
		contexts: cfg.Contexts,
	}
	v.table.SetRows(v.buildRows())
	return v
}

func (v *ContextsView) Init() tea.Cmd { return nil }

func (v *ContextsView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, navKeys.Up):
			v.table.MoveUp()
		case key.Matches(msg, navKeys.Down):
			v.table.MoveDown()
		case key.Matches(msg, navKeys.Top):
			v.table.GoToTop()
		case key.Matches(msg, navKeys.Bottom):
			v.table.GoToBottom()
		case key.Matches(msg, navKeys.Enter):
			ctx := v.selectedContext()
			if ctx != nil {
				return v, func() tea.Msg { return SwitchContextMsg{Context: *ctx} }
			}
		}
	}
	return v, nil
}

const ctxTitleHeight = 2

func (v *ContextsView) View(width, height int) string {
	v.table.SetSize(width, height-ctxTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Contexts")

	if len(v.contexts) == 0 {
		body := lipgloss.Place(width, height-ctxTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No contexts configured.\nAdd contexts to ~/.vaultui.yaml"))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *ContextsView) Title() string { return "Contexts" }

func (v *ContextsView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "switch"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *ContextsView) selectedContext() *config.Context {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.contexts) {
		return nil
	}
	return &v.contexts[idx]
}

func (v *ContextsView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.contexts))
	for i, ctx := range v.contexts {
		name := ctx.Name
		if ctx.Name == v.cfg.CurrentContext {
			name = "● " + name
		}
		ns := ctx.Namespace
		if ns == "" {
			ns = "root"
		}
		auth := ctx.Auth.Method
		if auth == "" {
			auth = "token"
		}
		rows[i] = components.Row{name, ctx.Address, ns, auth}
	}
	return rows
}
