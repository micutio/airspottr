package data

import (
	"io"
	"testing"
)

func TestClassifierCountryLookups(t *testing.T) {
	t.Parallel()

	c, err := NewClassifier(io.Discard)
	if err != nil {
		t.Fatalf("NewClassifier() error = %v", err)
	}

	// "VH-FGB" contains "F" and "G" inside the tail number, which previously matched France ("F-") or UK ("G-")
	// when using strings.Contains over a randomized map. It must deterministically match Australia ("VH-").
	country, ok := c.CountryByRegistration("VH-FGB")
	if !ok || country != "Australia" {
		t.Fatalf("CountryByRegistration(VH-FGB) = (%q, %v), want (Australia, true)", country, ok)
	}

	// Exact hex range lower bound (e.g. 004000 is the lower bound for Zimbabwe in ICAOHexRange.csv).
	hexCountry, hexOk := c.CountryByHex("004000")
	if !hexOk || hexCountry != "Zimbabwe" {
		t.Fatalf("CountryByHex(004000) = (%q, %v), want (Zimbabwe, true)", hexCountry, hexOk)
	}
}
