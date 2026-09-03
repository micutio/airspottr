// Package application provides the Dashboard type and all associated program logic.
package application

import (
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// Errors used by the Dashboard.
var (
	errParseIcaoAircraftMap      = errors.New("failed to parse ICAO to aircraft map")
	errParseIcaoAirlineMap       = errors.New("failed to parse ICAO to airline map")
	errParseRegToCountryMap      = errors.New("failed to parse reg-prefix to country map")
	errParseHexRangeToCountryMap = errors.New("failed to parse hex-range to country map")
	errParseMilCodeMap           = errors.New("failed to parse mil code to operator map")
)

type Dashboard struct {
	IsWarmup           bool
	Lat                float64
	Lon                float64
	Fastest            *obs.AircraftRecord
	Highest            *obs.AircraftRecord
	CurrentAircraft    []obs.AircraftRecord
	RareSightings      []obs.RareSighting
	CachedFlightRoutes map[string]*ref.FlightRouteRecord
	AircraftSightings  map[string]*obs.AircraftSighting // set of all seen aircraft, maps hex to last seen time
	TotalTypeCount     int
	TotalOperatorCount int
	TotalCountryCount  int
	SeenTypeCount      map[string]int // types mapped to how often seen
	SeenOperatorCount  map[string]int // airlines mapped to how often seen
	SeenCountryCount   map[string]int // airlines mapped to how often seen
	IcaoToAircraft     map[string]ref.IcaoAircraft
	IcaoToAirline      map[string]ref.IcaoOperator
	RegPrefixToCountry map[string]string
	HexRangeToCountry  map[ref.HexRange]string
	MilCodeToOperator  map[string]string
	ErrOut             log.Logger
}

func NewDashboard(lat float64, lon float64, stderr *io.Writer) (*Dashboard, error) {
	const initError = "newDashboard: %w caused by %w"

	icaoToAircraftMap, aircraftErr := ref.GetIcaoToAircraftMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAircraftMap, aircraftErr)
	}

	icaoToAirlineMap, airlineErr := ref.GetIcaoToAirlineMap()
	if airlineErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAirlineMap, airlineErr)
	}

	regPrefixToCountryMap, regErr := ref.GetRegPrefixMap()
	if regErr != nil {
		return nil, fmt.Errorf(initError, errParseRegToCountryMap, regErr)
	}

	hexRangeToCountryMap, hexRangeErr := ref.GetHexRangeToCountryMap()
	if hexRangeErr != nil {
		return nil, fmt.Errorf(initError, errParseHexRangeToCountryMap, hexRangeErr)
	}

	milCodeToOperatorMap, milCodeErr := ref.GetMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, fmt.Errorf(initError, errParseMilCodeMap, milCodeErr)
	}

	dashboard := Dashboard{
		IsWarmup:           true,
		Lat:                lat,
		Lon:                lon,
		Fastest:            nil,
		Highest:            nil,
		CurrentAircraft:    nil,
		RareSightings:      nil,
		CachedFlightRoutes: make(map[string]*ref.FlightRouteRecord),
		AircraftSightings:  make(map[string]*obs.AircraftSighting),
		TotalTypeCount:     0,
		TotalOperatorCount: 0,
		TotalCountryCount:  0,
		SeenTypeCount:      make(map[string]int),
		SeenOperatorCount:  make(map[string]int),
		SeenCountryCount:   make(map[string]int),
		IcaoToAircraft:     icaoToAircraftMap,
		IcaoToAirline:      icaoToAirlineMap,
		RegPrefixToCountry: regPrefixToCountryMap,
		HexRangeToCountry:  hexRangeToCountryMap,
		MilCodeToOperator:  milCodeToOperatorMap,
		ErrOut:             *log.New(*stderr, "dashboard ", log.LstdFlags),
	}

	dashboard.ErrOut.Println("Dashboard init")

	return &dashboard, nil
}

func (db *Dashboard) FinishWarmupPeriod() {
	db.IsWarmup = false
}

//////////////////////////////////////////////////////////////////////////////
/// Processing of all aircraft: civilian, military, government, private.    //
//////////////////////////////////////////////////////////////////////////////

