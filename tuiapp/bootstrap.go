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

// setupRequestsAndDashboard initializes the dashboard and notification system.
func setupAircraftRequest(
	requestOptions adsb.RequestOptions,
	errWriter io.Writer,
) (*adsb.AircraftRequest, error) {
	aircraftReq, aircraftReqErr := adsb.NewAircraftRequest(requestOptions, &errWriter)
	if aircraftReqErr != nil {
		return nil, fmt.Errorf("failed to create aircraft request: %w", aircraftReqErr)
	}

	return aircraftReq, nil
}

// setupRequestsAndDashboard initializes the dashboard and notification system.
func setupFlightrouteRequest(
	errWriter io.Writer,
) (*adsb.FlightrouteRequest, error) {
	flightrouteReq, flightrouteReqErr := adsb.NewFlightrouteRequest(&errWriter)
	if flightrouteReqErr != nil {
		return nil, fmt.Errorf("failed to create flight request: %w", flightrouteReqErr)
	}

	return flightrouteReq, nil
}

// setupRequestsAndDashboard initializes the dashboard and notification system.
func setupDashboard(
	requestOptions adsb.RequestOptions,
	state pers.AirspottrState,
	errWriter io.Writer,
) (*internal.Dashboard, error) {
	dashboard := internal.NewDashboard(requestOptions.Lat, requestOptions.Lon, &errWriter)
	if loadErr := state.LoadDashboardState(dashboard); loadErr != nil {
		return nil, fmt.Errorf("warning: unable to load persisted dashboard state: %w", loadErr)
	}

	return dashboard, nil
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
