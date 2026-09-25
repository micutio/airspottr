package repositories

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

type SightingRepo interface {
	// GetOrCreateSighting uses the given aircraft records to update existing sightings.
	// If no sighting can be found, a new one will be created.
	// The latitude, longitude coordinates are used to determine the current direction of the
	// sighted aircraft relative to the given position, i.e. our position.
	// Returns a sighting record for the corresponding aircraft and a boolean value indicating
	// whether this sighting new (`true`).
	GetOrCreateSighting(lat, lon float64, aircraft obs.AircraftRecord) (obs.AircraftSighting, bool)

	// UpdateSighting updates the sighting for a given aircraft identified by icao hex code.
	UpdateSighting(icaoHex string, sighting obs.AircraftSighting)

	// RestoreSightings adds a sighting to the repo, mostly to be used for persistence mgmt.
	RestoreSightings(sightings map[string]obs.AircraftSighting)

	// GetAllSightings retrieves all sightings for persisting them on disk.
	GetAllSightings() map[string]obs.AircraftSighting

	// UpdateFlightroute updates the flightroute of the given sighting.
	UpdateFlightroute(icaoHex string, flightroute ref.FlightrouteRecord)
}
