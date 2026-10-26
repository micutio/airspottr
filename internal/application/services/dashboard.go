// Package services provides the Dashboard type and all associated program logic.
package services

import (
	"errors"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"time"

	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// TODO: Remove and privatise as many fields as possible.
// TODO: Create specialized type for highest and fastest.

// Errors used by the Dashboard.
var (
	errCoordMismatch = errors.New("state coordinate mismatch")
)

type noopClassifier struct{}

func (noopClassifier) AircraftMake(string) (string, bool)               { return "", false }
func (noopClassifier) OperatorByIcao(string) (string, string, bool)     { return "", "", false }
func (noopClassifier) OperatorByMil(string) (string, bool)              { return "", false }
func (noopClassifier) CountryByHex(string) (string, bool)               { return "", false }
func (noopClassifier) CountryByRegistration(string) (string, bool)      { return "", false }

// Dashboard coordinates aircraft observation evaluation, route lookups, and notifications.
type Dashboard struct {
	obs.State
	currentSightings  []obs.AircraftSighting // volatile cache of recently sighted aircraft
	classifier        obs.Classifier
	notifier          NotificationService
	flightRouteLookup func([]string) map[string]ref.FlightrouteRecord
	IsWarmup          bool
	Lat               float64
	Lon               float64
	ErrOut            log.Logger
}

func NewDashboard(lat, lon float64,
	classifier obs.Classifier,
	stderr io.Writer,
) *Dashboard {
	if classifier == nil {
		classifier = noopClassifier{}
	}
	dashboard := Dashboard{
		State: obs.State{
			Observer:         ref.NewCoordinates(lat, lon),
			Sightings:        make(map[string]obs.AircraftSighting),
			SeenType:         make(map[string]int),
			SeenOperator:     make(map[string]int),
			SeenCountry:      make(map[string]int),
			SightedTypes:     0,
			SightedOperators: 0,
			SightedCountries: 0,
			Fastest:          obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
			Highest:          obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
		},
		currentSightings:  []obs.AircraftSighting{},
		classifier:        classifier,
		notifier:          nil,
		flightRouteLookup: nil,
		IsWarmup:          true,
		Lat:               lat,
		Lon:               lon,
		ErrOut:            *log.New(stderr, "dashboard ", log.LstdFlags),
	}

	dashboard.ErrOut.Println("Dashboard init")

	return &dashboard
}

// ConfigureRefresh wires optional notification and flight-route lookup callbacks for RefreshBatch.
func (db *Dashboard) ConfigureRefresh(
	notifier NotificationService,
	flightRouteLookup func([]string) map[string]ref.FlightrouteRecord,
) {
	db.notifier = notifier
	db.flightRouteLookup = flightRouteLookup
}

// RefreshBatch evaluates a batch of aircraft records, emits rarity notifications, and resolves missing routes.
func (db *Dashboard) RefreshBatch(
	records []obs.AircraftRecord,
	toggles obs.RarityNotifyToggles,
) []obs.AircraftSighting {
	db.ProcessAircraftRecords(records)
	if db.notifier != nil {
		db.notifier.EmitRarityNotifications(db.currentSightings, toggles)
	}
	if db.flightRouteLookup != nil {
		callsignsWithoutRoute := GetCallsignsRequiringRoutes(db.currentSightings)
		if len(callsignsWithoutRoute) > 0 {
			routes := db.flightRouteLookup(callsignsWithoutRoute)
			if len(routes) > 0 {
				db.AssignFlightRoutes(routes)
			}
		}
	}
	return db.currentSightings
}

func (db *Dashboard) FinishWarmupPeriod() {
	db.IsWarmup = false
}

//////////////////////////////////////////////////////////////////////////////
/// Processing of all aircraft: civilian, military, government, private.    //
//////////////////////////////////////////////////////////////////////////////

// ProcessAircraftRecords implements the interfaces.SpottingService method.
func (db *Dashboard) ProcessAircraftRecords(aircraftRecords []obs.AircraftRecord) {
	db.currentSightings = obs.EvaluateBatch(&db.State, aircraftRecords, db.classifier, time.Now())
}

// GetCurrentSightings implements the interfaces.SpottingService method.
func (db *Dashboard) GetCurrentSightings() []obs.AircraftSighting {
	return db.currentSightings
}

// GetFastest implements the interfaces.SpottingService method.
func (db *Dashboard) GetFastest() obs.AircraftRecord {
	return db.Fastest
}

// GetHighest implements the interfaces.SpottingService method.
func (db *Dashboard) GetHighest() obs.AircraftRecord {
	return db.Highest
}

// AircraftMake looks up the manufacturer/model description for an ICAO type code.
func (db *Dashboard) AircraftMake(icaoType string) (string, bool) {
	if db.classifier == nil {
		return "", false
	}
	return db.classifier.AircraftMake(icaoType)
}

// GetCallsignsRequiringRoutes attempts to assign cached route information to all
// sightings.
// It returns a list of callsigns without known routes, to allow querying
// online for these cases.
func GetCallsignsRequiringRoutes(currentSightings []obs.AircraftSighting) []string {
	emptyFlightroute := ref.FlightrouteRecord{} //nolint:exhaustruct_v5 // zero-value check
	var callsignsWithoutRoute []string
	for _, sighting := range currentSightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		if sighting.Flightroute != emptyFlightroute {
			// A Flight route is already set.
			continue
		}

		// No routes found, record this callsign to request route from adsbdb
		callsignsWithoutRoute = append(callsignsWithoutRoute, sighting.LastFlightNo)
	}
	return callsignsWithoutRoute
}

