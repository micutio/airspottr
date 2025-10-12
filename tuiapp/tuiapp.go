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
	"log" //nolint:depguard // Don't feel like using slog for now.
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

// setupLogger creates and configures the error log file.
func setupLogger() (*os.File, error) {
	errLogFile, err := os.OpenFile(errLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return errLogFile, nil
}

// setupDashboardAndNotifier initializes the dashboard and notification system.
func setupDashboardAndNotifier(
	appName string,
	requestOptions internal.RequestOptions,
	errWriter io.Writer,
) (*internal.Dashboard, *internal.Notify, error) {
	// Using io.Discard for notifications as we don't need to close it
	devNullWriter := io.Discard
	notify := internal.NewNotify(appName, &devNullWriter)

	dashboard, err := internal.NewDashboard(requestOptions.Lat, requestOptions.Lon, &errWriter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dashboard: %w", err)
	}

	return dashboard, notify, nil
}

type tableSetup struct {
	current   autoFormatTable
	types     autoFormatTable
	operators autoFormatTable
	countries autoFormatTable
	style     table.Styles
}

// initTables creates and configures all tables used in the TUI.
func initTables(theme Theme) tableSetup {
	tableStyle := table.DefaultStyles()
	tableStyle.Header.Padding(0)
	tableStyle.Cell.Padding(0)
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Highlight)

	return tableSetup{
		current:   newCurrentAircraftTable(tableStyle),
		types:     newTypeRarityTable(tableStyle),
		operators: newOperatorRarityTable(tableStyle),
		countries: newCountryRarityTable(tableStyle),
		style:     tableStyle,
	}
}

func Run(appName string, requestOptions internal.RequestOptions) {
	// Set up logging
	errLogFile, err := setupLogger()
	if err != nil {
		log.Fatalf("failed to set up logging: %v", err)
	}
	defer func() {
		if closeErr := errLogFile.Close(); err != nil {
			log.Printf("error closing log file: %v", closeErr)
		}
	}()

	// Initialize dashboard and notification system
	dashboard, notify, err := setupDashboardAndNotifier(appName, requestOptions, errLogFile)
	if err != nil {
		log.Printf("failed to set up dashboard and notifier: %v", err)
	}

	// TODO: Introduce extra command and message to finish warmup period.
	dashboard.FinishWarmupPeriod()

	// Initialize tables and theme
	theme := getDefaultTheme()
	tables := initTables(theme)

	// Initialize and run the application model
	appModel := model{
		width:              0,
		height:             0,
		baseStyle:          lipgloss.NewStyle(),
		viewStyle:          lipgloss.NewStyle(),
		theme:              theme,
		currentAircraftTbl: tables.current,
		typeRarityTbl:      tables.types,
		operatorRarityTbl:  tables.operators,
		countryRarityTbl:   tables.countries,
		tableStyle:         tables.style,
		startTime:          time.Now(),
		lastUpdate:         time.Unix(0, 0),
		dashboard:          dashboard,
		notify:             notify,
		options:            requestOptions,
	}

	// Create and run Bubble Tea program with alternate screen
	p := tea.NewProgram(&appModel, tea.WithAltScreen())
	if _, progErr := p.Run(); progErr != nil {
		log.Printf("error running program: %v", progErr)
	}
}
