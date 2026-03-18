package tuiapp

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/airspottr/internal"
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

	uiState    uiState
	startTime  time.Time
	lastUpdate time.Time
	request    *internal.Request
	dashboard  *internal.Dashboard
	notify     *internal.Notify
	options    internal.RequestOptions
}

// activeTable is the table that receives focus and keyboard navigation for the current view.
func (m *model) activeTable() *autoFormatTable {
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
	m.FocusSelectedTable()
	m.tableStyle.Selected = m.baseStyle
	for i := rarityByOperator; i < rarityTableCount; i++ {
		m.tables.rarities[i].table.SetStyles(m.tableStyle)
		m.tables.rarities[i].table.Blur()
	}
	return tea.Batch(updateTick(), aircraftQueryTick(), requestAircraftDataCmd(m.request))
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
