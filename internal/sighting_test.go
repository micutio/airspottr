package internal

import (
	"math"
	"testing"
)

// Point represents a geographic location.
type Point struct {
	Lat float64
	Lon float64
}

func TestBearing(t *testing.T) {
	tests := []struct {
		name     string
		p1       Point
		p2       Point
		expected float64
	}{
		{
			name:     "Due North",
			p1:       Point{Lat: 0, Lon: 0},
			p2:       Point{Lat: 10, Lon: 0},
			expected: 0.0,
		},
		{
			name:     "Due East",
			p1:       Point{Lat: 0, Lon: 0},
			p2:       Point{Lat: 0, Lon: 10},
			expected: 90.0,
		},
		{
			name:     "Due South",
			p1:       Point{Lat: 10, Lon: 0},
			p2:       Point{Lat: 0, Lon: 0},
			expected: 180.0,
		},
		{
			name:     "Due West",
			p1:       Point{Lat: 0, Lon: 10},
			p2:       Point{Lat: 0, Lon: 0},
			expected: 270.0,
		},
		{
			name:     "New York to London", // Long distance calculation
			p1:       Point{Lat: 40.7128, Lon: -74.0060},
			p2:       Point{Lat: 51.5074, Lon: -0.1278},
			expected: 51.21,
		},
		{
			name:     "London to New York", // Reciprocal of previous example
			p1:       Point{Lat: 51.5074, Lon: -0.1278},
			p2:       Point{Lat: 40.7128, Lon: -74.0060},
			expected: 288.33,
		},
		{
			name:     "Sydney to Tokyo", // Southern to Northern hemisphere
			p1:       Point{Lat: -33.8688, Lon: 151.2093},
			p2:       Point{Lat: 35.6895, Lon: 139.6917},
			expected: 350.09,
		},
		{
			name:     "Auckland to Honolulu", // Crossing International Date Line
			p1:       Point{Lat: -36.8485, Lon: 174.7633},
			p2:       Point{Lat: 21.3069, Lon: -157.8583},
			expected: 28.57,
		},
		{
			name:     "Quito to Nairobi", // Equator crossing (Eastward)
			p1:       Point{Lat: -0.1807, Lon: -78.4678},
			p2:       Point{Lat: -1.2921, Lon: 36.8219},
			expected: 91.51,
		},
		{
			name:     "South Pole to North Pole", // Maximum Latitudinal span
			p1:       Point{Lat: -90.0, Lon: 0.0},
			p2:       Point{Lat: 90.0, Lon: 0.0},
			expected: 0.0,
		},
	}

	// Precision threshold for floating point comparison
	const epsilon = 0.01

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBearing(tt.p1.Lat, tt.p1.Lon, tt.p2.Lat, tt.p2.Lon)
			if math.Abs(got-tt.expected) > epsilon {
				t.Errorf("Bearing() = %v, want %v", got, tt.expected)
			}
		})
	}
}
