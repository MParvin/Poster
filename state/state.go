package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const stateFileName = ".social_poster_state"

// ResolveStateFilePath returns the path used to persist the last processed commit SHA.
func ResolveStateFilePath(explicitPath, repoPath string) string {
	if explicitPath != "" {
		return explicitPath
	}
	if repoPath != "" {
		return filepath.Join(repoPath, stateFileName)
	}
	return stateFileName
}

// LoadLastProcessedSHA reads the last processed commit SHA from disk.
func LoadLastProcessedSHA(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read state file %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveLastProcessedSHA writes the last processed commit SHA to disk.
func SaveLastProcessedSHA(path, sha string) error {
	if sha == "" {
		return fmt.Errorf("cannot persist empty commit SHA")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state directory %s: %w", dir, err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(sha+"\n"), 0o600); err != nil {
		return fmt.Errorf("write temporary state file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace state file %s: %w", path, err)
	}
	return nil
}
