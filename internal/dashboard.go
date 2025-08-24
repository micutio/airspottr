// Package internal provides the Dashboard type and all associated program logic.
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log" //nolint:depguard // Don't feel like using slog
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

const (
	// Lat is Latitude of SIN Airport.
	Lat float64 = 1.359297
	// Lon is Longitude of SIN Airport.
	Lon float64 = 103.989348
	// appIconPath is the file path to the icon png for this application.
	appIconPath = "./assets/icon.png"
	// typeRarityThreshold denotes the maximum rate an aircraft type is seen to be considered rare.
	typeRarityThreshold = 0.001
	// operatorRarityThreshold denotes the maximum rate an operator is seen to be considered rare.
	operatorRarityThreshold = 0.001
	// countryRarityThreshold denotes the maximum rate a country is seen to be considered rare.
	countryRarityThreshold = 0.001
	// altitudeUnknown is what we use for aircraft without a given altitude.
	altitudeUnknown = "  n/a"
	// flightUnknown is what we use for aircraft with missing flight number.
	// Note: we're adding space at the end to have a length that is consistent with ICAO codes.
	flightUnknown = "unknown "
	// flightUnknownCode is a sentinel code we use for aircraft with missing flight number.
	flightUnknownCode = "n/a"
	// typeUnknown is what we use for aircraft with a type that's either empty or can't be found.
	typeUnknown = "unknown"
	// operatorUnknown is what we use for aircraft with a type that's either empty or can't be found.
	operatorUnknown = "unknown"
	// countryUnknown is what we use for aircraft with a type that's either empty or can't be found.
	countryUnknown = "unknown"
)

// Errors used by the Dashboard.
var (
	errParseIcaoAircraftMap      = errors.New("failed to parse ICAO to aircraft map")
	errParseIcaoAirlineMap       = errors.New("failed to parse ICAO to airline map")
	errParseRegToCountryMap      = errors.New("failed to parse reg-prefix to country map")
	errParseHexRangeToCountryMap = errors.New("failed to parse hex-range to country map")
	errParseMilCodeMap           = errors.New("failed to parse mil code to operator map")
)

type aircraftSighting struct {
	lastSeen     time.Time
	lastFlightNo string
	typeDesc     string // typeDesc is the full name of the aircraft type
	operator     string // operator can be either airline or military organisation
	country      string // country of registration
}

type Dashboard struct {
	isWarmup           bool
	Fastest            *aircraftRecord
	Highest            *aircraftRecord
	CurrentAircraft    []aircraftRecord
	aircraftSightings  map[string]aircraftSighting // set of all seen aircraft, maps hex to last seen time
	totalTypeCount     int
	totalOperatorCount int
	totalCountryCount  int
	seenTypeCount      map[string]int // types mapped to how often seen
	seenOperatorCount  map[string]int // airlines mapped to how often seen
	seenCountryCount   map[string]int // airlines mapped to how often seen
	IcaoToAircraft     map[string]icaoAircraft
	IcaoToAirline      map[string]icaoOperator
	regPrefixToCountry map[string]string
	hexRangeToCountry  map[hexRange]string
	milCodeToOperator  map[string]string
	consoleOut         log.Logger
	errOut             log.Logger
}

