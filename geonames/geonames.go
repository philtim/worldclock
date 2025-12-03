package geonames

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// GeoNamesURL is the download URL for cities with 15000+ population
	GeoNamesURL = "http://download.geonames.org/export/dump/cities15000.zip"
	// CacheFileName is the name of the cached cities file
	CacheFileName = "cities15000.txt"
)

// City represents a city from the GeoNames database
type City struct {
	Name        string
	CountryCode string
	Timezone    string
}

// Database holds the GeoNames cities data
type Database struct {
	cities []City
	ready  bool
	err    error
	mu     sync.RWMutex
}

// NewDatabase creates a new GeoNames database instance
func NewDatabase() *Database {
	return &Database{
		cities: []City{},
		ready:  false,
	}
}

// LoadAsync loads the GeoNames database asynchronously
func (db *Database) LoadAsync() {
	go func() {
		if err := db.load(); err != nil {
			db.mu.Lock()
			db.err = err
			db.mu.Unlock()
		}
	}()
}

// load downloads (if needed) and loads the GeoNames database
func (db *Database) load() error {
	cachePath, err := getCachePath()
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// Download and extract
		if err := downloadAndExtract(cachePath); err != nil {
			return fmt.Errorf("failed to download GeoNames data: %w", err)
		}
	}

	// Parse the file
	cities, err := parseFile(cachePath)
	if err != nil {
		return fmt.Errorf("failed to parse GeoNames data: %w", err)
	}

	db.mu.Lock()
	db.cities = cities
	db.ready = true
	db.mu.Unlock()

	return nil
}

// IsReady returns whether the database is loaded and ready
func (db *Database) IsReady() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.ready
}

// GetError returns any error that occurred during loading
func (db *Database) GetError() error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.err
}

// Search searches for cities matching the query
// Returns top maxResults matches
func (db *Database) Search(query string, maxResults int) []City {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.ready {
		return []City{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if len(query) < 3 {
		return []City{}
	}

	var exactMatches []City
	var partialMatches []City

	for _, city := range db.cities {
		cityNameLower := strings.ToLower(city.Name)

		// Exact match
		if cityNameLower == query {
			exactMatches = append(exactMatches, city)
		} else if strings.HasPrefix(cityNameLower, query) {
			// Prefix match
			partialMatches = append(partialMatches, city)
		} else if strings.Contains(cityNameLower, query) {
			// Contains match
			partialMatches = append(partialMatches, city)
		}

		// Stop if we have enough results
		if len(exactMatches)+len(partialMatches) >= maxResults {
			break
		}
	}

	// Combine results: exact matches first, then partial
	results := append(exactMatches, partialMatches...)
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// getCachePath returns the path to the cache file
func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "worldclock")
	return filepath.Join(cacheDir, CacheFileName), nil
}

// downloadAndExtract downloads the GeoNames zip file and extracts it
func downloadAndExtract(targetPath string) error {
	// Create cache directory
	cacheDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download zip file to temporary location
	tempZip := filepath.Join(cacheDir, "cities15000.zip")
	if err := downloadFile(GeoNamesURL, tempZip); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer os.Remove(tempZip) // Clean up zip file after extraction

	// Extract the txt file from zip
	if err := extractFile(tempZip, CacheFileName, targetPath); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}

// downloadFile downloads a file from URL to filepath
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractFile extracts a specific file from a zip archive
func extractFile(zipPath, fileName, targetPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == fileName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}

	return fmt.Errorf("file %s not found in zip archive", fileName)
}

// parseFile parses the GeoNames cities15000.txt file
func parseFile(path string) ([]City, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cities []City
	scanner := bufio.NewScanner(file)

	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")

		// We need at least 18 fields (timezone is at index 17)
		if len(fields) < 18 {
			continue
		}

		name := fields[1]        // City name
		countryCode := fields[8] // Country code
		timezone := fields[17]   // Timezone

		// Skip if timezone is empty
		if timezone == "" {
			continue
		}

		cities = append(cities, City{
			Name:        name,
			CountryCode: countryCode,
			Timezone:    timezone,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cities, nil
}
