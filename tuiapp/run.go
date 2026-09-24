package tuiapp

import (
	"io"
	"log" //nolint:depguard // TODO: Change later.
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	noti "github.com/micutio/airspottr/internal/application/services"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	"github.com/micutio/airspottr/internal/infrastructure/data"
	pers "github.com/micutio/airspottr/internal/infrastructure/persistence"
)

// Run starts the TUI. It exits the process if dashboard or request setup fails.
func Run(appName string, requestOptions adsb.RequestOptions) {
	// Map ANSI palette colors to this terminal (theme-aware selection/highlight).
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout))

	errLogFile, err := setupLogger()
	if err != nil {
		log.Fatalf("failed to set up logging: %v", err)
	}
	defer func() {
		if closeErr := errLogFile.Close(); closeErr != nil {
			log.Printf("error closing log file: %v", closeErr)
		}
	}()

	notify := noti.NewNotify(appName, new(io.Discard))

	aircraftReq, err := setupAircraftRequest(requestOptions, errLogFile)
	if err != nil {
		log.Printf("failed to setup aircraft request: %v", err)
	}

	appState, err := pers.LoadState(pers.StateFilePath())
	if err != nil {
		log.Printf("failed to load app state: %v", err)
	}

	flightrouteReq, err := setupFlightrouteRequest(errLogFile)
	if err != nil {
		log.Printf("failed to setup flightroute request: %v", err)
	}

	dashboard, err := setupDashboard(requestOptions, appState, errLogFile)
	if err != nil {
		log.Printf("failed to set up dashboard and request: %v", err)
		return
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

	dashboard.FinishWarmupPeriod()

	theme := getDefaultTheme()
	tables := initTables(theme)

	appModel := &model{
		width:             0,
		height:            0,
		baseStyle:         lipgloss.NewStyle(),
		viewStyle:         lipgloss.NewStyle(),
		tableStyle:        tables.style,
		theme:             theme,
		tables:            tables.tables,
		selectedRarityIdx: 0,
		aircraftSortCol:   1, // FNO (matches previous default flight sort)
		aircraftSortDesc:  false,
		raritySortCol:     [3]int{0, 0, 0},
		raritySortDesc:    [3]bool{false, false, false},
		uiState:           mainPage,
		startTime:         time.Now(),
		lastUpdate:        time.Unix(0, 0),
		aircraftRepo:      aircraftReq,
		flightrouteRepo:   flightrouteReq,
		typeRepo:          typeRepo,
		operatorRepo:      operatorRepo,
		countryRepo:       countryRepo,
		dashboard:         dashboard,
		notify:            notify,
		options:           requestOptions,
		currentSightings:  nil,
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
	if saveErr := pers.SaveState(pers.StateFilePath(), dashboard, flightrouteReq); saveErr != nil {
		log.Printf("failed to save persistent state: %v", saveErr)
	}
}
