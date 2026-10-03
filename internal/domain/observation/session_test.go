package observation

import (
	"testing"
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

func TestSpottingSessionSnapshotRestore(t *testing.T) {
	t.Parallel()

	observer := ref.NewCoordinates(53.55, 9.99)
	session := NewSpottingSession(observer)
	//nolint:exhaustruct_v5 // test state fixture
	state := State{
		Observer: observer,
		Sightings: map[string]AircraftSighting{
			"abc123": {
				LastFlightNo: "BAW123",
			},
		},
		SeenType:         map[string]int{"Boeing 737-800": 2},
		SeenOperator:     map[string]int{"British Airways": 1},
		SeenCountry:      map[string]int{"GB": 1},
		SightedTypes:     2,
		SightedOperators: 1,
		SightedCountries: 1,
		Fastest:          AircraftRecord{GroundSpeed: 500},
		Highest:          AircraftRecord{AltBaro: 35000.0},
	}

	if err := session.Restore(state); err != nil {
		t.Fatalf("Restore() unexpected error: %v", err)
	}

	snapshot := session.Snapshot()
	if snapshot.Observer != observer {
		t.Fatalf("Observer = %#v, want %#v", snapshot.Observer, observer)
	}
	if snapshot.SightedTypes != 2 {
		t.Fatalf("SightedTypes = %d, want 2", snapshot.SightedTypes)
	}
	if snapshot.Fastest.GroundSpeed != 500 {
		t.Fatalf("Fastest.GroundSpeed = %v, want 500", snapshot.Fastest.GroundSpeed)
	}
	if session.IsWarmup() != true {
		t.Fatal("expected new session to start warm")
	}
	session.SetWarmup(false)
	if session.IsWarmup() {
		t.Fatal("expected warmup to be disabled after SetWarmup(false)")
	}
}

func TestSpottingSessionRejectsObserverMismatch(t *testing.T) {
	t.Parallel()

	session := NewSpottingSession(ref.NewCoordinates(53.55, 9.99))
	//nolint:exhaustruct_v5 // test mismatch state
	state := State{Observer: ref.NewCoordinates(1.0, 2.0)}
	if err := session.Restore(state); err == nil {
		t.Fatal("expected observer mismatch error")
	}
}

func TestSpottingSessionObserveCreatesCurrentBatch(t *testing.T) {
	t.Parallel()

	observer := ref.NewCoordinates(53.55, 9.99)
	session := NewSpottingSession(observer)
	//nolint:exhaustruct_v5 // test fixture
	record := AircraftRecord{
		Hex:          "abc123",
		Flight:       "BAW123",
		IcaoType:     "B738",
		Lat:          53.6,
		Lon:          10.0,
		GroundSpeed:  400,
		AltBaro:      35000.0,
		Registration: "G-ABCD",
	}
	//nolint:exhaustruct_v5 // test fixture
	classifier := stubClassifier{makeByIcao: map[string]string{"B738": "Boeing 737-800"}}
	current := session.Observe([]AircraftRecord{record}, classifier, time.Now())
	if len(current) != 1 {
		t.Fatalf("expected 1 current sighting, got %d", len(current))
	}
	if session.Snapshot().SightedTypes != 1 {
		t.Fatalf("expected 1 sighted type, got %d", session.Snapshot().SightedTypes)
	}
}

func TestDetectRareSightingsBuildsDomainEvents(t *testing.T) {
	t.Parallel()

	sightings := []AircraftSighting{
		//nolint:exhaustruct_v5 // test event fixture
		{Rarities: NoRarity, LastFlightNo: "N123"},
		//nolint:exhaustruct_v5 // test event fixture
		{Rarities: RareType | RareOperator, LastFlightNo: "BAW123"},
	}

	events := DetectRareSightings(sightings)
	if len(events) != 1 {
		t.Fatalf("expected 1 rare event, got %d", len(events))
	}
	if events[0].EventName() != "RareSightingDetected" {
		t.Fatalf("event name = %q, want %q", events[0].EventName(), "RareSightingDetected")
	}
	if events[0].Rarity != (RareType | RareOperator) {
		t.Fatalf("event rarity = %v, want %v", events[0].Rarity, RareType|RareOperator)
	}
}
