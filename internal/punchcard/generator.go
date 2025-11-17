package punchcard

import (
	"fmt"
)

const (
	// CardWidth represents the number of columns in a standard Jacquard punchcard
	// Based on simplified home loom specifications (8 needles)
	CardWidth = 8

	// CardHeight represents the number of rows in a single punchcard
	// Each card represents one row of weave pattern
	CardHeight = 26
)

// Card represents a single Jacquard punchcard
type Card struct {
	Number int       // Sequential number for ordering
	Matrix [][]int   // Binary matrix: 1 = hole punched, 0 = no hole
	Width  int       // Number of columns (typically 8)
	Height int       // Number of rows (typically 26)
}

// Generator creates punchcards from binary image data
type Generator struct {
	CardsPerRow int // How many cards wide the pattern is (usually 1 for standard looms)
}

// NewGenerator creates a new punchcard generator
func NewGenerator() *Generator {
	return &Generator{
		CardsPerRow: 1,
	}
}

// Generate converts a binary matrix (from processed image) into a sequence of punchcards
// Each card represents CardHeight rows of the pattern
func (g *Generator) Generate(matrix [][]int) ([]*Card, error) {
	if len(matrix) == 0 {
		return nil, fmt.Errorf("empty matrix provided")
	}

	imageWidth := len(matrix[0])
	imageHeight := len(matrix)

	// Validate that the image width matches the card width
	// If not, we need to resize or crop
	if imageWidth != CardWidth {
		return nil, fmt.Errorf("image width (%d) does not match card width (%d)", imageWidth, CardWidth)
	}

	// Calculate the number of cards needed
	numCards := (imageHeight + CardHeight - 1) / CardHeight // Ceiling division

	cards := make([]*Card, numCards)

	// Split the matrix into individual cards
	for cardNum := 0; cardNum < numCards; cardNum++ {
		startRow := cardNum * CardHeight
		endRow := startRow + CardHeight
		if endRow > imageHeight {
			endRow = imageHeight
		}

		// Create the card matrix
		cardMatrix := make([][]int, CardHeight)
		for i := 0; i < CardHeight; i++ {
			cardMatrix[i] = make([]int, CardWidth)

			// Copy data from the source matrix if available
			sourceRow := startRow + i
			if sourceRow < imageHeight {
				copy(cardMatrix[i], matrix[sourceRow])
			}
			// If we're past the end of the image, leave as zeros (no holes)
		}

		cards[cardNum] = &Card{
			Number: cardNum + 1, // 1-indexed for user display
			Matrix: cardMatrix,
			Width:  CardWidth,
			Height: CardHeight,
		}
	}

	return cards, nil
}

// GetCardInfo returns information about a specific card
func (c *Card) GetCardInfo() string {
	holes := c.CountHoles()
	density := float64(holes) / float64(c.Width*c.Height) * 100

	return fmt.Sprintf("Card #%d: %dx%d, %d holes (%.1f%% density)",
		c.Number, c.Width, c.Height, holes, density)
}

// CountHoles returns the number of punched holes in the card
func (c *Card) CountHoles() int {
	count := 0
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if c.Matrix[y][x] == 1 {
				count++
			}
		}
	}
	return count
}

// GetRow returns a specific row of the card
func (c *Card) GetRow(rowIndex int) ([]int, error) {
	if rowIndex < 0 || rowIndex >= c.Height {
		return nil, fmt.Errorf("row index %d out of bounds (0-%d)", rowIndex, c.Height-1)
	}
	return c.Matrix[rowIndex], nil
}

// IsHolePunched checks if a hole is punched at the given coordinates
func (c *Card) IsHolePunched(x, y int) bool {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return false
	}
	return c.Matrix[y][x] == 1
}

// GetBinaryString returns a string representation of the card in binary form
// Useful for debugging and verification
func (c *Card) GetBinaryString() string {
	result := fmt.Sprintf("Card #%d:\n", c.Number)
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if c.Matrix[y][x] == 1 {
				result += "█"
			} else {
				result += "·"
			}
		}
		result += "\n"
	}
	return result
}

// Validate checks if the card has valid dimensions and data
func (c *Card) Validate() error {
	if c.Width <= 0 || c.Height <= 0 {
		return fmt.Errorf("invalid card dimensions: %dx%d", c.Width, c.Height)
	}

	if len(c.Matrix) != c.Height {
		return fmt.Errorf("matrix height (%d) does not match card height (%d)", len(c.Matrix), c.Height)
	}

	for y, row := range c.Matrix {
		if len(row) != c.Width {
			return fmt.Errorf("row %d width (%d) does not match card width (%d)", y, len(row), c.Width)
		}

		// Validate binary values
		for x, val := range row {
			if val != 0 && val != 1 {
				return fmt.Errorf("invalid value at (%d,%d): %d (must be 0 or 1)", x, y, val)
			}
		}
	}

	return nil
}

// Clone creates a deep copy of the card
func (c *Card) Clone() *Card {
	clone := &Card{
		Number: c.Number,
		Width:  c.Width,
		Height: c.Height,
		Matrix: make([][]int, c.Height),
	}

	for y := 0; y < c.Height; y++ {
		clone.Matrix[y] = make([]int, c.Width)
		copy(clone.Matrix[y], c.Matrix[y])
	}

	return clone
}

// Invert inverts the card (holes become no-holes and vice versa)
// Useful for creating negative patterns
func (c *Card) Invert() {
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			c.Matrix[y][x] = 1 - c.Matrix[y][x]
		}
	}
}

// GetMetadata returns metadata about the card set
type Metadata struct {
	TotalCards    int
	CardWidth     int
	CardHeight    int
	TotalRows     int
	HolesPerCard  []int
	AverageDensity float64
}

// GenerateMetadata creates metadata for a set of cards
func GenerateMetadata(cards []*Card) *Metadata {
	if len(cards) == 0 {
		return &Metadata{}
	}

	meta := &Metadata{
		TotalCards:   len(cards),
		CardWidth:    cards[0].Width,
		CardHeight:   cards[0].Height,
		TotalRows:    len(cards) * cards[0].Height,
		HolesPerCard: make([]int, len(cards)),
	}

	totalHoles := 0
	for i, card := range cards {
		holes := card.CountHoles()
		meta.HolesPerCard[i] = holes
		totalHoles += holes
	}

	totalPossibleHoles := len(cards) * cards[0].Width * cards[0].Height
	if totalPossibleHoles > 0 {
		meta.AverageDensity = float64(totalHoles) / float64(totalPossibleHoles) * 100
	}

	return meta
}
