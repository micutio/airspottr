package data

import (
	"errors"
	"fmt"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

var errParseIcaoAircraftMap = errors.New("failed to parse ICAO to aircraft map")

type AircraftSpecificationRepo struct {
	icaoToAircraft map[string]ref.IcaoAircraftSpec
}

func NewAircraftDescriptionRepo() (*AircraftSpecificationRepo, error) {
	const initError = "NewAircraftDescriptionRepo: %w caused by %w"
	icaoToAircraftMap, aircraftErr := ref.GetIcaoToAircraftMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAircraftMap, aircraftErr)
	}

	repo := AircraftSpecificationRepo{
		icaoToAircraft: icaoToAircraftMap,
	}

	return &repo, nil
}

// GetAircraftDescription implements the AicraftDescriptionRepo interface of
// the same name.
func (acr *AircraftSpecificationRepo) GetAircraftDescription(icaoCode string) ref.IcaoAircraftSpec {
	return acr.icaoToAircraft[icaoCode]
}
