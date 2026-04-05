package tuiapp

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/airspottr/internal"
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
	requestOptions internal.RequestOptions,
	errWriter io.Writer,
) (*internal.Request, *internal.Dashboard, error) {
	request, reqErr := internal.NewRequest(requestOptions, &errWriter)
	if reqErr != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", reqErr)
	}

	dashboard, dbErr := internal.NewDashboard(requestOptions.Lat, requestOptions.Lon, &errWriter)
	if dbErr != nil {
		return nil, nil, fmt.Errorf("failed to create dashboard: %w", dbErr)
	}

	if loadErr := internal.LoadState(internal.StateFilePath(), dashboard, request); loadErr != nil {
		fmt.Fprintf(errWriter, "warning: unable to load persisted state: %v\n", loadErr)
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
