package observation

import (
	"math"
	"strings"
	"time"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

// FlightIdentity describes how a new ADS-B flight number relates to a stored sighting.
type FlightIdentity struct {
	Identified bool
	Updated    bool
	NewFlight  bool
}

// IdentifyFlight compares previous and current flight numbers.
// A first-time hex is always a new flight. Identifying an unknown callsign
// is not a new statistical flight; changing from one known callsign to another is.
func IdentifyFlight(previousFlight, thisFlight string, isNewSighting bool) FlightIdentity {
	identified := previousFlight == FlightUnknown && thisFlight != FlightUnknown
	updated := previousFlight != FlightUnknown &&
		thisFlight != FlightUnknown &&
		previousFlight != thisFlight
	return FlightIdentity{
		Identified: identified,
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
func EvaluateBatch(
	state State,
	records []AircraftRecord,
	classifier Classifier,
	now time.Time,
) (State, []AircraftSighting) {
	state = cloneState(state)
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

		thisFlightNo := aircraft.GetFlightNoAsStr()
		identity := IdentifyFlight(sighting.LastFlightNo, thisFlightNo, isNew)
		if identity.Identified || identity.Updated {
			sighting.LastFlightNo = thisFlightNo
		}

		acPos := ref.NewCoordinates(aircraft.Lat, aircraft.Lon)
		sighting.Distance = ref.Distance(observer, acPos).Kilometers()

		state.Highest = updateHighest(state.Highest, aircraft)
		state.Fastest = updateFastest(state.Fastest, aircraft)

		newRarities := NoRarity
		newRarities |= classifyType(&state, &sighting, &aircraft, identity.NewFlight, classifier)
		newRarities |= classifyOperator(&state, &sighting, &aircraft, identity.NewFlight, classifier)
		newRarities |= classifyCountry(&state, &sighting, &aircraft, identity.NewFlight, classifier)
		sighting.Rarities = newRarities

		sighting.Info = sighting.SightingToString()
		currentSightings[idx] = sighting
		state.Sightings[aircraft.Hex] = sighting
	}

	return state, currentSightings
}

func cloneState(state State) State {
	state.Sightings = cloneSightings(state.Sightings)
	state.SeenType = cloneCounts(state.SeenType)
	state.SeenOperator = cloneCounts(state.SeenOperator)
	state.SeenCountry = cloneCounts(state.SeenCountry)
	return state
}

func cloneSightings(src map[string]AircraftSighting) map[string]AircraftSighting {
	out := make(map[string]AircraftSighting, len(src))
	for hex, sighting := range src {
		out[hex] = sighting
	}
	return out
}

func cloneCounts(src map[string]int) map[string]int {
	out := make(map[string]int, len(src))
	for key, count := range src {
		out[key] = count
	}
	return out
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
	thisTypeCountNew := state.SeenType[aType] + 1
	state.SeenType[aType] = thisTypeCountNew
	state.SightedTypes++
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

	flightNo := aircraft.GetFlightNoAsStr()
	if flightNo == "" {
		return NoRarity
	}

	flightCode := aircraft.GetFlightNoAsIcaoCode()
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

	thisOperatorCountNew := state.SeenOperator[sighting.Operator] + 1
	state.SeenOperator[sighting.Operator] = thisOperatorCountNew
	state.SightedOperators++
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

	flightNo := aircraft.GetFlightNoAsStr()
	if flightNo == "" {
		return NoRarity
	}

	flightCode := aircraft.GetFlightNoAsIcaoCode()
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

	thisCountryCountNew := state.SeenCountry[sighting.Country] + 1
	state.SeenCountry[sighting.Country] = thisCountryCountNew
	state.SightedCountries++
	if !isRare(thisCountryCountNew, state.SightedCountries) {
		return NoRarity
	}
	return RareCountry
}

func updateHighest(current, aircraft AircraftRecord) AircraftRecord {
	thisAltitude, thisAltOk := aircraft.AltBaro.(float64)
	if !thisAltOk {
		return current
	}

	highestAltitude, highestAltOk := aircraft.AltBaro.(float64)
	if !highestAltOk || highestAltitude > thisAltitude {
		return current
	}

	return aircraft
}

func updateFastest(current, aircraft AircraftRecord) AircraftRecord {
	if current.GroundSpeed > aircraft.GroundSpeed {
		return current
	}
	return aircraft
}
