package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// City represents a clock configuration for a city
type City struct {
	Name     string `yaml:"name"`
	Timezone string `yaml:"timezone"`
}

// Config represents the application configuration
type Config struct {
	Cities []City `yaml:"cities"`
}

// Load reads the configuration from ~/.config/worldclock.yaml
// If the file doesn't exist, it creates a default one
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate timezones
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that all timezone identifiers are valid
func (c *Config) Validate() error {
	if len(c.Cities) == 0 {
		return fmt.Errorf("no cities configured")
	}

	for i, city := range c.Cities {
		if city.Name == "" {
			return fmt.Errorf("city at index %d has no name", i)
		}
		if city.Timezone == "" {
			return fmt.Errorf("city '%s' has no timezone", city.Name)
		}
		// Validate timezone using time.LoadLocation
		if _, err := time.LoadLocation(city.Timezone); err != nil {
			return fmt.Errorf("invalid timezone '%s' for city '%s': %w", city.Timezone, city.Name, err)
		}
	}

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "worldclock.yaml"), nil
}

// createDefaultConfig creates a default configuration file with system timezone
func createDefaultConfig(path string) error {
	// Create .config directory if it doesn't exist
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Get system timezone
	systemTZ := getSystemTimezone()

	// Create default config with only system timezone
	defaultConfig := Config{
		Cities: []City{
			{Name: "Local", Timezone: systemTZ},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

// getSystemTimezone returns the system's IANA timezone name
func getSystemTimezone() string {
	// Get local timezone
	loc := time.Local
	if loc != nil {
		return loc.String()
	}

	// Fallback to UTC if we can't determine
	return "UTC"
}
