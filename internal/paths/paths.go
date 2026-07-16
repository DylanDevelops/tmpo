package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// TmpoDir returns the base directory for tmpo's data. Normally this is
// ~/.tmpo, but when the TMPO_DEV environment variable is set to "1" or
// "true" a separate ~/.tmpo-dev directory is used so development runs do
// not touch production data.
func TmpoDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	dir := ".tmpo"
	if devMode := os.Getenv("TMPO_DEV"); devMode == "1" || devMode == "true" {
		dir = ".tmpo-dev"
	}

	return filepath.Join(home, dir), nil
}
