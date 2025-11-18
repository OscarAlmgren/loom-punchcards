package punchcard

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewSVGExporter(t *testing.T) {
	exporter := NewSVGExporter()

	if exporter == nil {
		t.Fatal("NewSVGExporter() returned nil")
	}

	if !exporter.ShowGrid {
		t.Error("ShowGrid should be true by default")
	}

	if !exporter.ShowNumbers {
		t.Error("ShowNumbers should be true by default")
	}

	if exporter.HoleRadius != HoleRadius {
		t.Errorf("HoleRadius = %f, want %f", exporter.HoleRadius, HoleRadius)
	}

	if exporter.Scale != 1.0 {
		t.Error("Scale should be 1.0 by default")
	}
}

func TestExportCard(t *testing.T) {
	// Create a test card
	matrix := [][]int{
		{1, 0, 1, 0, 1, 0, 1, 0},
		{0, 1, 0, 1, 0, 1, 0, 1},
	}

	card := &Card{
		Number: 1,
		Width:  8,
		Height: 2,
		Matrix: matrix,
	}

	exporter := NewSVGExporter()
	var buf bytes.Buffer

	err := exporter.ExportCard(card, &buf)
	if err != nil {
		t.Fatalf("ExportCard() error = %v", err)
	}

	output := buf.String()

	// Verify SVG structure
	if !strings.Contains(output, "<?xml version") {
		t.Error("Output should contain XML declaration")
	}

	if !strings.Contains(output, "<svg") {
		t.Error("Output should contain SVG opening tag")
	}

	if !strings.Contains(output, "</svg>") {
		t.Error("Output should contain SVG closing tag")
	}

	if !strings.Contains(output, "Jacquard Loom Punchcard") {
		t.Error("Output should contain title")
	}

	if !strings.Contains(output, "Card #1") {
		t.Error("Output should contain card number")
	}

	// Verify circles are present
	if !strings.Contains(output, "<circle") {
		t.Error("Output should contain circle elements")
	}
}

func TestExportCardInvalid(t *testing.T) {
	// Create an invalid card
	card := &Card{
		Number: 1,
		Width:  2,
		Height: 2,
		Matrix: [][]int{{0, 1}}, // Wrong height
	}

	exporter := NewSVGExporter()
	var buf bytes.Buffer

	err := exporter.ExportCard(card, &buf)
	if err == nil {
		t.Error("ExportCard() with invalid card should return error")
	}
}

func TestExportCardWithoutGrid(t *testing.T) {
	card := createTestCard(1)

	exporter := NewSVGExporter()
	exporter.ShowGrid = false

	var buf bytes.Buffer
	err := exporter.ExportCard(card, &buf)
	if err != nil {
		t.Fatalf("ExportCard() error = %v", err)
	}

	output := buf.String()

	// Grid should not be present
	if strings.Contains(output, `id="grid"`) {
		t.Error("Output should not contain grid when ShowGrid is false")
	}
}

func TestExportCardWithoutNumbers(t *testing.T) {
	card := createTestCard(1)

	exporter := NewSVGExporter()
	exporter.ShowNumbers = false

	var buf bytes.Buffer
	err := exporter.ExportCard(card, &buf)
	if err != nil {
		t.Fatalf("ExportCard() error = %v", err)
	}

	output := buf.String()

	// Card number text should not be prominent
	// (it might still be in the title/desc, but not as visible text)
	if strings.Count(output, "<text") > 1 {
		t.Error("Output should have minimal text when ShowNumbers is false")
	}
}

func TestExportCards(t *testing.T) {
	// Create multiple test cards
	cards := []*Card{
		createTestCard(1),
		createTestCard(2),
		createTestCard(3),
	}

	exporter := NewSVGExporter()
	var buf bytes.Buffer

	err := exporter.ExportCards(cards, &buf)
	if err != nil {
		t.Fatalf("ExportCards() error = %v", err)
	}

	output := buf.String()

	// Verify all cards are present
	if !strings.Contains(output, "Card #1") {
		t.Error("Output should contain Card #1")
	}
	if !strings.Contains(output, "Card #2") {
		t.Error("Output should contain Card #2")
	}
	if !strings.Contains(output, "Card #3") {
		t.Error("Output should contain Card #3")
	}

	// Should have groups for each card
	groupCount := strings.Count(output, `<g id="card-`)
	if groupCount != 3 {
		t.Errorf("Output should contain 3 card groups, got %d", groupCount)
	}
}

