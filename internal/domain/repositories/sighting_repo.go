package repositories

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

type SightingRepo interface {
	// UpdateSighting updates the sighting for a given aircraft identified by icao hex code.
	UpdateSighting(icaoHex string, sighting obs.AircraftSighting)

	// RestoreSightings adds a sighting to the repo, mostly to be used for persistence mgmt.
	RestoreSightings(sightings map[string]obs.AircraftSighting)

	// GetAllSightings retrieves all sightings for persisting them on disk.
	GetAllSightings() map[string]obs.AircraftSighting

	// UpdateFlightroute updates the flightroute of the given sighting.
	UpdateFlightroute(icaoHex string, flightroute ref.FlightrouteRecord)
}
