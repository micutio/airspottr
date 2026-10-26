package tuiapp

import (
	"io"
	"log" //nolint:depguard // TODO: Change later.
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	srv "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	noti "github.com/micutio/airspottr/internal/infrastructure/notify"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
)

type Dependencies struct {
	AppName             string
	AircraftSource      srv.AircraftDataSource
	FlightRouteSource   srv.FlightRouteDataSource
	Dashboard           *srv.Dashboard
	NotificationService srv.NotificationService
	RequestOptions      adsb.RequestOptions
	ErrorLog            io.Writer
}

func RunWithDependencies(deps Dependencies) {
	if deps.ErrorLog == nil {
		deps.ErrorLog = io.Discard
	}
	log.SetOutput(deps.ErrorLog)
	if deps.Dashboard == nil {
		log.Printf("dashboard is required for TUI startup")
		return
	}
	if deps.AircraftSource == nil || deps.FlightRouteSource == nil {
		log.Printf("aircraft and route sources are required for TUI startup")
		return
	}
	notifyInstance := noti.NewBeeepNotifier(deps.AppName, io.Discard)
	if concreteNotify, ok := deps.NotificationService.(*noti.BeeepNotifier); ok {
		notifyInstance = concreteNotify
	}
	deps.NotificationService = notifyInstance

	theme := getDefaultTheme()
	tables := initTables(theme)
	deps.Dashboard.ConfigureRefresh(
		deps.NotificationService,
		deps.FlightRouteSource.GetFlightroutes,
	)
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout))
	appModel := &model{
		width:             0,
		height:            0,
		baseStyle:         lipgloss.NewStyle(),
		viewStyle:         lipgloss.NewStyle(),
		tableStyle:        tables.style,
		theme:             theme,
		tables:            tables.tables,
		selectedRarityIdx: 0,
		aircraftSortCol:   1,
		aircraftSortDesc:  false,
		raritySortCol:     [3]int{0, 0, 0},
		raritySortDesc:    [3]bool{false, false, false},
		uiState:           mainPage,
		startTime:         time.Now(),
		lastUpdate:        time.Unix(0, 0),
		aircraftRepo:      deps.AircraftSource,
		flightrouteRepo:   deps.FlightRouteSource,
		dashboard:         deps.Dashboard,
		notify:            notifyInstance,
		options:           deps.RequestOptions,
		inputFocus:        focusTable,
		notifyStripIdx:    notifyType,
		notifyOnType:      true,
		notifyOnOp:        true,
		notifyOnCountry:   true,
	}

	p := tea.NewProgram(appModel, tea.WithAltScreen())
	if _, progErr := p.Run(); progErr != nil {
		log.Printf("error running program: %v", progErr)
	}
	state := deps.Dashboard.SaveState(deps.FlightRouteSource.GetPendingCallsigns())
	if saveErr := pers.SaveState(pers.StateFilePath(), state); saveErr != nil {
		log.Printf("failed to save persistent state: %v", saveErr)
	}
}
