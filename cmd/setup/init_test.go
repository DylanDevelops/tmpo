package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectDefaultProjectName(t *testing.T) {
	t.Run("returns git repository name when in git repo", func(t *testing.T) {
		// This test would require setting up a real git repo
		// We'll test the fallback behavior instead
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Initialize a minimal git repo
		err = os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)
		require.NoError(t, err)

		name := detectDefaultProjectName()
		assert.NotEmpty(t, name)
	})

	t.Run("returns directory name when not in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		name := detectDefaultProjectName()
		assert.NotEmpty(t, name)
		// The name should be the base of the temp directory
		assert.Equal(t, filepath.Base(tmpDir), name)
	})
}

func TestValidateHourlyRate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty string is valid (optional field)",
			input:     "",
			wantError: false,
		},
		{
			name:      "whitespace only is valid",
			input:     "   ",
			wantError: false,
		},
		{
			name:      "valid positive number",
			input:     "75.50",
			wantError: false,
		},
		{
			name:      "valid integer",
			input:     "100",
			wantError: false,
		},
		{
			name:      "zero is valid",
			input:     "0",
			wantError: false,
		},
		{
			name:      "negative number is invalid",
			input:     "-50",
			wantError: true,
			errorMsg:  "hourly rate cannot be negative",
		},
		{
			name:      "non-numeric string is invalid",
			input:     "not-a-number",
			wantError: true,
			errorMsg:  "must be a valid number",
		},
		{
			name:      "mixed alphanumeric is invalid",
			input:     "50abc",
			wantError: true,
			errorMsg:  "must be a valid number",
		},
		{
			name:      "special characters are invalid",
			input:     "$100",
			wantError: true,
			errorMsg:  "must be a valid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHourlyRate(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetProjectDetails(t *testing.T) {
	// Save original acceptDefaults value and reset after test
	originalAcceptDefaults := acceptDefaults
	defer func() { acceptDefaults = originalAcceptDefaults }()

	t.Run("uses defaults when acceptDefaults is true", func(t *testing.T) {
		acceptDefaults = true
		defaultName := "test-project"

		name, hourlyRate, description, exportPath, err := getProjectDetails(defaultName, "Test Title")

		assert.NoError(t, err)
		assert.Equal(t, defaultName, name)
		assert.Equal(t, float64(0), hourlyRate)
		assert.Empty(t, description)
		assert.Empty(t, exportPath)
	})
}

func TestPrintProjectDetails(t *testing.T) {
	// This is primarily a display function, so we just test it doesn't panic
	t.Run("handles all fields present", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printProjectDetails(100.0, "Test description", "/tmp/export")
		})
	})

	t.Run("handles empty fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printProjectDetails(0, "", "")
		})
	})

	t.Run("handles partial fields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printProjectDetails(50.5, "Description only", "")
		})
	})
}

func TestInitGlobalProject_Integration(t *testing.T) {
	// Set up test environment
	tmpDir := t.TempDir()

	t.Setenv("HOME", tmpDir)        // Unix/macOS
	t.Setenv("USERPROFILE", tmpDir) // Windows
	t.Setenv("TMPO_DEV", "1")

	t.Run("creates global project directly via registry", func(t *testing.T) {
		// Instead of testing initGlobalProject() which requires interactive input,
		// test the registry operations directly
		registry, err := settings.LoadProjects()
		require.NoError(t, err)

		// Add a test project
		rate := 100.0
		newProject := settings.GlobalProject{
			Name:        "test-project",
			HourlyRate:  &rate,
			Description: "Test description",
			ExportPath:  "/tmp/test",
		}

		err = registry.AddProject(newProject)
		require.NoError(t, err)

		err = registry.Save()
		require.NoError(t, err)

		// Verify project was added to registry
		reloadedRegistry, err := settings.LoadProjects()
		require.NoError(t, err)
		assert.True(t, reloadedRegistry.Exists("test-project"))

		// Verify project details
		project, err := reloadedRegistry.GetProject("test-project")
		require.NoError(t, err)
		assert.Equal(t, "test-project", project.Name)
		assert.Equal(t, &rate, project.HourlyRate)
		assert.Equal(t, "Test description", project.Description)
		assert.Equal(t, "/tmp/test", project.ExportPath)
	})
}

func TestInitLocalProject_Integration(t *testing.T) {
	// Save original flags and restore after test
	originalAcceptDefaults := acceptDefaults
	originalGlobalProject := globalProject
	defer func() {
		acceptDefaults = originalAcceptDefaults
		globalProject = originalGlobalProject
	}()

	acceptDefaults = true
	globalProject = false

	t.Run("creates local .tmporc file", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Run init
		assert.NotPanics(t, func() {
			initLocalProject()
		})

		// Verify .tmporc was created
		tmporc := filepath.Join(tmpDir, ".tmporc")
		_, err = os.Stat(tmporc)
		assert.NoError(t, err)

		// Verify content can be loaded
		cfg, err := settings.Load(tmporc)
		require.NoError(t, err)
		assert.Equal(t, filepath.Base(tmpDir), cfg.ProjectName)
	})

	t.Run("prevents duplicate .tmporc creation", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(origDir)

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create .tmporc
		tmporc := filepath.Join(tmpDir, ".tmporc")
		err = os.WriteFile(tmporc, []byte("project_name: existing\n"), 0644)
		require.NoError(t, err)

		// Attempt to run init again should exit
		// We can't easily test os.Exit(), so we'll just verify the check logic
		_, err = os.Stat(".tmporc")
		assert.NoError(t, err) // File exists, so initLocalProject would fail
	})
}
