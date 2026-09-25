package interfaces

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// Example workflow
// aircraftRecords := aircraftSource.FetchNearbyAircraft(lat, lon)
// dashboard.ProcessAircraft(aircraftRecords)
// callsigns := dashboard.CallsignsNeedingRoutes()
// routes := flightRouteRepository.FetchRoutes(callsigns)
// dashboard.AssignRoutes(routes)
// notifications.EmitRareSightings(dashboard.RareSightings())
// stateRepo.Save(dashboard, pendingCallsigns)

// SpottingService or spotting orchestrator coordinates:
// - fetching aircraft data
// - processing the dashboard
// - requesting route data for missing callsigns
// - saving state
// - emitting notifications.
type SpottingService interface {
	// ProcessAircraftRecords takes currently observed aircraft messages to update
	// sightings and determine sighting rarity.
	// If an aircraft has not been recorded before -> create a new sighting.
	// If an aircraft has been recorded before on a different flight -> create a new sighting.
	// If an aircraft has been recorded before on the same flight -> update existing sighting.
	// If a sighting contains either a type, operator or country of origin that
	// has been counted below a certain threshold, then this sighting is now
	// considered rare and can be used to emit notifications to the user.
	ProcessAircraftRecords(aircraftRecords []obs.AircraftRecord)

	// GetCurrentSightings returns the most recent aircraft sightings.
	GetCurrentSightings() []obs.AircraftSighting

	// GetFastest returns the aircraft observed with the highest speed over ground.
	GetFastest() obs.AircraftRecord

	// GetHighest returns the aircraft observed at the highest altitude.
	GetHighest() obs.AircraftRecord

	// AssignFlightRoutes applies the given flight routes to all sightings they apply to.
	AssignFlightRoutes(flightRouteRecords map[string]ref.FlightrouteRecord)
}
