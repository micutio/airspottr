package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

func main() {
	// Setup data maps
	icaoAircraftMap := GetIcaoAircraftMap()

	// Define the URL for the HTTP GET request
	targetURL := "https://opendata.adsb.fi/api/v2/lat/1.359297/lon/103.989348/dist/25"
	//           "https://api.adsb.lol/v2/lat/1.359297/lon/103.989348/dist/25"

	// Create a ticker that fires every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Use a channel to gracefully stop the program if needed (though not strictly necessary for an infinite loop)
	done := make(chan bool)

	fmt.Printf("Starting HTTP GET request to %s every 30 seconds...\n", targetURL)
	fmt.Println("Press Ctrl+C to stop the program.")

	// Start a goroutine to perform the requests
	go func() {
		for {
			select {
			case <-ticker.C:
				// This case is executed every time the ticker "ticks"
				err := sendAndProcessRequest(targetURL, &icaoAircraftMap)
				if err != nil {
					log.Printf("Error during request: %v", err)
				}
			case <-done:
				// This case allows for graceful shutdown (not used in this example but good practice)
				fmt.Println("Stopping HTTP GET request routine.")
				return
			}
		}
	}()

	// Keep the main goroutine alive indefinitely, or until an interrupt signal is received
	// In a real application, you might have other logic here or use a wait group.
	select {} // Block indefinitely
}

// sendAndProcessRequest sends an HTTP GET request and processes the response
func sendAndProcessRequest(url string, icaoAircraftTypes *map[string]IcaoAircraft) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send GET request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close body reader: %v", err)
		}
	}(resp.Body) // Ensure the response body is closed

	// Check if the request was successful (status code 200 OK)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// — Start of Processing Logic —
	fmt.Printf("\n[%s] status: %s response body length: %d bytes\n",
		time.Now().Format("2006-01-02 15:04:05"),
		resp.Status,
		len(body),
	)

	// Example processing: print a snippet of the response body
	if len(body) == 0 {
		fmt.Println("Response body is empty.")
		// This is considered valid, no need to chuck an error.
		return nil
	}

	// Example processing: you could parse JSON/XML here, check for specific content, etc.
	if contentType := resp.Header.Get("Content-Type"); strings.Contains(contentType, "application/json") {
		var data AircraftRecord
		if err := json.Unmarshal(body, &data); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		foundAircraftCount := len(data.Aircraft)
		if foundAircraftCount == 0 {
			fmt.Println("No aircraft found.")
		} else {
			sort.Sort(ByFlight(data.Aircraft))
			for i := range foundAircraftCount {
				aircraft := data.Aircraft[i]
				flight := aircraft.Flight
				if len(flight) == 0 {
					flight = "unknown " // add space for consistent formatting with ICAO codes
				}
				aType := (*icaoAircraftTypes)[aircraft.IcaoType].ModelCode
				fmt.Printf("Flight %s on %s at %.0f feet, heading %.2f degrees\n",
					flight,
					aType,
					aircraft.AltBaro,
					aircraft.NavHeading)
			}
		}
	}
	// --- End of Processing Logic ---

	return nil
}
