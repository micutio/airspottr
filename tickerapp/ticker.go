// Package tickerapp launches the ticker application which writes out all updates to stdout and
// can be piped into other programs and processed further.
// This is in contrast to the TUI app, which works more like htop.
package tickerapp

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/micutio/airspottr/internal"
	"github.com/micutio/airspottr/internal/dash"
)

func Run(options internal.RequestOptions) {
	fmt.Printf("airspottr launching at Lat: %.3f, Lon: %.3f\n", options.Lat, options.Lon)

	// TODO: Replace slog logger!
	logger := slog.Default()

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	notify := internal.NewNotify(&stdout)

	flightDash, dashboardErr := dash.NewDashboard(options.Lon, options.Lon, &stderr)
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
	// (Though not strictly necessary for an infinite loop)
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
				}
			case <-summaryTicker.C:
				notify.PrintSummary(flightDash)
			case <-done:
				// This case allows for graceful shutdown (not used in this example but good practice)
				logger.Info("Stopping HTTP GET request routine.")

				return
			}
		}
	}()

	// Keep the main goroutine alive indefinitely, or until an interrupt signal is received
	// In a real application, you might have other logic here or use a wait group.
	select {} // Block indefinitely
}