func NewDashboard(logParams LogParams) (*Dashboard, error) {
	const initError = "newDashboard: %w caused by %w"

	icaoToAircraftMap, aircraftErr := getIcaoToAircraftMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAircraftMap, aircraftErr)
	}

	icaoToAirlineMap, airlineErr := getIcaoToAirlineMap()
	if airlineErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAirlineMap, airlineErr)
	}

	regPrefixToCountryMap, regErr := getRegPrefixMap()
	if regErr != nil {
		return nil, fmt.Errorf(initError, errParseRegToCountryMap, regErr)
	}

	hexRangeToCountryMap, hexRangeErr := getHexRangeToCountryMap()
	if hexRangeErr != nil {
		return nil, fmt.Errorf(initError, errParseHexRangeToCountryMap, hexRangeErr)
	}

	milCodeToOperatorMap, milCodeErr := getMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, fmt.Errorf(initError, errParseMilCodeMap, milCodeErr)
	}

	dash := Dashboard{
		isWarmup:           true,
		Fastest:            nil,
		Highest:            nil,
		CurrentAircraft:    nil,
		aircraftSightings:  make(map[string]aircraftSighting),
		totalTypeCount:     0,
		totalOperatorCount: 0,
		totalCountryCount:  0,
		seenTypeCount:      make(map[string]int),
		seenOperatorCount:  make(map[string]int),
		seenCountryCount:   make(map[string]int),
		IcaoToAircraft:     icaoToAircraftMap,
		IcaoToAirline:      icaoToAirlineMap,
		regPrefixToCountry: regPrefixToCountryMap,
		hexRangeToCountry:  hexRangeToCountryMap,
		milCodeToOperator:  milCodeToOperatorMap,
		consoleOut:         *log.New(logParams.ConsoleOut, "", 0),
		errOut:             *log.New(logParams.ErrorOut, "dashboard", log.LstdFlags),
	}

	return &dash, nil
}

func (db *Dashboard) FinishWarmupPeriod() {
	db.isWarmup = false
}

//////////////////////////////////////////////////////////////////////////////
/// Processing of all aircraft: civilian, military, government, private.    //
//////////////////////////////////////////////////////////////////////////////

// ProcessCivAircraftJSON takes a json record in form of a byte array, transforms it into a list
// of aircraft and performs some processing thereafter.
func (db *Dashboard) ProcessCivAircraftJSON(jsonBytes []byte) {
	var data civAircraftResult
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		db.errOut.Println(fmt.Errorf("failed to unmarshal Json: %w", err))
		return
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return // Valid outcome, no need to log an error.
	}

	db.CurrentAircraft = data.Aircraft
	db.processCivAircraftRecords()
}

func (db *Dashboard) processCivAircraftRecords() {
	sort.Sort(ByFlight(db.CurrentAircraft))
	thisPos := newCoordinates(Lat, Lon)

	for i := range len(db.CurrentAircraft) {
		// Step 1: Get aircraft and time of sighting
		aircraft := (db.CurrentAircraft)[i]
		lastSeenMsBeforeNow := time.Duration(aircraft.Seen) * time.Second
		lastSeenTime := time.Now().Add(-lastSeenMsBeforeNow)

		// Step 2: Retrieve previous sighting or create new one.
		sighting, exists := db.aircraftSightings[aircraft.Hex]
		if !exists {
			sighting = aircraftSighting{
				lastSeen:     lastSeenTime,
				lastFlightNo: flightUnknown,
				typeDesc:     typeUnknown,
				operator:     operatorUnknown,
				country:      countryUnknown,
			}
		}

		// Check whether we've seen this aircraft before by comparing last and current flight number.
		// If they differ, then we allow re-recording the statistics.
		thisFlightNo := aircraft.GetFlightNoAsStr()
		isNewFlight := sighting.lastFlightNo != thisFlightNo
		sighting.lastFlightNo = thisFlightNo

		// Step 3: Update distance
		acPos := newCoordinates(aircraft.Lat, aircraft.Lon)
		(db.CurrentAircraft)[i].CachedDist = Distance(thisPos, acPos).Kilometers()

		// Step 3: Update all aircraft, type, operator and country statistics
		db.updateHighest(&aircraft)
		db.updateFastest(&aircraft)
		db.updateType(&aircraft, &sighting, isNewFlight)
		db.updateOperator(&aircraft, &sighting, isNewFlight)
		db.updateCountry(&aircraft, &sighting, isNewFlight)

		// Finally, update the records
		db.aircraftSightings[aircraft.Hex] = sighting
	}
}

