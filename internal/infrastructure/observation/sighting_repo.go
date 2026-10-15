package observation

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// SightingRepo implements the repositories.SightingRepo interface.
type SightingRepo struct {
	sightings map[string]obs.AircraftSighting
}

func NewSightingRepo() *SightingRepo {
	return &SightingRepo{
		sightings: make(map[string]obs.AircraftSighting),
	}
}

// UpdateSighting implements the repositories.SightingRepo interface method.
func (sgtRepo *SightingRepo) UpdateSighting(icaoHex string, sighting obs.AircraftSighting) {
	sgtRepo.sightings[icaoHex] = sighting
}

// RestoreSightings implements the repositories.SightingRepo interface method.
// Replaces all internally kept sighting records.
func (sgtRepo *SightingRepo) RestoreSightings(sightings map[string]obs.AircraftSighting) {
	sgtRepo.sightings = sightings
}

// GetAllSightings implements the repositories.SightingRepo interface method.
func (sgtRepo *SightingRepo) GetAllSightings() map[string]obs.AircraftSighting {
	return sgtRepo.sightings
}

// UpdateFlightroute implements the repositories.SightingRepo interface method.
func (sgtRepo *SightingRepo) UpdateFlightroute(callsign string, flightroute ref.FlightrouteRecord) {
	for hex, sighting := range sgtRepo.sightings {
		if sighting.LastFlightNo == callsign {
			sighting.Flightroute = flightroute
			sgtRepo.sightings[hex] = sighting
		}
	}
}
