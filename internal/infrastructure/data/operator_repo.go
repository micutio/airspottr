package data

import (
	"errors"
	"fmt"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

var (
	errParseIcaoAirlineMap = errors.New("failed to parse ICAO to airline map")
	errParseMilCodeMap     = errors.New("failed to parse mil code to operator map")
)

type OperatorRepo struct {
	icaoToOperator    map[string]ref.IcaoOperator
	milCodeToOperator map[string]string
}

func NewOperatorRepo() (*OperatorRepo, error) {
	const initError = "NewOperatorRepo: %w caused by %w"
	icaoToOperatorMap, aircraftErr := ref.GetIcaoToAirlineMap()
	if aircraftErr != nil {
		return nil, fmt.Errorf(initError, errParseIcaoAirlineMap, aircraftErr)
	}

	milCodeToOperatorMap, milCodeErr := ref.GetMilCodeToOperatorMap()
	if milCodeErr != nil {
		return nil, fmt.Errorf(initError, errParseMilCodeMap, milCodeErr)
	}

	repo := OperatorRepo{
		icaoToOperator:    icaoToOperatorMap,
		milCodeToOperator: milCodeToOperatorMap,
	}

	return &repo, nil
}

// GetOperatorByIcao implements the OperatorRepo interface of the same name.
func (acr *OperatorRepo) GetOperatorByIcao(icaoCode string) ref.IcaoOperator {
	return acr.icaoToOperator[icaoCode]
}

// GetOperatorByMilCode implements the OperatorRepo interface of the same name.
func (acr *OperatorRepo) GetOperatorByMilCode(milCode string) ref.IcaoOperator {
	return acr.icaoToOperator[milCode]
}
