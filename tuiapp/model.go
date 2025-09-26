package tuiapp

import (
	"fmt"
	"math"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/airspottr/internal"
)

type UpdateTickMsg time.Time

func updateTick() tea.Cmd {
	return tea.Every(
		time.Second,
		func(t time.Time) tea.Msg {
			return UpdateTickMsg(t)
		},
	)
}

type AircraftQueryTickMsg time.Time

func aircraftQueryTick() tea.Cmd {
	return tea.Every(
		internal.AircraftUpdateInterval,
		func(t time.Time) tea.Msg {
			return AircraftQueryTickMsg(t)
		},
	)
}

type ADSBResponseMsg []byte

func requestADSBDataCmd(opts internal.RequestOptions) tea.Cmd {
	return func() tea.Msg {
		body, err := internal.RequestAndProcessCivAircraft(opts)
		if err != nil {
			// TODO: Log error
			return nil
		}
		return ADSBResponseMsg(body)
	}
}

// Model implements the bubbletea.Model interface, which requires three methods:
// - Init() Cmd
// - Update(Msg) (Model, Cmd)
// - View() string
// This forms the base for the TUI app.
// TODO: Consider wrapper struct Table {table, tableFormat, style}.
type model struct {
	width              int
	height             int
	baseStyle          lipgloss.Style
	viewStyle          lipgloss.Style
	theme              Theme
	currentAircraftTbl autoFormatTable
	typeRarityTbl      autoFormatTable
	operatorRarityTbl  autoFormatTable
	countryRarityTbl   autoFormatTable
	tableStyle         table.Styles
	startTime          time.Time
	lastUpdate         time.Time
	dashboard          *internal.Dashboard
	notify             *internal.Notify
	options            internal.RequestOptions
}

// Init calls the tickEvery function to set up a command that sends a TickMsg every second.
// This command will be executed immediately when the program starts, initiating the periodic updates.
func (m *model) Init() tea.Cmd {
	return tea.Batch(updateTick(), aircraftQueryTick(), requestADSBDataCmd(m.options))
}

// Update takes a tea.Msg as input and uses a type switch to handle different types of messages.
// Each case in the switch statement corresponds to a specific message type.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn // required by interface
	switch thisMsg := msg.(type) {
	// message is sent when the window size changes
	// save to reflect the new dimensions of the terminal window.
	case tea.WindowSizeMsg:
		m.height = thisMsg.Height
		m.width = thisMsg.Width
		m.resizeTables()
	// message is sent when a key is pressed.
	case tea.KeyMsg:
		return m, m.processKeyMsg(thisMsg)
	case UpdateTickMsg:
		return m, updateTick()
	case AircraftQueryTickMsg:
		return m, tea.Batch(requestADSBDataCmd(m.options), aircraftQueryTick())
	case ADSBResponseMsg:
		m.processADSBResponse(thisMsg)
		return m, nil
	}

	// If the message type does not match any of the handled cases, the model is returned unchanged,
	// and no new command is issued.
	return m, nil
}

func (m *model) resizeTables() {
	headerHeight := 10 // TODO: Make this cleaner and clearer.

	m.currentAircraftTbl.SetHeight(m.height - headerHeight)
	m.typeRarityTbl.SetHeight(m.height - headerHeight)
	m.operatorRarityTbl.SetHeight(m.height - headerHeight)
	m.countryRarityTbl.SetHeight(m.height - headerHeight)

	// TODO: Set type column width of current aircraft table to variable size.

	// Adjust widths of all tables
	leftSideWidthRatio := 0.5
	leftSideWidth := int(float64(m.width) * leftSideWidthRatio)
	rightSideWidth := m.width - leftSideWidth
	rightSideTableCount := 3.0
	rightSideTableRatio := 1.0 / rightSideTableCount
	rightSideTableWidth := int(float64(rightSideWidth) * rightSideTableRatio)

	caErr := m.currentAircraftTbl.resize(leftSideWidth - 2 - len(m.currentAircraftTbl.table.Columns()))
	if caErr != nil {
		m.notify.Stdout.Panicf("%s", caErr)
	}
	trErr := m.typeRarityTbl.resize(rightSideTableWidth - 2 - len(m.typeRarityTbl.table.Columns()))
	if trErr != nil {
		m.notify.Stdout.Panicf("%s", trErr)
	}
	orErr := m.operatorRarityTbl.resize(rightSideTableWidth - 2 - len(m.operatorRarityTbl.table.Columns()))
	if orErr != nil {
		m.notify.Stdout.Panicf("%s", orErr)
	}
	crErr := m.countryRarityTbl.resize(
		rightSideWidth -
			rightSideTableWidth -
			rightSideTableWidth -
			2 - len(m.countryRarityTbl.table.Columns()))
	if crErr != nil {
		m.notify.Stdout.Panicf("%s", crErr)
	}
}

