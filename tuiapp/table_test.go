package tuiapp

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
)

func TestSortableTableResizeAndSortHeaders(t *testing.T) {
	t.Parallel()

	specs := []colSpec{
		{title: "Count", width: 6, flex: false},
		{title: "Type", width: 8, flex: true},
	}
	st := newSortableTable(table.DefaultStyles(), specs, 0)

	// Default sort column 0 ascending -> "Count▲"
	cols := st.table.Columns()
	if cols[0].Title != "Count▲" || cols[1].Title != "Type" {
		t.Fatalf("unexpected initial titles: %+v", cols)
	}

	// Resize to 30 total width: 2 cols * 2 padding = 4 padding; 26 remaining; fixed=6 -> flex=20
	st.resize(30)
	cols = st.table.Columns()
	if st.table.Width() != 30 {
		t.Fatalf("table width = %d, want 30", st.table.Width())
	}
	if cols[0].Width != 6 || cols[1].Width != 20 {
		t.Fatalf("column widths = [%d, %d], want [6, 20]", cols[0].Width, cols[1].Width)
	}

	// Narrow terminal clamps flex column to its minimum width (8)
	st.resize(10)
	cols = st.table.Columns()
	if cols[1].Width != 8 {
		t.Fatalf("flex column width on narrow terminal = %d, want minimum 8", cols[1].Width)
	}

	// Cycle sort column and toggle direction
	st.cycleSortCol(1)
	st.toggleSortDir()
	cols = st.table.Columns()
	if cols[0].Title != "Count" || cols[1].Title != "Type▼" {
		t.Fatalf("unexpected titles after cycle + toggle: %+v", cols)
	}
}
