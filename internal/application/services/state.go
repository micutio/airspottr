package services

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// AirspottrState encapsulates the internal state of the app to hide the exported JSON fields of
// persistentState.
type AirspottrState struct {
	InternalState PersistentState
}

type PersistedRareSighting struct {
	Rarities obs.RarityFlag `json:"rarities"`
	Hex      string         `json:"hex"`
}

type FlightrouteRepoState struct {
	PendingCallsigns []string `json:"pending_callsigns"`
}

type PersistentState struct {
	DashboardState       DashboardState       `json:"dashboard"`
	FlightrouteRepoState FlightrouteRepoState `json:"request"`
}

type DashboardState struct {
	IsWarmup        bool                 `json:"is_warmup"`
	Lat             float64              `json:"lat"`
	Lon             float64              `json:"lon"`
	Fastest         obs.AircraftRecord   `json:"fastest"`
	Highest         obs.AircraftRecord   `json:"highest"`
	CurrentAircraft []obs.AircraftRecord `json:"current_aircraft"`
	// Deprecated: no longer in use
	RareSightings []PersistedRareSighting `json:"rare_sightings"`
	// Deprecated: no longer in use
	CachedFlightRoutes map[string]*ref.FlightrouteRecord `json:"cached_flight_routes"`
	AircraftSightings  map[string]obs.AircraftSighting   `json:"aircraft_sightings"`
	TotalTypeCount     int                               `json:"total_type_count"`
	TotalOperatorCount int                               `json:"total_operator_count"`
	TotalCountryCount  int                               `json:"total_country_count"`
	SeenTypeCount      map[string]int                    `json:"seen_type_count"`
	SeenOperatorCount  map[string]int                    `json:"seen_operator_count"`
	SeenCountryCount   map[string]int                    `json:"seen_country_count"`
}
