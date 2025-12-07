package cmd

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [description]",
	Short: "Start tracking time",
	Long:  `Start a new time tracking session for the current project.`,
	Run: func(cmd* cobra.Command, args []string) {
		db, err := storage.Initialize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		defer db.Close()

		running, err := db.GetRunningEntry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		if running != nil {
			fmt.Fprintf(os.Stderr, "Error: Already tracking time for `%s\n", running.ProjectName)
			fmt.Println("Use 'tmpo stop' to stop the current session first.")

			os.Exit(1)
		}

		projectName, err := DetectProjectName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting project: %v\n", err)
			
			os.Exit(1)
		}
				
		description := ""
		if len(args) > 0 {
			description = args[0]
		}

		entry, err := db.CreateEntry(projectName, description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		fmt.Printf("[tmpo] Started tracking time for '%s'\n", entry.ProjectName)

		if cfg, _, err := config.FindAndLoad(); err == nil && cfg != nil {
			fmt.Println("    Source: .tmporc")
		} else if project.IsInGitRepo() {
			fmt.Println("    Source: git repository")
		} else {
			fmt.Println("    Source: directory name")
		}

		if description != "" {
			fmt.Printf("    Description: %s\n", description)
		}
	},
}

// DetectProjectName returns the name of the current project.
// It first attempts to load a configuration via config.FindAndLoad; if a configuration
// is found and its ProjectName field is non-empty, that value is returned.
// If no configuration or project name is available, DetectProjectName falls back to
// project.DetectProject() to determine the project name from the repository or environment.
// The function returns the determined project name and any error encountered during
// configuration loading or fallback detection.
func DetectProjectName() (string, error) {
	if cfg, _, err := config.FindAndLoad(); err == nil && cfg != nil {
		if cfg.ProjectName != "" {
			return cfg.ProjectName, nil
		}
	}

	return project.DetectProject()
}

func init() {
	rootCmd.AddCommand(startCmd)
}
