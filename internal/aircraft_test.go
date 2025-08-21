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
	for _, ft := range getTestFlights() {
		aircraft := aircraftRecord{ //nolint:exhauststruct // convenience for testing
			Flight: ft.flightNo,
		}

		icaoCode := aircraft.GetFlightNoAsIcaoCode()
		if icaoCode != ft.expectedCode {
			t.Errorf("fail: want %v, got %v", ft.expectedCode, icaoCode)
		}
	}
}
