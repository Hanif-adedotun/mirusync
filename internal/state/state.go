package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SyncState struct {
	LastSync     time.Time `json:"last_sync"`
	LastDirection string   `json:"last_direction"` // "push", "pull", "sync"
	LastHost     string    `json:"last_host"`
	FileCount    int       `json:"file_count"`
	Conflicts    []string  `json:"conflicts,omitempty"`
}

func GetStateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	stateDir := filepath.Join(home, ".mirusync", "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return "", err
	}
	return stateDir, nil
}

func GetStatePath(folderName string) (string, error) {
	stateDir, err := GetStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, folderName+".json"), nil
}

func LoadState(folderName string) (*SyncState, error) {
	statePath, err := GetStatePath(folderName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(statePath)
	if os.IsNotExist(err) {
		// No previous state
		return &SyncState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	return &state, nil
}

func SaveState(folderName string, state *SyncState) error {
	statePath, err := GetStatePath(folderName)
	if err != nil {
		return err
	}

	state.LastSync = time.Now()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func RecordConflict(folderName string, conflictPath string) error {
	state, err := LoadState(folderName)
	if err != nil {
		return err
	}

	// Check if conflict already recorded
	for _, existing := range state.Conflicts {
		if existing == conflictPath {
			return nil // Already recorded
		}
	}

	state.Conflicts = append(state.Conflicts, conflictPath)
	return SaveState(folderName, state)
}

func ClearConflicts(folderName string) error {
	state, err := LoadState(folderName)
	if err != nil {
		return err
	}

	state.Conflicts = []string{}
	return SaveState(folderName, state)
}

