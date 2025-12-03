package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/philtim/worldclock/clock"
	"github.com/philtim/worldclock/config"
	"github.com/philtim/worldclock/geonames"
)

// viewState represents the current view state
type viewState int

const (
	viewMain viewState = iota
	viewAdd
	viewDelete
	viewConfirm
)

// tickMsg is sent every second to update the clocks
type tickMsg time.Time

// geonamesReadyMsg is sent when GeoNames database is ready
type geonamesReadyMsg struct{}

// geonamesErrorMsg is sent when GeoNames fails to load
type geonamesErrorMsg struct{ err error }

// model represents the application state
type model struct {
	// Core data
	cfg      *config.Config
	clocks   []*clock.Clock
	geonamesDB *geonames.Database
	systemTZ string

	// View state
	state    viewState
	viewport viewport.Model
	ready    bool
	err      error
	width    int
	height   int
	quitting bool

	// Add mode state
	searchInput       textinput.Model
	searchResults     []geonames.City
	selectedResult    int
	justEnteredAddMode bool // Flag to prevent initial key from appearing in input

	// Delete mode state
	deleteList     []string // List of city names
	deleteSelected map[int]bool
	deleteCursor   int

	// Confirm mode state
	confirmMsg     string
	confirmAction  func() error
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		checkGeoNamesCmd(m.geonamesDB),
	)
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Initialize viewport
			m.viewport = viewport.New(msg.Width, msg.Height-3) // Reserve space for command bar
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 3
		}

	case tickMsg:
		cmds = append(cmds, tickCmd())

	case geonamesReadyMsg:
		// GeoNames database is ready, no action needed

	case geonamesErrorMsg:
		m.err = msg.err

	case error:
		m.err = msg
		return m, tea.Quit
	}

	// Update sub-components based on state
	switch m.state {
	case viewAdd:
		// Only update searchInput if we didn't just enter add mode
		// (prevents the 'a' key from appearing in the input field)
		if !m.justEnteredAddMode {
			m.searchInput, cmd = m.searchInput.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Update search results when input changes
			if m.geonamesDB.IsReady() {
				m.searchResults = m.geonamesDB.Search(m.searchInput.Value(), 50)
				if m.selectedResult >= len(m.searchResults) {
					m.selectedResult = 0
				}
			}
		} else {
			// Reset the flag after first update cycle
			m.justEnteredAddMode = false
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input based on current view state
func (m *model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch m.state {
	case viewMain:
		return m.handleMainKeys(msg)
	case viewAdd:
		return m.handleAddKeys(msg)
	case viewDelete:
		return m.handleDeleteKeys(msg)
	case viewConfirm:
		return m.handleConfirmKeys(msg)
	}
	return nil
}

// handleMainKeys handles keys in main view
func (m *model) handleMainKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return tea.Quit

	case "a":
		// Enter add mode
		if m.geonamesDB.IsReady() {
			m.state = viewAdd
			m.searchInput.Reset()
			m.searchResults = []geonames.City{}
			m.selectedResult = 0
			m.justEnteredAddMode = true // Prevent 'a' key from appearing in input
			m.searchInput.Focus()
			return textinput.Blink
		}

	case "d":
		// Enter delete mode
		m.state = viewDelete
		m.deleteList = []string{}
		for _, city := range m.cfg.Cities {
			m.deleteList = append(m.deleteList, city.Name)
		}
		m.deleteSelected = make(map[int]bool)
		m.deleteCursor = 0
	}

	return nil
}

// handleAddKeys handles keys in add view
func (m *model) handleAddKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Cancel and return to main
		m.state = viewMain
		return nil

	case "up":
		if m.selectedResult > 0 {
			m.selectedResult--
		}

	case "down":
		if m.selectedResult < len(m.searchResults)-1 {
			m.selectedResult++
		}

	case "enter":
		// Add selected city
		if len(m.searchResults) > 0 && m.selectedResult < len(m.searchResults) {
			city := m.searchResults[m.selectedResult]
			if err := m.cfg.AddCity(city.Name, city.Timezone); err != nil {
				m.err = err
				return nil
			}
			if err := m.cfg.Save(); err != nil {
				m.err = err
				return nil
			}
			// Reload clocks
			return m.reloadClocks()
		}
	}

	return nil
}

