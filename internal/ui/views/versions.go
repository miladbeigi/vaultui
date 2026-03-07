package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type versionsLoadedMsg struct {
	versions []vault.VersionEntry
	err      error
}

// VersionsView displays the version history of a KV v2 secret.
type VersionsView struct {
	client   *vault.Client
	mount    string
	path     string
	table    *components.Table
	versions []vault.VersionEntry
	err      error
	loading  bool
}

var _ ui.View = (*VersionsView)(nil)

var versionColumns = []components.Column{
	{Title: "VERSION", MinWidth: 10},
	{Title: "CREATED", MinWidth: 24},
	{Title: "STATUS", MinWidth: 14, FlexFill: true},
}

func NewVersionsView(client *vault.Client, mount, path string) *VersionsView {
	return &VersionsView{
		client:  client,
		mount:   mount,
		path:    path,
		table:   components.NewTable(versionColumns),
		loading: true,
	}
}

func (v *VersionsView) Init() tea.Cmd {
	return v.fetchVersions
}

func (v *VersionsView) fetchVersions() tea.Msg {
	versions, err := v.client.ReadSecretMetadata(v.mount, v.path)
	return versionsLoadedMsg{versions: versions, err: err}
}

func (v *VersionsView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case versionsLoadedMsg:
		v.loading = false
		v.err = msg.err
		v.versions = msg.versions
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
		case msg.String() == "d":
			cmd := v.handleDiff()
			return v, cmd
		}
	}

	return v, nil
}

const versionsBreadcrumbHeight = 2

func (v *VersionsView) View(width, height int) string {
	v.table.SetSize(width, height-versionsBreadcrumbHeight)

	breadcrumb := v.renderBreadcrumb(width)

	if v.loading {
		body := lipgloss.Place(width, height-versionsBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading version history..."))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-versionsBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	if len(v.versions) == 0 {
		body := lipgloss.Place(width, height-versionsBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No versions found"))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, v.table.View())
}

func (v *VersionsView) renderBreadcrumb(width int) string {
	return components.Breadcrumb(v.mount, v.path, "versions", width)
}

func (v *VersionsView) Title() string {
	return v.mount + v.path + " (versions)"
}

func (v *VersionsView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "view version"},
		{Key: "d", Desc: "diff with prev"},
		{Key: "esc", Desc: "back"},
	}
}

func (v *VersionsView) handleEnter() tea.Cmd {
	ver := v.selectedVersion()
	if ver == nil {
		return nil
	}
	next := NewSecretDetailView(v.client, v.mount, v.path, true)
	next.version = ver.Version
	return func() tea.Msg {
		return ui.PushViewMsg{View: next}
	}
}

func (v *VersionsView) handleDiff() tea.Cmd {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.versions) || len(v.versions) < 2 {
		return nil
	}
	cur := v.versions[idx]
	// Find previous version
	var prev vault.VersionEntry
	if idx+1 < len(v.versions) {
		prev = v.versions[idx+1]
	} else {
		return nil
	}
	next := NewDiffView(v.client, v.mount, v.path, prev.Version, cur.Version)
	return func() tea.Msg {
		return ui.PushViewMsg{View: next}
	}
}

func (v *VersionsView) selectedVersion() *vault.VersionEntry {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.versions) {
		return nil
	}
	return &v.versions[idx]
}

func (v *VersionsView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.versions))
	for i, ver := range v.versions {
		status := styles.SuccessStyle.Render("current")
		if ver.Destroyed {
			status = styles.ErrorStyle.Render("destroyed")
		} else if ver.DeletionTime != "" {
			status = styles.ErrorStyle.Render("deleted")
		} else if i > 0 {
			status = styles.SubtleStyle.Render("old")
		}

		created := ver.CreatedTime.Format("2006-01-02 15:04:05")
		rows[i] = components.Row{
			fmt.Sprintf("v%d", ver.Version),
			created,
			status,
		}
	}
	return rows
}
