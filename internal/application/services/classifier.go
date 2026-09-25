package services

import (
	"log" //nolint:depguard // Don't feel like using slog

	obs "github.com/micutio/airspottr/internal/domain/observation"
	rep "github.com/micutio/airspottr/internal/domain/repositories"
)

// referenceClassifier adapts CSV-backed repositories to obs.Classifier.
type referenceClassifier struct {
	aircraftTypeRepo rep.AircraftTypeRepo
	operatorRepo     rep.OperatorRepository
	countryRepo      rep.CountryRepository
	errOut           *log.Logger
}

func newReferenceClassifier(
	aircraftTypeRepo rep.AircraftTypeRepo,
	operatorRepo rep.OperatorRepository,
	countryRepo rep.CountryRepository,
	errOut *log.Logger,
) *referenceClassifier {
	return &referenceClassifier{
		aircraftTypeRepo: aircraftTypeRepo,
		operatorRepo:     operatorRepo,
		countryRepo:      countryRepo,
		errOut:           errOut,
	}
}

func (c *referenceClassifier) AircraftMake(icaoType string) (string, bool) {
	spec, exists := c.aircraftTypeRepo.GetAircraftType(icaoType)
	if !exists {
		return "", false
	}
	return spec.Make, true
}

func (c *referenceClassifier) OperatorByIcao(code string) (string, string, bool) {
	operator, exists := c.operatorRepo.GetOperatorByIcao(code)
	if !exists {
		return "", "", false
	}
	return operator.Company, operator.Country, true
}

func (c *referenceClassifier) OperatorByMil(code string) (string, bool) {
	return c.operatorRepo.GetOperatorByMilCode(code)
}

func (c *referenceClassifier) CountryByHex(hex string) (string, bool) {
	country, countryErr := c.countryRepo.GetCountryByHexCode(hex)
	if countryErr != nil {
		if c.errOut != nil {
			c.errOut.Printf("warning: invalid hex code: %v", countryErr)
		}
		return "", false
	}
	return country, true
}

func (c *referenceClassifier) CountryByRegistration(reg string) (string, bool) {
	return c.countryRepo.GetCountryByRegistration(reg)
}

var _ obs.Classifier = (*referenceClassifier)(nil)
