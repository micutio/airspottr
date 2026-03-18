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
	tableHeight := m.height - layoutHeaderReservedRows
	m.currentAircraftTbl.SetHeight(tableHeight)
	m.typeRarityTbl.SetHeight(tableHeight)
	m.operatorRarityTbl.SetHeight(tableHeight)
	m.countryRarityTbl.SetHeight(tableHeight)

	leftSideWidth := m.width - 1
	rightSideWidth := m.width
	third := 1.0 / float64(globalStatsTableCount)
	rightSideTableWidth := int(float64(rightSideWidth) * third)

	if caErr := m.currentAircraftTbl.resize(leftSideWidth); caErr != nil {
		m.notify.Stdout.Panicf("%s", caErr)
	}
	if trErr := m.typeRarityTbl.resize(rightSideTableWidth); trErr != nil {
		m.notify.Stdout.Panicf("%s", trErr)
	}
	if orErr := m.operatorRarityTbl.resize(rightSideTableWidth); orErr != nil {
		m.notify.Stdout.Panicf("%s", orErr)
	}
	countryWidth := rightSideWidth -
		2*rightSideTableWidth -
		layoutGlobalStatsInterTableGap -
		len(m.countryRarityTbl.table.Columns())
	if crErr := m.countryRarityTbl.resize(countryWidth); crErr != nil {
		m.notify.Stdout.Panicf("%s", crErr)
	}
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
	m.currentAircraftTbl.table.SetRows(currentAircraftRows)

	setRarityRows(&m.typeRarityTbl, m.dashboard.SeenTypeCount)
	setRarityRows(&m.operatorRarityTbl, m.dashboard.SeenOperatorCount)
	setRarityRows(&m.countryRarityTbl, m.dashboard.SeenCountryCount)
}

func (m *model) selectTableToTheLeft() {
	if !m.selectedTable.table.Focused() {
		return
	}
	if m.selectedTable == &m.currentAircraftTbl {
		return
	}
	m.UnfocusSelectedTable()
	switch m.selectedTable {
	case &m.typeRarityTbl:
		m.selectedTable = &m.countryRarityTbl
	case &m.operatorRarityTbl:
		m.selectedTable = &m.typeRarityTbl
	case &m.countryRarityTbl:
		m.selectedTable = &m.operatorRarityTbl
	}
	m.FocusSelectedTable()
}

func (m *model) selectTableToTheRight() {
	if !m.selectedTable.table.Focused() {
		return
	}
	if m.selectedTable == &m.currentAircraftTbl {
		return
	}
	m.UnfocusSelectedTable()
	switch m.selectedTable {
	case &m.typeRarityTbl:
		m.selectedTable = &m.operatorRarityTbl
	case &m.operatorRarityTbl:
		m.selectedTable = &m.countryRarityTbl
	case &m.countryRarityTbl:
		m.selectedTable = &m.typeRarityTbl
	}
	m.FocusSelectedTable()
}

func (m *model) toggleGlobalView() {
	switch m.uiState {
	case mainPage:
		m.uiState = globalStats
		m.selectedTable.table.Blur()
		m.selectedTable = &m.typeRarityTbl
		m.selectedTable.table.Focus()
	case globalStats:
		m.uiState = mainPage
		m.selectedTable.table.Blur()
		m.selectedTable = &m.currentAircraftTbl
		m.selectedTable.table.Focus()
	case aircraftDetails:
	default:
	}
}
