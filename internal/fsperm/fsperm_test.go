package fsperm

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not enforced on Windows")
	}

	t.Run("creates a new directory with private permissions", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "new")

		assert.NoError(t, SecureDir(dir))

		info, err := os.Stat(dir)
		assert.NoError(t, err)
		assert.Equal(t, DirPerm, info.Mode().Perm())
	})

	t.Run("tightens an existing loose directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "loose")
		assert.NoError(t, os.MkdirAll(dir, 0755))

		assert.NoError(t, SecureDir(dir))

		info, err := os.Stat(dir)
		assert.NoError(t, err)
		assert.Equal(t, DirPerm, info.Mode().Perm())
	})
}

func TestSecureFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not enforced on Windows")
	}

	t.Run("tightens an existing loose file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "data")
		assert.NoError(t, os.WriteFile(path, []byte("secret"), 0644))

		assert.NoError(t, SecureFile(path))

		info, err := os.Stat(path)
		assert.NoError(t, err)
		assert.Equal(t, FilePerm, info.Mode().Perm())
	})

	t.Run("is a no-op when the file does not exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing")

		assert.NoError(t, SecureFile(path))
	})
}
