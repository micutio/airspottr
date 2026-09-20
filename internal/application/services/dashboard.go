// Package services provides the Dashboard type and all associated program logic.
package services

import (
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"math"
	"strings"
	"time"

	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	rep "github.com/micutio/airspottr/internal/domain/repositories"
)

// TODO: Remove and privatise as many fields as possible.
// TODO: Create specialized type for highest and fastest.

type Dashboard struct {
	IsWarmup              bool
	Lat                   float64
	Lon                   float64
	Fastest               obs.AircraftRecord
	Highest               obs.AircraftRecord
	CurrentSightings      []obs.AircraftSighting
	Sightings             map[string]obs.AircraftSighting
	SightedTypesCount     int            // total count of unique sighted types
	SightedOperatorsCount int            // total count of unique sighted operators
	SightedCountriesCount int            // total count of unique sighted countries of origin
	SeenTypeCount         map[string]int // types mapped to how often seen
	SeenOperatorCount     map[string]int // airlines mapped to how often seen
	SeenCountryCount      map[string]int // airlines mapped to how often seen
	ErrOut                log.Logger
}

func NewDashboard(lat float64, lon float64, stderr *io.Writer) *Dashboard {
	dashboard := Dashboard{
		IsWarmup:              true,
		Lat:                   lat,
		Lon:                   lon,
		Fastest:               obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
		Highest:               obs.AircraftRecord{}, //nolint:exhaustruct_v5 // using default values
		CurrentSightings:      nil,
		Sightings:             make(map[string]obs.AircraftSighting),
		SightedTypesCount:     0,
		SightedOperatorsCount: 0,
		SightedCountriesCount: 0,
		SeenTypeCount:         make(map[string]int),
		SeenOperatorCount:     make(map[string]int),
		SeenCountryCount:      make(map[string]int),
		ErrOut:                *log.New(*stderr, "dashboard ", log.LstdFlags),
	}

	dashboard.ErrOut.Println("Dashboard init")

	return &dashboard
}

func (db *Dashboard) FinishWarmupPeriod() {
	db.IsWarmup = false
}

//////////////////////////////////////////////////////////////////////////////
/// Processing of all aircraft: civilian, military, government, private.    //
//////////////////////////////////////////////////////////////////////////////

