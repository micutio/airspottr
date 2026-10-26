package tuiapp

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	srv "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
)

type uiState int

const (
	mainPage uiState = iota
	globalStats
)

// inputFocus determines whether keys target the table(s) or the notify checkbox strip.
type inputFocus int

const (
	focusTable inputFocus = iota
	focusNotifyStrip
)

// Vertical space reserved above and below tables.
const (
	layoutHeaderReservedRows = 8
	layoutFooterRows         = 1
)

// model implements tea.Model (Init, Update, View).
type model struct {
	width  int
	height int

	baseStyle  lipgloss.Style
	viewStyle  lipgloss.Style
	tableStyle table.Styles
	theme      Theme

	tables            tuiTables
	selectedRarityIdx int // rarityBy*; only used when uiState == globalStats

	uiState         uiState
	startTime       time.Time
	lastUpdate      time.Time
	aircraftRepo    srv.AircraftDataSource
	flightrouteRepo srv.FlightRouteDataSource
	dashboard       *srv.Dashboard
	options         adsb.RequestOptions

	inputFocus      inputFocus
	notifyStripIdx  int // 0=type, 1=operator, 2=country when focusNotifyStrip
	notifyOnType    bool
	notifyOnOp      bool
	notifyOnCountry bool
}

// activeTable is the table that receives focus and keyboard navigation for the current view.
func (m *model) activeTable() *sortableTable {
	if m.uiState == globalStats {
		return &m.tables.rarities[m.selectedRarityIdx]
	}
	return &m.tables.aircraft
}

// Init schedules ticks and the first aircraft fetch.
func (m *model) Init() tea.Cmd {
	tea.SetWindowTitle("airspottr")
	m.uiState = mainPage
	m.selectedRarityIdx = rarityByType
	m.inputFocus = focusTable
	m.FocusSelectedTable()
	m.tableStyle.Selected = m.baseStyle
	for i := rarityByOperator; i < rarityTableCount; i++ {
		m.tables.rarities[i].table.SetStyles(m.tableStyle)
		m.tables.rarities[i].table.Blur()
	}
	return tea.Batch(updateTick(), aircraftQueryTick(), requestAircraftDataCmd(m.aircraftRepo))
}

func (m *model) UnfocusSelectedTable() {
	m.tableStyle.Selected = m.baseStyle
	at := m.activeTable()
	at.table.SetStyles(m.tableStyle)
	at.table.Blur()
}

func (m *model) FocusSelectedTable() {
	m.tableStyle.Selected = m.tableStyle.Selected.Background(m.theme.Highlight)
	at := m.activeTable()
	at.table.SetStyles(m.tableStyle)
	at.table.Focus()
}

func (m *model) resizeTables() {
	m.tables.setAllHeights(m.height - layoutHeaderReservedRows - layoutFooterRows)
	m.tables.resizeForTerminal(m.width)
}

func (m *model) updateAllTables() {
	acTbl := &m.tables.aircraft
	records := filteredSortedSightings(
		m.dashboard.GetCurrentSightings(),
		acTbl.sortCol,
		acTbl.sortDesc,
	)
	acTbl.SetRows(buildAircraftRows(records))
	acTbl.syncColumns()

	raritySources := [rarityTableCount]map[string]int{
		m.dashboard.SeenType,
		m.dashboard.SeenOperator,
		m.dashboard.SeenCountry,
	}
	for idx := range m.tables.rarities {
		rTbl := &m.tables.rarities[idx]
		byProperty := rTbl.sortCol == 1
		tuples := sortedPropertyCounts(raritySources[idx], byProperty, rTbl.sortDesc)
		rarityRows := make([]table.Row, len(tuples))
		for j := range tuples {
			rarityRows[j] = propertyCountToRow(tuples[j])
		}
		rTbl.SetRows(rarityRows)
		rTbl.syncColumns()
	}
}

func (m *model) selectRarityNeighbour(direction int) {
	if m.uiState != globalStats || !m.activeTable().table.Focused() {
		return
	}
	m.UnfocusSelectedTable()
	m.selectedRarityIdx = (m.selectedRarityIdx + direction + rarityTableCount) % rarityTableCount
	m.FocusSelectedTable()
}

func (m *model) toggleGlobalView() {
	if m.inputFocus == focusNotifyStrip {
		return
	}
	m.UnfocusSelectedTable()
	if m.uiState == mainPage {
		m.uiState = globalStats
		m.selectedRarityIdx = rarityByType
	} else {
		m.uiState = mainPage
	}
	m.FocusSelectedTable()
}
