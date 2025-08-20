// Package internal provides the Dashboard type and all associated program logic.
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
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
	typeRarityThreshold = 0.01
	// airlineRarityThreshold denotes the maximum rate an airline is seen to be considered rare.
	airlineRarityThreshold = 0.01
	// altitudeUnknown is what we use for aircraft without a given altitude.
	altitudeUnknown = "  n/a"
	// flightUnknown is what we use for aircraft with missing flight number.
	// Note: we're adding space at the end to have a length that is consistent with ICAO codes.
	flightUnknown = "unknown "
	// flightUnknownCode is the three letter code we use for aircraft with missing flight number.
	flightUnknownCode = "n/a"
	// typeUnknown is what we use for aircraft with a type that's either empty or can't be found.
	typeUnknown = "unknown"
)

// Errors used by the Dashboard.
var (
	errParseIcaoAircraftMap = errors.New("failed to parse ICAO to aircraft map")
	errParseIcaoAirlineMap  = errors.New("failed to parse 3LTR to airline map")
	errParseMilCodeMap      = errors.New("failed to parse mil code to operator map")
)

type Dashboard struct {
	isWarmup          bool
	totalTypeCount    int
	totalAirlineCount int
	Fastest           *aircraftRecord
	Highest           *aircraftRecord
	CurrentAircraft   []aircraftRecord
	IcaoToAircraft    map[string]icaoAircraft
	IcaoToAirline     map[string]icaoAirline
	seenAircraft      map[string]time.Time // set of all seen aircraft, mapped to last seen time
	seenTypeCount     map[string]int       // types mapped to how often seen
	seenAirlineCount  map[string]int       // types mapped to how often seen
	milCodeToOperator map[string]string
	logger            slog.Logger
}

func NewDashboard() (*Dashboard, error) {
	icaoToAircraftMap, aircraftErr := getIcaoToAircraftMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf("newDashboard: %w caused by %w", errParseIcaoAircraftMap, aircraftErr)
	}

	icaoToAirlineMap, airlineErr := getIcaoToAirlineMap()
	if airlineErr != nil {
		return nil, fmt.Errorf("newDashboard: %w caused by %w", errParseIcaoAirlineMap, airlineErr)
	}

	milCodeToOperatorMap, milCodeErr := getMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, errParseMilCodeMap
	}

	dash := Dashboard{
		isWarmup:          true,
		totalTypeCount:    0,
		totalAirlineCount: 0,
		Fastest:           nil,
		Highest:           nil,
		CurrentAircraft:   nil,
		IcaoToAircraft:    icaoToAircraftMap,
		IcaoToAirline:     icaoToAirlineMap,
		seenAircraft:      make(map[string]time.Time),
		seenTypeCount:     make(map[string]int),
		seenAirlineCount:  make(map[string]int),
		milCodeToOperator: milCodeToOperatorMap,
		logger:            *slog.Default(),
	}

	return &dash, nil
}

func (db *Dashboard) FinishWarmupPeriod() {
	db.isWarmup = false
}

