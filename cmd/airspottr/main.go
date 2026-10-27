// Package main provides the flight tracking application
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/pflag"

	srv "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	"github.com/micutio/airspottr/internal/infrastructure/data"
	noti "github.com/micutio/airspottr/internal/infrastructure/notify"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
	"github.com/micutio/airspottr/tickerapp"
	"github.com/micutio/airspottr/tuiapp"
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

func buildCoreServices(
	options adsb.RequestOptions,
	errOut io.Writer,
) (*adsb.AircraftRequest, *adsb.FlightrouteRequest, *srv.Dashboard, error) {
	appState, appStateErr := pers.LoadState(pers.StateFilePath())
	if appStateErr != nil {
		return nil, nil, nil, fmt.Errorf("load app state: %w", appStateErr)
	}

	aircraftReq, aircraftErr := adsb.NewAircraftRequest(options, errOut)
	if aircraftErr != nil {
		return nil, nil, nil, fmt.Errorf("new aircraft request: %w", aircraftErr)
	}
	flightrouteReq, flightrouteErr := adsb.NewFlightrouteRequest(errOut)
	if flightrouteErr != nil {
		return nil, nil, nil, fmt.Errorf("new flight route request: %w", flightrouteErr)
	}

	classifier, classifierErr := data.NewClassifier(errOut)
	if classifierErr != nil {
		return nil, nil, nil, fmt.Errorf("load reference data: %w", classifierErr)
	}

	dashboard := srv.NewDashboard(options.Lat, options.Lon, classifier, errOut)
	if currentState, ok := pers.StateForLocation(appState, options.Lat, options.Lon); ok {
		if loadErr := dashboard.RestoreState(&currentState); loadErr != nil {
			return nil, nil, nil, fmt.Errorf("restore dashboard state: %w", loadErr)
		}
	}

	return aircraftReq, flightrouteReq, dashboard, nil
}

func buildTickerApp(appName string, options adsb.RequestOptions) (tickerapp.Dependencies, error) {
	stdout := os.Stdout
	stderr := os.Stderr
	aircraftReq, flightrouteReq, dashboard, err := buildCoreServices(options, stderr)
	if err != nil {
		return tickerapp.Dependencies{}, err
	}

	desktopNotify := noti.NewBeeepNotifier(appName, stdout)
	dashboard.ConfigureRefresh(desktopNotify, flightrouteReq.GetFlightroutes)
	return tickerapp.Dependencies{
		AppName:            appName,
		Options:            options,
		Logger:             slog.New(slog.NewTextHandler(stderr, nil)),
		AircraftRequest:    aircraftReq,
		FlightrouteRequest: flightrouteReq,
		Dashboard:          dashboard,
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
	aircraftReq, flightrouteReq, dashboard, err := buildCoreServices(options, errLogFile)
	if err != nil {
		_ = errLogFile.Close()
		return tuiapp.Dependencies{}, err
	}
	dashboard.FinishWarmupPeriod()

	return tuiapp.Dependencies{
		AppName:             appName,
		AircraftSource:      aircraftReq,
		FlightRouteSource:   flightrouteReq,
		Dashboard:           dashboard,
		NotificationService: noti.NewBeeepNotifier(appName, io.Discard),
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
