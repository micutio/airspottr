package observation

import (
	"math"
	"strings"
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// FlightIdentity describes how a new ADS-B callsign relates to a stored sighting.
type FlightIdentity struct {
	Identified bool // Unknown sighting matches new record with callsign
	Updated    bool // Known sighting matches new record with new data
	NewFlight  bool // No previously known sighting matches the new record
}

// IdentifyFlight compares previous and current flight numbers.
// An ICAO hex code seen for the first-time is always a new flight.
// Identifying an unknown callsign is not a new flight,
// but changing from one known callsign to another is.
// isNewSighting is `true` if no sighting of thisFlight has ocurred before.
func IdentifyFlight(previousFlight, thisFlight string, isNewSighting bool) FlightIdentity {
	isPreviousFlightKnown := previousFlight != FlightUnknown
	isThisFlightKnown := thisFlight != FlightUnknown

	updated := isPreviousFlightKnown &&
		isThisFlightKnown &&
		previousFlight != thisFlight
	return FlightIdentity{
		Identified: !isPreviousFlightKnown && isThisFlightKnown,
		Updated:    updated,
		NewFlight:  isNewSighting || updated,
	}
}

func isRare(thisCount, total int) bool {
	threshold := math.Log(float64(total)) - RarityConstant
	return float64(thisCount) < threshold
}

// EvaluateBatch applies flight identity, classification, rarity, and
// fastest/highest tracking to a batch of aircraft records.
// It mutates state in place and returns the current-batch sightings.
func EvaluateBatch(
	state *State,
	records []AircraftRecord,
	classifier Classifier,
	now time.Time,
) []AircraftSighting {
	state.EnsureInitialized()
	currentSightings := make([]AircraftSighting, len(records))
	observer := state.Observer

	for idx := range records {
		aircraft := records[idx]
		sighting, exists := state.Sightings[aircraft.Hex]
		isNew := !exists
		if isNew {
			sighting = NewSighting(observer.Latitude, observer.Longitude, aircraft, now)
		} else {
			sighting.LastRecord = aircraft
		}
		if sighting.Registration == "" {
			sighting.Registration = aircraft.Registration
		}

		thisFlightNo := aircraft.GetCallsignAsStr()
		identity := IdentifyFlight(sighting.LastFlightNo, thisFlightNo, isNew)
		if identity.Identified || identity.Updated {
			sighting.LastFlightNo = thisFlightNo
		}
		if identity.Updated {
			sighting.Flightroute = ref.FlightrouteRecord{} //nolint:exhaustruct_v5 // reset to unassigned
		}

		acPos := ref.NewCoordinates(aircraft.Lat, aircraft.Lon)
		sighting.Distance = ref.Distance(observer, acPos).Kilometers()

		state.Highest = updateHighest(state.Highest, sighting)
		state.Fastest = updateFastest(state.Fastest, sighting)

		newRarities := NoRarity
		newRarities |= classifyType(state, &sighting, &aircraft, identity.NewFlight, classifier)
		newRarities |= classifyOperator(state, &sighting, &aircraft, identity.NewFlight, classifier)
		newRarities |= classifyCountry(state, &sighting, &aircraft, identity.NewFlight, classifier)
		sighting.Rarities = newRarities

		sighting.Info = sighting.SightingToString()
		currentSightings[idx] = sighting
		state.Sightings[aircraft.Hex] = sighting
	}

	return currentSightings
}

func classifyType(
	state *State,
	sighting *AircraftSighting,
	aircraft *AircraftRecord,
	isNewFlight bool,
	classifier Classifier,
) RarityFlag {
	if sighting.TypeShort == "" && aircraft.Description != "" {
		sighting.TypeShort = aircraft.Description
	}

	isTypeKnown := sighting.TypeDesc != TypeUnknown
	if isTypeKnown && !isNewFlight {
		return NoRarity
	}

	aType, exists := classifier.AircraftMake(aircraft.IcaoType)
	if !exists {
		return NoRarity
	}

	sighting.TypeDesc = aType
	thisTypeCountNew := state.RecordType(aType)
	if !isRare(thisTypeCountNew, state.SightedTypes) {
		return NoRarity
	}
	return RareType
}

func classifyOperator(
	state *State,
	sighting *AircraftSighting,
	aircraft *AircraftRecord,
	isNewFlight bool,
	classifier Classifier,
) RarityFlag {
	if sighting.Operator != OperatorUnknown && !isNewFlight {
		return NoRarity
	}

	flightNo := aircraft.GetCallsignAsStr()
	if flightNo == "" {
		return NoRarity
	}

	flightCode := aircraft.GetCallsignAsIcaoCode()
	if flightCode != FlightUnknownCode {
		if company, _, opExists := classifier.OperatorByIcao(flightCode); opExists {
			sighting.Operator = company
		}
	}

	if sighting.Operator == OperatorUnknown {
		if militaryOperator, milOpExists := classifier.OperatorByMil(flightCode); milOpExists {
			sighting.Operator = militaryOperator
		}
	}

	if sighting.Operator == OperatorUnknown && aircraft.OwnOp != "" {
		sighting.Operator = aircraft.OwnOp
	}

	if sighting.Operator == OperatorUnknown {
		return NoRarity
	}

	thisOperatorCountNew := state.RecordOperator(sighting.Operator)
	if !isRare(thisOperatorCountNew, state.SightedOperators) {
		return NoRarity
	}
	return RareOperator
}

func classifyCountry(
	state *State,
	sighting *AircraftSighting,
	aircraft *AircraftRecord,
	isNewFlight bool,
	classifier Classifier,
) RarityFlag {
	if sighting.Country != ref.CountryUnknown && !isNewFlight {
		return NoRarity
	}

	flightNo := aircraft.GetCallsignAsStr()
	if flightNo == "" {
		return NoRarity
	}

	flightCode := aircraft.GetCallsignAsIcaoCode()
	if flightCode != FlightUnknownCode {
		if _, country, exists := classifier.OperatorByIcao(flightCode); exists {
			sighting.Country = strings.ToUpper(country)
		}
	}

	if sighting.Country == ref.CountryUnknown {
		if country, ok := classifier.CountryByHex(aircraft.Hex); ok {
			sighting.Country = strings.ToUpper(country)
		}
	}

	if sighting.Country == ref.CountryUnknown {
		if country, exists := classifier.CountryByRegistration(aircraft.Registration); exists {
			sighting.Country = strings.ToUpper(country)
		}
	}

	if sighting.Country == ref.CountryUnknown {
		return NoRarity
	}

	thisCountryCountNew := state.RecordCountry(sighting.Country)
	if !isRare(thisCountryCountNew, state.SightedCountries) {
		return NoRarity
	}
	return RareCountry
}

func updateHighest(current, sighting AircraftSighting) AircraftSighting {
	thisAltitude, thisAltOk := sighting.LastRecord.AltBaro.(float64)
	if !thisAltOk {
		return current
	}

	highestAltitude, highestAltOk := current.LastRecord.AltBaro.(float64)
	if highestAltOk && highestAltitude > thisAltitude {
		return current
	}

	return sighting
}

func updateFastest(current, sighting AircraftSighting) AircraftSighting {
	if current.LastRecord.GroundSpeed > sighting.LastRecord.GroundSpeed {
		return current
	}
	return sighting
}
