package repositories

import (
	ref "github.com/micutio/airspottr/internal/domain/reference"
)

type OperatorProvider interface {
	GetOperatorByIcao(icaoCode string) ref.IcaoOperator

	// TODO: Make this return ref.IcaoOperator!
	GetOperatorByMilCode(milCode string) string
}
