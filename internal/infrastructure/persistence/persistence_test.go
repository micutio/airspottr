package persistence

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	srv "github.com/micutio/airspottr/internal/application/services"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
	"github.com/micutio/airspottr/internal/infrastructure/observation"
)

type typeRepoMock struct{}

func (typeRepoMock) GetAircraftType(_ string) (ref.IcaoAircraftSpec, bool) {
	return ref.IcaoAircraftSpec{}, false //nolint:exhaustruct_v5 // shortened for testing
}

type operatorRepoMock struct{}

func (operatorRepoMock) GetOperatorByIcao(_ string) (ref.IcaoOperator, bool) {
	return ref.IcaoOperator{}, false //nolint:exhaustruct_v5 // shortened for testing
}

func (operatorRepoMock) GetOperatorByMilCode(_ string) (string, bool) {
	return "", false
}

type countryRepoMock struct{}

func (countryRepoMock) GetCountryByHexCode(_ string) (string, error) {
	return "", nil
}

func (countryRepoMock) GetCountryByRegistration(_ string) (string, bool) {
	return "", false
}

func TestSaveAndLoadMultipleLocations(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "airspottr_state.json")

	singaporeKey := ref.LocationKey(1.3521, 103.8198)
	//nolint:exhaustruct_v5 // test fixture
	state1 := &srv.PersistentState{
		LocationStates: map[string]srv.DashboardState{
			//nolint:exhaustruct_v5 // test fixture
			singaporeKey: {
				Lat: 1.3521,
				Lon: 103.8198,
				SeenTypeCount: map[string]int{
					"A320": 5,
				},
				TotalTypeCount: 1,
			},
		},
	}
	if saveErr := SaveState(statePath, state1); saveErr != nil {
		t.Fatal(saveErr)
	}

	loaded1, load1Err := LoadState(statePath)
	if load1Err != nil {
		t.Fatal(load1Err)
	}

	sin, okSin := StateForLocation(loaded1, 1.3521, 103.8198)
	if !okSin || sin.SeenTypeCount["A320"] != 5 {
		t.Fatalf("expected Singapore location state to be found, got ok=%v, state=%+v", okSin, sin)
	}

	// Unsaved location should not be found
	_, okHam := StateForLocation(loaded1, 53.5511, 9.9937)
	if okHam {
		t.Fatal("expected Hamburg location state to not be found initially")
	}

	// Save Hamburg location to the same state file
	hamburgKey := ref.LocationKey(53.5511, 9.9937)
	//nolint:exhaustruct_v5 // test fixture
	state2 := &srv.PersistentState{
		LocationStates: map[string]srv.DashboardState{
			//nolint:exhaustruct_v5 // test fixture
			hamburgKey: {
				Lat: 53.5511,
				Lon: 9.9937,
				SeenTypeCount: map[string]int{
					"B738": 3,
				},
				TotalTypeCount: 1,
			},
		},
	}
	if save2Err := SaveState(statePath, state2); save2Err != nil {
		t.Fatal(save2Err)
	}

	loaded2, load2Err := LoadState(statePath)
	if load2Err != nil {
		t.Fatal(load2Err)
	}

	// Both locations must now be present
	sin2, okSin2 := StateForLocation(loaded2, 1.3521, 103.8198)
	if !okSin2 || sin2.SeenTypeCount["A320"] != 5 {
		t.Fatalf("expected SG state to remain intact after adding HAM, got ok=%v, state=%+v",
			okSin2,
			sin2)
	}
	ham2, okHam2 := StateForLocation(loaded2, 53.5511, 9.9937)
	if !okHam2 || ham2.SeenTypeCount["B738"] != 3 {
		t.Fatalf("expected Hamburg location state to be found, got ok=%v, state=%+v", okHam2, ham2)
	}

	// Unsaved location must still return false
	_, okNY := StateForLocation(loaded2, 40.7128, -74.0060)
	if okNY {
		t.Fatal("expected New York location state to not be found")
	}

	// Ensure saved JSON on disk does not contain legacy "dashboard" field
	diskBytes, diskByteErr := os.ReadFile(statePath)
	if diskByteErr != nil {
		t.Fatal(diskByteErr)
	}
	if strings.Contains(string(diskBytes), `"dashboard"`) {
		t.Fatalf("expected 'dashboard' key not to be in JSON, got: %s", string(diskBytes))
	}
}

func TestSaveAndLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "airspottr_state.json")

	origWd, wdErr := os.Getwd()
	if wdErr != nil {
		t.Fatal(wdErr)
	}
	defer func() {
		t.Chdir(origWd)
	}()
	t.Chdir(findRepoRoot(t))

	sightingRepo := observation.NewSightingRepo()
	sighting := obs.AircraftSighting{
		Rarities:     obs.RareType,
		LastSeen:     now(),
		LastFlightNo: "TEST123",
		Registration: "N12345",
		Latitude:     1.23,
		Longitude:    4.56,
		Direction:    "north",
		Distance:     789.0,
		TypeShort:    "A320",
		TypeDesc:     "Airbus A320",
		Operator:     "TestAir",
		Country:      "US",
		Info:         "test info",
		Flightroute:  *ref.GetDefaultFlightrouteRecord(),
		LastRecord:   obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
	}
	sightingRepo.UpdateSighting("ABC123", sighting)

	dashboard := srv.NewDashboard(
		1.0,
		2.0,
		sightingRepo,
		typeRepoMock{},
		operatorRepoMock{},
		countryRepoMock{},
		io.Discard)

	request, reqErr := adsb.NewFlightrouteRequest(io.Discard)
	if reqErr != nil {
		t.Fatal(reqErr)
	}

	dashboard.IsWarmup = false
	dashboard.SeenTypeCount["A"] = 1
	dashboard.SeenOperatorCount["OP"] = 2
	dashboard.SeenCountryCount["US"] = 3
	dashboard.SightedTypesCount = 1
	dashboard.SightedOperatorsCount = 2
	dashboard.SightedCountriesCount = 3

	request.RestorePendingCallsigns([]string{"TEST123", "OTHER456"})

	stateToSave := dashboard.SaveState(request.GetPendingCallsigns())
	if saveErr := SaveState(statePath, stateToSave); saveErr != nil {
		t.Fatal(saveErr)
	}

	sightingRepo2 := observation.NewSightingRepo()
	dashboard2 := srv.NewDashboard(
		1.0,
		2.0,
		sightingRepo2,
		typeRepoMock{},
		operatorRepoMock{},
		countryRepoMock{},
		io.Discard)

	appState, appStateErr := LoadState(statePath)
	if appStateErr != nil {
		t.Fatal(appStateErr)
	}

	savedDashboard, ok := StateForLocation(appState, 1.0, 2.0)
	if !ok {
		t.Fatal("expected location state for (1.0, 2.0) to be found")
	}
	if loadDashboardErr := dashboard2.RestoreState(&savedDashboard); loadDashboardErr != nil {
		t.Fatal(loadDashboardErr)
	}

	diskBytes, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(diskBytes), `"dashboard"`) {
		t.Fatalf("expected 'dashboard' key not to be written by SaveState, got: %s", string(diskBytes))
	}

	if got := dashboard2.SeenTypeCount["A"]; got != 1 {
		t.Fatalf("expected SeenTypeCount A=1, got %d", got)
	}
	if got := dashboard2.SeenOperatorCount["OP"]; got != 2 {
		t.Fatalf("expected SeenOperatorCount OP=2, got %d", got)
	}
	if got := dashboard2.SeenCountryCount["US"]; got != 3 {
		t.Fatalf("expected SeenCountryCount US=3, got %d", got)
	}
	allSightings := sightingRepo2.GetAllSightings()
	if got := len(allSightings); got != 1 {
		t.Fatalf("expected 1 aircraft sighting, got %d", got)
	}
	if got := allSightings["ABC123"].LastFlightNo; got != "TEST123" {
		t.Fatalf("expected restored sighting flight TEST123, got %s", got)
	}
}

func now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func findRepoRoot(t *testing.T) string {
	workingDir, wdErr := os.Getwd()
	if wdErr != nil {
		t.Fatal(wdErr)
	}
	for range 10 {
		candidate := filepath.Join(workingDir, "data", "ICAOList.csv")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return workingDir
		}
		if workingDir == filepath.Dir(workingDir) {
			break
		}
		workingDir = filepath.Dir(workingDir)
	}
	t.Fatal("could not locate repository root")
	return ""
}
