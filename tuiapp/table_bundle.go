package tuiapp

import "github.com/charmbracelet/bubbles/table"

// Rarity table column order on the global-stats row: Type | Operator | Country.
const (
	rarityByType = iota
	rarityByOperator
	rarityByCountry
	rarityTableCount = 3
)

// Horizontal neighbour when pressing left / right between rarity tables.
var (
	rarityNavLeft  = [rarityTableCount]int{rarityByCountry, rarityByType, rarityByOperator}
	rarityNavRight = [rarityTableCount]int{rarityByOperator, rarityByCountry, rarityByType}
)

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

	for i := 0; i < rarityByCountry; i++ {
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
