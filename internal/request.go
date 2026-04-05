package internal

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// AircraftUpdateInterval determines the update rate for general aircraft.
	AircraftUpdateInterval = 30 * time.Second
	// SummaryInterval determines how often the summary is show.
	SummaryInterval = 1 * time.Hour
	// DashboardWarmup determines how long to 'warm up' before showing rarity reports.
	DashboardWarmup = 1 * time.Hour

	aircraftReqHost    = "opendata.adsb.fi"
	flightrouteReqHost = "api.adsbdb.com"

	requestTimeout = 25 * time.Second
	// FlightRouteQueryThreshold limits the number of concurrent flight route queries to avoid overwhelming the server.
	FlightRouteQueryThreshold = 10
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

// Request handles http request commands.
type Request struct {
	aircraftReqURL     string
	apiClient          *http.Client
	waitGroup          sync.WaitGroup
	errOut             log.Logger
	pendingCallsigns   []string
	pendingCallsignsMu sync.Mutex
}

func NewRequest(opts RequestOptions, stderr *io.Writer) (*Request, error) {
	aircraftReqURL, urlErr := createAircraftReqURL(opts)
	if urlErr != nil {
		return nil, fmt.Errorf("NewRequest: %w", urlErr)
	}

	client := &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ //nolint:exhaustruct // too large
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS13,
			},
		},
	}

	request := &Request{
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
	baseURL := &url.URL{Scheme: "https", Host: aircraftReqHost}
	fullURL := baseURL.JoinPath("api", "v2", "lat", latStr, "lon", lonStr, "dist", "250")
	targetURL := fullURL.String()
	validatedURL, valErr := validateURL(targetURL)
	if valErr != nil {
		return "", fmt.Errorf("sendRequest: error validating URL: %w", valErr)
	}
	return validatedURL, nil
}

func validateURL(targetURL string) (string, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}

	if parsed.Host != aircraftReqHost && parsed.Host != flightrouteReqHost {
		return "", ErrUnauthorizedHost
	}

	return targetURL, nil
}

func (r *Request) RequestAircraft() []AircraftRecord {
	body, requestErr := r.sendRequest(r.aircraftReqURL)
	if requestErr != nil {
		r.errOut.Println(fmt.Errorf("RequestAircraft: error during request: %w", requestErr))
		return []AircraftRecord{}
	}

	var data aircraftResult
	if err := json.Unmarshal(body, &data); err != nil {
		r.errOut.Println(fmt.Errorf("RequestAircraft: failed to unmarshal Json: %w", err))
		return []AircraftRecord{}
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return []AircraftRecord{} // Valid outcome, no need to log an error.
	}

	return data.Aircraft
}

func (r *Request) RequestFlightRoutesForCallsigns(callsigns []string) []FlightRouteRecord {
	r.pendingCallsignsMu.Lock()
	// Add new callsigns to the pending queue
	r.pendingCallsigns = append(r.pendingCallsigns, callsigns...)
	r.errOut.Printf("RequestFlightRoutesForCallsigns: %d callsigns requested, %d total pending\n", len(callsigns), len(r.pendingCallsigns))

	// Determine how many to process this time
	toProcess := FlightRouteQueryThreshold
	if len(r.pendingCallsigns) < toProcess {
		toProcess = len(r.pendingCallsigns)
	}

	// Take the first 'toProcess' callsigns from the queue
	selectedCallsigns := make([]string, toProcess)
	copy(selectedCallsigns, r.pendingCallsigns[:toProcess])

	// Remove the processed callsigns from the queue
	r.pendingCallsigns = r.pendingCallsigns[toProcess:]
	r.pendingCallsignsMu.Unlock()

	r.errOut.Printf("RequestFlightRoutesForCallsigns: processing %d callsigns this batch\n", len(selectedCallsigns))

	// 1. Build input urls for selected callsigns
	urlCount := len(selectedCallsigns)
	urls := make([]string, 0, urlCount)
	validCallsigns := make([]string, 0, urlCount)
	for _, callsign := range selectedCallsigns {
		callsignURL, urlErr := createFlightRouteRequestURL(callsign)
		if urlErr != nil {
			// Skip invalid urls.
			r.errOut.Println(
				fmt.Errorf(
					"RequestFlightRoutesForCallsigns: error constructing url for %s: %w",
					callsign, urlErr))
			continue
		}
		urls = append(urls, callsignURL)
		validCallsigns = append(validCallsigns, callsign)
	}

	results := make(chan []byte, len(urls))
	var waitGroup sync.WaitGroup

	// 2. Fan-out: Launch a goroutine for each URL
	for _, reqURL := range urls {
		waitGroup.Add(1)
		go func(urlStr string) {
			defer waitGroup.Done()

			body, reqErr := r.sendRequest(urlStr)
			// Only send body to results if there is no error.
			if reqErr != nil {
				r.errOut.Println(
					fmt.Errorf("RequestFlightRoutesForCallsigns: error requesting url: %s: %w",
						urlStr,
						reqErr))
			} else {
				results <- body
			}
		}(reqURL)
	}

	// 3. Wait and Close: Close the channel once all goroutines finish
	go func() {
		waitGroup.Wait()
		close(results)
	}()

	// 4. Fan-in: Collect and process results
	var flightrouteRecords []FlightRouteRecord
	for result := range results {
		flightrouteRecord, err := r.flightRouteJSONToRecord(result)
		if err != nil {
			r.errOut.Println(
				fmt.Errorf("RequestFlightRoutesForCallsigns: error parsing json: %w",
					err))
			continue
		}
		flightrouteRecords = append(flightrouteRecords, flightrouteRecord)
	}
	r.errOut.Printf(
		"RequestFlightRoutesForCallsigns: %d callsigns processed, %d routes found\n",
		len(selectedCallsigns), len(flightrouteRecords))
	return flightrouteRecords
}

func createFlightRouteRequestURL(callsign string) (string, error) {
	baseURL := &url.URL{Scheme: "https", Host: flightrouteReqHost}
	fullURL := baseURL.JoinPath("v0", "callsign", strings.TrimSpace(callsign))
	targetURL := fullURL.String()
	validatedURL, valErr := validateURL(targetURL)
	if valErr != nil {
		return "", fmt.Errorf("sendRequest: error validating URL: %w", valErr)
	}
	return validatedURL, nil
}

// flightRouteJSONToRecord takes a JSON record in form of a byte array and transforms it into a
// FlightRouteRecord.
// It is then assigned to all flights matching the callsign.
func (r *Request) flightRouteJSONToRecord(jsonBytes []byte) (FlightRouteRecord, error) {
	var data FlightrouteResponse
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		jsonErr := fmt.Errorf("RequestFlightRoutesForCallsigns: error parsing json: %w", err)
		r.errOut.Println(jsonErr)
		return data.Response.Flightroute, jsonErr
	}
	return data.Response.Flightroute, nil
}

// sendRequest builds the API URL from opts, sends an HTTP GET request, and returns the response body.
// The URL is constructed only from the fixed host and opts (lat/lon); no user-controlled URL input.
func (r *Request) sendRequest(targetURL string) ([]byte, error) {
	ctx := context.Background()
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("sendRequest: invalid request error: %s : %w", targetURL, reqErr)
	}

	resp, respErr := r.apiClient.Do(req)
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
