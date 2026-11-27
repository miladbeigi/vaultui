package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// FilterBarHeight is the vertical space a rendered FilterBar occupies
// (top border + input line + bottom border).
const FilterBarHeight = 3

// FilterBar is a `/`-triggered input that narrows table rows while typing.
// Text holds the applied filter; Active marks input mode, where every
// keystroke edits the filter instead of navigating the view.
type FilterBar struct {
	Text   string
	Active bool
}

// Update edits the filter from a key message. Enter commits the filter
// (it stays applied, input closes); Esc clears it; Ctrl+U resets the line.
func (f *FilterBar) Update(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEnter:
		f.Active = false
	case tea.KeyEsc:
		f.Text = ""
		f.Active = false
	case tea.KeyCtrlU:
		f.Text = ""
	case tea.KeyBackspace:
		if len(f.Text) > 0 {
			f.Text = f.Text[:len(f.Text)-1]
		}
	case tea.KeySpace:
		f.Text += " "
	case tea.KeyRunes:
		f.Text += string(msg.Runes)
	}
}

// View renders the filter input line, mirroring the command palette look.
func (f *FilterBar) View(width int) string {
	prompt := styles.SecondaryStyle.Render("/")
	line := prompt + " " + styles.HeaderValueStyle.Render(f.Text)
	if f.Active {
		line += styles.SecondaryStyle.Render("█")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		Width(width - 2).
		Render(line)
}

// Match reports whether pattern matches s as a case-insensitive
// subsequence: every character of pattern occurs in s in order, so
// "bnk" matches "bank". An empty pattern matches anything.
func Match(pattern, s string) bool {
	if pattern == "" {
		return true
	}

	p := strings.ToLower(pattern)
	str := strings.ToLower(s)

	i := 0
	for j := 0; i < len(p) && j < len(str); j++ {
		if p[i] == str[j] {
			i++
		}
	}
	return i == len(p)
}

// Filter returns the items whose key matches the pattern, preserving
// order. An empty pattern returns the items unchanged.
func Filter[T any](items []T, pattern string, key func(T) string) []T {
	if pattern == "" {
		return items
	}

	matched := make([]T, 0, len(items))
	for _, item := range items {
		if Match(pattern, key(item)) {
			matched = append(matched, item)
		}
	}
	return matched
}
