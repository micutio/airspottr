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
	if len(state.LocationStates) == 0 {
		state.LocationStates = map[string]rep.DashboardState{
			ref.LocationKey(state.DashboardState.Lat, state.DashboardState.Lon): state.DashboardState,
		}
	} else {
		state.LocationStates[ref.LocationKey(state.DashboardState.Lat, state.DashboardState.Lon)] = state.DashboardState
	}
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
	if len(internalState.LocationStates) == 0 &&
		(internalState.DashboardState.Lat != 0 ||
			internalState.DashboardState.Lon != 0 ||
			internalState.DashboardState.IsWarmup ||
			len(internalState.DashboardState.AircraftSightings) > 0) {
		internalState.LocationStates = map[string]rep.DashboardState{
			ref.LocationKey(internalState.DashboardState.Lat, internalState.DashboardState.Lon): internalState.DashboardState,
		}
	}
	state := rep.AirspottrState{InternalState: internalState}
	return state, nil
}

// StateForLocation returns the saved state for the provided latitude and longitude, or the current
// dashboard state when no location-specific history is available.
func StateForLocation(state rep.AirspottrState, latitude, longitude float64) rep.DashboardState {
	if state.InternalState.LocationStates != nil {
		if locationState, ok := state.InternalState.LocationStates[ref.LocationKey(latitude, longitude)]; ok {
			return locationState
		}
	}
	return state.InternalState.DashboardState
}
