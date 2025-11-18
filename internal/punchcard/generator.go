package punchcard

import (
	"fmt"
)

// CardType represents different loom card specifications
type CardType string

const (
	// CardType26x8 is the standard small card (26 columns × 8 rows = 208 holes)
	CardType26x8 CardType = "26x8"

	// CardType50x12 is a larger card for more detailed patterns (50 columns × 12 rows = 600 holes)
	CardType50x12 CardType = "50x12"
)

// CardDimensions holds the width and height for a card type
type CardDimensions struct {
	Width  int
	Height int
}

// GetCardDimensions returns the dimensions for a given card type
func GetCardDimensions(cardType CardType) CardDimensions {
	switch cardType {
	case CardType50x12:
		return CardDimensions{Width: 50, Height: 12}
	case CardType26x8:
		fallthrough
	default:
		return CardDimensions{Width: 26, Height: 8}
	}
}

// ValidateCardType checks if the card type is valid
func ValidateCardType(cardType string) error {
	if cardType != string(CardType26x8) && cardType != string(CardType50x12) {
		return fmt.Errorf("invalid card type: %s (must be '26x8' or '50x12')", cardType)
	}
	return nil
}

// Legacy constants for backward compatibility
const (
	// CardWidth represents the number of columns in a standard Jacquard punchcard
	// Horizontal orientation: 26 columns per card
	CardWidth = 26

	// CardHeight represents the number of rows in a single punchcard
	// Horizontal orientation: 8 rows per card
	// Each card represents one row of the source image (208 pixels = 26*8)
	CardHeight = 8
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
	CardsPerRow int            // How many cards wide the pattern is (usually 1 for standard looms)
	Dimensions  CardDimensions // Card dimensions (width and height)
}

// NewGenerator creates a new punchcard generator with default 26x8 card type
func NewGenerator() *Generator {
	return &Generator{
		CardsPerRow: 1,
		Dimensions:  GetCardDimensions(CardType26x8),
	}
}

// NewGeneratorWithType creates a new punchcard generator with a specific card type
func NewGeneratorWithType(cardType CardType) *Generator {
	return &Generator{
		CardsPerRow: 1,
		Dimensions:  GetCardDimensions(cardType),
	}
}

// Generate converts a binary matrix (from processed image) into a sequence of punchcards
// Each card represents one row of the image, arranged in a Width x Height grid
// The image should be resized to (Width * Height) pixels wide
func (g *Generator) Generate(matrix [][]int) ([]*Card, error) {
	if len(matrix) == 0 {
		return nil, fmt.Errorf("empty matrix provided")
	}

	imageWidth := len(matrix[0])
	imageHeight := len(matrix)

	// Expected width is Width * Height (e.g., 26 * 8 = 208 or 50 * 12 = 600)
	expectedWidth := g.Dimensions.Width * g.Dimensions.Height
	if imageWidth != expectedWidth {
		return nil, fmt.Errorf("image width (%d) does not match expected width (%d = %d x %d)",
			imageWidth, expectedWidth, g.Dimensions.Width, g.Dimensions.Height)
	}

	// Each row of the image becomes one card
	numCards := imageHeight
	cards := make([]*Card, numCards)

	// Convert each row into a card
	for cardNum := 0; cardNum < numCards; cardNum++ {
		// Get the source row (e.g., 208 or 600 pixels)
		sourceRow := matrix[cardNum]

		// Create the card matrix (Width columns x Height rows)
		cardMatrix := make([][]int, g.Dimensions.Height)

		// Reshape the pixel row into a Width x Height grid
		// We fill the grid row by row (left to right, top to bottom)
		for row := 0; row < g.Dimensions.Height; row++ {
			cardMatrix[row] = make([]int, g.Dimensions.Width)
			for col := 0; col < g.Dimensions.Width; col++ {
				pixelIndex := row*g.Dimensions.Width + col
				cardMatrix[row][col] = sourceRow[pixelIndex]
			}
		}

		cards[cardNum] = &Card{
			Number: cardNum + 1, // 1-indexed for user display
			Matrix: cardMatrix,
			Width:  g.Dimensions.Width,
			Height: g.Dimensions.Height,
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
