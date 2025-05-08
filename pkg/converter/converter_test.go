package converter

import (
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
