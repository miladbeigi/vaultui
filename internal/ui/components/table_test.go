package components

import (
	"strings"
	"testing"
)

func testColumns() []Column {
	return []Column{
		{Title: "NAME", MinWidth: 10},
		{Title: "TYPE", MinWidth: 10, FlexFill: true},
	}
}

func testRows() []Row {
	return []Row{
		{"alpha", "a"},
		{"beta", "b"},
		{"gamma", "c"},
		{"delta", "d"},
		{"epsilon", "e"},
	}
}

func TestTable_NewTable(t *testing.T) {
	tbl := NewTable(testColumns())

	if tbl.RowCount() != 0 {
		t.Errorf("expected 0 rows, got %d", tbl.RowCount())
	}
	if tbl.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", tbl.Cursor())
	}
	if tbl.SelectedRow() != nil {
		t.Error("expected nil selected row on empty table")
	}
}

func TestTable_SetRows(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	if tbl.RowCount() != 5 {
		t.Errorf("expected 5 rows, got %d", tbl.RowCount())
	}
	if tbl.SelectedRow()[0] != "alpha" {
		t.Errorf("expected first row selected, got %q", tbl.SelectedRow()[0])
	}
}

func TestTable_MoveDown(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	tbl.MoveDown()
	if tbl.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", tbl.Cursor())
	}
	if tbl.SelectedRow()[0] != "beta" {
		t.Errorf("expected 'beta', got %q", tbl.SelectedRow()[0])
	}
}

func TestTable_MoveDown_AtBottom(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	for range 10 {
		tbl.MoveDown()
	}
	if tbl.Cursor() != 4 {
		t.Errorf("expected cursor clamped at 4, got %d", tbl.Cursor())
	}
}

func TestTable_MoveUp(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	tbl.MoveDown()
	tbl.MoveDown()
	tbl.MoveUp()
	if tbl.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", tbl.Cursor())
	}
}

func TestTable_MoveUp_AtTop(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	tbl.MoveUp()
	if tbl.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", tbl.Cursor())
	}
}

func TestTable_GoToTop(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	tbl.GoToBottom()
	tbl.GoToTop()
	if tbl.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", tbl.Cursor())
	}
}

func TestTable_GoToBottom(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())

	tbl.GoToBottom()
	if tbl.Cursor() != 4 {
		t.Errorf("expected cursor 4, got %d", tbl.Cursor())
	}
}

func TestTable_PageDown(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())
	tbl.SetSize(30, 6) // 3 visible rows (6 - 3 for header + separator + gap)

	tbl.PageDown()
	if tbl.Cursor() != 3 {
		t.Errorf("expected cursor 3 after page down, got %d", tbl.Cursor())
	}
}

func TestTable_PageUp(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())
	tbl.SetSize(30, 6)

	tbl.GoToBottom()
	tbl.PageUp()
	if tbl.Cursor() != 1 {
		t.Errorf("expected cursor 1 after page up, got %d", tbl.Cursor())
	}
}

func TestTable_SetRows_ResetsCursor(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())
	tbl.GoToBottom()

	tbl.SetRows([]Row{{"only", "one"}})
	if tbl.Cursor() != 0 {
		t.Errorf("expected cursor reset to 0, got %d", tbl.Cursor())
	}
}

func TestTable_View_RendersHeader(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())
	tbl.SetSize(30, 10)

	view := tbl.View()
	if !strings.Contains(view, "NAME") {
		t.Error("expected view to contain column header 'NAME'")
	}
	if !strings.Contains(view, "TYPE") {
		t.Error("expected view to contain column header 'TYPE'")
	}
}

func TestTable_View_RendersRows(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetRows(testRows())
	tbl.SetSize(30, 10)

	view := tbl.View()
	if !strings.Contains(view, "alpha") {
		t.Error("expected view to contain row data 'alpha'")
	}
	if !strings.Contains(view, "beta") {
		t.Error("expected view to contain row data 'beta'")
	}
}

func TestTable_View_Empty(t *testing.T) {
	tbl := NewTable(testColumns())
	tbl.SetSize(30, 10)

	view := tbl.View()
	if !strings.Contains(view, "NAME") {
		t.Error("expected view to still show headers when empty")
	}
}
