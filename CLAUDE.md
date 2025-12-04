# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Terminal User Interface (TUI) application in Go that displays multiple world clocks showing current time for different cities. Features interactive add/delete functionality with GeoNames database integration for searching 15,000+ cities worldwide. Displays 4 clocks per row by default with responsive grid layout and symmetric spacing.

## Key Requirements

### Configuration
- Config file location: `~/.config/worldclock.yaml`
- Config structure:
  ```yaml
  cities:
    - name: "City Name"
      timezone: "IANA/Timezone"
  ```
- If config doesn't exist, returns empty config (no default creation)
- Empty state shows: "Press 'a' to add a new city"
- Use `time.LoadLocation()` to validate IANA timezone identifiers
- Config allows empty cities list (no minimum requirement)

### TUI Framework
- **bubbletea** - Framework using The Elm Architecture pattern
- **bubbles** - TUI components (viewport for scrolling, textinput for search)
- **lipgloss** - Styling terminal UI components
- Supports grid layouts, real-time updates, and modal views

### Display Specifications
- Grid layout with multiple clock cards/panels (4 columns by default)
- Each clock card displays (top to bottom):
  * City Name (UPPERCASE, header with padding)
  * Digital clock (24-hour format: HH:MM:SS)
  * Date and UTC offset (format: "YYYY-MM-DD - UTC±HH:MM")
- Clocks update every second
- Clocks sorted by UTC offset (west to east)
- Viewport scrolling for small terminal windows
- Visually styled with rounded borders
- Responsive to terminal window resizing (adapts columns: 4 → 2 → 1)
- Command bar anchored at very bottom of terminal (no extra lines)
- Loading spinner in lower right corner while GeoNames downloads
- Exit: Ctrl-C or 'q'