func (db *Dashboard) ProcessAircraftRecords(aircraftRecords []obs.AircraftRecord) {
	db.CurrentAircraft = aircraftRecords
	sort.Sort(obs.ByFlight(db.CurrentAircraft))
	thisPos := ref.NewCoordinates(db.Lat, db.Lon)
	var rareSightings []obs.RareSighting

	for idx := range len(db.CurrentAircraft) {
		// Get aircraft and time of sighting
		aircraft := &(db.CurrentAircraft)[idx]
		lastSeenMsBeforeNow := time.Duration(aircraft.Seen) * time.Second
		lastSeenTime := time.Now().Add(-lastSeenMsBeforeNow)

		// Retrieve previous sighting or create new one.
		sighting, exists := db.AircraftSightings[aircraft.Hex]
		if !exists {
			sighting = &obs.AircraftSighting{
				LastSeen:     lastSeenTime,
				LastFlightNo: obs.FlightUnknown,
				Registration: aircraft.Registration,
				Latitude:     aircraft.Lat,
				Longitude:    aircraft.Lon,
				Direction:    obs.GetDirection(db.Lat, db.Lon, aircraft.Lat, aircraft.Lon),
				Distance:     math.MaxInt,
				TypeShort:    "",
				TypeDesc:     obs.TypeUnknown,
				Operator:     obs.OperatorUnknown,
				Country:      obs.CountryUnknown,
				Info:         "",
				Flightroute:  nil,
			}
		}

		if sighting.Registration == "" {
			sighting.Registration = aircraft.Registration
		}

		// Check whether we've seen this aircraft before by comparing last and current Flight number.
		// If they differ, then we allow recording in the statistics again.
		thisFlightNo := aircraft.GetFlightNoAsStr()
		isFlightIdentified := sighting.LastFlightNo == obs.FlightUnknown && thisFlightNo != obs.FlightUnknown
		isFlightUpdated := sighting.LastFlightNo != obs.FlightUnknown &&
			thisFlightNo != obs.FlightUnknown &&
			sighting.LastFlightNo != thisFlightNo

		isNewFlight := !exists || isFlightUpdated

		if isFlightIdentified || isFlightUpdated {
			sighting.LastFlightNo = thisFlightNo
		}

		// Update distance
		acPos := ref.NewCoordinates(aircraft.Lat, aircraft.Lon)
		(db.CurrentAircraft)[idx].CachedDist = ref.Distance(thisPos, acPos).Kilometers()
		aircraft.CachedDist = ref.Distance(thisPos, acPos).Kilometers()
		sighting.Distance = aircraft.CachedDist

		// Update all aircraft, type, operator and country statistics
		db.updateHighest(aircraft)
		db.updateFastest(aircraft)

		newRarities := obs.NoRarity
		rareTypeFlag := db.updateType(sighting, aircraft, isNewFlight)
		rareOperatorFlag := db.updateOperator(sighting, aircraft, isNewFlight)
		rareCountryFlag := db.updateCountry(sighting, aircraft, isNewFlight)

		newRarities |= rareTypeFlag << 0
		newRarities |= rareOperatorFlag << 1
		newRarities |= rareCountryFlag << 2 //nolint:mnd // okay for bit shifting

		if newRarities != obs.NoRarity {
			rareSightings = append(rareSightings, obs.RareSighting{
				Rarities: newRarities,
				Sighting: sighting,
			})
		}

		// Finally, update the records
		sighting.Info = aircraft.AircraftToString()
		db.AircraftSightings[aircraft.Hex] = sighting
	}
	db.RareSightings = rareSightings
}

func (db *Dashboard) updateType(
	sighting *obs.AircraftSighting,
	aircraft *obs.AircraftRecord,
	isNewFlight bool,
) obs.RarityFlag {
	if sighting.TypeShort == "" && aircraft.Description != "" {
		sighting.TypeShort = aircraft.Description
	}

	// We already know the type or just saw this one recently, no need to update again.
	isTypeKnown := sighting.TypeDesc != obs.TypeUnknown
	isFlightKnown := !isNewFlight
	if isTypeKnown && isFlightKnown {
		aircraft.CachedType = sighting.TypeDesc
		return 0
	}

	// We couldn't find out the type of this aircraft, unable to update statistics.
	aType := db.IcaoToAircraft[aircraft.IcaoType].Make
	if aType == "" {
		return 0
	}

	sighting.TypeDesc = aType
	aircraft.CachedType = aType

	// Valid type found! Record type and update type rarities.
	thisTypeCountNew := db.SeenTypeCount[aType] + 1
	db.SeenTypeCount[aType] = thisTypeCountNew
	db.TotalTypeCount++
	rarityThreshold := math.Log(float64(db.TotalTypeCount)) - obs.RarityConstant
	isRareType := float64(thisTypeCountNew) < rarityThreshold

	// fmt.Println(
	//	"type rarity calculation: ",
	//	" aircraft Flight", aircraft.Flight,
	//	"type", sighting.typeDesc,
	//	"thisTypeCountNew", thisTypeCountNew,
	//	"totalTypeCount", db.totalTypeCount,
	//	"typeRarity", math.Log(float64(db.totalTypeCount))-5.0,
	//	"isRareType", isRareType)

	if !isRareType {
		return 0
	}

	// fmt.Println(
	//	"type rarity calculation: ",
	//	" aircraft Flight", aircraft.Flight,
	//	"type", sighting.typeDesc,
	//	"typeShort", sighting.typeShort,
	//	"thisTypeCountNew", thisTypeCountNew,
	//	"totalTypeCount", db.totalTypeCount,
	//	"typeRarity", rarityThreshold,
	//	"isRareType", isRareType)

	// db.logger.info(
	//	"type rarity calculation: ",
	//	"thisTypeCountNew", thisTypeCountNew,
	//	"totalTypeCount", db.totalTypeCount,
	//	"typeRarity", typeRarity,
	//	"typeRarityThreshold", typeRarityThreshold)

	return 1
}

