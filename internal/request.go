package internal

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

	allowedRequestHost = "opendata.adsb.fi"
	requestTimeout     = 25 * time.Second
	// UrlAdsbOne         = "https://api.adsb.one/v2/point/%.6f/%.6f/%d"
	// UrlAdsbLol         = "https://api.adsb.lol/v2/lat/%.6f/lon/%.6f/dist/%d"
)

var (
	ErrNonOkResponse     = errors.New("non-OK response")
	ErrEmptyResponseBody = errors.New("empty response body")
	ErrNonJSONContent    = errors.New("non-JSON content type")
	ErrInvalidURL        = errors.New("invalid or insecure URL")
	ErrUnauthorizedHost  = errors.New("unauthorized host")
)

type RequestOptions struct {
	Lat float64
	Lon float64
}

func RequestAndProcessCivAircraft(opts RequestOptions) ([]byte, error) {
	body, requestErr := sendRequest(opts)
	if requestErr != nil {
		return nil, fmt.Errorf("requestAndProcessCivAircraft: error during request: %w", requestErr)
	}
	return body, nil
}

// sendRequest builds the API URL from opts, sends an HTTP GET request, and returns the response body.
// The URL is constructed only from the fixed host and opts (lat/lon); no user-controlled URL input.
func sendRequest(opts RequestOptions) ([]byte, error) {
	latStr := strconv.FormatFloat(float64(opts.Lat), 'f', 6, 32)
	lonStr := strconv.FormatFloat(float64(opts.Lon), 'f', 6, 32)
	baseURL := &url.URL{Scheme: "https", Host: allowedRequestHost}
	fullURL := baseURL.JoinPath("api", "v2", "lat", latStr, "lon", lonStr, "dist", "250")
	targetURL := fullURL.String()
	validatedURL, valErr := validateURL(targetURL)
	if valErr != nil {
		return nil, fmt.Errorf("sendRequest: error validating URL: %w", valErr)
	}

	ctx := context.Background()
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, validatedURL, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("sendRequest: invalid request error: %s : %w", targetURL, reqErr)
	}

	apiClient := &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ //nolint:exhaustruct // too large
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS13,
			},
		},
	}

	// TODO: Remove once fixed linter version is public
	resp, respErr := apiClient.Do(req) //nolint:gosec // linter bug
	if respErr != nil {
		return nil, fmt.Errorf("sendRequest: failed to send GET request: %s: %w", targetURL, respErr)
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
		return nil, fmt.Errorf("sendRequest: %w, %s", ErrNonJSONContent, contentType)
	}

	return body, nil
}

func validateURL(targetURL string) (string, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}

	if parsed.Host != allowedRequestHost {
		return "", ErrUnauthorizedHost
	}

	return targetURL, nil
}
