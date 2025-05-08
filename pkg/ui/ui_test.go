package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelUpdate(t *testing.T) {
	// Get current working directory for absolute path tests
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	tests := []struct {
		name     string
		debug    bool
		total    int
		messages []tea.Msg
		wantDone bool
	}{
		{
			name:  "Normal mode progress",
			debug: false,
			total: 3,
			messages: []tea.Msg{
				ProgressMsg{CurrentFile: "file1.png", Processed: 0, Total: 3},
				ProgressMsg{CurrentFile: "file2.png", Processed: 1, Total: 3},
				ProgressMsg{CurrentFile: "file3.png", Processed: 2, Total: 3},
				ProgressMsg{CurrentFile: "Creating output GIF", Processed: 3, Total: 3, OutputFile: filepath.Join(cwd, "output.gif")},
			},
			wantDone: true,
		},
		{
			name:  "Debug mode progress",
			debug: true,
			total: 2,
			messages: []tea.Msg{
				ProgressMsg{CurrentFile: "file1.png", Processed: 0, Total: 2},
				ProgressMsg{CurrentFile: "file2.png", Processed: 1, Total: 2},
				ProgressMsg{CurrentFile: "Creating output GIF", Processed: 2, Total: 2, OutputFile: filepath.Join(cwd, "output.gif")},
			},
			wantDone: true,
		},
		{
			name:  "Quit on key press",
			debug: false,
			total: 3,
			messages: []tea.Msg{
				tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			},
			wantDone: false,
		},
		{
			name:  "Handle error",
			debug: false,
			total: 3,
			messages: []tea.Msg{
				errMsg{error: errors.New("test error")},
			},
			wantDone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel(tt.debug, tt.total)

			for _, msg := range tt.messages {
				newModel, newCmd := m.Update(msg)
				if newModel, ok := newModel.(model); ok {
					m = newModel
				}
				if newCmd != nil {
					// Execute the command to update the model
					newModel, _ = m.Update(tickMsg(time.Now()))
					if newModel, ok := newModel.(model); ok {
						m = newModel
					}
				}
			}

			if m.done != tt.wantDone {
				t.Errorf("Model.done = %v, want %v", m.done, tt.wantDone)
			}

			// Check processed files in debug mode
			if tt.debug && len(m.processedFiles) != tt.total {
				t.Errorf("Processed files count = %d, want %d", len(m.processedFiles), tt.total)
			}
		})
	}
}

func TestModelView(t *testing.T) {
	// Get current working directory for absolute path tests
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	tests := []struct {
		name       string
		debug      bool
		total      int
		processed  int
		done       bool
		err        error
		outputFile string
		want       string
	}{
		{
			name:       "Normal mode in progress",
			debug:      false,
			total:      3,
			processed:  1,
			done:       false,
			err:        nil,
			outputFile: "",
			want:       "Converting images",
		},
		{
			name:       "Debug mode completed",
			debug:      true,
			total:      2,
			processed:  2,
			done:       true,
			err:        nil,
			outputFile: filepath.Join(cwd, "output.gif"),
			want:       "Conversion completed",
		},
		{
			name:       "Error state",
			debug:      false,
			total:      3,
			processed:  0,
			done:       false,
			err:        errMsg{error: errors.New("test error")},
			outputFile: "",
			want:       "Error: test error\n",
		},
		{
			name:       "Normal mode completed",
			debug:      false,
			total:      3,
			processed:  3,
			done:       true,
			err:        nil,
			outputFile: filepath.Join(cwd, "output.gif"),
			want:       "Done! Processed 3 files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				debug:          tt.debug,
				totalFiles:     tt.total,
				processed:      tt.processed,
				done:           tt.done,
				err:            tt.err,
				processedFiles: make([]string, tt.total),
				outputFile:     tt.outputFile,
			}

			// Initialize processed files for debug mode
			if tt.debug {
				for i := 0; i < tt.total; i++ {
					m.processedFiles[i] = fmt.Sprintf("file%d.png", i+1)
				}
			}

			got := m.View()
			if tt.err != nil {
				if got != tt.want {
					t.Errorf("View() = %q, want %q", got, tt.want)
				}
			} else if !contains(got, tt.want) {
				t.Errorf("View() = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
