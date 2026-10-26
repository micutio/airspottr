package notify

import (
	"fmt"
	"io"

	"github.com/gen2brain/beeep"
	obs "github.com/micutio/airspottr/internal/domain/observation"
)

const appIconPath = "./assets/icon.png"

// BeeepNotifier emits desktop notifications for rare sighting events.
type BeeepNotifier struct {
	out io.Writer
}

func NewBeeepNotifier(appName string, consoleOut io.Writer) *BeeepNotifier {
	beeep.AppName = appName //nolint:reassign // This is the only way to set app name in beeep.
	if consoleOut == nil {
		consoleOut = io.Discard
	}
	return &BeeepNotifier{
		out: consoleOut,
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
	rarityFlag := enabledRarityFlags(sighting.Rarities, toggles)
	if rarityFlag == obs.NoRarity {
		return
	}

	title, body, logLine := formatRareSighting(sighting, rarityFlag)
	_, _ = fmt.Fprintln(n.out, logLine)
	if err := beeep.Notify(title, body, appIconPath); err != nil {
		_, _ = fmt.Fprintf(n.out, "warning: desktop notification failed: %v\n", err)
	}
}

func enabledRarityFlags(rarities obs.RarityFlag, toggles obs.RarityNotifyToggles) obs.RarityFlag {
	var flag obs.RarityFlag
	if toggles.Type && rarities&obs.RareType != 0 {
		flag |= obs.RareType
	}
	if toggles.Operator && rarities&obs.RareOperator != 0 {
		flag |= obs.RareOperator
	}
	if toggles.Country && rarities&obs.RareCountry != 0 {
		flag |= obs.RareCountry
	}
	return flag
}

func formatRareSighting(
	sighting obs.AircraftSighting,
	rarityFlag obs.RarityFlag,
) (string, string, string) {
	posLine := fmt.Sprintf("%3.0f %s", sighting.Distance, sighting.Direction)
	typeWithReg := fmt.Sprintf("%s (%s)", sighting.TypeDesc, sighting.Registration)

	switch rarityFlag { //nolint:exhaustive // NoRarity is filtered before calling formatRareSighting.
	case obs.RareType:
		return "Rare Aircraft Type Spotted",
			fmt.Sprintf("%s\n%s", typeWithReg, posLine),
			fmt.Sprintf("found rare type %s", sighting.Info)
	case obs.RareOperator:
		return "Rare Operator Spotted",
			fmt.Sprintf("%s flying %s\n%s", sighting.Operator, typeWithReg, posLine),
			fmt.Sprintf("found rare operator: %s", sighting.Operator)
	case obs.RareType | obs.RareOperator:
		return "Rare Type & Operator Spotted",
			fmt.Sprintf("%s operated by\n%s\n%s", typeWithReg, sighting.Operator, posLine),
			fmt.Sprintf("found rare type and operator: %s run by %s", sighting.Info, sighting.Operator)
	case obs.RareCountry:
		return "Rare Aircraft Country Spotted",
			fmt.Sprintf("%s-based %s\n%s", sighting.Country, typeWithReg, posLine),
			fmt.Sprintf("found rare country: %s", sighting.Country)
	case obs.RareType | obs.RareCountry:
		return "Rare Type & Country Spotted",
			fmt.Sprintf("%s registered in\n%s\n%s", typeWithReg, sighting.Country, posLine),
			fmt.Sprintf("found rare type and country: %s -> %s", sighting.Info, sighting.Country)
	case obs.RareOperator | obs.RareCountry:
		return "Rare Operator & Country Spotted",
			fmt.Sprintf("%s\nflying aircraft registered in\n%s\n%s", sighting.Operator, sighting.Country, posLine),
			fmt.Sprintf("found rare operator and country: %s -> %s", sighting.Operator, sighting.Country)
	case obs.RareType | obs.RareOperator | obs.RareCountry:
		aType := sighting.TypeDesc
		if sighting.TypeShort != "" {
			aType = sighting.TypeShort
		}
		return "TRIFECTA Spotted!",
			fmt.Sprintf("%s (%s),\nrun by %s,\nregistered in\n%s\n%s",
				aType, sighting.Registration, sighting.Operator, sighting.Country, posLine),
			fmt.Sprintf("found the TRIFECTA: %s -> %s -> %s",
				sighting.Info, sighting.Operator, sighting.Country)
	default:
		panic("unknown rare type")
	}
}
