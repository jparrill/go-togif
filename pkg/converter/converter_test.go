package converter

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandInputPattern(t *testing.T) {
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
		"other.txt",
		"frame1.png",
		"frame2.png",
		"frame3.png",
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	tests := []struct {
		name    string
		pattern string
		dir     string
		want    int
		wantErr bool
	}{
		{
			name:    "Glob pattern all PNGs",
			pattern: "*.png",
			dir:     tempDir,
			want:    5,
			wantErr: false,
		},
		{
			name:    "Glob pattern specific files",
			pattern: "test*.png",
			dir:     tempDir,
			want:    2,
			wantErr: false,
		},
		{
			name:    "Regex pattern frames",
			pattern: "frame[0-9]+\\.png",
			dir:     tempDir,
			want:    3,
			wantErr: false,
		},
		{
			name:    "Non-existent pattern",
			pattern: "nonexistent.png",
			dir:     tempDir,
			want:    0,
			wantErr: true,
		},
		{
			name:    "Invalid directory",
			pattern: "*.png",
			dir:     "nonexistent",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := filepath.Join(tt.dir, tt.pattern)
			got, err := ExpandInputPattern(pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandInputPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("ExpandInputPattern() got %d files, want %d", len(got), tt.want)
			}
		})
	}
}

func TestValidateInputFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "go-togif-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	validPNG := filepath.Join(tempDir, "valid.png")
	invalidExt := filepath.Join(tempDir, "invalid.txt")
	nonexistent := filepath.Join(tempDir, "nonexistent.png")

	// Create a valid PNG file
	f, err := os.Create(validPNG)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()

	tests := []struct {
		name    string
		files   []string
		wantErr bool
	}{
		{
			name:    "Valid PNG files",
			files:   []string{validPNG},
			wantErr: false,
		},
		{
			name:    "Invalid extension",
			files:   []string{invalidExt},
			wantErr: true,
		},
		{
			name:    "Nonexistent file",
			files:   []string{nonexistent},
			wantErr: true,
		},
		{
			name:    "Empty file list",
			files:   []string{},
			wantErr: true,
		},
		{
			name:    "Mixed valid and invalid",
			files:   []string{validPNG, invalidExt},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputFiles(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInputFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConvertPNGsToGIF(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "go-togif-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test PNG files with different colors
	testFiles := []string{
		"test1.png",
		"test2.png",
		"test3.png",
	}

	// Create test images with different colors
	colors := []color.RGBA{
		{255, 0, 0, 255},   // Red
		{0, 255, 0, 255},   // Green
		{0, 0, 255, 255},   // Blue
		{255, 255, 0, 255}, // Yellow
		{255, 0, 255, 255}, // Magenta
		{0, 255, 255, 255}, // Cyan
	}

	for _, file := range testFiles {
		// Create a new image with specific colors
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				// Use different colors for different regions
				colorIndex := (x + y) % len(colors)
				img.Set(x, y, colors[colorIndex])
			}
		}

		// Save the image
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("Failed to encode test image %s: %v", file, err)
		}
		f.Close()
	}

	// Test cases
	tests := []struct {
		name     string
		inputDir string
		output   string
		delay    int
		debug    bool
		wantErr  bool
		checkGIF bool
	}{
		{
			name:     "Basic conversion",
			inputDir: tempDir,
			output:   filepath.Join(tempDir, "output.gif"),
			delay:    100,
			debug:    false,
			wantErr:  false,
			checkGIF: true,
		},
		{
			name:     "Debug mode",
			inputDir: tempDir,
			output:   filepath.Join(tempDir, "output_debug.gif"),
			delay:    200,
			debug:    true,
			wantErr:  false,
			checkGIF: true,
		},
		{
			name:     "Invalid delay",
			inputDir: tempDir,
			output:   filepath.Join(tempDir, "output_invalid.gif"),
			delay:    -1,
			debug:    false,
			wantErr:  true,
			checkGIF: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get all PNG files in the input directory
			inputFiles, err := filepath.Glob(filepath.Join(tt.inputDir, "*.png"))
			if err != nil {
				t.Fatalf("Failed to glob input files: %v", err)
			}

			// Convert images
			err = ConvertPNGsToGIF(inputFiles, tt.output, tt.delay, tt.debug)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertPNGsToGIF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect success, verify the output GIF
			if !tt.wantErr && tt.checkGIF {
				// Check if output file exists
				if _, err := os.Stat(tt.output); os.IsNotExist(err) {
					t.Errorf("Output file %s was not created", tt.output)
					return
				}

				// Open and decode the GIF
				f, err := os.Open(tt.output)
				if err != nil {
					t.Errorf("Failed to open output file: %v", err)
					return
				}
				defer f.Close()

				gifImg, err := gif.DecodeAll(f)
				if err != nil {
					t.Errorf("Failed to decode output GIF: %v", err)
					return
				}

				// Verify GIF properties
				if len(gifImg.Image) != len(inputFiles) {
					t.Errorf("GIF has %d frames, want %d", len(gifImg.Image), len(inputFiles))
				}

				// Check that each frame has a valid palette
				for i, frame := range gifImg.Image {
					if len(frame.Palette) == 0 {
						t.Errorf("Frame %d has empty palette", i)
					}
					if len(frame.Palette) > 256 {
						t.Errorf("Frame %d has too many colors: %d", i, len(frame.Palette))
					}
				}

				// Check delay values
				for i, delay := range gifImg.Delay {
					expectedDelay := tt.delay / 10 // Convert to 100ths of a second
					if delay != expectedDelay {
						t.Errorf("Frame %d has delay %d, want %d", i, delay, expectedDelay)
					}
				}
			}
		})
	}
}

func TestPaletteGeneration(t *testing.T) {
	// Create a test image with specific colors
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testColors := []color.RGBA{
		{255, 0, 0, 255},   // Red
		{0, 255, 0, 255},   // Green
		{0, 0, 255, 255},   // Blue
		{255, 255, 0, 255}, // Yellow
		{255, 0, 255, 255}, // Magenta
		{0, 255, 255, 255}, // Cyan
	}

	// Fill the image with test colors
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			colorIndex := (x + y) % len(testColors)
			img.Set(x, y, testColors[colorIndex])
		}
	}

	// Create a temporary file
	tempDir, err := os.MkdirTemp("", "go-togif-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.png")
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("Failed to encode test image: %v", err)
	}
	f.Close()

	// Test conversion with debug mode to see palette size
	err = ConvertPNGsToGIF([]string{testFile}, filepath.Join(tempDir, "output.gif"), 100, true)
	if err != nil {
		t.Fatalf("Failed to convert image: %v", err)
	}

	// Verify the output GIF
	f, err = os.Open(filepath.Join(tempDir, "output.gif"))
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer f.Close()

	gifImg, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatalf("Failed to decode output GIF: %v", err)
	}

	// Check that the palette contains all test colors
	if len(gifImg.Image) == 0 {
		t.Fatal("GIF has no frames")
	}

	palette := gifImg.Image[0].Palette
	if len(palette) == 0 {
		t.Fatal("GIF has empty palette")
	}

	// Verify that all test colors are in the palette
	colorMap := make(map[color.Color]bool)
	for _, c := range palette {
		colorMap[c] = true
	}

	for _, testColor := range testColors {
		found := false
		for c := range colorMap {
			r1, g1, b1, a1 := testColor.RGBA()
			r2, g2, b2, a2 := c.RGBA()
			if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Test color %v not found in palette", testColor)
		}
	}
}
