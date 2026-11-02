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
	Fastest          AircraftSighting
	Highest          AircraftSighting
}

// NewState creates an initialized spotting session State for the given observer coordinates.
func NewState(observer ref.Coordinates) State {
	return State{
		Observer:         observer,
		Sightings:        make(map[string]AircraftSighting),
		SeenType:         make(map[string]int),
		SeenOperator:     make(map[string]int),
		SeenCountry:      make(map[string]int),
		SightedTypes:     0,
		SightedOperators: 0,
		SightedCountries: 0,
		Fastest:          AircraftSighting{}, //nolint:exhaustruct_v5 // default zero-value sighting
		Highest:          AircraftSighting{}, //nolint:exhaustruct_v5 // default zero-value sighting
	}
}

// EnsureInitialized guarantees all internal maps are non-nil
// (e.g. after JSON unmarshaling or zero-value construction).
func (s *State) EnsureInitialized() {
	if s.Sightings == nil {
		s.Sightings = make(map[string]AircraftSighting)
	}
	if s.SeenType == nil {
		s.SeenType = make(map[string]int)
	}
	if s.SeenOperator == nil {
		s.SeenOperator = make(map[string]int)
	}
	if s.SeenCountry == nil {
		s.SeenCountry = make(map[string]int)
	}
}

// RecordType increments the seen counter for the given aircraft make
// and total sighted types.
func (s *State) RecordType(makeName string) int {
	if s.SeenType == nil {
		s.SeenType = make(map[string]int)
	}
	s.SeenType[makeName]++
	s.SightedTypes++
	return s.SeenType[makeName]
}

// RecordOperator increments the seen counter for the given operator
// and total sighted operators.
func (s *State) RecordOperator(operator string) int {
	if s.SeenOperator == nil {
		s.SeenOperator = make(map[string]int)
	}
	s.SeenOperator[operator]++
	s.SightedOperators++
	return s.SeenOperator[operator]
}

// RecordCountry increments the seen counter for the given country
// and total sighted countries.
func (s *State) RecordCountry(country string) int {
	if s.SeenCountry == nil {
		s.SeenCountry = make(map[string]int)
	}
	s.SeenCountry[country]++
	s.SightedCountries++
	return s.SeenCountry[country]
}
