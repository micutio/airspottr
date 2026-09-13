package application

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
)

// SpottingService or spotting orchestrator coordinates:
// - fetching aircraft data
// - processing the dashboard
// - requesting route data for missing callsigns
// - saving state
// - emitting notifications.
type SpottingService interface {
	UpdateSightings(aircraftRecords []obs.AircraftRecord) obs.AircraftSighting
}
