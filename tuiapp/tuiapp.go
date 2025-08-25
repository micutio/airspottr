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

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/airspottr/internal"
)

const (
	errLogFilePath = "./airspottr.log"
)

// TODO: enable to write all errors to file instead of stdout when running as tui.

func Run() {
	theme := getDefaultTheme()

	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Highlight)

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

	errLogFile, fileErr := os.OpenFile(errLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if fileErr != nil {
		log.Fatalf("Failed to open log file: %v", fileErr)
	}
	defer errLogFile.Close()

	consoleParams := internal.LogParams{
		ConsoleOut: io.Discard,
		ErrorOut:   errLogFile,
	}

	flightDash, dashErr := internal.NewDashboard(consoleParams)
	if dashErr != nil {
		log.Fatal(dashErr)
	}

	flightDash.FinishWarmupPeriod()

	appModel := model{
		width:              0,
		height:             0,
		baseStyle:          lipgloss.NewStyle(),
		viewStyle:          lipgloss.NewStyle(),
		theme:              getDefaultTheme(),
		currentAircraftTbl: currentAircraftTbl,
		typeRarityTbl:      typeRarityTbl,
		tableStyle:         tableStyle,
		startTime:          time.Now(),
		lastUpdate:         time.Unix(0, 0),
		dashboard:          flightDash,
	}
	// Create a new Bubble Tea program with the appModel and enable alternate screen
	p := tea.NewProgram(&appModel, tea.WithAltScreen())

	// Run the program and handle any errors
	if _, progErr := p.Run(); progErr != nil {
		log.Fatalf("Error running program: %v", progErr)
	}
}
