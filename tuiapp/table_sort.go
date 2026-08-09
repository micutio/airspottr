package tuiapp

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	obs "github.com/micutio/airspottr/domain/observation"
	"github.com/micutio/airspottr/internal"
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

func routeFor(db *internal.Dashboard, ac *obs.AircraftRecord) *internal.FlightRouteRecord {
	r, ok := db.CachedFlightRoutes[ac.GetFlightNoAsStr()]
	if !ok {
		return internal.GetDefaultFlightrouteRecord()
	}
	return r
}

func altitudeSortKey(aircraftRecord *obs.AircraftRecord) float64 {
	if n, ok := aircraftRecord.AltBaro.(float64); ok {
		return n
	}
	if s, ok := aircraftRecord.AltBaro.(string); ok && strings.EqualFold(s, "ground") {
		return groundAltitudeZone
	}
	return infinity
}

// compareAircraftAscending reports whether a should sort before b (ascending).
func compareAircraftAscending(
	recordA, recordB *obs.AircraftRecord,
	col int,
	dashboard *internal.Dashboard,
) bool {
	dstCol := 0
	fnoCol := 1
	tidCol := 2
	depCol := 3
	arrCol := 4
	altCol := 5
	spdCol := 6
	hdgCol := 7
	routeA, routeB := routeFor(dashboard, recordA), routeFor(dashboard, recordB)
	switch col {
	case dstCol: // DST
		if recordA.CachedDist != recordB.CachedDist {
			return recordA.CachedDist < recordB.CachedDist
		}
	case fnoCol: // FNO
		sa, sb := recordA.GetFlightNoAsStr(), recordB.GetFlightNoAsStr()
		if sa != sb {
			return sa < sb
		}
	case tidCol: // TID
		ta := dashboard.IcaoToAircraft[recordA.IcaoType].Make
		tb := dashboard.IcaoToAircraft[recordB.IcaoType].Make
		if ta != tb {
			return ta < tb
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
		ka, kb := altitudeSortKey(recordA), altitudeSortKey(recordB)
		if ka != kb {
			return ka < kb
		}
	case spdCol: // SPD
		if recordA.GroundSpeed != recordB.GroundSpeed {
			return recordA.GroundSpeed < recordB.GroundSpeed
		}
	case hdgCol: // HDG
		if recordA.NavHeading != recordB.NavHeading {
			return recordA.NavHeading < recordB.NavHeading
		}
	}
	return recordA.Hex < recordB.Hex
}

func filteredSortedAircraft(dashboard *internal.Dashboard, sortCol int, desc bool) []obs.AircraftRecord {
	var rows []obs.AircraftRecord
	for _, ac := range dashboard.CurrentAircraft {
		aircraftType := dashboard.IcaoToAircraft[ac.IcaoType].Make
		if ac.GetFlightNoAsStr() == "" && aircraftType == "" {
			continue
		}
		rows = append(rows, ac)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		less := compareAircraftAscending(&rows[i], &rows[j], sortCol, dashboard)
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

func buildAircraftRows(db *internal.Dashboard, records []obs.AircraftRecord) []table.Row {
	rows := make([]table.Row, 0, len(records))
	for i := range records {
		ac := &records[i]
		route := routeFor(db, ac)
		rows = append(rows, aircraftToRow(ac, route))
	}
	return rows
}
