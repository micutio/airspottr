package repositories

type CountryRepository interface {
	GetCountryByHexCode(hexCode string) (string, error)
	GetCountryByRegistration(registration string) (string, bool)
}
