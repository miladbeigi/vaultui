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
	client           *vault.Client
	isEntity         bool
	id               string
	name             string
	entity           *vault.IdentityEntity
	group            *vault.IdentityGroup
	table            *components.Table
	rawView          *components.RawView
	rawMode          bool
	pendingRawFormat *components.RawFormat
	err              error
	loading          bool
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

func (v *IdentityDetailView) SetInitialRawFormat(format components.RawFormat) {
	v.pendingRawFormat = &format
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
		case "J":
			v.toggleRaw(components.FormatJSON)
		case "y":
			v.toggleRaw(components.FormatYAML)
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

	if v.rawMode && v.rawView != nil {
		v.rawView.SetSize(width, height-identityDetailTitleHeight)
		rawTitle := titleLine + "  " + styles.SecondaryStyle.Render("["+v.rawView.FormatLabel()+"]")
		return lipgloss.JoinVertical(lipgloss.Left, rawTitle, v.rawView.View())
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
	if v.rawMode {
		return []ui.KeyHint{
			{Key: "↑↓", Desc: "scroll"},
			{Key: "c", Desc: "copy"},
			{Key: "J/y", Desc: "json/yaml"},
			{Key: "esc", Desc: "table view"},
		}
	}
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "J/y", Desc: "json/yaml"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *IdentityDetailView) toggleRaw(format components.RawFormat) {
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

func (v *IdentityDetailView) buildData() map[string]interface{} {
	if v.isEntity && v.entity != nil {
		d := v.entity
		return map[string]interface{}{
			"Name":     d.Name,
			"ID":       d.ID,
			"Policies": strings.Join(d.Policies, ", "),
		}
	}
	if !v.isEntity && v.group != nil {
		d := v.group
		return map[string]interface{}{
			"Name":     d.Name,
			"ID":       d.ID,
			"Policies": strings.Join(d.Policies, ", "),
			"Type":     d.Type,
			"Members":  strings.Join(d.MemberEntityIDs, ", "),
		}
	}
	return nil
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
