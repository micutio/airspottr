package adsb

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

// See https://www.adsbexchange.com/version-2-api-wip/
// for further explanations of the fields

// aircraftResult mirrors the JSON which is returned for aircraft queries within a given distance.
type aircraftQueryResponse struct {
	Now         float64       `json:"now"`         // time this file was generated in [ms]
	ResultCount int           `json:"resultCount"` // total count of aircraft returned
	Ptime       float64       `json:"ptime"`       // server processing time required in [ms]
	Aircraft    []aircraftDTO `json:"aircraft"`    // list of Aircraft records
}

// AircraftRecord is used by both civilian and military aircraft queries.
type aircraftDTO struct {
	Alert           int      `json:"alert"`            // Flight status alert bit
	AltBaro         any      `json:"alt_baro"`         // altitude in [feet] or string "ground"
	AltGeom         int      `json:"alt_geom"`         // altitude in [feet]
	BaroRate        float64  `json:"baro_rate"`        // rate of change of baro alt in [feet/minute]
	EmitterCategory string   `json:"category"`         // emitter category to identify aircraft or vehicle classes (A0-D7)
	Emergency       string   `json:"emergency"`        // emergency/priority status, 7X00
	Flight          string   `json:"Flight"`           // Flight number, a.k.a. callsign
	GroundSpeed     float64  `json:"gs"`               // ground speed in [knots]
	Gva             float64  `json:"gva"`              // geometric vertical accuracy
	Hex             string   `json:"hex"`              // hex code ID for aircraft, assumed to be unique
	Lat             float64  `json:"lat"`              // Latitude in [decimal degrees]
	Lon             float64  `json:"lon"`              // Longitude in [decimal degrees]
	Messages        int      `json:"messages"`         // total number of Mode-S msg received from aircraft
	Mlat            []string `json:"mlat"`             // position calculation arrival time diffs
	NacP            float64  `json:"nac_p"`            // navigation accuracy for position
	NacV            float64  `json:"nac_v"`            // navigation accuracy for velocity
	NavAltitudeMcp  int      `json:"nav_altitude_mcp"` // selected from mode or Flight control panel (MCP)/(FCP) or other
	NavHeading      float64  `json:"nav_heading"`      // selected heading (True/Magnetic), magnetic is de-facto standard
	NavQNH          float64  `json:"nav_qnh"`          // altimeter setting (QFE  or QNH/QNE) in [hPa]
	Nic             int      `json:"nic"`              // Navigation Integrity Category
	NicBaro         int      `json:"nic_baro"`         // NIC for barometric altitude
	Registration    string   `json:"r"`                // Registration of the aircraft
	RadiusOfCtn     float64  `json:"rc"`               // Radius of containment, measure of position integrity in [meters]
	Rssi            float64  `json:"rssi"`             // recent average signal power, always negative, in [dbFS]
	Sda             int      `json:"sda"`              // system design assurance
	Seen            float64  `json:"seen"`             // last message received from aircraft in [seconds] from 'now'
	SeenPos         float64  `json:"seen_pos"`         // last update of position from aircraft in [seconds] from 'now'
	Sil             int      `json:"sil"`              // Source integrity level
	SilType         string   `json:"sil_type"`         // Source integrity level type
	Spi             int      `json:"spi"`              // Flight status special position identification bit
	Squawk          string   `json:"squawk"`           // Mode A code (Squawk) encoded as 4 octal digits
	IcaoType        string   `json:"t"`                // aircraft ICAO type pulled from database
	Tisb            []string `json:"tisb"`             // list of fields derived from TIS-B data
	Track           float64  `json:"track"`            // true track over ground in degrees (0-359)
	Type            string   `json:"type"`             // type of underlying messages
	Version         int      `json:"version"`          // ADS-B Version number 0,1,2 (3-7 are reserved)
	GeomRate        float64  `json:"geom_rate"`        // Rate of change of geometric (GNSS/INS) altitude in [ft/min]
	DBFlags         int      `json:"dbFlags"`          // bitfield for certain database flags (programming language)
	NavModes        []string `json:"nav_modes"`        // (autopilot, vnav, althold, approach, lnav, tcas)
	TrueHeading     float64  `json:"true_heading"`     // Heading clockwise from true north in [degrees]
	Ias             float64  `json:"ias"`              // indicated airspeed in [knots]
	Mach            float64  `json:"mach"`             // Mach number
	MagHeading      float64  `json:"mag_heading"`      // Heading clockwise from magnetic north in [degrees]
	Oat             float64  `json:"oat"`              // outer air temperature
	Roll            float64  `json:"roll"`             // roll, negative is left, in [degrees]
	Tas             float64  `json:"tas"`              // true airspeed in [knots]
	Tat             float32  `json:"tat"`              // total air temperature, might be inaccurate at lower alt, in [C]
	TrackRate       float64  `json:"track_rate"`       // rate of change of track in [degrees/second]
	WindDirection   float64  `json:"wd"`               // wind direction
	WindSpeed       float64  `json:"ws"`               // wind speed
	GpsOkBefore     float64  `json:"gpsOkBefore"`      // experimental, last timestamp of working GPS
	GpsOkLat        float64  `json:"gpsOkLat"`         // experimental, last timestamp of working Latitude
	GpsOkLon        float64  `json:"gpsOkLon"`         // experimental, last timestamp of working Longitude
	LastPosition    any      `json:"lastPosition"`     // TODO: Type
	RrLat           float64  `json:"rr_lat"`           // rough estimated latitude if no ADS-B or MLAT available
	RrLon           float64  `json:"rr_lon"`           // rough estimated longitude if no ADS-B or MLAT available
	CalcTrack       any      `json:"calc_track"`       // ? TODO
	NavAltitudeFMS  float64  `json:"nav_altitude_fms"` // selected altitude from the Flight management system (FMS)
	// found by my own investigation
	OwnOp       string `json:"ownOp"` // owner or operator, only rarely set
	Description string `json:"desc"`  // aircraft type description
	// cached data
	CachedDist float64
	CachedType string
}

