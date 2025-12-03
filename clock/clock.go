package clock

import (
	"fmt"
	"sort"
	"time"
)

// Clock represents a world clock for a specific timezone
type Clock struct {
	Name     string
	Location *time.Location
}

// New creates a new Clock instance
func New(name, timezone string) (*Clock, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone '%s': %w", timezone, err)
	}

	return &Clock{
		Name:     name,
		Location: loc,
	}, nil
}

// GetTime returns the current time in the clock's timezone
func (c *Clock) GetTime() time.Time {
	return time.Now().In(c.Location)
}

// FormatTime returns the time in 24-hour format (HH:MM:SS)
func (c *Clock) FormatTime() string {
	return c.GetTime().Format("15:04:05")
}

// FormatDate returns the date in YYYY-MM-DD format
func (c *Clock) FormatDate() string {
	return c.GetTime().Format("2006-01-02")
}

// FormatUTCOffset returns the UTC offset in ±HH:MM format
func (c *Clock) FormatUTCOffset() string {
	t := c.GetTime()
	_, offset := t.Zone()

	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}

	hours := offset / 3600
	minutes := (offset % 3600) / 60

	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, minutes)
}

// FormatDateWithOffset returns the date and UTC offset
// Format: "YYYY-MM-DD - UTC±HH:MM"
func (c *Clock) FormatDateWithOffset() string {
	return fmt.Sprintf("%s - %s", c.FormatDate(), c.FormatUTCOffset())
}

// GetUTCOffset returns the UTC offset in seconds
func (c *Clock) GetUTCOffset() int {
	t := c.GetTime()
	_, offset := t.Zone()
	return offset
}

// SortByUTCOffset sorts a slice of clocks by their UTC offset (west to east)
func SortByUTCOffset(clocks []*Clock) {
	sort.Slice(clocks, func(i, j int) bool {
		return clocks[i].GetUTCOffset() < clocks[j].GetUTCOffset()
	})
}
