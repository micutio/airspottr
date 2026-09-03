package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/micutio/airspottr/internal/application"
	obs "github.com/micutio/airspottr/internal/domain/observation"
	ref "github.com/micutio/airspottr/internal/domain/reference"
	repo "github.com/micutio/airspottr/internal/domain/repositories"
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
	Dashboard       dashboardState       `json:"dashboard"`
	FlightrouteRepo flightrouteRepoState `json:"request"`
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

type flightrouteRepoState struct {
	PendingCallsigns []string `json:"pending_callsigns"`
}

func saveState(dash *application.Dashboard, pendingCallsigns []string) *persistentState {
	aircraftSightings := make(map[string]obs.AircraftSighting, len(dash.AircraftSightings))
	sightingKeys := make(map[*obs.AircraftSighting]string, len(dash.AircraftSightings))
	for hex, sighting := range dash.AircraftSightings {
		if sighting == nil {
			continue
		}
		aircraftSightings[hex] = *sighting
		sightingKeys[sighting] = hex
	}

	raceSightings := make([]persistedRareSighting, 0, len(dash.RareSightings))
	for _, rare := range dash.RareSightings {
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
			IsWarmup:           dash.IsWarmup,
			Lat:                dash.Lat,
			Lon:                dash.Lon,
			CurrentAircraft:    dash.CurrentAircraft,
			RareSightings:      raceSightings,
			CachedFlightRoutes: dash.CachedFlightRoutes,
			AircraftSightings:  aircraftSightings,
			TotalTypeCount:     dash.TotalTypeCount,
			TotalOperatorCount: dash.TotalOperatorCount,
			TotalCountryCount:  dash.TotalCountryCount,
			SeenTypeCount:      dash.SeenTypeCount,
			SeenOperatorCount:  dash.SeenOperatorCount,
			SeenCountryCount:   dash.SeenCountryCount,
		},
		FlightrouteRepo: flightrouteRepoState{
			PendingCallsigns: append([]string(nil), pendingCallsigns...),
		},
	}
}

func restoreDashboardState(dash *application.Dashboard, state dashboardState) error {
	if state.Lat != dash.Lat || state.Lon != dash.Lon {
		return errCoordMismatch
	}

	dash.IsWarmup = state.IsWarmup
	dash.CurrentAircraft = state.CurrentAircraft
	dash.CachedFlightRoutes = state.CachedFlightRoutes
	dash.AircraftSightings = make(map[string]*obs.AircraftSighting, len(state.AircraftSightings))
	for hex, persisted := range state.AircraftSightings {
		dash.AircraftSightings[hex] = &persisted
	}
	dash.TotalTypeCount = state.TotalTypeCount
	dash.TotalOperatorCount = state.TotalOperatorCount
	dash.TotalCountryCount = state.TotalCountryCount
	dash.SeenTypeCount = state.SeenTypeCount
	dash.SeenOperatorCount = state.SeenOperatorCount
	dash.SeenCountryCount = state.SeenCountryCount

	dash.RareSightings = make([]obs.RareSighting, 0, len(state.RareSightings))
	for _, rare := range state.RareSightings {
		if sighting, ok := dash.AircraftSightings[rare.Hex]; ok {
			dash.RareSightings = append(
				dash.RareSightings,
				obs.RareSighting{Rarities: rare.Rarities, Sighting: sighting},
			)
		}
	}

	dash.RecomputeFastestAndHighest()
	return nil
}

func SaveState(filePath string, db *application.Dashboard, frr repo.FlightrouteRepository) error {
	pendingCallsigns := frr.GetPendingCallsigns()
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

func LoadState(
	filePath string,
	dashboard *application.Dashboard,
	frr repo.FlightrouteRepository,
) error {
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
	if restoreErr := restoreDashboardState(dashboard, state.Dashboard); restoreErr != nil {
		return fmt.Errorf("load state: %w", restoreErr)
	}
	frr.RestorePendingCallsigns(state.FlightrouteRepo.PendingCallsigns)
	return nil
}
