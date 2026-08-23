package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	internal "github.com/micutio/airspottr/internal"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	adsb "github.com/micutio/airspottr/internal/infrastructure/adsb"
)

const stateFileName = "airspottr_state.json"

// Errors used by the Dashboard.
var (
	errCoordMismatch = errors.New("state coordinate mismatch")
)

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
	IsWarmup           bool                              `json:"is_warmup"`
	Lat                float64                           `json:"lat"`
	Lon                float64                           `json:"lon"`
	CurrentAircraft    []obs.AircraftRecord              `json:"current_aircraft"`
	RareSightings      []persistedRareSighting           `json:"rare_sightings"`
	CachedFlightRoutes map[string]*ref.FlightRouteRecord `json:"cached_flight_routes"`
	AircraftSightings  map[string]obs.AircraftSighting   `json:"aircraft_sightings"`
	TotalTypeCount     int                               `json:"total_type_count"`
	TotalOperatorCount int                               `json:"total_operator_count"`
	TotalCountryCount  int                               `json:"total_country_count"`
	SeenTypeCount      map[string]int                    `json:"seen_type_count"`
	SeenOperatorCount  map[string]int                    `json:"seen_operator_count"`
	SeenCountryCount   map[string]int                    `json:"seen_country_count"`
}

type persistedRareSighting struct {
	Rarities obs.RarityFlag `json:"rarities"`
	Hex      string         `json:"hex"`
}

type requestState struct {
	PendingCallsigns []string `json:"pending_callsigns"`
}

func saveState(db *internal.Dashboard, pendingCallsigns []string) *persistentState {
	aircraftSightings := make(map[string]obs.AircraftSighting, len(db.AircraftSightings))
	sightingKeys := make(map[*obs.AircraftSighting]string, len(db.AircraftSightings))
	for hex, sighting := range db.AircraftSightings {
		if sighting == nil {
			continue
		}
		aircraftSightings[hex] = *sighting
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
			IsWarmup:           db.IsWarmup,
			Lat:                db.Lat,
			Lon:                db.Lon,
			CurrentAircraft:    db.CurrentAircraft,
			RareSightings:      raceSightings,
			CachedFlightRoutes: db.CachedFlightRoutes,
			AircraftSightings:  aircraftSightings,
			TotalTypeCount:     db.TotalTypeCount,
			TotalOperatorCount: db.TotalOperatorCount,
			TotalCountryCount:  db.TotalCountryCount,
			SeenTypeCount:      db.SeenTypeCount,
			SeenOperatorCount:  db.SeenOperatorCount,
			SeenCountryCount:   db.SeenCountryCount,
		},
		Request: requestState{PendingCallsigns: append([]string(nil), pendingCallsigns...)},
	}
}

func restoreState(r *adsb.Request, state requestState) {
	r.PendingCallsignsMu.Lock()
	defer r.PendingCallsignsMu.Unlock()
	r.PendingCallsigns = append([]string(nil), state.PendingCallsigns...)
}

func RestoreState(db *internal.Dashboard, state dashboardState) error {
	if state.Lat != db.Lat || state.Lon != db.Lon {
		return errCoordMismatch
	}

	db.IsWarmup = state.IsWarmup
	db.CurrentAircraft = state.CurrentAircraft
	db.CachedFlightRoutes = state.CachedFlightRoutes
	db.AircraftSightings = make(map[string]*obs.AircraftSighting, len(state.AircraftSightings))
	for hex, persisted := range state.AircraftSightings {
		db.AircraftSightings[hex] = &persisted
	}
	db.TotalTypeCount = state.TotalTypeCount
	db.TotalOperatorCount = state.TotalOperatorCount
	db.TotalCountryCount = state.TotalCountryCount
	db.SeenTypeCount = state.SeenTypeCount
	db.SeenOperatorCount = state.SeenOperatorCount
	db.SeenCountryCount = state.SeenCountryCount

	db.RareSightings = make([]obs.RareSighting, 0, len(state.RareSightings))
	for _, rare := range state.RareSightings {
		if sighting, ok := db.AircraftSightings[rare.Hex]; ok {
			db.RareSightings = append(db.RareSightings, obs.RareSighting{Rarities: rare.Rarities, Sighting: sighting})
		}
	}

	db.RecomputeFastestAndHighest()
	return nil
}

func SaveState(filePath string, db *internal.Dashboard, req *adsb.Request) error {
	req.PendingCallsignsMu.Lock()
	pendingCallsigns := append([]string(nil), req.PendingCallsigns...)
	req.PendingCallsignsMu.Unlock()
	state := saveState(db, pendingCallsigns)
	data, marshallErr := json.MarshalIndent(state, "", "  ")
	if marshallErr != nil {
		return fmt.Errorf("save state: marshal failed: %w", marshallErr)
	}
	if mkdirErr := os.MkdirAll(filepath.Dir(filePath), 0o700); mkdirErr != nil {
		return fmt.Errorf("save state: unable to create directory: %w", mkdirErr)
	}
	if writeFileErr := os.WriteFile(filePath, data, 0o600); writeFileErr != nil {
		return fmt.Errorf("save state: write failed: %w", writeFileErr)
	}
	return nil
}

func LoadState(filePath string, dashboard *internal.Dashboard, req *adsb.Request) error {
	data, readFileErr := os.ReadFile(filePath)
	if readFileErr != nil {
		if os.IsNotExist(readFileErr) {
			return nil
		}
		return fmt.Errorf("load state: unable to read file: %w", readFileErr)
	}
	var state persistentState
	if unmarshalErr := json.Unmarshal(data, &state); unmarshalErr != nil {
		return fmt.Errorf("load state: unmarshal failed: %w", unmarshalErr)
	}
	if restoreErr := RestoreState(dashboard, state.Dashboard); restoreErr != nil {
		return fmt.Errorf("load state: %w", restoreErr)
	}
	restoreState(req, state.Request)
	return nil
}
