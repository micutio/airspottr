package tuiapp

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Rarity table column order on the global-stats row: Type | Operator | Country.
const (
	rarityByType = iota
	rarityByOperator
	rarityByCountry
	rarityTableCount = 3
)

type tableSetup struct {
	tables tuiTables
	style  table.Styles
}

// initTables creates and configures all tables used in the TUI.
func initTables(theme Theme) tableSetup {
	tableStyle := table.DefaultStyles()
	tableStyle.Header.Padding(0)
	tableStyle.Cell.Padding(0)
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Highlight)

	return tableSetup{
		tables: newTuiTables(tableStyle),
		style:  tableStyle,
	}
}

// tuiTables groups all bubble tables and shared layout operations.
type tuiTables struct {
	aircraft autoFormatTable
	rarities [rarityTableCount]autoFormatTable
}

func newTuiTables(tableStyle table.Styles) tuiTables {
	return tuiTables{
		aircraft: newCurrentAircraftTable(tableStyle),
		rarities: [rarityTableCount]autoFormatTable{
			newRarityTable(tableStyle, "Type"),
			newRarityTable(tableStyle, "operator"),
			newRarityTable(tableStyle, "country"),
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
func (t *tuiTables) resizeForTerminal(terminalWidth int, onResizeErr func(err error)) {
	leftSideWidth := terminalWidth - 1
	if err := t.aircraft.resize(leftSideWidth); err != nil {
		onResizeErr(err)
		return
	}

	rightSideWidth := terminalWidth
	third := 1.0 / float64(globalStatsTableCount)
	sideW := int(float64(rightSideWidth) * third)

	for i := range rarityByCountry {
		if err := t.rarities[i].resize(sideW); err != nil {
			onResizeErr(err)
			return
		}
	}
	countryW := rightSideWidth -
		2*sideW -
		layoutGlobalStatsInterTableGap -
		len(t.rarities[rarityByCountry].table.Columns())
	if err := t.rarities[rarityByCountry].resize(countryW); err != nil {
		onResizeErr(err)
	}
}
