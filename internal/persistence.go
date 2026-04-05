package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const stateFileName = "airspottr_state.json"

// StateFilePath returns the platform-specific path to the persisted state file.
func StateFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return stateFileName
	}
	return filepath.Join(configDir, "airspottr", stateFileName)
}

type persistentState struct {
	Dashboard dashboardState `json:"dashboard"`
	Request   requestState   `json:"request"`
}

type dashboardState struct {
	IsWarmup           bool                                 `json:"is_warmup"`
	Lat                float64                              `json:"lat"`
	Lon                float64                              `json:"lon"`
	CurrentAircraft    []AircraftRecord                     `json:"current_aircraft"`
	RareSightings      []persistedRareSighting              `json:"rare_sightings"`
	CachedFlightRoutes map[string]*FlightRouteRecord        `json:"cached_flight_routes"`
	AircraftSightings  map[string]persistedAircraftSighting `json:"aircraft_sightings"`
	TotalTypeCount     int                                  `json:"total_type_count"`
	TotalOperatorCount int                                  `json:"total_operator_count"`
	TotalCountryCount  int                                  `json:"total_country_count"`
	SeenTypeCount      map[string]int                       `json:"seen_type_count"`
	SeenOperatorCount  map[string]int                       `json:"seen_operator_count"`
	SeenCountryCount   map[string]int                       `json:"seen_country_count"`
}

type persistedRareSighting struct {
	Rarities RarityFlag `json:"rarities"`
	Hex      string     `json:"hex"`
}

type persistedAircraftSighting struct {
	LastSeen     time.Time          `json:"last_seen"`
	LastFlightNo string             `json:"last_flight_no"`
	Registration string             `json:"registration"`
	Latitude     float64            `json:"latitude"`
	Longitude    float64            `json:"longitude"`
	Direction    string             `json:"direction"`
	Distance     float64            `json:"distance"`
	TypeShort    string             `json:"type_short"`
	TypeDesc     string             `json:"type_desc"`
	Operator     string             `json:"operator"`
	Country      string             `json:"country"`
	Info         string             `json:"info"`
	Flightroute  *FlightRouteRecord `json:"flightroute"`
}

type requestState struct {
	PendingCallsigns []string `json:"pending_callsigns"`
}

func (s *persistedAircraftSighting) toAircraftSighting() *AircraftSighting {
	return &AircraftSighting{
		lastSeen:     s.LastSeen,
		lastFlightNo: s.LastFlightNo,
		registration: s.Registration,
		latitude:     s.Latitude,
		longitude:    s.Longitude,
		direction:    s.Direction,
		distance:     s.Distance,
		typeShort:    s.TypeShort,
		typeDesc:     s.TypeDesc,
		operator:     s.Operator,
		country:      s.Country,
		info:         s.Info,
		flightroute:  s.Flightroute,
	}
}

func (s *AircraftSighting) persisted() persistedAircraftSighting {
	return persistedAircraftSighting{
		LastSeen:     s.lastSeen,
		LastFlightNo: s.lastFlightNo,
		Registration: s.registration,
		Latitude:     s.latitude,
		Longitude:    s.longitude,
		Direction:    s.direction,
		Distance:     s.distance,
		TypeShort:    s.typeShort,
		TypeDesc:     s.typeDesc,
		Operator:     s.operator,
		Country:      s.country,
		Info:         s.info,
		Flightroute:  s.flightroute,
	}
}

func (db *Dashboard) saveState(pendingCallsigns []string) *persistentState {
	aircraftSightings := make(map[string]persistedAircraftSighting, len(db.aircraftSightings))
	sightingKeys := make(map[*AircraftSighting]string, len(db.aircraftSightings))
	for hex, sighting := range db.aircraftSightings {
		if sighting == nil {
			continue
		}
		aircraftSightings[hex] = sighting.persisted()
		sightingKeys[sighting] = hex
	}

	raceSightings := make([]persistedRareSighting, 0, len(db.RareSightings))
	for _, rare := range db.RareSightings {
		if rare.Sighting == nil {
			continue
		}
		if hex, ok := sightingKeys[rare.Sighting]; ok {
			raceSightings = append(raceSightings, persistedRareSighting{
				Rarities: rare.Rarities,
				Hex:      hex,
			})
		}
	}

	return &persistentState{
		Dashboard: dashboardState{
			IsWarmup:           db.isWarmup,
			Lat:                db.Lat,
			Lon:                db.Lon,
			CurrentAircraft:    db.CurrentAircraft,
			RareSightings:      raceSightings,
			CachedFlightRoutes: db.CachedFlightRoutes,
			AircraftSightings:  aircraftSightings,
			TotalTypeCount:     db.totalTypeCount,
			TotalOperatorCount: db.totalOperatorCount,
			TotalCountryCount:  db.totalCountryCount,
			SeenTypeCount:      db.SeenTypeCount,
			SeenOperatorCount:  db.SeenOperatorCount,
			SeenCountryCount:   db.SeenCountryCount,
		},
		Request: requestState{PendingCallsigns: append([]string(nil), pendingCallsigns...)},
	}
}

func (r *Request) RestoreState(state requestState) {
	r.pendingCallsignsMu.Lock()
	defer r.pendingCallsignsMu.Unlock()
	r.pendingCallsigns = append([]string(nil), state.PendingCallsigns...)
}

func (db *Dashboard) RestoreState(state dashboardState) error {
	if state.Lat != db.Lat || state.Lon != db.Lon {
		return fmt.Errorf("state coordinates mismatch")
	}

	db.isWarmup = state.IsWarmup
	db.CurrentAircraft = state.CurrentAircraft
	db.CachedFlightRoutes = state.CachedFlightRoutes
	db.aircraftSightings = make(map[string]*AircraftSighting, len(state.AircraftSightings))
	for hex, persisted := range state.AircraftSightings {
		db.aircraftSightings[hex] = persisted.toAircraftSighting()
	}
	db.totalTypeCount = state.TotalTypeCount
	db.totalOperatorCount = state.TotalOperatorCount
	db.totalCountryCount = state.TotalCountryCount
	db.SeenTypeCount = state.SeenTypeCount
	db.SeenOperatorCount = state.SeenOperatorCount
	db.SeenCountryCount = state.SeenCountryCount

	db.RareSightings = make([]RareSighting, 0, len(state.RareSightings))
	for _, rare := range state.RareSightings {
		if sighting, ok := db.aircraftSightings[rare.Hex]; ok {
			db.RareSightings = append(db.RareSightings, RareSighting{Rarities: rare.Rarities, Sighting: sighting})
		}
	}

	db.recomputeFastestAndHighest()
	return nil
}

func SaveState(filePath string, db *Dashboard, req *Request) error {
	req.pendingCallsignsMu.Lock()
	pendingCallsigns := append([]string(nil), req.pendingCallsigns...)
	req.pendingCallsignsMu.Unlock()
	state := db.saveState(pendingCallsigns)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("save state: marshal failed: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
		return fmt.Errorf("save state: unable to create directory: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("save state: write failed: %w", err)
	}
	return nil
}

func LoadState(filePath string, db *Dashboard, req *Request) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load state: unable to read file: %w", err)
	}
	var state persistentState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("load state: unmarshal failed: %w", err)
	}
	if err := db.RestoreState(state.Dashboard); err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	req.RestoreState(state.Request)
	return nil
}
