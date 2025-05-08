package cmd

import (
	"fmt"

	"github.com/jparrill/go-togif/pkg/converter"
	"github.com/spf13/cobra"
)

var (
	delay int
	debug bool
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert PNG images to GIF",
	Long: `Convert one or more PNG images to a GIF file.
You can use glob patterns (e.g., "*.png") or regex patterns (e.g., "^frame.*\\.png$") to specify input files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get input pattern from flag
		inputPattern, err := cmd.Flags().GetString("input")
		if err != nil {
			return err
		}

		// Get output file from flag
		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}

		// Expand input pattern
		inputFiles, err := converter.ExpandInputPattern(inputPattern)
		if err != nil {
			return fmt.Errorf("error expanding pattern %s: %v", inputPattern, err)
		}

		// Validate input files
		if err := converter.ValidateInputFiles(inputFiles); err != nil {
			return err
		}

		// Convert files
		return converter.ConvertPNGsToGIF(inputFiles, outputFile, delay, debug)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Add flags
	convertCmd.Flags().StringP("input", "i", "", "Input PNG file(s) pattern (required)")
	convertCmd.Flags().StringP("output", "o", "", "Output GIF file path (required)")
	convertCmd.Flags().IntVarP(&delay, "delay", "d", 100, "Delay between frames in milliseconds")
	convertCmd.Flags().BoolVarP(&debug, "debug", "", false, "Enable debug mode to show detailed progress")

	// Mark required flags
	convertCmd.MarkFlagRequired("input")
	convertCmd.MarkFlagRequired("output")
}
