# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Terminal User Interface (TUI) application in Go that displays multiple world clocks showing current time for different cities.

## Key Requirements

### Configuration
- Config file location: `~/.config/worldclock.yaml`
- Config structure:
  ```yaml
  cities:
    - name: "City Name"
      timezone: "IANA/Timezone"
  ```
- If config doesn't exist, create it with default entry using system timezone (labeled as "Local")
- Use `time.LoadLocation()` to validate IANA timezone identifiers

### TUI Framework
- **bubbletea** - Chosen framework using The Elm Architecture pattern
- **lipgloss** - Used for styling terminal UI components
- Supports grid layouts and real-time updates via tick messages

### Display Specifications
- Grid layout with multiple clock cards/panels
- Each clock card displays (top to bottom):
  * City Name (header)
  * Digital clock (24-hour format: HH:MM:SS)
  * Date and UTC offset (format: "YYYY-MM-DD - UTC±HH:MM")
- Clocks update every second
- Visually styled with borders/boxes
- Responsive to terminal window resizing
- Exit: Ctrl-C

## Project Structure

```
worldclock/
├── main.go           # Main application with bubbletea model, Update/View logic
├── config/
│   └── config.go     # Config loading, YAML parsing, validation, default generation
├── clock/
│   └── clock.go      # Clock struct, time formatting (24h, date, UTC offset)
├── go.mod
├── go.sum
├── README.md
└── CLAUDE.md
```

### Architecture

**Three-layer separation:**

1. **config package** (`config/config.go`)
   - `Load()` - Reads from `~/.config/worldclock.yaml`, creates default if missing
   - `Validate()` - Validates all timezones using `time.LoadLocation()`
   - `createDefaultConfig()` - Generates config with system timezone

2. **clock package** (`clock/clock.go`)
   - `Clock` struct - Holds city name and `*time.Location`
   - `New()` - Creates clock with validated timezone
   - Format methods: `FormatTime()`, `FormatDate()`, `FormatUTCOffset()`, `FormatDateWithOffset()`

3. **main package** (`main.go`)
   - `model` - Bubbletea model holding clocks array, dimensions, error state
   - `Init()` - Returns initial tick command
   - `Update()` - Handles KeyMsg, WindowSizeMsg, tickMsg (every second)
   - `View()` - Renders grid layout using lipgloss
   - `renderClockCard()` - Creates individual clock cards with borders
   - `calculateColumns()` - Dynamic grid layout based on terminal width

## Development Commands

### Build
```bash
go build -o worldclock .
```

### Run
```bash
# Run the built binary
./worldclock

# Or run directly without building
go run .
```

### Install Dependencies
```bash
go get github.com/charmbracelet/bubbletea \
       github.com/charmbracelet/lipgloss \
       gopkg.in/yaml.v3
```

### Testing
```bash
go test ./...
```

## Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework using The Elm Architecture
- `github.com/charmbracelet/lipgloss` - Style definitions for terminal UIs
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Key Implementation Details

- **Real-time updates**: Uses `tea.Tick(time.Second, ...)` to send tickMsg every second
- **Grid layout**: `calculateColumns()` determines optimal column count based on terminal width
- **Styling**: lipgloss styles for title (cyan), time (magenta), date (gray), borders (blue)
- **Alt screen**: Uses `tea.WithAltScreen()` to preserve terminal history on exit
- **Exit handling**: Responds to both 'q' key and Ctrl+C
