//nolint:exhaustruct_v5
package tuiapp

import (
	"testing"

	"github.com/micutio/airspottr/internal/application"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

type TypeRepoMock struct{}

// GetAircraftType implements the AircraftTypeRepo interface of
// the same name.
func (mock *TypeRepoMock) GetAircraftType(icaoCode string) (ref.IcaoAircraftSpec, bool) {
	return ref.IcaoAircraftSpec{
		Class:  icaoCode,
		Engine: icaoCode,
		Make:   icaoCode,
	}, true
}

func TestFilteredSortedSightingsByDistance(t *testing.T) {
	t.Parallel()
	dashboard := &application.Dashboard{
		CurrentSightings: []obs.AircraftSighting{
			{
				LastRecord: obs.AircraftRecord{
					Hex:    "a",
					Flight: "B",
				},
				Distance:     100,
				LastFlightNo: "B",
			},
			{
				LastRecord: obs.AircraftRecord{
					Hex:    "b",
					Flight: "A",
				},
				Distance:     10,
				LastFlightNo: "A",
			},
		},
	}
	typeRepo := TypeRepoMock{}

	out := filteredSortedSightings(dashboard, &typeRepo, 0, false) // DST asc
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
	if out[0].LastRecord.Hex != "b" || out[1].LastRecord.Hex != "a" {
		t.Errorf("order %+v", out)
	}
	out = filteredSortedSightings(dashboard, &typeRepo, 0, true)
	if out[0].LastRecord.Hex != "a" {
		t.Errorf("desc first want a got %s", out[0].LastRecord.Hex)
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