func (db *Dashboard) updateType(aircraft *aircraftRecord, sighting *aircraftSighting, isNewFlight bool) {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.typeDesc != typeUnknown && !isNewFlight {
		return
	}

	// We couldn't find out the type of this aircraft, unable to update statistics.
	aType := db.IcaoToAircraft[aircraft.IcaoType].ModelCode
	if aType == "" {
		return
	}

	sighting.typeDesc = aType

	// Valid type found! Record type and update type rarities.
	thisTypeCountNew := db.seenTypeCount[aType] + 1
	db.seenTypeCount[aType] = thisTypeCountNew
	db.totalTypeCount++
	typeRarity := float32(thisTypeCountNew) / float32(db.totalTypeCount)

	// db.logger.Debug(
	//	"type rarity calculation: ",
	//	" aircraft flight", aircraft.Flight,
	//	"type", sighting.typeDesc,
	//	"thisTypeCountNew", thisTypeCountNew,
	//	"totalTypeCount", db.totalTypeCount,
	//	"typeRarity", typeRarity,
	//	"typeRarityThreshold", typeRarityThreshold)

	if typeRarity > typeRarityThreshold {
		return
	}

	// db.logger.Info(
	//	"type rarity calculation: ",
	//	"thisTypeCountNew", thisTypeCountNew,
	//	"totalTypeCount", db.totalTypeCount,
	//	"typeRarity", typeRarity,
	//	"typeRarityThreshold", typeRarityThreshold)

	db.consoleOut.Printf("found rare type %s\n", db.aircraftToString(aircraft))

	if !db.isWarmup {
		db.notifyRareType(aircraft, sighting)
	}
}

func (db *Dashboard) updateOperator(aircraft *aircraftRecord, sighting *aircraftSighting, isNewFlight bool) {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.operator != operatorUnknown && !isNewFlight {
		return
	}

	flightNo := aircraft.Flight
	if flightNo == "" {
		return
	}

	// First option: try to detect the airline and get operator & country from it.
	flightCode := aircraft.GetFlightNoAsIcaoCode()
	if flightCode != flightUnknownCode {
		if operatorRecord, opExists := db.IcaoToAirline[flightCode]; opExists {
			sighting.operator = operatorRecord.Company
		}
	}

	// Unable to detect airline, maybe it's military or government.
	if sighting.operator == operatorUnknown {
		if militaryOperator, milOpExists := db.milCodeToOperator[flightCode]; milOpExists {
			sighting.operator = militaryOperator
		}
	}

	// Operator still not found, check whether the 'ownOp' field in the aircraft record is set.
	if sighting.operator == operatorUnknown && aircraft.OwnOp != "" {
		sighting.operator = aircraft.OwnOp
	}

	// Did not manage to find out the operator of this aircraft.
	if sighting.operator == operatorUnknown {
		return
	}

	thisOperatorCountNew := db.seenOperatorCount[sighting.operator] + 1
	db.seenOperatorCount[sighting.operator] = thisOperatorCountNew
	db.totalOperatorCount++
	operatorRarity := float32(thisOperatorCountNew) / float32(db.totalOperatorCount)

	// db.logger.Debug(
	//	"operator rarity calculation:",
	//	"operator", sighting.operator,
	//	"thisOperatorCountNew", thisOperatorCountNew,
	//	"totalOperatorCount", db.totalOperatorCount,
	//	"operatorRarity", operatorRarity,
	//	"operatorRarityThreshold", operatorRarityThreshold)

	if operatorRarity > operatorRarityThreshold {
		return
	}

	// db.logger.Debug(
	//	"operator rarity calculation: ",
	//	"thisOperatorCountNew", thisOperatorCountNew,
	//	"totalOperatorCount", db.totalOperatorCount,
	//	"operatorRarity", operatorRarity,
	//	"operatorRarityThreshold", operatorRarityThreshold)

	db.consoleOut.Printf("found rare operator: %s\n", sighting.operator)

	if !db.isWarmup {
		db.notifyRareOperator(aircraft, sighting)
	}
}

