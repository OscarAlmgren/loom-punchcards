package punchcard

import (
	"bytes"
	"strings"
	"testing"
)

func TestTextExporter_ExportCards(t *testing.T) {
	// Create test cards
	cards := []*Card{
		{
			Number: 1,
			Width:  CardWidth,
			Height: CardHeight,
			Matrix: [][]int{
				{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0},
				{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
				{1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1},
				{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0},
				{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
			},
		},
		{
			Number: 2,
			Width:  CardWidth,
			Height: CardHeight,
			Matrix: [][]int{
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
		},
	}

	exporter := NewTextExporter()
	exporter.SetTitle("Test Pattern", 2)

	var buf bytes.Buffer
	err := exporter.ExportCards(cards, &buf)
	if err != nil {
		t.Fatalf("ExportCards failed: %v", err)
	}

	output := buf.String()

	// Verify header
	if !strings.Contains(output, "Title: Test Pattern") {
		t.Errorf("Missing title in output")
	}
	if !strings.Contains(output, "Cards: 2") {
		t.Errorf("Missing card count in output")
	}
	if !strings.Contains(output, "Holes per card: 208") {
		t.Errorf("Missing holes per card in output")
	}

	// Verify card headers
	if !strings.Contains(output, "Card 1:") {
		t.Errorf("Missing Card 1 header")
	}
	if !strings.Contains(output, "Card 2:") {
		t.Errorf("Missing Card 2 header")
	}

	// Verify output contains holes and no-holes
	if !strings.Contains(output, "#") {
		t.Errorf("Missing hole characters (#)")
	}
	if !strings.Contains(output, ".") {
		t.Errorf("Missing no-hole characters (.)")
	}

	// Verify each card has 8 rows (CardHeight)
	lines := strings.Split(output, "\n")
	cardLines := 0
	inCard := false
	for _, line := range lines {
		if strings.HasPrefix(line, "Card ") {
			inCard = true
			cardLines = 0
		} else if inCard && len(line) == CardWidth {
			cardLines++
		} else if inCard && line == "" {
			if cardLines != CardHeight {
				t.Errorf("Expected %d rows per card, got %d", CardHeight, cardLines)
			}
			inCard = false
		}
	}
}

func TestTextParser_Parse(t *testing.T) {
	// Create a test text file
	input := `Title: Test Pattern
Cards: 2
Holes per card: 208

Card 1:
#.#.#.#.#.#.#.#.#.#.#.#.#.
.#.#.#.#.#.#.#.#.#.#.#.#.#
####....####....####....##
....####....####....####..
#.........................
..........................
##########################
.#.#.#.#.#.#.#.#.#.#.#.#.#

Card 2:
..........................
##########################
..........................
##########################
..........................
##########################
..........................
##########################
`

	parser := NewTextParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify metadata
	if result.Title != "Test Pattern" {
		t.Errorf("Expected title 'Test Pattern', got '%s'", result.Title)
	}
	if result.TotalCards != 2 {
		t.Errorf("Expected 2 cards, got %d", result.TotalCards)
	}
	if result.HolesPerCard != 208 {
		t.Errorf("Expected 208 holes per card, got %d", result.HolesPerCard)
	}
	if len(result.Cards) != 2 {
		t.Fatalf("Expected 2 cards, got %d", len(result.Cards))
	}

	// Verify card structure
	for i, card := range result.Cards {
		if card.Number != i+1 {
			t.Errorf("Card %d has wrong number: %d", i+1, card.Number)
		}
		if card.Width != CardWidth {
			t.Errorf("Card %d has wrong width: expected %d, got %d", i+1, CardWidth, card.Width)
		}
		if card.Height != CardHeight {
			t.Errorf("Card %d has wrong height: expected %d, got %d", i+1, CardHeight, card.Height)
		}
		if err := card.Validate(); err != nil {
			t.Errorf("Card %d validation failed: %v", i+1, err)
		}
	}

	// Verify first card first row pattern
	expectedFirstRow := []int{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0}
	for x := 0; x < CardWidth; x++ {
		if result.Cards[0].Matrix[0][x] != expectedFirstRow[x] {
			t.Errorf("Card 1 row 0 col %d: expected %d, got %d",
				x, expectedFirstRow[x], result.Cards[0].Matrix[0][x])
		}
	}
}

func TestTextParser_ParseWithO(t *testing.T) {
	// Test parsing with 'O' character for holes
	input := `Title: O Pattern
Cards: 1
Holes per card: 208

Card 1:
O.O.O.O.O.O.O.O.O.O.O.O.O.
.O.O.O.O.O.O.O.O.O.O.O.O.O
OOOO....OOOO....OOOO....OO
....OOOO....OOOO....OOOO..
O.........................
..........................
OOOOOOOOOOOOOOOOOOOOOOOOOO
.O.O.O.O.O.O.O.O.O.O.O.O.O
`

	parser := NewTextParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse with O failed: %v", err)
	}

	if len(result.Cards) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(result.Cards))
	}

	// Verify 'O' was parsed as hole (1)
	if result.Cards[0].Matrix[0][0] != 1 {
		t.Errorf("Expected 'O' to be parsed as hole (1), got %d", result.Cards[0].Matrix[0][0])
	}
}

func TestTextParser_ParseInvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing header",
			input: "Card 1:\n##########\n",
		},
		{
			name:  "invalid card width",
			input: "Title: Test\nCards: 1\nHoles per card: 208\n\nCard 1:\n###\n",
		},
		{
			name:  "wrong card count",
			input: "Title: Test\nCards: 2\nHoles per card: 208\n\nCard 1:\n" + strings.Repeat("#", CardWidth) + "\n",
		},
	}

	parser := NewTextParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.input)
			if err == nil {
				t.Errorf("Expected parse to fail for %s, but it succeeded", tt.name)
			}
		})
	}
}

func TestTextRoundTrip(t *testing.T) {
	// Create test cards
	cards := []*Card{
		{
			Number: 1,
			Width:  CardWidth,
			Height: CardHeight,
			Matrix: [][]int{
				{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0},
				{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
				{1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1},
				{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0},
				{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
			},
		},
	}

	// Export to text
	exporter := NewTextExporter()
	exporter.SetTitle("Round Trip Test", 1)

	var buf bytes.Buffer
	err := exporter.ExportCards(cards, &buf)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Parse back
	parser := NewTextParser()
	result, err := parser.Parse(buf.String())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify the round trip
	if len(result.Cards) != len(cards) {
		t.Fatalf("Card count mismatch: expected %d, got %d", len(cards), len(result.Cards))
	}

	for i := 0; i < len(cards); i++ {
		for y := 0; y < CardHeight; y++ {
			for x := 0; x < CardWidth; x++ {
				if cards[i].Matrix[y][x] != result.Cards[i].Matrix[y][x] {
					t.Errorf("Card %d [%d][%d]: expected %d, got %d",
						i+1, y, x, cards[i].Matrix[y][x], result.Cards[i].Matrix[y][x])
				}
			}
		}
	}
}

