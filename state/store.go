package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	StatusSuccess = "success"
	deliveriesName = ".social_poster_deliveries.json"
)

// Store holds checkpoint and per-platform delivery records.
type Store struct {
	LastSHA    string            `json:"last_sha"`
	Deliveries map[string]string `json:"deliveries"`
	UpdatedAt  string            `json:"updated_at,omitempty"`
}

// DeliveryKey builds a stable key for a post/platform delivery.
func DeliveryKey(commitSHA, path, platform string) string {
	return commitSHA + "|" + path + "|" + platform
}

// ResolveDeliveriesPath returns the deliveries ledger path.
func ResolveDeliveriesPath(explicitPath, repoPath, stateFilePath string) string {
	if explicitPath != "" {
		return explicitPath
	}
	if stateFilePath != "" {
		return stateFilePath + ".deliveries.json"
	}
	if repoPath != "" {
		return filepath.Join(repoPath, deliveriesName)
	}
	return deliveriesName
}

// LoadStore reads the checkpoint SHA and deliveries ledger.
func LoadStore(statePath, deliveriesPath string) (*Store, error) {
	store := &Store{Deliveries: map[string]string{}}

	sha, err := LoadLastProcessedSHA(statePath)
	if err != nil {
		return nil, err
	}
	store.LastSHA = sha

	data, err := os.ReadFile(deliveriesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, fmt.Errorf("read deliveries file %s: %w", deliveriesPath, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return store, nil
	}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, fmt.Errorf("parse deliveries file %s: %w", deliveriesPath, err)
	}
	if store.Deliveries == nil {
		store.Deliveries = map[string]string{}
	}
	// Prefer the dedicated SHA file as source of truth when present.
	if sha != "" {
		store.LastSHA = sha
	}
	return store, nil
}

// IsDelivered reports whether a platform delivery already succeeded.
func (s *Store) IsDelivered(commitSHA, path, platform string) bool {
	if s == nil || s.Deliveries == nil {
		return false
	}
	return s.Deliveries[DeliveryKey(commitSHA, path, platform)] == StatusSuccess
}

// MarkDelivered records a successful platform delivery in memory.
func (s *Store) MarkDelivered(commitSHA, path, platform string) {
	if s.Deliveries == nil {
		s.Deliveries = map[string]string{}
	}
	s.Deliveries[DeliveryKey(commitSHA, path, platform)] = StatusSuccess
}

// SaveCheckpoint persists the last processed SHA.
func (s *Store) SaveCheckpoint(statePath, sha string) error {
	s.LastSHA = sha
	return SaveLastProcessedSHA(statePath, sha)
}

// SaveDeliveries atomically writes the deliveries ledger.
func (s *Store) SaveDeliveries(path string) error {
	if s.Deliveries == nil {
		s.Deliveries = map[string]string{}
	}
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	payload, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal deliveries: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create deliveries directory %s: %w", dir, err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write temporary deliveries file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace deliveries file %s: %w", path, err)
	}
	return nil
}
