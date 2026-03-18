package tuiapp

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/micutio/airspottr/internal"
)

// Vertical space reserved above tables (header + stats blocks).
const layoutHeaderReservedRows = 8

// Rarity tables on the global stats view share width in thirds.
const globalStatsTableCount = 3

// Gap between joined global-stats tables (lipgloss borders/spacing).
const layoutGlobalStatsInterTableGap = 2

func (m *model) resizeTables() {
	m.tables.setAllHeights(m.height - layoutHeaderReservedRows)
	m.tables.resizeForTerminal(m.width, func(err error) {
		m.notify.Stdout.Panicf("%s", err)
	})
}

func setRarityRows(tbl *autoFormatTable, counts map[string]int) {
	rarities := internal.GetSortedCountsForProperty(counts)
	rows := make([]table.Row, len(rarities))
	for i := range rarities {
		rows[i] = propertyCountToRow(rarities[i])
	}
	tbl.table.SetRows(rows)
}

func (m *model) updateAllTables() {
	var currentAircraftRows []table.Row
	for _, aircraft := range m.dashboard.CurrentAircraft {
		aircraftType := m.dashboard.IcaoToAircraft[aircraft.IcaoType].Make
		flightRoute, ok := m.dashboard.CachedFlightRoutes[aircraft.GetFlightNoAsStr()]
		if !ok {
			flightRoute = internal.GetDefaultFlightrouteRecord()
		}
		if aircraft.GetFlightNoAsStr() == "" && aircraftType == "" {
			continue
		}
		currentAircraftRows = append(currentAircraftRows, aircraftToRow(&aircraft, flightRoute))
	}
	m.tables.aircraft.table.SetRows(currentAircraftRows)

	raritySources := [rarityTableCount]map[string]int{
		m.dashboard.SeenTypeCount,
		m.dashboard.SeenOperatorCount,
		m.dashboard.SeenCountryCount,
	}
	for i := range m.tables.rarities {
		setRarityRows(&m.tables.rarities[i], raritySources[i])
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
