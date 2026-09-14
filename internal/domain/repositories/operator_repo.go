package repositories

import (
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

type OperatorRepository interface {
	GetOperatorByIcao(icaoCode string) (ref.IcaoOperator, bool)

	// TODO: Make this return ref.IcaoOperator!

	GetOperatorByMilCode(milCode string) (string, bool)
}
