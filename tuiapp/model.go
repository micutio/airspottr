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

	currentAircraftTbl autoFormatTable
	typeRarityTbl      autoFormatTable
	operatorRarityTbl  autoFormatTable
	countryRarityTbl   autoFormatTable
	selectedTable      *autoFormatTable

	uiState    uiState
	startTime  time.Time
	lastUpdate time.Time
	request    *internal.Request
	dashboard  *internal.Dashboard
	notify     *internal.Notify
	options    internal.RequestOptions
}

// Init schedules ticks and the first aircraft fetch.
func (m *model) Init() tea.Cmd {
	tea.SetWindowTitle("airspottr")
	m.selectedTable = &m.currentAircraftTbl
	m.FocusSelectedTable()
	m.tableStyle.Selected = m.baseStyle
	m.countryRarityTbl.table.SetStyles(m.tableStyle)
	m.countryRarityTbl.table.Blur()
	m.operatorRarityTbl.table.SetStyles(m.tableStyle)
	m.operatorRarityTbl.table.Blur()
	return tea.Batch(updateTick(), aircraftQueryTick(), requestAircraftDataCmd(m.request))
}

func (m *model) UnfocusSelectedTable() {
	m.tableStyle.Selected = m.baseStyle
	m.selectedTable.table.SetStyles(m.tableStyle)
	m.selectedTable.table.Blur()
}

func (m *model) FocusSelectedTable() {
	m.tableStyle.Selected = m.tableStyle.Selected.Background(m.theme.Highlight)
	m.selectedTable.table.SetStyles(m.tableStyle)
	m.selectedTable.table.Focus()
}
