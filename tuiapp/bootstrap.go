package tuiapp

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	srv "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/domain/repositories"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
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
	state srv.AirspottrState,
	requestOptions adsb.RequestOptions,
	sightingRepo repositories.SightingRepo,
	aircraftTypeRepo repositories.AircraftTypeRepo,
	operatorRepo repositories.OperatorRepository,
	countryRepo repositories.CountryRepository,
	errWriter io.Writer,
) (*srv.Dashboard, error) {
	dashboard := srv.NewDashboard(
		requestOptions.Lat,
		requestOptions.Lon,
		sightingRepo,
		aircraftTypeRepo,
		operatorRepo,
		countryRepo,
		&errWriter)
	if loadErr := dashboard.RestoreState(&state.InternalState.DashboardState); loadErr != nil {
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