// ProcessCivAircraftJSON takes a json record in form of a byte array, transforms it into a list
// of aircraft and performs some processing thereafter.
func (db *Dashboard) ProcessCivAircraftJSON(jsonBytes []byte) {
	var data civAircraftResult
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		db.logger.Error("dashboard:", slog.Any("failed to unmarshal Json", err))
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

	for i := range len(db.CurrentAircraft) {
		aircraft := (db.CurrentAircraft)[i]

		db.checkHighest(&aircraft)
		db.checkFastest(&aircraft)

		db.logger.Debug("debug", "aircraft.seen", aircraft.Seen)
		lastSeenMsBeforeNow := time.Duration(aircraft.Seen) * time.Second
		lastSeenTime := time.Now().Add(-lastSeenMsBeforeNow)
		db.logger.Debug("debug", "lastSeenTime", lastSeenTime)

		value, exists := db.seenAircraft[aircraft.Hex]
		db.seenAircraft[aircraft.Hex] = lastSeenTime

		// If the aircraft is new or has already been spotted, but enough time has passed,
		// then we can report it again.
		if exists && lastSeenTime.Sub(value) < time.Hour*24 {
			// Aircraft has already been spotted recently (within the last day).
			// No need to report this sighting again.
			// NOTE: Alternatively we might check for different flight number as well.
			return
		}

		aType := db.IcaoToAircraft[aircraft.IcaoType].ModelCode
		if aType == "" {
			aType = typeUnknown
		}

		// Ignore aircraft with empty registration
		if aircraft.Registration == "" {
			continue
		}

		// record type and update type rarity
		thisTypeCountCurrent := db.seenTypeCount[aType]
		thisTypeCountNew := thisTypeCountCurrent + 1
		db.seenTypeCount[aType] = thisTypeCountNew
		db.totalTypeCount++
		typeRarity := float32(thisTypeCountNew) / float32(db.totalTypeCount)

		db.logger.Debug(
			"type rarity calculation: ",
			"thisTypeCountNew", thisTypeCountNew,
			"totalTypeCount", db.totalTypeCount,
			"typeRarity", typeRarity,
			"typeRarityThreshold", typeRarityThreshold)

		if typeRarity < typeRarityThreshold {
			db.logger.Info(
				"type rarity calculation: ",
				"thisTypeCountNew", thisTypeCountNew,
				"totalTypeCount", db.totalTypeCount,
				"typeRarity", typeRarity,
				"typeRarityThreshold", typeRarityThreshold)
			db.logger.Info(
				"found rare",
				"type",
				db.aircraftToString(&aircraft),
				"Δt(ms)",
				time.Since(lastSeenTime))

			if !db.isWarmup {
				db.notifyRareType(&aircraft)
			}
		}

		// record airline and update airline rarity
		// TODO: Also look into military operators and ownOp field
		// TODO: Rename airline to operator
		// Goal: detect operator and country
		// Strategy:
		//	1. look a flight number, try to extract icao code and look up airline, like it's done already
		//  2. get country from hex value
		//  3. get country from registration prefix
		//  4. get operator from mil-icao-operator lookup, NOTE: operator code has variable length!
		//  5. get operator from ownOp field
		airlineCode := aircraft.GetFlightNoAsIcaoCode()
		if airlineCode == flightUnknownCode {
			continue
		}

		airline := db.IcaoToAirline[airlineCode]
		if airline.Company == "" {
			continue
		}

		thisAirlineCountCurrent := db.seenAirlineCount[airlineCode]
		thisAirlineCountNew := thisAirlineCountCurrent + 1
		db.seenAirlineCount[airlineCode] = thisAirlineCountNew
		db.totalAirlineCount++
		airlineRarity := float32(thisAirlineCountNew) / float32(db.totalAirlineCount)

		db.logger.Debug(
			"airline rarity calculation: ",
			"thisAirlineCountNew", thisAirlineCountNew,
			"totalAirlineCount", db.totalAirlineCount,
			"airlineRarity", airlineRarity,
			"airlineRarityThreshold", airlineRarityThreshold)

		if airlineRarity < airlineRarityThreshold {
			db.logger.Info(
				"airline rarity calculation: ",
				"thisAirlineCountNew", thisAirlineCountNew,
				"totalAirlineCount", db.totalAirlineCount,
				"airlineRarity", airlineRarity,
				"airlineRarityThreshold", airlineRarityThreshold)
			db.logger.Info(
				"found rare",
				"airline",
				airline.Company,
				"country",
				airline.Country,
				"Δt(ms)",
				time.Since(lastSeenTime))

			if !db.isWarmup {
				db.notifyRareAirline(&aircraft)
			}
		}
	}
}

