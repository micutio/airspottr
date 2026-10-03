package observation

import (
	"fmt"
	"strings"
	"unicode"
)

// See https://www.adsbexchange.com/version-2-api-wip/
// for further explanations of the fields

const (
	// AltitudeUnknown is what we use for aircraft without a given altitude.
	AltitudeUnknown = "  n/a"
	// FlightUnknown is what we use for aircraft with missing Flight number.
	// Note: we're adding space at the end to have a length that is consistent with ICAO codes.
	FlightUnknown = "unknown "
	// FlightUnknownCode is a sentinel code we use for aircraft with missing Flight number.
	FlightUnknownCode = "n/a"
	// TypeUnknown is what we use for aircraft with a type that's either empty or can't be found.
	TypeUnknown = "unknown"
	// OperatorUnknown is what we use for aircraft with a type that's either empty or can't be found.
	OperatorUnknown = "unknown"
)

// AircraftResult is the domain-level result wrapper for an aircraft query.
// External data adapters map ADS-B payloads into this shape.
type AircraftResult struct {
	Now         float64
	ResultCount int
	Ptime       float64
	Aircraft    []AircraftRecord
}

// AircraftRecord is the lean domain representation of an observed aircraft.
// ADS-B transport types live in infrastructure/adsb and are converted here.
type AircraftRecord struct {
	Alert           int
	AltBaro         any
	AltGeom         int
	BaroRate        float64
	EmitterCategory string
	Emergency       string
	Flight          string
	GroundSpeed     float64
	Gva             float64
	Hex             string
	Lat             float64
	Lon             float64
	Messages        int
	Mlat            []string
	NacP            float64
	NacV            float64
	NavAltitudeMcp  int
	NavHeading      float64
	NavQNH          float64
	Nic             int
	NicBaro         int
	Registration    string
	RadiusOfCtn     float64
	Rssi            float64
	Sda             int
	Seen            float64
	SeenPos         float64
	Sil             int
	SilType         string
	Spi             int
	Squawk          string
	IcaoType        string
	Tisb            []string
	Track           float64
	Type            string
	Version         int
	GeomRate        float64
	DBFlags         int
	NavModes        []string
	TrueHeading     float64
	Ias             float64
	Mach            float64
	MagHeading      float64
	Oat             float64
	Roll            float64
	Tas             float64
	Tat             float32
	TrackRate       float64
	WindDirection   float64
	WindSpeed       float64
	GpsOkBefore     float64
	GpsOkLat        float64
	GpsOkLon        float64
	LastPosition    any
	RrLat           float64
	RrLon           float64
	CalcTrack       any
	NavAltitudeFMS  float64
	OwnOp           string
	Description     string
}

// GetAltitudeAsStr reads the altitude of an aircraft and returns it as a string.
// The altitude is stored either as a string 'ground' or as a float denoting the measured
// barometric altitude.
// If the latter is the case, the float will be formatted without any decimal places
// (unnecessary accuracy) and converted to string.
func (ac *AircraftRecord) GetAltitudeAsStr() string {
	if num, numOk := ac.AltBaro.(float64); numOk {
		return fmt.Sprintf("%5.0f", num)
	}

	if str, strOk := ac.AltBaro.(string); strOk {
		return str
	}

	return AltitudeUnknown
}

// GetFlightNoAsStr converts the Flight number to a string.
// Returns either the full Flight number or 'unknown ' if it was not transmitted.
// TODO: Move or copy this to sighting!
func (ac *AircraftRecord) GetFlightNoAsStr() string {
	if ac.Flight == "" {
		return FlightUnknown
	}
	callsign, err := NewFlightCallsign(ac.Flight)
	if err != nil {
		return FlightUnknown
	}
	return strings.TrimSpace(callsign.String())
}

// FlightNo returns the validated flight identifier as a domain value object.
func (ac *AircraftRecord) FlightNo() (FlightCallsign, error) {
	return NewFlightCallsign(ac.Flight)
}

// HexCode returns the validated aircraft hex as a domain value object.
func (ac *AircraftRecord) HexCode() (AircraftHex, error) {
	return NewAircraftHex(ac.Hex)
}

// GetFlightNoAsIcaoCode trims whitespaces and digits from the Flight number,
// resulting in the three-digit icao code for civilian flights and arbitrary length codes
// for military, government and private flights.
func (ac *AircraftRecord) GetFlightNoAsIcaoCode() string {
	if len(ac.Flight) == 0 {
		return FlightUnknownCode
	}
	callsign, err := NewFlightCallsign(ac.Flight)
	if err != nil {
		return FlightUnknownCode
	}
	return stripDigits(strings.TrimSpace(callsign.String()))
}

// GetRegistrationPrefix returns the prefix of the registration if it exists,
// otherwise it returns the entire registration.
func (ac *AircraftRecord) GetRegistrationPrefix() string {
	prefixEnd := strings.Index(ac.Registration, "-")

	if prefixEnd != -1 {
		return ac.Registration[0:prefixEnd]
	}
	return ac.Registration
}

func stripDigits(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return -1 // Remove the digit
		}
		return r // Keep the character
	}, str)
}

// AircraftToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func (ac *AircraftRecord) AircraftToString() string {
	flight := ac.GetFlightNoAsStr()
	altitude := ac.GetAltitudeAsStr()
	var aType string
	if ac.Description != "" {
		aType = ac.Description
	}

	return fmt.Sprintf("FNO %s km ALT %s SPD %3.0f HDG %3.0f TID %s (%s)",
		flight,
		altitude,
		ac.GroundSpeed,
		ac.NavHeading,
		aType,
		ac.Registration)
}
