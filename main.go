// Package main provides the flight tracking application
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	srv "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	"github.com/micutio/airspottr/internal/infrastructure/data"
	noti "github.com/micutio/airspottr/internal/infrastructure/notify"
	"github.com/micutio/airspottr/internal/infrastructure/observation"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
	"github.com/micutio/airspottr/tickerapp"
	"github.com/micutio/airspottr/tuiapp"
	"github.com/spf13/pflag"
)

const (
	// thisAppName is the name of this application as shown on notifications.
	thisAppName = "airspottr"
)

func main() {
	predefinedLocations := map[string][]float64{
		"hamburg":   {53.5511, 9.9937},
		"new-york":  {40.7128, -74.0060},
		"singapore": {1.3521, 103.8198},
	}

	var argIsUseTicker bool
	var argLatLon []float64
	var argLocation string

	setupCommandLineFlags(&argIsUseTicker, &argLatLon, &argLocation)
	pflag.Parse()

	if val, ok := predefinedLocations[argLocation]; ok {
		argLatLon = val
	}

	options := adsb.RequestOptions{
		Lat: argLatLon[0],
		Lon: argLatLon[1],
	}

	if argIsUseTicker {
		app, err := buildTickerApp(thisAppName, options)
		if err != nil {
			panic(err)
		}
		tickerapp.RunWithDependencies(app)
		return
	}

	app, err := buildTUIApp(thisAppName, options)
	if err != nil {
		panic(err)
	}
	tuiapp.RunWithDependencies(app)
}

func buildTickerApp(appName string, options adsb.RequestOptions) (tickerapp.Dependencies, error) {
	stdout := os.Stdout
	stderr := os.Stderr
	logger := slog.Default()
	desktopNotify := noti.NewBeeepNotifier(appName, stdout)

	appState, appStateErr := pers.LoadState(pers.StateFilePath())
	if appStateErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("load app state: %w", appStateErr)
	}

	aircraftRequest, aircraftRequestErr := adsb.NewAircraftRequest(options, stderr)
	if aircraftRequestErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("new aircraft request: %w", aircraftRequestErr)
	}
	flightrouteRequest, flightrouteRequestErr := adsb.NewFlightrouteRequest(stderr)
	if flightrouteRequestErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("new flight route request: %w", flightrouteRequestErr)
	}

	typeRepo, typeRepoErr := data.NewAircraftTypeRepo()
	if typeRepoErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("new aircraft type repo: %w", typeRepoErr)
	}
	operatorRepo, operatorRepoErr := data.NewOperatorRepo()
	if operatorRepoErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("new operator repo: %w", operatorRepoErr)
	}
	countryRepo, countryRepoErr := data.NewCountryRepo()
	if countryRepoErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("new country repo: %w", countryRepoErr)
	}

	dashboard := srv.NewDashboard(
		options.Lat,
		options.Lon,
		observation.NewSightingRepo(),
		typeRepo,
		operatorRepo,
		countryRepo,
		stderr,
	)
	currentState := pers.StateForLocation(appState, options.Lat, options.Lon)
	if loadErr := dashboard.RestoreState(&currentState); loadErr != nil {
		return tickerapp.Dependencies{}, fmt.Errorf("restore dashboard state: %w", loadErr)
	}

	refreshUseCase := srv.NewRefreshUseCase(dashboard, desktopNotify, flightrouteRequest.GetFlightroutes)
	return tickerapp.Dependencies{
		AppName:            appName,
		Options:            options,
		Logger:             logger,
		AircraftRequest:    aircraftRequest,
		FlightrouteRequest: flightrouteRequest,
		Dashboard:          dashboard,
		RefreshUseCase:     refreshUseCase,
		Notify:             desktopNotify,
		SummaryNotify:      srv.NewNotify(appName, stdout),
		Done:               make(chan bool),
		Stdout:             stdout,
		Stderr:             stderr,
	}, nil
}

func buildTUIApp(appName string, options adsb.RequestOptions) (tuiapp.Dependencies, error) {
	errLogFile, err := os.OpenFile("./airspottr.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("open log file: %w", err)
	}
	aircraftReq, aircraftErr := adsb.NewAircraftRequest(options, errLogFile)
	if aircraftErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("new aircraft request: %w", aircraftErr)
	}
	flightrouteReq, flightrouteErr := adsb.NewFlightrouteRequest(errLogFile)
	if flightrouteErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("new flight route request: %w", flightrouteErr)
	}

	appState, appStateErr := pers.LoadState(pers.StateFilePath())
	if appStateErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("load app state: %w", appStateErr)
	}
	typeRepo, typeRepoErr := data.NewAircraftTypeRepo()
	if typeRepoErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("new aircraft type repo: %w", typeRepoErr)
	}
	operatorRepo, operatorRepoErr := data.NewOperatorRepo()
	if operatorRepoErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("new operator repo: %w", operatorRepoErr)
	}
	countryRepo, countryRepoErr := data.NewCountryRepo()
	if countryRepoErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("new country repo: %w", countryRepoErr)
	}

	dashboard := srv.NewDashboard(
		options.Lat,
		options.Lon,
		observation.NewSightingRepo(),
		typeRepo,
		operatorRepo,
		countryRepo,
		errLogFile,
	)
	currentState := pers.StateForLocation(appState, options.Lat, options.Lon)
	if loadErr := dashboard.RestoreState(&currentState); loadErr != nil {
		return tuiapp.Dependencies{}, fmt.Errorf("restore dashboard state: %w", loadErr)
	}
	dashboard.FinishWarmupPeriod()

	notify := noti.NewBeeepNotifier(appName, io.Discard)
	return tuiapp.Dependencies{
		AppName:             appName,
		AircraftSource:      aircraftReq,
		FlightRouteSource:   flightrouteReq,
		TypeRepo:            typeRepo,
		Dashboard:           dashboard,
		NotificationService: notify,
		RequestOptions:      options,
		ErrorLog:            errLogFile,
	}, nil
}

func setupCommandLineFlags(argIsUseTicker *bool, argLatLon *[]float64, argLocation *string) {
	// Whether to launch the Ticker or TUI app.
	pflag.BoolVarP(
		argIsUseTicker,
		"ticker",
		"t",
		false,
		"print plane spotting information on the command line without TUI")
	pflag.Lookup("ticker").NoOptDefVal = "true"

	// Location to plane spot, provided as lat,lon coordinates
	pflag.Float64SliceVarP(
		argLatLon,
		"latlon",
		"l",
		[]float64{0, 0},
		"define the location where to spot planes")

	pflag.StringVarP(
		argLocation,
		"location",
		"L",
		"",
		"define a predefined location, e.g. hamburg, new-york, singapore",
	)
}
