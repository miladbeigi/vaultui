package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type policiesLoadedMsg struct {
	policies []string
	err      error
}

// PoliciesView displays the list of ACL policies.
type PoliciesView struct {
	client   *vault.Client
	table    *components.Table
	policies []string
	err      error
	loading  bool
}

var _ ui.View = (*PoliciesView)(nil)

var policyColumns = []components.Column{
	{Title: "NAME", MinWidth: 30, FlexFill: true},
	{Title: "TYPE", MinWidth: 10},
}

func NewPoliciesView(client *vault.Client) *PoliciesView {
	return &PoliciesView{
		client:  client,
		table:   components.NewTable(policyColumns),
		loading: true,
	}
}

func (v *PoliciesView) Init() tea.Cmd {
	return v.fetchPolicies
}

func (v *PoliciesView) fetchPolicies() tea.Msg {
	policies, err := v.client.ListPolicies()
	return policiesLoadedMsg{policies: policies, err: err}
}

func (v *PoliciesView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.policies = msg.policies
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
		case key.Matches(msg, navKeys.Enter):
			cmd := v.handleEnter()
			return v, cmd
		}
	}

	return v, nil
}

const policiesTitleHeight = 2

func (v *PoliciesView) View(width, height int) string {
	v.table.SetSize(width, height-policiesTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Policies")

	if v.loading {
		body := lipgloss.Place(width, height-policiesTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading policies..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-policiesTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if len(v.policies) == 0 {
		body := lipgloss.Place(width, height-policiesTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No policies found"))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *PoliciesView) Title() string {
	return "Policies"
}

func (v *PoliciesView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "view"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func (v *PoliciesView) handleEnter() tea.Cmd {
	name := v.selectedPolicy()
	if name == "" {
		return nil
	}
	next := NewPolicyDetailView(v.client, name)
	return func() tea.Msg {
		return ui.PushViewMsg{View: next}
	}
}

func (v *PoliciesView) selectedPolicy() string {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.policies) {
		return ""
	}
	return v.policies[idx]
}

func (v *PoliciesView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.policies))
	for i, p := range v.policies {
		pType := "acl"
		if p == "root" {
			pType = "root"
		}
		rows[i] = components.Row{p, pType}
	}
	return rows
}
