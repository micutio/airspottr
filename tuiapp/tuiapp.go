// Package tuiapp provides the TUI app which displays flight tracking data, updates continuously
// and can be interacted with.
// Layout idea:
// +-------------------------------------------------+
// | last update time: 00:00:00                      |
// |                                                 |
// | Highest Aircraft                                |
// | ALT: ... FNO: ... Type: ... REG: ...            |
// | Fastest Aircraft                                |
// | SPD: ... FNO: ... Type: ... REG: ...            |
// |  ________________________       ______________  |
// | | current aircraft table |     | rarity table | |
// | | entry 0                |     | entry 0      | |
// | | ...                    |     | ...          | |
// | | entry N                |     | entry M      | |
// |  ------------------------       --------------  |
// +-------------------------------------------------+
// .
package tuiapp

// TODO:
// - [ ] Set up ticks for ADS-B API requests to periodically update the dashboard
// - [ ] refactor the ADS-B API request to make them reusable for both ticker and tui apps.

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/flighttrack/internal"
)

func Run() {
	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(Color.Highlight)

	// Create a new table with specified columns and initial empty rows.
	currentAircraftTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "DST", Width: 8},
				{Title: "FNO", Width: 8},
				{Title: "TID", Width: 20}, // TODO: Find max length in icao table
				{Title: "ALT", Width: 8},
				{Title: "SPD", Width: 8},
				{Title: "HDG", Width: 8},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(20), // TODO: Tie to console window height
		table.WithStyles(tableStyle),
	)

	// Create a new table with specified columns and initial empty rows.
	typeRarityTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "Count", Width: 8},
				{Title: "Type", Width: 8},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(40), // TODO: Tie to console window height
		table.WithStyles(tableStyle),
	)
	m := model{
		currentAircraftTbl: currentAircraftTbl,
		typeRarityTbl:      typeRarityTbl,
		tableStyle:         tableStyle,
		baseStyle:          lipgloss.NewStyle(),
		viewStyle:          lipgloss.NewStyle(),
	}
	// Create a new Bubble Tea program with the model and enable alternate screen
	p := tea.NewProgram(&m, tea.WithAltScreen())

	// Run the program and handle any errors
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

// Model implements the bubbletea.Model interface, which requires three methods:
// - Init() Cmd
// - Update(Msg) (Model, Cmd)
// - View() string
// This forms the base for the TUI app.
type model struct {
	width     int
	height    int
	baseStyle lipgloss.Style
	viewStyle lipgloss.Style

	currentAircraftTbl table.Model
	typeRarityTbl      table.Model
	tableStyle         table.Styles

	lastUpdate time.Time
	dashboard  internal.Dashboard
}

type TickMsg time.Time
type FinishCooldownMsg struct{}

type ADSBRequestMsg struct{}
type ADSBResponseMsg []byte

func scheduleFinishCooldown() tea.Cmd {
	return tea.Every(
		1*time.Hour,
		func(t time.Time) tea.Msg { return FinishCooldownMsg{} },
	)
}

