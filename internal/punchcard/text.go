package punchcard

import (
	"fmt"
	"io"
	"strings"
)

// TextExporter handles exporting punchcards to text format
// The text format is human-readable and editable, using:
// - # or O for punched holes
// - . for no holes
type TextExporter struct {
	Title        string // Pattern title
	TotalCards   int    // Total number of cards in the series
	HoleChar     rune   // Character to represent holes (default: #)
	NoHoleChar   rune   // Character to represent no holes (default: .)
}

// NewTextExporter creates a new text exporter with default settings
func NewTextExporter() *TextExporter {
	return &TextExporter{
		HoleChar:   '#',
		NoHoleChar: '.',
	}
}

// SetTitle sets the title and total card count
func (e *TextExporter) SetTitle(title string, totalCards int) {
	e.Title = title
	e.TotalCards = totalCards
}

// ExportCards exports multiple cards to text format
func (e *TextExporter) ExportCards(cards []*Card, w io.Writer) error {
	if len(cards) == 0 {
		return fmt.Errorf("no cards to export")
	}

	// Calculate total holes per card (all cards should have the same dimensions)
	holesPerCard := cards[0].Width * cards[0].Height

	// Write header
	if e.Title != "" {
		fmt.Fprintf(w, "Title: %s\n", e.Title)
	} else {
		fmt.Fprintf(w, "Title: Untitled Pattern\n")
	}
	fmt.Fprintf(w, "Cards: %d\n", len(cards))
	fmt.Fprintf(w, "Holes per card: %d\n", holesPerCard)
	fmt.Fprintf(w, "\n")

	// Write each card
	for i, card := range cards {
		if err := card.Validate(); err != nil {
			return fmt.Errorf("invalid card %d: %w", i+1, err)
		}

		// Card header
		fmt.Fprintf(w, "Card %d:\n", card.Number)

		// Write the card matrix
		// Each row is CardWidth (26) columns wide
		for y := 0; y < card.Height; y++ {
			for x := 0; x < card.Width; x++ {
				if card.Matrix[y][x] == 1 {
					fmt.Fprintf(w, "%c", e.HoleChar)
				} else {
					fmt.Fprintf(w, "%c", e.NoHoleChar)
				}
			}
			fmt.Fprintf(w, "\n")
		}

		// Add empty line between cards (except after the last card)
		if i < len(cards)-1 {
			fmt.Fprintf(w, "\n")
		}
	}

	return nil
}

// TextParser handles parsing text format back into cards
type TextParser struct{}

// NewTextParser creates a new text parser
func NewTextParser() *TextParser {
	return &TextParser{}
}

// ParseResult contains the parsed data
type ParseResult struct {
	Title      string
	Cards      []*Card
	TotalCards int
	HolesPerCard int
}

// Parse parses a text format punchcard file
func (p *TextParser) Parse(content string) (*ParseResult, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 4 {
		return nil, fmt.Errorf("invalid file format: too few lines")
	}

	result := &ParseResult{}

	// Parse header
	lineIdx := 0

	// Parse Title
	if !strings.HasPrefix(lines[lineIdx], "Title: ") {
		return nil, fmt.Errorf("missing Title header on line %d", lineIdx+1)
	}
	result.Title = strings.TrimPrefix(lines[lineIdx], "Title: ")
	lineIdx++

	// Parse Cards
	if !strings.HasPrefix(lines[lineIdx], "Cards: ") {
		return nil, fmt.Errorf("missing Cards header on line %d", lineIdx+1)
	}
	_, err := fmt.Sscanf(lines[lineIdx], "Cards: %d", &result.TotalCards)
	if err != nil {
		return nil, fmt.Errorf("invalid Cards value on line %d: %w", lineIdx+1, err)
	}
	lineIdx++

	// Parse Holes per card
	if !strings.HasPrefix(lines[lineIdx], "Holes per card: ") {
		return nil, fmt.Errorf("missing Holes per card header on line %d", lineIdx+1)
	}
	_, err = fmt.Sscanf(lines[lineIdx], "Holes per card: %d", &result.HolesPerCard)
	if err != nil {
		return nil, fmt.Errorf("invalid Holes per card value on line %d: %w", lineIdx+1, err)
	}
	lineIdx++

	// Skip empty line after header
	if lineIdx < len(lines) && strings.TrimSpace(lines[lineIdx]) == "" {
		lineIdx++
	}

	// Parse cards
	result.Cards = make([]*Card, 0, result.TotalCards)
	cardNumber := 1

	for lineIdx < len(lines) {
		// Skip empty lines
		if strings.TrimSpace(lines[lineIdx]) == "" {
			lineIdx++
			continue
		}

		// Parse card header "Card N:"
		var parsedCardNum int
		if !strings.HasPrefix(lines[lineIdx], "Card ") {
			// If we've parsed all expected cards, we're done
			if len(result.Cards) == result.TotalCards {
				break
			}
			return nil, fmt.Errorf("expected Card header on line %d, got: %s", lineIdx+1, lines[lineIdx])
		}
		_, err = fmt.Sscanf(lines[lineIdx], "Card %d:", &parsedCardNum)
		if err != nil {
			return nil, fmt.Errorf("invalid Card header on line %d: %w", lineIdx+1, err)
		}
		lineIdx++

		// Parse card matrix (CardHeight rows of CardWidth columns)
		matrix := make([][]int, 0, CardHeight)

		for row := 0; row < CardHeight; row++ {
			if lineIdx >= len(lines) {
				return nil, fmt.Errorf("unexpected end of file while parsing card %d row %d", parsedCardNum, row+1)
			}

			line := lines[lineIdx]
			lineIdx++

			// Parse the row
			if len(line) != CardWidth {
				return nil, fmt.Errorf("card %d row %d has incorrect width: expected %d, got %d",
					parsedCardNum, row+1, CardWidth, len(line))
			}

			rowData := make([]int, CardWidth)
			for col, char := range line {
				switch char {
				case '#', 'O', 'o':
					rowData[col] = 1
				case '.':
					rowData[col] = 0
				default:
					return nil, fmt.Errorf("invalid character '%c' in card %d row %d col %d (expected #, O, or .)",
						char, parsedCardNum, row+1, col+1)
				}
			}
			matrix = append(matrix, rowData)
		}

		// Create the card
		card := &Card{
			Number: cardNumber,
			Matrix: matrix,
			Width:  CardWidth,
			Height: CardHeight,
		}

		// Validate the card
		if err := card.Validate(); err != nil {
			return nil, fmt.Errorf("invalid card %d: %w", cardNumber, err)
		}

		result.Cards = append(result.Cards, card)
		cardNumber++
	}

	// Verify we got all cards
	if len(result.Cards) != result.TotalCards {
		return nil, fmt.Errorf("expected %d cards but found %d", result.TotalCards, len(result.Cards))
	}

	return result, nil
}
