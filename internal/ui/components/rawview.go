package components

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"

	"github.com/miladbeigi/vaultui/internal/clipboard"
	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// RawFormat selects JSON or YAML rendering.
type RawFormat int

const (
	FormatJSON RawFormat = iota
	FormatYAML
)

// RawView renders structured data as formatted JSON or YAML with scrolling and copy.
type RawView struct {
	data    map[string]interface{}
	format  RawFormat
	content string
	scroll  int
	width   int
	height  int
	Status  string
}

// NewRawView creates a new RawView for the given data and format.
func NewRawView(data map[string]interface{}, format RawFormat) *RawView {
	rv := &RawView{
		data:   data,
		format: format,
	}
	rv.render()
	return rv
}

// SetData replaces the data and re-renders.
func (r *RawView) SetData(data map[string]interface{}) {
	r.data = data
	r.scroll = 0
	r.render()
}

// SetFormat changes the format and re-renders.
func (r *RawView) SetFormat(f RawFormat) {
	r.format = f
	r.scroll = 0
	r.render()
}

// Format returns the current format.
func (r *RawView) Format() RawFormat {
	return r.format
}

// Content returns the raw rendered text.
func (r *RawView) Content() string {
	return r.content
}

// SetSize sets the viewport dimensions.
func (r *RawView) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// ScrollDown moves the viewport down.
func (r *RawView) ScrollDown() {
	maxScroll := r.maxScroll()
	if r.scroll < maxScroll {
		r.scroll++
	}
}

// ScrollUp moves the viewport up.
func (r *RawView) ScrollUp() {
	if r.scroll > 0 {
		r.scroll--
	}
}

// GoToTop scrolls to the beginning.
func (r *RawView) GoToTop() {
	r.scroll = 0
}

// GoToBottom scrolls to the end.
func (r *RawView) GoToBottom() {
	r.scroll = r.maxScroll()
}

// PageDown moves the viewport down by half a page.
func (r *RawView) PageDown() {
	r.scroll += r.height / 2
	if r.scroll > r.maxScroll() {
		r.scroll = r.maxScroll()
	}
}

// PageUp moves the viewport up by half a page.
func (r *RawView) PageUp() {
	r.scroll -= r.height / 2
	if r.scroll < 0 {
		r.scroll = 0
	}
}

// CopyContent copies the raw content to clipboard with auto-clear.
func (r *RawView) CopyContent() error {
	return clipboard.WriteWithAutoClear(r.content, 30*time.Second)
}

// FormatLabel returns "JSON" or "YAML".
func (r *RawView) FormatLabel() string {
	if r.format == FormatYAML {
		return "YAML"
	}
	return "JSON"
}

// View renders the visible portion of the formatted content.
func (r *RawView) View() string {
	if r.content == "" {
		return lipgloss.Place(r.width, r.height, lipgloss.Center, lipgloss.Center,
			styles.SubtleStyle.Render("No data"))
	}

	lines := strings.Split(r.content, "\n")

	if r.scroll > r.maxScroll() {
		r.scroll = r.maxScroll()
	}

	end := r.scroll + r.height
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[r.scroll:end]

	keyStyle := lipgloss.NewStyle().Foreground(styles.SecondaryColor)
	valStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
	punctStyle := lipgloss.NewStyle().Foreground(styles.DimTextColor)

	rendered := make([]string, len(visible))
	for i, line := range visible {
		rendered[i] = colorizeLine(line, r.format, keyStyle, valStyle, punctStyle)
	}

	content := strings.Join(rendered, "\n")
	padded := lipgloss.NewStyle().Width(r.width).Height(r.height).Render(content)

	if r.Status != "" {
		padLines := strings.Split(padded, "\n")
		if len(padLines) > 0 {
			padLines[len(padLines)-1] = styles.SuccessStyle.Render(r.Status)
			padded = strings.Join(padLines, "\n")
		}
	}

	return padded
}

func (r *RawView) render() {
	if r.data == nil {
		r.content = ""
		return
	}

	var out []byte
	var err error

	switch r.format {
	case FormatYAML:
		out, err = yaml.Marshal(r.data)
	default:
		out, err = json.MarshalIndent(r.data, "", "  ")
	}

	if err != nil {
		r.content = fmt.Sprintf("render error: %v", err)
		return
	}
	r.content = strings.TrimRight(string(out), "\n")
}

func (r *RawView) maxScroll() int {
	lines := strings.Split(r.content, "\n")
	m := len(lines) - r.height
	if m < 0 {
		return 0
	}
	return m
}

func colorizeLine(line string, format RawFormat, keyStyle, valStyle, punctStyle lipgloss.Style) string {
	if format == FormatJSON {
		return colorizeJSON(line, keyStyle, valStyle, punctStyle)
	}
	return colorizeYAML(line, keyStyle, valStyle, punctStyle)
}

func colorizeJSON(line string, keyStyle, valStyle, punctStyle lipgloss.Style) string {
	trimmed := strings.TrimSpace(line)

	if trimmed == "{" || trimmed == "}" || trimmed == "}," ||
		trimmed == "[" || trimmed == "]" || trimmed == "]," {
		return punctStyle.Render(line)
	}

	indent := line[:len(line)-len(strings.TrimLeft(line, " "))]

	colonIdx := strings.Index(trimmed, ":")
	if colonIdx > 0 && strings.HasPrefix(trimmed, "\"") {
		key := trimmed[:colonIdx]
		rest := trimmed[colonIdx:]
		return indent + keyStyle.Render(key) + punctStyle.Render(":") + valStyle.Render(rest[1:])
	}

	return valStyle.Render(line)
}

func colorizeYAML(line string, keyStyle, valStyle, punctStyle lipgloss.Style) string {
	if strings.TrimSpace(line) == "" {
		return line
	}

	indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "- ") {
		return indent + punctStyle.Render("- ") + valStyle.Render(trimmed[2:])
	}

	colonIdx := strings.Index(trimmed, ":")
	if colonIdx > 0 {
		key := trimmed[:colonIdx]
		rest := trimmed[colonIdx:]
		if len(rest) > 1 {
			return indent + keyStyle.Render(key) + punctStyle.Render(":") + valStyle.Render(rest[1:])
		}
		return indent + keyStyle.Render(key) + punctStyle.Render(":")
	}

	return valStyle.Render(line)
}
