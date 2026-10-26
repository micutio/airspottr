package tuiapp

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// Rarity table column order on the global-stats row: Type | Operator | Country.
const (
	rarityByType = iota
	rarityByOperator
	rarityByCountry
	rarityTableCount = 3
)

const (
	cellHorizontalPadding = 2 // default bubbles/table cell padding (1 left + 1 right)
	initialTableHeight    = 5
)

// colSpec defines a single column's base title and width behavior.
type colSpec struct {
	title string
	width int  // fixed width, or minimum width when flex is true
	flex  bool // absorbs remaining horizontal space
}

// sortableTable wraps a bubbles/table.Model with declarative column specs and sort state.
type sortableTable struct {
	table    table.Model
	specs    []colSpec
	sortCol  int
	sortDesc bool
	width    int
}

func newSortableTable(style table.Styles, specs []colSpec, defaultSortCol int) sortableTable {
	cols := make([]table.Column, len(specs))
	for i, s := range specs {
		cols[i] = table.Column{Title: s.title, Width: s.width}
	}
	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
		table.WithStyles(style),
	)
	tbl.Blur()

	st := sortableTable{
		table:    tbl,
		specs:    specs,
		sortCol:  defaultSortCol,
		sortDesc: false,
		width:    0,
	}
	st.syncColumns()
	return st
}

// resize updates the target inner width of the table and recalculates column widths.
func (st *sortableTable) resize(totalWidth int) {
	st.width = totalWidth
	st.syncColumns()
}

// syncColumns recalculates column widths for the current table width and applies sort arrows.
func (st *sortableTable) syncColumns() {
	colCount := len(st.specs)
	if colCount == 0 {
		return
	}

	fixedTotal := 0
	flexCount := 0
	for _, spec := range st.specs {
		if spec.flex {
			flexCount++
		} else {
			fixedTotal += spec.width
		}
	}

	availableForCols := st.width - colCount*cellHorizontalPadding
	flexWidth := 0
	if flexCount > 0 && availableForCols > fixedTotal {
		flexWidth = (availableForCols - fixedTotal) / flexCount
	}

	cols := make([]table.Column, colCount)
	for i, spec := range st.specs {
		w := spec.width
		if spec.flex {
			w = max(spec.width, flexWidth)
		}
		title := spec.title
		if i == st.sortCol {
			title += arrowForSort(st.sortDesc)
		}
		cols[i] = table.Column{Title: title, Width: w}
	}

	st.table.SetColumns(cols)
	if st.width > 0 {
		st.table.SetWidth(st.width)
	}
}

func (st *sortableTable) SetHeight(height int) {
	st.table.SetHeight(height)
}

func (st *sortableTable) SetRows(rows []table.Row) {
	st.table.SetRows(rows)
}

func (st *sortableTable) cycleSortCol(delta int) {
	n := len(st.specs)
	if n == 0 {
		return
	}
	st.sortCol = (st.sortCol + delta + n) % n
	st.syncColumns()
}

func (st *sortableTable) setSortCol(col int) bool {
	if col < 0 || col >= len(st.specs) {
		return false
	}
	st.sortCol = col
	st.syncColumns()
	return true
}

func (st *sortableTable) toggleSortDir() {
	st.sortDesc = !st.sortDesc
	st.syncColumns()
}

func arrowForSort(desc bool) string {
	if desc {
		return "▼"
	}
	return "▲"
}

func initTableStyle(theme Theme) table.Styles {
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.Padding(0, 1)
	tableStyle.Cell = tableStyle.Cell.Padding(0, 1)
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Highlight)
	return tableStyle
}

func newCurrentAircraftTable(tableStyle table.Styles) sortableTable {
	specs := []colSpec{
		{title: "DST", width: 4, flex: false},
		{title: "FNO", width: 8, flex: false},
		{title: "TID", width: 8, flex: true},
		{title: "DEP", width: 4, flex: false},
		{title: "ARR", width: 4, flex: false},
		{title: "ALT", width: 7, flex: false},
		{title: "SPD", width: 4, flex: false},
		{title: "HDG", width: 4, flex: false},
	}
	return newSortableTable(tableStyle, specs, fnoCol)
}

func newRarityTable(tableStyle table.Styles, valueColumnTitle string) sortableTable {
	specs := []colSpec{
		{title: "Count", width: 6, flex: false},
		{title: valueColumnTitle, width: 8, flex: true},
	}
	return newSortableTable(tableStyle, specs, 0)
}

// tuiTables groups the aircraft table and the three rarity tables.
type tuiTables struct {
	aircraft sortableTable
	rarities [rarityTableCount]sortableTable
}

func newTuiTables(tableStyle table.Styles) tuiTables {
	return tuiTables{
		aircraft: newCurrentAircraftTable(tableStyle),
		rarities: [rarityTableCount]sortableTable{
			newRarityTable(tableStyle, "Type"),
			newRarityTable(tableStyle, "Operator"),
			newRarityTable(tableStyle, "Country"),
		},
	}
}

func (t *tuiTables) setAllHeights(height int) {
	t.aircraft.SetHeight(height)
	for i := range t.rarities {
		t.rarities[i].SetHeight(height)
	}
}

// resizeForTerminal applies widths for main view (aircraft full width) and global stats (thirds).
func (t *tuiTables) resizeForTerminal(terminalWidth int) {
	const borderWidth = 2 // left + right rounded border characters
	t.aircraft.resize(max(0, terminalWidth-borderWidth))

	panelWidth := terminalWidth / rarityTableCount
	innerWidth := max(0, panelWidth-borderWidth)
	for i := range rarityByCountry {
		t.rarities[i].resize(innerWidth)
	}
	lastPanelWidth := terminalWidth - 2*panelWidth
	t.rarities[rarityByCountry].resize(max(0, lastPanelWidth-borderWidth))
}

func aircraftToRow(sighting obs.AircraftSighting, route ref.FlightrouteRecord) table.Row {
	return table.Row{
		fmt.Sprintf("%3.0f", sighting.Distance),
		sighting.LastRecord.GetCallsignAsStr(),
		sighting.TypeDesc,
		route.Origin.IataCode,
		route.Destination.IataCode,
		sighting.LastRecord.GetAltitudeAsStr(),
		fmt.Sprintf("%3.0f", sighting.LastRecord.GroundSpeed),
		fmt.Sprintf("%3.0f", sighting.LastRecord.NavHeading),
	}
}

func propertyCountToRow(propCount ref.PropertyCountTuple) table.Row {
	return table.Row{fmt.Sprintf("%5d", propCount.Count), propCount.Property}
}
