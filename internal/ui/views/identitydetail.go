package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type identityDetailLoadedMsg struct {
	entity *vault.IdentityEntity
	group  *vault.IdentityGroup
	err    error
}

// IdentityDetailView displays entity or group details in a key-value table.
type IdentityDetailView struct {
	client   *vault.Client
	isEntity bool
	id       string
	name     string
	entity   *vault.IdentityEntity
	group    *vault.IdentityGroup
	table    *components.Table
	err      error
	loading  bool
}

var _ ui.View = (*IdentityDetailView)(nil)

var identityDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

// NewIdentityDetailView creates a detail view for an entity (isEntity=true) or group (isEntity=false).
func NewIdentityDetailView(client *vault.Client, isEntity bool, id, name string) *IdentityDetailView {
	return &IdentityDetailView{
		client:   client,
		isEntity: isEntity,
		id:       id,
		name:     name,
		table:    components.NewTable(identityDetailColumns),
		loading:  true,
	}
}

func (v *IdentityDetailView) Init() tea.Cmd {
	if v.isEntity {
		return v.fetchEntity
	}
	return v.fetchGroup
}

func (v *IdentityDetailView) fetchEntity() tea.Msg {
	entity, err := v.client.ReadIdentityEntity(v.id)
	return identityDetailLoadedMsg{entity: entity, err: err}
}

func (v *IdentityDetailView) fetchGroup() tea.Msg {
	group, err := v.client.ReadIdentityGroup(v.id)
	return identityDetailLoadedMsg{group: group, err: err}
}

func (v *IdentityDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case identityDetailLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.entity = msg.entity
		v.group = msg.group
		v.table.SetRows(v.buildRows())
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			v.table.MoveDown()
		case "k", "up":
			v.table.MoveUp()
		}
	}

	return v, nil
}

const identityDetailTitleHeight = 2

func (v *IdentityDetailView) View(width, height int) string {
	v.table.SetSize(width, height-identityDetailTitleHeight)

	title := "Entity: " + v.name
	if !v.isEntity {
		title = "Group: " + v.name
	}
	titleLine := styles.ViewTitleStyle.Width(width).Render(title)

	if v.loading {
		body := lipgloss.Place(width, height-identityDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading details..."))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-identityDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, titleLine, v.table.View())
}

func (v *IdentityDetailView) Title() string {
	if v.isEntity {
		return "Entity: " + v.name
	}
	return "Group: " + v.name
}

func (v *IdentityDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *IdentityDetailView) buildRows() []components.Row {
	if v.isEntity && v.entity != nil {
		d := v.entity
		rows := []components.Row{
			{"Name", d.Name},
			{"ID", d.ID},
			{"Policies", strings.Join(d.Policies, ", ")},
		}
		return rows
	}
	if !v.isEntity && v.group != nil {
		d := v.group
		rows := []components.Row{
			{"Name", d.Name},
			{"ID", d.ID},
			{"Policies", strings.Join(d.Policies, ", ")},
			{"Type", d.Type},
			{"Members", strings.Join(d.MemberEntityIDs, ", ")},
		}
		return rows
	}
	return nil
}