func (db *Dashboard) notifyRareType(aircraft *aircraftRecord) {
	aType := db.IcaoToAircraft[aircraft.IcaoType].ModelCode

	msgBody := fmt.Sprintf("%s (%s)", aType, aircraft.Registration)
	err := beeep.Notify("Rare Aircraft Type Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func (db *Dashboard) notifyRareAirline(aircraft *aircraftRecord) {
	airline := db.IcaoToAirline[aircraft.GetFlightNoAsIcaoCode()]

	msgBody := fmt.Sprintf("%s (%s)", airline.Company, airline.Country)
	err := beeep.Notify("Rare Aircraft Type Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func (db *Dashboard) ProcessMilAircraftJSON(jsonBytes []byte) {
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

	db.logger.Info("Military aircraft in increasing distance from here:\n")

	for i := range len(*allAircraft) {
		aircraft := (*allAircraft)[i]
		if (aircraft.Lat == 0 && aircraft.Lon == 0) || aircraft.CachedDist > 1000.0 {
			continue
		}

		db.logger.Info(db.aircraftToString(&aircraft))
	}
}

func (db *Dashboard) checkHighest(aircraft *aircraftRecord) {
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

func (db *Dashboard) checkFastest(aircraft *aircraftRecord) {
	if db.Fastest != nil && db.Fastest.GroundSpeed > aircraft.GroundSpeed {
		return
	}

	db.Fastest = aircraft
}

type aircraftTypeCountTuple struct {
	acType string
	count  int
}

type ByTypeCount []aircraftTypeCountTuple

func (a ByTypeCount) Len() int           { return len(a) }
func (a ByTypeCount) Less(i, j int) bool { return a[i].count < a[j].count }
func (a ByTypeCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type airlineCountTuple struct {
	airline string
	count   int
}

type ByAirlineCount []airlineCountTuple

func (a ByAirlineCount) Len() int           { return len(a) }
func (a ByAirlineCount) Less(i, j int) bool { return a[i].count < a[j].count }
func (a ByAirlineCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// PrintSummary prints the highest, fastest and the most and the least common types.
func (db *Dashboard) PrintSummary() {
	fmt.Println("=== Summary ===")
	db.listTypesByRarity()
	db.listAirlineByRarity()
	db.logger.Info("Fastest Aircraft:")
	db.logger.Info(db.aircraftToString(db.Fastest))
	db.logger.Info("Highest Aircraft:")
	db.logger.Info(db.aircraftToString(db.Highest))
	fmt.Println("=== End Summary ===")
}

func (db *Dashboard) listTypesByRarity() {
	typeCounts := make([]aircraftTypeCountTuple, len(db.seenTypeCount))
	i := 0
	for key, value := range db.seenTypeCount {
		typeCounts[i] = aircraftTypeCountTuple{acType: key, count: value}
		i++
	}

	sort.Sort(ByTypeCount(typeCounts))

	db.logger.Info("aircraft types from least to most common")
	for j := range typeCounts {
		db.logger.Info(fmt.Sprintf("%6d - %q\n", typeCounts[j].count, typeCounts[j].acType))
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

	return fmt.Sprintf("FNO %s DST %4.0f km ALT %s SPD %3.0f HDG %3.0f TID %q (%s)\n",
		flight,
		aircraft.CachedDist,
		altitude,
		aircraft.GroundSpeed,
		aircraft.NavHeading,
		aType,
		aircraft.Registration)
}

func (db *Dashboard) listAirlineByRarity() {
	airlineCounts := make([]airlineCountTuple, len(db.seenAirlineCount))
	i := 0
	for key, value := range db.seenAirlineCount {
		operator := db.IcaoToAirline[key].Company
		airlineCounts[i] = airlineCountTuple{airline: operator, count: value}
		i++
	}

	sort.Sort(ByAirlineCount(airlineCounts))

	db.logger.Info("airline from least to most common")
	for j := range airlineCounts {
		db.logger.Info(fmt.Sprintf("%6d - %q\n", airlineCounts[j].count, airlineCounts[j].airline))
	}
}
