package tuiapp

import "testing"

func TestUIStateValuesAreConsecutive(t *testing.T) {
	t.Parallel()
	if mainPage != 0 || aircraftDetails != 1 || globalStats != 2 {
		t.Errorf("mainPage=%d aircraftDetails=%d globalStats=%d want 0,1,2",
			mainPage, aircraftDetails, globalStats)
	}
}
