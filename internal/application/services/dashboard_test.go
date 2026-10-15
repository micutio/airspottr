//nolint:exhaustruct_v5
package services

import (
	"io"
	"slices"
	"testing"

	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	infraobs "github.com/micutio/airspottr/internal/infrastructure/observation"
)

type stubTypeRepo struct{}

func (stubTypeRepo) GetAircraftType(string) (ref.IcaoAircraftSpec, bool) {
	return ref.IcaoAircraftSpec{}, false
}

type stubOperatorRepo struct{}

func (stubOperatorRepo) GetOperatorByIcao(string) (ref.IcaoOperator, bool) {
	return ref.IcaoOperator{}, false
}

func (stubOperatorRepo) GetOperatorByMilCode(string) (string, bool) {
	return "", false
}

type stubCountryRepo struct{}

func (stubCountryRepo) GetCountryByHexCode(string) (string, error) {
	return "", nil
}

func (stubCountryRepo) GetCountryByRegistration(string) (string, bool) {
	return "", false
}

func TestGetCallsignsRequiringRoutes(t *testing.T) {
	t.Parallel()

	assignedRoute := *ref.GetDefaultFlightrouteRecord()
	assignedRoute.Callsign = "DLH456"

	sightings := []obs.AircraftSighting{
		{LastFlightNo: obs.FlightUnknown},
		{LastFlightNo: "BAW123", Flightroute: ref.FlightrouteRecord{}},
		{LastFlightNo: "DLH456", Flightroute: assignedRoute},
		{LastFlightNo: "AFR789", Flightroute: *ref.GetDefaultFlightrouteRecord()},
	}

	got := GetCallsignsRequiringRoutes(sightings)
	want := []string{"BAW123"}
	if !slices.Equal(got, want) {
		t.Fatalf("GetCallsignsRequiringRoutes() = %v, want %v", got, want)
	}
}

func TestRefreshBatchAssignsFlightRoutesByCallsign(t *testing.T) {
	t.Parallel()

	sightingRepo := infraobs.NewSightingRepo()
	dashboard := NewDashboard(
		53.55,
		9.99,
		sightingRepo,
		stubTypeRepo{},
		stubOperatorRepo{},
		stubCountryRepo{},
		io.Discard,
	)

	expectedRoute := *ref.GetDefaultFlightrouteRecord()
	expectedRoute.Callsign = "BAW123"
	expectedRoute.Origin.IataCode = "LHR"
	expectedRoute.Destination.IataCode = "JFK"

	var requestedCallsigns []string
	useCase := NewRefreshUseCase(
		dashboard,
		nil,
		func(callsigns []string) map[string]ref.FlightrouteRecord {
			requestedCallsigns = append(requestedCallsigns, callsigns...)
			return map[string]ref.FlightrouteRecord{
				"BAW123": expectedRoute,
			}
		})

	records := []obs.AircraftRecord{
		{
			Hex:    "abc123",
			Flight: "BAW123",
			Lat:    53.6,
			Lon:    10.0,
		},
	}

	current := useCase.RefreshBatch(records, obs.RarityNotifyToggles{})
	if !slices.Equal(requestedCallsigns, []string{"BAW123"}) {
		t.Fatalf("requested callsigns = %v, want [BAW123]", requestedCallsigns)
	}
	if len(current) != 1 {
		t.Fatalf("len(current) = %d, want 1", len(current))
	}
	if current[0].Flightroute.Origin.IataCode != "LHR" ||
		current[0].Flightroute.Destination.IataCode != "JFK" {
		t.Fatalf("current sighting route not updated: %#v", current[0].Flightroute)
	}

	stored, exists := sightingRepo.GetAllSightings()["abc123"]
	if !exists {
		t.Fatal("expected sighting stored in repo under hex abc123")
	}
	if stored.Flightroute != expectedRoute {
		t.Fatalf("stored sighting route = %#v, want %#v", stored.Flightroute, expectedRoute)
	}

	// Subsequent refresh for the same aircraft/flight should preserve the route without re-requesting.
	requestedCallsigns = nil
	current = useCase.RefreshBatch(records, obs.RarityNotifyToggles{})
	if len(requestedCallsigns) != 0 {
		t.Fatalf("expected no route lookup on second pass, got %v", requestedCallsigns)
	}
	if current[0].Flightroute != expectedRoute {
		t.Fatalf("expected route preserved on second pass, got %#v", current[0].Flightroute)
	}
}
