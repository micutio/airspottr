package repositories

import ref "github.com/micutio/airspottr/internal/domain/reference"

type AircraftSpecificationRepo interface {
	GetAircraftDescription(icaoCode string) ref.IcaoAircraftSpec
}
