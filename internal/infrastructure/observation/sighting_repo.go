package observation

import (
	"math"
	"time"

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

// GetOrCreateSighting implements the repositories.SightingRepo interface method.
func (sgtRepo *SightingRepo) GetOrCreateSighting(
	lat, lon float64,
	aircraft obs.AircraftRecord,
) (obs.AircraftSighting, bool) {
	lastSeenMsBeforeNow := time.Duration(aircraft.Seen) * time.Second
	lastSeenTime := time.Now().Add(-lastSeenMsBeforeNow)

	// Retrieve previous sighting or create new one.
	sighting, exists := sgtRepo.sightings[aircraft.Hex]
	if !exists {
		sighting = obs.AircraftSighting{
			Rarities:     obs.NoRarity,
			LastSeen:     lastSeenTime,
			LastFlightNo: obs.FlightUnknown,
			Registration: aircraft.Registration,
			Latitude:     aircraft.Lat,
			Longitude:    aircraft.Lon,
			Direction:    ref.GetDirection(lat, lon, aircraft.Lat, aircraft.Lon),
			Distance:     math.MaxInt,
			TypeShort:    "",
			TypeDesc:     obs.TypeUnknown,
			Operator:     obs.OperatorUnknown,
			Country:      ref.CountryUnknown,
			Info:         "",
			Flightroute:  ref.FlightrouteRecord{}, //nolint:exhaustruct_v5 // using default values
			LastRecord:   aircraft,
		}
	} else {
		sighting.LastRecord = aircraft
	}

	if sighting.Registration == "" {
		sighting.Registration = aircraft.Registration
	}

	return sighting, !exists
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
func (sgtRepo *SightingRepo) UpdateFlightroute(icaoHex string, flightroute ref.FlightrouteRecord) {
	sighting, exists := sgtRepo.sightings[icaoHex]
	if !exists {
		return
	}

	sighting.Flightroute = flightroute
	sgtRepo.sightings[icaoHex] = sighting
}
