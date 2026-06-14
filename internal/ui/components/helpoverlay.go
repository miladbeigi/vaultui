package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// HelpSection groups keybinding hints under a title.
type HelpSection struct {
	Title string
	Hints []ui.KeyHint
}

// HelpOverlay renders a centered modal listing keybinding help sections.
type HelpOverlay struct {
	Sections []HelpSection
}

var (
	helpSectionTitleStyle = lipgloss.NewStyle().
				Foreground(styles.SecondaryColor).
				Bold(true).
				Underline(true)
	helpSectionDividerStyle = lipgloss.NewStyle().Foreground(styles.SubtleColor)
)

// View renders the help overlay centered in the given area.
func (h HelpOverlay) View(width, height int) string {
	boxWidth := min(width-4, 72)
	if boxWidth < 30 {
		boxWidth = width - 2
	}

	innerWidth := boxWidth - 4
	twoCol := boxWidth >= 56

	var parts []string
	parts = append(parts, styles.ViewTitleStyle.Render("Keyboard Shortcuts"))
	parts = append(parts, renderSectionDivider(innerWidth))

	firstSection := true
	for _, section := range h.Sections {
		if len(section.Hints) == 0 {
			continue
		}
		if !firstSection {
			parts = append(parts, "", renderSectionDivider(innerWidth), "")
		}
		firstSection = false

		parts = append(parts, helpSectionTitleStyle.Render(section.Title))
		parts = append(parts, renderHintRows(section.Hints, innerWidth, twoCol))
	}

	parts = append(parts, "", renderSectionDivider(innerWidth), "")
	parts = append(parts, styles.SubtleStyle.Render("Press esc or ? to close"))

	content := strings.Join(parts, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func renderSectionDivider(width int) string {
	if width < 8 {
		width = 8
	}
	return helpSectionDividerStyle.Render(strings.Repeat("─", width))
}

func renderHintRows(hints []ui.KeyHint, innerWidth int, twoCol bool) string {
	if !twoCol {
		return renderHintColumn(hints, innerWidth)
	}

	mid := (len(hints) + 1) / 2
	left := renderHintColumn(hints[:mid], innerWidth/2-1)
	right := renderHintColumn(hints[mid:], innerWidth/2-1)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

func renderHintColumn(hints []ui.KeyHint, width int) string {
	lines := make([]string, len(hints))
	for i, h := range hints {
		keyPart := styles.HintKeyStyle.Render(h.Key)
		descPart := styles.HintDescStyle.Render(" " + h.Desc)
		line := keyPart + descPart
		if lipgloss.Width(line) > width && width > 0 {
			line = lipgloss.NewStyle().Width(width).Render(line)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
