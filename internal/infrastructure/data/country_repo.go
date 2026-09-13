package data

import (
	"errors"
	"fmt"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

var (
	errParseRegToCountryMap      = errors.New("failed to parse reg-prefix to country map")
	errParseHexRangeToCountryMap = errors.New("failed to parse hex-range to country map")
)

type CountryRepo struct {
	hexRangeToCountry  map[ref.HexRange]string
	regPrefixToCountry map[string]string
}

func NewCountryRepo() (*CountryRepo, error) {
	const initError = "NewCountryRepo: %w caused by %w"
	hexRangeToCountryMap, hexRangeErr := ref.GetHexRangeToCountryMap()
	if hexRangeErr != nil {
		return nil, fmt.Errorf(initError, errParseHexRangeToCountryMap, hexRangeErr)
	}

	regPrefixToCountryMap, regPrefixErr := ref.GetRegPrefixMap()
	if regPrefixErr != nil {
		return nil, fmt.Errorf(initError, errParseRegToCountryMap, regPrefixErr)
	}

	repo := CountryRepo{
		hexRangeToCountry:  hexRangeToCountryMap,
		regPrefixToCountry: regPrefixToCountryMap,
	}

	return &repo, nil
}

// GetCountryByHexRange implements the CountryRepo interface of the same name.
func (cr *CountryRepo) GetCountryByHexRange(hexRange ref.HexRange) string {
	return cr.hexRangeToCountry[hexRange]
}

// GetCountryByRegPrefix implements the CountryRepo interface of the same name.
func (cr *CountryRepo) GetCountryByRegPrefix(regPrefix string) string {
	return cr.regPrefixToCountry[regPrefix]
}
