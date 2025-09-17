package dash

import (
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
	return Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	}
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
