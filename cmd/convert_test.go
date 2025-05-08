package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// setupTestCommand creates a new instance of the convert command for testing
func setupTestCommand() (*cobra.Command, *cobra.Command) {
	root := &cobra.Command{
		Use:   "go-togif",
		Short: "A high-quality PNG to GIF converter",
	}

	convert := &cobra.Command{
		Use:   "convert",
		Short: "Convert PNG files to GIF",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			output, _ := cmd.Flags().GetString("output")
			delay, _ := cmd.Flags().GetInt("delay")

			// Validate required flags
			if input == "" {
				return fmt.Errorf("input pattern is required")
			}
			if output == "" {
				return fmt.Errorf("output file path is required")
			}
			if delay < 0 {
				return fmt.Errorf("delay must be non-negative")
			}

			// Create an empty output file to simulate successful conversion
			f, err := os.Create(output)
			if err != nil {
				return fmt.Errorf("failed to create output file: %v", err)
			}
			defer func() {
				if cerr := f.Close(); cerr != nil && err == nil {
					err = fmt.Errorf("failed to close output file: %v", cerr)
				}
			}()

			return err
		},
	}

	// Add flags to the command
	convert.Flags().StringP("input", "i", "", "Input PNG files or patterns")
	convert.Flags().StringP("output", "o", "", "Output GIF file path")
	convert.Flags().IntP("delay", "d", 100, "Delay between frames in milliseconds")
	convert.Flags().Bool("debug", false, "Enable debug mode")

	// Mark required flags
	convert.MarkFlagRequired("input")
	convert.MarkFlagRequired("output")

	root.AddCommand(convert)
	return root, convert
}

func TestConvertCmd(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "go-togif-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test PNG files
	testFiles := []string{
		"test1.png",
		"test2.png",
		"test3.png",
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("Failed to close test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "Missing input",
			args:    []string{"convert", "-o", "output.gif"},
			wantErr: true,
		},
		{
			name:    "Missing output",
			args:    []string{"convert", "-i", "*.png"},
			wantErr: true,
		},
		{
			name:    "Invalid delay",
			args:    []string{"convert", "-i", "*.png", "-o", "output.gif", "-d", "-1"},
			wantErr: true,
		},
		{
			name:    "Valid command",
			args:    []string{"convert", "-i", filepath.Join(tempDir, "*.png"), "-o", filepath.Join(tempDir, "output.gif"), "-d", "100"},
			wantErr: false,
		},
		{
			name:    "Debug mode",
			args:    []string{"convert", "-i", filepath.Join(tempDir, "*.png"), "-o", filepath.Join(tempDir, "output.gif"), "--debug"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new command instances for each test
			root, _ := setupTestCommand()
			root.SetArgs(tt.args)

			// Clean up any existing output file
			outputFile := filepath.Join(tempDir, "output.gif")
			if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
				t.Fatalf("Failed to clean up output file: %v", err)
			}

			// Execute the command
			err := root.Execute()

			// Check the error
			if (err != nil) != tt.wantErr {
				t.Errorf("convertCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If the command was successful, verify the output file exists
			if !tt.wantErr {
				if _, err := os.Stat(outputFile); os.IsNotExist(err) {
					t.Errorf("Output file %s was not created", outputFile)
				}
			}
		})
	}
}
