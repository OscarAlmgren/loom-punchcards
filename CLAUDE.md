# CLAUDE.md - Jacquard Loom Punchcard Generator

## Project Overview

This is a full-stack web application that converts digital images into Jacquard loom punchcard patterns. Inspired by Joseph Marie Jacquard's 1801 invention that revolutionized textile manufacturing, this tool allows weavers to create historical-style punchcards from modern digital images.

**Technology Stack:**
- Backend: Go (HTTP server, image processing, punchcard generation)
- Frontend: HTML, CSS, JavaScript
- Image Processing: Floyd-Steinberg dithering for quality conversion
- Output Formats: SVG (with PDF planned)

## Architecture

### Project Structure

```
loom-punchcards/
├── cmd/
│   └── server/
│       └── main.go              # HTTP server entry point
├── internal/
│   ├── handler/
│   │   └── handler.go           # HTTP request handlers
│   ├── image/
│   │   ├── processor.go         # Image processing and dithering
│   │   └── processor_test.go
│   └── punchcard/
│       ├── generator.go         # Punchcard generation logic
│       ├── svg_exporter.go      # SVG export functionality
│       ├── generator_test.go
│       └── svg_exporter_test.go
├── web/
│   ├── static/
│   │   └── style.css
│   └── templates/
│       └── index.html
└── go.mod
```

### Key Components

#### 1. Image Processor (`internal/image/processor.go`)

Handles image conversion with the following pipeline:
1. **Decode**: Load image from uploaded file
2. **Grayscale Conversion**: Convert to grayscale using standard luminosity formula
3. **Resize**: Resize to 8 columns (CardWidth) with auto-calculated height based on aspect ratio
4. **Dithering**: Apply Floyd-Steinberg dithering for visual quality with limited color levels
5. **Binary Conversion**: Convert to binary matrix (1 = punch hole, 0 = no hole)

**Color Modes:**
- `2-color`: Binary black/white using dithering
- `4-color`: 4 grayscale levels using dithering patterns
- `8-color`: 8 grayscale levels using dithering patterns

**Important Design Decision:**
- Width is fixed at 8 (CardWidth - number of needles in a simplified home loom)
- Height is auto-calculated from image aspect ratio (set to 0 in processor initialization)
- The resize function handles height=0 by calculating: `height = width * (srcHeight / srcWidth)`

#### 2. Punchcard Generator (`internal/punchcard/generator.go`)

Converts binary matrices into punchcard sequences:
- **CardWidth**: 8 columns (represents 8 needles)
- **CardHeight**: 26 rows per card (standard punchcard size)
- Splits long images into multiple sequential cards
- Validates that matrix width matches CardWidth (8)
- Pads partial cards with zeros (no holes)

#### 3. HTTP Handlers (`internal/handler/handler.go`)

**Endpoints:**
- `GET /` - Main page with upload interface
- `POST /upload` - Process image and download punchcard file
- `POST /preview` - Generate preview (first 3 cards only)
- `POST /info` - Get metadata about generated punchcards
- `GET /health` - Health check endpoint
- `GET /static/*` - Serve static assets

**Safety Features:**
- Matrix empty check before accessing elements (prevents panics)
- File size limit: 10MB
- Validates color mode (2, 4, or 8)
- Validates output format (svg or pdf)

#### 4. SVG Exporter (`internal/punchcard/svg_exporter.go`)

Generates SVG files with:
- Visual representation of punchcard holes
- Card numbering and metadata
- Professional styling for printing
- Scalable vector format suitable for manufacturing

## Running the Application

### Development

```bash
# Run the server
go run ./cmd/server/main.go

# Server starts on http://localhost:8080
```

### Build

