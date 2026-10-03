package tuiapp

import (
	"io"
	"log" //nolint:depguard // TODO: Change later.
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	appint "github.com/micutio/airspottr/internal/application/interfaces"
	srv "github.com/micutio/airspottr/internal/application/services"
	repo "github.com/micutio/airspottr/internal/domain/repositories"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	"github.com/micutio/airspottr/internal/infrastructure/data"
	noti "github.com/micutio/airspottr/internal/infrastructure/notify"
	"github.com/micutio/airspottr/internal/infrastructure/observation"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
)

type Dependencies struct {
	AppName             string
	AircraftSource      appint.AircraftDataSource
	FlightRouteSource   appint.FlightRouteDataSource
	TypeRepo            repo.AircraftTypeRepo
	Dashboard           *srv.Dashboard
	NotificationService appint.NotificationService
	RequestOptions      adsb.RequestOptions
	ErrorLog            io.Writer
}

// Run starts the TUI. It exits the process if dashboard or request setup fails.
func Run(appName string, requestOptions adsb.RequestOptions) {
	errLogFile, err := setupLogger()
	if err != nil {
		log.Fatalf("failed to set up logging: %v", err)
	}
	defer func() {
		if closeErr := errLogFile.Close(); closeErr != nil {
			log.Printf("error closing log file: %v", closeErr)
		}
	}()

	appState, appStateErr := pers.LoadState(pers.StateFilePath())
	if appStateErr != nil {
		log.Printf("failed to load app state: %v", appStateErr)
	}

	typeRepo, typeRepoErr := data.NewAircraftTypeRepo()
	if typeRepoErr != nil {
		log.Printf("failed to create aircraft type repo: %v", typeRepoErr)
		return
	}
	operatorRepo, operatorRepoErr := data.NewOperatorRepo()
	if operatorRepoErr != nil {
		log.Printf("failed to set up operator repository: %v", operatorRepoErr)
		return
	}
	countryRepo, countryRepoErr := data.NewCountryRepo()
	if countryRepoErr != nil {
		log.Printf("unable to create country repository: %v", countryRepoErr)
		return
	}

	dashboard, err := setupDashboard(
		appState,
		requestOptions,
		observation.NewSightingRepo(),
		typeRepo,
		operatorRepo,
		countryRepo,
		errLogFile,
	)
	if err != nil {
		log.Printf("failed to set up dashboard and request: %v", err)
		return
	}
	dashboard.FinishWarmupPeriod()
	flightrouteReq, flightrouteErr := setupFlightrouteRequest(errLogFile)
	if flightrouteErr != nil {
		log.Printf("failed to setup flightroute request: %v", flightrouteErr)
		return
	}
	aircraftReq, aircraftErr := setupAircraftRequest(requestOptions, errLogFile)
	if aircraftErr != nil {
		log.Printf("failed to setup aircraft request: %v", aircraftErr)
		return
	}
	notify := noti.NewBeeepNotifier(appName, io.Discard)
	RunWithDependencies(Dependencies{
		AppName:             appName,
		AircraftSource:      aircraftReq,
		FlightRouteSource:   flightrouteReq,
		TypeRepo:            typeRepo,
		Dashboard:           dashboard,
		NotificationService: notify,
		RequestOptions:      requestOptions,
		ErrorLog:            errLogFile,
	})
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
	if deps.TypeRepo == nil {
		typeRepo, typeRepoErr := data.NewAircraftTypeRepo()
		if typeRepoErr != nil {
			log.Printf("failed to create aircraft type repo for TUI: %v", typeRepoErr)
			return
		}
		deps.TypeRepo = typeRepo
	}
	notifyInstance := noti.NewBeeepNotifier(deps.AppName, io.Discard)
	if concreteNotify, ok := deps.NotificationService.(*noti.BeeepNotifier); ok {
		notifyInstance = concreteNotify
	}
	deps.NotificationService = notifyInstance

	theme := getDefaultTheme()
	tables := initTables(theme)
	refreshUseCase := srv.NewRefreshUseCase(
		deps.Dashboard,
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
		typeRepo:          deps.TypeRepo,
		dashboard:         deps.Dashboard,
		refreshUseCase:    refreshUseCase,
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