func (db *Dashboard) updateCountry(aircraft *aircraftRecord, sighting *aircraftSighting, isNewFlight bool) {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.country != countryUnknown && !isNewFlight {
		return
	}

	flightNo := aircraft.Flight
	if flightNo == "" {
		return
	}

	// Option #1: Try to detect the airline and get operator & country from it.
	flightCode := aircraft.GetFlightNoAsIcaoCode()
	if flightCode != flightUnknownCode {
		if operatorRecord, exists := db.IcaoToAirline[flightCode]; exists {
			sighting.country = strings.ToUpper(operatorRecord.Country)
		}
	}

	// Option #2: Detect country by the range of it's hex registration.
	if sighting.country == countryUnknown {
		sighting.country = strings.ToUpper(db.getCountryByHexRange(aircraft.Hex))
	}

	// Option #3: Detect country by its ICAO registration prefix.
	if sighting.country == countryUnknown {
		if country, exists := db.getCountryByRegPrefix(aircraft.Registration); exists {
			sighting.country = strings.ToUpper(country)
		}
	}

	// Unable to detect country of this aircraft.
	if sighting.country == countryUnknown {
		return
	}

	thisCountryCountNew := db.seenCountryCount[sighting.country] + 1
	db.seenCountryCount[sighting.country] = thisCountryCountNew
	db.totalCountryCount++
	countryRarity := float32(thisCountryCountNew) / float32(db.totalCountryCount)

	// db.logger.Debug(
	//	"country rarity calculation:",
	//	"country", sighting.country,
	//	"thisCountryCountNew", thisCountryCountNew,
	//	"totalCountryCount", db.totalCountryCount,
	//	"countryRarity", countryRarity,
	//	"countryRarityThreshold", countryRarityThreshold)

	if countryRarity > countryRarityThreshold {
		return
	}

	// db.logger.Debug(
	//	"country rarity calculation: ",
	//	"thisCountryCountNew", thisCountryCountNew,
	//	"totalCountryCount", db.totalCountryCount,
	//	"countryRarity", countryRarity,
	//	"countryRarityThreshold", countryRarityThreshold)
	db.consoleOut.Printf("found rare country: %s\n", sighting.country)

	if !db.isWarmup {
		db.notifyRareCountry(aircraft, sighting)
	}
}

func (db *Dashboard) getCountryByHexRange(hexAsStr string) string {
	hexAsInt, err := strconv.ParseInt(hexAsStr, 16, 64)
	if err != nil {
		db.errOut.Printf("unable to convert hex to int: %s\n", hexAsStr)
		return countryUnknown
	}
	for key, value := range db.hexRangeToCountry {
		if hexAsInt > key.LowerBound && hexAsInt < key.UpperBound {
			return value
		}
	}
	return countryUnknown
}

func (db *Dashboard) getCountryByRegPrefix(reg string) (string, bool) {
	for key, value := range db.regPrefixToCountry {
		if strings.Contains(reg, key) {
			return value, true
		}
	}

	return "", false
}

