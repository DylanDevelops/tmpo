package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	acceptDefaults bool
	globalProject  bool
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a project configuration",
		Long:  `Create a project configuration using an interactive form. By default, creates a .tmporc file in the current directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			// accept all is incompatible with initialization of global project
			if acceptDefaults && globalProject {
				ui.PrintError(ui.EmojiError, "Cannot use --accept-defaults with --global. Global projects require an explicit project configuration.")
				return ui.ErrHandled
			}

			var err error
			if globalProject {
				err = initGlobalProject()
			} else {
				err = initLocalProject()
			}
			if err != nil {
				return err
			}

			ui.NewlineBelow()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&acceptDefaults, "accept-defaults", "a", false, "Accept all defaults and skip interactive prompts")
	cmd.Flags().BoolVarP(&globalProject, "global", "g", false, "Create a global project that can be tracked from any directory")

	return cmd
}

func initLocalProject() error {
	if _, err := os.Stat(".tmporc"); err == nil {
		ui.PrintError(ui.EmojiError, ".tmporc already exists in this directory")
		return ui.ErrHandled
	}

	defaultName := detectDefaultProjectName()
	name, hourlyRate, description, exportPath, err := getProjectDetails(defaultName, "Initialize Project Configuration")
	if err != nil {
		return err
	}

	// create a .tmporc file
	err = settings.CreateWithTemplate(name, hourlyRate, description, exportPath)
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
		return err
	}

	fmt.Println()
	ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Created .tmporc for project %s", ui.Bold(name)))
	printProjectDetails(hourlyRate, description, exportPath)

	fmt.Println()
	ui.PrintMuted(0, "You can edit .tmporc to customize your project settings.")
	ui.PrintMuted(0, "Use 'tmpo config' to set global preferences like currency and time formats.")

	return nil
}

func initGlobalProject() error {
	registry, err := settings.LoadProjects()
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("failed to load projects registry: %v", err))
		return err
	}

	// global projects require project name type in
	name, hourlyRate, description, exportPath, err := getProjectDetails("", "Initialize Global Project")
	if err != nil {
		return err
	}

	if registry.Exists(name) {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("global project '%s' already exists", name))
		return ui.ErrHandled
	}

	// create the project
	var hourlyRatePtr *float64
	if hourlyRate > 0 {
		hourlyRatePtr = &hourlyRate
	}

	newProject := settings.GlobalProject{
		Name:        name,
		HourlyRate:  hourlyRatePtr,
		Description: description,
		ExportPath:  exportPath,
	}

	err = registry.AddProject(newProject)
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("failed to add project: %v", err))
		return err
	}

	err = registry.Save()
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("failed to save projects registry: %v", err))
		return err
	}

	fmt.Println()
	ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Created global project %s", ui.Bold(name)))
	printProjectDetails(hourlyRate, description, exportPath)

	fmt.Println()
	ui.PrintMuted(0, "You can now track time for this project from any directory:")
	ui.PrintMuted(0, fmt.Sprintf("  tmpo start --project \"%s\"", name))
	ui.PrintMuted(0, "")
	ui.PrintMuted(0, "Use 'tmpo config' to set global preferences like currency and time formats.")

	return nil
}

func getProjectDetails(defaultName, title string) (name string, hourlyRate float64, description, exportPath string, err error) {
	if acceptDefaults {
		name = defaultName
		hourlyRate = 0
		description = ""
		exportPath = ""
		return
	}

	ui.PrintSuccess(ui.EmojiInit, title)
	fmt.Println()

	// project name prompt
	var namePrompt promptui.Prompt
	if defaultName != "" {
		// local project
		namePrompt = promptui.Prompt{
			Label:     fmt.Sprintf("Project name (%s)", defaultName),
			AllowEdit: true,
		}
	} else {
		// global project
		namePrompt = promptui.Prompt{
			Label: "Project name",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("project name is required")
				}
				return nil
			},
		}
	}

	nameInput, err := namePrompt.Run()
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
		return
	}

	name = strings.TrimSpace(nameInput)
	if name == "" && defaultName != "" {
		name = defaultName
	}

	// hourly Rate prompt
	ratePrompt := promptui.Prompt{
		Label:    "Hourly rate (press Enter to skip)",
		Validate: validateHourlyRate,
	}

	rateInput, err := ratePrompt.Run()
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
		return
	}

	rateInput = strings.TrimSpace(rateInput)
	if rateInput != "" {
		hourlyRate, err = strconv.ParseFloat(rateInput, 64)
		if err != nil {
			ui.PrintError(ui.EmojiError, fmt.Sprintf("parsing hourly rate: %v", err))
			return
		}
	}

	// description prompt
	descPrompt := promptui.Prompt{
		Label: "Description (press Enter to skip)",
	}

	descInput, err := descPrompt.Run()
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
		return
	}

	description = strings.TrimSpace(descInput)

	// export path prompt
	exportPathPrompt := promptui.Prompt{
		Label: "Export path (press Enter to skip)",
	}

	exportPathInput, err := exportPathPrompt.Run()
	if err != nil {
		ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
		return
	}

	exportPath = strings.TrimSpace(exportPathInput)

	return
}

func printProjectDetails(hourlyRate float64, description, exportPath string) {
	if hourlyRate > 0 {
		ui.PrintInfo(4, ui.Bold("Hourly Rate"), fmt.Sprintf("%.2f", hourlyRate))
	}

	if description != "" {
		ui.PrintInfo(4, ui.Bold("Description"), description)
	}

	if exportPath != "" {
		ui.PrintInfo(4, ui.Bold("Export path"), exportPath)
	}
}

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

func validateHourlyRate(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil // optional field
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

func validateCurrency(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil // empty for default
	}

	// check formatting for currency codes
	if len(input) != 3 {
		return fmt.Errorf("currency code must be 3 letters (e.g., USD, EUR, GBP)")
	}

	for _, char := range input {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			return fmt.Errorf("currency code must contain only letters")
		}
	}

	return nil
}
