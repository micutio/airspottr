package interfaces

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// AircraftDataSource represents a source of fresh aircraft observations.
type AircraftDataSource interface {
	RequestAircraft() []obs.AircraftRecord
}

// FlightRouteDataSource represents a source of flight routes keyed by callsign.
type FlightRouteDataSource interface {
	GetFlightroutes(callsigns []string) map[string]ref.FlightrouteRecord
	GetPendingCallsigns() []string
	RestorePendingCallsigns(pendingCallsigns []string)
}
