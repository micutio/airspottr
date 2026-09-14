package interfaces

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
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
	UpdateSightings(aircraftRecords []obs.AircraftRecord) obs.AircraftSighting
}
