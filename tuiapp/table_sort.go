package tuiapp

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	"github.com/micutio/airspottr/internal/domain/reference"
)

const (
	dstCol = iota
	fnoCol
	tidCol
	depCol
	arrCol
	altCol
	spdCol
	hdgCol
)

const (
	groundAltitudeZone = -1000
	infinity           = 1e9
)

func altitudeSortKey(aircraftSighting obs.AircraftSighting) float64 {
	if n, ok := aircraftSighting.LastRecord.AltBaro.(float64); ok {
		return n
	}
	if s, ok := aircraftSighting.LastRecord.AltBaro.(string); ok && strings.EqualFold(s, "ground") {
		return groundAltitudeZone
	}
	return infinity
}

// compareSightingsAscending reports whether a should sort before b (ascending).
//
//nolint:gocognit
func compareSightingsAscending(
	sightingA, sightingB obs.AircraftSighting,
	col int,
) bool {
	routeA, routeB := sightingA.Flightroute, sightingB.Flightroute
	switch col {
	case dstCol:
		if sightingA.Distance != sightingB.Distance {
			return sightingA.Distance < sightingB.Distance
		}
	case fnoCol:
		if sightingA.LastFlightNo != sightingB.LastFlightNo {
			return sightingA.LastFlightNo < sightingB.LastFlightNo
		}
	case tidCol:
		if sightingA.TypeDesc != sightingB.TypeDesc {
			return sightingA.TypeDesc < sightingB.TypeDesc
		}
	case depCol:
		if routeA.Origin.IataCode != routeB.Origin.IataCode {
			return routeA.Origin.IataCode < routeB.Origin.IataCode
		}
	case arrCol:
		if routeA.Destination.IataCode != routeB.Destination.IataCode {
			return routeA.Destination.IataCode < routeB.Destination.IataCode
		}
	case altCol:
		ka, kb := altitudeSortKey(sightingA), altitudeSortKey(sightingB)
		if ka != kb {
			return ka < kb
		}
	case spdCol:
		if sightingA.LastRecord.GroundSpeed != sightingB.LastRecord.GroundSpeed {
			return sightingA.LastRecord.GroundSpeed < sightingB.LastRecord.GroundSpeed
		}
	case hdgCol:
		if sightingA.LastRecord.NavHeading != sightingB.LastRecord.NavHeading {
			return sightingA.LastRecord.NavHeading < sightingB.LastRecord.NavHeading
		}
	}
	return sightingA.LastRecord.Hex < sightingB.LastRecord.Hex
}

func filteredSortedSightings(
	currentSightings []obs.AircraftSighting,
	sortCol int,
	desc bool,
) []obs.AircraftSighting {
	var rows []obs.AircraftSighting
	for _, ac := range currentSightings {
		if ac.TypeDesc == obs.TypeUnknown || (ac.LastFlightNo == "" && ac.TypeDesc == "") {
			continue
		}
		rows = append(rows, ac)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		less := compareSightingsAscending(rows[i], rows[j], sortCol)
		if desc {
			return !less
		}
		return less
	})
	return rows
}

func compareRarityAscending(propertyA, propertyB reference.PropertyCountTuple, byProperty bool) bool {
	if byProperty {
		if propertyA.Property != propertyB.Property {
			return propertyA.Property < propertyB.Property
		}
		return propertyA.Count < propertyB.Count
	}
	if propertyA.Count != propertyB.Count {
		return propertyA.Count < propertyB.Count
	}
	return propertyA.Property < propertyB.Property
}

func sortedPropertyCounts(m map[string]int, byProperty, desc bool) []reference.PropertyCountTuple {
	tuples := make([]reference.PropertyCountTuple, 0, len(m))
	for k, v := range m {
		tuples = append(tuples, reference.PropertyCountTuple{Property: k, Count: v})
	}
	sort.SliceStable(tuples, func(i, j int) bool {
		less := compareRarityAscending(tuples[i], tuples[j], byProperty)
		if desc {
			return !less
		}
		return less
	})
	return tuples
}

func buildAircraftRows(records []obs.AircraftSighting) []table.Row {
	rows := make([]table.Row, 0, len(records))
	for i := range records {
		ac := records[i]
		rows = append(rows, aircraftToRow(ac, ac.Flightroute))
	}
	return rows
}
