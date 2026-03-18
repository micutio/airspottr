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
func (notify *Notify) PrintSummary(dash *Dashboard) {
	notify.Stdout.Println("=== Summary ===")
	notify.listByRarity("aircraft", dash.SeenTypeCount)
	notify.listByRarity("operator", dash.SeenOperatorCount)
	notify.listByRarity("country", dash.SeenCountryCount)
	notify.Stdout.Println("Fastest Aircraft:")
	notify.Stdout.Println(aircraftToString(dash.Fastest))
	notify.Stdout.Println("Highest Aircraft:")
	notify.Stdout.Println(aircraftToString(dash.Highest))
	notify.Stdout.Println("=== End Summary ===")
}

func (notify *Notify) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := GetSortedCountsForProperty(propertyCountMap)

	notify.Stdout.Printf("Rarity from least to most common %s", propertyName)
	for j := range propertyCounts {
		notify.Stdout.Printf("%6d - %s\n", propertyCounts[j].Count, propertyCounts[j].Property)
	}
}

// RarityNotifyToggles selects which rarity dimensions may trigger desktop notifications.
type RarityNotifyToggles struct {
	Type, Operator, Country bool
}

// DefaultRarityNotifyToggles enables notifications for all rarity kinds.
func DefaultRarityNotifyToggles() RarityNotifyToggles {
	return RarityNotifyToggles{Type: true, Operator: true, Country: true}
}

// EmitRarityNotifications sends desktop notifications for sightings, respecting toggles.
// Combined rarities (e.g. type+operator) degrade to the best matching template for the enabled subset.
func (notify *Notify) EmitRarityNotifications(sightings []RareSighting, toggles RarityNotifyToggles) {
	for i := range sightings {
		notify.emitRarityWithToggles(&sightings[i], toggles)
	}
}

func (notify *Notify) emitRarityWithToggles(rs *RareSighting, t RarityNotifyToggles) {
	if rs.Rarities == NoRarity || rs.Sighting == nil {
		return
	}
	f := rs.Rarities
	hasT := f&RareType != 0
	hasO := f&RareOperator != 0
	hasC := f&RareCountry != 0

	wt := t.Type && hasT
	wo := t.Operator && hasO
	wc := t.Country && hasC

	eff := RarityFlag(0)
	if wt {
		eff |= RareType
	}
	if wo {
		eff |= RareOperator
	}
	if wc {
		eff |= RareCountry
	}
	if eff == NoRarity {
		return
	}

	s := rs.Sighting
	switch eff {
	case RareType:
		notify.Stdout.Printf("found rare type %s\n", s.info)
		notifyRareType(s)
	case RareOperator:
		notify.Stdout.Printf("found rare operator: %s\n", s.operator)
		notifyRareOperator(s)
	case RareType | RareOperator:
		notify.Stdout.Printf(
			"found rare type and operator: %s run by %s\n", s.info, s.operator)
		notifyRareTypeAndOperator(s)
	case RareCountry:
		notify.Stdout.Printf("found rare country: %s\n", s.country)
		notifyRareCountry(s)
	case RareType | RareCountry:
		notify.Stdout.Printf("found rare type and country: %s -> %s\n", s.info, s.country)
		notifyRareTypeAndCountry(s)
	case RareOperator | RareCountry:
		notify.Stdout.Printf(
			"found rare operator and country: %s -> %s\n", s.operator, s.country)
		notifyRareOperatorAndCountry(s)
	case RareType | RareOperator | RareCountry:
		notify.Stdout.Printf(
			"found the TRIFECTA: %s -> %s -> %s\n", s.info, s.operator, s.country)
		notifyRareTypeOperatorCountry(s)
	}
}

func notifyRareType(sighting *AircraftSighting) {
	msgTitle := "Rare Aircraft Type Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s)\n%3.0f %s",
		sighting.typeDesc,
		sighting.registration,
		sighting.distance,
		sighting.direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperator(sighting *AircraftSighting) {
	operator := sighting.operator
	msgTitle := "Rare Operator Spotted"
	msgBody := fmt.Sprintf(
		"%s flying %s (%s)\n%3.0f %s",
		operator,
		sighting.typeDesc,
		sighting.registration,
		sighting.distance,
		sighting.direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareCountry(sighting *AircraftSighting) {
	country := sighting.country
	msgTitle := "Rare Aircraft Country Spotted"
	msgBody := fmt.Sprintf(
		"%s-based %s (%s)\n%3.0f %s",
		country,
		sighting.typeDesc,
		sighting.registration,
		sighting.distance,
		sighting.direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeAndOperator(sighting *AircraftSighting) {
	operator := sighting.operator
	msgTitle := "Rare Type & Operator Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s) operated by\n%s\n%3.0f %s",
		sighting.typeDesc,
		sighting.registration,
		operator,
		sighting.distance,
		sighting.direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeAndCountry(sighting *AircraftSighting) {
	country := sighting.country
	msgTitle := "Rare Type & Country Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s) registered in\n%s\n%3.0f %s",
		sighting.typeDesc,
		sighting.registration,
		country,
		sighting.distance,
		sighting.direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperatorAndCountry(sighting *AircraftSighting) {
	operator := sighting.operator
	country := sighting.country
	msgTitle := "Rare Operator & Country Spotted"
	msgBody := fmt.Sprintf(
		"%s\nflying aircraft registered in\n%s\n%3.0f %s",
		operator,
		country,
		sighting.distance,
		sighting.direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeOperatorCountry(sighting *AircraftSighting) {
	var aType string
	if sighting.typeShort != "" {
		aType = sighting.typeShort
	} else {
		aType = sighting.typeDesc
	}

	operator := sighting.operator
	country := sighting.country
	msgTitle := "TRIFECTA Spotted!"
	msgBody := fmt.Sprintf(
		"%s (%s),\nrun by %s,\nregistered in\n%s\n%3.0f %s",
		aType,
		sighting.registration,
		operator,
		country,
		sighting.distance,
		sighting.direction)
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
