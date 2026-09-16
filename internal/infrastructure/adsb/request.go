package adsb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	reqProtocolHTTPS = "https"
	reqTimeout       = 25 * time.Second
)

func validateURL(targetURL string) (string, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Scheme != reqProtocolHTTPS {
		return "", ErrInvalidURL
	}

	if parsed.Host != reqHostAircraft {
		return "", ErrUnauthorizedHost
	}

	return targetURL, nil
}

// sendRequest builds the API URL from opts, sends an HTTP GET request, and returns the response body.
// The URL is constructed only from the fixed host and opts (lat/lon); no user-controlled URL input.
func sendRequest(targetURL string, apiClient *http.Client) ([]byte, error) {
	ctx := context.Background()
	req, reqCreateErr := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if reqCreateErr != nil {
		return nil, fmt.Errorf(
			"sendRequest: invalid request error: %s : %w",
			targetURL,
			reqCreateErr)
	}

	resp, respErr := apiClient.Do(req)
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
