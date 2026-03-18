package tuiapp

import "github.com/charmbracelet/bubbles/table"

// Vertical space reserved above tables (header + stats blocks).
const layoutHeaderReservedRows = 8

// Row for sort hotkey hint below tables.
const layoutSortHintRows = 1

// Rarity tables on the global stats view share width in thirds.
const globalStatsTableCount = 3

// Gap between joined global-stats tables (lipgloss borders/spacing).
const layoutGlobalStatsInterTableGap = 2

func (m *model) resizeTables() {
	m.tables.setAllHeights(m.height - layoutHeaderReservedRows - layoutSortHintRows)
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
	for i := range m.tables.rarities {
		byProperty := m.raritySortCol[i] == 1
		tuples := sortedPropertyCounts(raritySources[i], byProperty, m.raritySortDesc[i])
		rarityRows := make([]table.Row, len(tuples))
		for j := range tuples {
			rarityRows[j] = propertyCountToRow(tuples[j])
		}
		m.tables.rarities[i].table.SetRows(rarityRows)
		applyRaritySortHeaders(&m.tables.rarities[i].table, i, m.raritySortCol[i], m.raritySortDesc[i])
	}
}

func (m *model) selectRarityNeighbour(direction int) {
	if m.uiState != globalStats || !m.activeTable().table.Focused() {
		return
	}
	m.UnfocusSelectedTable()
	if direction < 0 {
		m.selectedRarityIdx = rarityNavLeft[m.selectedRarityIdx]
	} else {
		m.selectedRarityIdx = rarityNavRight[m.selectedRarityIdx]
	}
	m.FocusSelectedTable()
}

func (m *model) toggleGlobalView() {
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
