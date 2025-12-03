# World Clock TUI

A Terminal User Interface (TUI) application that displays multiple world clocks showing the current time for different cities.

![World Clock Demo](https://via.placeholder.com/800x400?text=World+Clock+TUI)

## Features

- **Multiple Time Zones**: Display clocks for multiple cities simultaneously
- **Real-time Updates**: Clocks tick every second
- **Responsive Grid Layout**: Automatically adjusts to terminal window size
- **Beautiful UI**: Styled with borders, colors, and clean formatting
- **24-hour Format**: Time displayed in HH:MM:SS format
- **UTC Offset Display**: Shows date and UTC offset for each timezone
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

- `q` or `Ctrl+C` - Quit the application

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

## Project Structure

```
worldclock/
├── main.go           # Main application and bubbletea model
├── config/
│   └── config.go     # Configuration loading and validation
├── clock/
│   └── clock.go      # Clock logic and time formatting
├── go.mod            # Go module definition
└── go.sum            # Go dependencies
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework using The Elm Architecture
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

## License

MIT

## Author

Phil Tim
