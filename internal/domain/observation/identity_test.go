package observation

import "testing"

func TestNewAircraftHex(t *testing.T) {
	t.Parallel()

	hex, err := NewAircraftHex("abc123")
	if err != nil {
		t.Fatalf("NewAircraftHex() unexpected error: %v", err)
	}
	if hex.String() != "ABC123" {
		t.Fatalf("hex.String() = %q, want %q", hex.String(), "ABC123")
	}
	if _, err = NewAircraftHex("abc"); err == nil {
		t.Fatal("NewAircraftHex() accepted invalid length")
	}
}

func TestNewFlightCallsign(t *testing.T) {
	t.Parallel()

	callsign, err := NewFlightCallsign(" baw123 ")
	if err != nil {
		t.Fatalf("NewFlightCallsign() unexpected error: %v", err)
	}
	if callsign.String() != "BAW123" {
		t.Fatalf("callsign.String() = %q, want %q", callsign.String(), "BAW123")
	}
	if _, err = NewFlightCallsign(""); err == nil {
		t.Fatal("NewFlightCallsign() accepted empty value")
	}
}
