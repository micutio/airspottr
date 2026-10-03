package notify

import (
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog

	"github.com/gen2brain/beeep"
	obs "github.com/micutio/airspottr/internal/domain/observation"
)

const appIconPath = "./assets/icon.png"

// BeeepNotifier emits desktop notifications for rare sighting events.
type BeeepNotifier struct {
	Stdout log.Logger
}

func NewBeeepNotifier(appName string, consoleOut io.Writer) *BeeepNotifier {
	beeep.AppName = appName //nolint:reassign // This is the only way to set app name in beeep.
	return &BeeepNotifier{
		Stdout: *log.New(consoleOut, "", 0),
	}
}

func DefaultRarityNotifyToggles() obs.RarityNotifyToggles {
	return obs.RarityNotifyToggles{Type: true, Operator: true, Country: true}
}

func (n *BeeepNotifier) EmitRarityNotifications(
	sightings []obs.AircraftSighting,
	toggles obs.RarityNotifyToggles,
) {
	for i := range sightings {
		if sightings[i].Rarities == obs.NoRarity {
			continue
		}
		n.emitRarityWithToggles(sightings[i], toggles)
	}
}

func (n *BeeepNotifier) DefaultRarityNotifyToggles() obs.RarityNotifyToggles {
	return DefaultRarityNotifyToggles()
}

func (n *BeeepNotifier) emitRarityWithToggles(
	sighting obs.AircraftSighting,
	toggles obs.RarityNotifyToggles,
) {
	if sighting.Rarities == obs.NoRarity {
		return
	}
	f := sighting.Rarities
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

	switch rarityFlag { //nolint:exhaustive // By definition noFlag is false when this is called.
	case obs.RareType:
		n.Stdout.Printf("found rare type %sighting\n", sighting.Info)
		n.notifyRareType(sighting)
	case obs.RareOperator:
		n.Stdout.Printf("found rare operator: %sighting\n", sighting.Operator)
		n.notifyRareOperator(sighting)
	case obs.RareType | obs.RareOperator:
		n.Stdout.Printf(
			"found rare type and operator: %sighting run by %sighting\n", sighting.Info, sighting.Operator)
		n.notifyRareTypeAndOperator(sighting)
	case obs.RareCountry:
		n.Stdout.Printf("found rare country: %sighting\n", sighting.Country)
		n.notifyRareCountry(sighting)
	case obs.RareType | obs.RareCountry:
		n.Stdout.Printf("found rare type and country: %sighting -> %sighting\n", sighting.Info, sighting.Country)
		n.notifyRareTypeAndCountry(sighting)
	case obs.RareOperator | obs.RareCountry:
		n.Stdout.Printf(
			"found rare operator and country: %sighting -> %sighting\n", sighting.Operator, sighting.Country)
		n.notifyRareOperatorAndCountry(sighting)
	case obs.RareType | obs.RareOperator | obs.RareCountry:
		n.Stdout.Printf(
			"found the TRIFECTA: %sighting -> %sighting -> %sighting\n",
			sighting.Info,
			sighting.Operator,
			sighting.Country,
		)
		n.notifyRareTypeOperatorCountry(sighting)
	default:
		panic("unknown rare type")
	}
}

func (n *BeeepNotifier) notifyRareType(sighting obs.AircraftSighting) {
	msgTitle := "Rare Aircraft Type Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s)\n%3.0f %s",
		sighting.TypeDesc,
		sighting.Registration,
		sighting.Distance,
		sighting.Direction)
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}

func (n *BeeepNotifier) notifyRareOperator(sighting obs.AircraftSighting) {
	operator := sighting.Operator
	msgTitle := "Rare Operator Spotted"
	msgBody := fmt.Sprintf(
		"%s flying %s (%s)\n%3.0f %s",
		operator,
		sighting.TypeDesc,
		sighting.Registration,
		sighting.Distance,
		sighting.Direction)
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}

func (n *BeeepNotifier) notifyRareCountry(sighting obs.AircraftSighting) {
	country := sighting.Country
	msgTitle := "Rare Aircraft Country Spotted"
	msgBody := fmt.Sprintf(
		"%s-based %s (%s)\n%3.0f %s",
		country,
		sighting.TypeDesc,
		sighting.Registration,
		sighting.Distance,
		sighting.Direction)
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}

func (n *BeeepNotifier) notifyRareTypeAndOperator(sighting obs.AircraftSighting) {
	operator := sighting.Operator
	msgTitle := "Rare Type & Operator Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s) operated by\n%s\n%3.0f %s",
		sighting.TypeDesc,
		sighting.Registration,
		operator,
		sighting.Distance,
		sighting.Direction)
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}

func (n *BeeepNotifier) notifyRareTypeAndCountry(sighting obs.AircraftSighting) {
	country := sighting.Country
	msgTitle := "Rare Type & Country Spotted"
	msgBody := fmt.Sprintf(
		"%s (%s) registered in\n%s\n%3.0f %s",
		sighting.TypeDesc,
		sighting.Registration,
		country,
		sighting.Distance,
		sighting.Direction)
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}

func (n *BeeepNotifier) notifyRareOperatorAndCountry(sighting obs.AircraftSighting) {
	operator := sighting.Operator
	country := sighting.Country
	msgTitle := "Rare Operator & Country Spotted"
	msgBody := fmt.Sprintf(
		"%s\nflying aircraft registered in\n%s\n%3.0f %s",
		operator,
		country,
		sighting.Distance,
		sighting.Direction)
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}

func (n *BeeepNotifier) notifyRareTypeOperatorCountry(sighting obs.AircraftSighting) {
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
	if err := beeep.Notify(msgTitle, msgBody, appIconPath); err != nil {
		panic(err)
	}
}
