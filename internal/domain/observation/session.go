package observation

import (
	"errors"
	"fmt"
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

var ErrObserverMismatch = errors.New("spotting session observer mismatch")

// SpottingSession is the aggregate boundary for a single observation session at a fixed location.
// It owns the session state and enforces coordinate-level invariants for the current observation run.
type SpottingSession struct {
	observer         ref.Coordinates
	sightings        map[string]AircraftSighting
	seenType         map[string]int
	seenOperator     map[string]int
	seenCountry      map[string]int
	sightedTypes     int
	sightedOperators int
	sightedCountries int
	fastest          AircraftRecord
	highest          AircraftRecord
	warmup           bool
}

func NewSpottingSession(observer ref.Coordinates) *SpottingSession {
	return &SpottingSession{
		observer:         observer,
		sightings:        make(map[string]AircraftSighting),
		seenType:         make(map[string]int),
		seenOperator:     make(map[string]int),
		seenCountry:      make(map[string]int),
		sightedTypes:     0,
		sightedOperators: 0,
		sightedCountries: 0,
		fastest:          AircraftRecord{}, //nolint:exhaustruct_v5 // zero-value aggregate snapshot
		highest:          AircraftRecord{}, //nolint:exhaustruct_v5 // zero-value aggregate snapshot
		warmup:           true,
	}
}

func (s *SpottingSession) Observer() ref.Coordinates {
	return s.observer
}

func (s *SpottingSession) SetWarmup(warmup bool) {
	s.warmup = warmup
}

func (s *SpottingSession) IsWarmup() bool {
	return s.warmup
}

func (s *SpottingSession) Snapshot() State {
	return State{
		Observer:         s.observer,
		Sightings:        cloneSightings(s.sightings),
		SeenType:         cloneCounter(s.seenType),
		SeenOperator:     cloneCounter(s.seenOperator),
		SeenCountry:      cloneCounter(s.seenCountry),
		SightedTypes:     s.sightedTypes,
		SightedOperators: s.sightedOperators,
		SightedCountries: s.sightedCountries,
		Fastest:          s.fastest,
		Highest:          s.highest,
	}
}

func (s *SpottingSession) Restore(state State) error {
	if state.Observer != s.observer {
		return fmt.Errorf("%w: got %v, want %v", ErrObserverMismatch, state.Observer, s.observer)
	}

	s.sightings = cloneSightings(state.Sightings)
	s.seenType = cloneCounter(state.SeenType)
	s.seenOperator = cloneCounter(state.SeenOperator)
	s.seenCountry = cloneCounter(state.SeenCountry)
	s.sightedTypes = state.SightedTypes
	s.sightedOperators = state.SightedOperators
	s.sightedCountries = state.SightedCountries
	s.fastest = state.Fastest
	s.highest = state.Highest
	return nil
}

func (s *SpottingSession) Observe(records []AircraftRecord, classifier Classifier, now time.Time) []AircraftSighting {
	state := s.Snapshot()
	currentSightings := EvaluateBatch(&state, records, classifier, now)
	if err := s.Restore(state); err != nil {
		panic(err)
	}
	return currentSightings
}

func cloneSightings(in map[string]AircraftSighting) map[string]AircraftSighting {
	if in == nil {
		return map[string]AircraftSighting{}
	}
	out := make(map[string]AircraftSighting, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneCounter(in map[string]int) map[string]int {
	if in == nil {
		return map[string]int{}
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
