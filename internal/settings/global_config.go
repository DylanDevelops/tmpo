package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DylanDevelops/tmpo/internal/currency"
	"go.yaml.in/yaml/v3"
)

// GlobalConfig represents the user's global configuration as loaded from a YAML file.
// It contains user-wide settings that apply across all projects.
//
// Currency is the ISO 4217 currency code (e.g., USD, EUR, GBP) for billing display.
// DateFormat is the preferred date format (e.g., MM/DD/YYYY, DD/MM/YYYY, YYYY-MM-DD).
// TimeFormat is the preferred time format (e.g., 24-hour, 12-hour).
// Timezone is an optional IANA timezone name (e.g., America/New_York, UTC).
type GlobalConfig struct {
	Currency   string `yaml:"currency"`
	DateFormat string `yaml:"date_format,omitempty"`
	TimeFormat string `yaml:"time_format,omitempty"`
	Timezone   string `yaml:"timezone,omitempty"`
}

// DefaultGlobalConfig returns a GlobalConfig with sensible default values.
// Currency defaults to USD, while DateFormat, TimeFormat, and Timezone are empty
// (meaning the system will use defaults and local timezone respectively).
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Currency:   currency.DefaultCurrency,
		DateFormat: "",
		TimeFormat: "",
		Timezone:   "",
	}
}

// GetGlobalConfigPath returns the absolute path to the global configuration file.
// The config is stored at $HOME/.tmpo/config.yaml alongside the database.
// Returns an error if the home directory cannot be determined.
func GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".tmpo", "config.yaml"), nil
}

// LoadGlobalConfig loads the global configuration from ~/.tmpo/config.yaml.
// If the file doesn't exist, it returns a default configuration without error.
// If the file exists but cannot be read or parsed, it returns an error.
func LoadGlobalConfig() (*GlobalConfig, error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse global config at %s: %w (check file syntax)", configPath, err)
	}

	// Ensure currency has a default if empty
	if config.Currency == "" {
		config.Currency = currency.DefaultCurrency
	}

	return &config, nil
}

// Save writes the GlobalConfig to ~/.tmpo/config.yaml.
// It creates the ~/.tmpo directory if it doesn't exist.
// Returns an error if the directory cannot be created or the file cannot be written.
func (gc *GlobalConfig) Save() error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

	// Ensure ~/.tmpo directory exists
	tmpoDir := filepath.Dir(configPath)
	if err := os.MkdirAll(tmpoDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(gc)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	return nil
}

// FormatTime formats a time according to the user's global time format preference.
// Returns time in either 24-hour format (15:04) or 12-hour format (3:04 PM).
func FormatTime(t time.Time) string {
	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.TimeFormat == "" || cfg.TimeFormat == "Keep current" {
		// Default to 12-hour format
		return t.Format("3:04 PM")
	}

	if cfg.TimeFormat == "24-hour" {
		return t.Format("15:04")
	}

	// 12-hour (AM/PM)
	return t.Format("3:04 PM")
}

// FormatTimePadded formats a time with zero-padded hours according to the user's time format preference.
// Returns time in either 24-hour format (15:04) or 12-hour format (03:04 PM).
func FormatTimePadded(t time.Time) string {
	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.TimeFormat == "" || cfg.TimeFormat == "Keep current" {
		// Default to 12-hour format with padding
		return t.Format("03:04 PM")
	}

	if cfg.TimeFormat == "24-hour" {
		return t.Format("15:04")
	}

	// 12-hour (AM/PM) with padding
	return t.Format("03:04 PM")
}

// FormatDate formats a date according to the user's global date format preference.
// Returns date in MM/DD/YYYY, DD/MM/YYYY, or YYYY-MM-DD format based on config.
func FormatDate(t time.Time) string {
	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.DateFormat == "" || cfg.DateFormat == "Keep current" {
		// Default to MM/DD/YYYY
		return t.Format("01/02/2006")
	}

	switch cfg.DateFormat {
	case "MM/DD/YYYY":
		return t.Format("01/02/2006")
	case "DD/MM/YYYY":
		return t.Format("02/01/2006")
	case "YYYY-MM-DD":
		return t.Format("2006-01-02")
	default:
		return t.Format("01/02/2006")
	}
}

// FormatDateDashed formats a date with dashes according to the user's date format preference.
// Returns date in MM-DD-YYYY, DD-MM-YYYY, or YYYY-MM-DD format based on config.
func FormatDateDashed(t time.Time) string {
	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.DateFormat == "" || cfg.DateFormat == "Keep current" {
		// Default to MM-DD-YYYY
		return t.Format("01-02-2006")
	}

	switch cfg.DateFormat {
	case "MM/DD/YYYY":
		return t.Format("01-02-2006")
	case "DD/MM/YYYY":
		return t.Format("02-01-2006")
	case "YYYY-MM-DD":
		return t.Format("2006-01-02")
	default:
		return t.Format("01-02-2006")
	}
}

// FormatDateTime formats a date and time according to the user's global preferences.
// Returns combined date and time string (e.g., "01/02/2006 3:04 PM" or "2006-01-02 15:04").
func FormatDateTime(t time.Time) string {
	return FormatDate(t) + " " + FormatTime(t)
}

// FormatDateTimeDashed formats a date and time with dashes according to preferences.
// Returns combined date and time string (e.g., "01-02-2006 3:04 PM" or "2006-01-02 15:04").
func FormatDateTimeDashed(t time.Time) string {
	return FormatDateDashed(t) + " " + FormatTime(t)
}

// FormatDateLong formats a date in a long human-readable format.
// Returns date as "Mon, Jan 2, 2006" regardless of user preferences (for headers).
func FormatDateLong(t time.Time) string {
	return t.Format("Mon, Jan 2, 2006")
}

// FormatDateTimeLong formats a date and time in a long human-readable format.
// Returns "Jan 2, 2006 at 3:04 PM" or "Jan 2, 2006 at 15:04" based on time preference.
func FormatDateTimeLong(t time.Time) string {
	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.TimeFormat == "" || cfg.TimeFormat == "Keep current" {
		return t.Format("Jan 2, 2006 at 3:04 PM")
	}

	if cfg.TimeFormat == "24-hour" {
		return t.Format("Jan 2, 2006 at 15:04")
	}

	return t.Format("Jan 2, 2006 at 3:04 PM")
}
