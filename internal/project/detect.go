package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DylanDevelops/tmpo/internal/config"
)

// DetectProject attempts to determine the project name using a prioritized strategy:
// 1. If a tmporc config file is found (via FindTmporc) and a non-empty path is returned,
//    the project name is the base name of the directory that contains that config file.
// 2. Otherwise, if the repository name can be determined from Git (via GetGitRepoName)
//    and it is non-empty, that name is used.
// 3. If neither of the above produce a name, the base name of the current working
//    directory is returned.
//
// The function returns the detected project name. An error is returned only if the
// current working directory cannot be obtained (os.Getwd failure). Errors from
// FindTmporc and GetGitRepoName are not propagated; they are treated as an absence
// of a discovered name and the detection falls through to the next strategy.
func DetectProject() (string, error) {
	configPath, err := FindTmporc()
	if err == nil && configPath != "" {
		dir := filepath.Dir(configPath)

		return filepath.Base(dir), nil
	}

	gitName, err := GetGitRepoName()
	if err == nil && gitName != "" {
		return gitName, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return filepath.Base(cwd), nil
}


// DetectConfiguredProject returns the project name specified in the repository
// configuration, if present, otherwise it falls back to auto-detection.
// It loads configuration via config.FindAndLoad(); if a non-empty cfg.ProjectName
// is found that value is returned with a nil error. If no configured project
// name exists or loading fails, DetectProject() is invoked and its result is
// returned (project name, error).
func DetectConfiguredProject() (string, error) {
	if cfg, _, err := config.FindAndLoad(); err == nil && cfg != nil {
		if cfg.ProjectName != "" {
			return cfg.ProjectName, nil
		}
	}

	return DetectProject()
}

// FindTmporc searches the current working directory and each parent directory
// up to the filesystem root for a file named ".tmporc". If found, it returns
// the path to that file and a nil error. If no ".tmporc" file is found, it
// returns an empty string and a nil error. If obtaining the current working
// directory fails, the underlying error is returned.
func FindTmporc() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		tmporc := filepath.Join(dir, ".tmporc")
		if _, err := os.Stat(tmporc); err == nil {
			return tmporc, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", nil
}

// GetGitRepoName returns the name of the current Git repository.
//
// It runs "git rev-parse --show-toplevel" to determine the repository root
// directory, trims any surrounding whitespace from the command output, and
// returns the base name of that directory.
//
// The function uses the current working directory to locate the repository.
// If the git command fails (for example, if "git" is not available, the
// working directory is not inside a Git repository, or the command returns
// another error), it returns an empty string and the underlying error.
func GetGitRepoName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	gitRoot := strings.TrimSpace(string(output))

	return filepath.Base(gitRoot), nil
}

// IsInGitRepo reports whether the current working directory is inside a Git repository.
// It runs "git rev-parse --git-dir" and returns true if the command exits successfully.
// If git is not available in PATH, the command fails, or the directory is not a Git
// repository, the function returns false. This function spawns an external process and
// does not modify the current working directory.
func IsInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()

	return err == nil
}

// GetGitRoot returns the absolute path to the top-level directory of the
// current Git repository by invoking "git rev-parse --show-toplevel".
// It trims any trailing newline or whitespace from the command output and
// returns that path as a string. If the working directory is not inside a
// Git repository or the git command fails, an error is returned.
// Note: this function expects the "git" executable to be available in PATH.
func GetGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}

	return strings.TrimSpace(string(output)), nil
}
