package persistence

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	ref "github.com/micutio/airspottr/internal/domain/reference"
	rep "github.com/micutio/airspottr/internal/domain/repositories"
)

var errNilState = errors.New("save state: state is nil")

const stateFileName = "airspottr_state.json"

// StateFilePath returns the platform-specific path to the persisted state file.
func StateFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return stateFileName
	}
	return filepath.Join(configDir, "airspottr", stateFileName)
}

func SaveState(filePath string, state *rep.PersistentState) error {
	if state == nil {
		return errNilState
	}
	if state.DashboardState != nil {
		if hasDashboardState(*state.DashboardState) {
			if state.LocationStates == nil {
				state.LocationStates = make(map[string]rep.DashboardState)
			}
			key := ref.LocationKey(state.DashboardState.Lat, state.DashboardState.Lon)
			if _, exists := state.LocationStates[key]; !exists {
				state.LocationStates[key] = *state.DashboardState
			}
		}
		state.DashboardState = nil
	}
	if existingState, err := LoadState(filePath); err == nil {
		if state.LocationStates == nil {
			state.LocationStates = make(map[string]rep.DashboardState)
		}
		for key, existing := range existingState.InternalState.LocationStates {
			if _, exists := state.LocationStates[key]; !exists {
				state.LocationStates[key] = existing
			}
		}
	}
	return saveStateFile(filePath, state)
}

func saveStateFile(filePath string, state *rep.PersistentState) error {
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

func LoadState(filePath string) (rep.AirspottrState, error) {
	data, readFileErr := os.ReadFile(filePath)
	defaultState := rep.AirspottrState{} //nolint:exhaustruct_v5 // using default values
	if readFileErr != nil {
		if os.IsNotExist(readFileErr) {
			return defaultState, nil
		}
		return defaultState, fmt.Errorf("load state: unable to read file: %w", readFileErr)
	}
	var internalState rep.PersistentState
	if unmarshalErr := json.Unmarshal(data, &internalState); unmarshalErr != nil {
		return defaultState, fmt.Errorf("load state: unmarshal failed: %w", unmarshalErr)
	}
	if internalState.LocationStates == nil {
		internalState.LocationStates = make(map[string]rep.DashboardState)
	}
	needsMigration := false
	if internalState.DashboardState != nil {
		if hasDashboardState(*internalState.DashboardState) {
			key := ref.LocationKey(internalState.DashboardState.Lat, internalState.DashboardState.Lon)
			if _, exists := internalState.LocationStates[key]; !exists {
				internalState.LocationStates[key] = *internalState.DashboardState
			}
		}
		internalState.DashboardState = nil
		needsMigration = true
	}
	if needsMigration {
		_ = saveStateFile(filePath, &internalState)
	}
	state := rep.AirspottrState{InternalState: internalState}
	return state, nil
}

// StateForLocation returns the saved state for the provided latitude and longitude.
// The boolean return value indicates whether saved state for that location exists.
func StateForLocation(state rep.AirspottrState, latitude, longitude float64) (rep.DashboardState, bool) {
	key := ref.LocationKey(latitude, longitude)
	if state.InternalState.LocationStates != nil {
		if locationState, ok := state.InternalState.LocationStates[key]; ok {
			return locationState, true
		}
	}
	if state.InternalState.DashboardState != nil &&
		ref.LocationKey(state.InternalState.DashboardState.Lat, state.InternalState.DashboardState.Lon) == key &&
		hasDashboardState(*state.InternalState.DashboardState) {
		return *state.InternalState.DashboardState, true
	}
	return rep.DashboardState{}, false
}

func hasDashboardState(d rep.DashboardState) bool {
	return d.Lat != 0 ||
		d.Lon != 0 ||
		d.IsWarmup ||
		len(d.AircraftSightings) > 0 ||
		d.TotalTypeCount > 0 ||
		d.TotalOperatorCount > 0 ||
		d.TotalCountryCount > 0 ||
		len(d.SeenTypeCount) > 0 ||
		len(d.SeenOperatorCount) > 0 ||
		len(d.SeenCountryCount) > 0 ||
		d.Fastest.Hex != "" ||
		d.Highest.Hex != ""
}
