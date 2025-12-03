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

// GetSystemTimezone returns the system's IANA timezone name (exported version)
func GetSystemTimezone() string {
	return getSystemTimezone()
}

// Save writes the configuration to ~/.config/worldclock.yaml atomically
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Validate before saving
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Atomic write: write to temp file, then rename
	configDir := filepath.Dir(configPath)
	tempFile, err := os.CreateTemp(configDir, "worldclock-*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Write data
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Close temp file
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename temp file to actual config file
	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// AddCity adds a new city to the configuration
func (c *Config) AddCity(name, timezone string) error {
	// Check if city already exists
	for _, city := range c.Cities {
		if city.Name == name && city.Timezone == timezone {
			return fmt.Errorf("city '%s' already exists", name)
		}
	}

	// Validate timezone
	if _, err := time.LoadLocation(timezone); err != nil {
		return fmt.Errorf("invalid timezone '%s': %w", timezone, err)
	}

	// Add city
	c.Cities = append(c.Cities, City{
		Name:     name,
		Timezone: timezone,
	})

	return nil
}

// DeleteCities removes cities by name from the configuration
func (c *Config) DeleteCities(names []string) error {
	// Create a map for quick lookup
	toDelete := make(map[string]bool)
	for _, name := range names {
		toDelete[name] = true
	}

	// Filter cities
	var remaining []City
	for _, city := range c.Cities {
		if !toDelete[city.Name] {
			remaining = append(remaining, city)
		}
	}

	// Check if we have at least one city left
	if len(remaining) == 0 {
		return fmt.Errorf("cannot delete all cities")
	}

	c.Cities = remaining
	return nil
}

// HasCity checks if a city with the given name exists
func (c *Config) HasCity(name string) bool {
	for _, city := range c.Cities {
		if city.Name == name {
			return true
		}
	}
	return false
}
