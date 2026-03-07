package views

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
	"github.com/miladbeigi/vaultui/internal/vault"
)

type diffLoadedMsg struct {
	oldData *vault.SecretData
	newData *vault.SecretData
	err     error
}

type diffLine struct {
	kind   string // "unchanged", "added", "removed", "changed"
	key    string
	oldVal string
	newVal string
}

// DiffView shows differences between two versions of a KV v2 secret.
type DiffView struct {
	client     *vault.Client
	mount      string
	path       string
	oldVersion int
	newVersion int
	lines      []diffLine
	err        error
	loading    bool
	scroll     int
}

var _ ui.View = (*DiffView)(nil)

func NewDiffView(client *vault.Client, mount, path string, oldVer, newVer int) *DiffView {
	return &DiffView{
		client:     client,
		mount:      mount,
		path:       path,
		oldVersion: oldVer,
		newVersion: newVer,
		loading:    true,
	}
}

func (v *DiffView) Init() tea.Cmd {
	return v.fetchDiff
}

func (v *DiffView) fetchDiff() tea.Msg {
	oldData, err := v.client.ReadSecretVersion(v.mount, v.path, v.oldVersion)
	if err != nil {
		return diffLoadedMsg{err: err}
	}
	newData, err := v.client.ReadSecretVersion(v.mount, v.path, v.newVersion)
	if err != nil {
		return diffLoadedMsg{err: err}
	}
	return diffLoadedMsg{oldData: oldData, newData: newData}
}

func (v *DiffView) Update(msg tea.Msg) (ui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case diffLoadedMsg:
		v.loading = false
		v.err = msg.err
		if msg.oldData != nil && msg.newData != nil {
			v.lines = computeDiff(msg.oldData, msg.newData)
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			v.scroll++
		case "k", "up":
			if v.scroll > 0 {
				v.scroll--
			}
		case "g", "home":
			v.scroll = 0
		case "G", "end":
			v.scroll = max(0, len(v.lines)-5)
		}
	}

	return v, nil
}

const diffTitleHeight = 2

func (v *DiffView) View(width, height int) string {
	title := v.renderTitle()

	if v.loading {
		body := lipgloss.Place(width, height-diffTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("Loading diff..."))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if v.err != nil {
		body := lipgloss.Place(width, height-diffTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.ErrorStyle.Render("Error: "+v.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	if len(v.lines) == 0 {
		body := lipgloss.Place(width, height-diffTitleHeight, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No differences"))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}

	bodyHeight := height - diffTitleHeight
	if v.scroll > len(v.lines)-bodyHeight {
		v.scroll = max(0, len(v.lines)-bodyHeight)
	}

	end := v.scroll + bodyHeight
	if end > len(v.lines) {
		end = len(v.lines)
	}
	visible := v.lines[v.scroll:end]

	addedStyle := lipgloss.NewStyle().Foreground(styles.AccentColor)
	removedStyle := lipgloss.NewStyle().Foreground(styles.ErrorColor)
	unchangedStyle := lipgloss.NewStyle().Foreground(styles.DimTextColor)
	keyStyle := lipgloss.NewStyle().Foreground(styles.TextColor).Width(24)

	rendered := make([]string, len(visible))
	for i, dl := range visible {
		switch dl.kind {
		case "added":
			rendered[i] = addedStyle.Render("+ ") + keyStyle.Render(dl.key) + addedStyle.Render(dl.newVal)
		case "removed":
			rendered[i] = removedStyle.Render("- ") + keyStyle.Render(dl.key) + removedStyle.Render(dl.oldVal)
		case "changed":
			rendered[i] = removedStyle.Render("- ") + keyStyle.Render(dl.key) + removedStyle.Render(dl.oldVal) + "\n" +
				addedStyle.Render("+ ") + keyStyle.Render(dl.key) + addedStyle.Render(dl.newVal)
		default:
			rendered[i] = unchangedStyle.Render("  ") + keyStyle.Render(dl.key) + unchangedStyle.Render(dl.oldVal)
		}
	}

	content := strings.Join(rendered, "\n")
	padded := lipgloss.NewStyle().Width(width).Height(bodyHeight).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, title, padded)
}

func (v *DiffView) renderTitle() string {
	label := styles.SubtleStyle.Render("Diff: ")
	path := styles.SecondaryStyle.Render(v.mount + v.path)
	versions := styles.SubtleStyle.Render(fmt.Sprintf("  v%d → v%d", v.oldVersion, v.newVersion))
	return lipgloss.NewStyle().PaddingBottom(1).Render(label + path + versions)
}

func (v *DiffView) Title() string {
	return fmt.Sprintf("Diff v%d→v%d %s%s", v.oldVersion, v.newVersion, v.mount, v.path)
}

func (v *DiffView) KeyHints() []ui.KeyHint {
	return []ui.KeyHint{
		{Key: "↑↓", Desc: "scroll"},
		{Key: "esc", Desc: "back"},
	}
}

func computeDiff(oldData, newData *vault.SecretData) []diffLine {
	allKeys := make(map[string]bool)
	for k := range oldData.Data {
		allKeys[k] = true
	}
	for k := range newData.Data {
		allKeys[k] = true
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var lines []diffLine
	for _, k := range sorted {
		oldVal, inOld := oldData.Data[k]
		newVal, inNew := newData.Data[k]

		switch {
		case inOld && !inNew:
			lines = append(lines, diffLine{kind: "removed", key: k, oldVal: oldVal})
		case !inOld && inNew:
			lines = append(lines, diffLine{kind: "added", key: k, newVal: newVal})
		case oldVal != newVal:
			lines = append(lines, diffLine{kind: "changed", key: k, oldVal: oldVal, newVal: newVal})
		default:
			lines = append(lines, diffLine{kind: "unchanged", key: k, oldVal: oldVal, newVal: newVal})
		}
	}
	return lines
}
