package observation

// Classifier looks up reference data used when classifying a sighting.
// Fallback order (ICAO → mil → ownOp, operator country → hex → registration)
// lives in EvaluateBatch, not in the adapter.
type Classifier interface {
	AircraftMake(icaoType string) (makeName string, ok bool)
	OperatorByIcao(code string) (company, country string, ok bool)
	OperatorByMil(code string) (name string, ok bool)
	CountryByHex(hex string) (country string, ok bool)
	CountryByRegistration(reg string) (country string, ok bool)
}
