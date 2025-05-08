# go-togif

A high-quality PNG to GIF converter CLI tool written in Go.

## Features

- Converts multiple PNG images to a single GIF
- Maintains original image quality and dimensions
- Configurable frame delay
- Cross-platform support
- Simple and intuitive CLI interface
- Support for glob patterns and regular expressions for input files

## Installation

### From Binary

Download the latest release from the [releases page](https://github.com/jparrill/go-togif/releases) and extract the binary to your PATH.

### From Source

```bash
# Clone the repository
git clone https://github.com/jparrill/go-togif.git
cd go-togif

# Install dependencies
make deps

# Build the binary
make build
```

## Usage

```bash
# Basic usage with specific files
go-togif convert -i image1.png -i image2.png -o output.gif

# Using glob pattern
go-togif convert -i "*.png" -o output.gif

# Using regex pattern
go-togif convert -i "^frame.*\.png$" -o output.gif

# With custom delay (in milliseconds)
go-togif convert -i "*.png" -o output.gif -d 200

# Get help
go-togif --help
```

### Input Patterns

The tool supports two types of patterns for input files:

1. **Glob Patterns**: Traditional file matching patterns like `*.png` or `images/*.png`
2. **Regular Expressions**: Full regex support for complex matching patterns
   - Must start with `^` or contain regex special characters (`.`, `*`, `+`, `?`, etc.)
   - Example: `^frame[0-9]+\.png$` matches files like `frame1.png`, `frame2.png`, etc.

### Flags

- `-i, --input`: Input PNG files or patterns (can be specified multiple times)
- `-o, --output`: Output GIF file path (default: "output.gif")
- `-d, --delay`: Delay between frames in milliseconds (default: 100)

## Development

### Prerequisites

- Go 1.16 or later
- Make
- Goreleaser (for releases)

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Creating a Release

```bash
# Create a snapshot release
make release-snapshot

# Create a real release
make release
```

## AI Assistance

This project was developed with the assistance of an AI coding companion (Claude 3.5 Sonnet).

## License

MIT License - see LICENSE file for details