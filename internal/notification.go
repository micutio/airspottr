package internal

import (
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog

	"github.com/gen2brain/beeep"
	obs "github.com/micutio/airspottr/internal/domain/observation"
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
func (notify *Notify) EmitRarityNotifications(sightings []obs.RareSighting, toggles RarityNotifyToggles) {
	for i := range sightings {
		notify.emitRarityWithToggles(&sightings[i], toggles)
	}
}

func (notify *Notify) emitRarityWithToggles(rareSighting *obs.RareSighting, toggles RarityNotifyToggles) {
	if rareSighting.Rarities == obs.NoRarity || rareSighting.Sighting == nil {
		return
	}
	f := rareSighting.Rarities
	hasT := f&obs.RareType != 0
	hasO := f&obs.RareOperator != 0
	hasC := f&obs.RareCountry != 0

	toggleType := toggles.Type && hasT
	toggleOperator := toggles.Operator && hasO
	toggleCountry := toggles.Country && hasC

	rarityFlag := obs.RarityFlag(0)
	if toggleType {
		rarityFlag |= obs.RareType
	}
	if toggleOperator {
		rarityFlag |= obs.RareOperator
	}
	if toggleCountry {
		rarityFlag |= obs.RareCountry
	}
	if rarityFlag == obs.NoRarity {
		return
	}

	sighting := rareSighting.Sighting
	switch rarityFlag { //nolint:exhaustive // By definition noFlag is false when this is called.
	case obs.RareType:
		notify.Stdout.Printf("found rare type %sighting\n", sighting.Info)
		notifyRareType(sighting)
	case obs.RareOperator:
		notify.Stdout.Printf("found rare operator: %sighting\n", sighting.Operator)
		notifyRareOperator(sighting)
	case obs.RareType | obs.RareOperator:
		notify.Stdout.Printf(
			"found rare type and operator: %sighting run by %sighting\n", sighting.Info, sighting.Operator)
		notifyRareTypeAndOperator(sighting)
	case obs.RareCountry:
		notify.Stdout.Printf("found rare country: %sighting\n", sighting.Country)
		notifyRareCountry(sighting)
	case obs.RareType | obs.RareCountry:
		notify.Stdout.Printf("found rare type and country: %sighting -> %sighting\n", sighting.Info, sighting.Country)
		notifyRareTypeAndCountry(sighting)
	case obs.RareOperator | obs.RareCountry:
		notify.Stdout.Printf(
			"found rare operator and country: %sighting -> %sighting\n", sighting.Operator, sighting.Country)
		notifyRareOperatorAndCountry(sighting)
	case obs.RareType | obs.RareOperator | obs.RareCountry:
		notify.Stdout.Printf(
			"found the TRIFECTA: %sighting -> %sighting -> %sighting\n",
			sighting.Info,
			sighting.Operator,
			sighting.Country,
		)
		notifyRareTypeOperatorCountry(sighting)
	default:
		panic("unknown rare type")
	}
}

func notifyRareType(sighting *obs.AircraftSighting) {
	msgTitle := "Rare Aircraft Type Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s)\n%3.0f %s",
		sighting.TypeDesc,
		sighting.Registration,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperator(sighting *obs.AircraftSighting) {
	operator := sighting.Operator
	msgTitle := "Rare Operator Spotted"
	msgBody := fmt.Sprintf(
		"%s flying %s (%s)\n%3.0f %s",
		operator,
		sighting.TypeDesc,
		sighting.Registration,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareCountry(sighting *obs.AircraftSighting) {
	country := sighting.Country
	msgTitle := "Rare Aircraft Country Spotted"
	msgBody := fmt.Sprintf(
		"%s-based %s (%s)\n%3.0f %s",
		country,
		sighting.TypeDesc,
		sighting.Registration,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeAndOperator(sighting *obs.AircraftSighting) {
	operator := sighting.Operator
	msgTitle := "Rare Type & Operator Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s) operated by\n%s\n%3.0f %s",
		sighting.TypeDesc,
		sighting.Registration,
		operator,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeAndCountry(sighting *obs.AircraftSighting) {
	country := sighting.Country
	msgTitle := "Rare Type & Country Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s) registered in\n%s\n%3.0f %s",
		sighting.TypeDesc,
		sighting.Registration,
		country,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperatorAndCountry(sighting *obs.AircraftSighting) {
	operator := sighting.Operator
	country := sighting.Country
	msgTitle := "Rare Operator & Country Spotted"
	msgBody := fmt.Sprintf(
		"%s\nflying aircraft registered in\n%s\n%3.0f %s",
		operator,
		country,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareTypeOperatorCountry(sighting *obs.AircraftSighting) {
	var aType string
	if sighting.TypeShort != "" {
		aType = sighting.TypeShort
	} else {
		aType = sighting.TypeDesc
	}

	operator := sighting.Operator
	country := sighting.Country
	msgTitle := "TRIFECTA Spotted!"
	msgBody := fmt.Sprintf(
		"%s (%s),\nrun by %s,\nregistered in\n%s\n%3.0f %s",
		aType,
		sighting.Registration,
		operator,
		country,
		sighting.Distance,
		sighting.Direction)
	err := beeep.Notify(msgTitle, msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

// aircraftToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func aircraftToString(aircraft *obs.AircraftRecord) string {
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
