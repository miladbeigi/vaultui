package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/milad/vaultui/internal/ui/styles"
)

// Column defines a table column with a header and fixed width.
type Column struct {
	Title string
	Width int
}

// Row is a slice of cell values corresponding to the columns.
type Row []string

// Table is a simple, vim-navigable table component.
type Table struct {
	columns []Column
	rows    []Row
	cursor  int
	offset  int
	height  int
	width   int
}

// NewTable creates a table with the given columns.
func NewTable(columns []Column) *Table {
	return &Table{
		columns: columns,
	}
}

// SetRows replaces the table data, resetting cursor if out of bounds.
func (t *Table) SetRows(rows []Row) {
	t.rows = rows
	if t.cursor >= len(rows) {
		t.cursor = max(0, len(rows)-1)
	}
}

// SetSize sets the available rendering dimensions.
func (t *Table) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// Cursor returns the current cursor index.
func (t *Table) Cursor() int {
	return t.cursor
}

// SelectedRow returns the currently highlighted row, or nil if empty.
func (t *Table) SelectedRow() Row {
	if len(t.rows) == 0 || t.cursor >= len(t.rows) {
		return nil
	}
	return t.rows[t.cursor]
}

// MoveUp moves the cursor up by one row.
func (t *Table) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
		if t.cursor < t.offset {
			t.offset = t.cursor
		}
	}
}

// MoveDown moves the cursor down by one row.
func (t *Table) MoveDown() {
	if t.cursor < len(t.rows)-1 {
		t.cursor++
		if t.cursor >= t.offset+t.visibleRows() {
			t.offset = t.cursor - t.visibleRows() + 1
		}
	}
}

// GoToTop moves the cursor to the first row.
func (t *Table) GoToTop() {
	t.cursor = 0
	t.offset = 0
}

// GoToBottom moves the cursor to the last row.
func (t *Table) GoToBottom() {
	t.cursor = max(0, len(t.rows)-1)
	vis := t.visibleRows()
	if len(t.rows) > vis {
		t.offset = len(t.rows) - vis
	}
}

// PageDown moves the cursor down by one page.
func (t *Table) PageDown() {
	vis := t.visibleRows()
	t.cursor = min(t.cursor+vis, len(t.rows)-1)
	if t.cursor >= t.offset+vis {
		t.offset = t.cursor - vis + 1
	}
}

// PageUp moves the cursor up by one page.
func (t *Table) PageUp() {
	vis := t.visibleRows()
	t.cursor = max(t.cursor-vis, 0)
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
}

// RowCount returns the total number of rows.
func (t *Table) RowCount() int {
	return len(t.rows)
}

func (t *Table) visibleRows() int {
	h := t.height - 2 // header row + separator
	if h < 1 {
		return 1
	}
	return h
}

// View renders the table.
func (t *Table) View() string {
	var b strings.Builder

	// Header
	var headerCells []string
	for _, col := range t.columns {
		cell := styles.TableHeaderStyle.Width(col.Width).Render(col.Title)
		headerCells = append(headerCells, cell)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	// Rows
	vis := t.visibleRows()
	end := min(t.offset+vis, len(t.rows))

	for i := t.offset; i < end; i++ {
		row := t.rows[i]
		var cells []string
		for j, col := range t.columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			style := lipgloss.NewStyle().Width(col.Width)
			if i == t.cursor {
				style = styles.SelectedRowStyle.Width(col.Width)
			}
			cells = append(cells, style.Render(truncate(val, col.Width)))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}
