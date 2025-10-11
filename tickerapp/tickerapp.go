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
	appName    string
	options    internal.RequestOptions
	logger     *slog.Logger
	notify     *internal.Notify
	flightDash *internal.Dashboard
	done       chan bool
	wg         sync.WaitGroup
}

// New creates and initializes a new TickerApp.
func New(appName string, options internal.RequestOptions, stdout, stderr io.Writer) (*TickerApp, error) {
	logger := slog.Default() // Or your custom logger
	notify := internal.NewNotify(appName, &stdout)

	flightDash, err := internal.NewDashboard(options.Lat, options.Lon, &stderr)
	if err != nil {
		return nil, fmt.Errorf("unable to create dashboard: %w", err)
	}

	return &TickerApp{ //nolint:exhaustruct // no need to init waitgroup
		appName:    appName,
		options:    options,
		logger:     logger,
		notify:     notify,
		flightDash: flightDash,
		done:       make(chan bool),
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
		app.flightDash.FinishWarmupPeriod()
	})

	aircraftUpdateTicker := time.NewTicker(internal.AircraftUpdateInterval)
	summaryTicker := time.NewTicker(internal.SummaryInterval)

	app.wg.Go(func() {
		defer aircraftUpdateTicker.Stop()
		defer summaryTicker.Stop()

		for {
			select {
			case <-aircraftUpdateTicker.C:
				if body, err := internal.RequestAndProcessCivAircraft(app.options); err != nil {
					app.logger.Error("main: ", slog.Any("error", err))
				} else {
					app.flightDash.ProcessCivAircraftJSON(body)
					app.notify.EmitRarityNotifications(app.flightDash.RareSightings)
				}
			case <-summaryTicker.C:
				app.notify.PrintSummary(app.flightDash)
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
}
