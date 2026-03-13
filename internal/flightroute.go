package internal

// See https://www.adsbdb.com for Flight route data definition.

const (
	NotAvailable string = "N/A"
)

// FlightrouteResponse reflects the JSON received from calls to adsbdb.
type FlightrouteResponse struct {
	Response FlightrouteResult `json:"response"`
}

// FlightrouteResult reflects the result contained withing FlightrouteResponse.
type FlightrouteResult struct {
	Flightroute FlightRouteRecord `json:"flightroute"`
}

// FlightRouteRecord reflects the actual flightroute data contained in the result.
type FlightRouteRecord struct {
	Callsign     string         `json:"callsign"`
	CallsignIcao string         `json:"callsign_icao"`
	CallsignIata string         `json:"callsign_iata"`
	Airline      AirlineRecord  `json:"airline"`
	Origin       LocationRecord `json:"origin"`
	Destination  LocationRecord `json:"destination"`
}

// AirlineRecord reflects the airline data within FlightRouteRecord.
type AirlineRecord struct {
	Name       string `json:"name"`
	Icao       string `json:"icao"`
	Iata       string `json:"iata"`
	Country    string `json:"country"`
	CountryIso string `json:"country_iso"`
	Callsign   string `json:"callsign"`
}

// LocationRecord is used to store data for Flight origin and destination.
type LocationRecord struct {
	CountryIsoName string  `json:"country_iso_name"`
	CountryName    string  `json:"country_name"`
	Elevation      int     `json:"elevation"`
	IataCode       string  `json:"iata_code"`
	IcaoCode       string  `json:"icao_code"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	Municipality   string  `json:"municipality"`
	Airport        string  `json:"name"`
}

func GetDefaultFlightrouteRecord() *FlightRouteRecord {
	return &FlightRouteRecord{
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