func (db *Dashboard) notifyRareType(aircraft *aircraftRecord, sighting *aircraftSighting) {
	var aType string
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = sighting.typeDesc
	}
	msgBody := fmt.Sprintf("%s (%s)", aType, aircraft.Registration)
	err := beeep.Notify("Rare Aircraft Type Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func (db *Dashboard) notifyRareOperator(aircraft *aircraftRecord, sighting *aircraftSighting) {
	operator := sighting.operator

	msgBody := fmt.Sprintf("%s flying %s (%s)", operator, aircraft.Description, aircraft.Registration)
	err := beeep.Notify("Rare Operator Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func (db *Dashboard) notifyRareCountry(aircraft *aircraftRecord, sighting *aircraftSighting) {
	country := sighting.country

	msgBody := fmt.Sprintf("%s-based %s (%s)", country, aircraft.Description, aircraft.Registration)
	err := beeep.Notify("Rare Country Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func (db *Dashboard) updateHighest(aircraft *aircraftRecord) {
	thisAltitude, thisAltOk := aircraft.AltBaro.(float64)
	if !thisAltOk {
		return
	}

	//nolint:errcheck // If highest is initialized the altBaro is always valid.
	if db.Highest != nil && db.Highest.AltBaro.(float64) > thisAltitude {
		return
	}

	db.Highest = aircraft
}

func (db *Dashboard) updateFastest(aircraft *aircraftRecord) {
	if db.Fastest != nil && db.Fastest.GroundSpeed > aircraft.GroundSpeed {
		return
	}

	db.Fastest = aircraft
}

// PrintSummary prints the highest, fastest and the most and the least common types.
func (db *Dashboard) PrintSummary() {
	db.consoleOut.Println("=== Summary ===")
	db.listByRarity("aircraft", db.seenTypeCount)
	db.listByRarity("operator", db.seenOperatorCount)
	db.listByRarity("country", db.seenCountryCount)
	db.consoleOut.Println("Fastest Aircraft:")
	db.consoleOut.Println(db.aircraftToString(db.Fastest))
	db.consoleOut.Println("Highest Aircraft:")
	db.consoleOut.Println(db.aircraftToString(db.Highest))
	db.consoleOut.Println("=== End Summary ===")
}

type propertyCountTuple struct {
	property string
	count    int
}

type ByCount []propertyCountTuple

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].count < a[j].count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (db *Dashboard) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := make([]propertyCountTuple, len(propertyCountMap))
	i := 0
	for key, value := range propertyCountMap {
		propertyCounts[i] = propertyCountTuple{property: key, count: value}
		i++
	}

	sort.Sort(ByCount(propertyCounts))

	db.consoleOut.Printf("Rarity from least to most common %s", propertyName)
	for j := range propertyCounts {
		db.consoleOut.Printf("%6d - %s\n", propertyCounts[j].count, propertyCounts[j].property)
	}
}

// aircraftToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func (db *Dashboard) aircraftToString(aircraft *aircraftRecord) string {
	thisPos := newCoordinates(Lat, Lon)
	acPos := newCoordinates(aircraft.Lat, aircraft.Lon)
	aircraft.CachedDist = Distance(thisPos, acPos).Kilometers()

	flight := aircraft.GetFlightNoAsStr()
	altitude := aircraft.GetAltitudeAsStr()
	aType := db.IcaoToAircraft[aircraft.IcaoType].ModelCode

	return fmt.Sprintf("FNO %s DST %4.0f km ALT %s SPD %3.0f HDG %3.0f TID %s (%s)",
		flight,
		aircraft.CachedDist,
		altitude,
		aircraft.GroundSpeed,
		aircraft.NavHeading,
		aType,
		aircraft.Registration)
}

//////////////////////////////////////////////////////////////////////////////
/// Processing of military & government aircraft in a wider area            //
//////////////////////////////////////////////////////////////////////////////

func (db *Dashboard) ProcessMilAircraftJSON(jsonBytes []byte) {
	var data milAircraftResult
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		db.errOut.Println(fmt.Errorf("failed to unmarshal military aircraft JSON %w", err))
		return
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return // Valid outcome, no need to log an error.
	}

	db.processMilAircraftRecords(&(data.Aircraft))
}

func (db *Dashboard) processMilAircraftRecords(allAircraft *[]aircraftRecord) {
	thisPos := newCoordinates(Lat, Lon)
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

	db.consoleOut.Println("Military aircraft in increasing distance from here:")

	for i := range len(*allAircraft) {
		aircraft := (*allAircraft)[i]
		if (aircraft.Lat == 0 && aircraft.Lon == 0) || aircraft.CachedDist > 1000.0 {
			continue
		}

		db.consoleOut.Println(db.aircraftToString(&aircraft))
	}
}
