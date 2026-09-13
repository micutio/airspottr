package tuiapp

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	internal "github.com/micutio/airspottr/internal/application"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
)

const errLogFilePath = "./airspottr.log"

// setupLogger creates and configures the error log file.
func setupLogger() (*os.File, error) {
	errLogFile, err := os.OpenFile(errLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return errLogFile, nil
}

// setupRequestAndDashboard initializes the dashboard and notification system.
func setupRequestAndDashboard(
	requestOptions adsb.RequestOptions,
	errWriter io.Writer,
) (*adsb.Request, *internal.Dashboard, error) {
	request, reqErr := adsb.NewRequest(requestOptions, &errWriter)
	if reqErr != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", reqErr)
	}

	dashboard, dbErr := internal.NewDashboard(requestOptions.Lat, requestOptions.Lon, &errWriter)
	if dbErr != nil {
		return nil, nil, fmt.Errorf("failed to create dashboard: %w", dbErr)
	}

	if loadErr := pers.LoadState(pers.StateFilePath(), dashboard, request); loadErr != nil {
		return nil, nil, fmt.Errorf("warning: unable to load persisted state: %w", loadErr)
	}

	return request, dashboard, nil
}

type tableSetup struct {
	tables tuiTables
	style  table.Styles
}

// initTables creates and configures all tables used in the TUI.
func initTables(theme Theme) tableSetup {
	tableStyle := table.DefaultStyles()
	tableStyle.Header.Padding(0)
	tableStyle.Cell.Padding(0)
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Highlight)

	return tableSetup{
		tables: newTuiTables(tableStyle),
		style:  tableStyle,
	}
}
