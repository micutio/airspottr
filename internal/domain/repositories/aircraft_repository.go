package repositories

import (
	obs "github.com/micutio/airspottr/internal/domain/observation"
)

type AircraftRepository interface {
	RequestAircraft() []obs.AircraftRecord
}