func (db *Dashboard) updateOperator(
	sighting *obs.AircraftSighting,
	aircraft *obs.AircraftRecord,
	isNewFlight bool,
) obs.RarityFlag {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.Operator != obs.OperatorUnknown && !isNewFlight {
		return 0
	}

	flightNo := aircraft.GetFlightNoAsStr()
	if flightNo == "" {
		return 0
	}

	// First option: try to detect the airline and get operator & country from it.
	flightCode := aircraft.GetFlightNoAsIcaoCode()
	if flightCode != obs.FlightUnknownCode {
		if operatorRecord, opExists := db.IcaoToAirline[flightCode]; opExists {
			sighting.Operator = operatorRecord.Company
		}
	}

	// Unable to detect airline, maybe it's military or government.
	if sighting.Operator == obs.OperatorUnknown {
		if militaryOperator, milOpExists := db.MilCodeToOperator[flightCode]; milOpExists {
			sighting.Operator = militaryOperator
		}
	}

	// operator still not found, check whether the 'ownOp' field in the aircraft record is set.
	if sighting.Operator == obs.OperatorUnknown && aircraft.OwnOp != "" {
		sighting.Operator = aircraft.OwnOp
	}

	// Did not manage to find out the operator of this aircraft.
	if sighting.Operator == obs.OperatorUnknown {
		return 0
	}

	thisOperatorCountNew := db.SeenOperatorCount[sighting.Operator] + 1
	db.SeenOperatorCount[sighting.Operator] = thisOperatorCountNew
	db.TotalOperatorCount++
	rarityThreshold := math.Log(float64(db.TotalOperatorCount)) - obs.RarityConstant
	isRareOperator := float64(thisOperatorCountNew) < rarityThreshold

	// fmt.Println(
	//	"operator rarity calculation:",
	//	"operator", sighting.operator,
	//	"thisOperatorCountNew", thisOperatorCountNew,
	//	"totalOperatorCount", db.totalOperatorCount,
	//	"operatorRarity", math.Log(float64(db.totalOperatorCount))-5.0,
	//	"isRareOperator", isRareOperator)

	if !isRareOperator {
		return 0
	}

	// fmt.Println(
	//	"operator rarity calculation: ",
	//	"thisOperatorCountNew", thisOperatorCountNew,
	//	"totalOperatorCount", db.totalOperatorCount,
	//	"operatorRarity", rarityThreshold,
	//	"isRareOperator", isRareOperator)

	return 1
}

