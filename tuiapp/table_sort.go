package tuiapp

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/micutio/airspottr/internal"
)

const aircraftColumnCount = 8

// Base titles (sort arrows appended in applyAircraftSortHeaders).
var aircraftColumnTitles = [aircraftColumnCount]string{
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

var rarityValueColumnTitle = [rarityTableCount]string{"Type", "operator", "country"}

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

func routeFor(db *internal.Dashboard, ac *internal.AircraftRecord) *internal.FlightRouteRecord {
	r, ok := db.CachedFlightRoutes[ac.GetFlightNoAsStr()]
	if !ok {
		return internal.GetDefaultFlightrouteRecord()
	}
	return r
}

func altitudeSortKey(ac *internal.AircraftRecord) float64 {
	if n, ok := ac.AltBaro.(float64); ok {
		return n
	}
	if s, ok := ac.AltBaro.(string); ok && strings.EqualFold(s, "ground") {
		return -1000
	}
	return 1e9
}

// compareAircraftAscending reports whether a should sort before b (ascending).
func compareAircraftAscending(a, b *internal.AircraftRecord, col int, db *internal.Dashboard) bool {
	ra, rb := routeFor(db, a), routeFor(db, b)
	switch col {
	case 0: // DST
		if a.CachedDist != b.CachedDist {
			return a.CachedDist < b.CachedDist
		}
	case 1: // FNO
		sa, sb := a.GetFlightNoAsStr(), b.GetFlightNoAsStr()
		if sa != sb {
			return sa < sb
		}
	case 2: // TID
		ta := db.IcaoToAircraft[a.IcaoType].Make
		tb := db.IcaoToAircraft[b.IcaoType].Make
		if ta != tb {
			return ta < tb
		}
	case 3: // DEP
		da, dbi := ra.Origin.IataCode, rb.Origin.IataCode
		if da != dbi {
			return da < dbi
		}
	case 4: // ARR
		da, dbi := ra.Destination.IataCode, rb.Destination.IataCode
		if da != dbi {
			return da < dbi
		}
	case 5: // ALT
		ka, kb := altitudeSortKey(a), altitudeSortKey(b)
		if ka != kb {
			return ka < kb
		}
	case 6: // SPD
		if a.GroundSpeed != b.GroundSpeed {
			return a.GroundSpeed < b.GroundSpeed
		}
	case 7: // HDG
		if a.NavHeading != b.NavHeading {
			return a.NavHeading < b.NavHeading
		}
	}
	return a.Hex < b.Hex
}

func filteredSortedAircraft(db *internal.Dashboard, sortCol int, desc bool) []internal.AircraftRecord {
	var rows []internal.AircraftRecord
	for _, ac := range db.CurrentAircraft {
		aircraftType := db.IcaoToAircraft[ac.IcaoType].Make
		if ac.GetFlightNoAsStr() == "" && aircraftType == "" {
			continue
		}
		rows = append(rows, ac)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		less := compareAircraftAscending(&rows[i], &rows[j], sortCol, db)
		if desc {
			return !less
		}
		return less
	})
	return rows
}

func compareRarityAscending(a, b internal.PropertyCountTuple, byProperty bool) bool {
	if byProperty {
		if a.Property != b.Property {
			return a.Property < b.Property
		}
		return a.Count < b.Count
	}
	if a.Count != b.Count {
		return a.Count < b.Count
	}
	return a.Property < b.Property
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

func buildAircraftRows(db *internal.Dashboard, records []internal.AircraftRecord) []table.Row {
	rows := make([]table.Row, 0, len(records))
	for i := range records {
		ac := &records[i]
		route := routeFor(db, ac)
		rows = append(rows, aircraftToRow(ac, route))
	}
	return rows
}
