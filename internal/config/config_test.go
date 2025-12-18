package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("loads valid config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "test.tmporc")
		content := `project_name: test-project
hourly_rate: 100.5
description: "Test description"
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		assert.NoError(t, err)

		cfg, err := Load(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test-project", cfg.ProjectName)
		assert.Equal(t, 100.5, cfg.HourlyRate)
		assert.Equal(t, "Test description", cfg.Description)
	})

	t.Run("handles minimal config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "minimal.tmporc")
		content := `project_name: minimal-project
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		assert.NoError(t, err)

		cfg, err := Load(configPath)
		assert.NoError(t, err)
		assert.Equal(t, "minimal-project", cfg.ProjectName)
		assert.Equal(t, float64(0), cfg.HourlyRate)
		assert.Empty(t, cfg.Description)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := Load(filepath.Join(tmpDir, "nonexistent.tmporc"))
		assert.Error(t, err)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "invalid.tmporc")
		content := `project_name: test
invalid yaml syntax: [ unclosed
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		assert.NoError(t, err)

		_, err = Load(configPath)
		assert.Error(t, err)
	})
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("saves config successfully", func(t *testing.T) {
		cfg := &Config{
			ProjectName: "save-test",
			HourlyRate:  75.0,
			Description: "Saved config",
		}

		configPath := filepath.Join(tmpDir, "saved.tmporc")
		err := cfg.Save(configPath)
		assert.NoError(t, err)

		// Verify file was created and can be loaded
		loaded, err := Load(configPath)
		assert.NoError(t, err)
		assert.Equal(t, cfg.ProjectName, loaded.ProjectName)
		assert.Equal(t, cfg.HourlyRate, loaded.HourlyRate)
		assert.Equal(t, cfg.Description, loaded.Description)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "overwrite.tmporc")

		// Create initial config
		cfg1 := &Config{ProjectName: "original"}
		err := cfg1.Save(configPath)
		assert.NoError(t, err)

		// Overwrite with new config
		cfg2 := &Config{ProjectName: "updated"}
		err = cfg2.Save(configPath)
		assert.NoError(t, err)

		// Verify updated content
		loaded, err := Load(configPath)
		assert.NoError(t, err)
		assert.Equal(t, "updated", loaded.ProjectName)
	})

	t.Run("omits empty optional fields", func(t *testing.T) {
		cfg := &Config{
			ProjectName: "minimal",
			HourlyRate:  0,
			Description: "",
		}

		configPath := filepath.Join(tmpDir, "minimal.tmporc")
		err := cfg.Save(configPath)
		assert.NoError(t, err)

		// Read the raw file content
		content, err := os.ReadFile(configPath)
		assert.NoError(t, err)

		// Should only contain project_name
		assert.Contains(t, string(content), "project_name:")
		// Should omit hourly_rate and description when they're zero/empty
		assert.NotContains(t, string(content), "hourly_rate:")
		assert.NotContains(t, string(content), "description:")
	})
}

func TestCreate(t *testing.T) {
	t.Run("creates config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		err = Create("new-project", 125.0)
		assert.NoError(t, err)

		// Verify file was created
		tmporc := filepath.Join(tmpDir, ".tmporc")
		_, err = os.Stat(tmporc)
		assert.NoError(t, err)

		// Verify content
		cfg, err := Load(tmporc)
		assert.NoError(t, err)
		assert.Equal(t, "new-project", cfg.ProjectName)
		assert.Equal(t, 125.0, cfg.HourlyRate)
	})

	t.Run("returns error if file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		// Create initial file
		err = Create("first", 100.0)
		assert.NoError(t, err)

		// Try to create again
		err = Create("second", 200.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestCreateWithTemplate(t *testing.T) {
	t.Run("creates templated config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		err = CreateWithTemplate("templated-project", 99.99, "Test description")
		assert.NoError(t, err)

		// Verify file was created
		tmporc := filepath.Join(tmpDir, ".tmporc")
		content, err := os.ReadFile(tmporc)
		assert.NoError(t, err)

		// Verify template includes comments and formatting
		assert.Contains(t, string(content), "# tmpo project configuration")
		assert.Contains(t, string(content), "project_name: templated-project")
		assert.Contains(t, string(content), "hourly_rate: 99.99")
		assert.Contains(t, string(content), "description: \"Test description\"")
		assert.Contains(t, string(content), "# [OPTIONAL]")

		// Verify it can be loaded
		cfg, err := Load(tmporc)
		assert.NoError(t, err)
		assert.Equal(t, "templated-project", cfg.ProjectName)
		assert.Equal(t, 99.99, cfg.HourlyRate)
		assert.Equal(t, "Test description", cfg.Description)
	})

	t.Run("returns error if file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		// Create initial file
		err = CreateWithTemplate("first", 100.0, "desc")
		assert.NoError(t, err)

		// Try to create again
		err = CreateWithTemplate("second", 200.0, "desc2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestFindAndLoad(t *testing.T) {
	t.Run("finds config in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		configPath := filepath.Join(tmpDir, ".tmporc")
		cfg := &Config{ProjectName: "current-dir"}
		err = cfg.Save(configPath)
		assert.NoError(t, err)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		found, path, err := FindAndLoad()
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "current-dir", found.ProjectName)

		// Resolve both paths to handle symlinks (e.g., /var -> /private/var on macOS)
		expectedPath, _ := filepath.EvalSymlinks(configPath)
		actualPath, _ := filepath.EvalSymlinks(path)
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("finds config in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		configPath := filepath.Join(tmpDir, ".tmporc")
		cfg := &Config{ProjectName: "parent-dir"}
		err = cfg.Save(configPath)
		assert.NoError(t, err)

		// Create and change to subdirectory
		subDir := filepath.Join(tmpDir, "subdir", "nested")
		err = os.MkdirAll(subDir, 0755)
		assert.NoError(t, err)
		err = os.Chdir(subDir)
		assert.NoError(t, err)

		found, path, err := FindAndLoad()
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "parent-dir", found.ProjectName)

		// Resolve both paths to handle symlinks
		expectedPath, _ := filepath.EvalSymlinks(configPath)
		actualPath, _ := filepath.EvalSymlinks(path)
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		found, path, err := FindAndLoad()
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Empty(t, path)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("uses nearest config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create config in root
		rootConfig := filepath.Join(tmpDir, ".tmporc")
		cfg := &Config{ProjectName: "root"}
		err = cfg.Save(rootConfig)
		assert.NoError(t, err)

		// Create config in subdirectory
		subDir := filepath.Join(tmpDir, "project")
		err = os.MkdirAll(subDir, 0755)
		assert.NoError(t, err)
		subConfig := filepath.Join(subDir, ".tmporc")
		cfg = &Config{ProjectName: "project"}
		err = cfg.Save(subConfig)
		assert.NoError(t, err)

		// Change to subdirectory
		err = os.Chdir(subDir)
		assert.NoError(t, err)

		// Should find the nearest one (in current dir)
		found, path, err := FindAndLoad()
		assert.NoError(t, err)
		assert.Equal(t, "project", found.ProjectName)

		// Resolve both paths to handle symlinks
		expectedPath, _ := filepath.EvalSymlinks(subConfig)
		actualPath, _ := filepath.EvalSymlinks(path)
		assert.Equal(t, expectedPath, actualPath)
	})
}
