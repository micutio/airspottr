package services

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// AircraftDataSource represents a source of fresh aircraft observations.
type AircraftDataSource interface {
	RequestAircraft() []obs.AircraftRecord
}

// FlightRouteDataSource represents a source of flight routes keyed by callsign.
type FlightRouteDataSource interface {
	GetFlightroutes(callsigns []string) map[string]ref.FlightrouteRecord
	GetPendingCallsigns() []string
	RestorePendingCallsigns(pendingCallsigns []string)
}

// NotificationService emits rarity notifications for newly observed sightings.
type NotificationService interface {
	EmitRarityNotifications(sightings []obs.AircraftSighting, toggles obs.RarityNotifyToggles)
	DefaultRarityNotifyToggles() obs.RarityNotifyToggles
}

// AirspottrState encapsulates the internal state of the app to hide the exported JSON fields of
// PersistentState.
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
	FlightrouteRepoState FlightrouteRepoState      `json:"request"`
	LocationStates       map[string]DashboardState `json:"location_states"`
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
