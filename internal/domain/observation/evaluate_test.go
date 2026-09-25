//nolint:exhaustruct_v5
package observation

import (
	"testing"
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

type stubClassifier struct {
	makeByIcao   map[string]string
	opByIcao     map[string][2]string // company, country
	opByMil      map[string]string
	countryByHex map[string]string
	countryByReg map[string]string
}

func (s stubClassifier) AircraftMake(icaoType string) (string, bool) {
	makeName, ok := s.makeByIcao[icaoType]
	return makeName, ok
}

func (s stubClassifier) OperatorByIcao(code string) (string, string, bool) {
	pair, ok := s.opByIcao[code]
	return pair[0], pair[1], ok
}

func (s stubClassifier) OperatorByMil(code string) (string, bool) {
	name, ok := s.opByMil[code]
	return name, ok
}

func (s stubClassifier) CountryByHex(hex string) (string, bool) {
	country, ok := s.countryByHex[hex]
	return country, ok
}

func (s stubClassifier) CountryByRegistration(reg string) (string, bool) {
	country, ok := s.countryByReg[reg]
	return country, ok
}

func emptyState() State {
	return State{
		Observer:     ref.NewCoordinates(53.55, 9.99),
		Sightings:    map[string]AircraftSighting{},
		SeenType:     map[string]int{},
		SeenOperator: map[string]int{},
		SeenCountry:  map[string]int{},
	}
}

func baseRecord() AircraftRecord {
	return AircraftRecord{ //nolint:exhaustruct_v5 // test fixture
		Hex:          "abc123",
		Flight:       "BAW123",
		IcaoType:     "B738",
		Lat:          53.6,
		Lon:          10.0,
		GroundSpeed:  400,
		AltBaro:      35000.0,
		Registration: "G-ABCD",
	}
}

func TestIdentifyFlight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		previous, current string
		isNew             bool
		want              FlightIdentity
	}{
		{
			name:     "new hex with known flight is identified and new",
			previous: FlightUnknown,
			current:  "BAW123",
			isNew:    true,
			want:     FlightIdentity{Identified: true, Updated: false, NewFlight: true},
		},
		{
			name:     "same known flight is not new",
			previous: "BAW123",
			current:  "BAW123",
			isNew:    false,
			want:     FlightIdentity{Identified: false, Updated: false, NewFlight: false},
		},
		{
			name:     "known flight change is updated and new",
			previous: "BAW123",
			current:  "DLH456",
			isNew:    false,
			want:     FlightIdentity{Identified: false, Updated: true, NewFlight: true},
		},
		{
			name:     "unknown to known is identified but not a statistical new flight",
			previous: FlightUnknown,
			current:  "BAW123",
			isNew:    false,
			want:     FlightIdentity{Identified: true, Updated: false, NewFlight: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IdentifyFlight(tc.previous, tc.current, tc.isNew)
			if got != tc.want {
				t.Fatalf("IdentifyFlight() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestEvaluateBatchNewAircraftUnknownUntilIdentified(t *testing.T) {
	t.Parallel()

	record := baseRecord()
	record.Flight = ""
	state, current := EvaluateBatch(emptyState(), []AircraftRecord{record}, stubClassifier{}, time.Now())
	if len(current) != 1 {
		t.Fatalf("expected 1 current sighting, got %d", len(current))
	}
	if current[0].LastFlightNo != FlightUnknown {
		t.Fatalf("expected unknown flight, got %q", current[0].LastFlightNo)
	}
	if _, exists := state.Sightings[record.Hex]; !exists {
		t.Fatal("expected sighting stored by hex")
	}
}

func TestEvaluateBatchSameFlightDoesNotIncrementAgain(t *testing.T) {
	t.Parallel()

	classifier := stubClassifier{
		makeByIcao: map[string]string{"B738": "Boeing 737-800"},
	}
	record := baseRecord()
	state, _ := EvaluateBatch(emptyState(), []AircraftRecord{record}, classifier, time.Now())
	if state.SightedTypes != 1 {
		t.Fatalf("first pass SightedTypes = %d, want 1", state.SightedTypes)
	}
	state, _ = EvaluateBatch(state, []AircraftRecord{record}, classifier, time.Now())
	if state.SightedTypes != 1 {
		t.Fatalf("second pass SightedTypes = %d, want 1", state.SightedTypes)
	}
	if state.SeenType["Boeing 737-800"] != 1 {
		t.Fatalf("SeenType = %d, want 1", state.SeenType["Boeing 737-800"])
	}
}

func TestEvaluateBatchDifferentFlightIncrementsAgain(t *testing.T) {
	t.Parallel()

	classifier := stubClassifier{
		makeByIcao: map[string]string{"B738": "Boeing 737-800"},
	}
	first := baseRecord()
	state, _ := EvaluateBatch(emptyState(), []AircraftRecord{first}, classifier, time.Now())

	second := first
	second.Flight = "DLH456"
	state, current := EvaluateBatch(state, []AircraftRecord{second}, classifier, time.Now())
	if current[0].LastFlightNo != "DLH456" {
		t.Fatalf("LastFlightNo = %q, want DLH456", current[0].LastFlightNo)
	}
	if state.SightedTypes != 2 {
		t.Fatalf("SightedTypes = %d, want 2", state.SightedTypes)
	}
}

func TestEvaluateBatchIdentifyingCallsignDoesNotCountAsNewFlight(t *testing.T) {
	t.Parallel()

	classifier := stubClassifier{
		makeByIcao: map[string]string{"B738": "Boeing 737-800"},
	}
	unknown := baseRecord()
	unknown.Flight = ""
	state, current := EvaluateBatch(emptyState(), []AircraftRecord{unknown}, classifier, time.Now())
	if current[0].LastFlightNo != FlightUnknown {
		t.Fatalf("expected unknown flight after first pass, got %q", current[0].LastFlightNo)
	}
	if state.SightedTypes != 1 {
		t.Fatalf("first pass should still classify type, SightedTypes = %d", state.SightedTypes)
	}

	identified := unknown
	identified.Flight = "BAW123"
	state, current = EvaluateBatch(state, []AircraftRecord{identified}, classifier, time.Now())
	if current[0].LastFlightNo != "BAW123" {
		t.Fatalf("LastFlightNo = %q, want BAW123", current[0].LastFlightNo)
	}
	if state.SightedTypes != 1 {
		t.Fatalf("identifying callsign must not increment counts, SightedTypes = %d", state.SightedTypes)
	}
}

func TestEvaluateBatchRareType(t *testing.T) {
	t.Parallel()

	classifier := stubClassifier{
		makeByIcao: map[string]string{"B738": "Boeing 737-800"},
	}
	state := emptyState()
	state.SightedTypes = 1099
	record := baseRecord()
	_, current := EvaluateBatch(state, []AircraftRecord{record}, classifier, time.Now())
	if current[0].Rarities&RareType == 0 {
		t.Fatalf("expected RareType flag, got %b", current[0].Rarities)
	}
}

func TestEvaluateBatchOperatorIcaoFallback(t *testing.T) {
	t.Parallel()

	classifier := stubClassifier{
		opByIcao: map[string][2]string{"BAW": {"British Airways", "united kingdom"}},
	}
	record := baseRecord()
	_, current := EvaluateBatch(emptyState(), []AircraftRecord{record}, classifier, time.Now())
	if current[0].Operator != "British Airways" {
		t.Fatalf("Operator = %q, want British Airways", current[0].Operator)
	}
}

func TestEvaluateBatchOperatorMilFallback(t *testing.T) {
	t.Parallel()

	classifier := stubClassifier{
		opByMil: map[string]string{"RCH": "US Air Mobility Command"},
	}
	record := baseRecord()
	record.Flight = "RCH801"
	_, current := EvaluateBatch(emptyState(), []AircraftRecord{record}, classifier, time.Now())
	if current[0].Operator != "US Air Mobility Command" {
		t.Fatalf("Operator = %q, want mil operator", current[0].Operator)
	}
}

func TestEvaluateBatchOperatorOwnOpFallback(t *testing.T) {
	t.Parallel()

	record := baseRecord()
	record.OwnOp = "Private Owner"
	_, current := EvaluateBatch(emptyState(), []AircraftRecord{record}, stubClassifier{}, time.Now())
	if current[0].Operator != "Private Owner" {
		t.Fatalf("Operator = %q, want Private Owner", current[0].Operator)
	}
}

func TestEvaluateBatchCountryOperatorThenHexThenRegistration(t *testing.T) {
	t.Parallel()

	t.Run("operator country", func(t *testing.T) {
		t.Parallel()
		classifier := stubClassifier{
			opByIcao: map[string][2]string{"BAW": {"British Airways", "united kingdom"}},
		}
		_, current := EvaluateBatch(emptyState(), []AircraftRecord{baseRecord()}, classifier, time.Now())
		if current[0].Country != "UNITED KINGDOM" {
			t.Fatalf("Country = %q, want UNITED KINGDOM", current[0].Country)
		}
	})

	t.Run("hex country", func(t *testing.T) {
		t.Parallel()
		classifier := stubClassifier{
			countryByHex: map[string]string{"abc123": "germany"},
		}
		_, current := EvaluateBatch(emptyState(), []AircraftRecord{baseRecord()}, classifier, time.Now())
		if current[0].Country != "GERMANY" {
			t.Fatalf("Country = %q, want GERMANY", current[0].Country)
		}
	})

	t.Run("registration country", func(t *testing.T) {
		t.Parallel()
		classifier := stubClassifier{
			countryByReg: map[string]string{"G-ABCD": "united kingdom"},
		}
		_, current := EvaluateBatch(emptyState(), []AircraftRecord{baseRecord()}, classifier, time.Now())
		if current[0].Country != "UNITED KINGDOM" {
			t.Fatalf("Country = %q, want UNITED KINGDOM", current[0].Country)
		}
	})
}
