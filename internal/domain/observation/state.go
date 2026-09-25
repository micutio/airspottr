package observation

import (
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// State is the spotting-session snapshot used by EvaluateBatch.
type State struct {
	Observer         ref.Coordinates
	Sightings        map[string]AircraftSighting // keyed by ICAO hex
	SeenType         map[string]int
	SeenOperator     map[string]int
	SeenCountry      map[string]int
	SightedTypes     int
	SightedOperators int
	SightedCountries int
	Fastest          AircraftRecord
	Highest          AircraftRecord
}
