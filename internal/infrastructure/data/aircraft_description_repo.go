package data

import (
	"errors"
	"fmt"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

var errParseIcaoAircraftMap = errors.New("failed to parse ICAO to aircraft map")

type AircraftDescriptionRepo struct {
	icaoToAircraft map[string]ref.IcaoAircraft
}

func NewAircraftDescriptionRepo() (*AircraftDescriptionRepo, error) {
	const initError = "NewAircraftDescriptionRepo: %w caused by %w"
	icaoToAircraftMap, aircraftErr := ref.GetIcaoToAircraftMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAircraftMap, aircraftErr)
	}

	repo := AircraftDescriptionRepo{
		icaoToAircraft: icaoToAircraftMap,
	}

	return &repo, nil
}

// GetAircraftDescription implements the AicraftDescriptionRepo interface of
// the same name.
func (acr *AircraftDescriptionRepo) GetAircraftDescription(icaoCode string) ref.IcaoAircraft {
	return acr.icaoToAircraft[icaoCode]
}