func (m *model) processKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	// Toggles the focus state of the aircraft table
	case "esc":
		if m.currentAircraftTbl.table.Focused() {
			m.tableStyle.Selected = m.baseStyle
			m.currentAircraftTbl.table.SetStyles(m.tableStyle)
			m.currentAircraftTbl.table.Blur()
		} else {
			m.tableStyle.Selected = m.tableStyle.Selected.Background(m.theme.Highlight)
			m.currentAircraftTbl.table.SetStyles(m.tableStyle)
			m.currentAircraftTbl.table.Focus()
		}
	// Moves the focus up in the aircraft table if the table is focused.
	case "up", "k":
		if m.currentAircraftTbl.table.Focused() {
			m.currentAircraftTbl.table.MoveUp(1)
		}
	// Moves the focus down in the aircraft table if the table is focused.
	case "down", "j":
		if m.currentAircraftTbl.table.Focused() {
			m.currentAircraftTbl.table.MoveDown(1)
		}
	case "h":
		if m.currentAircraftTbl.table.Focused() {
			m.currentAircraftTbl.table.HelpView()
		}
	// Quits the program by returning the tea.Quit command.
	case "q", "ctrl+c":
		return tea.Quit
	}
	return nil
}

// processADSBResponse processes new data from the ADS-B data source and updates the tables accordingly.
func (m *model) processADSBResponse(msg ADSBResponseMsg) {
	m.lastUpdate = time.Now()
	responseBody := []byte(msg)
	m.dashboard.ProcessCivAircraftJSON(responseBody)

	// Update current aircraft table.
	currentAircraftRows := make([]table.Row, len(m.dashboard.CurrentAircraft))
	for idx, aircraft := range m.dashboard.CurrentAircraft {
		aircraftType := m.dashboard.IcaoToAircraft[aircraft.IcaoType].Make

		// Filter out aircraft where both flight number and type are unknown.
		if aircraft.Flight == "" && aircraftType == "" {
			continue
		}

		currentAircraftRows[idx] = aircraftToRow(&aircraft)
	}
	m.currentAircraftTbl.table.SetRows(currentAircraftRows)

	// Update current type rarity table.
	// typeRarities := m.dashboard.GetTypeRarities()
	typeRarities := internal.GetSortedCountsForProperty(m.dashboard.SeenTypeCount)
	typeRarityRows := make([]table.Row, len(typeRarities))
	for typeIdx := range typeRarities {
		propertyCountToRow(typeRarities[typeIdx])
	}
	m.typeRarityTbl.table.SetRows(typeRarityRows)

	// Update current operator rarity table.
	// operatorRarities := m.dashboard.GetOperatorRarities()
	operatorRarities := internal.GetSortedCountsForProperty(m.dashboard.SeenOperatorCount)
	operatorRarityRows := make([]table.Row, len(operatorRarities))
	for operatorIdx := range operatorRarities {
		propertyCountToRow(operatorRarities[operatorIdx])
	}
	m.operatorRarityTbl.table.SetRows(operatorRarityRows)

	// Update current type rarity table.
	// countryRarities := m.dashboard.GetCountryRarities()
	countryRarities := internal.GetSortedCountsForProperty(m.dashboard.SeenCountryCount)
	countryRarityRows := make([]table.Row, len(countryRarities))
	for countryIdx := range countryRarities {
		propertyCountToRow(countryRarities[countryIdx])
	}
	m.countryRarityTbl.table.SetRows(countryRarityRows)

	// finally send out notifications for any rare sightings that occurred
	m.notify.EmitRarityNotifications(m.dashboard.RareSightings)

	// since we've already scheduled the next request, there is nothing to return now.
}

