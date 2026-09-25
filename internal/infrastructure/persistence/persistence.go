package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	srv "github.com/micutio/airspottr/internal/application/services"
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

func SaveState(filePath string, state *srv.PersistentState) error {
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

func LoadState(filePath string) (srv.AirspottrState, error) {
	data, readFileErr := os.ReadFile(filePath)
	defaultState := srv.AirspottrState{} //nolint:exhaustruct_v5 // using default values
	if readFileErr != nil {
		if os.IsNotExist(readFileErr) {
			return defaultState, nil
		}
		return defaultState, fmt.Errorf("load state: unable to read file: %w", readFileErr)
	}
	var internalState srv.PersistentState
	if unmarshalErr := json.Unmarshal(data, &internalState); unmarshalErr != nil {
		return defaultState, fmt.Errorf("load state: unmarshal failed: %w", unmarshalErr)
	}

	state := srv.AirspottrState{
		InternalState: internalState,
	}

	return state, nil
}
