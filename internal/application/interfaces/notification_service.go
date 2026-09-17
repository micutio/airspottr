package interfaces

import obs "github.com/micutio/airspottr/internal/domain/observation"

type NotificationService interface {
	EmitRarityNotifications(sightings []obs.AircraftSighting, toggles obs.RarityNotifyToggles)
}
