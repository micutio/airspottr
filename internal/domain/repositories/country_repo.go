package repositories

import ref "github.com/micutio/airspottr/internal/domain/reference"

type CountryProvider interface {
	GetCountryByHexRange(ref.HexRange) string
	GetCountryByRegistration(string) string
}