// handleDeleteKeys handles keys in delete view
func (m *model) handleDeleteKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Cancel and return to main
		m.state = viewMain
		return nil

	case "up":
		if m.deleteCursor > 0 {
			m.deleteCursor--
		}

	case "down":
		if m.deleteCursor < len(m.deleteList)-1 {
			m.deleteCursor++
		}

	case " ":
		// Toggle selection (but not for protected cities)
		if !m.isCityProtected(m.deleteList[m.deleteCursor]) {
			m.deleteSelected[m.deleteCursor] = !m.deleteSelected[m.deleteCursor]
		}

	case "enter":
		// Delete selected cities
		if len(m.deleteSelected) == 0 {
			m.err = fmt.Errorf("no cities selected")
			return nil
		}

		// Collect selected city names
		var toDelete []string
		for idx := range m.deleteSelected {
			if m.deleteSelected[idx] {
				toDelete = append(toDelete, m.deleteList[idx])
			}
		}

		// Set up confirmation
		m.state = viewConfirm
		if len(toDelete) == 1 {
			m.confirmMsg = fmt.Sprintf("Delete '%s'? (y/n)", toDelete[0])
		} else {
			m.confirmMsg = fmt.Sprintf("Delete %d selected cities? (y/n)", len(toDelete))
		}
		m.confirmAction = func() error {
			if err := m.cfg.DeleteCities(toDelete); err != nil {
				return err
			}
			return m.cfg.Save()
		}
	}

	return nil
}

// handleConfirmKeys handles keys in confirm view
func (m *model) handleConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y":
		// Confirm action
		if err := m.confirmAction(); err != nil {
			m.err = err
			m.state = viewMain
			return nil
		}
		// Reload clocks and return to main
		return m.reloadClocks()

	case "n", "esc":
		// Cancel and return to main
		m.state = viewMain
		return nil
	}

	return nil
}

// reloadClocks reloads the configuration and recreates clocks
func (m *model) reloadClocks() tea.Cmd {
	// Reload config
	cfg, err := config.Load()
	if err != nil {
		m.err = err
		m.state = viewMain
		return nil
	}
	m.cfg = cfg

	// Recreate clocks
	var clocks []*clock.Clock
	for _, city := range m.cfg.Cities {
		clk, err := clock.New(city.Name, city.Timezone)
		if err != nil {
			m.err = err
			m.state = viewMain
			return nil
		}
		clocks = append(clocks, clk)
	}

	// Sort by UTC offset
	clock.SortByUTCOffset(clocks)
	m.clocks = clocks

	// Return to main view
	m.state = viewMain
	return nil
}

// isCityProtected checks if a city is protected (matches system timezone)
func (m *model) isCityProtected(cityName string) bool {
	for _, city := range m.cfg.Cities {
		if city.Name == cityName && city.Timezone == m.systemTZ {
			return true
		}
	}
	return false
}

// View renders the UI
func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	if !m.ready {
		return "Initializing..."
	}

	switch m.state {
	case viewMain:
		return m.renderMain()
	case viewAdd:
		return m.renderAdd()
	case viewDelete:
		return m.renderDelete()
	case viewConfirm:
		return m.renderConfirm()
	}

	return ""
}

// renderMain renders the main clock view
func (m model) renderMain() string {
	// Render clocks
	content := renderClocks(m.clocks, m.width, m.viewport.Height)
	m.viewport.SetContent(content)

	// Command bar
	commandBar := m.renderCommandBar()

	return fmt.Sprintf("%s\n%s", m.viewport.View(), commandBar)
}

// renderAdd renders the add city view
func (m model) renderAdd() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 0)
	b.WriteString(titleStyle.Render("Add City"))
	b.WriteString("\n\n")

	// Check if GeoNames is ready
	if !m.geonamesDB.IsReady() {
		if m.geonamesDB.GetError() != nil {
			b.WriteString(fmt.Sprintf("Error loading city database: %v\n", m.geonamesDB.GetError()))
		} else {
			b.WriteString("Loading city database...\n")
		}
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press ESC to cancel"))
		return b.String()
	}

	// Search input
	b.WriteString("Search city (min 3 characters):\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	// Results
	if len(m.searchInput.Value()) < 3 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Type at least 3 characters to search..."))
	} else if len(m.searchResults) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("No cities found"))
	} else {
		b.WriteString(fmt.Sprintf("Results (%d):\n", len(m.searchResults)))
		// Show results (limit visible results)
		maxVisible := 10
		start := 0
		if m.selectedResult >= maxVisible {
			start = m.selectedResult - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(m.searchResults) {
			end = len(m.searchResults)
		}

		for i := start; i < end; i++ {
			city := m.searchResults[i]
			line := fmt.Sprintf("  %s, %s (%s)", city.Name, city.CountryCode, city.Timezone)

			if i == m.selectedResult {
				line = lipgloss.NewStyle().
					Foreground(lipgloss.Color("205")).
					Bold(true).
					Render("> " + line)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑/↓: Navigate | Enter: Select | ESC: Cancel"))

	return b.String()
}

// renderDelete renders the delete city view
func (m model) renderDelete() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 0)
	b.WriteString(titleStyle.Render("Delete Cities"))
	b.WriteString("\n\n")

	// List cities
	for i, cityName := range m.deleteList {
		isProtected := m.isCityProtected(cityName)
		isSelected := m.deleteSelected[i]
		isCursor := i == m.deleteCursor

		var line string
		if isProtected {
			line = fmt.Sprintf("  [ ] %s (protected)", cityName)
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(line)
		} else {
			checkbox := " "
			if isSelected {
				checkbox = "x"
			}
			line = fmt.Sprintf("  [%s] %s", checkbox, cityName)
		}

		if isCursor {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Render("> " + line)
		} else {
			line = "  " + line
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑/↓: Navigate | Space: Toggle | Enter: Delete | ESC: Cancel"))

	return b.String()
}

// renderConfirm renders the confirmation dialog
func (m model) renderConfirm() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 0)
	b.WriteString(titleStyle.Render("Confirm"))
	b.WriteString("\n\n")

	b.WriteString(m.confirmMsg)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("y: Yes | n/ESC: No"))

	return b.String()
}

