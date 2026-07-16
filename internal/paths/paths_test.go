package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTmpoDir(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)

	t.Run("uses .tmpo by default", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "")

		dir, err := TmpoDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".tmpo"), dir)
	})

	t.Run("uses .tmpo-dev when TMPO_DEV=1", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "1")

		dir, err := TmpoDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".tmpo-dev"), dir)
	})

	t.Run("uses .tmpo-dev when TMPO_DEV=true", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "true")

		dir, err := TmpoDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".tmpo-dev"), dir)
	})

	t.Run("uses .tmpo for other TMPO_DEV values", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "0")

		dir, err := TmpoDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".tmpo"), dir)
	})
}
