package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/spf13/cobra"
)

var (
	hourlyRate float64
	projectName string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a .tmporc config file",
	Long:  `Create a .tmporc configuration file in the current directory.`,
	Run: func(cmd* cobra.Command, args []string) {
		if _, err := os.Stat(".tmporc"); err == nil {
			fmt.Println("Error: .tmporc already exists in this directory")

			os.Exit(1)
		}

		name := projectName
		if name == "" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)

				os.Exit(1)
			}

			if project.IsInGitRepo() {
				gitName, _ := project.GetGitRoot()
				if gitName != "" {
					name = filepath.Base(gitName)
				}
			}

			if name == "" {
				name = filepath.Base(cwd)
			}
		}

		err := config.Create(name, hourlyRate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		fmt.Printf("[tmpo] Created .tmporc for project '%s'\n", name)
		if hourlyRate > 0 {
			fmt.Printf("	Hourly Rate: $%.2f\n", hourlyRate)
		}

		fmt.Println("\nYou can edit .tmporc to customize your project settings.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Float64VarP(&hourlyRate, "rate", "r", 0, "Hourly rate for this project")
	initCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name (defaults to directory/repo name)")
}
