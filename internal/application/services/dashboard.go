// Package services provides the Dashboard type and all associated program logic.
package services

import (
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"time"

	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	rep "github.com/micutio/airspottr/internal/domain/repositories"
)

// TODO: Remove and privatise as many fields as possible.
// TODO: Create specialized type for highest and fastest.

// Errors used by the Dashboard.
var (
	errCoordMismatch = errors.New("state coordinate mismatch")
)

// Dashboard implements the interfaces.SpottingService interface.
type Dashboard struct {
	session               *obs.SpottingSession
	currentSightings      []obs.AircraftSighting // volatile cache of recently sighted aircraft
	sightingRepo          rep.SightingRepo
	classifier            obs.Classifier
	IsWarmup              bool
	Lat                   float64
	Lon                   float64
	fastest               obs.AircraftRecord
	highest               obs.AircraftRecord
	SightedTypesCount     int            // total count of unique sighted types
	SightedOperatorsCount int            // total count of unique sighted operators
	SightedCountriesCount int            // total count of unique sighted countries of origin
	SeenTypeCount         map[string]int // types mapped to how often seen
	SeenOperatorCount     map[string]int // airlines mapped to how often seen
	SeenCountryCount      map[string]int // airlines mapped to how often seen
	ErrOut                log.Logger
}

func NewDashboard(lat, lon float64,
	sightingRepo rep.SightingRepo,
	aircraftTypeRepo rep.AircraftTypeRepo,
	operatorRepo rep.OperatorRepository,
	countryRepo rep.CountryRepository,
	stderr io.Writer,
) *Dashboard {
	classifier := newReferenceClassifier(
		aircraftTypeRepo,
		operatorRepo,
		countryRepo,
		log.New(stderr, "classifier ", log.LstdFlags),
	)
	session := obs.NewSpottingSession(ref.NewCoordinates(lat, lon))
	dashboard := Dashboard{
		session:               session,
		currentSightings:      []obs.AircraftSighting{},
		sightingRepo:          sightingRepo,
		classifier:            classifier,
		IsWarmup:              true,
		Lat:                   lat,
		Lon:                   lon,
		fastest:               obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
		highest:               obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
		SightedTypesCount:     0,
		SightedOperatorsCount: 0,
		SightedCountriesCount: 0,
		SeenTypeCount:         make(map[string]int),
		SeenOperatorCount:     make(map[string]int),
		SeenCountryCount:      make(map[string]int),
		ErrOut:                *log.New(stderr, "dashboard ", log.LstdFlags),
	}

	dashboard.ErrOut.Println("Dashboard init")
	dashboard.syncStateFromSession()

	return &dashboard
}

func (db *Dashboard) FinishWarmupPeriod() {
	db.IsWarmup = false
	if db.session != nil {
		db.session.SetWarmup(false)
	}
}

func (db *Dashboard) syncStateFromSession() {
	if db.session == nil {
		return
	}
	state := db.session.Snapshot()
	db.fastest = state.Fastest
	db.highest = state.Highest
	db.SightedTypesCount = state.SightedTypes
	db.SightedOperatorsCount = state.SightedOperators
	db.SightedCountriesCount = state.SightedCountries
	db.SeenTypeCount = state.SeenType
	db.SeenOperatorCount = state.SeenOperator
	db.SeenCountryCount = state.SeenCountry
	db.Lat = state.Observer.Latitude
	db.Lon = state.Observer.Longitude
}

func (db *Dashboard) syncSessionFromDashboard() {
	if db.session == nil {
		db.session = obs.NewSpottingSession(ref.NewCoordinates(db.Lat, db.Lon))
	}
	state := obs.State{
		Observer:         ref.NewCoordinates(db.Lat, db.Lon),
		Sightings:        db.sightingRepo.GetAllSightings(),
		SeenType:         db.SeenTypeCount,
		SeenOperator:     db.SeenOperatorCount,
		SeenCountry:      db.SeenCountryCount,
		SightedTypes:     db.SightedTypesCount,
		SightedOperators: db.SightedOperatorsCount,
		SightedCountries: db.SightedCountriesCount,
		Fastest:          db.fastest,
		Highest:          db.highest,
	}
	if restoreErr := db.session.Restore(state); restoreErr != nil {
		db.ErrOut.Printf("sync spotting session from dashboard: %v", restoreErr)
	}
}

//////////////////////////////////////////////////////////////////////////////
/// Processing of all aircraft: civilian, military, government, private.    //
//////////////////////////////////////////////////////////////////////////////

// ProcessAircraftRecords implements the interfaces.SpottingService method.
func (db *Dashboard) ProcessAircraftRecords(aircraftRecords []obs.AircraftRecord) {
	state := db.observationState()
	currentSightings := obs.EvaluateBatch(&state, aircraftRecords, db.classifier, time.Now())
	db.applyObservationState(state, currentSightings)
}

