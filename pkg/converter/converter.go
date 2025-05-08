package converter

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jparrill/go-togif/pkg/ui"
	xdraw "golang.org/x/image/draw"
)

// ConvertPNGsToGIF converts a series of PNG images to a GIF
func ConvertPNGsToGIF(inputFiles []string, outputFile string, delay int, debug bool) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files specified")
	}

	// Validate delay
	if delay < 0 {
		return fmt.Errorf("delay must be non-negative")
	}

	// Create a channel for progress updates
	progressChan := ui.RunUI(debug, len(inputFiles))

	// First, read all images and get dimensions
	var firstImgBounds image.Rectangle
	var images []*image.Paletted
	var err error

	// Create a color map to store unique colors
	colorMap := make(map[color.Color]bool)
	var palette []color.Color

	// Process each image
	for i, inputFile := range inputFiles {
		// Update progress
		progressChan <- ui.ProgressMsg{
			CurrentFile: inputFile,
			Processed:   i,
			Total:       len(inputFiles),
		}

		// Open and decode the PNG file
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("error opening file %s: %v", inputFile, err)
		}
		defer file.Close()

		img, err := png.Decode(file)
		if err != nil {
			return fmt.Errorf("error decoding PNG file %s: %v", inputFile, err)
		}

		// If this is the first image, store its bounds
		if i == 0 {
			firstImgBounds = img.Bounds()
		}

		// Resize image if dimensions don't match
		if img.Bounds().Dx() != firstImgBounds.Dx() || img.Bounds().Dy() != firstImgBounds.Dy() {
			resized := image.NewRGBA(firstImgBounds)
			xdraw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), xdraw.Over, nil)
			img = resized
		}

		// Sample colors from the image
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				colorMap[img.At(x, y)] = true
			}
		}
	}

	// Convert color map to palette
	for c := range colorMap {
		palette = append(palette, c)
	}

	// Ensure we have at least one color in the palette
	if len(palette) == 0 {
		// Add basic colors if no colors were found
		palette = []color.Color{
			color.RGBA{0, 0, 0, 255},       // Black
			color.RGBA{255, 255, 255, 255}, // White
		}
	}

	// If we have too many colors, reduce the palette
	if len(palette) > 256 {
		// Sort colors by frequency
		colorFreq := make(map[color.Color]int)
		for _, inputFile := range inputFiles {
			file, err := os.Open(inputFile)
			if err != nil {
				return fmt.Errorf("error opening file %s: %v", inputFile, err)
			}
			defer file.Close()

			img, err := png.Decode(file)
			if err != nil {
				return fmt.Errorf("error decoding PNG file %s: %v", inputFile, err)
			}

			bounds := img.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					colorFreq[img.At(x, y)]++
				}
			}
		}

		// Sort colors by frequency
		type colorCount struct {
			color color.Color
			count int
		}
		var sortedColors []colorCount
		for c, count := range colorFreq {
			sortedColors = append(sortedColors, colorCount{c, count})
		}
		sort.Slice(sortedColors, func(i, j int) bool {
			return sortedColors[i].count > sortedColors[j].count
		})

		// Take the most frequent colors
		palette = make([]color.Color, 0, 256)
		for i := 0; i < len(sortedColors) && i < 256; i++ {
			palette = append(palette, sortedColors[i].color)
		}
	}

	if debug {
		fmt.Printf("Generated palette with %d colors\n", len(palette))
	}

	// Process each image again with the final palette
	for _, inputFile := range inputFiles {
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("error opening file %s: %v", inputFile, err)
		}
		defer file.Close()

		img, err := png.Decode(file)
		if err != nil {
			return fmt.Errorf("error decoding PNG file %s: %v", inputFile, err)
		}

		// Resize image if dimensions don't match
		if img.Bounds().Dx() != firstImgBounds.Dx() || img.Bounds().Dy() != firstImgBounds.Dy() {
			resized := image.NewRGBA(firstImgBounds)
			xdraw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), xdraw.Over, nil)
			img = resized
		}

		// Create a paletted image with our color palette
		paletted := image.NewPaletted(img.Bounds(), palette)
		xdraw.Draw(paletted, paletted.Bounds(), img, img.Bounds().Min, xdraw.Src)

		images = append(images, paletted)
	}

	// Create the output GIF
	outGif := &gif.GIF{
		Image: images,
		Delay: make([]int, len(images)),
	}

	// Set the same delay for all frames
	for i := range outGif.Delay {
		outGif.Delay[i] = delay / 10 // Convert to 100ths of a second
	}

	// Create the output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	// Get absolute path for the output file
	absOutputPath, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	// Update progress for final step
	progressChan <- ui.ProgressMsg{
		CurrentFile: "Creating output GIF",
		Processed:   len(inputFiles),
		Total:       len(inputFiles),
		OutputFile:  absOutputPath,
	}

	// Encode the GIF
	if err := gif.EncodeAll(outFile, outGif); err != nil {
		return fmt.Errorf("error encoding GIF: %v", err)
	}

	return nil
}

// ExpandInputPattern expands a glob pattern or regex into a list of matching PNG files
func ExpandInputPattern(pattern string) ([]string, error) {
	// Get the directory and base pattern
	dir := "."
	basePattern := pattern
	if strings.Contains(pattern, "/") {
		dir = filepath.Dir(pattern)
		basePattern = filepath.Base(pattern)
	}

	// Ensure the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	var matches []string

	// Try glob pattern first
	globMatches, err := filepath.Glob(filepath.Join(dir, basePattern))
	if err == nil && len(globMatches) > 0 {
		// Filter for PNG files
		for _, match := range globMatches {
			if strings.HasSuffix(strings.ToLower(match), ".png") {
				matches = append(matches, match)
			}
		}
		if len(matches) > 0 {
			sort.Strings(matches)
			return matches, nil
		}
	}

	// If glob pattern didn't work, try regex
	if strings.HasPrefix(basePattern, "^") || strings.ContainsAny(basePattern, ".*+?[](){}|") {
		re, err := regexp.Compile(basePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}

		// Read all files in the directory
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("error reading directory: %v", err)
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".png") {
				if re.MatchString(file.Name()) {
					matches = append(matches, filepath.Join(dir, file.Name()))
				}
			}
		}
		if len(matches) > 0 {
			sort.Strings(matches)
			return matches, nil
		}
	}

	// If no matches found, read the directory manually for simple patterns
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".png") {
			// For *.png pattern, match all PNG files
			if basePattern == "*.png" {
				matches = append(matches, filepath.Join(dir, file.Name()))
				continue
			}

			// For other patterns, try to match the filename
			matched, err := filepath.Match(basePattern, file.Name())
			if err == nil && matched {
				matches = append(matches, filepath.Join(dir, file.Name()))
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no PNG files found matching pattern: %s", pattern)
	}

	// Sort matches for consistent ordering
	sort.Strings(matches)
	return matches, nil
}

// ValidateInputFiles checks if all input files exist and are PNGs
func ValidateInputFiles(inputFiles []string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files specified")
	}

	for _, file := range inputFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return err
		}
		if !strings.HasSuffix(strings.ToLower(file), ".png") {
			return fmt.Errorf("file %s is not a PNG", file)
		}
	}
	return nil
}
