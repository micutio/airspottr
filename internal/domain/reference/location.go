package reference

import (
	"fmt"
	"math"
)

const defaultLocationPrecision = 4

// LocationID is the canonical identity for a spotting location.
// It quantizes lat/lon to a stable grid so histories can be kept separate per spot.
type LocationID struct {
	Latitude  float64
	Longitude float64
	Precision int
}

// NewLocationID creates a canonical location identity for the given coordinates.
func NewLocationID(latitude, longitude float64) LocationID {
	return LocationID{
		Latitude:  roundToPrecision(latitude, defaultLocationPrecision),
		Longitude: roundToPrecision(longitude, defaultLocationPrecision),
		Precision: defaultLocationPrecision,
	}
}

// Key returns the canonical string key for the location.
func (id LocationID) Key() string {
	return fmt.Sprintf("%.4f,%.4f", id.Latitude, id.Longitude)
}

// LocationKey creates a location key for a coordinate pair.
func LocationKey(latitude, longitude float64) string {
	return NewLocationID(latitude, longitude).Key()
}

func roundToPrecision(value float64, precision int) float64 {
	factor := math.Pow10(precision)
	return math.Round(value*factor) / factor
}