// renderCommandBar renders the command bar at the bottom
func (m model) renderCommandBar() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	var commands string
	if m.geonamesDB.IsReady() {
		commands = "a: Add City | d: Delete Cities | q: Quit"
	} else {
		commands = "Loading city database... | d: Delete Cities | q: Quit"
	}

	return style.Render(commands)
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// checkGeoNamesCmd checks if GeoNames database is ready
func checkGeoNamesCmd(db *geonames.Database) tea.Cmd {
	return func() tea.Msg {
		// Check periodically until ready
		for i := 0; i < 300; i++ { // Check for up to 5 minutes
			time.Sleep(100 * time.Millisecond)
			if db.IsReady() {
				return geonamesReadyMsg{}
			}
			if err := db.GetError(); err != nil {
				return geonamesErrorMsg{err: err}
			}
		}
		return geonamesErrorMsg{err: fmt.Errorf("timeout waiting for GeoNames database")}
	}
}

// renderClocks renders all clocks in a grid layout
func renderClocks(clocks []*clock.Clock, width, height int) string {
	if len(clocks) == 0 {
		return "No clocks configured\n"
	}

	// Calculate grid dimensions
	numClocks := len(clocks)
	cols := calculateColumns(numClocks, width)
	rows := (numClocks + cols - 1) / cols // Ceiling division

	// Determine clock card width
	// Account for: margin (2 left + 1 right), border (2), padding (4), and some buffer
	cardWidth := (width / cols) - 10

	// Create clock cards
	var clockCards []string
	for _, clk := range clocks {
		clockCards = append(clockCards, renderClockCard(clk, cardWidth))
	}

	// Arrange cards in grid
	var rows_content []string
	for row := 0; row < rows; row++ {
		var rowCards []string
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx < len(clockCards) {
				rowCards = append(rowCards, clockCards[idx])
			}
		}
		if len(rowCards) > 0 {
			rows_content = append(rows_content, lipgloss.JoinHorizontal(lipgloss.Top, rowCards...))
		}
	}

	return strings.Join(rows_content, "\n")
}

// renderClockCard renders a single clock card
func renderClockCard(clk *clock.Clock, width int) string {
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Align(lipgloss.Center).
		Width(width).
		PaddingTop(1).
		PaddingBottom(1)

	timeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Width(width).
		MarginTop(1).
		MarginBottom(1)

	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(width).
		PaddingBottom(1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 2).
		MarginLeft(1).
		MarginRight(0).
		MarginTop(1).
		MarginBottom(0)

	// Build card content with visual spacing
	title := titleStyle.Render(strings.ToUpper(clk.Name))

	timeStr := timeStyle.Render(clk.FormatTime())

	dateStr := dateStyle.Render(clk.FormatDateWithOffset())

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		timeStr,
		dateStr,
	)

	return cardStyle.Render(content)
}

// calculateColumns determines the number of columns based on terminal width
func calculateColumns(numClocks, width int) int {
	// Each clock needs approximately 30 characters minimum
	minCardWidth := 30

	maxCols := width / minCardWidth
	if maxCols < 1 {
		maxCols = 1
	}

	// Determine optimal number of columns
	if numClocks <= 2 {
		return numClocks
	} else if numClocks <= 4 && maxCols >= 2 {
		return 2
	} else if numClocks <= 6 && maxCols >= 3 {
		return 3
	} else if maxCols >= 4 {
		return 4
	}

	return maxCols
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create clocks from config
	var clocks []*clock.Clock
	for _, city := range cfg.Cities {
		clk, err := clock.New(city.Name, city.Timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating clock for %s: %v\n", city.Name, err)
			os.Exit(1)
		}
		clocks = append(clocks, clk)
	}

	// Sort clocks by UTC offset (west to east)
	clock.SortByUTCOffset(clocks)

	// Initialize GeoNames database (async)
	geonamesDB := geonames.NewDatabase()
	geonamesDB.LoadAsync()

	// Get system timezone
	systemTZ := config.GetSystemTimezone()

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search city..."
	ti.CharLimit = 50
	ti.Width = 50

	// Initialize model
	m := model{
		cfg:            cfg,
		clocks:         clocks,
		geonamesDB:     geonamesDB,
		systemTZ:       systemTZ,
		state:          viewMain,
		searchInput:    ti,
		searchResults:  []geonames.City{},
		selectedResult: 0,
		deleteSelected: make(map[int]bool),
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
