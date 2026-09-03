package application

import obs "github.com/micutio/airspottr/internal/domain/observation"

type NotificationService interface {
	EmitRarityNotifications(sightings []obs.RareSighting, toggles obs.RarityNotifyToggles)
}
