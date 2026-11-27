package views

import (
	"charm.land/bubbles/v2/key"
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
	client   *vault.Client
	mount    string
	path     string
	kvV2     bool
	table    *components.Table
	entries  []vault.PathEntry
	filtered []vault.PathEntry
	filter   components.FilterBar
	err      error
	loading  bool
}

var _ ui.View = (*PathBrowserView)(nil)
var _ ui.Filterable = (*PathBrowserView)(nil)

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
		v.applyFilter()
		return v, nil

	case tea.KeyMsg:
		if v.filter.Active {
			v.filter.Update(msg)
			v.applyFilter()
			return v, nil
		}

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
		case msg.String() == "/":
			v.filter.Active = true
			return v, nil
		case msg.String() == "J":
			cmd := v.handleRawOpen(components.FormatJSON)
			return v, cmd
		case msg.String() == "y":
			cmd := v.handleRawOpen(components.FormatYAML)
			return v, cmd
		case msg.String() == "r":
			v.loading = true
			listPath := v.mount + v.path
			if v.kvV2 {
				listPath = v.mount + "metadata/" + v.path
			}
			v.client.InvalidateCache("list:" + listPath)
			return v, v.fetchEntries
		}
	}

	return v, nil
}

func (v *PathBrowserView) handleRawOpen(format components.RawFormat) tea.Cmd {
	entry := v.selectedEntry()
	if entry == nil || entry.IsDir {
		return nil
	}
	secretPath := v.path + entry.Name
	next := NewSecretDetailView(v.client, v.mount, secretPath, v.kvV2)
	next.SetInitialRawFormat(format)
	return func() tea.Msg { return ui.PushViewMsg{View: next} }
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
	filterHeight := 0
	parts := []string{v.renderBreadcrumb(width)}
	if v.filter.Active || v.filter.Text != "" {
		parts = append(parts, v.filter.View(width))
		filterHeight = components.FilterBarHeight
	}

	v.table.SetSize(width, height-pathBreadcrumbHeight-filterHeight)

	if v.loading {
		parts = append(parts, lipgloss.Place(width, height-pathBreadcrumbHeight-filterHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading...")))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	if v.err != nil {
		parts = append(parts, lipgloss.Place(width, height-pathBreadcrumbHeight-filterHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error())))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	if len(v.entries) == 0 {
		parts = append(parts, lipgloss.Place(width, height-pathBreadcrumbHeight-filterHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Empty — no keys at this path")))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	if len(v.filtered) == 0 {
		parts = append(parts, lipgloss.Place(width, height-pathBreadcrumbHeight-filterHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No matches for '"+v.filter.Text+"'")))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	return lipgloss.JoinVertical(lipgloss.Left, append(parts, v.table.View())...)
}

func (v *PathBrowserView) renderBreadcrumb(width int) string {
	return components.Breadcrumb(v.mount, v.path, "", width)
}

func (v *PathBrowserView) Title() string {
	return v.mount + v.path
}

func (v *PathBrowserView) KeyHints() []ui.KeyHint {
	if v.filter.Active {
		return []ui.KeyHint{
			{Key: "⏎", Desc: "apply filter"},
			{Key: "esc", Desc: "clear filter"},
		}
	}
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "⏎", Desc: "open"},
		{Key: "/", Desc: "filter"},
		{Key: "r", Desc: "refresh"},
		{Key: "J/y", Desc: "raw view"},
		{Key: "esc", Desc: "back"},
		{Key: "q", Desc: "quit"},
	}
}

// FilterActive reports whether the filter input is capturing keys.
func (v *PathBrowserView) FilterActive() bool {
	return v.filter.Active
}

// HasFilter reports whether a filter is currently applied.
func (v *PathBrowserView) HasFilter() bool {
	return v.filter.Text != ""
}

// ClearFilter drops the applied filter and shows all entries.
func (v *PathBrowserView) ClearFilter() {
	v.filter.Text = ""
	v.filter.Active = false
	v.applyFilter()
}

func (v *PathBrowserView) selectedEntry() *vault.PathEntry {
	idx := v.table.Cursor()
	if idx < 0 || idx >= len(v.filtered) {
		return nil
	}
	return &v.filtered[idx]
}

// applyFilter rebuilds the filtered entries and table rows, resetting
// the cursor to the top.
func (v *PathBrowserView) applyFilter() {
	v.filtered = components.Filter(v.entries, v.filter.Text, func(e vault.PathEntry) string {
		return e.Name
	})
	v.table.SetRows(v.buildRows())
	v.table.GoToTop()
}

func (v *PathBrowserView) buildRows() []components.Row {
	rows := make([]components.Row, len(v.filtered))
	for i, e := range v.filtered {
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
