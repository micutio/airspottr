package repositories

import ref "github.com/micutio/airspottr/internal/domain/reference"

type AircraftTypeRepo interface {
	GetAircraftType(icaoCode string) (ref.IcaoAircraftSpec, bool)
}
