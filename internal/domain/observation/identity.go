package observation

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const aircraftHexLength = 6

var (
	ErrEmptyAircraftHex    = errors.New("empty aircraft hex")
	ErrInvalidAircraftHex  = errors.New("invalid aircraft hex")
	ErrEmptyFlightCallsign = errors.New("empty flight callsign")
)

// AircraftHex is a validated ICAO 24-bit aircraft identifier.
type AircraftHex string

// NewAircraftHex returns a normalized aircraft hex identifier.
func NewAircraftHex(raw string) (AircraftHex, error) {
	normalized := strings.TrimSpace(strings.ToUpper(raw))
	if normalized == "" {
		return "", ErrEmptyAircraftHex
	}
	if len(normalized) != aircraftHexLength {
		return "", fmt.Errorf("%w: %q expected %d characters", ErrInvalidAircraftHex, raw, aircraftHexLength)
	}
	for _, r := range normalized {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", fmt.Errorf("%w: %q", ErrInvalidAircraftHex, raw)
		}
	}
	return AircraftHex(normalized), nil
}

// String returns the underlying raw identifier string.
func (h AircraftHex) String() string {
	return string(h)
}

// FlightCallsign is a validated flight identifier.
type FlightCallsign string

// NewFlightCallsign returns a normalized flight callsign.
func NewFlightCallsign(raw string) (FlightCallsign, error) {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return "", ErrEmptyFlightCallsign
	}
	return FlightCallsign(strings.ToUpper(normalized)), nil
}

// String returns the underlying raw identifier string.
func (c FlightCallsign) String() string {
	return string(c)
}
