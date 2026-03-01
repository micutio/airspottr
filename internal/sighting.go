package internal

import (
	"math"
	"time"
)

const (
	dirUnknown string = "unknown"
	dirN       string = "north"
	dirNbE     string = "north by east"
	dirNNE     string = "north-northeast"
	dirNEbN    string = "northeast by north"
	dirNE      string = "northeast"
	dirNEbE    string = "northeast by east"
	dirENE     string = "east-northeast"
	dirEbN     string = "east by north"
	dirE       string = "east"
	dirEbS     string = "east by south"
	dirESE     string = "east-southeast"
	dirSEbE    string = "southeast by east"
	dirSE      string = "southeast"
	dirSEbS    string = "southeast by south"
	dirSSE     string = "south-southeast"
	dirSbE     string = "south by east"
	dirS       string = "south"
	dirSbW     string = "south by west"
	dirSSW     string = "south-southwest"
	dirSWbS    string = "southwest by south"
	dirSW      string = "southwest"
	dirSWbW    string = "southwest by west"
	dirWSW     string = "west-southwest"
	dirWbS     string = "west by south"
	dirW       string = "west"
	dirWbN     string = "west by north"
	dirWNW     string = "west-northwest"
	dirNWbW    string = "northwest by west"
	dirNW      string = "northwest"
	dirNWbN    string = "northwest by north"
	dirNNW     string = "north-northwest"
	dirNbW     string = "north by west"
)

var directions = []string{ //nolint: gochecknoglobals // Can't be bothered to fix for now
	dirN,
	dirNbE,
	dirNNE,
	dirNEbN,
	dirNE,
	dirNEbE,
	dirENE,
	dirEbN,
	dirE,
	dirEbS,
	dirESE,
	dirSEbE,
	dirSE,
	dirSEbS,
	dirSSE,
	dirSbE,
	dirS,
	dirSbW,
	dirSSW,
	dirSWbS,
	dirSW,
	dirSWbW,
	dirWSW,
	dirWbS,
	dirW,
	dirWbN,
	dirWNW,
	dirNWbW,
	dirNW,
	dirNWbN,
	dirNNW,
	dirNbW,
}

// AircraftSighting represents signals received from an aircraft in flight.
// This includes aircraft on the ground as long as a valid flight number is
// being broadcast.
// All signals received from an aircraft during the same flight (number) will
// be treated as one singular sighting.
// Once the aircraft lands and departs again with a new flight number, this
// will be considered a _new_ sighting.
// Since individual ADS-B messages may contain incomplete data, we are
// continuously updating the AircraftSighting struct fields with data received
// from an ongoing flight.
type AircraftSighting struct {
	lastSeen     time.Time
	lastFlightNo string
	registration string
	latitude     float64
	longitude    float64
	direction    string
	distance     float64 // distance is the current distance of the aircraft to our location [m]
	typeShort    string  // typeShort is a shorter name of the type, directly from the record
	typeDesc     string  // typeDesc is the full name of the aircraft type
	operator     string  // operator can be either airline or military organisation
	country      string  // country of registration
	info         string  // info contains the aircraft information represented as string
	// TODO: Think about adding a short type description, for when the _desc_ field in Aircraft is set.
}

// TODO: Try to remove *AircraftRecord from this by caching required fields in the sighting instead!

// RareSighting combines an aircraft sighting with a rarity flag.
type RareSighting struct {
	Rarities RarityFlag
	Sighting *AircraftSighting
}

func getDirection(originLat, originLon, destLat, destLon float64) string {
	// TODO: Get bearing from (lat, lon) towards sighting location
	bearing := calculateBearing(originLat, originLon, destLat, destLon)

	start := 5.625
	step := 11.25

	for i := range 32 {
		if bearing <= start {
			return directions[i]
		}

		start += step
	}
	return dirUnknown
}

// toRadians converts degrees to radians.
func toRadians(deg float64) float64 {
	return deg * math.Pi / 180.0 //nolint: mnd // readability
}

// toDegrees converts radians to degrees.
func toDegrees(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// calculateBearing calculates the initial bearing (forward azimuth) from point 1 to point 2.
func calculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert degrees to radians
	fLat := toRadians(lat1)
	fLong := toRadians(lon1)
	tLat := toRadians(lat2)
	tLong := toRadians(lon2)

	dLon := tLong - fLong

	y := math.Sin(dLon) * math.Cos(tLat)
	x := math.Cos(fLat)*math.Sin(tLat) - math.Sin(fLat)*math.Cos(tLat)*math.Cos(dLon)

	// Calculate bearing in radians
	brng := math.Atan2(y, x)

	// Convert bearing to degrees
	brngDeg := toDegrees(brng)

	// Normalize the bearing to a value between 0 and 360 degrees
	// The result from Atan2 ranges from -180 to +180
	normalizedBearing := math.Mod(brngDeg+360.0, 360.0) //nolint: mnd // readability

	return normalizedBearing
}