### Interactive Features
- **Add City ('a' key)**: Search GeoNames database, real-time filtering, select and add cities
- **Delete City ('d' key)**: Multi-select list with checkboxes, confirmation dialog
- **Modal Views**: Separate views for add/delete/confirm operations
- **Keyboard Navigation**: Full keyboard-driven interface with ESC to cancel
- **GeoNames Status**: Spinner animation shows download progress, "Ready" when complete

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
├── .editorconfig        # Editor configuration for consistent formatting
├── .gitignore           # Git ignore patterns
├── README.md
├── CLAUDE.md
└── worldclock.yaml.example  # Example configuration file
```

### Architecture

**Four-layer separation:**

1. **config package** (`config/config.go`)
   - `Load()` - Reads from `~/.config/worldclock.yaml`, returns empty config if missing
   - `Validate()` - Validates all timezones using `time.LoadLocation()`, allows empty cities list
   - `Save()` - Atomically writes config (temp file + rename)
   - `AddCity()` - Adds new city, checks for duplicates
   - `DeleteCities()` - Removes cities by name, allows deleting all cities
   - `HasCity()` - Checks if city exists by name
   - `ConfigExists()` - Checks if config file exists
   - `CreateDefaultConfigWithCity()` - Creates config with specified city name
   - `GetSystemTimezone()` - Returns system IANA timezone

2. **clock package** (`clock/clock.go`)
   - `Clock` struct - Holds city name and `*time.Location`
   - `New()` - Creates clock with validated timezone
   - Format methods: `FormatTime()`, `FormatDate()`, `FormatUTCOffset()`, `FormatDateWithOffset()`
   - `GetUTCOffset()` - Returns UTC offset in seconds
   - `SortByUTCOffset()` - Sorts clock slice by UTC offset (west to east)

3. **geonames package** (`geonames/geonames.go`)
   - `Database` struct - Holds parsed cities data with thread-safe access (RWMutex)
   - `City` struct - Name, CountryCode, Timezone, Population
   - `NewDatabase()` - Creates database instance
   - `LoadAsync()` - Downloads and loads data in background goroutine
   - `LoadSync()` - Blocking load for synchronous operations
   - `IsReady()` / `GetError()` - Thread-safe status checks
   - `Search()` - Searches cities, returns exact matches first, then partial matches
   - `FindBestCityForTimezone()` - Returns most populous city in given timezone
   - Downloads from: http://download.geonames.org/export/dump/cities15000.zip
   - Caches to: `~/.cache/worldclock/cities15000.txt`

4. **main package** (`main.go`)
   - **View States**: `viewMain`, `viewAdd`, `viewDelete`, `viewConfirm`
   - `model` - Holds: config, clocks, geonames DB, viewport, view state, search/delete state, spinner state
   - **Messages**: `tickMsg`, `spinnerTickMsg`, `geonamesReadyMsg`, `geonamesErrorMsg`
   - `Init()` - Returns tick command, spinner tick, and GeoNames check command
   - `Update()` - Routes to state-specific key handlers, updates sub-components, manages spinner
   - `View()` - Routes to state-specific renderers
   - **Key Handlers**: `handleMainKeys()`, `handleAddKeys()`, `handleDeleteKeys()`, `handleConfirmKeys()`
   - **Renderers**: `renderMain()`, `renderAdd()`, `renderDelete()`, `renderConfirm()`, `renderCommandBar()`, `renderClocks()`, `renderClockCard()`
   - `reloadClocks()` - Reloads config, recreates and sorts clocks
   - `calculateColumns()` - Determines optimal column count based on width and city name lengths
   - **Spinner**: Braille pattern animation (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏) updates every 100ms

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
- **Grid layout**: `calculateColumns()` determines optimal column count (4 → 2 → 1) based on terminal width and city name lengths
- **Clock sorting**: `SortByUTCOffset()` orders clocks west to east automatically
- **Viewport scrolling**: `bubbles/viewport` handles content overflow for small terminals
- **Alt screen**: `tea.WithAltScreen()` preserves terminal history on exit
- **Loading indicator**: Spinner animation shows GeoNames download progress in command bar

### Layout System
- **CSS-style spacing**: Cards handle their own margins, no global window padding
- **Card overhead**: 8 characters (border: 2, padding: 4, margins: 2)
- **Width calculation**: `widthPerCard = terminalWidth / columns`
- **Content width**: `cardWidth = widthPerCard - 8`
- **Symmetric spacing**: Left and right margins on each card create equal gaps
- **Viewport height**: `terminalHeight - 2` (reserves 1 newline + 1 command bar line)
- **Command bar position**: Always anchored at very bottom with no extra lines

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
- **Validation**: All timezones validated before save using `time.LoadLocation()`
- **Empty config allowed**: No minimum city requirement, can delete all cities
- **Duplicate check**: Prevents adding same city/timezone combination twice
- **No auto-creation**: Returns empty config if file doesn't exist

### Styling
- **Clock cards**: Rounded borders with blue foreground (62)
- **City names**: Bold cyan (86), UPPERCASE, with padding
- **Time**: Bold magenta (205), centered
- **Date**: Gray (241), centered
- **Command bar**: Dark background (235) with gray text (240)
- **Selected items**: Bold magenta (205) with ">" prefix
- **Card spacing**: 1-char margins on all sides (top, right, bottom, left)
- **Text alignment**: All clock content centered within card

### Code Formatting
- **EditorConfig**: `.editorconfig` enforces consistent formatting across all editors
- **Go files**: 4-space indentation (not tabs)
- **YAML files**: 2-space indentation
- **Line endings**: Unix-style (LF)
- **Charset**: UTF-8
- **Trailing whitespace**: Automatically trimmed
- **Final newline**: Always inserted
- All Go files in the project use spaces for indentation to ensure consistency

## Recent Improvements

### Layout and Spacing (2025-12-04)
- **CSS-style layout**: Removed global window padding, cards now handle their own margins
- **Symmetric spacing**: Equal spacing on left and right sides via card margins
- **Card margins**: `Margin(1, 1, 0, 1)` provides 1-char spacing on top/right/left
- **Width distribution**: Percentage-based calculation divides terminal width equally among columns
- **Reduced spacing**: Removed extra line between city name and time for compact display
- **Command bar anchoring**: Fixed to sit at very bottom with no extra empty lines below

### Code Consistency (2025-12-04)
- **EditorConfig added**: Enforces consistent formatting across all editors and IDEs
- **Tab to space conversion**: All Go files converted from tabs to 4-space indentation
- **Build verification**: Confirmed all files compile successfully after conversion

### Feature Simplification (2025-12-04)
- **Removed default timezone**: No auto-creation of config file on first run
- **Empty state**: Shows helpful "Press 'a' to add a new city" message when no clocks configured
- **No minimum cities**: Allows deleting all cities, returning to empty state
- **Removed protected cities**: Users can delete any city including system timezone matches
