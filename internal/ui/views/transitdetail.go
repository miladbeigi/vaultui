package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type transitKeyLoadedMsg struct {
	detail *vault.TransitKeyDetail
	err    error
}

// TransitKeyDetailView displays details about a transit encryption key.
type TransitKeyDetailView struct {
	client  *vault.Client
	mount   string
	keyName string
	detail  *vault.TransitKeyDetail
	table   *components.Table
	err     error
	loading bool
}

var _ ui.View = (*TransitKeyDetailView)(nil)

var transitDetailColumns = []components.Column{
	{Title: "PROPERTY", MinWidth: 24},
	{Title: "VALUE", MinWidth: 30, FlexFill: true},
}

func NewTransitKeyDetailView(client *vault.Client, mount, keyName string) *TransitKeyDetailView {
	return &TransitKeyDetailView{
		client:  client,
		mount:   mount,
		keyName: keyName,
		table:   components.NewTable(transitDetailColumns),
		loading: true,
	}
}

func (v *TransitKeyDetailView) Init() tea.Cmd {
	return v.fetchDetail
}

func (v *TransitKeyDetailView) fetchDetail() tea.Msg {
	detail, err := v.client.ReadTransitKey(v.mount, v.keyName)
	return transitKeyLoadedMsg{detail: detail, err: err}
}

func (v *TransitKeyDetailView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case transitKeyLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.detail = msg.detail
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

const transitDetailTitleHeight = 2

func (v *TransitKeyDetailView) View(width, height int) string {
	v.table.SetSize(width, height-transitDetailTitleHeight)

	title := styles.ViewTitleStyle.Width(width).Render("Transit Key: " + v.keyName)

	if v.loading {
		body := lipgloss.Place(width, height-transitDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading key details..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-transitDetailTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *TransitKeyDetailView) Title() string {
	return "Transit Key: " + v.keyName
}

func (v *TransitKeyDetailView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *TransitKeyDetailView) buildRows() []components.Row {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	return []components.Row{
		{"Name", d.Name},
		{"Type", d.Type},
		{"Latest Version", fmt.Sprintf("%d", d.LatestVersion)},
		{"Min Decrypt Version", fmt.Sprintf("%d", d.MinDecryptVersion)},
		{"Min Encrypt Version", fmt.Sprintf("%d", d.MinEncryptVersion)},
		{"Exportable", fmt.Sprintf("%v", d.Exportable)},
		{"Deletion Allowed", fmt.Sprintf("%v", d.DeletionAllowed)},
	}
}
