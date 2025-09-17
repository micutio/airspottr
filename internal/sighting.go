package internal

import "time"

type AircraftSighting struct {
	lastSeen     time.Time
	lastFlightNo string
	TypeDesc     string // TypeDesc is the full name of the aircraft type
	Operator     string // Operator can be either airline or military organisation
	Country      string // Country of registration
	// TODO: Think about adding a short type description, for when the _desc_ field in Aircraft is set.
}

// TODO: Try to remove *AircraftRecord from this by caching required fields in the sighting instead!
type RareSighting struct {
	Rarities RarityFlag
	Sighting *AircraftSighting
	Aircraft *AircraftRecord
}
