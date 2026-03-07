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

type pkiLoadedMsg struct {
	certs []vault.PKICert
	roles []vault.PKIRole
	err   error
}

// PKIView displays certificates and roles from a PKI engine.
type PKIView struct {
	client  *vault.Client
	mount   string
	table   *components.Table
	certs   []vault.PKICert
	roles   []vault.PKIRole
	err     error
	loading bool
	tab     int // 0 = certs, 1 = roles
}

var _ ui.View = (*PKIView)(nil)

var pkiCertColumns = []components.Column{
	{Title: "SERIAL NUMBER", MinWidth: 50, FlexFill: true},
}

var pkiRoleColumns = []components.Column{
	{Title: "ROLE NAME", MinWidth: 30, FlexFill: true},
}

func NewPKIView(client *vault.Client, mount string) *PKIView {
	return &PKIView{
		client:  client,
		mount:   mount,
		table:   components.NewTable(pkiCertColumns),
		loading: true,
	}
}

func (v *PKIView) Init() tea.Cmd {
	return v.fetchData
}

func (v *PKIView) fetchData() tea.Msg {
	certs, err := v.client.ListPKICerts(v.mount)
	if err != nil {
		return pkiLoadedMsg{err: err}
	}
	roles, _ := v.client.ListPKIRoles(v.mount)
	return pkiLoadedMsg{certs: certs, roles: roles}
}

func (v *PKIView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case pkiLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.certs = msg.certs
		v.roles = msg.roles
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
			v.tab = (v.tab + 1) % 2
			v.rebuildTable()
		case key.Matches(msg, navKeys.Enter):
			return v, v.handleEnter()
		}
	}

	return v, nil
}

func (v *PKIView) rebuildTable() {
	if v.tab == 0 {
		v.table = components.NewTable(pkiCertColumns)
		rows := make([]components.Row, len(v.certs))
		for i, c := range v.certs {
			rows[i] = components.Row{c.SerialNumber}
		}
		v.table.SetRows(rows)
	} else {
		v.table = components.NewTable(pkiRoleColumns)
		rows := make([]components.Row, len(v.roles))
		for i, r := range v.roles {
			rows[i] = components.Row{r.Name}
		}
		v.table.SetRows(rows)
	}
}

func (v *PKIView) handleEnter() tea.Cmd {
	if v.tab == 0 {
		idx := v.table.Cursor()
		if idx < 0 || idx >= len(v.certs) {
			return nil
		}
		next := NewPKICertDetailView(v.client, v.mount, v.certs[idx].SerialNumber)
		return func() tea.Msg { return ui.PushViewMsg{View: next} }
	}
	return nil
}

const pkiTitleHeight = 2

func (v *PKIView) View(width, height int) string {
	v.table.SetSize(width, height-pkiTitleHeight)

	tabNames := []string{"Certificates", "Roles"}
	tabs := ""
	for i, name := range tabNames {
		if i == v.tab {
			tabs += styles.SecondaryStyle.Render("["+name+"]") + "  "
		} else {
			tabs += styles.SubtleStyle.Render(" "+name+" ") + "  "
		}
	}
	title := lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(
		styles.ViewTitleStyle.Render("PKI: "+v.mount) + "  " + tabs)

	if v.loading {
		body := lipgloss.Place(width, height-pkiTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading PKI data..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-pkiTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *PKIView) Title() string {
	return "PKI: " + v.mount
}

func (v *PKIView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "tab", Desc: "switch tab"},
		{Key: "⏎", Desc: "view"},
		{Key: "esc", Desc: "back"},
	}
}
