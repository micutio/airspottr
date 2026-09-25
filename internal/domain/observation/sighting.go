package observation

import (
	"fmt"
	"math"
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

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
	Rarities     RarityFlag            `json:"rarities"`
	LastSeen     time.Time             `json:"last_seen"`
	LastFlightNo string                `json:"last_flight_no"`
	Registration string                `json:"registration"`
	Latitude     float64               `json:"latitude"`
	Longitude    float64               `json:"longitude"`
	Direction    ref.Direction         `json:"direction"`
	Distance     float64               `json:"distance"`    // distance of the aircraft to our location [m]
	TypeShort    string                `json:"type_short"`  // short type name, directly from the record
	TypeDesc     string                `json:"type_desc"`   // typeDesc is the full name of the aircraft type
	Operator     string                `json:"operator"`    // operator can be either airline or military organisation
	Country      string                `json:"country"`     // country of registration
	Info         string                `json:"info"`        // info contains the aircraft information represented as string
	Flightroute  ref.FlightrouteRecord `json:"flightroute"` // flightroute contains airline, origin and destination
	LastRecord   AircraftRecord        `json:"last_record"`
}

// NewSighting builds a default sighting for an aircraft not seen before.
func NewSighting(observerLat, observerLon float64, aircraft AircraftRecord, now time.Time) AircraftSighting {
	lastSeenOffset := time.Duration(aircraft.Seen) * time.Second
	return AircraftSighting{
		Rarities:     NoRarity,
		LastSeen:     now.Add(-lastSeenOffset),
		LastFlightNo: FlightUnknown,
		Registration: aircraft.Registration,
		Latitude:     aircraft.Lat,
		Longitude:    aircraft.Lon,
		Direction:    ref.GetDirection(observerLat, observerLon, aircraft.Lat, aircraft.Lon),
		Distance:     math.MaxInt,
		TypeShort:    "",
		TypeDesc:     TypeUnknown,
		Operator:     OperatorUnknown,
		Country:      ref.CountryUnknown,
		Info:         "",
		Flightroute:  ref.FlightrouteRecord{}, //nolint:exhaustruct_v5 // using default values
		LastRecord:   aircraft,
	}
}

// ByFlight implements the comparator interface and allows sorting a list of aircraft records
// by Flight.
type ByFlight []AircraftSighting

func (a ByFlight) Len() int           { return len(a) }
func (a ByFlight) Less(i, j int) bool { return a[i].LastFlightNo < a[j].LastFlightNo }
func (a ByFlight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ByDistance implements the comparator interface and allows sorting a list of aircraft records.
// by distance to a given lon,lat coordinate.
type ByDistance []AircraftSighting

func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// SightingToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func (ac *AircraftSighting) SightingToString() string {
	flight := ac.LastFlightNo
	altitude := ac.LastRecord.GetAltitudeAsStr()

	return fmt.Sprintf("FNO %s DST %4.0f km ALT %s SPD %3.0f HDG %3.0f TID %s (%s)",
		flight,
		ac.Distance,
		altitude,
		ac.LastRecord.GroundSpeed,
		ac.LastRecord.NavHeading,
		ac.TypeDesc,
		ac.Registration)
}
