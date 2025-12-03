package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/philtim/worldclock/clock"
	"github.com/philtim/worldclock/config"
)

// tickMsg is sent every second to update the clocks
type tickMsg time.Time

// model represents the application state
type model struct {
	clocks   []*clock.Clock
	viewport viewport.Model
	ready    bool
	err      error
	width    int
	height   int
	quitting bool
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Initialize viewport when we first get dimensions
			headerHeight := 0
			footerHeight := 2
			verticalMarginHeight := headerHeight + footerHeight

			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			// Update viewport dimensions
			headerHeight := 0
			footerHeight := 2
			verticalMarginHeight := headerHeight + footerHeight
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	case tickMsg:
		// Return command to tick again
		return m, tickCmd()

	case error:
		m.err = msg
		return m, tea.Quit
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the UI
func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	if !m.ready {
		return "Initializing..."
	}

	// Render the clocks content
	content := renderClocks(m.clocks, m.width, m.viewport.Height)

	// Set viewport content
	m.viewport.SetContent(content)

	// Build the full view with footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press 'q' or 'Ctrl+C' to quit")

	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer)
}

// tickCmd returns a command that sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
	cardWidth := (width / cols) - 4 // Leave some margin

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
		Width(width)

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
		Width(width)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Margin(1)

	// Build card content
	title := titleStyle.Render(clk.Name)
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

	// Initialize model
	m := model{
		clocks: clocks,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
