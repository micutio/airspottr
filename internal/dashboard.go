// Package internal provides the Dashboard type and all associated program logic.
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	set "github.com/deckarep/golang-set/v2"
	"github.com/gen2brain/beeep"
)

const (
	// Lat is Latitude of SIN Airport.
	Lat float64 = 1.359297
	// Lon is Longitude of SIN Airport.
	Lon float64 = 103.989348
	// appIconPath is the file path to the icon png for this application.
	appIconPath = "./assets/icon.png"
	// aircraftRarityThreshold denotes the maximum rate an aircraft type is seen to be considered rare.
	aircraftRarityThreshold = 0.02
	// altitudeUnknown is what we use for aircraft without a given altitude.
	altitudeUnknown = "  n/a"
	// flightUnknown is what we use for aircraft with missing flight number.
	// Note: we're adding space at the end to have a length that is consistent with ICAO codes.
	flightUnknown = "unknown "
	// typeUnknown is what we use for aircraft with a type that's either empty or can't be found.
	typeUnknown = "unknown"
)

// Errors used by the Dashboard.
var (
	errParseIcaoMap    = errors.New("failed to parse ICAO to aircraft map")
	errParseMilCodeMap = errors.New("failed to parse mil code to operator map")
)

// TODO: change seenAircraft to map [string -> timestamp] to record the last sighting

type Dashboard struct {
	fastest           *aircraftRecord
	highest           *aircraftRecord
	isWarmup          bool
	seenAircraft      set.Set[string] // set of all seen aircraft, mapped to their type
	seenTypeCount     map[string]int  // types mapped to how often seen
	totalTypeCount    int
	milCodeToOperator map[string]string
	icaoToAircraft    map[string]icaoAircraft
	logger            slog.Logger
}

func NewDashboard() (*Dashboard, error) {
	icaoToAircraftMap, icaoErr := getIcaoToAircraftMap()
	if icaoErr != nil {
		return nil, fmt.Errorf("newDashboard: %w caused by %w", errParseIcaoMap, icaoErr)
	}

	milCodeToOperatorMap, milCodeErr := getMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, errParseMilCodeMap
	}

	dash := Dashboard{
		fastest:           nil,
		highest:           nil,
		isWarmup:          true,
		seenAircraft:      set.NewSet[string](),
		seenTypeCount:     make(map[string]int),
		totalTypeCount:    0,
		icaoToAircraft:    icaoToAircraftMap,
		milCodeToOperator: milCodeToOperatorMap,
		logger:            *slog.Default(),
	}

	return &dash, nil
}

func (db *Dashboard) EndWarmupPeriod() {
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

	db.processCivAircraftRecords(&data.Aircraft)
}

func (db *Dashboard) processCivAircraftRecords(allAircraft *[]aircraftRecord) {
	sort.Sort(ByFlight(*allAircraft))

	for i := range len(*allAircraft) {
		aircraft := (*allAircraft)[i]

		db.checkHighest(&aircraft)
		db.checkFastest(&aircraft)

		isNewAircraft := db.seenAircraft.Add(aircraft.Hex)
		if !isNewAircraft {
			return
		}

		aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode
		if aType == "" {
			aType = typeUnknown
		}

		// TODO: Throw away aircraft with empty registration or type \"\"

		currentCount := db.seenTypeCount[aType]
		newCount := currentCount + 1
		db.seenTypeCount[aType] = newCount
		db.totalTypeCount++
		aircraftRarity := float32(newCount) / float32(db.totalTypeCount)

		// TODO: Define a good rarity metric.
		if !db.isWarmup && aircraftRarity < aircraftRarityThreshold {
			db.logger.Info("found rare", "aircraft", db.aircraftToString(&aircraft))
			db.notifyRareAircraft(&aircraft)
		}
	}
}

func (db *Dashboard) notifyRareAircraft(aircraft *aircraftRecord) {
	aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode

	msgBody := fmt.Sprintf("%s (%s)", aType, aircraft.Registration)
	err := beeep.Notify("Rare Aircraft Detected", msgBody, appIconPath)
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
	if db.highest != nil && db.highest.AltBaro.(float64) > thisAltitude {
		return
	}

	db.highest = aircraft
}

func (db *Dashboard) checkFastest(aircraft *aircraftRecord) {
	if db.fastest != nil && db.fastest.GroundSpeed > aircraft.GroundSpeed {
		return
	}

	db.fastest = aircraft
}

type aircraftTypeCountTuple struct {
	acType string
	count  int
}

type ByCount []aircraftTypeCountTuple

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].count < a[j].count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// PrintSummary prints the highest, fastest and the most and the least common types.
func (db *Dashboard) PrintSummary() {
	fmt.Println("=== Summary ===")
	db.listTypesByRarity()
	db.logger.Info("Fastest Aircraft:")
	db.logger.Info(db.aircraftToString(db.fastest))
	db.logger.Info("Highest Aircraft:")
	db.logger.Info(db.aircraftToString(db.highest))
	fmt.Println("=== End Summary ===")
}

func (db *Dashboard) listTypesByRarity() {
	typeCounts := make([]aircraftTypeCountTuple, len(db.seenTypeCount))
	i := 0
	for key, value := range db.seenTypeCount {
		typeCounts[i] = aircraftTypeCountTuple{acType: key, count: value}
		i++
	}

	sort.Sort(ByCount(typeCounts))

	db.logger.Info("aircraft types from least to most common")
	for i := range typeCounts {
		db.logger.Info(fmt.Sprintf("%6d - %q\n", typeCounts[i].count, typeCounts[i].acType))
	}
}

// aircraftToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func (db *Dashboard) aircraftToString(aircraft *aircraftRecord) string {
	thisPos := newCoordinates(Lat, Lon)
	acPos := newCoordinates(aircraft.Lat, aircraft.Lon)
	aircraft.CachedDist = Distance(thisPos, acPos).Kilometers()

	flight := aircraft.Flight

	if len(flight) == 0 {
		flight = flightUnknown
	}

	altitude := getAltitudeAsString(aircraft.AltBaro)
	aType := db.icaoToAircraft[aircraft.IcaoType].ModelCode

	return fmt.Sprintf("FLT %s DST %4.0f km ALT %s SPD %3.0f HDG %3.0f ID %q (%s)\n",
		flight,
		aircraft.CachedDist,
		altitude,
		aircraft.GroundSpeed,
		aircraft.NavHeading,
		aType,
		aircraft.Registration)
}

// getAltitudeAsString reads the altitude of an aircraft and returns it as a string.
// The altitude is stored either as a string 'ground' or as a float denoting the measured
// barometric altitude.
// If the latter is the case, the float will be formatted without any decimal places
// (unnecessary accuracy) and converted to string.
func getAltitudeAsString(altBaro any) string {
	if num, numOk := altBaro.(float64); numOk {
		return fmt.Sprintf("%5.0f", num)
	}

	if str, strOk := altBaro.(string); strOk {
		return str
	}

	return altitudeUnknown
}
