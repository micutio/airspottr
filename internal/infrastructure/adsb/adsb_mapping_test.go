package adsb

import (
	"encoding/json"
	"testing"
)

func TestAircraftJSONMapsToDomainModel(t *testing.T) {
	raw := []byte(`{
		"now": 1,
		"resultCount": 1,
		"ptime": 2,
		"aircraft": [
			{
				"hex": "abc123",
				"Flight": "BAW123",
				"alt_baro": 39000,
				"lat": 51.5,
				"lon": -0.12,
				"r": "G-ABCD",
				"t": "A319",
				"desc": "Airbus A319",
				"ownOp": "British Airways"
			}
		]
	}`)

	var data aircraftQueryResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal aircraft payload: %v", err)
	}
	if len(data.Aircraft) != 1 {
		t.Fatalf("want one aircraft, got %d", len(data.Aircraft))
	}

	aircraft := data.Aircraft[0].toDomain()
	if aircraft.Hex != "abc123" {
		t.Fatalf("hex mismatch: got %q", aircraft.Hex)
	}
	if aircraft.GetFlightNoAsStr() != "BAW123" {
		t.Fatalf("callsign mismatch: got %q", aircraft.GetFlightNoAsStr())
	}
	if aircraft.Description != "Airbus A319" {
		t.Fatalf("description mismatch: got %q", aircraft.Description)
	}
	if aircraft.OwnOp != "British Airways" {
		t.Fatalf("owner mismatch: got %q", aircraft.OwnOp)
	}
}

func TestFlightRouteJSONMapsToDomainModel(t *testing.T) {
	raw := []byte(`{
		"response": {
			"flightroute": {
				"callsign": "BAW123",
				"callsign_icao": "BAW",
				"callsign_iata": "BA123",
				"airline": {
					"name": "British Airways",
					"icao": "BAW",
					"iata": "BA",
					"country": "United Kingdom",
					"country_iso": "GB",
					"callsign": "BAW"
				},
				"origin": {
					"country_iso_name": "GB",
					"country_name": "United Kingdom",
					"elevation": 0,
					"iata_code": "LHR",
					"icao_code": "EGLL",
					"latitude": 51.47,
					"longitude": -0.45,
					"municipality": "London",
					"name": "Heathrow"
				},
				"destination": {
					"country_iso_name": "US",
					"country_name": "United States",
					"elevation": 0,
					"iata_code": "JFK",
					"icao_code": "KJFK",
					"latitude": 40.64,
					"longitude": -73.78,
					"municipality": "New York",
					"name": "John F. Kennedy"
				}
			}
		}
	}`)

	var data flightRouteQueryResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal flightroute payload: %v", err)
	}

	route := data.Response.Flightroute.toDomain()
	if route.Callsign != "BAW123" {
		t.Fatalf("callsign mismatch: got %q", route.Callsign)
	}
	if route.Airline.Country != "United Kingdom" {
		t.Fatalf("airline country mismatch: got %q", route.Airline.Country)
	}
	if route.Origin.Airport != "Heathrow" {
		t.Fatalf("origin airport mismatch: got %q", route.Origin.Airport)
	}
	if route.Destination.Airport != "John F. Kennedy" {
		t.Fatalf("destination airport mismatch: got %q", route.Destination.Airport)
	}
}
