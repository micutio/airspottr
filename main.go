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
	Lat     float64 = 1.359297
	Lon     float64 = 103.989348
	TimeFmt string  = "2006-01-02 15:04:05"
)

func main() {
	// Setup dashboard
	dashboard := NewDashboard()

	// Create a aircraftUpdateTicker that fires every 30 seconds
	aircraftUpdateTicker := time.NewTicker(30 * time.Second)
	defer aircraftUpdateTicker.Stop()

	milAircraftUpdateTicker := time.NewTicker(1 * time.Hour)
	defer milAircraftUpdateTicker.Stop()
	time.AfterFunc(15*time.Second, func() {
		milAircraftUpdateTicker.Reset(15 * time.Minute)
	})

	raritySummaryTicker := time.NewTicker(1 * time.Hour)
	defer raritySummaryTicker.Stop()
	time.AfterFunc(15*time.Minute, func() {
		raritySummaryTicker.Reset(15 * time.Minute)
	})

	// Use a channel to gracefully stop the program if needed (though not strictly necessary for an infinite loop)
	done := make(chan bool)

	fmt.Println("Aircraft Tracking Dashboard")
	fmt.Println("Press Ctrl+C to stop the program.")

	// Start a goroutine to perform the requests
	go func() {
		for {
			select {
			case <-aircraftUpdateTicker.C:
				requestAndProcessCivAircraft(&dashboard)
				dashboard.ListTypesByRarity()
			case <-milAircraftUpdateTicker.C:
				requestAndProcessMilAircraft(&dashboard)
			case <-raritySummaryTicker.C:
				dashboard.ListTypesByRarity()
			case <-done:
				// This case allows for graceful shutdown (not used in this example but good practice)
				fmt.Println("Stopping HTTP GET request routine.")
				return
			}
		}
	}()

	// Just for testing, remove this call later.
	requestAndProcessMilAircraft(&dashboard)

	// Keep the main goroutine alive indefinitely, or until an interrupt signal is received
	// In a real application, you might have other logic here or use a wait group.
	select {} // Block indefinitely
}

func requestAndProcessCivAircraft(dashboard *Dashboard) {
	// Define the URL for the HTTP GET request
	targetURL := fmt.Sprintf(
		"https://opendata.adsb.fi/api/v2/lat/%.6f/lon/%.6f/dist/250",
		Lat,
		Lon,
	)
	// This case is executed every time the ticker "ticks"
	body, requestErr := sendRequest(targetURL)
	if requestErr != nil {
		log.Printf("Error during request: %v", requestErr)
		return
	}
	processingErr := (*dashboard).ProcessCivAircraftJson(body)
	if processingErr != nil {
		log.Printf("Error during processing: %v", processingErr)
	}
}

func requestAndProcessMilAircraft(dashboard *Dashboard) {
	// Define the URL for the HTTP GET request
	targetURL := "https://opendata.adsb.fi/api/v2/mil"
	// This case is executed every time the ticker "ticks"
	body, requestErr := sendRequest(targetURL)
	if requestErr != nil {
		log.Printf("Error during request: %v", requestErr)
		return
	}
	processingErr := (*dashboard).ProcessMilAircraftJson(body)
	if processingErr != nil {
		log.Printf("Error during processing: %v", processingErr)
	}
}

// sendRequest sends an HTTP GET request and returns a valid byte slice of the response body
func sendRequest(url string) ([]byte, error) {
	resp, respErr := http.Get(url)
	if respErr != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", respErr)
	}
	defer func(bodyReader io.ReadCloser) {
		err := bodyReader.Close()
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

	// Debug output
	//fmt.Printf(
	//	"[%s] status: %s response body length: %d bytes\n",
	//	time.Now().Format("2006-01-02 15:04:05"),
	//	resp.Status,
	//	len(body),
	//)

	return body, nil
}
