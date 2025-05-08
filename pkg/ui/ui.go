package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	fileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
)

// ProgressMsg represents a progress update message
type ProgressMsg struct {
	CurrentFile string
	Processed   int
	Total       int
	OutputFile  string
}

type model struct {
	spinner        spinner.Model
	progress       progress.Model
	debug          bool
	totalFiles     int
	processed      int
	currentFile    string
	done           bool
	err            error
	processedFiles []string
	outputFile     string
}

type tickMsg time.Time
type errMsg struct{ error }

func (e errMsg) Error() string { return e.error.Error() }

func initialModel(debug bool, totalFiles int) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	p := progress.New(progress.WithDefaultGradient())

	return model{
		spinner:        s,
		progress:       p,
		debug:          debug,
		totalFiles:     totalFiles,
		processed:      0,
		done:           false,
		processedFiles: make([]string, 0, totalFiles),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	case tickMsg:
		if m.done {
			return m, nil
		}
		return m, tickCmd()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	case ProgressMsg:
		m.processed = msg.Processed
		m.currentFile = msg.CurrentFile
		if msg.Processed < m.totalFiles {
			m.processedFiles = append(m.processedFiles, msg.CurrentFile)
		}
		if msg.Processed >= msg.Total {
			m.done = true
			m.outputFile = msg.OutputFile
			return m, tea.Quit
		}
		return m, m.progress.IncrPercent(1.0 / float64(m.totalFiles))
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.done {
		if m.debug {
			var s strings.Builder
			s.WriteString("\n" + titleStyle.Render("Conversion completed! 🎉\n"))
			s.WriteString(fmt.Sprintf("\nProcessed %d files:\n", m.totalFiles))

			// Calculate the maximum width needed for the index
			maxIndexWidth := len(fmt.Sprintf("%d", len(m.processedFiles)))

			// Format each line with proper padding and truncate long paths
			for i, file := range m.processedFiles[:m.totalFiles] {
				// Remove the "temp/" prefix for cleaner output
				displayFile := strings.TrimPrefix(file, "temp/")
				if displayFile == file && len(file) > 50 {
					// If it's not in temp/ and the path is too long, truncate it
					displayFile = "..." + file[len(file)-47:]
				}

				indexStr := fmt.Sprintf("%*d", maxIndexWidth, i+1)
				s.WriteString(fmt.Sprintf("%s. %s\n", indexStr, displayFile))
			}
			if m.outputFile != "" {
				s.WriteString(fmt.Sprintf("\nGIF file generated at: %s\n", m.outputFile))
			}
			return s.String()
		}
		var s strings.Builder
		s.WriteString(fmt.Sprintf("\nDone! Processed %d files.\n", m.totalFiles))
		if m.outputFile != "" {
			s.WriteString(fmt.Sprintf("GIF file generated at: %s\n", m.outputFile))
		}
		return s.String()
	}

	var s strings.Builder
	if !m.debug {
		s.WriteString(fmt.Sprintf("\n%s Converting images...\n", m.spinner.View()))
		s.WriteString(fmt.Sprintf("Progress: %s\n", m.progress.ViewAs(float64(m.processed)/float64(m.totalFiles))))
		s.WriteString(helpStyle("\nPress q to quit"))
	}

	return s.String()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// RunUI starts the UI and returns a channel to send progress updates
func RunUI(debug bool, totalFiles int) chan ProgressMsg {
	progressChan := make(chan ProgressMsg)
	go func() {
		p := tea.NewProgram(initialModel(debug, totalFiles))
		go func() {
			for msg := range progressChan {
				p.Send(msg)
			}
		}()
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running UI: %v\n", err)
		}
	}()
	return progressChan
}