func TestExportCardsEmpty(t *testing.T) {
	exporter := NewSVGExporter()
	var buf bytes.Buffer

	err := exporter.ExportCards([]*Card{}, &buf)
	if err == nil {
		t.Error("ExportCards() with empty slice should return error")
	}
}

func TestExportCardWithScale(t *testing.T) {
	card := createTestCard(1)

	exporter := NewSVGExporter()
	exporter.Scale = 2.0

	var buf bytes.Buffer
	err := exporter.ExportCard(card, &buf)
	if err != nil {
		t.Fatalf("ExportCard() error = %v", err)
	}

	output := buf.String()

	// Just verify it doesn't error with different scale
	if !strings.Contains(output, "<svg") {
		t.Error("Output should contain SVG with custom scale")
	}
}

func TestPrepareTemplateData(t *testing.T) {
	card := createTestCard(1)
	exporter := NewSVGExporter()

	data := exporter.prepareTemplateData(card)

	if data == nil {
		t.Fatal("prepareTemplateData() returned nil")
	}

	if data.CardNumber != 1 {
		t.Errorf("CardNumber = %d, want 1", data.CardNumber)
	}

	if data.Width <= 0 {
		t.Error("Width should be positive")
	}

	if data.Height <= 0 {
		t.Error("Height should be positive")
	}

	if len(data.Holes) == 0 {
		t.Error("Holes should not be empty")
	}

	expectedHoles := card.Width * card.Height
	if len(data.Holes) != expectedHoles {
		t.Errorf("Holes count = %d, want %d", len(data.Holes), expectedHoles)
	}
}

func TestExportCardTemplate(t *testing.T) {
	card := createTestCard(1)
	exporter := NewSVGExporter()

	var buf bytes.Buffer
	err := exporter.ExportCardTemplate(card, &buf)
	if err != nil {
		t.Fatalf("ExportCardTemplate() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "<svg") {
		t.Error("Template output should contain SVG")
	}

	if !strings.Contains(output, "Card #1") {
		t.Error("Template output should contain card number")
	}
}

func TestSVGHoleColors(t *testing.T) {
	// Create a card with specific pattern
	matrix := [][]int{
		{1, 0}, // One hole, one no-hole
	}

	card := &Card{
		Number: 1,
		Width:  2,
		Height: 1,
		Matrix: matrix,
	}

	exporter := NewSVGExporter()
	var buf bytes.Buffer

	err := exporter.ExportCard(card, &buf)
	if err != nil {
		t.Fatalf("ExportCard() error = %v", err)
	}

	output := buf.String()

	// Should have both filled (black) and unfilled (lightgray) holes
	if !strings.Contains(output, `fill="black"`) {
		t.Error("Output should contain black filled holes")
	}

	if !strings.Contains(output, `stroke="lightgray"`) {
		t.Error("Output should contain lightgray guide marks")
	}
}

// Helper functions

func createTestCard(number int) *Card {
	matrix := make([][]int, CardHeight)
	for y := 0; y < CardHeight; y++ {
		matrix[y] = make([]int, CardWidth)
		for x := 0; x < CardWidth; x++ {
			matrix[y][x] = (x + y) % 2
		}
	}

	return &Card{
		Number: number,
		Width:  CardWidth,
		Height: CardHeight,
		Matrix: matrix,
	}
}

// Benchmark tests

func BenchmarkExportCard(b *testing.B) {
	card := createTestCard(1)
	exporter := NewSVGExporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		exporter.ExportCard(card, &buf)
	}
}

func BenchmarkExportCards(b *testing.B) {
	cards := []*Card{
		createTestCard(1),
		createTestCard(2),
		createTestCard(3),
	}
	exporter := NewSVGExporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		exporter.ExportCards(cards, &buf)
	}
}

func BenchmarkPrepareTemplateData(b *testing.B) {
	card := createTestCard(1)
	exporter := NewSVGExporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exporter.prepareTemplateData(card)
	}
}
