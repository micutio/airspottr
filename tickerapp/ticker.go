// Package tickerapp launches the ticker application which writes out all updates to stdout and
// can be piped into other programs and processed further.
// This is in contrast to the TUI app, which works more like htop.
package tickerapp

import (
	"log/slog"
	"os"
	"time"

	"github.com/micutio/airspottr/internal"
)

func Run() {
	// Initialize logging, notifications and dashboard

	logger := slog.Default()

	logParams := internal.LogParams{
		ConsoleOut: os.Stdout,
		ErrorOut:   os.Stderr,
	}

	flightDash, dashboardErr := internal.NewDashboard(logParams)
	if dashboardErr != nil {
		logger.Error("unable to create dashboard, exiting", slog.Any("dashboard error", dashboardErr))
		os.Exit(1)
	}

	// Set a timeout for the warmup period. After that point in time we will show rare aircraft immediately
	time.AfterFunc(internal.DashboardWarmup, func() {
		flightDash.FinishWarmupPeriod()
	})

	// Create a aircraftUpdateTicker that fires every 30 seconds
	aircraftUpdateTicker := time.NewTicker(internal.AircraftUpdateInterval)
	defer aircraftUpdateTicker.Stop()

	// aircraft and military aircraft updates should not coincide to avoid exceeding the api limit.
	// Hence, stagger them by 15 seconds.
	milAircraftUpdateTicker := time.NewTicker(internal.MilAircraftUpdateInterval)
	defer milAircraftUpdateTicker.Stop()

	time.AfterFunc(internal.MilAircraftUpdateDelay, func() {
		milAircraftUpdateTicker.Reset(internal.MilAircraftUpdateInterval)
	})

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
				if body, err := internal.RequestAndProcessCivAircraft(); err != nil {
					logger.Error("main: ", slog.Any("error", err))
				} else {
					flightDash.ProcessCivAircraftJSON(body)
				}
			case <-milAircraftUpdateTicker.C:
				if body, err := internal.RequestAndProcessMilAircraft(); err != nil {
					logger.Error("main: %w", slog.Any("error", err))
				} else {
					flightDash.ProcessMilAircraftJSON(body)
				}
			case <-summaryTicker.C:
				flightDash.PrintSummary()
			case <-done:
				// This case allows for graceful shutdown (not used in this example but good practice)
				logger.Info("Stopping HTTP GET request routine.")

				return
			}
		}
	}()

	// Run once in the beginning.
	if body, err := internal.RequestAndProcessMilAircraft(); err != nil {
		logger.Error("main: ", slog.Any("error", err))
	} else {
		flightDash.ProcessMilAircraftJSON(body)
	}

	// Keep the main goroutine alive indefinitely, or until an interrupt signal is received
	// In a real application, you might have other logic here or use a wait group.
	select {} // Block indefinitely
}
