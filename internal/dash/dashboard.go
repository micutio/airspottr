// Package dash provides the Dashboard type and all associated program logic.
package dash

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// Lat is Latitude of SIN Airport.
	// Lat float64 = 1.359297
	// Lon is Longitude of SIN Airport.
	// Lon float64 = 103.989348
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
	Lat                float32
	Lon                float32
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
	// TODO: Move out to notifications
	consoleOut log.Logger
	// TODO: Check whether possible to move out to notifications
	errOut log.Logger
}

func NewDashboard(lat float32, lon float32, stderr *io.Writer) (*Dashboard, error) {
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
		Lat:                lat,
		Lon:                lon,
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
		errOut:             *log.New(*stderr, "dashboard", log.LstdFlags),
	}

	dash.errOut.Println("Dashboard init")

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
	thisPos := newCoordinates(float64(db.Lat), float64(db.Lon))

	for idx := range len(db.CurrentAircraft) {
		// Step 1: Get aircraft and time of sighting
		aircraft := (db.CurrentAircraft)[idx]
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
		// If they differ, then we allow recording in the statistics again.
		thisFlightNo := aircraft.GetFlightNoAsStr()
		isFlightIdentified := sighting.lastFlightNo == flightUnknown && thisFlightNo != flightUnknown
		isFlightUpdated := sighting.lastFlightNo != flightUnknown &&
			thisFlightNo != flightUnknown &&
			sighting.lastFlightNo != thisFlightNo

		isNewFlight := !exists || isFlightUpdated

		if isFlightIdentified || isFlightUpdated {
			sighting.lastFlightNo = thisFlightNo
		}

		// Step 3: Update distance
		acPos := newCoordinates(aircraft.Lat, aircraft.Lon)
		(db.CurrentAircraft)[idx].CachedDist = Distance(thisPos, acPos).Kilometers()
		aircraft.CachedDist = Distance(thisPos, acPos).Kilometers()

		// Step 3: Update all aircraft, type, operator and country statistics
		db.updateHighest(&aircraft)
		db.updateFastest(&aircraft)

		newRarities := noRarity
		rareTypeFlag := db.updateType(&sighting, &aircraft, isNewFlight)
		rareOperatorFlag := db.updateOperator(&sighting, &aircraft, isNewFlight)
		rareCountryFlag := db.updateCountry(&sighting, &aircraft, isNewFlight)

		newRarities |= rareTypeFlag << 0
		newRarities |= rareOperatorFlag << 1
		newRarities |= rareCountryFlag << 2

		emitRarityNotifications(newRarities, &aircraft, &sighting, &db.consoleOut, db.isWarmup)

		// Finally, update the records
		db.aircraftSightings[aircraft.Hex] = sighting
	}
}

func (db *Dashboard) updateType(
	sighting *aircraftSighting,
	aircraft *aircraftRecord,
	isNewFlight bool,
) (foundRareFlag rarityFlag) {
	// We already know the type or just saw this one recently, no need to update again.
	isTypeKnown := sighting.typeDesc != typeUnknown
	isFlightKnown := !isNewFlight
	if isTypeKnown && isFlightKnown {
		return 0
	}

	// We couldn't find out the type of this aircraft, unable to update statistics.
	aType := db.IcaoToAircraft[aircraft.IcaoType].Make
	if aType == "" {
		return 0
	}

	sighting.typeDesc = aType
	aircraft.CachedType = aType

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
		return 0
	}

	// db.logger.Info(
	//	"type rarity calculation: ",
	//	"thisTypeCountNew", thisTypeCountNew,
	//	"totalTypeCount", db.totalTypeCount,
	//	"typeRarity", typeRarity,
	//	"typeRarityThreshold", typeRarityThreshold)

	return 1
}

func (db *Dashboard) updateOperator(
	sighting *aircraftSighting,
	aircraft *aircraftRecord,
	isNewFlight bool,
) (foundRareFlag rarityFlag) {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.operator != operatorUnknown && !isNewFlight {
		return 0
	}

	flightNo := aircraft.Flight
	if flightNo == "" {
		return 0
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
		return 0
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
		return 0
	}

	// db.logger.Debug(
	//	"operator rarity calculation: ",
	//	"thisOperatorCountNew", thisOperatorCountNew,
	//	"totalOperatorCount", db.totalOperatorCount,
	//	"operatorRarity", operatorRarity,
	//	"operatorRarityThreshold", operatorRarityThreshold)

	return 1
}

func (db *Dashboard) updateCountry(
	sighting *aircraftSighting,
	aircraft *aircraftRecord,
	isNewFlight bool,
) (foundRareFlag rarityFlag) {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.country != countryUnknown && !isNewFlight {
		return 0
	}

	flightNo := aircraft.Flight
	if flightNo == "" {
		return 0
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
		return 0
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
		return 0
	}

	// db.logger.Debug(
	//	"country rarity calculation: ",
	//	"thisCountryCountNew", thisCountryCountNew,
	//	"totalCountryCount", db.totalCountryCount,
	//	"countryRarity", countryRarity,
	//	"countryRarityThreshold", countryRarityThreshold)
	return 1
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

func (db *Dashboard) GetMaxTypeNameLength() int {
	// Create a new table with specified columns and initial empty rows.
	maxTypeLen := 0
	for _, value := range db.IcaoToAircraft {
		if len(value.Make) > maxTypeLen {
			maxTypeLen = len(value.Make)
		}
	}
	return maxTypeLen
}
