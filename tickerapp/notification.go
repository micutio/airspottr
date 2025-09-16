package tickerapp

import (
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"sort"

	"github.com/gen2brain/beeep"
	"github.com/micutio/airspottr/internal/dash"
)

type rarityFlag int

const (
	noRarity                rarityFlag = 0b000
	rareType                rarityFlag = 0b001
	rareOperator            rarityFlag = 0b010
	rareCountry             rarityFlag = 0b100
	rareTypeAndOperator     rarityFlag = 0b011
	rareTypeAndCountry      rarityFlag = 0b101
	rareOperatorAndCountry  rarityFlag = 0b110
	rareTypeOperatorCountry rarityFlag = 0b111
)

type Notify struct {
	Stdout log.Logger
}

func NewNotify(consoleOut *io.Writer) *Notify {
	return &Notify{
		Stdout: *log.New(*consoleOut, "", 0),
	}
}

// PrintSummary prints the highest, fastest and the most and the least common types.
func (notify *Notify) PrintSummary(db *dash.Dashboard) {
	notify.Stdout.Println("=== Summary ===")
	notify.listByRarity("aircraft", db.seenTypeCount)
	notify.listByRarity("operator", db.seenOperatorCount)
	notify.listByRarity("country", db.seenCountryCount)
	notify.Stdout.Println("Fastest Aircraft:")
	notify.Stdout.Println(aircraftToString(db.Fastest))
	notify.Stdout.Println("Highest Aircraft:")
	notify.Stdout.Println(aircraftToString(db.Highest))
	notify.Stdout.Println("=== End Summary ===")
}

type PropertyCountTuple struct {
	Property string
	Count    int
}

type ByCount []PropertyCountTuple

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Less(i, j int) bool { return a[i].Count < a[j].Count }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func getSortedCountsForProperty(propertyCountMap map[string]int) []PropertyCountTuple {
	propertyCounts := make([]PropertyCountTuple, len(propertyCountMap))
	i := 0
	for key, value := range propertyCountMap {
		propertyCounts[i] = PropertyCountTuple{Property: key, Count: value}
		i++
	}

	sort.Sort(ByCount(propertyCounts))
	return propertyCounts
}

// GetTypeRarities contains all seen types ever, sorted by their counts in increasing order.
func GetTypeRarities() []PropertyCountTuple {
	return getSortedCountsForProperty(db.seenTypeCount)
}

// GetOperatorRarities contains all seen operators ever, sorted by their counts in increasing order.
func GetOperatorRarities() []PropertyCountTuple {
	return getSortedCountsForProperty(db.seenOperatorCount)
}

// GetCountryRarities contains all seen countries ever, sorted by their counts in increasing order.
func GetCountryRarities() []PropertyCountTuple {
	return getSortedCountsForProperty(db.seenCountryCount)
}
func (notify *Notify) listByRarity(propertyName string, propertyCountMap map[string]int) {
	propertyCounts := getSortedCountsForProperty(propertyCountMap)

	notify.Stdout.Printf("Rarity from least to most common %s", propertyName)
	for j := range propertyCounts {
		notify.Stdout.Printf("%6d - %s\n", propertyCounts[j].Count, propertyCounts[j].Property)
	}
}

func emitRarityNotifications(
	newRarities rarityFlag,
	aircraft *aircraftRecord,
	sighting *aircraftSighting,
	stdout *log.Logger,
	isWarmup bool,
) {
	switch newRarities {
	case noRarity:
		return
	case rareType:
		stdout.Printf("found rare type %s\n", aircraftToString(aircraft))
		if isWarmup {
			notifyRareType(aircraft, sighting)
		}
	case rareOperator:
		stdout.Printf("found rare operator: %s\n", sighting.operator)
		if !isWarmup {
			notifyRareOperator(aircraft, sighting)
		}
	case rareCountry:
		stdout.Printf("found rare country: %s\n", sighting.country)
		if !isWarmup {
			notifyRareCountry(aircraft, sighting)
		}
	case rareTypeAndOperator:
	case rareTypeAndCountry:
	case rareOperatorAndCountry:
	case rareTypeOperatorCountry:
	}
}

func notifyRareType(aircraft *aircraftRecord, sighting *aircraftSighting) {
	var aType string
	// Use the shorter description, if available.
	if aircraft.Description != "" {
		aType = aircraft.Description
	} else {
		aType = sighting.typeDesc
	}
	msgBody := fmt.Sprintf("%s (%s)", aType, aircraft.Registration)
	err := beeep.Notify("Rare Aircraft Type Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareOperator(aircraft *aircraftRecord, sighting *aircraftSighting) {
	operator := sighting.operator

	msgBody := fmt.Sprintf("%s flying %s (%s)", operator, aircraft.Description, aircraft.Registration)
	err := beeep.Notify("Rare Operator Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

func notifyRareCountry(aircraft *aircraftRecord, sighting *aircraftSighting) {
	country := sighting.country

	msgBody := fmt.Sprintf("%s-based %s (%s)", country, aircraft.Description, aircraft.Registration)
	err := beeep.Notify("Rare Country Spotted", msgBody, appIconPath)
	if err != nil {
		panic(err)
	}
}

// aircraftToString generates a one-liner consisting of the most relevant information about the
// given aircraft.
func aircraftToString(aircraft *aircraftRecord) string {
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
