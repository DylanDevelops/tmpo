package settings

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DylanDevelops/tmpo/internal/fsperm"
	"github.com/stretchr/testify/assert"
)

func TestGlobalConfigSave_SetsPrivatePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not enforced on Windows")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	cfg := &GlobalConfig{Currency: "USD"}
	assert.NoError(t, cfg.Save())

	path, err := GetGlobalConfigPath()
	assert.NoError(t, err)

	fileInfo, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, fsperm.FilePerm, fileInfo.Mode().Perm())

	dirInfo, err := os.Stat(filepath.Dir(path))
	assert.NoError(t, err)
	assert.Equal(t, fsperm.DirPerm, dirInfo.Mode().Perm())
}
