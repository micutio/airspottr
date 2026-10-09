package main

import (
	"testing"

	"github.com/micutio/airspottr/internal/infrastructure/adsb"
)

func TestBuildAppsWithDifferentLocations(t *testing.T) {
	t.Chdir("../..")

	locations := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{name: "hamburg", lat: 53.5511, lon: 9.9937},
		{name: "singapore", lat: 1.3521, lon: 103.8198},
		{name: "new-york", lat: 40.7128, lon: -74.0060},
	}

	for _, loc := range locations {
		t.Run(loc.name+"_ticker", func(t *testing.T) {
			options := adsb.RequestOptions{
				Lat: loc.lat,
				Lon: loc.lon,
			}
			deps, err := buildTickerApp("testapp", options)
			if err != nil {
				t.Fatalf("failed to build ticker app for %s: %v", loc.name, err)
			}
			if deps.Dashboard.Lat != loc.lat || deps.Dashboard.Lon != loc.lon {
				t.Fatalf("expected dashboard coords (%f,%f), got (%f,%f)",
					loc.lat, loc.lon, deps.Dashboard.Lat, deps.Dashboard.Lon)
			}
		})

		t.Run(loc.name+"_tui", func(t *testing.T) {
			options := adsb.RequestOptions{
				Lat: loc.lat,
				Lon: loc.lon,
			}
			deps, err := buildTUIApp("testapp", options)
			if err != nil {
				t.Fatalf("failed to build TUI app for %s: %v", loc.name, err)
			}
			if closer, ok := deps.ErrorLog.(interface{ Close() error }); ok {
				_ = closer.Close()
			}
			if deps.Dashboard.Lat != loc.lat || deps.Dashboard.Lon != loc.lon {
				t.Fatalf("expected dashboard coords (%f,%f), got (%f,%f)",
					loc.lat, loc.lon, deps.Dashboard.Lat, deps.Dashboard.Lon)
			}
		})
	}
}
