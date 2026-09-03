package internal

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	internal "github.com/micutio/airspottr/internal/application"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	"github.com/micutio/airspottr/internal/infrastructure/adsb"
)

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

	dashboard, dashErr := internal.NewDashboard(1.0, 2.0, new(io.Discard))
	if dashErr != nil {
		t.Fatal(dashErr)
	}

	request, reqErr := adsb.NewRequest(adsb.RequestOptions{Lat: 1.0, Lon: 2.0}, new(io.Discard))
	if reqErr != nil {
		t.Fatal(reqErr)
	}

	dashboard.IsWarmup = false
	dashboard.SeenTypeCount["A"] = 1
	dashboard.SeenOperatorCount["OP"] = 2
	dashboard.SeenCountryCount["US"] = 3
	dashboard.TotalTypeCount = 1
	dashboard.TotalOperatorCount = 2
	dashboard.TotalCountryCount = 3
	dashboard.CachedFlightRoutes["TEST123"] = ref.GetDefaultFlightrouteRecord()

	sighting := &obs.AircraftSighting{
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
		Flightroute:  ref.GetDefaultFlightrouteRecord(),
	}
	dashboard.AircraftSightings["ABC123"] = sighting
	dashboard.RareSightings = []obs.RareSighting{{Rarities: obs.RareType, Sighting: sighting}}

	request.PendingCallsignsMu.Lock()
	request.PendingCallsigns = []string{"TEST123", "OTHER456"}
	request.PendingCallsignsMu.Unlock()

	if saveErr := SaveState(statePath, dashboard, request); saveErr != nil {
		t.Fatal(saveErr)
	}

	dashboard2, dashErr := internal.NewDashboard(1.0, 2.0, new(io.Discard))
	if dashErr != nil {
		t.Fatal(dashErr)
	}

	request2, requestErr := adsb.NewRequest(adsb.RequestOptions{Lat: 1.0, Lon: 2.0}, new(io.Discard))
	if requestErr != nil {
		t.Fatal(requestErr)
	}

	if err := LoadState(statePath, dashboard2, request2); err != nil {
		t.Fatal(err)
	}

	if got, want := len(request2.PendingCallsigns), 2; got != want {
		t.Fatalf("expected %d pending callsigns, got %d", want, got)
	}
	if got := request2.PendingCallsigns[0]; got != "TEST123" {
		t.Fatalf("expected first pending callsign TEST123, got %s", got)
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
	if got := len(dashboard2.AircraftSightings); got != 1 {
		t.Fatalf("expected 1 aircraft sighting, got %d", got)
	}
	if got := dashboard2.AircraftSightings["ABC123"].LastFlightNo; got != "TEST123" {
		t.Fatalf("expected restored sighting flight TEST123, got %s", got)
	}
	if len(dashboard2.RareSightings) != 1 {
		t.Fatalf("expected 1 rare sighting, got %d", len(dashboard2.RareSightings))
	}
	if got := dashboard2.RareSightings[0].Sighting.LastFlightNo; got != "TEST123" {
		t.Fatalf("expected rare sighting to reference restored sighting, got %s", got)
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
