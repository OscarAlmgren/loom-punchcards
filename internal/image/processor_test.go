package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		colorMode ColorMode
	}{
		{"2-color mode", 8, 26, TwoColor},
		{"4-color mode", 8, 26, FourColor},
		{"8-color mode", 8, 26, EightColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(tt.width, tt.height, tt.colorMode)
			if p.Width != tt.width {
				t.Errorf("Width = %d, want %d", p.Width, tt.width)
			}
			if p.Height != tt.height {
				t.Errorf("Height = %d, want %d", p.Height, tt.height)
			}
			if p.ColorMode != tt.colorMode {
				t.Errorf("ColorMode = %d, want %d", p.ColorMode, tt.colorMode)
			}
		})
	}
}

func TestValidateColorMode(t *testing.T) {
	tests := []struct {
		name      string
		mode      int
		wantError bool
	}{
		{"valid 2-color", 2, false},
		{"valid 4-color", 4, false},
		{"valid 8-color", 8, false},
		{"invalid 1-color", 1, true},
		{"invalid 3-color", 3, true},
		{"invalid 16-color", 16, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColorMode(tt.mode)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateColorMode() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRGBToGray(t *testing.T) {
	tests := []struct {
		name     string
		r, g, b  uint8
		expected uint8
	}{
		{"black", 0, 0, 0, 0},
		{"white", 255, 255, 255, 255},
		{"red", 255, 0, 0, 76},    // 0.299 * 255 ≈ 76
		{"green", 0, 255, 0, 150}, // 0.587 * 255 ≈ 150
		{"blue", 0, 0, 255, 29},   // 0.114 * 255 ≈ 29
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RGBToGray(tt.r, tt.g, tt.b)
			// Allow small margin of error due to rounding
			if abs(int(result)-int(tt.expected)) > 1 {
				t.Errorf("RGBToGray(%d, %d, %d) = %d, want %d", tt.r, tt.g, tt.b, result, tt.expected)
			}
		})
	}
}

func TestGetPixelIntensity(t *testing.T) {
	tests := []struct {
		name     string
		color    color.Color
		expected float64
	}{
		{"black", color.Black, 0.0},
		{"white", color.White, 1.0},
		{"gray", color.Gray{Y: 128}, 0.5019607843137255}, // 128/255
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPixelIntensity(tt.color)
			if abs(int(result*1000)-int(tt.expected*1000)) > 5 {
				t.Errorf("GetPixelIntensity() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	// Create a simple test image (8x8 checkerboard)
	img := createCheckerboardImage(16, 16, 2)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}

	// Process the image
	processor := NewProcessor(8, 8, TwoColor)
	matrix, err := processor.Process(&buf)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify matrix dimensions
	if len(matrix) != 8 {
		t.Errorf("Matrix height = %d, want 8", len(matrix))
	}
	if len(matrix[0]) != 8 {
		t.Errorf("Matrix width = %d, want 8", len(matrix[0]))
	}

	// Verify binary values
	for y := 0; y < len(matrix); y++ {
		for x := 0; x < len(matrix[y]); x++ {
			if matrix[y][x] != 0 && matrix[y][x] != 1 {
				t.Errorf("Matrix[%d][%d] = %d, want 0 or 1", y, x, matrix[y][x])
			}
		}
	}
}

func TestProcessInvalidImage(t *testing.T) {
	processor := NewProcessor(8, 8, TwoColor)

	// Test with invalid data
	invalidData := []byte("not an image")
	_, err := processor.Process(bytes.NewReader(invalidData))
	if err == nil {
		t.Error("Process() with invalid data should return error")
	}
}

func TestDescribeColorMode(t *testing.T) {
	tests := []struct {
		mode ColorMode
		want string
	}{
		{TwoColor, "2-color (binary: black/white using dithering)"},
		{FourColor, "4-color (4 grayscale levels using dithering patterns)"},
		{EightColor, "8-color (8 grayscale levels using dithering patterns)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			p := NewProcessor(8, 8, tt.mode)
			if got := p.DescribeColorMode(); got != tt.want {
				t.Errorf("DescribeColorMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetColorLevels(t *testing.T) {
	tests := []struct {
		mode ColorMode
		want int
	}{
		{TwoColor, 2},
		{FourColor, 4},
		{EightColor, 8},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.want)), func(t *testing.T) {
			p := NewProcessor(8, 8, tt.mode)
			if got := p.GetColorLevels(); got != tt.want {
				t.Errorf("GetColorLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToGrayscale(t *testing.T) {
	// Create a colored image
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})   // Red
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})   // Green
	img.Set(2, 0, color.RGBA{R: 0, G: 0, B: 255, A: 255})   // Blue
	img.Set(3, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // White

	gray := toGrayscale(img)

	// Verify it's grayscale
	if gray.Bounds() != img.Bounds() {
		t.Error("Grayscale image has different bounds")
	}

	// Check that the image is actually grayscale (simplified check)
	if gray.GrayAt(0, 0).Y == 0 {
		t.Error("Red should not convert to black")
	}
	if gray.GrayAt(3, 0).Y != 255 {
		t.Error("White should convert to white")
	}
}

func TestResize(t *testing.T) {
	// Create a test image
	img := image.NewGray(image.Rect(0, 0, 16, 16))

	// Resize to 8x8
	resized := resize(img, 8, 8)

	if resized.Bounds().Dx() != 8 {
		t.Errorf("Resized width = %d, want 8", resized.Bounds().Dx())
	}
	if resized.Bounds().Dy() != 8 {
		t.Errorf("Resized height = %d, want 8", resized.Bounds().Dy())
	}
}

// Helper functions

func createCheckerboardImage(width, height, squareSize int) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create checkerboard pattern
			if (x/squareSize+y/squareSize)%2 == 0 {
				img.SetGray(x, y, color.Gray{Y: 255}) // White
			} else {
				img.SetGray(x, y, color.Gray{Y: 0}) // Black
			}
		}
	}

	return img
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests

func BenchmarkProcess(b *testing.B) {
	img := createCheckerboardImage(100, 100, 10)
	var buf bytes.Buffer
	png.Encode(&buf, img)
	data := buf.Bytes()

	processor := NewProcessor(8, 26, TwoColor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.Process(bytes.NewReader(data))
	}
}

func BenchmarkDithering(b *testing.B) {
	processor := NewProcessor(8, 26, TwoColor)
	img := createCheckerboardImage(8, 26, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.applyDithering(img)
	}
}
