package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DylanDevelops/tmpo/internal/currency"
	"go.yaml.in/yaml/v3"
)

type GlobalConfig struct {
	Currency   string `yaml:"currency"`
	DateFormat string `yaml:"date_format,omitempty"`
	TimeFormat string `yaml:"time_format,omitempty"`
	Timezone   string `yaml:"timezone,omitempty"`
	ExportPath string `yaml:"export_path,omitempty"`
}

func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Currency:   currency.DefaultCurrency,
		DateFormat: "",
		TimeFormat: "",
		Timezone:   "",
		ExportPath: "",
	}
}

// GetGlobalConfigPath returns the path to the global config file. If
// the TMPO_DEV environment variable is set to 1 or true, a
// developer-specific config path is returned.
func GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	tmpoDir := filepath.Join(home, ".tmpo")
	if devMode := os.Getenv("TMPO_DEV"); devMode == "1" || devMode == "true" {
		tmpoDir = filepath.Join(home, ".tmpo-dev")
	}

	return filepath.Join(tmpoDir, "config.yaml"), nil
}

// LoadGlobalConfig loads the global config from disk. If the config
// file does not exist the default config is returned.
func LoadGlobalConfig() (*GlobalConfig, error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

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

	if config.Currency == "" {
		config.Currency = currency.DefaultCurrency
	}

	return &config, nil
}

// Save writes the GlobalConfig to the config file, creating the
// config directory if necessary.
func (gc *GlobalConfig) Save() error {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}

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

// GetDisplayTimezone returns the user's configured timezone or the
// local timezone as a fallback.
func GetDisplayTimezone() *time.Location {
	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.Timezone == "" {
		return time.Local
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return time.Local
	}

	return loc
}

// ToDisplayTime converts the provided time to the user's display
// timezone.
func ToDisplayTime(t time.Time) time.Time {
	return t.In(GetDisplayTimezone())
}

// FormatTime formats t according to the user's configured time
// format. Falls back to 12-hour format "3:04 PM".
func FormatTime(t time.Time) string {
	t = ToDisplayTime(t)

	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.TimeFormat == "" || cfg.TimeFormat == "Keep current" {
		return t.Format("3:04 PM")
	}

	if cfg.TimeFormat == "24-hour" {
		return t.Format("15:04")
	}

	return t.Format("3:04 PM")
}

// FormatTimePadded formats t with a padded hour for 12-hour formats
// ("03:04 PM") or returns the 24-hour format when configured.
func FormatTimePadded(t time.Time) string {
	t = ToDisplayTime(t)

	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.TimeFormat == "" || cfg.TimeFormat == "Keep current" {
		return t.Format("03:04 PM")
	}

	if cfg.TimeFormat == "24-hour" {
		return t.Format("15:04")
	}

	return t.Format("03:04 PM")
}

// FormatDate formats t according to the user's configured date
// format. Defaults to "01/02/2006" (MM/DD/YYYY).
func FormatDate(t time.Time) string {
	t = ToDisplayTime(t)

	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.DateFormat == "" || cfg.DateFormat == "Keep current" {
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

// FormatDateDashed formats t using dashes between date components.
func FormatDateDashed(t time.Time) string {
	t = ToDisplayTime(t)

	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.DateFormat == "" || cfg.DateFormat == "Keep current" {
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

// FormatDateTime returns the date and time formatted according to
// user preferences.
func FormatDateTime(t time.Time) string {
	return FormatDate(t) + " " + FormatTime(t)
}

// FormatDateTimeDashed returns the dashed date and time formatted
// according to user preferences.
func FormatDateTimeDashed(t time.Time) string {
	return FormatDateDashed(t) + " " + FormatTime(t)
}

// FormatDateLong returns a long human-readable date string like
// "Mon, Jan 2, 2006".
func FormatDateLong(t time.Time) string {
	t = ToDisplayTime(t)

	return t.Format("Mon, Jan 2, 2006")
}

// FormatDateTimeLong returns a long human-readable date/time string
// honoring the user's time format preference.
func FormatDateTimeLong(t time.Time) string {
	t = ToDisplayTime(t)

	cfg, err := LoadGlobalConfig()
	if err != nil || cfg.TimeFormat == "" || cfg.TimeFormat == "Keep current" {
		return t.Format("Jan 2, 2006 at 3:04 PM")
	}

	if cfg.TimeFormat == "24-hour" {
		return t.Format("Jan 2, 2006 at 15:04")
	}

	return t.Format("Jan 2, 2006 at 3:04 PM")
}