func (db *Dashboard) observationState() obs.State {
	if db.session == nil {
		db.session = obs.NewSpottingSession(ref.NewCoordinates(db.Lat, db.Lon))
	}
	state := db.session.Snapshot()
	state.Sightings = db.sightingRepo.GetAllSightings()
	state.SeenType = db.SeenTypeCount
	state.SeenOperator = db.SeenOperatorCount
	state.SeenCountry = db.SeenCountryCount
	state.SightedTypes = db.SightedTypesCount
	state.SightedOperators = db.SightedOperatorsCount
	state.SightedCountries = db.SightedCountriesCount
	state.Fastest = db.fastest
	state.Highest = db.highest
	return state
}

func (db *Dashboard) applyObservationState(
	state obs.State,
	currentSightings []obs.AircraftSighting,
) {
	if db.session == nil {
		db.session = obs.NewSpottingSession(ref.NewCoordinates(db.Lat, db.Lon))
	}
	if restoreErr := db.session.Restore(state); restoreErr != nil {
		db.ErrOut.Printf("restore spotting session: %v", restoreErr)
	}
	db.sightingRepo.RestoreSightings(state.Sightings)
	db.syncStateFromSession()
	db.currentSightings = currentSightings
}

// GetCurrentSightings implements the interfaces.SpottingService method.
func (db *Dashboard) GetCurrentSightings() []obs.AircraftSighting {
	return db.currentSightings
}

// GetFastest implements the interfaces.SpottingService method.
func (db *Dashboard) GetFastest() obs.AircraftRecord {
	return db.fastest
}

// GetHighest implements the interfaces.SpottingService method.
func (db *Dashboard) GetHighest() obs.AircraftRecord {
	return db.highest
}

// GetCallsignsRequiringRoutes attempts to assign cached route information to all
// sightings.
// It returns a list of callsigns without known routes, to allow querying
// online for these cases.
func GetCallsignsRequiringRoutes(currentSightings []obs.AircraftSighting) []string {
	defaultFlightrouteRecord := *ref.GetDefaultFlightrouteRecord()
	var callsignsWithoutRoute []string
	for _, sighting := range currentSightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		if sighting.Flightroute != defaultFlightrouteRecord {
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
	defaultFlightrouteRecord := *ref.GetDefaultFlightrouteRecord()
	for _, sighting := range db.currentSightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		if sighting.Flightroute != defaultFlightrouteRecord {
			// A Flight route is already set.
			continue
		}

		if flightRoute, ok := flightRouteRecords[sighting.LastFlightNo]; ok {
			sighting.Flightroute = flightRoute
			continue
		}
	}

	for hexCode, flightRoute := range flightRouteRecords {
		db.sightingRepo.UpdateFlightroute(hexCode, flightRoute)
	}
}

func (db *Dashboard) SaveState(
	pendingCallsigns []string,
) *PersistentState {
	if db.session == nil {
		db.session = obs.NewSpottingSession(ref.NewCoordinates(db.Lat, db.Lon))
	}
	db.syncSessionFromDashboard()
	state := db.session.Snapshot()
	aircraftSightings := db.sightingRepo.GetAllSightings()
	for hex, sighting := range aircraftSightings {
		aircraftSightings[hex] = sighting
	}

	dashboardState := DashboardState{ //nolint:exhaustruct_v5 // removed deprecated items
		IsWarmup:           db.IsWarmup,
		Lat:                db.Lat,
		Lon:                db.Lon,
		Fastest:            state.Fastest,
		Highest:            state.Highest,
		CurrentAircraft:    nil,
		AircraftSightings:  aircraftSightings,
		TotalTypeCount:     state.SightedTypes,
		TotalOperatorCount: state.SightedOperators,
		TotalCountryCount:  state.SightedCountries,
		SeenTypeCount:      state.SeenType,
		SeenOperatorCount:  state.SeenOperator,
		SeenCountryCount:   state.SeenCountry,
	}
	locationKey := ref.LocationKey(db.Lat, db.Lon)

	return &PersistentState{
		DashboardState: dashboardState,
		FlightrouteRepoState: FlightrouteRepoState{
			PendingCallsigns: append([]string(nil), pendingCallsigns...),
		},
		LocationStates: map[string]DashboardState{locationKey: dashboardState},
	}
}

func (db *Dashboard) RestoreState(state *DashboardState) error {
	if state.Lat != db.Lat || state.Lon != db.Lon {
		return errCoordMismatch
	}

	if db.session == nil {
		db.session = obs.NewSpottingSession(ref.NewCoordinates(db.Lat, db.Lon))
	}
	if restoreErr := db.session.Restore(obs.State{
		Observer:         ref.NewCoordinates(state.Lat, state.Lon),
		Sightings:        state.AircraftSightings,
		SeenType:         state.SeenTypeCount,
		SeenOperator:     state.SeenOperatorCount,
		SeenCountry:      state.SeenCountryCount,
		SightedTypes:     state.TotalTypeCount,
		SightedOperators: state.TotalOperatorCount,
		SightedCountries: state.TotalCountryCount,
		Fastest:          state.Fastest,
		Highest:          state.Highest,
	}); restoreErr != nil {
		return fmt.Errorf("restore spotting session: %w", restoreErr)
	}

	db.IsWarmup = state.IsWarmup
	db.session.SetWarmup(state.IsWarmup)
	db.sightingRepo.RestoreSightings(state.AircraftSightings)
	db.syncStateFromSession()

	return nil
}
