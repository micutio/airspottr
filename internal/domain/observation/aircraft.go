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

// AircraftRecord is the lean domain representation of an observed aircraft.
// ADS-B transport types live in infrastructure/adsb and are converted here.
type AircraftRecord struct {
	Hex          string
	Flight       string
	Registration string
	IcaoType     string
	Description  string
	OwnOp        string
	Lat          float64
	Lon          float64
	AltBaro      any
	GroundSpeed  float64
	NavHeading   float64
	Seen         float64
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

// GetCallsignAsStr converts the callsign (a.k.a. flight number) to a string.
// Returns either
// - the full Flight number
// or
// - 'unknown ' if it was not transmitted.
func (ac *AircraftRecord) GetCallsignAsStr() string {
	normalized := strings.ToUpper(strings.TrimSpace(ac.Flight))
	if normalized == "" {
		return FlightUnknown
	}
	return normalized
}

// GetCallsignAsIcaoCode trims whitespaces and digits from the Flight number,
// resulting in the three-digit icao code for civilian flights and arbitrary
// length codes for military, government and private flights.
func (ac *AircraftRecord) GetCallsignAsIcaoCode() string {
	normalized := strings.ToUpper(strings.TrimSpace(ac.Flight))
	if normalized == "" {
		return FlightUnknownCode
	}
	return stripDigits(normalized)
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
	flight := ac.GetCallsignAsStr()
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