func scheduleNextADSBRequest() tea.Cmd {
	return tea.Every(
		30*time.Second,
		func(t time.Time) tea.Msg { return ADSBRequestMsg{} },
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

type Theme struct {
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Highlight lipgloss.AdaptiveColor
	Border    lipgloss.AdaptiveColor
	Green     lipgloss.AdaptiveColor
	Red       lipgloss.AdaptiveColor
}

var Color = Theme{
	Primary:   lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
	Secondary: lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"},
	Highlight: lipgloss.AdaptiveColor{Light: "#8b2def", Dark: "#8b2def"},
	Border:    lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"},
	Green:     lipgloss.AdaptiveColor{Light: "#00FF00", Dark: "#00FF00"},
	Red:       lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"},
}

// Init calls the tickEvery function to set up a command that sends a TickMsg every second.
// This command will be executed immediately when the program starts, initiating the periodic updates.
func (m *model) Init() tea.Cmd {
	// TODO: Setup all callbacks for ADS-B requests.
	return tea.Batch(tickEvery(), scheduleNextADSBRequest(), scheduleFinishCooldown())
}

// tickEvery returns a command which emits a TickMsg every second.
func tickEvery() tea.Cmd {
	// tea.Every function is a helper function from the Bubble Tea framework
	// that schedules a command to run at regular intervals.
	return tea.Every(time.Second,
		// Callback function that takes the current time (t time.Time) as a parameter and returns a message (tea.Msg).
		// This callback is invoked every second.
		func(t time.Time) tea.Msg {
			return TickMsg(t)
		})
}

// Update takes a tea.Msg as input and uses a type switch to handle different types of messages.
// Each case in the switch statement corresponds to a specific message type.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn // required by interface
	switch msg := msg.(type) {
	// message is sent when the window size changes
	// save to reflect the new dimensions of the terminal window.
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	// message is sent when a key is pressed.
	case tea.KeyMsg:
		switch msg.String() {
		// Toggles the focus state of the process table
		case "esc":
			if m.currentAircraftTbl.Focused() {
				m.tableStyle.Selected = m.baseStyle
				m.currentAircraftTbl.SetStyles(m.tableStyle)
				m.currentAircraftTbl.Blur()
			} else {
				m.tableStyle.Selected = m.tableStyle.Selected.Background(Color.Highlight)
				m.currentAircraftTbl.SetStyles(m.tableStyle)
				m.currentAircraftTbl.Focus()
			}
		// Moves the focus up in the process table if the table is focused.
		case "up", "k":
			if m.currentAircraftTbl.Focused() {
				m.currentAircraftTbl.MoveUp(1)
			}
		// Moves the focus down in the process table if the table is focused.
		case "down", "j":
			if m.currentAircraftTbl.Focused() {
				m.currentAircraftTbl.MoveDown(1)
			}
		// Quits the program by returning the tea.Quit command.
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case ADSBRequestMsg:
		return m, tea.Batch(requestADSBDataCmd(), scheduleNextADSBRequest())
	case FinishCooldownMsg:
		// TODO: complete this branch
	// This custom message is sent periodically by the tickEvery function.
	// The model's lastUpdate field is updated to the current time.
	// Fetching CPU Stats, Memory Stats & Processes
	// Returning Command: The tickEvery command is returned to ensure that the TickMsg
	//                    continues to be sent periodically.
	case TickMsg:
		m.lastUpdate = time.Time(msg)
		// cpuStats, err := GetCPUStats()
		// if err != nil {
		// 	slog.Error("Could not get CPU info", "error", err)
		// } else {
		//	m.CpuUsage = cpuStats
		// }

		// memStats, err := GetMEMStats()
		// if err != nil {
		// 	slog.Error("Could not get memory info", "error", err)
		// } else {
		// 	m.MemUsage = memStats
		// }

		// TODO: Update currentAircraftTbl with aircraft from dashboard
		// procs, err := GetProcesses(5)
		// if err != nil {
		// 	slog.Error("Could not get processes", "error", err)
		// } else {
		// 	rows := []table.Row{}
		//	for _, p := range procs {
		//		memString, memUnit := convertBytes(p.Memory)
		//		rows = append(rows, table.Row{
		//			fmt.Sprintf("%d", p.PID),
		//			p.Name,
		//			fmt.Sprintf("%.2f%%", p.CPUPercent),
		//			fmt.Sprintf("%s %s", memString, memUnit),
		//			p.Username,
		//			p.RunningTime,
		//		})
		//	}
		//	m.processTable.SetRows(rows)
		// }

		return m, tickEvery()
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
		BorderForeground(Color.Border).
		Height(4).
		Padding(0, 1)

	// Applies bold styling to the text.
	listHeader := m.baseStyle.Bold(true).Render

	// Helper function that formats a key-value pair with an optional suffix.
	// It aligns the value to the right and renders it with the specified style.
	listItem := func(key string, value string, suffix ...string) string {
		finalSuffix := ""
		if len(suffix) > 0 {
			finalSuffix = suffix[0]
		}

		listItemValue := m.baseStyle.Align(lipgloss.Right).Render(fmt.Sprintf("%s%s", value, finalSuffix))

		listItemKey := func(key string) string {
			return m.baseStyle.Render(key + ":")
		}

		return fmt.Sprintf("%s %s", listItemKey(key), listItemValue)
	}

	highest := m.dashboard.Highest
	fastest := m.dashboard.Fastest

	return m.viewStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			fmt.Sprintf("Last update: %d milliseconds ago\n", time.Since(m.lastUpdate).Milliseconds()),
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
