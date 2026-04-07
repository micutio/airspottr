package tuiapp

import "github.com/charmbracelet/bubbles/table"

// Vertical space reserved above tables (header + stats blocks).
const layoutHeaderReservedRows = 8

// Row below tables (sort hotkeys only; notify lives in header).
const layoutFooterRows = 1

// Rarity tables on the global stats view share width in thirds.
const globalStatsTableCount = 3

// Gap between joined global-stats tables (lipgloss borders/spacing).
const layoutGlobalStatsInterTableGap = 2

func (m *model) resizeTables() {
	m.tables.setAllHeights(m.height - layoutHeaderReservedRows - layoutFooterRows)
	m.tables.resizeForTerminal(m.width, func(err error) {
		m.notify.Stdout.Panicf("%s", err)
	})
	m.applySortHeadersAfterResize()
}

// applySortHeadersAfterResize restores ▲/▼ titles; resize only touches widths.
func (m *model) applySortHeadersAfterResize() {
	applyAircraftSortHeaders(&m.tables.aircraft.table, m.aircraftSortCol, m.aircraftSortDesc)
	for i := range m.tables.rarities {
		applyRaritySortHeaders(
			&m.tables.rarities[i].table,
			i,
			m.raritySortCol[i],
			m.raritySortDesc[i],
		)
	}
}

func (m *model) updateAllTables() {
	records := filteredSortedAircraft(m.dashboard, m.aircraftSortCol, m.aircraftSortDesc)
	rows := buildAircraftRows(m.dashboard, records)
	m.tables.aircraft.table.SetRows(rows)
	applyAircraftSortHeaders(&m.tables.aircraft.table, m.aircraftSortCol, m.aircraftSortDesc)

	raritySources := [rarityTableCount]map[string]int{
		m.dashboard.SeenTypeCount,
		m.dashboard.SeenOperatorCount,
		m.dashboard.SeenCountryCount,
	}
	for idx := range m.tables.rarities {
		byProperty := m.raritySortCol[idx] == 1
		tuples := sortedPropertyCounts(raritySources[idx], byProperty, m.raritySortDesc[idx])
		rarityRows := make([]table.Row, len(tuples))
		for j := range tuples {
			rarityRows[j] = propertyCountToRow(tuples[j])
		}
		m.tables.rarities[idx].table.SetRows(rarityRows)
		applyRaritySortHeaders(
			&m.tables.rarities[idx].table,
			idx, m.raritySortCol[idx],
			m.raritySortDesc[idx])
	}
}

func (m *model) selectRarityNeighbour(direction int) {
	// Horizontal neighbour when pressing left / right between rarity tables.
	if m.uiState != globalStats || !m.activeTable().table.Focused() {
		return
	}

	rarityNavLeft := [rarityTableCount]int{rarityByCountry, rarityByType, rarityByOperator}
	rarityNavRight := [rarityTableCount]int{rarityByOperator, rarityByCountry, rarityByType}
	m.UnfocusSelectedTable()
	if direction < 0 {
		m.selectedRarityIdx = rarityNavLeft[m.selectedRarityIdx]
	} else {
		m.selectedRarityIdx = rarityNavRight[m.selectedRarityIdx]
	}
	m.FocusSelectedTable()
}

func (m *model) toggleGlobalView() {
	if m.inputFocus == focusNotifyStrip {
		return
	}
	switch m.uiState {
	case mainPage:
		m.UnfocusSelectedTable()
		m.uiState = globalStats
		m.selectedRarityIdx = rarityByType
		m.FocusSelectedTable()
	case globalStats:
		m.UnfocusSelectedTable()
		m.uiState = mainPage
		m.FocusSelectedTable()
	case aircraftDetails:
	default:
	}
}