func (dto aircraftDTO) toDomain() obs.AircraftRecord {
	return obs.AircraftRecord{
		Hex:          dto.Hex,
		Flight:       dto.Flight,
		Registration: dto.Registration,
		IcaoType:     dto.IcaoType,
		Description:  dto.Description,
		OwnOp:        dto.OwnOp,
		Lat:          dto.Lat,
		Lon:          dto.Lon,
		AltBaro:      dto.AltBaro,
		GroundSpeed:  dto.GroundSpeed,
		NavHeading:   dto.NavHeading,
		Seen:         dto.Seen,
	}
}

func mapAircraftRecords(aircraft []aircraftDTO) []obs.AircraftRecord {
	out := make([]obs.AircraftRecord, len(aircraft))
	for i := range aircraft {
		out[i] = aircraft[i].toDomain()
	}
	return out
}

// AircraftRequest handles http request commands.
type AircraftRequest struct {
	aircraftReqURL string
	apiClient      *http.Client
	waitGroup      sync.WaitGroup
	errOut         *slog.Logger
}

func NewAircraftRequest(opts RequestOptions, stderr io.Writer) (*AircraftRequest, error) {
	aircraftReqURL, urlErr := createAircraftReqURL(opts)
	if urlErr != nil {
		return nil, fmt.Errorf("NewRequest: %w", urlErr)
	}

	if stderr == nil {
		stderr = io.Discard
	}

	client := &http.Client{
		Timeout: reqTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS13,
			},
		},
	}

	request := &AircraftRequest{
		aircraftReqURL: aircraftReqURL,
		apiClient:      client,
		waitGroup:      sync.WaitGroup{},
		errOut:         slog.New(slog.NewTextHandler(stderr, nil)),
	}

	request.errOut.Info("Request init")

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
		r.errOut.Error("RequestAircraft: error during request", slog.Any("error", requestErr))
		return []obs.AircraftRecord{}
	}

	var data aircraftQueryResponse
	if err := json.Unmarshal(body, &data); err != nil {
		r.errOut.Error("RequestAircraft: failed to unmarshal Json", slog.Any("error", err))
		return []obs.AircraftRecord{}
	}

	foundAircraftCount := len(data.Aircraft)
	if foundAircraftCount == 0 {
		return []obs.AircraftRecord{} // Valid outcome, no need to log an error.
	}

	return mapAircraftRecords(data.Aircraft)
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