// AssignFlightRoutes assigns the given Flight routes to all flights matching the callsign.
func (db *Dashboard) AssignFlightRoutes(flightRouteRecords map[string]ref.FlightrouteRecord) {
	emptyFlightroute := ref.FlightrouteRecord{} //nolint:exhaustruct_v5 // zero-value check
	for idx := range db.currentSightings {
		sighting := &db.currentSightings[idx]
		if sighting.LastFlightNo == obs.FlightUnknown || sighting.Flightroute != emptyFlightroute {
			continue
		}

		if flightRoute, ok := flightRouteRecords[sighting.LastFlightNo]; ok {
			sighting.Flightroute = flightRoute
		}
	}

	for callsign, flightRoute := range flightRouteRecords {
		for hex, sighting := range db.Sightings {
			if sighting.LastFlightNo == callsign {
				sighting.Flightroute = flightRoute
				db.Sightings[hex] = sighting
			}
		}
	}
}

func (db *Dashboard) SaveState(
	pendingCallsigns []string,
) *PersistentState {
	if db.Sightings == nil {
		db.Sightings = make(map[string]obs.AircraftSighting)
	}

	dashboardState := DashboardState{ //nolint:exhaustruct_v5 // removed deprecated items
		IsWarmup:           db.IsWarmup,
		Lat:                db.Lat,
		Lon:                db.Lon,
		Fastest:            db.Fastest,
		Highest:            db.Highest,
		CurrentAircraft:    nil,
		AircraftSightings:  db.Sightings,
		TotalTypeCount:     db.SightedTypes,
		TotalOperatorCount: db.SightedOperators,
		TotalCountryCount:  db.SightedCountries,
		SeenTypeCount:      db.SeenType,
		SeenOperatorCount:  db.SeenOperator,
		SeenCountryCount:   db.SeenCountry,
	}
	locationKey := ref.LocationKey(db.Lat, db.Lon)

	return &PersistentState{
		FlightrouteRepoState: FlightrouteRepoState{
			PendingCallsigns: append([]string(nil), pendingCallsigns...),
		},
		LocationStates: map[string]DashboardState{locationKey: dashboardState},
	}
}

func (db *Dashboard) RestoreState(state *DashboardState) error {
	if state == nil {
		return nil
	}
	if ref.LocationKey(state.Lat, state.Lon) != ref.LocationKey(db.Lat, db.Lon) {
		return errCoordMismatch
	}

	db.State = obs.State{
		Observer:         ref.NewCoordinates(db.Lat, db.Lon),
		Sightings:        state.AircraftSightings,
		SeenType:         state.SeenTypeCount,
		SeenOperator:     state.SeenOperatorCount,
		SeenCountry:      state.SeenCountryCount,
		SightedTypes:     state.TotalTypeCount,
		SightedOperators: state.TotalOperatorCount,
		SightedCountries: state.TotalCountryCount,
		Fastest:          state.Fastest,
		Highest:          state.Highest,
	}
	if db.Sightings == nil {
		db.Sightings = make(map[string]obs.AircraftSighting)
	}
	if db.SeenType == nil {
		db.SeenType = make(map[string]int)
	}
	if db.SeenOperator == nil {
		db.SeenOperator = make(map[string]int)
	}
	if db.SeenCountry == nil {
		db.SeenCountry = make(map[string]int)
	}
	db.IsWarmup = state.IsWarmup

	return nil
}
