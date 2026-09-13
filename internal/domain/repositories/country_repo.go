package repositories

type CountryRepository interface {
	GetCountryByHexCode(hexCode string) string
	GetCountryByRegistration(registration string) (string, bool)
}
