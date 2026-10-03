package observation

import "time"

// RareSightingDetected is a domain event emitted when a newly evaluated sighting
// qualifies as rare under one or more rarity dimensions.
type RareSightingDetected struct {
	Sighting AircraftSighting
	Rarity   RarityFlag
	At       time.Time
}

// EventName identifies the event type for the domain event contract.
func (e RareSightingDetected) EventName() string {
	return "RareSightingDetected"
}

// DetectRareSightings converts current sightings into explicit domain events for
// any rarity that was detected. This keeps rarity classification in the domain
// instead of treating it as an ad-hoc UI polling concern.
func DetectRareSightings(sightings []AircraftSighting) []RareSightingDetected {
	events := make([]RareSightingDetected, 0, len(sightings))
	for i := range sightings {
		sighting := sightings[i]
		if sighting.Rarities == NoRarity {
			continue
		}
		events = append(events, RareSightingDetected{
			Sighting: sighting,
			Rarity:   sighting.Rarities,
			At:       time.Now(),
		})
	}
	return events
}
