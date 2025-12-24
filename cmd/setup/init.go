package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DylanDevelops/tmpo/internal/config"
	"github.com/DylanDevelops/tmpo/internal/currency"
	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	acceptDefaults bool
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .tmporc config file",
		Long:  `Create a .tmporc configuration file in the current directory using an interactive form.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			if _, err := os.Stat(".tmporc"); err == nil {
				ui.PrintError(ui.EmojiError, ".tmporc already exists in this directory")
				os.Exit(1)
			}

			// Detect default project name
			defaultName := detectDefaultProjectName()

			var name string
			var hourlyRate float64
			var description string
			var currencyCode string

			if acceptDefaults {
				// Use all defaults without prompting
				name = defaultName
				hourlyRate = 0
				description = ""
				currencyCode = "USD"
			} else {
				// Interactive form
				ui.PrintSuccess(ui.EmojiInit, "Initialize Project Configuration")
				fmt.Println()

				// Project Name prompt
				namePrompt := promptui.Prompt{
					Label:     fmt.Sprintf("Project name (%s)", defaultName),
					AllowEdit: true,
				}

				nameInput, err := namePrompt.Run()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				name = strings.TrimSpace(nameInput)
				if name == "" {
					name = defaultName
				}

				// Hourly Rate prompt
				ratePrompt := promptui.Prompt{
					Label:    "Hourly rate (press Enter to skip)",
					Validate: validateHourlyRate,
				}

				rateInput, err := ratePrompt.Run()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				rateInput = strings.TrimSpace(rateInput)
				if rateInput != "" {
					hourlyRate, err = strconv.ParseFloat(rateInput, 64)
					if err != nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing hourly rate: %v", err))
						os.Exit(1)
					}
				}

				// Description prompt
				descPrompt := promptui.Prompt{
					Label: "Description (press Enter to skip)",
				}

				descInput, err := descPrompt.Run()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				description = strings.TrimSpace(descInput)

				// Currency prompt
				currencyPrompt := promptui.Prompt{
					Label:    "Currency code (press Enter for USD)",
					Validate: validateCurrency,
				}

				currencyInput, err := currencyPrompt.Run()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}

				currencyCode = strings.ToUpper(strings.TrimSpace(currencyInput))
				if currencyCode == "" {
					currencyCode = "USD"
				}
			}

			// Create the .tmporc file
			err := config.CreateWithTemplate(name, hourlyRate, description, currencyCode)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Created .tmporc for project %s", ui.Bold(name)))
			if hourlyRate > 0 {
				ui.PrintInfo(4, ui.Bold("Hourly Rate"), currency.FormatCurrency(hourlyRate, currencyCode))
			}
			if description != "" {
				ui.PrintInfo(4, ui.Bold("Description"), description)
			}
			if currencyCode != "" && currencyCode != "USD" {
				ui.PrintInfo(4, ui.Bold("Currency"), currencyCode)
			}

			fmt.Println()
			ui.PrintMuted(0, "You can edit .tmporc to customize your project settings.")

			ui.NewlineBelow()
		},
	}

	cmd.Flags().BoolVarP(&acceptDefaults, "accept-defaults", "a", false, "Accept all defaults and skip interactive prompts")

	return cmd
}

// detectDefaultProjectName returns the auto-detected project name
func detectDefaultProjectName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "my-project"
	}

	name := ""
	if project.IsInGitRepo() {
		gitName, _ := project.GetGitRoot()
		if gitName != "" {
			name = filepath.Base(gitName)
		}
	}

	if name == "" {
		name = filepath.Base(cwd)
	}

	return name
}

// validateHourlyRate validates that the input is empty or a valid positive number
func validateHourlyRate(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil // Allow empty for optional field
	}

	rate, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return fmt.Errorf("must be a valid number")
	}

	if rate < 0 {
		return fmt.Errorf("hourly rate cannot be negative")
	}

	return nil
}

// validateCurrency validates that the input is empty or a valid currency code format
func validateCurrency(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil // Allow empty for default
	}

	// Currency codes should be 3 letters
	if len(input) != 3 {
		return fmt.Errorf("currency code must be 3 letters (e.g., USD, EUR, GBP)")
	}

	// Check that it's all letters
	for _, char := range input {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			return fmt.Errorf("currency code must contain only letters")
		}
	}

	return nil
}
