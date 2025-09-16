package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// AircraftUpdateInterval determines the update rate for general aircraft.
	AircraftUpdateInterval = 30 * time.Second
	// SummaryInterval determines how often the summary is show.
	SummaryInterval = 1 * time.Hour
	// DashboardWarmup determines how long to 'warm up' before showing rarity reports.
	DashboardWarmup = 1 * time.Hour

	URLAdsbOpenData = "https://opendata.adsb.fi/api/v2/lat/%.6f/lon/%.6f/dist/250"
	// UrlAdsbOne         = "https://api.adsb.one/v2/point/%.6f/%.6f/%d"
	// UrlAdsbLol         = "https://api.adsb.lol/v2/lat/%.6f/lon/%.6f/dist/%d"
)

var (
	ErrNonOkResponse     = errors.New("non-OK response")
	ErrEmptyResponseBody = errors.New("empty response body")
	ErrNonJSONContent    = errors.New("non-JSON content type")
)

type RequestOptions struct {
	Lat float32
	Lon float32
}

func RequestAndProcessCivAircraft(opts RequestOptions) ([]byte, error) {
	// Define the URL for the HTTP GET request
	targetURL := fmt.Sprintf(URLAdsbOpenData, opts.Lat, opts.Lon)

	// This case is executed every time the ticker "ticks"
	body, requestErr := sendRequest(targetURL)
	if requestErr != nil {
		return nil, fmt.Errorf("requestAndProcessCivAircraft: error during request: %w", requestErr)
	}

	return body, nil
}

// sendRequest sends an HTTP GET request and returns a valid byte slice of the response body.
func sendRequest(url string) ([]byte, error) {
	ctx := context.Background()
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("sendRequest: invalid request error: %s : %w", url, reqErr)
	}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("sendRequest: failed to send GET request: %s: %w", url, respErr)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			respErr = fmt.Errorf("sendRequest: error while closing response body: %w", closeErr)
		}
	}()

	// Check if the request was successful (status code 200 OK)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sendRequest: %w %s", ErrNonOkResponse, resp.Status)
	}

	// Read the response body
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", bodyErr)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("sendRequest: %w", ErrEmptyResponseBody)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf("senRequest: %w, %s", ErrNonJSONContent, contentType)
	}

	return body, nil
}
