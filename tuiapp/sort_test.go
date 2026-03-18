package tuiapp

import (
	"testing"

	"github.com/micutio/airspottr/internal"
	"github.com/micutio/airspottr/internal/dash"
)

func TestFilteredSortedAircraftByDistance(t *testing.T) {
	t.Parallel()
	db := &internal.Dashboard{
		CurrentAircraft: []internal.AircraftRecord{
			{Hex: "a", CachedDist: 100, Flight: "B"},
			{Hex: "b", CachedDist: 10, Flight: "A"},
		},
		IcaoToAircraft: map[string]dash.IcaoAircraft{},
	}
	out := filteredSortedAircraft(db, 0, false) // DST asc
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
	if out[0].Hex != "b" || out[1].Hex != "a" {
		t.Errorf("order %+v", out)
	}
	out = filteredSortedAircraft(db, 0, true)
	if out[0].Hex != "a" {
		t.Errorf("desc first want a got %s", out[0].Hex)
	}
}

func TestSortedPropertyCounts(t *testing.T) {
	t.Parallel()
	m := map[string]int{"z": 1, "a": 2, "b": 2}
	byCount := sortedPropertyCounts(m, false, false)
	if byCount[0].Count != 1 || byCount[len(byCount)-1].Count != 2 {
		t.Errorf("by count order %+v", byCount)
	}
	byName := sortedPropertyCounts(m, true, false)
	if byName[0].Property != "a" || byName[len(byName)-1].Property != "z" {
		t.Errorf("by name order %+v", byName)
	}
}
