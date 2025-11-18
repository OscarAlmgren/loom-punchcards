package image

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
)

// ColorMode defines the number of color variations supported
type ColorMode int

const (
	TwoColor   ColorMode = 2  // Binary: black/white
	FourColor  ColorMode = 4  // 4 grayscale levels
	EightColor ColorMode = 8  // 8 grayscale levels
)

// Processor handles image processing for punchcard conversion
type Processor struct {
	Width     int
	Height    int
	ColorMode ColorMode
}

// NewProcessor creates a new image processor
// For Jacquard looms: width typically represents the number of needles (8 for simplified version)
// height represents the number of rows in the image
func NewProcessor(width, height int, mode ColorMode) *Processor {
	return &Processor{
		Width:     width,
		Height:    height,
		ColorMode: mode,
	}
}

// Process converts an uploaded image to a binary matrix suitable for punchcard generation
// Uses Floyd-Steinberg dithering for better visual quality with limited colors
func (p *Processor) Process(r io.Reader) ([][]int, error) {
	// Decode the image
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Convert to grayscale and resize
	grayImg := toGrayscale(img)
	resized := resize(grayImg, p.Width, p.Height)

	// Apply dithering based on color mode
	dithered := p.applyDithering(resized)

	return dithered, nil
}

// toGrayscale converts an image to grayscale
func toGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	return gray
}

// resize uses nearest-neighbor interpolation to resize the image
func resize(img *image.Gray, width, height int) *image.Gray {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// If height is 0, calculate it based on aspect ratio
	if height == 0 && width > 0 {
		aspectRatio := float64(srcHeight) / float64(srcWidth)
		height = int(float64(width) * aspectRatio)
		if height == 0 {
			height = 1 // Ensure at least 1 row
		}
	}

	// If width is 0, calculate it based on aspect ratio
	if width == 0 && height > 0 {
		aspectRatio := float64(srcWidth) / float64(srcHeight)
		width = int(float64(height) * aspectRatio)
		if width == 0 {
			width = 1 // Ensure at least 1 column
		}
	}

	// Safety check: ensure both dimensions are positive
	if width <= 0 || height <= 0 {
		// Return a minimal 1x1 image if dimensions are invalid
		dst := image.NewGray(image.Rect(0, 0, 1, 1))
		dst.Set(0, 0, img.At(0, 0))
		return dst
	}

	dst := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map destination coordinates to source coordinates
			srcX := x * srcWidth / width
			srcY := y * srcHeight / height

			dst.Set(x, y, img.At(srcX, srcY))
		}
	}

	return dst
}

// applyDithering applies Floyd-Steinberg dithering to create visual patterns
// with limited color levels, mimicking old-school pixel art techniques
func (p *Processor) applyDithering(img *image.Gray) [][]int {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a copy of the image data for dithering
	pixels := make([][]float64, height)
	for y := 0; y < height; y++ {
		pixels[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			gray := img.GrayAt(x, y)
			// Normalize to 0-1 range
			pixels[y][x] = float64(gray.Y) / 255.0
		}
	}

	// Determine the number of levels based on color mode
	levels := int(p.ColorMode)

	// Apply Floyd-Steinberg dithering
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			oldPixel := pixels[y][x]

			// Find the nearest color level
			newPixel := math.Round(oldPixel*float64(levels-1)) / float64(levels-1)
			pixels[y][x] = newPixel

			// Calculate quantization error
			err := oldPixel - newPixel

			// Distribute error to neighboring pixels (Floyd-Steinberg)
			if x+1 < width {
				pixels[y][x+1] += err * 7.0 / 16.0
			}
			if y+1 < height {
				if x > 0 {
					pixels[y+1][x-1] += err * 3.0 / 16.0
				}
				pixels[y+1][x] += err * 5.0 / 16.0
				if x+1 < width {
					pixels[y+1][x+1] += err * 1.0 / 16.0
				}
			}
		}
	}

	// Convert to binary matrix
	// In Jacquard weaving: 1 = hole punched (thread raised), 0 = no hole (thread lowered)
	// We'll map darker pixels to 1 (punch) and lighter pixels to 0 (no punch)
	result := make([][]int, height)
	threshold := 0.5 // Middle gray as threshold

	for y := 0; y < height; y++ {
		result[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if pixels[y][x] < threshold {
				result[y][x] = 1 // Dark = punch hole
			} else {
				result[y][x] = 0 // Light = no punch
			}
		}
	}

	return result
}

// GetColorLevels returns the number of distinct visual levels achievable
// by combining dithering patterns
func (p *Processor) GetColorLevels() int {
	return int(p.ColorMode)
}

// DescribeColorMode returns a human-readable description of the color mode
func (p *Processor) DescribeColorMode() string {
	switch p.ColorMode {
	case TwoColor:
		return "2-color (binary: black/white using dithering)"
	case FourColor:
		return "4-color (4 grayscale levels using dithering patterns)"
	case EightColor:
		return "8-color (8 grayscale levels using dithering patterns)"
	default:
		return fmt.Sprintf("%d-color mode", p.ColorMode)
	}
}

// ValidateColorMode checks if the color mode is valid
func ValidateColorMode(mode int) error {
	if mode != 2 && mode != 4 && mode != 8 {
		return fmt.Errorf("invalid color mode: %d (must be 2, 4, or 8)", mode)
	}
	return nil
}

// RGBToGray converts an RGB color to grayscale using the luminosity method
func RGBToGray(r, g, b uint8) uint8 {
	// Using standard luminosity formula
	return uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

// GetPixelIntensity returns the intensity (0-1) of a color
func GetPixelIntensity(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit
	gray := RGBToGray(uint8(r>>8), uint8(g>>8), uint8(b>>8))
	return float64(gray) / 255.0
}
