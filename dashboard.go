package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"
)

type dashboard struct {
	// Fields for tracking some statistics
	fastest *aircraft
	highest *aircraft
	// Data
	seenAircraft      map[string]string // set of all seen aircraft, mapped to their type
	milCodeToOperator map[string]string
	icaoToAircraft    map[string]icaoAircraft
}

func newDashboard() dashboard {
	return dashboard{
		fastest:           nil,
		highest:           nil,
		seenAircraft:      make(map[string]string),
		icaoToAircraft:    GetIcaoToAircraftMap(),
		milCodeToOperator: GetMilCodeToOperatorMap(),
	}
}

// processCivAircraftJSON takes a json record in form of a byte array, transforms it into a list
// of aircraft and performs some processing thereafter.
func (db *dashboard) processCivAircraftJSON(jsonBytes []byte) error {
	var data civAircraftRecord
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

func (db *dashboard) processCivAircraftRecords(aircraft *[]aircraft) {
	sort.Sort(ByFlight(*aircraft))

	for i := range len(*aircraft) {
		ac := (*aircraft)[i]
		aType := db.icaoToAircraft[ac.IcaoType].ModelCode
		db.seenAircraft[ac.Hex] = aType
		db.checkHighest(&ac)
		db.checkFastest(&ac)
	}
}

func (db *dashboard) processMilAircraftJSON(jsonBytes []byte) error {
	var data milAircraftRecord
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

func (db *dashboard) processMilAircraftRecords(aircraft *[]aircraft) {
	thisPos := newCoordinates(lat, lon)

	for i := range len(*aircraft) {
		ac := (*aircraft)[i]
		acPos := newCoordinates(ac.Lat, ac.Lon)
		(*aircraft)[i].CachedDist = Distance(thisPos, acPos).Kilometers()
	}

	sort.Sort(ByDistance(*aircraft))

	// Only print something if there are any miliary aircraft within range.
	if len(*aircraft) == 0 || (*aircraft)[0].CachedDist > 1000 {
		return
	}

	log.Printf("[%s] Military aircraft in increasing distance from here:\n", time.Now().Format(timeFmt))

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

		log.Printf(
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

func (db *dashboard) checkHighest(ac *aircraft) {
	if val, ok := ac.AltBaro.(float64); ok {
		if db.highest != nil && db.highest.AltBaro.(float64) > val {
			return
		}

		db.highest = ac

		flight := ac.Flight

		if len(flight) == 0 {
			flight = "unknown " // add space for consistent formatting with ICAO codes
		}

		var altBaro string
		if num, numOK := ac.AltBaro.(float64); numOK {
			altBaro = fmt.Sprintf("%5.0f", num)
		}

		if str, strOk := ac.AltBaro.(string); strOk {
			altBaro = str
		}

		aType := db.icaoToAircraft[ac.IcaoType].ModelCode

		log.Printf("[%s] highest -> FLT %s ALT %s SPD %3.0f HDG %6.2f ID %q (%s)\n",
			time.Now().Format(timeFmt),
			flight,
			altBaro,
			ac.GroundSpeed,
			ac.NavHeading,
			aType,
			ac.Registration,
		)
	}
}

func (db *dashboard) checkFastest(ac *aircraft) {
	if db.fastest != nil && db.fastest.GroundSpeed > ac.GroundSpeed {
		return
	}

	db.fastest = ac

	flight := ac.Flight

	if len(flight) == 0 {
		flight = "unknown " // add space for consistent formatting with ICAO codes
	}

	var altBaro string
	if num, numOk := ac.AltBaro.(float64); numOk {
		altBaro = fmt.Sprintf("%.0f", num)
	}

	if str, strOk := ac.AltBaro.(string); strOk {
		altBaro = str
	}

	aType := db.icaoToAircraft[ac.IcaoType].ModelCode

	log.Printf("[%s] fastest -> FLT %s ALT %s SPD %3.0f HDG %6.2f ID %q (%s)\n",
		time.Now().Format(timeFmt),
		flight,
		altBaro,
		ac.GroundSpeed,
		ac.NavHeading,
		aType,
		ac.Registration,
	)
}

type typeCountTuple struct {
	typ   string
	count int
}

type ByCount []typeCountTuple

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].count < a[j].count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (db *dashboard) listTypesByRarity() {
	typeCountMap := make(map[string]int)
	for _, value := range db.seenAircraft {
		typeCountMap[value] += 1
	}

	typeCountList := make([]typeCountTuple, len(typeCountMap))
	for key, value := range typeCountMap {
		typeCountList = append(typeCountList, typeCountTuple{typ: key, count: value})
	}

	sort.Sort(ByCount(typeCountList))

	log.Printf("[%s] aircraft types from least to most common\n", time.Now().Format(timeFmt))

	for i := range len(typeCountList) {
		log.Printf("%6d - %q\n", typeCountList[i].count, typeCountList[i].typ)
	}
}
