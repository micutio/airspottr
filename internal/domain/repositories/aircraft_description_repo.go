package repositories

import ref "github.com/micutio/airspottr/internal/domain/reference"

type AircraftDescriptionRepo interface {
	GetAircraftDescription(icaoCode string) ref.IcaoAircraft
}
