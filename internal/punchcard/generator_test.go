package punchcard

import (
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Error("NewGenerator() returned nil")
	}
	if g.CardsPerRow != 1 {
		t.Errorf("CardsPerRow = %d, want 1", g.CardsPerRow)
	}
}

func TestGenerate(t *testing.T) {
	// Expected image width is CardWidth * CardHeight (26 * 8 = 208)
	expectedWidth := CardWidth * CardHeight

	tests := []struct {
		name          string
		matrixHeight  int
		matrixWidth   int
		expectedCards int
	}{
		{"single card (one row)", 1, expectedWidth, 1},
		{"two cards (two rows)", 2, expectedWidth, 2},
		{"ten cards (ten rows)", 10, expectedWidth, 10},
		{"hundred cards (hundred rows)", 100, expectedWidth, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test matrix
			matrix := createTestMatrix(tt.matrixHeight, tt.matrixWidth)

			generator := NewGenerator()
			cards, err := generator.Generate(matrix)

			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(cards) != tt.expectedCards {
				t.Errorf("Generate() returned %d cards, want %d", len(cards), tt.expectedCards)
			}

			// Verify card dimensions
			for i, card := range cards {
				if card.Width != CardWidth {
					t.Errorf("Card %d width = %d, want %d", i, card.Width, CardWidth)
				}
				if card.Height != CardHeight {
					t.Errorf("Card %d height = %d, want %d", i, card.Height, CardHeight)
				}
				if card.Number != i+1 {
					t.Errorf("Card %d number = %d, want %d", i, card.Number, i+1)
				}
			}
		})
	}
}

func TestGenerateEmptyMatrix(t *testing.T) {
	generator := NewGenerator()
	_, err := generator.Generate([][]int{})
	if err == nil {
		t.Error("Generate() with empty matrix should return error")
	}
}

func TestGenerateInvalidWidth(t *testing.T) {
	generator := NewGenerator()

	// Create matrix with wrong width (should be CardWidth * CardHeight = 208)
	wrongWidth := CardWidth * CardHeight + 1
	matrix := createTestMatrix(1, wrongWidth)

	_, err := generator.Generate(matrix)
	if err == nil {
		t.Error("Generate() with invalid width should return error")
	}
}

