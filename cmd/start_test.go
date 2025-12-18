package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectProjectName(t *testing.T) {
	t.Run("returns project name from .tmporc config", func(t *testing.T) {
		// Create a temporary directory with a .tmporc file
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create a .tmporc file with a project name
		cfg := &config.Config{
			ProjectName: "test-project-from-config",
			HourlyRate:  75.0,
		}
		err = cfg.Save(filepath.Join(tmpDir, ".tmporc"))
		require.NoError(t, err)

		// Test
		projectName, err := DetectProjectName()
		assert.NoError(t, err)
		assert.Equal(t, "test-project-from-config", projectName)
	})

	t.Run("falls back to git repository name when no .tmporc", func(t *testing.T) {
		// Create a temporary directory
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Initialize a git repository
		err = os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)
		require.NoError(t, err)

		// Test - should use directory name as fallback since it's not a real git repo
		projectName, err := DetectProjectName()
		assert.NoError(t, err)
		assert.NotEmpty(t, projectName)
	})

	t.Run("falls back to directory name when no .tmporc and not in git repo", func(t *testing.T) {
		// Create a temporary directory
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Test
		projectName, err := DetectProjectName()
		assert.NoError(t, err)
		// Should use the directory name
		assert.NotEmpty(t, projectName)
		assert.Contains(t, tmpDir, projectName) // The project name should be part of the temp dir path
	})

	t.Run("empty project name in .tmporc falls back to detection", func(t *testing.T) {
		// Create a temporary directory with a .tmporc file that has empty project name
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create a .tmporc file with empty project name
		cfg := &config.Config{
			ProjectName: "",
			HourlyRate:  50.0,
		}
		err = cfg.Save(filepath.Join(tmpDir, ".tmporc"))
		require.NoError(t, err)

		// Test
		projectName, err := DetectProjectName()
		assert.NoError(t, err)
		assert.NotEmpty(t, projectName)
	})
}