// ProcessAircraftRecords takes currently observed aircraft messages to update
// sightings and determine sighting rarity.
// If an aircraft has not been recorded before -> create a new sighting.
// If an aircraft has been recorded before on a different flight -> create a new sighting.
// If an aircraft has been recorded before on the same flight -> update existing sighting.
// If a sighting contains either a type, operator or country of origin that
// has been counted below a certain threshold, then this sighting is now
// considered rare and can be used to emit notifications to the user.
func (db *Dashboard) ProcessAircraftRecords(
	aircraftSpecRepo rep.AircraftTypeRepo,
	operatorRepo rep.OperatorRepository,
	countryRepo rep.CountryRepository,
	aircraftRecords []obs.AircraftRecord,
) {
	thisPos := ref.NewCoordinates(db.Lat, db.Lon)
	currentSightings := make([]obs.AircraftSighting, len(aircraftRecords))

	for idx := range aircraftRecords {
		// Get aircraft and time of sighting
		aircraft := aircraftRecords[idx]
		lastSeenMsBeforeNow := time.Duration(aircraft.Seen) * time.Second
		lastSeenTime := time.Now().Add(-lastSeenMsBeforeNow)

		// Retrieve previous sighting or create new one.
		sighting, exists := db.Sightings[aircraft.Hex]
		if !exists {
			sighting = obs.AircraftSighting{
				Rarities:     obs.NoRarity,
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
				Country:      ref.CountryUnknown,
				Info:         "",
				Flightroute:  ref.FlightrouteRecord{}, //nolint:exhaustruct_v5 // using default values
				LastRecord:   aircraft,
			}
		} else {
			sighting.LastRecord = aircraft
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
		dist := ref.Distance(thisPos, acPos).Kilometers()
		sighting.Distance = dist

		// Update all aircraft, type, operator and country statistics
		db.updateHighest(aircraft)
		db.updateFastest(aircraft)

		newRarities := obs.NoRarity
		rareTypeFlag := db.updateType(aircraftSpecRepo, &sighting, &aircraft, isNewFlight)
		rareOperatorFlag := db.updateOperator(operatorRepo, &sighting, &aircraft, isNewFlight)
		rareCountryFlag := db.updateCountry(
			operatorRepo, countryRepo, &sighting, &aircraft, isNewFlight)

		newRarities |= rareTypeFlag << 0
		newRarities |= rareOperatorFlag << 1
		newRarities |= rareCountryFlag << 2 //nolint:mnd // okay for bit shifting

		sighting.Rarities = newRarities
		currentSightings[idx] = sighting

		// Finally, update the records
		sighting.Info = sighting.SightingToString()
		db.Sightings[aircraft.Hex] = sighting
	}
	db.CurrentSightings = currentSightings
}

func (db *Dashboard) updateType(
	aircraftSpecRepo rep.AircraftTypeRepo,
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
		return 0
	}

	// We couldn't find out the type of this aircraft, unable to update statistics.
	spec, exists := aircraftSpecRepo.GetAircraftType(aircraft.IcaoType)
	if !exists {
		return 0
	}

	aType := spec.Make

	sighting.TypeDesc = aType

	// Valid type found! Record type and update type rarities.
	thisTypeCountNew := db.SeenTypeCount[aType] + 1
	db.SeenTypeCount[aType] = thisTypeCountNew
	db.SightedTypesCount++
	rarityThreshold := math.Log(float64(db.SightedTypesCount)) - obs.RarityConstant
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
	operatorRepo rep.OperatorRepository,
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
		if operatorRecord, opExists := operatorRepo.GetOperatorByIcao(flightCode); opExists {
			sighting.Operator = operatorRecord.Company
		}
	}

	// Unable to detect airline, maybe it's military or government.
	if sighting.Operator == obs.OperatorUnknown {
		if militaryOperator, milOpExists := operatorRepo.GetOperatorByMilCode(flightCode); milOpExists {
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
	db.SightedOperatorsCount++
	rarityThreshold := math.Log(float64(db.SightedOperatorsCount)) - obs.RarityConstant
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
	operatorRepo rep.OperatorRepository,
	countryRepo rep.CountryRepository,
	sighting *obs.AircraftSighting,
	aircraft *obs.AircraftRecord,
	isNewFlight bool,
) obs.RarityFlag {
	// We already know the type or just saw this one recently, no need to update again.
	if sighting.Country != ref.CountryUnknown && !isNewFlight {
		return 0
	}

	flightNo := aircraft.GetFlightNoAsStr()
	if flightNo == "" {
		return 0
	}

	// Option #1: Try to detect the airline and get operator & country from it.
	flightCode := aircraft.GetFlightNoAsIcaoCode()
	if flightCode != obs.FlightUnknownCode {
		if operatorRecord, exists := operatorRepo.GetOperatorByIcao(flightCode); exists {
			sighting.Country = strings.ToUpper(operatorRecord.Country)
		}
	}

	// Option #2: Detect country by the range of it's hex registration.
	if sighting.Country == ref.CountryUnknown {
		country, countryErr := countryRepo.GetCountryByHexCode(aircraft.Hex)
		if countryErr != nil {
			db.ErrOut.Printf("warning: invalid hex code: %v", countryErr)
		} else {
			sighting.Country = strings.ToUpper(country)
		}
	}

	// Option #3: Detect country by its ICAO registration prefix.
	if sighting.Country == ref.CountryUnknown {
		if country, exists := countryRepo.GetCountryByRegistration(aircraft.Registration); exists {
			sighting.Country = strings.ToUpper(country)
		}
	}

	// Unable to detect country of this aircraft.
	if sighting.Country == ref.CountryUnknown {
		return 0
	}

	thisCountryCountNew := db.SeenCountryCount[sighting.Country] + 1
	db.SeenCountryCount[sighting.Country] = thisCountryCountNew
	db.SightedCountriesCount++
	rarityThreshold := math.Log(float64(db.SightedCountriesCount)) - obs.RarityConstant
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

func (db *Dashboard) updateHighest(aircraft obs.AircraftRecord) {
	thisAltitude, thisAltOk := aircraft.AltBaro.(float64)
	if !thisAltOk {
		return
	}

	highestAltitude, highestAltOk := aircraft.AltBaro.(float64)
	if !highestAltOk || highestAltitude > thisAltitude {
		return
	}

	db.Highest = aircraft
}

func (db *Dashboard) updateFastest(aircraft obs.AircraftRecord) {
	if db.Fastest.GroundSpeed > aircraft.GroundSpeed {
		return
	}

	db.Fastest = aircraft
}

// GetCallsignsRequiringRoutes attempts to assign cached route information to all
// sightings.
// It returns a list of callsigns without known routes, to allow querying
// online for these cases.
func (db *Dashboard) GetCallsignsRequiringRoutes() []string {
	defaultFlightrouteRecord := *ref.GetDefaultFlightrouteRecord()
	var callsignsWithoutRoute []string
	for _, sighting := range db.CurrentSightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		// TODO: Verify struct comparison.
		if sighting.Flightroute != defaultFlightrouteRecord {
			// A Flight route is already set.
			continue
		}

		// No routes found, record this callsign to request route from adsbdb
		callsignsWithoutRoute = append(callsignsWithoutRoute, sighting.LastFlightNo)
	}
	return callsignsWithoutRoute
}

// AssignFlightRoutes assigns the given Flight routes to all flights matching the callsign.
func (db *Dashboard) AssignFlightRoutes(flightRouteRecords map[string]ref.FlightrouteRecord) {
	defaultFlightrouteRecord := *ref.GetDefaultFlightrouteRecord()
	for _, sighting := range db.Sightings {
		if sighting.LastFlightNo == obs.FlightUnknown {
			// Can't get Flight routes for unknown Flight.
			continue
		}

		if sighting.Flightroute != defaultFlightrouteRecord {
			// A Flight route is already set.
			continue
		}

		if flightRoute, ok := flightRouteRecords[sighting.LastFlightNo]; ok {
			sighting.Flightroute = flightRoute
			continue
		}
	}
}
