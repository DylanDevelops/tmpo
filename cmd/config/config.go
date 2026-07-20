package config

import (
	"fmt"
	"strings"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"settings", "preferences"},
		Short:   "Configure global tmpo settings",
		Long:    `Set up global configuration for tmpo including currency, date/time format, and timezone.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.NewlineAbove()

			// Load current global config
			currentConfig, err := settings.LoadGlobalConfig()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Failed to load config: %v", err))
				return err
			}

			// Display header
			ui.PrintSuccess(ui.EmojiInit, "Global tmpo Configuration")
			fmt.Println()

			// Show current settings
			fmt.Println(ui.Bold("Current settings:"))
			fmt.Printf("  Currency:    %s\n", ui.Muted(currentConfig.Currency))

			dateFormatDisplay := "(default)"
			if currentConfig.DateFormat != "" {
				dateFormatDisplay = currentConfig.DateFormat
			}
			fmt.Printf("  Date format: %s\n", ui.Muted(dateFormatDisplay))

			timeFormatDisplay := "(default)"
			if currentConfig.TimeFormat != "" {
				timeFormatDisplay = currentConfig.TimeFormat
			}
			fmt.Printf("  Time format: %s\n", ui.Muted(timeFormatDisplay))

			timezoneDisplay := "(local)"
			if currentConfig.Timezone != "" {
				timezoneDisplay = currentConfig.Timezone
			}
			fmt.Printf("  Timezone:    %s\n", ui.Muted(timezoneDisplay))

			exportPathDisplay := "(current directory)"
			if currentConfig.ExportPath != "" {
				exportPathDisplay = currentConfig.ExportPath
			}
			fmt.Printf("  Export path: %s\n", ui.Muted(exportPathDisplay))
			fmt.Println()

			// Currency prompt
			currencyPrompt := promptui.Prompt{
				Label:    fmt.Sprintf("Currency code (press Enter for %s)", currentConfig.Currency),
				Validate: validateCurrency,
			}

			currencyInput, err := currencyPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			currencyCode := strings.ToUpper(strings.TrimSpace(currencyInput))
			if currencyCode == "" {
				currencyCode = currentConfig.Currency
			}

			// Date format selection
			fmt.Println()
			dateFormatOptions := []string{"Keep current", "MM/DD/YYYY", "DD/MM/YYYY", "YYYY-MM-DD"}
			dateFormatSelect := promptui.Select{
				Label: "Select date format",
				Items: dateFormatOptions,
			}

			_, dateFormat, err := dateFormatSelect.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			// Keep current format if selected
			if dateFormat == "Keep current" {
				dateFormat = currentConfig.DateFormat
			}

			// Time format selection
			fmt.Println()
			timeFormatOptions := []string{"Keep current", "24-hour", "12-hour (AM/PM)"}
			timeFormatSelect := promptui.Select{
				Label: "Select time format",
				Items: timeFormatOptions,
			}

			_, timeFormat, err := timeFormatSelect.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			// Keep current format if selected
			if timeFormat == "Keep current" {
				timeFormat = currentConfig.TimeFormat
			}

			// Timezone prompt with validation
			fmt.Println()
			fmt.Println(ui.Muted("IANA timezone (e.g., America/New_York, Europe/London, Asia/Tokyo, UTC)"))
			fmt.Println(ui.Muted("Full list: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones"))
			timezonePrompt := promptui.Prompt{
				Label:    "Timezone (press Enter for local)",
				Validate: validateTimezone,
			}

			timezoneInput, err := timezonePrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			timezone := strings.TrimSpace(timezoneInput)
			if timezone == "" {
				timezone = currentConfig.Timezone
			}

			// Export path prompt
			fmt.Println()
			fmt.Println(ui.Muted("Default export directory for time entries (supports ~ for home directory)"))
			fmt.Println(ui.Muted("Type 'clear' to remove the export path setting"))
			exportPathPrompt := promptui.Prompt{
				Label: "Export path (press Enter to keep current)",
			}

			exportPathInput, err := exportPathPrompt.Run()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				return err
			}

			exportPath := strings.TrimSpace(exportPathInput)
			// Check for special clear keywords
			if strings.ToLower(exportPath) == "clear" || strings.ToLower(exportPath) == "none" {
				exportPath = ""
			} else if exportPath == "" {
				exportPath = currentConfig.ExportPath
			}

			// Create new config with updated values
			newConfig := &settings.GlobalConfig{
				Currency:   currencyCode,
				DateFormat: dateFormat,
				TimeFormat: timeFormat,
				Timezone:   timezone,
				ExportPath: exportPath,
			}

			// Save the config
			if err := newConfig.Save(); err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Failed to save config: %v", err))
				return err
			}

			// Display success message
			configPath, _ := settings.GetGlobalConfigPath()
			fmt.Println()
			ui.PrintSuccess(ui.EmojiSuccess, fmt.Sprintf("Configuration saved to %s", ui.Muted(configPath)))
			ui.PrintInfo(4, ui.Bold("Currency"), currencyCode)

			if dateFormat != "" {
				ui.PrintInfo(4, ui.Bold("Date format"), dateFormat)
			}

			if timeFormat != "" {
				ui.PrintInfo(4, ui.Bold("Time format"), timeFormat)
			}

			if timezone != "" {
				ui.PrintInfo(4, ui.Bold("Timezone"), timezone)
			}

			exportPathDisplay = "(current directory)"
			if exportPath != "" {
				exportPathDisplay = exportPath
			}
			ui.PrintInfo(4, ui.Bold("Export path"), exportPathDisplay)

			ui.NewlineBelow()

			return nil
		},
	}

	return cmd
}

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

func validateTimezone(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil // Allow empty for local timezone
	}

	// Basic validation: should contain at least one slash (e.g., America/New_York)
	// or be a common abbreviation like UTC, GMT
	commonTimezones := []string{"UTC", "GMT", "EST", "PST", "MST", "CST"}

	// Check if it's a common abbreviation
	upperInput := strings.ToUpper(input)
	for _, tz := range commonTimezones {
		if upperInput == tz {
			return nil
		}
	}

	// Check for IANA format (Region/City)
	if !strings.Contains(input, "/") {
		return fmt.Errorf("timezone should be in format Region/City (e.g., America/New_York) or UTC")
	}

	// Check that it doesn't have spaces
	if strings.Contains(input, " ") {
		return fmt.Errorf("timezone should not contain spaces (use underscores instead)")
	}

	return nil
}