```bash
# Build the binary
go build -o punchcard-server ./cmd/server/main.go

# Run the binary
./punchcard-server
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Usage Flow

1. **Upload Image**: User uploads an image (PNG, JPEG supported)
2. **Select Options**:
   - Color Mode: 2, 4, or 8 color levels
   - Format: SVG or PDF
3. **Preview** (optional): View first 3 cards before downloading
4. **Get Info** (optional): See metadata (total cards, dimensions, density)
5. **Download**: Generate and download complete punchcard file

## Common Issues and Solutions

### Issue: Upload Panic (Index Out of Range)

**Symptoms**:
```
runtime error: index out of range [0] with length 0
```

**Root Cause**:
The resize function received height=0 but didn't implement auto-calculation, resulting in an empty matrix.

**Solution** (Fixed in commit `264c53f`):
- Resize function now auto-calculates height from aspect ratio when height=0
- Added safety checks to prevent accessing empty matrices
- All handlers validate matrix before use

### Issue: Image Width Doesn't Match CardWidth

**Symptoms**:
```
image width (X) does not match card width (8)
```

**Root Cause**:
After resizing, the image should be exactly 8 pixels wide, but something went wrong in the resize process.

**Solution**:
The resize function is supposed to force the width to exactly CardWidth (8). Check that the resize logic is working correctly.

## Development Guidelines

### Adding New Features

1. **New Color Modes**:
   - Update `ValidateColorMode()` in `internal/image/processor.go`
   - Add new constants to ColorMode type
   - Update dithering logic in `applyDithering()`

2. **New Export Formats**:
   - Create new exporter (e.g., `pdf_exporter.go`) implementing Exporter interface
   - Update handler to support new format
   - Add format validation

3. **New Card Dimensions**:
   - Modify `CardWidth` and `CardHeight` constants in `internal/punchcard/generator.go`
   - Update tests to reflect new dimensions
   - Consider making dimensions configurable

### Code Quality

- All public functions must have documentation comments
- Write tests for new functionality
- Use `go fmt` before committing
- Run `go vet` to catch common mistakes
- Maintain test coverage above 80%

### Error Handling

- Always validate user input
- Log errors with context (e.g., filename, dimensions)
- Return user-friendly error messages
- Never panic in handlers - use safety checks

## Key Algorithms

### Floyd-Steinberg Dithering

The image processor uses Floyd-Steinberg dithering to create visually pleasing patterns with limited color levels:

```
Current pixel error is distributed to neighbors:
    *   7/16
3/16 5/16 1/16
```

This creates the illusion of more colors through spatial patterns.

### Aspect Ratio Preservation

When height=0 is specified:
```go
aspectRatio := float64(srcHeight) / float64(srcWidth)
height = int(float64(width) * aspectRatio)
```

This ensures the output maintains the visual proportions of the input image.

### Card Splitting

Long images are split into 26-row cards:
```go
numCards = (imageHeight + CardHeight - 1) / CardHeight  // Ceiling division
```

Partial cards are padded with zeros (no holes).

## Future Improvements

### Planned Features

1. **PDF Export**: Real PDF generation (currently exports SVG with PDF content-type)
2. **Image Preview**: Show before/after comparison
3. **Batch Processing**: Upload multiple images at once
4. **Custom Dimensions**: Allow users to specify card width/height
5. **Pattern Library**: Save and reuse common patterns
6. **Print Optimization**: Add printer-friendly layouts with cut lines
7. **Color Mapping**: Support for color-coded threads

### Technical Debt

1. **PDF Export**: Currently returns SVG with PDF mime type - needs proper PDF library
2. **Frontend**: Could benefit from modern framework (React/Vue/Svelte)
3. **Image Validation**: Add more robust image format validation
4. **Configuration**: Make CardWidth/CardHeight configurable via env vars or config file
5. **Async Processing**: For large images, consider background processing with progress updates

## Historical Context

Jacquard looms use punchcards to control which warp threads are raised for each pass of the weft thread. Each hole in the card corresponds to a thread being raised, creating the pattern.

- **Hole punched (1)**: Thread raised (creates dark pixel in fabric)
- **No hole (0)**: Thread lowered (creates light pixel in fabric)

This binary system was an early form of programming and directly inspired Charles Babbage's Analytical Engine and Herman Hollerith's punched card machines.

## Contributing

When contributing to this project:

1. Create a feature branch: `git checkout -b feature/your-feature-name`
2. Write tests for new functionality
3. Ensure all tests pass: `go test ./...`
4. Update documentation if needed
5. Create a pull request with clear description

## Resources

- [Jacquard Loom History](https://en.wikipedia.org/wiki/Jacquard_machine)
- [Floyd-Steinberg Dithering](https://en.wikipedia.org/wiki/Floyd%E2%80%93Steinberg_dithering)
- [Go Image Package Documentation](https://pkg.go.dev/image)

## License

See LICENSE file for details.

---

**For Claude Code Users**: This file serves as context for understanding the project architecture, common issues, and development practices. When making changes, refer to this document to ensure consistency with the project's design philosophy.
