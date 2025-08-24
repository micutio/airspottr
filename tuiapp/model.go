package tuiapp

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/flighttrack/internal"
)

type TickMsg time.Time

type ADSBResponseMsg []byte

func tick() tea.Cmd {
	return tea.Every(
		internal.AircraftUpdateInterval,
		func(t time.Time) tea.Msg {
			return TickMsg(t)
		},
	)
}

func requestADSBDataCmd() tea.Cmd {
	return func() tea.Msg {
		body, err := internal.RequestAndProcessCivAircraft()
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
type model struct {
	width              int
	height             int
	baseStyle          lipgloss.Style
	viewStyle          lipgloss.Style
	theme              Theme
	currentAircraftTbl table.Model
	typeRarityTbl      table.Model
	tableStyle         table.Styles
	lastUpdate         time.Time
	dashboard          *internal.Dashboard
}

// Init calls the tickEvery function to set up a command that sends a TickMsg every second.
// This command will be executed immediately when the program starts, initiating the periodic updates.
func (m *model) Init() tea.Cmd {
	return tick()
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

	// message is sent when a key is pressed.
	case tea.KeyMsg:
		switch thisMsg.String() {
		// Toggles the focus state of the aircraft table
		case "esc":
			if m.currentAircraftTbl.Focused() {
				m.tableStyle.Selected = m.baseStyle
				m.currentAircraftTbl.SetStyles(m.tableStyle)
				m.currentAircraftTbl.Blur()
			} else {
				m.tableStyle.Selected = m.tableStyle.Selected.Background(m.theme.Highlight)
				m.currentAircraftTbl.SetStyles(m.tableStyle)
				m.currentAircraftTbl.Focus()
			}
		// Moves the focus up in the aircraft table if the table is focused.
		case "up", "k":
			if m.currentAircraftTbl.Focused() {
				m.currentAircraftTbl.MoveUp(1)
			}
		// Moves the focus down in the aircraft table if the table is focused.
		case "down", "j":
			if m.currentAircraftTbl.Focused() {
				m.currentAircraftTbl.MoveDown(1)
			}
		// Quits the program by returning the tea.Quit command.
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case TickMsg:
		m.lastUpdate = time.Time(thisMsg)
		return m, tea.Batch(requestADSBDataCmd(), tick())
	case ADSBResponseMsg:
		responseBody := []byte(thisMsg)
		m.dashboard.ProcessCivAircraftJSON(responseBody)
		var rows []table.Row
		for _, aircraft := range m.dashboard.CurrentAircraft {
			rows = append(rows, table.Row{
				fmt.Sprintf("%3.0f", aircraft.CachedDist),
				aircraft.GetFlightNoAsStr(),
				m.dashboard.IcaoToAircraft[aircraft.IcaoType].ModelCode,
				aircraft.GetAltitudeAsStr(),
				fmt.Sprintf("%3.0f", aircraft.GroundSpeed),
				fmt.Sprintf("%3.0f", aircraft.NavHeading),
			})
			m.currentAircraftTbl.SetRows(rows)
		}
		return m, nil // since we've already scheduled the next request, there is nothing to do now.
	}

	// If the message type does not match any of the handled cases, the model is returned unchanged,
	// and no new command is issued.
	return m, nil
}

func (m *model) View() string {
	// Sets the width of the column to the width of the terminal (m.width) and adds padding of 1 unit
	// on the top.
	// Render is a method from the lipgloss package that applies the defined style and returns
	// a function that can render styled content.
	column := m.baseStyle.Width(m.width).Padding(1, 0, 0, 0).Render
	// Set the content to match the terminal dimensions (m.width and m.height).
	content := m.baseStyle.
		Width(m.width).
		Height(m.height).
		Render(
			// Vertically join multiple elements aligned to the left.
			lipgloss.JoinVertical(lipgloss.Left,
				column(m.viewHeader()),
				column(m.viewAircraft()),
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
		Height(4).
		Padding(0, 1)

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

	return m.viewStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			fmt.Sprintf("Last update: %d seconds ago\n", time.Since(m.lastUpdate).Seconds()),
			list.Border(lipgloss.NormalBorder(), false).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					listHeader("Highest"),
					lipgloss.JoinHorizontal(
						lipgloss.Left,
						listItem("ALT", highest.GetAltitudeAsStr()),
						listItem("FNO", highest.GetFlightNoAsStr()),
						listItem("TID", m.dashboard.IcaoToAircraft[highest.IcaoType].ModelCode),
						listItem("REG", highest.Registration),
					),
				),
			),
			list.Border(lipgloss.NormalBorder(), false).Render(
				lipgloss.JoinVertical(lipgloss.Left,
					listHeader("Fastest"),
					lipgloss.JoinHorizontal(
						lipgloss.Left,
						listItem("SPD", fmt.Sprintf("%3.0f", fastest.GroundSpeed)),
						listItem("FNO", fastest.GetFlightNoAsStr()),
						listItem("TID", m.dashboard.IcaoToAircraft[fastest.IcaoType].ModelCode),
						listItem("REG", fastest.Registration),
					),
				),
			),
		),
	)
}

func (m *model) viewAircraft() string {
	return m.viewStyle.Render(m.currentAircraftTbl.View())
}
