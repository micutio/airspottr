package reference

const (

	// CountryUnknown is what we use for aircraft with a type that's either empty or can't be found.
	CountryUnknown = "unknown"
)

type IcaoAircraftSpec struct {
	Class  string
	Engine string
	Make   string
}

type IcaoOperator struct {
	Company string
	Country string
}

type HexRange struct {
	LowerBound int64
	UpperBound int64
}
