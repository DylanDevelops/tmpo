package fsperm

import (
	"fmt"
	"os"
)

const (
	DirPerm  os.FileMode = 0700
	FilePerm os.FileMode = 0600
)

// SecureDir creates dir (and any parents) if needed and sets it to owner-only
// access (DirPerm, 0700). The Chmod also applies to a pre-existing directory,
// remediating installs created with looser permissions.
func SecureDir(dir string) error {
	if err := os.MkdirAll(dir, DirPerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	if err := os.Chmod(dir, DirPerm); err != nil {
		return fmt.Errorf("failed to secure directory %s: %w", dir, err)
	}
	return nil
}

// SecureFile sets an existing file to owner-only read/write (FilePerm, 0600).
// It is a no-op if the file does not exist yet, so callers may invoke it
// defensively on optional files.
func SecureFile(path string) error {
	if err := os.Chmod(path, FilePerm); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to secure file %s: %w", path, err)
	}
	return nil
}