func (db *Dashboard) updateCountry(
	sighting *obs.AircraftSighting,
	aircraft *obs.AircraftRecord,
	isNewFlight bool,
) obs.RarityFlag {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.Country != obs.CountryUnknown && !isNewFlight {
		return 0
	}

	flightNo := aircraft.GetFlightNoAsStr()
	if flightNo == "" {
		return 0
	}

	// Option #1: Try to detect the airline and get operator & country from it.
	flightCode := aircraft.GetFlightNoAsIcaoCode()
	if flightCode != obs.FlightUnknownCode {
		if operatorRecord, exists := db.IcaoToAirline[flightCode]; exists {
			sighting.Country = strings.ToUpper(operatorRecord.Country)
		}
	}

	// Option #2: Detect country by the range of it's hex registration.
	if sighting.Country == obs.CountryUnknown {
		sighting.Country = strings.ToUpper(db.getCountryByHexRange(aircraft.Hex))
	}

	// Option #3: Detect country by its ICAO registration prefix.
	if sighting.Country == obs.CountryUnknown {
		if country, exists := db.getCountryByRegPrefix(aircraft.Registration); exists {
			sighting.Country = strings.ToUpper(country)
		}
	}

	// Unable to detect country of this aircraft.
	if sighting.Country == obs.CountryUnknown {
		return 0
	}

	thisCountryCountNew := db.SeenCountryCount[sighting.Country] + 1
	db.SeenCountryCount[sighting.Country] = thisCountryCountNew
	db.TotalCountryCount++
	rarityThreshold := math.Log(float64(db.TotalCountryCount)) - obs.RarityConstant
	isRareCountry := float64(thisCountryCountNew) < rarityThreshold

	// db.logger.Debug(
	//	"country rarity calculation:",
	//	"country", sighting.country,
	//	"thisCountryCountNew", thisCountryCountNew,
	//	"totalCountryCount", db.totalCountryCount,
	//	"countryRarity", countryRarity,
	//	"countryRarityThreshold", countryRarityThreshold)

	if !isRareCountry {
		return 0
	}

	// fmt.Println(
	//	"country rarity calculation: ",
	//	"thisCountryCountNew", thisCountryCountNew,
	//	"totalCountryCount", db.totalCountryCount,
	//	"countryRarity", rarityThreshold,
	//	"isRareCountry", isRareCountry)

	return 1
}

func (db *Dashboard) getCountryByHexRange(hexAsStr string) string {
	hexAsInt, err := strconv.ParseInt(hexAsStr, 16, 64)
	if err != nil {
		db.ErrOut.Printf("unable to convert hex to int: %s\n", hexAsStr)
		return obs.CountryUnknown
	}
	for key, value := range db.HexRangeToCountry {
		if hexAsInt > key.LowerBound && hexAsInt < key.UpperBound {
			return value
		}
	}
	return obs.CountryUnknown
}

func (db *Dashboard) getCountryByRegPrefix(reg string) (string, bool) {
	for key, value := range db.RegPrefixToCountry {
		if strings.Contains(reg, key) {
			return value, true
		}
	}

	return "", false
}

func (db *Dashboard) updateHighest(aircraft *obs.AircraftRecord) {
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

func (db *Dashboard) updateFastest(aircraft *obs.AircraftRecord) {
	if db.Fastest != nil && db.Fastest.GroundSpeed > aircraft.GroundSpeed {
		return
	}

	db.Fastest = aircraft
}

func (db *Dashboard) RecomputeFastestAndHighest() {
	db.Fastest = nil
	db.Highest = nil
	for idx := range db.CurrentAircraft {
		aircraft := &(db.CurrentAircraft)[idx]
		db.updateHighest(aircraft)
		db.updateFastest(aircraft)
	}
}

func (db *Dashboard) AssignRouteToCallsigns() []string {
	var callsignsWithoutRoute []string
	for _, sighting := range db.AircraftSightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		if sighting.Flightroute != nil {
			// A Flight route is already set.
			continue
		}

		if flightRoute, ok := db.CachedFlightRoutes[sighting.LastFlightNo]; ok {
			// Found a cached route for this Flight, reuse it!
			sighting.Flightroute = flightRoute
			continue
		}

		// No routes found, record this callsign to request route from adsbdb
		callsignsWithoutRoute = append(callsignsWithoutRoute, sighting.LastFlightNo)
	}
	return callsignsWithoutRoute
}

// AssignFlightRoutes assigns the given Flight routes to all flights matching the callsign.
func (db *Dashboard) AssignFlightRoutes(flightRouteRecords []ref.FlightRouteRecord) {
	for _, flightrouteRecord := range flightRouteRecords {
		callsign := flightrouteRecord.Callsign
		db.CachedFlightRoutes[callsign] = &flightrouteRecord
	}
	for _, sighting := range db.AircraftSightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		if sighting.Flightroute != nil {
			// A Flight route is already set.
			continue
		}

		if flightRoute, ok := db.CachedFlightRoutes[sighting.LastFlightNo]; ok {
			// Found a cached route for this Flight, reuse it!
			sighting.Flightroute = flightRoute
			continue
		}

		// Route cannot be found: use a dummy and also cache a dummy to prevent unnecessary requests
		// for the same callsign again.
		sighting.Flightroute = ref.GetDefaultFlightrouteRecord()
		db.CachedFlightRoutes[sighting.LastFlightNo] = ref.GetDefaultFlightrouteRecord()
	}
}
