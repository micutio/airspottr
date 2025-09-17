package internal

import "testing"

type testFlight struct {
	flightNo        string
	expectedCode    string
	expectedAirline string
}

func getTestFlights() []testFlight {
	return []testFlight{
		{"SIA106  ", "SIA", "SINGAPORE AIRLINES LIMITED"},
	}
}

func TestFlightToAirlineConversion(t *testing.T) {
	for _, flight := range getTestFlights() {
		aircraft := AircraftRecord{ //nolint:exhaustruct // convenience for testing
			Flight: flight.flightNo,
		}

		icaoCode := aircraft.GetFlightNoAsIcaoCode()

		if icaoCode != flight.expectedCode {
			t.Errorf("fail: want %v, got %v", flight.expectedCode, icaoCode)
		}
	}
}
