// Package tickerapp launches the ticker application which writes out all updates to stdout and
// can be piped into other programs and processed further.
// This is in contrast to the TUI app, which works more like htop.
package tickerapp

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/micutio/airspottr/internal"
)

// TickerApp holds the state and dependencies for the ticker application.
type TickerApp struct {
	appName   string
	options   internal.RequestOptions
	logger    *slog.Logger
	request   *internal.Request
	dashboard *internal.Dashboard
	notify    *internal.Notify
	done      chan bool
	wg        sync.WaitGroup
}

// New creates and initializes a new TickerApp.
func New(appName string, options internal.RequestOptions, stdout, stderr io.Writer) (*TickerApp, error) {
	logger := slog.Default() // Or a custom logger
	notify := internal.NewNotify(appName, &stdout)

	dashboard, dashboardErr := internal.NewDashboard(options.Lat, options.Lon, &stderr)
	if dashboardErr != nil {
		return nil, fmt.Errorf("unable to create dashboard: %w", dashboardErr)
	}

	request, requestErr := internal.NewRequest(options, &stderr)
	if requestErr != nil {
		return nil, fmt.Errorf("unable to create request: %w", requestErr)
	}

	if loadErr := internal.LoadState(internal.StateFilePath(), dashboard, request); loadErr != nil {
		fmt.Fprintf(stderr, "warning: unable to load persisted state: %v\n", loadErr)
	}

	return &TickerApp{ //nolint:exhaustruct // no need to init waitgroup
		appName:   appName,
		options:   options,
		logger:    logger,
		request:   request,
		dashboard: dashboard,
		notify:    notify,
		done:      make(chan bool),
	}, nil
}

// Run is the main entry point for the ticker application.
func Run(appName string, options internal.RequestOptions) {
	app, err := New(appName, options, os.Stdout, os.Stderr)
	if err != nil {
		slog.Default().Error("failed to initialize ticker app", slog.Any("error", err))
		os.Exit(1)
	}

	fmt.Printf("%s launching at Lat: %.3f, Lon: %.3f\n", appName, options.Lat, options.Lon)

	app.start()
	app.waitForShutdown()
}

// start begins the application's main event loop in a goroutine.
func (app *TickerApp) start() {
	// Set a timeout for the warmup period.
	time.AfterFunc(internal.DashboardWarmup, func() {
		app.dashboard.FinishWarmupPeriod()
	})

	aircraftUpdateTicker := time.NewTicker(internal.AircraftUpdateInterval)
	summaryTicker := time.NewTicker(internal.SummaryInterval)

	app.wg.Go(func() {
		defer aircraftUpdateTicker.Stop()
		defer summaryTicker.Stop()

		for {
			select {
			case <-aircraftUpdateTicker.C:
				aircraftRecords := app.request.RequestAircraft()
				app.dashboard.ProcessAircraftRecords(aircraftRecords)
				app.notify.EmitRarityNotifications(
					app.dashboard.RareSightings,
					internal.DefaultRarityNotifyToggles(),
				)

				// This method checks whether we have flight routes in the cache for all sightings.
				callsignsWithoutRoute := app.dashboard.AssignRouteToCallsigns()
				if len(callsignsWithoutRoute) > 0 {
					// For flights without known route we query data from adsbdb.com.
					routes := app.request.RequestFlightRoutesForCallsigns(callsignsWithoutRoute)
					app.dashboard.AssignFlightRoutes(routes)
				}
			case <-summaryTicker.C:
				app.notify.PrintSummary(app.dashboard)
			case <-app.done:
				slog.Info("Stopping HTTP GET request routine.")
				return
			}
		}
	})
	// WaitGroup.Wait() is called in waitForShutdown() below
}

// waitForShutdown blocks until an interrupt or terminate signal is received.
func (app *TickerApp) waitForShutdown() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	app.logger.Info("Shutdown signal received, stopping...")
	close(app.done)
	// Wait for the main goroutine to finish.
	app.wg.Wait()
	if saveErr := internal.SaveState(internal.StateFilePath(), app.dashboard, app.request); saveErr != nil {
		app.logger.Error("failed to save persistent state", slog.Any("error", saveErr))
	}
}
