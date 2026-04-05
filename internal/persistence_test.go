package internal

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "airspottr_state.json")

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origWd)
	}()
	if err := os.Chdir(findRepoRoot(t)); err != nil {
		t.Fatal(err)
	}

	dashboard, err := NewDashboard(1.0, 2.0, new(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	request, err := NewRequest(RequestOptions{Lat: 1.0, Lon: 2.0}, new(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	dashboard.isWarmup = false
	dashboard.SeenTypeCount["A"] = 1
	dashboard.SeenOperatorCount["OP"] = 2
	dashboard.SeenCountryCount["US"] = 3
	dashboard.totalTypeCount = 1
	dashboard.totalOperatorCount = 2
	dashboard.totalCountryCount = 3
	dashboard.CachedFlightRoutes["TEST123"] = GetDefaultFlightrouteRecord()

	sighting := &AircraftSighting{
		lastSeen:     now(),
		lastFlightNo: "TEST123",
		registration: "N12345",
		latitude:     1.23,
		longitude:    4.56,
		direction:    "north",
		distance:     789.0,
		typeShort:    "A320",
		typeDesc:     "Airbus A320",
		operator:     "TestAir",
		country:      "US",
		info:         "test info",
		flightroute:  GetDefaultFlightrouteRecord(),
	}
	dashboard.aircraftSightings["ABC123"] = sighting
	dashboard.RareSightings = []RareSighting{{Rarities: RareType, Sighting: sighting}}

	request.pendingCallsignsMu.Lock()
	request.pendingCallsigns = []string{"TEST123", "OTHER456"}
	request.pendingCallsignsMu.Unlock()

	if err := SaveState(statePath, dashboard, request); err != nil {
		t.Fatal(err)
	}

	dashboard2, err := NewDashboard(1.0, 2.0, new(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	request2, err := NewRequest(RequestOptions{Lat: 1.0, Lon: 2.0}, new(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	if err := LoadState(statePath, dashboard2, request2); err != nil {
		t.Fatal(err)
	}

	if got, want := len(request2.pendingCallsigns), 2; got != want {
		t.Fatalf("expected %d pending callsigns, got %d", want, got)
	}
	if got := request2.pendingCallsigns[0]; got != "TEST123" {
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
	if got := len(dashboard2.aircraftSightings); got != 1 {
		t.Fatalf("expected 1 aircraft sighting, got %d", got)
	}
	if got := dashboard2.aircraftSightings["ABC123"].lastFlightNo; got != "TEST123" {
		t.Fatalf("expected restored sighting flight TEST123, got %s", got)
	}
	if len(dashboard2.RareSightings) != 1 {
		t.Fatalf("expected 1 rare sighting, got %d", len(dashboard2.RareSightings))
	}
	if got := dashboard2.RareSightings[0].Sighting.lastFlightNo; got != "TEST123" {
		t.Fatalf("expected rare sighting to reference restored sighting, got %s", got)
	}
}

func now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func findRepoRoot(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		candidate := filepath.Join(wd, "data", "ICAOList.csv")
		if _, err := os.Stat(candidate); err == nil {
			return wd
		}
		if wd == filepath.Dir(wd) {
			break
		}
		wd = filepath.Dir(wd)
	}
	t.Fatal("could not locate repository root")
	return ""
}
