package adsb

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	obs "github.com/micutio/airspottr/internal/domain/observation"
)

const (
	// AircraftUpdateInterval determines the update rate for general aircraft.
	AircraftUpdateInterval = 30 * time.Second
	// SummaryInterval determines how often the summary is show.
	SummaryInterval = 1 * time.Hour
	// DashboardWarmup determines how long to 'warm up' before showing rarity reports.
	DashboardWarmup = 1 * time.Hour

	reqHostAircraft = "opendata.adsb.fi"
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

// AircraftRequest handles http request commands.
// Should implement:
//   - application.AircraftRepository
type AircraftRequest struct {
	aircraftReqURL string
	apiClient      *http.Client
	waitGroup      sync.WaitGroup
	errOut         log.Logger
}

func NewAircraftRequest(opts RequestOptions, stderr *io.Writer) (*AircraftRequest, error) {
	aircraftReqURL, urlErr := createAircraftReqURL(opts)
	if urlErr != nil {
		return nil, fmt.Errorf("NewRequest: %w", urlErr)
	}

	client := &http.Client{
		Timeout: reqTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ //nolint:exhaustruct_v5 // too large
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS13,
			},
		},
	}

	request := &AircraftRequest{
		aircraftReqURL: aircraftReqURL,
		apiClient:      client,
		waitGroup:      sync.WaitGroup{},
		errOut:         *log.New(*stderr, "request ", log.LstdFlags),
	}

	request.errOut.Println("Request init")

	return request, nil
}

func createAircraftReqURL(opts RequestOptions) (string, error) {
	latStr := strconv.FormatFloat(opts.Lat, 'f', 6, 32)
	lonStr := strconv.FormatFloat(opts.Lon, 'f', 6, 32)
	baseURL := &url.URL{Scheme: reqProtocolHTTPS, Host: reqHostAircraft}
	fullURL := baseURL.JoinPath("api", "v2", "lat", latStr, "lon", lonStr, "dist", "250")
	targetURL := fullURL.String()
	validatedURL, valErr := validateURL(targetURL)
	if valErr != nil {
		return "", fmt.Errorf("sendRequest: error validating URL: %w", valErr)
	}
	return validatedURL, nil
}

func (r *AircraftRequest) RequestAircraft() []obs.AircraftRecord {
	body, requestErr := r.sendRequest(r.aircraftReqURL)
	if requestErr != nil {
		r.errOut.Println(fmt.Errorf("RequestAircraft: error during request: %w", requestErr))
		return []obs.AircraftRecord{}
	}

	var data obs.AircraftResult
	if err := json.Unmarshal(body, &data); err != nil {
		r.errOut.Println(fmt.Errorf("RequestAircraft: failed to unmarshal Json: %w", err))
		return []obs.AircraftRecord{}
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return []obs.AircraftRecord{} // Valid outcome, no need to log an error.
	}

	return data.Aircraft
}

// sendRequest builds the API URL from opts, sends an HTTP GET request, and returns the response body.
// The URL is constructed only from the fixed host and opts (lat/lon); no user-controlled URL input.
func (r *AircraftRequest) sendRequest(targetURL string) ([]byte, error) {
	body, requestSendErr := sendRequest(targetURL, r.apiClient)
	if requestSendErr != nil {
		return nil, fmt.Errorf("sendRequest: error sending request: %w", requestSendErr)
	}

	return body, nil
}
