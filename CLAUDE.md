# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Terminal User Interface (TUI) application in Go that displays multiple world clocks showing current time for different cities. Features interactive add/delete functionality with GeoNames database integration for searching 15,000+ cities worldwide.

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
- **bubbletea** - Framework using The Elm Architecture pattern
- **bubbles** - TUI components (viewport for scrolling, textinput for search)
- **lipgloss** - Styling terminal UI components
- Supports grid layouts, real-time updates, and modal views

### Display Specifications
- Grid layout with multiple clock cards/panels
- Each clock card displays (top to bottom):
  * City Name (header)
  * Digital clock (24-hour format: HH:MM:SS)
  * Date and UTC offset (format: "YYYY-MM-DD - UTC±HH:MM")
- Clocks update every second
- Clocks sorted by UTC offset (west to east)
- Viewport scrolling for small terminal windows
- Visually styled with borders/boxes
- Responsive to terminal window resizing
- Command bar at bottom showing available shortcuts
- Exit: Ctrl-C or 'q'

### Interactive Features
- **Add City ('a' key)**: Search GeoNames database, real-time filtering, select and add cities
- **Delete City ('d' key)**: Multi-select list, protected city detection, confirmation dialog
- **Protected Cities**: Cities matching system timezone cannot be deleted
- **Modal Views**: Separate views for add/delete/confirm operations
- **Keyboard Navigation**: Full keyboard-driven interface with ESC to cancel

## Project Structure

```
worldclock/
├── main.go              # Main application with view states, bubbletea model, Update/View logic
├── config/
│   └── config.go        # Config loading, YAML parsing, validation, add/delete operations
├── clock/
│   └── clock.go         # Clock struct, time formatting, UTC offset sorting
├── geonames/
│   └── geonames.go      # GeoNames database download, parsing, searching
├── go.mod
├── go.sum
├── README.md
├── CLAUDE.md
└── worldclock.yaml.example  # Example configuration file
```

### Architecture

**Four-layer separation:**

1. **config package** (`config/config.go`)
   - `Load()` - Reads from `~/.config/worldclock.yaml`, creates default if missing
   - `Validate()` - Validates all timezones using `time.LoadLocation()`
   - `Save()` - Atomically writes config (temp file + rename)
   - `AddCity()` - Adds new city, checks for duplicates
   - `DeleteCities()` - Removes cities by name, ensures at least one remains
   - `GetSystemTimezone()` - Returns system IANA timezone

2. **clock package** (`clock/clock.go`)
   - `Clock` struct - Holds city name and `*time.Location`
   - `New()` - Creates clock with validated timezone
   - Format methods: `FormatTime()`, `FormatDate()`, `FormatUTCOffset()`, `FormatDateWithOffset()`
   - `GetUTCOffset()` - Returns UTC offset in seconds
   - `SortByUTCOffset()` - Sorts clock slice by UTC offset (west to east)

3. **geonames package** (`geonames/geonames.go`)
   - `Database` struct - Holds parsed cities data with thread-safe access
   - `NewDatabase()` - Creates database instance
   - `LoadAsync()` - Downloads and loads data in background
   - `IsReady()` / `GetError()` - Thread-safe status checks
   - `Search()` - Searches cities, returns exact matches first, then partial matches
   - Downloads from: http://download.geonames.org/export/dump/cities15000.zip
   - Caches to: `~/.cache/worldclock/cities15000.txt`

4. **main package** (`main.go`)
   - **View States**: `viewMain`, `viewAdd`, `viewDelete`, `viewConfirm`
   - `model` - Holds: config, clocks, geonames DB, viewport, view state, search/delete state
   - `Init()` - Returns tick command and GeoNames check command
   - `Update()` - Routes to state-specific key handlers, updates sub-components
   - `View()` - Routes to state-specific renderers
   - **Key Handlers**: `handleMainKeys()`, `handleAddKeys()`, `handleDeleteKeys()`, `handleConfirmKeys()`
   - **Renderers**: `renderMain()`, `renderAdd()`, `renderDelete()`, `renderConfirm()`, `renderCommandBar()`
   - `reloadClocks()` - Reloads config, recreates and sorts clocks
   - `isCityProtected()` - Checks if city matches system timezone

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
       github.com/charmbracelet/bubbles/viewport \
       github.com/charmbracelet/bubbles/textinput \
       github.com/charmbracelet/lipgloss \
       gopkg.in/yaml.v3
```

### Testing
```bash
go test ./...
```

## Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework using The Elm Architecture
- `github.com/charmbracelet/bubbles` - TUI components (viewport, textinput)
- `github.com/charmbracelet/lipgloss` - Style definitions for terminal UIs
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Key Implementation Details

### Core Features
- **Real-time updates**: `tea.Tick(time.Second, ...)` sends tickMsg every second
- **Grid layout**: `calculateColumns()` determines optimal column count based on terminal width
- **Clock sorting**: `SortByUTCOffset()` orders clocks west to east automatically
- **Viewport scrolling**: `bubbles/viewport` handles content overflow for small terminals
- **Alt screen**: `tea.WithAltScreen()` preserves terminal history on exit

### View State Management
- **State enum**: `viewMain`, `viewAdd`, `viewDelete`, `viewConfirm`
- **State transitions**: ESC always cancels, Enter confirms actions
- **Modal rendering**: Each view has dedicated render function
- **Key routing**: `handleKeyPress()` routes to state-specific handlers

### GeoNames Integration
- **Async download**: Background goroutine downloads on first run
- **Thread-safe**: RWMutex protects shared state
- **Efficient search**: Exact matches first, then prefix, then contains
- **Max results**: Limited to 50 to keep UI responsive
- **Cache management**: Single file at `~/.cache/worldclock/cities15000.txt`

### Config Management
- **Atomic writes**: Write to temp file, then rename for safety
- **Validation**: All timezones validated before save
- **Protected cities**: System timezone cities cannot be deleted
- **Duplicate check**: Prevents adding same city twice

### Styling
- **Colors**: Title (cyan/86), Time (magenta/205), Date (gray/241), Borders (blue/62)
- **Command bar**: Dark background (235) with gray text (240)
- **Protected items**: Grayed out (240) in delete mode
- **Selected items**: Bold magenta (205) for current selection
