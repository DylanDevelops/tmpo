package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DylanDevelops/tmpo/internal/currency"
	"go.yaml.in/yaml/v3"
)

// Config represents the project's configuration as loaded from a YAML file.
// It contains identifying and billing information along with an optional description.
//
// ProjectName is the human-readable name of the project.
// HourlyRate is the billable hourly rate for the project; when zero it will be omitted from YAML.
// Description is an optional free-form description of the project; when empty it will be omitted from YAML.
// Currency is the ISO 4217 currency code (e.g., USD, EUR, GBP) for billing display; when empty it will be omitted from YAML.
//
// ! IMPORTANT When adding new fields to this struct, also update configTemplate below. !
type Config struct {
	ProjectName string `yaml:"project_name"`
	HourlyRate float64 `yaml:"hourly_rate,omitempty"`
	Description string `yaml:"description,omitempty"`
	Currency    string `yaml:"currency,omitempty"`
}

// configTemplate is the template used when creating new .tmporc files via CreateWithTemplate.
// It includes all available configuration options with helpful comments.
//
// Format placeholders:
//   %s - project name (string)
//   %.2f - hourly rate (float64, 2 decimal places)
//   %s - description (string)
//   %s - currency code (string)
//
// ! IMPORTANT: When adding new fields to the Config struct above, update this template. !
const configTemplate = `# tmpo project configuration
# This file configures time tracking settings for this project

# Project name (used to identify time entries)
project_name: %s

# [OPTIONAL] Hourly rate for billing calculations (set to 0 to disable)
hourly_rate: %.2f

# [OPTIONAL] Description for this project
description: "%s"

# [OPTIONAL] Currency code for billing display (USD, EUR, GBP, JPY, etc.)
currency: %s
`

// Load reads a YAML configuration file from the provided path and unmarshals it into a Config.
// It returns a pointer to the populated Config on success. If the file cannot be read or the
// contents cannot be parsed as YAML, Load returns a wrapped error describing the failure.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// Save marshals the Config into YAML and writes it to the provided filesystem path.
// The configuration is encoded using yaml.Marshal and written with file mode 0644.
// If a file already exists at path it will be overwritten. An error is returned
// if marshaling fails or if writing the file to disk is unsuccessful.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Create creates a new Config populated with projectName and hourlyRate and writes it
// to a ".tmporc" file in the current working directory ("./.tmporc").
// If a ".tmporc" file already exists, Create returns an error and does not overwrite it.
// Any error encountered while saving the configuration is returned to the caller.
func Create(projectName string, hourlyRate float64) error {
	config := &Config{
		ProjectName: projectName,
		HourlyRate: hourlyRate,
	}

	tmporc := filepath.Join(".", ".tmporc")
	if _, err := os.Stat(tmporc); err == nil {
		return fmt.Errorf(".tmporc already exists")
	}

	return config.Save(tmporc)
}

// CreateWithTemplate creates a new .tmporc file with a user-friendly format that includes
// all fields (even if empty) and helpful comments. This provides a better user experience
// by showing all available configuration options.
func CreateWithTemplate(projectName string, hourlyRate float64, description string, currency string) error {
	tmporc := filepath.Join(".", ".tmporc")
	if _, err := os.Stat(tmporc); err == nil {
		return fmt.Errorf(".tmporc already exists")
	}

	content := fmt.Sprintf(configTemplate, projectName, hourlyRate, description, currency)

	if err := os.WriteFile(tmporc, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// FindAndLoad searches upward from the current working directory for a file named
// ".tmporc". Starting at os.Getwd(), it ascends parent directories until it either
// finds the file or reaches the filesystem root.
//
// If a ".tmporc" file is found, FindAndLoad calls Load(path) and returns the resulting
// *Config, the absolute path to the discovered file, and any error returned by Load.
// If Load returns an error the returned *Config may be nil.
//
// If the current working directory cannot be determined, or if no ".tmporc" is found
// before reaching the root, FindAndLoad returns (nil, "", err) describing the failure.
func FindAndLoad() (*Config, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	for {
		tmporc := filepath.Join(dir, ".tmporc")
		if _, err := os.Stat(tmporc); err == nil {
			config, err := Load(tmporc)

			return config, tmporc, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return nil, "", fmt.Errorf(".tmporc not found")
}

// GetCurrencyOrDefault returns the configured currency code, or "USD" if not set.
// This provides a convenient way to get the currency with a sensible default.
func (c *Config) GetCurrencyOrDefault() string {
	if c.Currency == "" {
		return currency.DefaultCurrency
	}
	return c.Currency
}