func (m *model) View() string {
	// Sets the width of the column to the width of the terminal (m.width) and adds padding of 1 unit
	// on the top.
	// Render is a method from the lipgloss package that applies the defined style and returns
	// a function that can render styled content.
	column := m.baseStyle.Width(m.width).Padding(0, 0, 0, 0).Render
	// Set the content to match the terminal dimensions (m.width and m.height).
	content := m.baseStyle.
		Width(m.width).
		Height(m.height).
		Render(
			// Vertically join multiple elements aligned to the left.
			lipgloss.JoinVertical(lipgloss.Left,
				column(m.viewHeader()),
				column(
					lipgloss.JoinHorizontal(
						lipgloss.Top,
						m.viewAircraft(),
						m.viewTypeRarity(),
						m.viewOperatorRarity(),
						m.viewCountryRarity(),
					),
				),
			),
		)

	return content
}

// Uses lipgloss.JoinVertical and lipgloss.JoinHorizontal to arrange the header content.
// It displays the last update time and aircraft information a structured format.
func (m *model) viewHeader() string {
	// defines the style for list items, including borders, border color, height, and padding.
	list := m.baseStyle.
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(m.theme.Border).
		Height(1).
		Padding(0, 0)

	// Applies bold styling to the text.
	listHeader := m.baseStyle.Bold(true).Render

	// Helper function that formats a key-value pair with an optional suffix.
	// It aligns the value to the right and renders it with the specified style.
	listItem := func(key string, value string) string {
		listItemValue := m.baseStyle.Align(lipgloss.Right).Render(value)

		listItemKey := func(key string) string {
			return m.baseStyle.Render(key + ":")
		}

		return fmt.Sprintf("%s %s", listItemKey(key), listItemValue)
	}

	highest := m.dashboard.Highest
	fastest := m.dashboard.Fastest

	if highest == nil || fastest == nil {
		return ""
	}

	minutesInHour := 60.0
	secsInMinute := 60.0
	tSince := time.Since(m.startTime)
	hours := tSince.Hours()
	mins := math.Mod(math.Floor(tSince.Minutes()), minutesInHour)
	secs := math.Mod(math.Floor(tSince.Seconds()), secsInMinute)

	return m.viewStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			list.Border(lipgloss.RoundedBorder()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("Location   : %6.3f, %6.3f", m.dashboard.Lat, m.dashboard.Lon),
					fmt.Sprintf("UpTime     : %.0f Hr %2.0f Min %2.0f Sec", hours, mins, secs),
					fmt.Sprintf("Last Update: %2.0f seconds ago", time.Since(m.lastUpdate).Seconds())),
			),
			list.Border(lipgloss.RoundedBorder()).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					listHeader("Highest"),
					lipgloss.JoinHorizontal(
						lipgloss.Left,
						listItem("ALT", highest.GetAltitudeAsStr()),
						listItem("FNO", highest.GetFlightNoAsStr()),
						listItem("TID", m.dashboard.IcaoToAircraft[highest.IcaoType].Make),
						listItem("REG", highest.Registration),
					),
					listHeader("Fastest"),
					lipgloss.JoinHorizontal(
						lipgloss.Left,
						listItem("SPD", fmt.Sprintf("%3.0f", fastest.GroundSpeed)),
						listItem("FNO", fastest.GetFlightNoAsStr()),
						listItem("TID", m.dashboard.IcaoToAircraft[fastest.IcaoType].Make),
						listItem("REG", fastest.Registration),
					),
				),
			),
		),
	)
}

func (m *model) viewAircraft() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.currentAircraftTbl.table.View())
}

func (m *model) viewTypeRarity() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.typeRarityTbl.table.View())
}

func (m *model) viewOperatorRarity() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.operatorRarityTbl.table.View())
}

func (m *model) viewCountryRarity() string {
	return m.viewStyle.Border(lipgloss.RoundedBorder()).Render(m.countryRarityTbl.table.View())
}
