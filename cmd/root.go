package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-togif",
	Short: "Convert PNG images to GIF with high quality",
	Long: `A CLI application that converts a series of PNG images into a high-quality GIF.
The output GIF will maintain the same quality and dimensions as the input images.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "output.gif", "Output GIF file path")
	rootCmd.PersistentFlags().IntP("delay", "d", 100, "Delay between frames in milliseconds")
	rootCmd.PersistentFlags().StringSliceP("input", "i", []string{}, "Input PNG files (can be specified multiple times)")
}
