package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	srv "github.com/micutio/airspottr/internal/application/services"
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

// AirspottrState encapsulates the internal state of the app to hide the exported JSON fields of
// persistentState.
type AirspottrState struct {
	internalState persistentState
}

type persistentState struct {
	DashboardState       dashboardState       `json:"dashboard"`
	FlightrouteRepoState flightrouteRepoState `json:"request"`
}

type dashboardState struct {
	IsWarmup        bool                 `json:"is_warmup"`
	Lat             float64              `json:"lat"`
	Lon             float64              `json:"lon"`
	Fastest         obs.AircraftRecord   `json:"fastest"`
	Highest         obs.AircraftRecord   `json:"highest"`
	CurrentAircraft []obs.AircraftRecord `json:"current_aircraft"`
	// Deprecated: no longer in use
	RareSightings []persistedRareSighting `json:"rare_sightings"`
	// Deprecated: no longer in use
	CachedFlightRoutes map[string]*ref.FlightrouteRecord `json:"cached_flight_routes"`
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

func saveState(dash *srv.Dashboard,
	pendingCallsigns []string,
) *persistentState {
	aircraftSightings := make(map[string]obs.AircraftSighting, len(dash.Sightings))
	sightingKeys := make(map[*obs.AircraftSighting]string, len(dash.Sightings))
	for hex, sighting := range dash.Sightings {
		aircraftSightings[hex] = sighting
		sightingKeys[&sighting] = hex
	}

	return &persistentState{
		DashboardState: dashboardState{ //nolint:exhaustruct_v5 // removed deprecated items
			IsWarmup:           dash.IsWarmup,
			Lat:                dash.Lat,
			Lon:                dash.Lon,
			Fastest:            dash.Fastest,
			Highest:            dash.Highest,
			CurrentAircraft:    nil,
			AircraftSightings:  aircraftSightings,
			TotalTypeCount:     dash.SightedTypesCount,
			TotalOperatorCount: dash.SightedOperatorsCount,
			TotalCountryCount:  dash.SightedCountriesCount,
			SeenTypeCount:      dash.SeenTypeCount,
			SeenOperatorCount:  dash.SeenOperatorCount,
			SeenCountryCount:   dash.SeenCountryCount,
		},
		FlightrouteRepoState: flightrouteRepoState{
			PendingCallsigns: append([]string(nil), pendingCallsigns...),
		},
	}
}

func restoreDashboardState(dash *srv.Dashboard, state dashboardState) error {
	if state.Lat != dash.Lat || state.Lon != dash.Lon {
		return errCoordMismatch
	}

	dash.IsWarmup = state.IsWarmup
	dash.Fastest = state.Fastest
	dash.Highest = state.Highest
	dash.Sightings = make(map[string]obs.AircraftSighting, len(state.AircraftSightings))
	maps.Copy(dash.Sightings, state.AircraftSightings)
	dash.SightedTypesCount = state.TotalTypeCount
	dash.SightedOperatorsCount = state.TotalOperatorCount
	dash.SightedCountriesCount = state.TotalCountryCount
	dash.SeenTypeCount = state.SeenTypeCount
	dash.SeenOperatorCount = state.SeenOperatorCount
	dash.SeenCountryCount = state.SeenCountryCount
	dash.CurrentSightings = nil

	return nil
}

func SaveState(filePath string, db *srv.Dashboard, frr repo.FlightrouteRepository) error {
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

func LoadState(filePath string) (AirspottrState, error) {
	data, readFileErr := os.ReadFile(filePath)
	defaultState := AirspottrState{} //nolint:exhaustruct_v5 // using default values
	if readFileErr != nil {
		if os.IsNotExist(readFileErr) {
			return defaultState, nil
		}
		return defaultState, fmt.Errorf("load state: unable to read file: %w", readFileErr)
	}
	var internalState persistentState
	if unmarshalErr := json.Unmarshal(data, &internalState); unmarshalErr != nil {
		return defaultState, fmt.Errorf("load state: unmarshal failed: %w", unmarshalErr)
	}

	state := AirspottrState{
		internalState: internalState,
	}

	return state, nil
}

func (as *AirspottrState) LoadDashboardState(
	dashboard *srv.Dashboard,
) error {
	if restoreErr := restoreDashboardState(dashboard, as.internalState.DashboardState); restoreErr != nil {
		return fmt.Errorf("load state: %w", restoreErr)
	}
	return nil
}
