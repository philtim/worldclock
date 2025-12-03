# World Clock TUI

A Terminal User Interface (TUI) application that displays multiple world clocks showing the current time for different cities.

![World Clock Demo](https://via.placeholder.com/800x400?text=World+Clock+TUI)

## Features

- **Multiple Time Zones**: Display clocks for multiple cities simultaneously
- **Real-time Updates**: Clocks tick every second
- **Responsive Grid Layout**: Automatically adjusts to terminal window size with scrolling support
- **Beautiful UI**: Styled with borders, colors, and clean formatting
- **24-hour Format**: Time displayed in HH:MM:SS format
- **UTC Offset Display**: Shows date and UTC offset for each timezone
- **Sorted by Timezone**: Clocks automatically sorted west to east by UTC offset
- **Add Cities**: Search and add cities from GeoNames database (15,000+ cities)
- **Delete Cities**: Multi-select delete with protected system timezone
- **Interactive TUI**: Full keyboard-driven interface with modal views
- **YAML Configuration**: Easy configuration via `~/.config/worldclock.yaml`
- **Automatic Setup**: Creates default config with system timezone on first run

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from Source

```bash
# Clone the repository
git clone https://github.com/philtim/worldclock.git
cd worldclock

# Build the application
go build -o worldclock .

# Run it
./worldclock
```

### Install

```bash
# Install to $GOPATH/bin
go install github.com/philtim/worldclock@latest
```

## Configuration

The application reads configuration from `~/.config/worldclock.yaml`.

### Configuration Format

```yaml
cities:
  - name: "City Name"
    timezone: "IANA/Timezone"
  - name: "Another City"
    timezone: "Another/Timezone"
```

### Example Configuration

```yaml
cities:
  - name: "Kailua-Kona"
    timezone: "Pacific/Honolulu"
  - name: "Medicine Hat"
    timezone: "America/Edmonton"
  - name: "Germany"
    timezone: "Europe/Berlin"
  - name: "Manila"
    timezone: "Asia/Manila"
```

### Default Configuration

On first run, if no configuration file exists, the application will create one with your current system timezone:

```yaml
cities:
  - name: "Local"
    timezone: "America/Los_Angeles"  # Your system timezone
```

### Finding Timezone Names

Use IANA timezone database names. Common examples:

- **North America**: `America/New_York`, `America/Chicago`, `America/Denver`, `America/Los_Angeles`
- **Europe**: `Europe/London`, `Europe/Paris`, `Europe/Berlin`, `Europe/Moscow`
- **Asia**: `Asia/Tokyo`, `Asia/Shanghai`, `Asia/Dubai`, `Asia/Manila`
- **Pacific**: `Pacific/Honolulu`, `Pacific/Auckland`, `Pacific/Fiji`
- **UTC**: `UTC`

Full list: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

## Usage

### Run the Application

```bash
./worldclock
```

### Keyboard Controls

#### Main View
- `a` - Add a new city (search from GeoNames database)
- `d` - Delete cities (multi-select mode)
- `q` or `Ctrl+C` - Quit the application
- `↑/↓` or `PgUp/PgDn` - Scroll through clocks (if terminal is small)

#### Add City Mode
- Type to search cities (minimum 3 characters)
- `↑/↓` - Navigate search results
- `Enter` - Add selected city
- `ESC` - Cancel and return to main view

#### Delete City Mode
- `↑/↓` - Navigate city list
- `Space` - Toggle selection (protected cities cannot be selected)
- `Enter` - Confirm deletion (shows confirmation dialog)
- `ESC` - Cancel and return to main view

#### Confirmation Dialog
- `y` - Confirm action
- `n` or `ESC` - Cancel action

### Display Layout

Each clock card shows:
```
┌──────────────────────┐
│      City Name       │
│                      │
│      15:04:05       │
│                      │
│ 2025-12-03 - UTC-08:00│
└──────────────────────┘
```

Clocks are automatically sorted by UTC offset (west to east).

### Adding Cities

Press `a` to enter Add City mode. The application uses the GeoNames database containing over 15,000 cities worldwide.

**First Run**: The GeoNames database (cities15000.zip, ~4MB) will be downloaded automatically in the background to `~/.cache/worldclock/`. The add feature becomes available once the download completes (usually takes a few seconds).

**Search Tips**:
- Type at least 3 characters to start searching
- Search is case-insensitive
- Exact matches appear first, followed by partial matches
- Results show: City Name, Country Code, and Timezone

**Example**:
1. Press `a`
2. Type "berl" to search for Berlin
3. Use `↑/↓` to select "Berlin, DE (Europe/Berlin)"
4. Press `Enter` to add

### Deleting Cities

Press `d` to enter Delete Cities mode with multi-select functionality.

**Protected Cities**: Cities matching your system timezone are automatically protected and cannot be deleted. They appear grayed out with a "(protected)" label.

**Example**:
1. Press `d`
2. Use `↑/↓` to navigate
3. Press `Space` to select/deselect cities
4. Press `Enter` to confirm deletion
5. Press `y` in the confirmation dialog

### GeoNames Database

- **Source**: http://download.geonames.org/export/dump/cities15000.zip
- **Cache Location**: `~/.cache/worldclock/cities15000.txt`
- **Size**: ~4MB compressed, ~12MB uncompressed
- **Updates**: Delete the cache file to re-download latest data

## Project Structure

```
worldclock/
├── main.go              # Main application with view states and TUI logic
├── config/
│   └── config.go        # Configuration loading, validation, add/delete
├── clock/
│   └── clock.go         # Clock logic, time formatting, and sorting
├── geonames/
│   └── geonames.go      # GeoNames database download, parsing, and search
├── go.mod               # Go module definition
└── go.sum               # Go dependencies
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework using The Elm Architecture
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components (viewport, textinput)
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UIs
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML parser

## Development

### Build

```bash
go build -o worldclock .
```

### Run without Building

```bash
go run .
```

### Run Tests

```bash
go test ./...
```

## Troubleshooting

### Invalid Timezone Error

If you see an error like `invalid timezone 'XXX' for city 'YYY'`, check that:
1. The timezone name is a valid IANA timezone identifier
2. The timezone name is spelled correctly (case-sensitive)
3. Use forward slashes `/` not backslashes `\`

### Config File Not Found

The config file should be at `~/.config/worldclock.yaml`. If the directory doesn't exist, the application will create it automatically on first run.

### GeoNames Download Failed

If the GeoNames database fails to download:
1. Check your internet connection
2. The download URL may be temporarily unavailable
3. Try manually downloading from: http://download.geonames.org/export/dump/cities15000.zip
4. Extract `cities15000.txt` to `~/.cache/worldclock/cities15000.txt`

### Cannot Delete Last City

The application requires at least one city to be configured. If you try to delete all cities, you'll see an error.

### "City Already Exists" Error

When adding a city, if it already exists in your configuration (same name and timezone), you'll see this error. Check your current cities with `d` (Delete mode) to see what's configured.

## License

MIT

## Author

Phil Tim
