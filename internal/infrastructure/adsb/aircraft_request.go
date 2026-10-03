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

type aircraftQueryResponse struct {
	Now         float64       `json:"now"`
	ResultCount int           `json:"resultCount"`
	Ptime       float64       `json:"ptime"`
	Aircraft    []aircraftDTO `json:"aircraft"`
}

type aircraftDTO struct {
	Alert           int      `json:"alert"`
	AltBaro         any      `json:"alt_baro"`
	AltGeom         int      `json:"alt_geom"`
	BaroRate        float64  `json:"baro_rate"`
	EmitterCategory string   `json:"category"`
	Emergency       string   `json:"emergency"`
	Flight          string   `json:"Flight"`
	GroundSpeed     float64  `json:"gs"`
	Gva             float64  `json:"gva"`
	Hex             string   `json:"hex"`
	Lat             float64  `json:"lat"`
	Lon             float64  `json:"lon"`
	Messages        int      `json:"messages"`
	Mlat            []string `json:"mlat"`
	NacP            float64  `json:"nac_p"`
	NacV            float64  `json:"nac_v"`
	NavAltitudeMcp  int      `json:"nav_altitude_mcp"`
	NavHeading      float64  `json:"nav_heading"`
	NavQNH          float64  `json:"nav_qnh"`
	Nic             int      `json:"nic"`
	NicBaro         int      `json:"nic_baro"`
	Registration    string   `json:"r"`
	RadiusOfCtn     float64  `json:"rc"`
	Rssi            float64  `json:"rssi"`
	Sda             int      `json:"sda"`
	Seen            float64  `json:"seen"`
	SeenPos         float64  `json:"seen_pos"`
	Sil             int      `json:"sil"`
	SilType         string   `json:"sil_type"`
	Spi             int      `json:"spi"`
	Squawk          string   `json:"squawk"`
	IcaoType        string   `json:"t"`
	Tisb            []string `json:"tisb"`
	Track           float64  `json:"track"`
	Type            string   `json:"type"`
	Version         int      `json:"version"`
	GeomRate        float64  `json:"geom_rate"`
	DBFlags         int      `json:"dbFlags"`
	NavModes        []string `json:"nav_modes"`
	TrueHeading     float64  `json:"true_heading"`
	Ias             float64  `json:"ias"`
	Mach            float64  `json:"mach"`
	MagHeading      float64  `json:"mag_heading"`
	Oat             float64  `json:"oat"`
	Roll            float64  `json:"roll"`
	Tas             float64  `json:"tas"`
	Tat             float32  `json:"tat"`
	TrackRate       float64  `json:"track_rate"`
	WindDirection   float64  `json:"wd"`
	WindSpeed       float64  `json:"ws"`
	GpsOkBefore     float64  `json:"gpsOkBefore"`
	GpsOkLat        float64  `json:"gpsOkLat"`
	GpsOkLon        float64  `json:"gpsOkLon"`
	LastPosition    any      `json:"lastPosition"`
	RrLat           float64  `json:"rr_lat"`
	RrLon           float64  `json:"rr_lon"`
	CalcTrack       any      `json:"calc_track"`
	NavAltitudeFMS  float64  `json:"nav_altitude_fms"`
	OwnOp           string   `json:"ownOp"`
	Description     string   `json:"desc"`
}

func (dto aircraftDTO) toDomain() obs.AircraftRecord {
	return obs.AircraftRecord{
		Alert:           dto.Alert,
		AltBaro:         dto.AltBaro,
		AltGeom:         dto.AltGeom,
		BaroRate:        dto.BaroRate,
		EmitterCategory: dto.EmitterCategory,
		Emergency:       dto.Emergency,
		Flight:          dto.Flight,
		GroundSpeed:     dto.GroundSpeed,
		Gva:             dto.Gva,
		Hex:             dto.Hex,
		Lat:             dto.Lat,
		Lon:             dto.Lon,
		Messages:        dto.Messages,
		Mlat:            dto.Mlat,
		NacP:            dto.NacP,
		NacV:            dto.NacV,
		NavAltitudeMcp:  dto.NavAltitudeMcp,
		NavHeading:      dto.NavHeading,
		NavQNH:          dto.NavQNH,
		Nic:             dto.Nic,
		NicBaro:         dto.NicBaro,
		Registration:    dto.Registration,
		RadiusOfCtn:     dto.RadiusOfCtn,
		Rssi:            dto.Rssi,
		Sda:             dto.Sda,
		Seen:            dto.Seen,
		SeenPos:         dto.SeenPos,
		Sil:             dto.Sil,
		SilType:         dto.SilType,
		Spi:             dto.Spi,
		Squawk:          dto.Squawk,
		IcaoType:        dto.IcaoType,
		Tisb:            dto.Tisb,
		Track:           dto.Track,
		Type:            dto.Type,
		Version:         dto.Version,
		GeomRate:        dto.GeomRate,
		DBFlags:         dto.DBFlags,
		NavModes:        dto.NavModes,
		TrueHeading:     dto.TrueHeading,
		Ias:             dto.Ias,
		Mach:            dto.Mach,
		MagHeading:      dto.MagHeading,
		Oat:             dto.Oat,
		Roll:            dto.Roll,
		Tas:             dto.Tas,
		Tat:             dto.Tat,
		TrackRate:       dto.TrackRate,
		WindDirection:   dto.WindDirection,
		WindSpeed:       dto.WindSpeed,
		GpsOkBefore:     dto.GpsOkBefore,
		GpsOkLat:        dto.GpsOkLat,
		GpsOkLon:        dto.GpsOkLon,
		LastPosition:    dto.LastPosition,
		RrLat:           dto.RrLat,
		RrLon:           dto.RrLon,
		CalcTrack:       dto.CalcTrack,
		NavAltitudeFMS:  dto.NavAltitudeFMS,
		OwnOp:           dto.OwnOp,
		Description:     dto.Description,
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
// Should implement:
//   - application.AircraftRepository
type AircraftRequest struct {
	aircraftReqURL string
	apiClient      *http.Client
	waitGroup      sync.WaitGroup
	errOut         log.Logger
}

func NewAircraftRequest(opts RequestOptions, stderr io.Writer) (*AircraftRequest, error) {
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
		errOut:         *log.New(stderr, "request ", log.LstdFlags),
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

	var data aircraftQueryResponse
	if err := json.Unmarshal(body, &data); err != nil {
		r.errOut.Println(fmt.Errorf("RequestAircraft: failed to unmarshal Json: %w", err))
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
