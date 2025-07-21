// Inspired by https://github.com/LucaTheHacker/go-haversine

package main

import "math"

// Constants

const (
	earthRadiusKilometers    = 6371 // Radius of Earth in kilometers
	earthRadiusMiles         = 3958 // Radius of Earth in miles
	earthRadiusNauticalMiles = 3443 // Radius of Earth in miles
)

// Conversion function

func degreesToRadian(d float64) float64 {
	return d * math.Pi / 180
}

// Coordinate type

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func (c Coordinates) toRadians() Coordinates {
	return Coordinates{
		Latitude:  degreesToRadian(c.Latitude),
		Longitude: degreesToRadian(c.Longitude),
	}
}

// NewCoordinates returns a Coordinates struct based on parameters passed
func NewCoordinates(latitude, longitude float64) Coordinates {
	return Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	}
}

// distance type

type DistanceStruct struct {
	C float64 // Must be multiplied to obtain distance. Public in order to allow unexpected calculations.
}

func NewDistanceStruct(distance float64) DistanceStruct {
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

// Distance calculates distance using the haversine formula
func Distance(p, q Coordinates) DistanceStruct {
	from := p.toRadians()
	to := q.toRadians()

	deltaLat := to.Latitude - from.Latitude
	deltaLon := to.Longitude - from.Longitude

	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Cos(from.Latitude)*math.Cos(to.Latitude)*math.Pow(math.Sin(deltaLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return DistanceStruct{
		C: c,
	}
}
