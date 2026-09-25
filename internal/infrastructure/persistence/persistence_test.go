package persistence

import (
	"io"
	"os"
	"path/filepath"
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
		new(io.Discard))

	request, reqErr := adsb.NewFlightrouteRequest(new(io.Discard))
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
		new(io.Discard))

	appState, appStateErr := LoadState(statePath)
	if appStateErr != nil {
		t.Fatal(appStateErr)
	}

	if loadDashboardErr := dashboard2.RestoreState(&appState.InternalState.DashboardState); loadDashboardErr != nil {
		t.Fatal(loadDashboardErr)
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
