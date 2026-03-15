package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewRawView_JSON(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"port": 8080,
	}
	rv := NewRawView(data, FormatJSON)
	content := rv.Content()

	if !strings.Contains(content, `"name"`) {
		t.Error("expected JSON to contain key 'name'")
	}
	if !strings.Contains(content, `"test"`) {
		t.Error("expected JSON to contain value 'test'")
	}
	if !strings.Contains(content, "8080") {
		t.Error("expected JSON to contain value 8080")
	}
}

func TestNewRawView_YAML(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"port": 8080,
	}
	rv := NewRawView(data, FormatYAML)
	content := rv.Content()

	if !strings.Contains(content, "name:") {
		t.Error("expected YAML to contain key 'name:'")
	}
	if !strings.Contains(content, "test") {
		t.Error("expected YAML to contain value 'test'")
	}
}

func TestRawView_SetFormat(t *testing.T) {
	data := map[string]interface{}{"key": "val"}
	rv := NewRawView(data, FormatJSON)

	if rv.Format() != FormatJSON {
		t.Error("expected format to be JSON")
	}

	rv.SetFormat(FormatYAML)
	if rv.Format() != FormatYAML {
		t.Error("expected format to be YAML after SetFormat")
	}
	if !strings.Contains(rv.Content(), "key:") {
		t.Error("expected YAML content after format change")
	}
}

func TestRawView_SetData(t *testing.T) {
	rv := NewRawView(map[string]interface{}{"a": 1}, FormatJSON)
	if !strings.Contains(rv.Content(), `"a"`) {
		t.Error("expected initial data in content")
	}

	rv.SetData(map[string]interface{}{"b": 2})
	if !strings.Contains(rv.Content(), `"b"`) {
		t.Error("expected new data in content")
	}
	if strings.Contains(rv.Content(), `"a"`) {
		t.Error("expected old data to be gone")
	}
}

func TestRawView_NilData(t *testing.T) {
	rv := NewRawView(nil, FormatJSON)
	if rv.Content() != "" {
		t.Error("expected empty content for nil data")
	}
}

func TestRawView_FormatLabel(t *testing.T) {
	rv := NewRawView(nil, FormatJSON)
	if rv.FormatLabel() != "JSON" {
		t.Errorf("expected 'JSON', got %q", rv.FormatLabel())
	}
	rv.SetFormat(FormatYAML)
	if rv.FormatLabel() != "YAML" {
		t.Errorf("expected 'YAML', got %q", rv.FormatLabel())
	}
}

func TestRawView_Scroll(t *testing.T) {
	data := map[string]interface{}{
		"a": "1", "b": "2", "c": "3", "d": "4", "e": "5",
		"f": "6", "g": "7", "h": "8", "i": "9", "j": "10",
	}
	rv := NewRawView(data, FormatJSON)
	rv.SetSize(80, 3)

	rv.ScrollDown()
	rv.ScrollDown()

	rv.GoToTop()
	rv.ScrollUp()

	rv.GoToBottom()

	rv.GoToTop()
	if rv.scroll != 0 {
		t.Errorf("expected scroll 0 after GoToTop, got %d", rv.scroll)
	}
}

func TestRawView_PageDownUp(t *testing.T) {
	data := map[string]interface{}{
		"a": "1", "b": "2", "c": "3", "d": "4", "e": "5",
		"f": "6", "g": "7", "h": "8", "i": "9", "j": "10",
	}
	rv := NewRawView(data, FormatJSON)
	rv.SetSize(80, 5)

	rv.PageDown()
	if rv.scroll == 0 {
		t.Error("expected scroll > 0 after PageDown")
	}

	rv.PageUp()
	if rv.scroll != 0 {
		t.Errorf("expected scroll 0 after PageUp, got %d", rv.scroll)
	}
}

func TestRawView_View_NoData(t *testing.T) {
	rv := NewRawView(nil, FormatJSON)
	rv.SetSize(40, 10)
	view := rv.View()
	if !strings.Contains(view, "No data") {
		t.Error("expected 'No data' placeholder")
	}
}

func TestRawView_View_WithData(t *testing.T) {
	rv := NewRawView(map[string]interface{}{"key": "val"}, FormatJSON)
	rv.SetSize(80, 20)
	view := rv.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestRawView_View_Status(t *testing.T) {
	rv := NewRawView(map[string]interface{}{"key": "val"}, FormatJSON)
	rv.SetSize(80, 20)
	rv.Status = "Copied!"
	view := rv.View()
	if !strings.Contains(view, "Copied!") {
		t.Error("expected status message in view")
	}
}

func TestColorizeJSON(t *testing.T) {
	s := lipgloss.NewStyle()
	lines := []string{
		`{`,
		`  "name": "test",`,
		`  "count": 42`,
		`}`,
	}
	for _, line := range lines {
		result := colorizeJSON(line, s, s, s)
		if result == "" {
			t.Errorf("expected non-empty colorized output for %q", line)
		}
	}
}

func TestColorizeYAML(t *testing.T) {
	s := lipgloss.NewStyle()
	lines := []string{
		`name: test`,
		`items:`,
		`  - first`,
		`  - second`,
		``,
	}
	for _, line := range lines {
		result := colorizeYAML(line, s, s, s)
		if line != "" && result == "" {
			t.Errorf("expected non-empty colorized output for %q", line)
		}
	}
}
