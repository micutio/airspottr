package observation

import (
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// TODO: Remove dependency on Math, if possible.
// TODO: Make flightroute record value instead of pointer.
// TODO: Avoid duplicating data by adding field: lastStatus AicraftRecord `json:"lastStatus"`.

// AircraftSighting represents signals received from an aircraft in Flight.
// This includes aircraft on the ground as long as a valid Flight number is
// being broadcast.
// All signals received from an aircraft during the same Flight (number) will
// be treated as one singular sighting.
// Once the aircraft lands and departs again with a new Flight number, this
// will be considered a _new_ sighting.
// Since individual ADS-B messages may contain incomplete data, we are
// continuously updating the AircraftSighting struct fields with data received
// from an ongoing Flight.
type AircraftSighting struct {
	LastSeen     time.Time              `json:"last_seen"`
	LastFlightNo string                 `json:"last_flight_no"`
	Registration string                 `json:"registration"`
	Latitude     float64                `json:"latitude"`
	Longitude    float64                `json:"longitude"`
	Direction    Direction              `json:"direction"`
	Distance     float64                `json:"distance"`    // distance of the aircraft to our location [m]
	TypeShort    string                 `json:"type_short"`  // short type name, directly from the record
	TypeDesc     string                 `json:"type_desc"`   // typeDesc is the full name of the aircraft type
	Operator     string                 `json:"operator"`    // operator can be either airline or military organization
	Country      string                 `json:"country"`     // country of registration
	Info         string                 `json:"info"`        // info contains the aircraft information represented as string
	Flightroute  *ref.FlightRouteRecord `json:"flightroute"` // flightroute contains airline, origin and destination
}

// RareSighting combines an aircraft sighting with a rarity flag.
type RareSighting struct {
	Rarities RarityFlag
	Sighting *AircraftSighting
}
