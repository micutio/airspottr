package repositories

import ref "github.com/micutio/airspottr/internal/domain/reference"

// FlightrouteRepository is an abstract source of flightroutes.
// TODO: Maybe return an error if flightroute request fails.
type FlightrouteRepository interface {
	// TODO: Dedicated type for callsigns, maybe `VerifiedCallsigns`.

	// RequestFlightroutesForCallsigns collects known flight routes for the given callsigns.
	// a.k.a. flight numbers.
	RequestFlightroutesForCallsigns(callsigns []string) []ref.FlightRouteRecord

	// The following methods should be used only for saving and loading application state in-between
	// program shutdowns and starts.
	// TODO: Decide whether these methods should be moved elsewhere.

	// GetPendingCallsigns returns all currently pending callsigns to allow persisting the state.
	GetPendingCallsigns() []string

	// RestorePendingCallsigns restores previously exported callsigns to the internal state of the
	// FlightrouteRepository
	RestorePendingCallsigns(pendingCallsigns []string)
}

// TODO:
//    - Reference repositories for looking up ICAO/operator/country data
//    - StateRepository for saving and loading state
//    - NotificationSender for emitting summaries and rarity notifications
