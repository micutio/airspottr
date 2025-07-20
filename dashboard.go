package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type Dashboard struct {
	// Fields for tracking some statistics
	fastest *Aircraft
	highest *Aircraft
	// Data
	icaoToAircraft    map[string]IcaoAircraft
	milCodeToOperator map[string]string
}

func NewDashboard() Dashboard {
	return Dashboard{
		fastest:           nil,
		highest:           nil,
		icaoToAircraft:    GetIcaoToAircraftMap(),
		milCodeToOperator: GetMilCodeToOperatorMap(),
	}
}

// ProcessCivAircraftJson takes a json record in form of a byte array, transforms it into a list
// of aircraft and performs some processing thereafter.
func (db *Dashboard) ProcessCivAircraftJson(jsonBytes []byte) error {
	var data CivAircraftRecord
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return fmt.Errorf("no aircraft found")
	}

	db.processCivAircraftRecords(&data.Aircraft)
	return nil
}

func (db *Dashboard) processCivAircraftRecords(aircraft *[]Aircraft) {
	sort.Sort(ByFlight(*aircraft))
	for i := range len(*aircraft) {
		aircraft := (*aircraft)[i]
		db.checkHighest(&aircraft)
		db.checkFastest(&aircraft)
	}

}

func (db *Dashboard) ProcessMilAircraftJson(jsonBytes []byte) error {
	var data MilAircraftRecord
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal military aircraft JSON: %w", err)
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return fmt.Errorf("no military aircraft found")
	}

	db.processMilAircraftRecords(&(data.Aircraft))
	return nil
}

func (db *Dashboard) processMilAircraftRecords(aircraft *[]Aircraft) {
	thisPos := NewCoordinates(Lat, Lon)
	for i := range len(*aircraft) {
		ac := (*aircraft)[i]
		acPos := NewCoordinates(ac.Lat, ac.Lon)
		(*aircraft)[i].CachedDist = Distance(thisPos, acPos).Miles()
	}
	sort.Sort(ByDistance(*aircraft))

	fmt.Printf("[%s] Military aircraft in increasing distance from here:\n", time.Now().Format(TimeFmt))
	for i := range len(*aircraft) {
		ac := (*aircraft)[i]
		if ac.Lat == 0 && ac.Lon == 0 {
			continue
		}

		var altBaro string
		if num, ok := ac.AltBaro.(float64); ok {
			altBaro = fmt.Sprintf("%.0f", num)
		}
		if str, ok := ac.AltBaro.(string); ok {
			altBaro = str
		}
		aType := db.icaoToAircraft[ac.IcaoType].ModelCode

		fmt.Printf(
			"(%.1f NM) (%s) %s ALT %s, SPD %.0f knots (Mach %.2f), POS (lat %.6f, lon %.6f) , HDG %.2f deg\n",
			ac.CachedDist,
			ac.Registration,
			aType,
			altBaro,
			ac.GroundSpeed,
			ac.Mach,
			ac.Lat,
			ac.Lon,
			ac.TrueHeading,
		)
	}
}

func (db *Dashboard) checkHighest(ac *Aircraft) {
	if val, ok := ac.AltBaro.(float64); ok {
		if db.highest == nil || db.highest.AltBaro.(float64) < val {
			db.highest = ac

			flight := ac.Flight

			if len(flight) == 0 {
				flight = "unknown " // add space for consistent formatting with ICAO codes
			}
			var altBaro string
			if num, ok := ac.AltBaro.(float64); ok {
				altBaro = fmt.Sprintf("%.0f", num)
			}
			if str, ok := ac.AltBaro.(string); ok {
				altBaro = str
			}

			aType := db.icaoToAircraft[ac.IcaoType].ModelCode

			fmt.Printf("[%s] new highest -> flight %s (%s) %s at %s feet, %.0f knots, heading %.2f degrees\n",
				time.Now().Format("2006-01-02 15:04:05"),
				flight,
				ac.Registration,
				aType,
				altBaro,
				ac.GroundSpeed,
				ac.NavHeading,
			)
		}
	}
}

func (db *Dashboard) checkFastest(ac *Aircraft) {
	if db.fastest == nil || db.fastest.GroundSpeed < ac.GroundSpeed {
		db.fastest = ac

		flight := ac.Flight

		if len(flight) == 0 {
			flight = "unknown " // add space for consistent formatting with ICAO codes
		}
		var altBaro string
		if num, ok := ac.AltBaro.(float64); ok {
			altBaro = fmt.Sprintf("%.0f", num)
		}
		if str, ok := ac.AltBaro.(string); ok {
			altBaro = str
		}

		aType := db.icaoToAircraft[ac.IcaoType].ModelCode

		fmt.Printf("[%s] new fastest -> flight %s (%s) %s at %s feet, %.0f knots (mach %.2f), heading %.2f degrees\n",
			time.Now().Format(TimeFmt),
			flight,
			ac.Registration,
			aType,
			altBaro,
			ac.GroundSpeed,
			ac.Mach,
			ac.NavHeading,
		)
	}
}
