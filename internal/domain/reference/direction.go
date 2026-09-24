package reference

import "math"

type Direction string

const (
	dirUnknown Direction = "unknown"
	dirN       Direction = "north"
	dirNbE     Direction = "north by east"
	dirNNE     Direction = "north-northeast"
	dirNEbN    Direction = "northeast by north"
	dirNE      Direction = "northeast"
	dirNEbE    Direction = "northeast by east"
	dirENE     Direction = "east-northeast"
	dirEbN     Direction = "east by north"
	dirE       Direction = "east"
	dirEbS     Direction = "east by south"
	dirESE     Direction = "east-southeast"
	dirSEbE    Direction = "southeast by east"
	dirSE      Direction = "southeast"
	dirSEbS    Direction = "southeast by south"
	dirSSE     Direction = "south-southeast"
	dirSbE     Direction = "south by east"
	dirS       Direction = "south"
	dirSbW     Direction = "south by west"
	dirSSW     Direction = "south-southwest"
	dirSWbS    Direction = "southwest by south"
	dirSW      Direction = "southwest"
	dirSWbW    Direction = "southwest by west"
	dirWSW     Direction = "west-southwest"
	dirWbS     Direction = "west by south"
	dirW       Direction = "west"
	dirWbN     Direction = "west by north"
	dirWNW     Direction = "west-northwest"
	dirNWbW    Direction = "northwest by west"
	dirNW      Direction = "northwest"
	dirNWbN    Direction = "northwest by north"
	dirNNW     Direction = "north-northwest"
	dirNbW     Direction = "north by west"
)

// directions contains the eight principal winds, eight half-winds and sixteen
// quarter-winds to combine into 32 directions total, sorted by angle from
// zero to 360.
var directions = []Direction{ //nolint: gochecknoglobals // Can't be bothered to fix for now
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

// GetDirection calculates the bearing between two given coordinates and
// converts it into a human-readable direction, e.g.: north north-west.
func GetDirection(originLat, originLon, destLat, destLon float64) Direction {
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
