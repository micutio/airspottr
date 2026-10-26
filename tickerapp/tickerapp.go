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

	srv "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	noti "github.com/micutio/airspottr/internal/infrastructure/notify"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
)

type Dependencies struct {
	AppName            string
	Options            adsb.RequestOptions
	Logger             *slog.Logger
	AircraftRequest    *adsb.AircraftRequest
	FlightrouteRequest *adsb.FlightrouteRequest
	Dashboard          *srv.Dashboard
	Notify             *noti.BeeepNotifier
	SummaryNotify      *srv.Notify
	Done               chan bool
	Stdout             io.Writer
	Stderr             io.Writer
}

// TickerApp holds the state and dependencies for the ticker application.
type TickerApp struct {
	appName            string
	options            adsb.RequestOptions
	logger             *slog.Logger
	aircraftRequest    *adsb.AircraftRequest
	flightrouteRequest *adsb.FlightrouteRequest
	dashboard          *srv.Dashboard
	notify             *noti.BeeepNotifier
	summaryNotify      *srv.Notify
	done               chan bool
	wg                 sync.WaitGroup
}

func NewWithDependencies(deps Dependencies) *TickerApp {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.Done == nil {
		deps.Done = make(chan bool)
	}
	if deps.Dashboard != nil && deps.FlightrouteRequest != nil {
		deps.Dashboard.ConfigureRefresh(deps.Notify, deps.FlightrouteRequest.GetFlightroutes)
	}
	return &TickerApp{
		appName:            deps.AppName,
		options:            deps.Options,
		logger:             deps.Logger,
		aircraftRequest:    deps.AircraftRequest,
		flightrouteRequest: deps.FlightrouteRequest,
		dashboard:          deps.Dashboard,
		notify:             deps.Notify,
		summaryNotify:      deps.SummaryNotify,
		done:               deps.Done,
		wg:                 sync.WaitGroup{},
	}
}

func RunWithDependencies(deps Dependencies) {
	app := NewWithDependencies(deps)
	fmt.Printf("%s launching at Lat: %.3f, Lon: %.3f\n", app.appName, deps.Options.Lat, deps.Options.Lon)
	app.start()
	app.waitForShutdown()
}

// start begins the application's main event loop in a goroutine.
func (app *TickerApp) start() {
	// Set a timeout for the warmup period.
	time.AfterFunc(adsb.DashboardWarmup, func() {
		app.dashboard.FinishWarmupPeriod()
	})

	aircraftUpdateTicker := time.NewTicker(adsb.AircraftUpdateInterval)
	summaryTicker := time.NewTicker(adsb.SummaryInterval)

	app.wg.Go(func() {
		defer aircraftUpdateTicker.Stop()
		defer summaryTicker.Stop()

		for {
			select {
			case <-aircraftUpdateTicker.C:
				aircraftRecords := app.aircraftRequest.RequestAircraft()
				app.dashboard.RefreshBatch(aircraftRecords, noti.DefaultRarityNotifyToggles())
			case <-summaryTicker.C:
				app.summaryNotify.PrintSummary(app.dashboard)
			case <-app.done:
				app.logger.Info("Stopping HTTP GET request routine.")
				return
			}
		}
	})
}

// waitForShutdown blocks until an interrupt or terminate signal is received.
func (app *TickerApp) waitForShutdown() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	app.logger.Info("Shutdown signal received, stopping...")
	close(app.done)
	app.wg.Wait()
	state := app.dashboard.SaveState(app.flightrouteRequest.GetPendingCallsigns())
	if saveErr := pers.SaveState(pers.StateFilePath(), state); saveErr != nil {
		app.logger.Error("failed to save persistent state", slog.Any("error", saveErr))
	}
}
