package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/components"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type pathListMsg struct {
	entries []vault.PathEntry
	err     error
}

// PathBrowserView browses directories and secrets within a secret engine.
type PathBrowserView struct {
	client  *vault.Client
	mount   string
	path    string
	kvV2    bool
	table   *components.Table
	entries []vault.PathEntry
	err     error
	loading bool
}

var _ ui.View = (*PathBrowserView)(nil)

var pathColumns = []components.Column{
	{Title: "NAME", MinWidth: 30, FlexFill: true},
	{Title: "TYPE", MinWidth: 12},
}

// NewPathBrowserView creates a path browser for the given mount and sub-path.
func NewPathBrowserView(client *vault.Client, mount, path string, kvV2 bool) *PathBrowserView {
	return &PathBrowserView{
		client:  client,
		mount:   mount,
		path:    path,
		kvV2:    kvV2,
		table:   components.NewTable(pathColumns),
		loading: true,
	}
}

func (v *PathBrowserView) Init() tea.Cmd {
	return v.fetchEntries
}

func (v *PathBrowserView) fetchEntries() tea.Msg {
	entries, err := v.client.ListSecrets(v.mount, v.path, v.kvV2)
	return pathListMsg{entries: entries, err: err}
}

func (v *PathBrowserView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case pathListMsg:
		v.loading = false
		v.err = msg.err
		v.entries = msg.entries
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

func (v *PathBrowserView) handleEnter() tea.Cmd {
	entry := v.selectedEntry()
	if entry == nil {
		return nil
	}

	if entry.IsDir {
		newPath := v.path + entry.Name
		next := NewPathBrowserView(v.client, v.mount, newPath, v.kvV2)
		return func() tea.Msg {
			return ui.PushViewMsg{View: next}
		}
	}

	secretPath := v.path + entry.Name
	next := NewSecretDetailView(v.client, v.mount, secretPath, v.kvV2)
	return func() tea.Msg {
		return ui.PushViewMsg{View: next}
	}
}

const pathBreadcrumbHeight = 2 // breadcrumb + blank line

func (v *PathBrowserView) View(width, height int) string {
	v.table.SetSize(width, height-pathBreadcrumbHeight)

	breadcrumb := v.renderBreadcrumb(width)

	if v.loading {
		body := lipgloss.Place(width, height-pathBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading..."))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-pathBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	if len(v.entries) == 0 {
		body := lipgloss.Place(width, height-pathBreadcrumbHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Empty — no keys at this path"))
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, body)
	}

	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, v.table.View())
}

func (v *PathBrowserView) renderBreadcrumb(width int) string {
	sep := styles.BreadcrumbStyle.Render(" ▸ ")

	parts := []string{styles.BreadcrumbActiveStyle.Render(v.mount)}

	if v.path != "" {
		segments := strings.Split(strings.TrimSuffix(v.path, "/"), "/")
		for i, seg := range segments {
			if i == len(segments)-1 {
				parts = append(parts, styles.BreadcrumbActiveStyle.Render(seg+"/"))
			} else {
				parts = append(parts, styles.SubtleStyle.Render(seg+"/"))
			}
		}
	}

	crumb := strings.Join(parts, sep)
	return lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(crumb)
}

func (v *PathBrowserView) Title() string {
	return v.mount + v.path
}

func (v *PathBrowserView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "open"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

func (v *PathBrowserView) selectedEntry() *vault.PathEntry {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.entries) {
		return nil
	}
	return &v.entries[idx]
}

func (v *PathBrowserView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.entries))
	for i, e := range v.entries {
		icon := "📄 "
		kind := "secret"
		if e.IsDir {
			icon = "📁 "
			kind = "dir"
		}
		rows[i] = components.Row{icon + e.Name, kind}
	}
	return rows
}
