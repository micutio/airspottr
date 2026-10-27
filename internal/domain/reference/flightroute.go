package reference

// See https://www.adsbdb.com for Flight route data definition.

const (
	NotAvailable string = "N/A"
)

// FlightrouteRecord is the domain model for route metadata for a flight callsign.
// External API adapters map ADS-B DB payloads into this structure.
type FlightrouteRecord struct {
	Callsign     string
	CallsignIcao string
	CallsignIata string
	Airline      AirlineRecord
	Origin       LocationRecord
	Destination  LocationRecord
}

// AirlineRecord reflects the airline data within FlightrouteRecord.
type AirlineRecord struct {
	Name       string
	Icao       string
	Iata       string
	Country    string
	CountryIso string
	Callsign   string
}

// LocationRecord is used to store data for Flight origin and destination.
type LocationRecord struct {
	CountryIsoName string
	CountryName    string
	Elevation      int
	IataCode       string
	IcaoCode       string
	Latitude       float32
	Longitude      float32
	Municipality   string
	Airport        string
}

func GetDefaultFlightrouteRecord() *FlightrouteRecord {
	return &FlightrouteRecord{
		Callsign:     NotAvailable,
		CallsignIcao: NotAvailable,
		CallsignIata: NotAvailable,
		Airline: AirlineRecord{
			Name:       NotAvailable,
			Icao:       NotAvailable,
			Iata:       NotAvailable,
			Country:    NotAvailable,
			CountryIso: NotAvailable,
			Callsign:   NotAvailable,
		},
		Origin: LocationRecord{
			CountryIsoName: NotAvailable,
			CountryName:    NotAvailable,
			Elevation:      0,
			IataCode:       NotAvailable,
			IcaoCode:       NotAvailable,
			Latitude:       0,
			Longitude:      0,
			Municipality:   NotAvailable,
			Airport:        NotAvailable,
		},
		Destination: LocationRecord{
			CountryIsoName: NotAvailable,
			CountryName:    NotAvailable,
			Elevation:      0,
			IataCode:       NotAvailable,
			IcaoCode:       NotAvailable,
			Latitude:       0,
			Longitude:      0,
			Municipality:   NotAvailable,
			Airport:        NotAvailable,
		},
	}
}
