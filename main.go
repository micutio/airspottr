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
	targetURL := "https://opendata.adsb.fi/api/v2/lat/1.359297/lon/103.989348/dist/250"
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
				body, requestErr := sendRequest(targetURL)
				if requestErr != nil {
					log.Printf("Error during request: %v", requestErr)
					break
				}
				processingErr := processJsonBody(body, &icaoAircraftMap)
				if processingErr != nil {
					log.Printf("Error during processing: %v", processingErr)
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

// sendRequest sends an HTTP GET request and returns a valid byte slice of the response body
func sendRequest(url string) ([]byte, error) {
	resp, respErr := http.Get(url)
	if respErr != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", respErr)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close body reader: %v", err)
		}
	}(resp.Body) // Ensure the response body is closed

	// Check if the request was successful (status code 200 OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// Read the response body
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", bodyErr)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf("received non-JSON content-type: %s", contentType)
	}

	fmt.Printf(
		"\n[%s] status: %s response body length: %d bytes\n",
		time.Now().Format("2006-01-02 15:04:05"),
		resp.Status,
		len(body),
	)

	return body, nil
}

// processJsonBody processes data contained in the response body
func processJsonBody(body []byte, icaoAircraftTypes *map[string]IcaoAircraft) error {
	// Actual processing takes place here
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
			altBaro := aircraft.AltBaro
			if altBaro == "" {
				altBaro = "unknown"
			}
			aType := (*icaoAircraftTypes)[aircraft.IcaoType].ModelCode
			fmt.Printf("Flight %s on %s at %.0f feet, heading %.2f degrees\n",
				flight,
				aType,
				aircraft.AltBaro,
				aircraft.NavHeading)
		}
	}
	return nil
}
