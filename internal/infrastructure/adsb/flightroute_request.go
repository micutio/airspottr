package adsb

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log" //nolint:depguard // Don't feel like using slog
	"net/http"
	"net/url"
	"strings"
	"sync"

	ref "github.com/micutio/airspottr/internal/domain/reference"
)

const (
	reqHostFlightroute = "api.adsbdb.com"
	// FlightRouteQueryThreshold limits the number of concurrent flight route queries to avoid overwhelming the server.
	FlightRouteQueryThreshold = 10
)

// FlightrouteRequest handles http request commands.
// Should implement:
//   - application.FlightrouteRepository
type FlightrouteRequest struct {
	apiClient            *http.Client
	waitGroup            sync.WaitGroup
	errOut               log.Logger
	pendingCallsigns     []string
	pendingCallsignsMu   sync.Mutex
	cachedFlightroutes   map[string]ref.FlightrouteRecord
	cachedFlightroutesMu sync.Mutex
}

func NewFlightrouteRequest(stderr *io.Writer) (*FlightrouteRequest, error) {
	client := &http.Client{
		Timeout: reqTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ //nolint:exhaustruct_v5 // too large
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS13,
			},
		},
	}

	request := &FlightrouteRequest{
		apiClient:            client,
		waitGroup:            sync.WaitGroup{},
		errOut:               *log.New(*stderr, "request ", log.LstdFlags),
		pendingCallsigns:     []string{},
		pendingCallsignsMu:   sync.Mutex{},
		cachedFlightroutes:   make(map[string]ref.FlightrouteRecord),
		cachedFlightroutesMu: sync.Mutex{},
	}

	request.errOut.Println("Request init")

	return request, nil
}

func (r *FlightrouteRequest) GetFlightroutes(callsigns []string) map[string]ref.FlightrouteRecord {
	// Retrieve routes from cache first, if available.
	r.cachedFlightroutesMu.Lock()
	flightrouteRecords := make(map[string]ref.FlightrouteRecord)
	var callsignsWithoutRoute []string
	for _, callsign := range callsigns {
		if route, exists := r.cachedFlightroutes[callsign]; exists {
			flightrouteRecords[callsign] = route
		} else {
			callsignsWithoutRoute = append(callsignsWithoutRoute, callsign)
		}
	}
	r.cachedFlightroutesMu.Unlock()

	// Best case: all routes have been found, return right here
	if len(callsignsWithoutRoute) == 0 {
		return flightrouteRecords
	}

	// For missing routes make a web request to adsbdb.
	r.requestFlightroutesFromWeb(flightrouteRecords, callsignsWithoutRoute)
	return flightrouteRecords
}

func (r *FlightrouteRequest) requestFlightroutesFromWeb(
	flightrouteRecords map[string]ref.FlightrouteRecord,
	callsigns []string,
) {
	r.pendingCallsignsMu.Lock()
	// Add new callsigns to the pending queue
	r.pendingCallsigns = append(r.pendingCallsigns, callsigns...)
	r.errOut.Printf(
		"RequestFlightRoutesForCallsigns: %d callsigns requested, %d total pending\n",
		len(callsigns),
		len(r.pendingCallsigns),
	)

	// Determine how many to process this time
	toProcess := min(len(r.pendingCallsigns), FlightRouteQueryThreshold)

	// Take the first 'toProcess' callsigns from the queue
	selectedCallsigns := make([]string, toProcess)
	copy(selectedCallsigns, r.pendingCallsigns[:toProcess])

	// Remove the processed callsigns from the queue
	r.pendingCallsigns = r.pendingCallsigns[toProcess:]
	r.pendingCallsignsMu.Unlock()

	r.errOut.Printf("RequestFlightRoutesForCallsigns: processing %d callsigns this batch\n", len(selectedCallsigns))

	// 0. Put dummies for the selected callsigns into the cache, so that we do not query for them
	// again if the database does not have them.
	for _, callsign := range selectedCallsigns {
		r.cachedFlightroutesMu.Lock()
		r.cachedFlightroutes[callsign] = *ref.GetDefaultFlightrouteRecord()
		r.cachedFlightroutesMu.Unlock()
	}

	// 1. Build input urls for selected callsigns
	urls := r.createFlightrouteRequestURLs(selectedCallsigns)

	// 2. Fan-out: Launch a goroutine for each URL
	results := make(chan []byte, len(urls))
	var waitGroup sync.WaitGroup
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
	for result := range results {
		flightrouteRecord, err := r.flightRouteJSONToRecord(result)
		if err != nil {
			r.errOut.Println(
				fmt.Errorf("RequestFlightRoutesForCallsigns: error parsing json: %w",
					err))
			continue
		}
		flightrouteRecords[flightrouteRecord.Callsign] = flightrouteRecord
		// Cache the found flightroutes.
		r.cachedFlightroutesMu.Lock()
		r.cachedFlightroutes[flightrouteRecord.Callsign] = flightrouteRecord
		r.cachedFlightroutesMu.Unlock()
	}
	r.errOut.Printf(
		"RequestFlightRoutesForCallsigns: %d callsigns processed, %d routes found\n",
		len(callsigns),
		len(flightrouteRecords))
	r.errOut.Printf(
		"RequestFlightRoutesForCallsigns: total flightroutes cached: %d\n",
		len(r.cachedFlightroutes))
}

func (r *FlightrouteRequest) createFlightrouteRequestURLs(selectedCallsigns []string) []string {
	urlCount := len(selectedCallsigns)
	urls := make([]string, 0, urlCount)
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
	}
	return urls
}

func createFlightRouteRequestURL(callsign string) (string, error) {
	baseURL := &url.URL{Scheme: reqProtocolHTTPS, Host: reqHostFlightroute}
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
func (r *FlightrouteRequest) flightRouteJSONToRecord(jsonBytes []byte) (ref.FlightrouteRecord, error) {
	var data ref.FlightrouteResponse
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		jsonErr := fmt.Errorf("RequestFlightRoutesForCallsigns: error parsing json: %w", err)
		r.errOut.Println(jsonErr)
		return data.Response.Flightroute, jsonErr
	}
	return data.Response.Flightroute, nil
}

// sendRequest builds the API URL from opts, sends an HTTP GET request, and returns the response body.
// The URL is constructed only from the fixed host and opts (lat/lon); no user-controlled URL input.
func (r *FlightrouteRequest) sendRequest(targetURL string) ([]byte, error) {
	body, requestSendErr := sendRequest(targetURL, r.apiClient)
	if requestSendErr != nil {
		return nil, fmt.Errorf("sendRequest: error sending request: %w", requestSendErr)
	}

	return body, nil
}

func (r *FlightrouteRequest) ClearFlightrouteCache() {
	r.cachedFlightroutes = make(map[string]ref.FlightrouteRecord)
}

func (r *FlightrouteRequest) GetPendingCallsigns() []string {
	r.pendingCallsignsMu.Lock()
	defer r.pendingCallsignsMu.Unlock()
	cp := make([]string, len(r.pendingCallsigns))
	copy(cp, r.pendingCallsigns)
	return cp
}

func (r *FlightrouteRequest) RestorePendingCallsigns(pendingCallsigns []string) {
	r.pendingCallsignsMu.Lock()
	defer r.pendingCallsignsMu.Unlock()
	r.pendingCallsigns = append([]string(nil), pendingCallsigns...)
}
