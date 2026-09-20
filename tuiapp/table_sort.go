package tuiapp

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/micutio/airspottr/internal"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	"github.com/micutio/airspottr/internal/domain/repositories"
)

const (
	aircraftColumnCount = 8
	groundAltitudeZone  = -1000
	infinity            = 1e9
)

// Base titles (sort arrows appended in applyAircraftSortHeaders).
var aircraftColumnTitles = [aircraftColumnCount]string{ //nolint:gochecknoglobals // TODO: Fix
	"DST", "FNO", "TID", "DEP", "ARR", "ALT", "SPD", "HDG",
}

func arrowForSort(desc bool) string {
	if desc {
		return "▼"
	}
	return "▲"
}

func applyAircraftSortHeaders(tbl *table.Model, sortCol int, desc bool) {
	old := tbl.Columns()
	cols := make([]table.Column, len(old))
	copy(cols, old)
	for i := range cols {
		title := aircraftColumnTitles[i]
		if i == sortCol {
			title += arrowForSort(desc)
		}
		cols[i].Title = title
	}
	tbl.SetColumns(cols)
}

//nolint:gochecknoglobals // TODO: Fix
var rarityValueColumnTitle = [rarityTableCount]string{"Type", "Operator", "Country"}

func applyRaritySortHeaders(tbl *table.Model, rarityIdx, sortCol int, desc bool) {
	old := tbl.Columns()
	cols := make([]table.Column, len(old))
	copy(cols, old)
	cols[0].Title = "Count"
	if sortCol == 0 {
		cols[0].Title = "Count" + arrowForSort(desc)
	}
	second := rarityValueColumnTitle[rarityIdx]
	if sortCol == 1 {
		second += arrowForSort(desc)
	}
	cols[1].Title = second
	tbl.SetColumns(cols)
}

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
	typeRepo repositories.AircraftTypeRepo,
) bool {
	dstCol := 0
	fnoCol := 1
	tidCol := 2
	depCol := 3
	arrCol := 4
	altCol := 5
	spdCol := 6
	hdgCol := 7
	routeA, routeB := sightingA.Flightroute, sightingB.Flightroute
	switch col {
	case dstCol: // DST
		if sightingA.Distance != sightingB.Distance {
			return sightingA.Distance < sightingB.Distance
		}
	case fnoCol: // FNO
		sa, sb := sightingA.LastFlightNo, sightingB.LastFlightNo
		if sa != sb {
			return sa < sb
		}
	case tidCol: // TID
		// TODO: Can we cache AircraftType somewhere?
		typeA, taExists := typeRepo.GetAircraftType(sightingA.LastRecord.IcaoType)
		if !taExists {
			return false
		}
		typeB, tbExists := typeRepo.GetAircraftType(sightingB.LastRecord.IcaoType)
		if !tbExists {
			return false
		}
		if typeA.Make != typeB.Make {
			return typeA.Make < typeB.Make
		}
	case depCol: // DEP
		da, dbi := routeA.Origin.IataCode, routeB.Origin.IataCode
		if da != dbi {
			return da < dbi
		}
	case arrCol: // ARR
		da, dbi := routeA.Destination.IataCode, routeB.Destination.IataCode
		if da != dbi {
			return da < dbi
		}
	case altCol: // ALT
		ka, kb := altitudeSortKey(sightingA), altitudeSortKey(sightingB)
		if ka != kb {
			return ka < kb
		}
	case spdCol: // SPD
		if sightingA.LastRecord.GroundSpeed != sightingB.LastRecord.GroundSpeed {
			return sightingA.LastRecord.GroundSpeed < sightingB.LastRecord.GroundSpeed
		}
	case hdgCol: // HDG
		if sightingA.LastRecord.NavHeading != sightingB.LastRecord.NavHeading {
			return sightingA.LastRecord.NavHeading < sightingB.LastRecord.NavHeading
		}
	}
	return sightingA.LastRecord.Hex < sightingB.LastRecord.Hex
}

func filteredSortedSightings(
	currentSightings []obs.AircraftSighting,
	typeRepo repositories.AircraftTypeRepo,
	sortCol int,
	desc bool,
) []obs.AircraftSighting {
	var rows []obs.AircraftSighting
	for _, ac := range currentSightings {
		aircraftType, atExists := typeRepo.GetAircraftType(ac.LastRecord.IcaoType)
		if !atExists || (ac.LastFlightNo == "" && aircraftType.Make == "") {
			continue
		}
		rows = append(rows, ac)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		less := compareSightingsAscending(rows[i], rows[j], sortCol, typeRepo)
		if desc {
			return !less
		}
		return less
	})
	return rows
}

func compareRarityAscending(propertyA, propertyB internal.PropertyCountTuple, byProperty bool) bool {
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

func sortedPropertyCounts(m map[string]int, byProperty, desc bool) []internal.PropertyCountTuple {
	tuples := make([]internal.PropertyCountTuple, 0, len(m))
	for k, v := range m {
		tuples = append(tuples, internal.PropertyCountTuple{Property: k, Count: v})
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

// cycleSortColumn steps sort column (focused table). dir +1 or -1.
func (m *model) cycleSortColumn(dir int) {
	if m.uiState == mainPage {
		m.aircraftSortCol = (m.aircraftSortCol + dir + aircraftColumnCount) % aircraftColumnCount
	} else {
		idx := m.selectedRarityIdx
		//nolint:mnd // Don't care about this magic number.
		m.raritySortCol[idx] = (m.raritySortCol[idx] + dir + 2) % 2
	}
	m.updateAllTables()
}

func (m *model) toggleSortDirection() {
	if m.uiState == mainPage {
		m.aircraftSortDesc = !m.aircraftSortDesc
	} else {
		idx := m.selectedRarityIdx
		m.raritySortDesc[idx] = !m.raritySortDesc[idx]
	}
	m.updateAllTables()
}

func buildAircraftRows(records []obs.AircraftSighting) []table.Row {
	rows := make([]table.Row, 0, len(records))
	for i := range records {
		ac := records[i]
		route := ac.Flightroute
		rows = append(rows, aircraftToRow(ac, route))
	}
	return rows
}