func TestCardValidate(t *testing.T) {
	tests := []struct {
		name      string
		card      *Card
		wantError bool
	}{
		{
			name: "valid card",
			card: &Card{
				Number: 1,
				Width:  CardWidth,
				Height: CardHeight,
				Matrix: createTestMatrix(CardHeight, CardWidth),
			},
			wantError: false,
		},
		{
			name: "invalid dimensions",
			card: &Card{
				Number: 1,
				Width:  0,
				Height: 0,
				Matrix: [][]int{},
			},
			wantError: true,
		},
		{
			name: "mismatched height",
			card: &Card{
				Number: 1,
				Width:  CardWidth,
				Height: CardHeight,
				Matrix: createTestMatrix(CardHeight-1, CardWidth),
			},
			wantError: true,
		},
		{
			name: "mismatched width",
			card: &Card{
				Number: 1,
				Width:  CardWidth,
				Height: CardHeight,
				Matrix: createTestMatrix(CardHeight, CardWidth-1),
			},
			wantError: true,
		},
		{
			name: "invalid value",
			card: &Card{
				Number: 1,
				Width:  2,
				Height: 2,
				Matrix: [][]int{{0, 1}, {2, 0}}, // 2 is invalid
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCardCountHoles(t *testing.T) {
	tests := []struct {
		name     string
		matrix   [][]int
		expected int
	}{
		{
			name: "no holes",
			matrix: [][]int{
				{0, 0, 0},
				{0, 0, 0},
			},
			expected: 0,
		},
		{
			name: "all holes",
			matrix: [][]int{
				{1, 1, 1},
				{1, 1, 1},
			},
			expected: 6,
		},
		{
			name: "mixed",
			matrix: [][]int{
				{1, 0, 1},
				{0, 1, 0},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &Card{
				Number: 1,
				Width:  len(tt.matrix[0]),
				Height: len(tt.matrix),
				Matrix: tt.matrix,
			}

			count := card.CountHoles()
			if count != tt.expected {
				t.Errorf("CountHoles() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestCardGetRow(t *testing.T) {
	matrix := createTestMatrix(3, 3)
	card := &Card{
		Number: 1,
		Width:  3,
		Height: 3,
		Matrix: matrix,
	}

	// Valid row
	row, err := card.GetRow(1)
	if err != nil {
		t.Errorf("GetRow(1) error = %v", err)
	}
	if len(row) != 3 {
		t.Errorf("GetRow(1) length = %d, want 3", len(row))
	}

	// Invalid row (negative)
	_, err = card.GetRow(-1)
	if err == nil {
		t.Error("GetRow(-1) should return error")
	}

	// Invalid row (too large)
	_, err = card.GetRow(3)
	if err == nil {
		t.Error("GetRow(3) should return error for 3-row card")
	}
}

func TestCardIsHolePunched(t *testing.T) {
	matrix := [][]int{
		{1, 0, 1},
		{0, 1, 0},
		{1, 1, 1},
	}

	card := &Card{
		Number: 1,
		Width:  3,
		Height: 3,
		Matrix: matrix,
	}

	tests := []struct {
		x, y     int
		expected bool
	}{
		{0, 0, true},  // matrix[0][0] = 1
		{1, 0, false}, // matrix[0][1] = 0
		{2, 0, true},  // matrix[0][2] = 1
		{1, 1, true},  // matrix[1][1] = 1
		{0, 1, false}, // matrix[1][0] = 0
		{5, 5, false}, // out of bounds
		{-1, 0, false}, // out of bounds
	}

	for _, tt := range tests {
		result := card.IsHolePunched(tt.x, tt.y)
		if result != tt.expected {
			t.Errorf("IsHolePunched(%d, %d) = %v, want %v", tt.x, tt.y, result, tt.expected)
		}
	}
}

func TestCardClone(t *testing.T) {
	original := &Card{
		Number: 1,
		Width:  3,
		Height: 3,
		Matrix: createTestMatrix(3, 3),
	}

	clone := original.Clone()

	// Verify clone has same values
	if clone.Number != original.Number {
		t.Error("Clone has different number")
	}
	if clone.Width != original.Width {
		t.Error("Clone has different width")
	}
	if clone.Height != original.Height {
		t.Error("Clone has different height")
	}

	// Verify it's a deep copy
	clone.Matrix[0][0] = 1 - clone.Matrix[0][0]
	if clone.Matrix[0][0] == original.Matrix[0][0] {
		t.Error("Clone is not a deep copy - modifying clone affected original")
	}
}

func TestCardInvert(t *testing.T) {
	matrix := [][]int{
		{1, 0, 1},
		{0, 1, 0},
	}

	card := &Card{
		Number: 1,
		Width:  3,
		Height: 2,
		Matrix: matrix,
	}

	card.Invert()

	expected := [][]int{
		{0, 1, 0},
		{1, 0, 1},
	}

	for y := 0; y < card.Height; y++ {
		for x := 0; x < card.Width; x++ {
			if card.Matrix[y][x] != expected[y][x] {
				t.Errorf("After invert, Matrix[%d][%d] = %d, want %d",
					y, x, card.Matrix[y][x], expected[y][x])
			}
		}
	}
}

func TestCardGetBinaryString(t *testing.T) {
	matrix := [][]int{
		{1, 0},
		{0, 1},
	}

	card := &Card{
		Number: 1,
		Width:  2,
		Height: 2,
		Matrix: matrix,
	}

	str := card.GetBinaryString()
	if str == "" {
		t.Error("GetBinaryString() returned empty string")
	}
	// Should contain card number
	if len(str) < 10 {
		t.Error("GetBinaryString() output seems too short")
	}
}

func TestCardGetCardInfo(t *testing.T) {
	matrix := createTestMatrix(CardHeight, CardWidth)
	card := &Card{
		Number: 5,
		Width:  CardWidth,
		Height: CardHeight,
		Matrix: matrix,
	}

	info := card.GetCardInfo()
	if info == "" {
		t.Error("GetCardInfo() returned empty string")
	}
	// Should contain card number
	if len(info) < 10 {
		t.Error("GetCardInfo() output seems too short")
	}
}

func TestGenerateMetadata(t *testing.T) {
	// Create test cards (3 rows, each becomes one card)
	expectedWidth := CardWidth * CardHeight
	matrix := createTestMatrix(3, expectedWidth)
	generator := NewGenerator()
	cards, err := generator.Generate(matrix)
	if err != nil {
		t.Fatalf("Failed to generate cards: %v", err)
	}

	meta := GenerateMetadata(cards)

	if meta.TotalCards != 3 {
		t.Errorf("TotalCards = %d, want 3", meta.TotalCards)
	}
	if meta.CardWidth != CardWidth {
		t.Errorf("CardWidth = %d, want %d", meta.CardWidth, CardWidth)
	}
	if meta.CardHeight != CardHeight {
		t.Errorf("CardHeight = %d, want %d", meta.CardHeight, CardHeight)
	}
	if len(meta.HolesPerCard) != 3 {
		t.Errorf("HolesPerCard length = %d, want 3", len(meta.HolesPerCard))
	}
	if meta.AverageDensity < 0 || meta.AverageDensity > 100 {
		t.Errorf("AverageDensity = %f, should be between 0 and 100", meta.AverageDensity)
	}
}

func TestGenerateMetadataEmpty(t *testing.T) {
	meta := GenerateMetadata([]*Card{})

	if meta.TotalCards != 0 {
		t.Errorf("TotalCards = %d, want 0", meta.TotalCards)
	}
}

// Helper functions

func createTestMatrix(height, width int) [][]int {
	matrix := make([][]int, height)
	for y := 0; y < height; y++ {
		matrix[y] = make([]int, width)
		for x := 0; x < width; x++ {
			// Create a checkerboard pattern
			matrix[y][x] = (x + y) % 2
		}
	}
	return matrix
}

// Benchmark tests

func BenchmarkGenerate(b *testing.B) {
	expectedWidth := CardWidth * CardHeight
	matrix := createTestMatrix(10, expectedWidth) // 10 rows = 10 cards
	generator := NewGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Generate(matrix)
	}
}

func BenchmarkCardCountHoles(b *testing.B) {
	card := &Card{
		Number: 1,
		Width:  CardWidth,
		Height: CardHeight,
		Matrix: createTestMatrix(CardHeight, CardWidth),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		card.CountHoles()
	}
}

func BenchmarkCardClone(b *testing.B) {
	card := &Card{
		Number: 1,
		Width:  CardWidth,
		Height: CardHeight,
		Matrix: createTestMatrix(CardHeight, CardWidth),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		card.Clone()
	}
}
