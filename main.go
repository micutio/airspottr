package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	// lat is Latitude of SIN Airport.
	lat float64 = 1.359297
	// lon is Longitude of SIN Airport.
	lon float64 = 103.989348
	// timeFmt is a formatting string for timestamp outputs.
	timeFmt string = "2006-01-02 15:04:05"
	// aircraftUpdateInterval determines the update rate for general aircraft.
	aircraftUpdateInterval = 30 * time.Second
	// milAircraftUpdateInterval determines the update rate for military aircraft.
	milAircraftUpdateInterval = 15 * time.Minute
	// milAircraftUpdateDelay determines interleaving between general and mil aircraft api calls.
	milAircraftUpdateDelay = 15 * time.Second
	// summaryInterval determines how often the summary is show.
	summaryInterval = 1 * time.Hour
)

func main() {
	flightDash := newDashboard()

	// Create a aircraftUpdateTicker that fires every 30 seconds
	aircraftUpdateTicker := time.NewTicker(aircraftUpdateInterval)
	defer aircraftUpdateTicker.Stop()

	milAircraftUpdateTicker := time.NewTicker(milAircraftUpdateInterval)
	defer milAircraftUpdateTicker.Stop()
	// aircraft and military aircraft updates should not coincide to avoid exceeding the api limit.
	// Hence, stagger them by 15 seconds.
	time.AfterFunc(milAircraftUpdateDelay, func() {
		milAircraftUpdateTicker.Reset(milAircraftUpdateInterval)
	})

	raritySummaryTicker := time.NewTicker(summaryInterval)
	defer raritySummaryTicker.Stop()

	// Use a channel to gracefully stop the program if needed.
	// (Though not strictly necessary for an infinite loop)
	done := make(chan bool)

	log.Println("aircraft Tracking dashboard")
	log.Println("Press Ctrl+C to stop the program.")

	// Start a goroutine to perform the requests
	go func() {
		for {
			select {
			case <-aircraftUpdateTicker.C:
				requestAndProcessCivAircraft(&flightDash)
			case <-milAircraftUpdateTicker.C:
				requestAndProcessMilAircraft(&flightDash)
			case <-raritySummaryTicker.C:
				flightDash.listTypesByRarity()
			case <-done:
				// This case allows for graceful shutdown (not used in this example but good practice)
				log.Println("Stopping HTTP GET request routine.")

				return
			}
		}
	}()

	// Just for testing, remove this call later.
	requestAndProcessMilAircraft(&flightDash)

	// Keep the main goroutine alive indefinitely, or until an interrupt signal is received
	// In a real application, you might have other logic here or use a wait group.
	select {} // Block indefinitely
}

func requestAndProcessCivAircraft(dashboard *dashboard) {
	// Define the URL for the HTTP GET request
	targetURL := fmt.Sprintf(
		"https://opendata.adsb.fi/api/v2/lat/%.6f/lon/%.6f/dist/250",
		lat,
		lon,
	)
	// This case is executed every time the ticker "ticks"
	body, requestErr := sendRequest(targetURL)
	if requestErr != nil {
		log.Printf("Error during request: %v", requestErr)

		return
	}

	processingErr := dashboard.processCivAircraftJSON(body)

	if processingErr != nil {
		log.Printf("Error during processing: %v", processingErr)
	}
}

func requestAndProcessMilAircraft(dashboard *dashboard) {
	// Define the URL for the HTTP GET request
	targetURL := "https://opendata.adsb.fi/api/v2/mil"
	// This case is executed every time the ticker "ticks"
	body, requestErr := sendRequest(targetURL)
	if requestErr != nil {
		log.Printf("Error during request: %v", requestErr)

		return
	}

	processingErr := dashboard.processMilAircraftJSON(body)

	if processingErr != nil {
		log.Printf("Error during processing: %v", processingErr)
	}
}

// sendRequest sends an HTTP GET request and returns a valid byte slice of the response body.
func sendRequest(url string) ([]byte, error) {
	resp, respErr := http.Get(url)
	if respErr != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", respErr)
	}
	defer func(bodyReader io.ReadCloser) {
		err := bodyReader.Close()
		if err != nil {
			log.Printf("failed to close body reader: %v", err)
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

	return body, nil
}
