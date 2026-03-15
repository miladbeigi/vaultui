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
	client           *vault.Client
	mount            string
	keyName          string
	detail           *vault.TransitKeyDetail
	table            *components.Table
	rawView          *components.RawView
	rawMode          bool
	pendingRawFormat *components.RawFormat
	err              error
	loading          bool
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

func (v *TransitKeyDetailView) SetInitialRawFormat(format components.RawFormat) {
	v.pendingRawFormat = &format
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

	if v.rawMode && v.rawView != nil {
		v.rawView.SetSize(width, height-transitDetailTitleHeight)
		rawTitle := title + "  " + styles.SecondaryStyle.Render("["+v.rawView.FormatLabel()+"]")
		return lipgloss.JoinVertical(lipgloss.Left, rawTitle, v.rawView.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, v.table.View())
}

func (v *TransitKeyDetailView) Title() string {
	return "Transit Key: " + v.keyName
}

func (v *TransitKeyDetailView) KeyHints() []ui.KeyHint {
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

func (v *TransitKeyDetailView) toggleRaw(format components.RawFormat) {
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

func (v *TransitKeyDetailView) buildData() map[string]interface{} {
	if v.detail == nil {
		return nil
	}
	d := v.detail
	return map[string]interface{}{
		"Name":                d.Name,
		"Type":                d.Type,
		"Latest Version":      d.LatestVersion,
		"Min Decrypt Version": d.MinDecryptVersion,
		"Min Encrypt Version": d.MinEncryptVersion,
		"Exportable":          d.Exportable,
		"Deletion Allowed":    d.DeletionAllowed,
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
