// Package main provides the flight tracking application
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
)

var (
	errParseIcaoMap    = errors.New("failed to parse ICAO to aircraft map")
	errParseMilCodeMap = errors.New("failed to parse mil code to operator map")
)

type dashboard struct {
	// Fields for tracking some statistics
	fastest *aircraftRecord
	highest *aircraftRecord
	// Data
	seenAircraft      map[string]string // set of all seen aircraft, mapped to their type
	milCodeToOperator map[string]string
	icaoToAircraft    map[string]icaoAircraft
	logger            slog.Logger
}

func newDashboard() (*dashboard, error) {
	icaoToAircraftMap, icaoErr := getIcaoToAircraftMap()
	if icaoErr != nil {
		return nil, errParseIcaoMap
	}

	milCodeToOperatorMap, milCodeErr := getMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, errParseMilCodeMap
	}

	dash := dashboard{
		fastest:           nil,
		highest:           nil,
		seenAircraft:      make(map[string]string),
		icaoToAircraft:    icaoToAircraftMap,
		milCodeToOperator: milCodeToOperatorMap,
		logger:            *slog.Default(),
	}

	return &dash, nil
}

// processCivAircraftJSON takes a json record in form of a byte array, transforms it into a list
// of aircraft and performs some processing thereafter.
func (db *dashboard) processCivAircraftJSON(jsonBytes []byte) {
	var data civAircraftResult
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		db.logger.Error("dashboard:", slog.Any("failed to unmarshal Json", err))
		return
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return // Valid outcome, no need to log an error.
	}

	db.processCivAircraftRecords(&data.Aircraft)
}

func (db *dashboard) processCivAircraftRecords(allAircraft *[]aircraftRecord) {
	sort.Sort(ByFlight(*allAircraft))

	for i := range len(*allAircraft) {
		aircraft := (*allAircraft)[i]
		aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode
		db.seenAircraft[aircraft.Hex] = aType
		db.checkHighest(&aircraft)
		db.checkFastest(&aircraft)
	}
}

func (db *dashboard) processMilAircraftJSON(jsonBytes []byte) {
	var data milAircraftResult
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		db.logger.Error("dashboard:", slog.Any("failed to unmarshal military aircraft JSON", err))
		return
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return // Valid outcome, no need to log an error.
	}

	db.processMilAircraftRecords(&(data.Aircraft))
}

func (db *dashboard) processMilAircraftRecords(allAircraft *[]aircraftRecord) {
	thisPos := newCoordinates(lat, lon)

	for i := range len(*allAircraft) {
		aircraft := (*allAircraft)[i]
		acPos := newCoordinates(aircraft.Lat, aircraft.Lon)
		(*allAircraft)[i].CachedDist = Distance(thisPos, acPos).Kilometers()
	}

	sort.Sort(ByDistance(*allAircraft))

	// Only print something if there are any military aircraft within range.
	if len(*allAircraft) == 0 || (*allAircraft)[0].CachedDist > 1000 {
		return
	}

	db.logger.Info("Military aircraft in increasing distance from here:\n")

	for i := range len(*allAircraft) {
		aircraft := (*allAircraft)[i]
		if (aircraft.Lat == 0 && aircraft.Lon == 0) || aircraft.CachedDist > 1000.0 {
			continue
		}

		var altBaro string
		if num, ok := aircraft.AltBaro.(float64); ok {
			altBaro = fmt.Sprintf("%5.0f", num)
		}

		if str, ok := aircraft.AltBaro.(string); ok {
			altBaro = str
		}

		aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode

		db.logger.Info(
			fmt.Sprintf(
				"(%5.1f Km) ALT %s SPD %3.0f POS (%7.3f, %7.3f) HDG %6.2f ID %q (%s)\n",
				aircraft.CachedDist,
				altBaro,
				aircraft.GroundSpeed,
				aircraft.Lat,
				aircraft.Lon,
				aircraft.TrueHeading,
				aType,
				aircraft.Registration),
		)
	}
}

func (db *dashboard) checkHighest(aircraft *aircraftRecord) {
	thisAltitude, thisAltOk := aircraft.AltBaro.(float64)
	if !thisAltOk {
		return
	}

	//nolint:errcheck // If highest is initialized the altBaro is always valid.
	if db.highest != nil && db.highest.AltBaro.(float64) > thisAltitude {
		return
	}

	db.highest = aircraft

	flight := aircraft.Flight

	if len(flight) == 0 {
		flight = "unknown " // add space for consistent formatting with ICAO codes
	}

	var altBaro string
	if num, numOK := aircraft.AltBaro.(float64); numOK {
		altBaro = fmt.Sprintf("%5.0f", num)
	}

	if str, strOk := aircraft.AltBaro.(string); strOk {
		altBaro = str
	}

	aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode

	db.logger.Info(
		fmt.Sprintf("highest -> FLT %s ALT %s SPD %3.0f HDG %6.2f ID %q (%s)\n",
			flight,
			altBaro,
			aircraft.GroundSpeed,
			aircraft.NavHeading,
			aType,
			aircraft.Registration))
}

func (db *dashboard) checkFastest(aircraft *aircraftRecord) {
	if db.fastest != nil && db.fastest.GroundSpeed > aircraft.GroundSpeed {
		return
	}

	db.fastest = aircraft

	flight := aircraft.Flight

	if len(flight) == 0 {
		flight = "unknown " // add space for consistent formatting with ICAO codes
	}

	var altBaro string
	if num, numOk := aircraft.AltBaro.(float64); numOk {
		altBaro = fmt.Sprintf("%.0f", num)
	}

	if str, strOk := aircraft.AltBaro.(string); strOk {
		altBaro = str
	}

	aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode

	db.logger.Info(fmt.Sprintf("fastest -> FLT %s ALT %s SPD %3.0f HDG %6.2f ID %q (%s)\n",
		flight,
		altBaro,
		aircraft.GroundSpeed,
		aircraft.NavHeading,
		aType,
		aircraft.Registration))
}

type aircraftTypeCountTuple struct {
	acType string
	count  int
}

type ByCount []aircraftTypeCountTuple

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].count < a[j].count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (db *dashboard) listTypesByRarity() {
	typeCountMap := make(map[string]int)
	for _, value := range db.seenAircraft {
		typeCountMap[value]++
	}

	typeCountList := make([]aircraftTypeCountTuple, len(typeCountMap))
	i := 0
	for key, value := range typeCountMap {
		typeCountList[i] = aircraftTypeCountTuple{acType: key, count: value}
		i++
	}

	sort.Sort(ByCount(typeCountList))

	db.logger.Info("aircraft types from least to most common")
	for i := range typeCountList {
		db.logger.Info(fmt.Sprintf("%6d - %q\n", typeCountList[i].count, typeCountList[i].acType))
	}
}
