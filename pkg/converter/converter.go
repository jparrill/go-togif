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

	// Create a channel for progress updates
	progressChan := ui.RunUI(debug, len(inputFiles))

	// First, read all images and get dimensions
	var firstImgBounds image.Rectangle
	var images []*image.Paletted
	var err error

	// Create a basic color palette
	palette := []color.Color{
		color.RGBA{0, 0, 0, 255},       // Black
		color.RGBA{255, 255, 255, 255}, // White
		color.RGBA{255, 0, 0, 255},     // Red
		color.RGBA{0, 255, 0, 255},     // Green
		color.RGBA{0, 0, 255, 255},     // Blue
		color.RGBA{255, 255, 0, 255},   // Yellow
		color.RGBA{255, 0, 255, 255},   // Magenta
		color.RGBA{0, 255, 255, 255},   // Cyan
		color.RGBA{128, 128, 128, 255}, // Gray
	}

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

	// Update progress for final step
	progressChan <- ui.ProgressMsg{
		CurrentFile: "Creating output GIF",
		Processed:   len(inputFiles),
		Total:       len(inputFiles),
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
