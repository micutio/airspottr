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
		ac := (*aircraft)[i]
		db.checkHighest(&ac)
		db.checkFastest(&ac)
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
		(*aircraft)[i].CachedDist = Distance(thisPos, acPos).Kilometers()
	}
	sort.Sort(ByDistance(*aircraft))

	fmt.Printf("[%s] Military aircraft in increasing distance from here:\n", time.Now().Format(TimeFmt))
	for i := range len(*aircraft) {
		ac := (*aircraft)[i]
		if (ac.Lat == 0 && ac.Lon == 0) || ac.CachedDist > 1000.0 {
			continue
		}

		var altBaro string
		if num, ok := ac.AltBaro.(float64); ok {
			altBaro = fmt.Sprintf("%5.0f", num)
		}
		if str, ok := ac.AltBaro.(string); ok {
			altBaro = str
		}
		aType := db.icaoToAircraft[ac.IcaoType].ModelCode

		fmt.Printf(
			"(%5.1f Km) ALT %s SPD %3.0f POS (%7.3f, %7.3f) HDG %6.2f ID %q (%s)\n",
			ac.CachedDist,
			altBaro,
			ac.GroundSpeed,
			ac.Lat,
			ac.Lon,
			ac.TrueHeading,
			aType,
			ac.Registration,
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
				altBaro = fmt.Sprintf("%5.0f", num)
			}
			if str, ok := ac.AltBaro.(string); ok {
				altBaro = str
			}

			aType := db.icaoToAircraft[ac.IcaoType].ModelCode

			fmt.Printf("[%s] highest -> FLT %s ALT %s SPD %3.0f HDG %6.2f ID %q (%s)\n",
				time.Now().Format("2006-01-02 15:04:05"),
				flight,
				altBaro,
				ac.GroundSpeed,
				ac.NavHeading,
				aType,
				ac.Registration,
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

		fmt.Printf("[%s] fastest -> FLT %s ALT %s SPD %3.0f HDG %6.2f ID %q (%s)\n",
			time.Now().Format(TimeFmt),
			flight,
			altBaro,
			ac.GroundSpeed,
			ac.NavHeading,
			aType,
			ac.Registration,
		)
	}
}
