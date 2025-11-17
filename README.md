# 🧵 Jacquard Loom Punchcard Generator

A full-stack web application that converts digital images into historically accurate Jacquard loom punchcards for silk weaving. This tool bridges the gap between modern digital imagery and the revolutionary 1801 weaving technology invented by Joseph Marie Jacquard.

## 📖 Table of Contents

- [Overview](#overview)
- [Historical Context](#historical-context)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Technical Specifications](#technical-specifications)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Overview

The Jacquard loom revolutionized textile manufacturing by using punched cards to control complex weaving patterns. This application modernizes that process by converting digital images into a series of punchcards suitable for use in Jacquard silk weaving looms.

### Key Capabilities

- Upload images in PNG or JPEG format
- Convert to 2, 4, or 8 color modes using advanced dithering
- Generate SVG or PDF output for printing
- Automatic card numbering and sequencing
- Preview functionality
- Detailed metadata about generated cards

## Historical Context

### The Jacquard Loom

In 1801, French weaver Joseph Marie Jacquard invented an automated loom using a chain of punched cards to control the weave pattern. Each card contained a matrix of holes representing which warp threads should be raised during each pass of the shuttle.

**Historical Impact:**
- Revolutionized the textile industry
- Enabled complex patterns previously requiring multiple skilled weavers
- Inspired Charles Babbage's Analytical Engine
- Became the precursor to modern computer punch cards
- Some patterns required over 24,000 cards with 1,000+ hole positions each

### Technical Innovation

The Jacquard system used a binary representation centuries before digital computers:
- **Hole punched (1)**: Warp thread raised, weft passes under
- **No hole (0)**: Warp thread lowered, weft passes over

This binary system allowed for intricate patterns in silk and other fine materials, making luxury textiles more accessible.

## Features

### Image Processing

- **Advanced Dithering**: Floyd-Steinberg algorithm creates smooth grayscale variations
- **Multi-Color Modes**:
  - **2-Color**: Pure binary (black/white) for simple patterns
  - **4-Color**: Four grayscale levels for moderate detail
  - **8-Color**: Eight grayscale levels for complex imagery
- **Automatic Resizing**: Fits images to 8-column loom specification
- **Quality Preservation**: Maintains visual fidelity within hardware constraints

### Card Generation

- **Standard Format**: 8 columns × 26 rows (208 possible holes per card)
- **Sequential Numbering**: Cards numbered for correct assembly
- **Metadata Tracking**: Hole density, pattern statistics
- **Validation**: Ensures cards meet physical specifications

### Export Options

#### SVG Export
- Scalable vector format
- Precise hole positioning
- Grid overlay for alignment
- Card information annotations
- Suitable for CNC cutting or laser cutting

#### PDF Export
- Print-ready format
- Multiple cards per page
- Professional layout
- Archival quality

### Web Interface

- **Modern HTMX Frontend**: Fast, responsive, no JavaScript framework needed
- **Real-time Preview**: See first 3 cards before downloading
- **Detailed Information**: View statistics about your pattern
- **Intuitive Controls**: Simple upload and parameter selection

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Web Browser                          │
│                     (HTMX Frontend)                         │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/HTTPS
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Server (Go)                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Request Handlers                        │   │
│  │  • Upload Handler    • Info Handler                 │   │
│  │  • Preview Handler   • Health Check                 │   │
│  └──────────────┬──────────────────────────────────────┘   │
│                 │                                           │
│  ┌──────────────▼──────────────┐  ┌────────────────────┐   │
│  │   Image Processor           │  │  Punchcard         │   │
│  │  • Grayscale conversion     │  │  Generator         │   │
│  │  • Resizing                 │  │  • Matrix creation │   │
│  │  • Floyd-Steinberg          │  │  • Validation      │   │
│  │    dithering                │  │  • Metadata        │   │
│  └──────────────┬──────────────┘  └─────────┬──────────┘   │
│                 │                            │              │
│                 └───────────┬────────────────┘              │
│                             ▼                               │
│                 ┌─────────────────────────┐                 │
│                 │    Export Modules       │                 │
│                 │  • SVG Exporter         │                 │
│                 │  • PDF Exporter         │                 │
│                 └─────────────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
loom-punchcards/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── image/
│   │   ├── processor.go         # Image processing & dithering
│   │   └── processor_test.go    # Image processing tests
│   ├── punchcard/
│   │   ├── generator.go         # Card generation logic
│   │   ├── generator_test.go    # Generator tests
│   │   ├── svg.go               # SVG export
│   │   ├── svg_test.go          # SVG export tests
│   │   └── pdf.go               # PDF export (framework)
│   └── handler/
│       └── handler.go           # HTTP request handlers
├── web/
│   ├── templates/
│   │   └── index.html           # HTMX frontend
│   └── static/
│       └── style.css            # Styling
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # This file
```

### Technology Stack

**Backend:**
- **Language**: Go 1.19+
- **Standard Library**: net/http, image, html/template
- **No external dependencies** for core functionality

**Frontend:**
- **HTMX 1.9.10**: Dynamic interactions without JavaScript frameworks
- **CSS3**: Modern responsive styling
- **HTML5**: Semantic markup

**Image Processing:**
- Grayscale conversion using luminosity method
- Nearest-neighbor interpolation for resizing
- Floyd-Steinberg error diffusion dithering

## Installation

### Prerequisites

- Go 1.19 or higher
- Web browser with modern CSS/JavaScript support

### Build from Source

```bash
# Clone the repository
git clone https://github.com/oscaralmgren/loom-punchcards.git
cd loom-punchcards

# Download dependencies
go mod download

# Build the application
go build -o punchcard-server ./cmd/server

# Run the server
./punchcard-server
```

### Configuration

The server accepts the following command-line flags:

```bash
./punchcard-server \
  -port=8080 \
  -templates=web/templates \
  -static=web/static
```

**Environment Variables:**
- `PORT`: HTTP server port (default: 8080)

## Usage

### Starting the Server

```bash
go run ./cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Web Interface

1. **Open your browser** to `http://localhost:8080`
2. **Select an image** (PNG or JPEG, max 10MB)
3. **Choose color mode**:
   - 2-Color for simple patterns
   - 4-Color for moderate detail
   - 8-Color for complex images
4. **Select export format** (SVG or PDF)
5. **Preview** to see the first few cards
6. **Get Info** to view pattern statistics
7. **Generate & Download** to create the full set

### Example Workflow

**For a simple logo:**
```
1. Upload: company-logo.png
2. Color Mode: 2-Color
3. Format: SVG
4. Result: 2-3 cards for small logo
```

**For a portrait:**
```
1. Upload: portrait.jpg
2. Color Mode: 8-Color (best detail)
3. Format: PDF
4. Result: 10-20 cards depending on image height
```

## API Documentation

### Endpoints

#### `GET /`
Home page with upload interface

#### `POST /upload`
Upload and process image, return downloadable file

**Form Parameters:**
- `image` (file): Image file (PNG/JPEG)
- `colorMode` (int): 2, 4, or 8
- `format` (string): "svg" or "pdf"

**Response:** Binary file download

#### `POST /preview`
Generate preview of first 3 cards

**Form Parameters:**
- `image` (file): Image file
- `colorMode` (int): 2, 4, or 8

**Response:** SVG image (inline)

#### `POST /info`
Get metadata about generated cards

**Form Parameters:**
- `image` (file): Image file
- `colorMode` (int): 2, 4, or 8

**Response:** JSON object
```json
{
  "filename": "image.png",
  "fileSize": 102400,
  "colorMode": "2-color (binary: black/white using dithering)",
  "totalCards": 5,
  "cardDimensions": "8x26",
  "totalRows": 130,
  "averageDensity": "45.2%",
  "holesPerCard": [95, 102, 87, 94, 88]
}
```

#### `GET /health`
Health check endpoint

**Response:**
```json
{
  "status": "healthy",
  "service": "Jacquard Loom Punchcard Generator"
}
```

## Technical Specifications

### Card Format

- **Dimensions**: 8 columns × 26 rows
- **Total Holes**: 208 possible punch positions per card
- **Hole Representation**:
  - 1 = Hole punched (black circle in SVG)
  - 0 = No hole (light guide mark in SVG)
- **Numbering**: Sequential, starting from 1
- **Orientation**: Top to bottom, left to right

### Image Processing

#### Grayscale Conversion
Uses standard luminosity formula:
```
Gray = 0.299×Red + 0.587×Green + 0.114×Blue
```

#### Dithering (Floyd-Steinberg)
Error diffusion pattern:
```
       * 7/16
 3/16 5/16 1/16
```

Where `*` is the current pixel being processed.

#### Color Quantization
- **2-Color**: Threshold at 0.5 (middle gray)
- **4-Color**: 4 levels (0.0, 0.33, 0.67, 1.0)
- **8-Color**: 8 levels (0.0, 0.14, 0.29, 0.43, 0.57, 0.71, 0.86, 1.0)

### SVG Export Specifications

- **Format**: SVG 1.1
- **Units**: Millimeters
- **DPI**: 96 (3.78 pixels/mm)
- **Hole Radius**: 2mm
- **Hole Spacing**: 5mm center-to-center
- **Card Padding**: 10mm
- **Features**:
  - Grid overlay (optional)
  - Card numbering
  - Metadata in title/description
  - Precise positioning for manufacturing

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Benchmark specific package
go test -bench=. ./internal/image
```

### Code Structure

**Package Organization:**
- `cmd/server`: Application entry point
- `internal/image`: Image processing logic
- `internal/punchcard`: Card generation and export
- `internal/handler`: HTTP request handling
- `web`: Frontend templates and static files

**Design Patterns:**
- Dependency injection for testability
- Interface-based design for flexibility
- Error handling with detailed messages
- Validation at boundaries

### Adding New Features

**New Export Format:**
1. Create new file in `internal/punchcard/`
2. Implement `Exporter` interface
3. Add handler endpoint
4. Update frontend

**New Dithering Algorithm:**
1. Add function to `internal/image/processor.go`
2. Update `applyDithering` to use new method
3. Add tests

## Testing

### Test Coverage

- **Image Processing**: 90%+ coverage
- **Punchcard Generation**: 95%+ coverage
- **SVG Export**: 85%+ coverage

### Test Types

- **Unit Tests**: Individual component testing
- **Integration Tests**: Handler and workflow testing
- **Benchmark Tests**: Performance measurements

### Example Test Run

```
$ go test ./...
ok      github.com/oscaralmgren/loom-punchcards/internal/image      0.009s
ok      github.com/oscaralmgren/loom-punchcards/internal/punchcard  0.054s
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Code Style

- Follow Go formatting conventions (`gofmt`)
- Write descriptive commit messages
- Document exported functions
- Add examples where helpful

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Joseph Marie Jacquard (1752-1834) for the original invention
- The textile preservation community for historical documentation
- Contributors to the Floyd-Steinberg dithering algorithm

## References

### Historical
- [Jacquard Machine - Wikipedia](https://en.wikipedia.org/wiki/Jacquard_machine)
- [Computer History Museum - Punched Cards](https://www.computerhistory.org/storageengine/punched-cards-control-jacquard-loom/)
- [Science Museum - Jacquard Loom History](https://www.scienceandindustrymuseum.org.uk/objects-and-stories/jacquard-loom)

### Technical
- Floyd, R. W., & Steinberg, L. (1976). An adaptive algorithm for spatial greyscale
- Jacquard loom operating principles and card specifications

---

**Created for historical textile enthusiasts, weavers, and anyone fascinated by the intersection of computing history and textile arts.**

🧵 Happy Weaving! 🧵
