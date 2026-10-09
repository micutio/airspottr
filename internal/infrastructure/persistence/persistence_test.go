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

func TestLoadStateMigratesLegacyStateToLocationHistory(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "airspottr_state.json")
	//nolint:exhaustruct_v5 // test migration fixture
	legacyState := &srv.PersistentState{
		//nolint:exhaustruct_v5 // test migration fixture
		DashboardState: &srv.DashboardState{
			Lat: 1.3521,
			Lon: 103.8198,
			SeenTypeCount: map[string]int{
				"A": 4,
			},
		},
	}
	if err := SaveState(statePath, legacyState); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	key := ref.LocationKey(1.3521, 103.8198)
	if _, ok := loaded.InternalState.LocationStates[key]; !ok {
		t.Fatalf("expected location history for %s to be present", key)
	}
	selected, ok := StateForLocation(loaded, 1.3521, 103.8198)
	if !ok {
		t.Fatalf("expected state for location %s to be found", key)
	}
	if selected.Lat != 1.3521 || selected.Lon != 103.8198 {
		t.Fatalf("StateForLocation() = (%f,%f), want (1.3521,103.8198)", selected.Lat, selected.Lon)
	}

	// Another location should not find state or return legacy state.
	_, okDiff := StateForLocation(loaded, 53.5511, 9.9937)
	if okDiff {
		t.Fatal("expected no state for different location")
	}
}

func TestLoadStateMigratesRawLegacyJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "airspottr_state.json")

	// Raw legacy JSON without location_states key
	rawLegacyJSON := []byte(`{
  "dashboard": {
    "is_warmup": false,
    "lat": 1.3521,
    "lon": 103.8198,
    "seen_type_count": {
      "A320": 5
    },
    "total_type_count": 1
  },
  "request": {
    "pending_callsigns": ["SIA123"]
  }
}`)

	if err := os.WriteFile(statePath, rawLegacyJSON, 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	key := ref.LocationKey(1.3521, 103.8198)
	if _, exists := loaded.InternalState.LocationStates[key]; !exists {
		t.Fatalf("expected legacy state to be migrated into LocationStates under key %s", key)
	}
	if loaded.InternalState.DashboardState != nil {
		t.Fatalf("expected DashboardState to be nil in memory after migration")
	}

	// Check that the file on disk was migrated to the new format (dashboard key removed).
	diskBytes, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(diskBytes), `"dashboard"`) {
		t.Fatalf("expected 'dashboard' key to be removed from migrated JSON file on disk, got: %s", string(diskBytes))
	}
	if !strings.Contains(string(diskBytes), `"location_states"`) {
		t.Fatalf("expected 'location_states' to be present in migrated JSON file on disk, got: %s", string(diskBytes))
	}

	state, ok := StateForLocation(loaded, 1.3521, 103.8198)
	if !ok {
		t.Fatal("expected StateForLocation to return true for legacy location")
	}
	if state.TotalTypeCount != 1 || state.SeenTypeCount["A320"] != 5 {
		t.Fatalf("unexpected state content: %+v", state)
	}

	// Querying another location must return false, not legacy state
	_, okOther := StateForLocation(loaded, 53.5511, 9.9937)
	if okOther {
		t.Fatal("expected StateForLocation to return false for non-matching location")
	}

	// Saving a new location to this file must preserve the migrated legacy location
	hamburgState := &srv.PersistentState{
		DashboardState: &srv.DashboardState{
			Lat: 53.5511,
			Lon: 9.9937,
			SeenTypeCount: map[string]int{
				"B738": 3,
			},
			TotalTypeCount: 1,
		},
	}
	if err := SaveState(statePath, hamburgState); err != nil {
		t.Fatal(err)
	}

	reloaded, err := LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	// Both locations should now be present
	singapore, okSin := StateForLocation(reloaded, 1.3521, 103.8198)
	if !okSin || singapore.SeenTypeCount["A320"] != 5 {
		t.Fatalf("expected Singapore state to be preserved, got ok=%v, state=%+v", okSin, singapore)
	}

	hamburg, okHam := StateForLocation(reloaded, 53.5511, 9.9937)
	if !okHam || hamburg.SeenTypeCount["B738"] != 3 {
		t.Fatalf("expected Hamburg state to be saved, got ok=%v, state=%+v", okHam, hamburg)
	}

	// Check that the re-saved file also does NOT contain the legacy dashboard key
	diskBytesAfterSave, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(diskBytesAfterSave), `"dashboard"`) {
		t.Fatalf("expected 'dashboard' key to NOT be written by SaveState, got: %s", string(diskBytesAfterSave))
	}
}

func TestLoadStateCleansUpDuplicateLegacyAndNewFormat(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "airspottr_state.json")

	// File with duplicate data in both old and new formats
	duplicateJSON := []byte(`{
  "dashboard": {
    "is_warmup": false,
    "lat": 1.3521,
    "lon": 103.8198,
    "seen_type_count": {
      "A320": 5
    },
    "total_type_count": 1
  },
  "request": {
    "pending_callsigns": null
  },
  "location_states": {
    "1.3521,103.8198": {
      "is_warmup": false,
      "lat": 1.3521,
      "lon": 103.8198,
      "seen_type_count": {
        "A320": 5
      },
      "total_type_count": 1
    },
    "53.5511,9.9937": {
      "is_warmup": false,
      "lat": 53.5511,
      "lon": 9.9937,
      "seen_type_count": {
        "B738": 3
      },
      "total_type_count": 1
    }
  }
}`)

	if err := os.WriteFile(statePath, duplicateJSON, 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.InternalState.DashboardState != nil {
		t.Fatal("expected DashboardState to be nil after migration")
	}

	// Check file on disk: "dashboard" must be removed
	diskBytes, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(diskBytes), `"dashboard"`) {
		t.Fatalf("expected 'dashboard' key to be stripped from file on disk, got: %s", string(diskBytes))
	}

	// Both locations must still be available
	sin, okSin := StateForLocation(loaded, 1.3521, 103.8198)
	if !okSin || sin.SeenTypeCount["A320"] != 5 {
		t.Fatalf("expected Singapore location state to remain intact, got: %+v", sin)
	}
	ham, okHam := StateForLocation(loaded, 53.5511, 9.9937)
	if !okHam || ham.SeenTypeCount["B738"] != 3 {
		t.Fatalf("expected Hamburg location state to remain intact, got: %+v", ham)
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
