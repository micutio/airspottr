package reference

import (
	"errors"
	"fmt"
	"math"
)

// Inspired by https://github.com/LucaTheHacker/go-haversine

// Constants

const (
	earthRadiusKilometers    float64 = 6371 // Radius of Earth in kilometers
	earthRadiusMiles         float64 = 3958 // Radius of Earth in miles
	earthRadiusNauticalMiles float64 = 3443 // Radius of Earth in miles
	piHalf                   float64 = math.Pi / 180
)

// Coordinate type

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// NewCoordinates returns a coordinates struct based on parameters passed.
func NewCoordinates(latitude, longitude float64) Coordinates {
	return Coordinates{Latitude: latitude, Longitude: longitude}
}

var (
	ErrLatitudeOutOfRange  = errors.New("latitude out of range")
	ErrLongitudeOutOfRange = errors.New("longitude out of range")
)

// ParseCoordinates validates and returns a coordinate pair.
func ParseCoordinates(latitude, longitude float64) (Coordinates, error) {
	coords := Coordinates{Latitude: latitude, Longitude: longitude}
	if err := coords.Validate(); err != nil {
		return Coordinates{}, err
	}
	return coords, nil
}

// MustCoordinates validates a coordinate pair and panics on invalid input.
func MustCoordinates(latitude, longitude float64) Coordinates {
	coords := NewCoordinates(latitude, longitude)
	if err := coords.Validate(); err != nil {
		panic(err)
	}
	return coords
}

// Validate ensures the coordinate falls within the valid lat/lon ranges.
func (c Coordinates) Validate() error {
	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf("%w: got %v", ErrLatitudeOutOfRange, c.Latitude)
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf("%w: got %v", ErrLongitudeOutOfRange, c.Longitude)
	}
	return nil
}

func (c Coordinates) toRadians() Coordinates {
	return Coordinates{
		Latitude:  degreesToRadian(c.Latitude),
		Longitude: degreesToRadian(c.Longitude),
	}
}

// Conversion function

func degreesToRadian(d float64) float64 {
	return d * piHalf
}

// distance type

type DistanceStruct struct {
	C float64 // Must be multiplied to obtain distance. Public in order to allow unexpected calculations.
}

func newDistanceStruct(distance float64) DistanceStruct {
	return DistanceStruct{C: distance}
}

func (d DistanceStruct) Kilometers() float64 {
	return d.C * earthRadiusKilometers
}

func (d DistanceStruct) Miles() float64 {
	return d.C * earthRadiusMiles
}

func (d DistanceStruct) NauticalMiles() float64 {
	return d.C * earthRadiusNauticalMiles
}

// Distance calculates distance using the haversine formula.
//
//nolint:mnd // readability of mathmatic formula
func Distance(p, q Coordinates) DistanceStruct {
	fromPos := p.toRadians()
	toPos := q.toRadians()

	deltaLat := toPos.Latitude - fromPos.Latitude
	deltaLon := toPos.Longitude - fromPos.Longitude

	a := math.Pow(math.Sin(deltaLat/2), 2) +
		math.Cos(fromPos.Latitude)*
			math.Cos(toPos.Latitude)*
			math.Pow(math.Sin(deltaLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return newDistanceStruct(c)
}
