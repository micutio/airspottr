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
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
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

func Run(requestOptions internal.RequestOptions) {
	theme := getDefaultTheme()

	// STEP 1: Create logs and dashboard. ////////////////////////////////////////////////////////
	errLogFile, fileErr := os.OpenFile(
		errLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) // 0o666
	if fileErr != nil {
		log.Fatalf("Failed to open log file: %v", fileErr)
	}
	defer func() {
		closeErr := errLogFile.Close()
		if closeErr != nil {
			fileErr = fmt.Errorf(
				"tuiApp.Run(): error while closing file %s: %w",
				errLogFilePath,
				closeErr)
		}
	}()

	consoleParams := internal.LogParams{
		ConsoleOut: io.Discard,
		ErrorOut:   errLogFile,
	}

	dashboard, dashErr := internal.NewDashboard(requestOptions.Lat, requestOptions.Lon, consoleParams)
	if dashErr != nil {
		log.Println(fmt.Errorf("tuiapp.Run: %w", dashErr))
		return
	}

	// TODO: Introduce extra command and message to finish warmup period.
	dashboard.FinishWarmupPeriod()

	// STEP 2: Initialise visual styles for tables. //////////////////////////////////////////////
	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Highlight)

	// Create a new table with specified columns and initial empty rows.
	maxTypeLen := 0
	for _, value := range dashboard.IcaoToAircraft {
		if len(value.Make) > maxTypeLen {
			maxTypeLen = len(value.Make)
		}
	}

	dstLen := 7
	fnoLen := 10
	spdLen := 5
	initialTableHeight := 5

	currentAircraftTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "DST", Width: dstLen},
				{Title: "FNO", Width: fnoLen},
				{Title: "TID", Width: maxTypeLen},
				{Title: "ALT", Width: dstLen},
				{Title: "SPD", Width: spdLen},
				{Title: "HDG", Width: spdLen},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
		table.WithStyles(tableStyle),
	)

	// Create a new table with specified columns and initial empty rows.
	typeRarityTbl := table.New(
		// table header
		table.WithColumns(
			[]table.Column{
				{Title: "Count", Width: fnoLen},
				{Title: "Type", Width: maxTypeLen},
			},
		),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(initialTableHeight),
		table.WithStyles(tableStyle),
	)

	// STEP 3: Initialise model and run the application. /////////////////////////////////////////
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
		dashboard:          dashboard,
		options:            requestOptions,
	}
	// Create a new Bubble Tea program with the appModel and enable alternate screen
	p := tea.NewProgram(&appModel, tea.WithAltScreen())

	// Run the program and handle any errors
	if _, progErr := p.Run(); progErr != nil {
		log.Println(fmt.Errorf("error running program: %w", progErr))
	}
}
