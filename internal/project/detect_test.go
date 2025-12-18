package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTmporc(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create nested directories
	projectDir := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(projectDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	assert.NoError(t, err)

	// Create .tmporc in project root
	tmporc := filepath.Join(projectDir, ".tmporc")
	err = os.WriteFile(tmporc, []byte("project_name: test\n"), 0644)
	assert.NoError(t, err)

	t.Run("finds tmporc in current directory", func(t *testing.T) {
		err := os.Chdir(projectDir)
		assert.NoError(t, err)

		path, err := FindTmporc()
		assert.NoError(t, err)

		// Resolve both paths to handle symlinks (e.g., /var -> /private/var on macOS)
		expectedPath, _ := filepath.EvalSymlinks(tmporc)
		actualPath, _ := filepath.EvalSymlinks(path)
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("finds tmporc in parent directory", func(t *testing.T) {
		err := os.Chdir(subDir)
		assert.NoError(t, err)

		path, err := FindTmporc()
		assert.NoError(t, err)

		// Resolve both paths to handle symlinks
		expectedPath, _ := filepath.EvalSymlinks(tmporc)
		actualPath, _ := filepath.EvalSymlinks(path)
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("returns empty string when not found", func(t *testing.T) {
		noConfigDir := filepath.Join(tmpDir, "no-config")
		err := os.MkdirAll(noConfigDir, 0755)
		assert.NoError(t, err)

		err = os.Chdir(noConfigDir)
		assert.NoError(t, err)

		path, err := FindTmporc()
		assert.NoError(t, err)
		assert.Empty(t, path)
	})
}

func TestGetGitRepoName(t *testing.T) {
	// This test depends on running in a git repository
	// It will work in the tmpo repository itself

	t.Run("returns repo name when in git repo", func(t *testing.T) {
		if !IsInGitRepo() {
			t.Skip("Not in a git repository")
		}

		name, err := GetGitRepoName()
		assert.NoError(t, err)
		assert.NotEmpty(t, name)
		// Should be "tmpo" when running in tmpo repo
		assert.Equal(t, "tmpo", name)
	})
}

func TestIsInGitRepo(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	t.Run("returns true when in git repo", func(t *testing.T) {
		// This test assumes we're running in the tmpo git repo
		result := IsInGitRepo()
		assert.True(t, result)
	})

	t.Run("returns false when not in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		result := IsInGitRepo()
		assert.False(t, result)
	})
}

func TestGetGitRoot(t *testing.T) {
	t.Run("returns git root when in git repo", func(t *testing.T) {
		if !IsInGitRepo() {
			t.Skip("Not in a git repository")
		}

		root, err := GetGitRoot()
		assert.NoError(t, err)
		assert.NotEmpty(t, root)
		// Should end with "tmpo"
		assert.Equal(t, "tmpo", filepath.Base(root))
	})

	t.Run("returns error when not in git repo", func(t *testing.T) {
		// Save original directory
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		assert.NoError(t, err)

		_, err = GetGitRoot()
		assert.Error(t, err)
	})
}

func TestDetectProject(t *testing.T) {
	t.Run("detects from tmporc", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		projectDir := filepath.Join(tmpDir, "my-cool-project")
		err = os.MkdirAll(projectDir, 0755)
		assert.NoError(t, err)

		// Create .tmporc
		tmporc := filepath.Join(projectDir, ".tmporc")
		err = os.WriteFile(tmporc, []byte("project_name: test\n"), 0644)
		assert.NoError(t, err)

		err = os.Chdir(projectDir)
		assert.NoError(t, err)

		name, err := DetectProject()
		assert.NoError(t, err)
		assert.Equal(t, "my-cool-project", name)
	})

	t.Run("detects from git when no tmporc", func(t *testing.T) {
		if !IsInGitRepo() {
			t.Skip("Not in a git repository")
		}

		// Change to git root (which shouldn't have .tmporc in this test)
		root, err := GetGitRoot()
		assert.NoError(t, err)

		err = os.Chdir(root)
		assert.NoError(t, err)

		name, err := DetectProject()
		assert.NoError(t, err)
		assert.NotEmpty(t, name)
	})

	t.Run("falls back to directory name", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer os.Chdir(originalDir)

		projectDir := filepath.Join(tmpDir, "fallback-project")
		err = os.MkdirAll(projectDir, 0755)
		assert.NoError(t, err)

		err = os.Chdir(projectDir)
		assert.NoError(t, err)

		name, err := DetectProject()
		assert.NoError(t, err)
		assert.Equal(t, "fallback-project", name)
	})
}
