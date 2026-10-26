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
	if deps.NotificationService == nil {
		deps.NotificationService = noti.NewBeeepNotifier(deps.AppName, io.Discard)
	}

	theme := getDefaultTheme()
	tableStyle := initTableStyle(theme)
	deps.Dashboard.ConfigureRefresh(
		deps.NotificationService,
		nil, // Flight routes are fetched asynchronously via tea.Cmd in Update.
	)
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout))
	appModel := &model{
		width:             0,
		height:            0,
		baseStyle:         lipgloss.NewStyle(),
		viewStyle:         lipgloss.NewStyle(),
		tableStyle:        tableStyle,
		theme:             theme,
		tables:            newTuiTables(tableStyle),
		selectedRarityIdx: rarityByType,
		uiState:           mainPage,
		startTime:         time.Now(),
		lastUpdate:        time.Unix(0, 0),
		aircraftRepo:      deps.AircraftSource,
		flightrouteRepo:   deps.FlightRouteSource,
		dashboard:         deps.Dashboard,
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
