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
	"syscall"
	"time"

	"github.com/micutio/airspottr/internal"
)

func Run(appName string, options internal.RequestOptions) {
	fmt.Printf("%s launching at Lat: %.3f, Lon: %.3f\n", appName, options.Lat, options.Lon)

	// TODO: Replace slog logger!
	logger := slog.Default()

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	notify := internal.NewNotify(appName, &stdout)

	flightDash, dashboardErr := internal.NewDashboard(options.Lat, options.Lon, &stderr)
	if dashboardErr != nil {
		logger.Error("unable to create dashboard, exiting", slog.Any("dashboard error", dashboardErr))
		os.Exit(1)
	}

	// Set a timeout for the warmup period. After that point in time we will show rare aircraft immediately
	time.AfterFunc(internal.DashboardWarmup, func() {
		flightDash.FinishWarmupPeriod()
	})

	// Create an aircraft update ticker that fires in a given interval
	aircraftUpdateTicker := time.NewTicker(internal.AircraftUpdateInterval)
	defer aircraftUpdateTicker.Stop()

	// Create a summary ticker that fires in a given interval
	summaryTicker := time.NewTicker(internal.SummaryInterval)
	defer summaryTicker.Stop()

	// Use a channel to gracefully stop the program if needed.
	done := make(chan bool)

	// Start a goroutine to perform the requests
	go func() {
		for {
			select {
			case <-aircraftUpdateTicker.C:
				if body, err := internal.RequestAndProcessCivAircraft(options); err != nil {
					logger.Error("main: ", slog.Any("error", err))
				} else {
					flightDash.ProcessCivAircraftJSON(body)
					notify.EmitRarityNotifications(flightDash.RareSightings)
				}
			case <-summaryTicker.C:
				notify.PrintSummary(flightDash)
			case <-done:
				// This case allows for graceful shutdown (not used in this example but good practice)
				slog.Info("Stopping HTTP GET request routine.")

				return
			}
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	slog.Info("Shutdown signal received, stopping...")
	close(done)
}
