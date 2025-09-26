package internal

import (
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog

	"github.com/gen2brain/beeep"
)

const (
	// appIconPath is the file path to the icon png for this application.
	appIconPath = "./assets/icon.png"
)

type Notify struct {
	Stdout log.Logger
}

func NewNotify(appName string, consoleOut *io.Writer) *Notify {
	beeep.AppName = appName //nolint:reassign // This is the only way to set app name in beeep.
	return &Notify{
		Stdout: *log.New(*consoleOut, "", 0),
	}
}

// PrintSummary prints the highest, fastest and the most and the least common types.
func (notify *Notify) PrintSummary(db *Dashboard) {
	notify.Stdout.Println("=== Summary ===")
	notify.listByRarity("aircraft", db.SeenTypeCount)
	notify.listByRarity("operator", db.SeenOperatorCount)
	notify.listByRarity("country", db.SeenCountryCount)
	notify.Stdout.Println("Fastest Aircraft:")
	notify.Stdout.Println(aircraftToString(db.Fastest))
	notify.Stdout.Println("Highest Aircraft:")
	notify.Stdout.Println(aircraftToString(db.Highest))
	notify.Stdout.Println("=== End Summary ===")
}

func (notify *Notify) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := GetSortedCountsForProperty(propertyCountMap)

	notify.Stdout.Printf("Rarity from least to most common %s", propertyName)
	for j := range propertyCounts {
		notify.Stdout.Printf("%6d - %s\n", propertyCounts[j].Count, propertyCounts[j].Property)
	}
}

func (notify *Notify) EmitRarityNotifications(rareSightings []RareSighting) {
	for _, rareSighting := range rareSightings {
		switch rareSighting.Rarities {
		case NoRarity:
			return
		case RareType:
			notify.Stdout.Printf("found rare type %s\n", aircraftToString(rareSighting.Aircraft))
			notifyRareType(rareSighting.Aircraft, rareSighting.Sighting)
		case RareOperator:
			notify.Stdout.Printf("found rare operator: %s\n", rareSighting.Sighting.Operator)
			notifyRareOperator(rareSighting.Aircraft, rareSighting.Sighting)
		case RareCountry:
			notify.Stdout.Printf("found rare country: %s\n", rareSighting.Sighting.Country)
			notifyRareCountry(rareSighting.Aircraft, rareSighting.Sighting)
		case RareTypeAndOperator:
			notify.Stdout.Printf(
				"found rare type and operator: %s run by %s\n",
				aircraftToString(rareSighting.Aircraft),
				rareSighting.Sighting.Operator)
			notifyRareTypeAndOperator(rareSighting.Aircraft, rareSighting.Sighting)
		case RareTypeAndCountry:
			notify.Stdout.Printf(
				"found rare type and country: %s -> %s\n",
				aircraftToString(rareSighting.Aircraft),
				rareSighting.Sighting.Country)
			notifyRareTypeAndCountry(rareSighting.Aircraft, rareSighting.Sighting)
		case RareOperatorAndCountry:
			notify.Stdout.Printf(
				"found rare operator and country: %s -> %s\n",
				rareSighting.Sighting.Operator,
				rareSighting.Sighting.Country)
			notifyRareOperatorAndCountry(rareSighting.Aircraft, rareSighting.Sighting)
		case RareTypeOperatorCountry:
			notify.Stdout.Printf(
				"found the TRIFECTA: %s -> %s -> %s\n",
				aircraftToString(rareSighting.Aircraft),
				rareSighting.Sighting.Operator,
				rareSighting.Sighting.Country)
			notifyRareCountry(rareSighting.Aircraft, rareSighting.Sighting)
		}
	}
}

func notifyRareType(aircraft *AircraftRecord, sighting *AircraftSighting) {
	var aType string
	// Use the shorter description, if available.
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = sighting.TypeDesc
	}
	msgBody := fmt.Sprintf("%s (%s)", aType, aircraft.Registration)
	err := beeep.Notify("Rare Aircraft Type Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperator(aircraft *AircraftRecord, sighting *AircraftSighting) {
	operator := sighting.Operator

	msgBody := fmt.Sprintf("%s flying %s (%s)", operator, aircraft.Description, aircraft.Registration)
	err := beeep.Notify("Rare Operator Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareCountry(aircraft *AircraftRecord, sighting *AircraftSighting) {
	country := sighting.Country

	msgBody := fmt.Sprintf("%s-based %s (%s)", country, aircraft.Description, aircraft.Registration)
	err := beeep.Notify("Rare Country Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeAndOperator(aircraft *AircraftRecord, sighting *AircraftSighting) {
	var aType string
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = sighting.TypeDesc
	}

	operator := sighting.Operator
	msgTitle := "Rare Type & Operator Spotted"
	msgBody := fmt.Sprintf("%s (%s) operated by\n%s", aType, aircraft.Registration, operator)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeAndCountry(aircraft *AircraftRecord, sighting *AircraftSighting) {
	var aType string
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = sighting.TypeDesc
	}

	country := sighting.Country
	msgTitle := "Rare Type & Country Spotted"
	msgBody := fmt.Sprintf("%s (%s) registered in\n%s", aType, aircraft.Registration, country)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperatorAndCountry(aircraft *AircraftRecord, sighting *AircraftSighting) {
	operator := sighting.Operator
	country := sighting.Country
	msgTitle := "Rare Operator & Country Spotted"
	msgBody := fmt.Sprintf("%s\nflying aircraft registered in\n%s", operator, country)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeOperatorCountry(aircraft *AircraftRecord, sighting *AircraftSighting) {
	var aType string
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = sighting.TypeDesc
	}

	operator := sighting.Operator
	country := sighting.Country
	msgTitle := "TRIFECTA spotted!"
	msgBody := fmt.Sprintf(
		"%s (%s),\nrunby %s,\nregistered in%s",
		aType,
		aircraft.Registration,
		operator,
		country)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

// aircraftToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func aircraftToString(aircraft *AircraftRecord) string {
	flight := aircraft.GetFlightNoAsStr()
	altitude := aircraft.GetAltitudeAsStr()
	var aType string
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = aircraft.CachedType
	}

	return fmt.Sprintf("FNO %s DST %4.0f km ALT %s SPD %3.0f HDG %3.0f TID %s (%s)",
		flight,
		aircraft.CachedDist,
		altitude,
		aircraft.GroundSpeed,
		aircraft.NavHeading,
		aType,
		aircraft.Registration)
}
