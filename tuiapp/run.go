package tuiapp

import (
	"io"
	"log" //nolint:depguard // TODO: Change later.
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/micutio/airspottr/internal"
	adsb "github.com/micutio/airspottr/internal/infrastructure/adsb"
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

	notify := internal.NewNotify(appName, new(io.Discard))

	request, dashboard, err := setupRequestAndDashboard(requestOptions, errLogFile)
	if err != nil {
		log.Printf("failed to set up dashboard and request: %v", err)
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
		request:           request,
		dashboard:         dashboard,
		notify:            notify,
		options:           requestOptions,
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
	if saveErr := pers.SaveState(pers.StateFilePath(), dashboard, request); saveErr != nil {
		log.Printf("failed to save persistent state: %v", saveErr)
	}
}
