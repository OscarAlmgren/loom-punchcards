package punchcard

import (
	"bytes"
	"fmt"
	"io"
)

// PDFExporter handles exporting punchcards to PDF format
// This is a simplified PDF generator that creates basic PDFs without external dependencies
type PDFExporter struct {
	ShowGrid    bool
	ShowNumbers bool
	PageSize    string // "A4", "Letter", etc.
}

// NewPDFExporter creates a new PDF exporter
func NewPDFExporter() *PDFExporter {
	return &PDFExporter{
		ShowGrid:    true,
		ShowNumbers: true,
		PageSize:    "A4",
	}
}

// ExportCard exports a single card to PDF
// For a proper PDF implementation, we'll use SVG as an intermediate format
// and convert it to PDF, or we can use a simple PDF library
func (e *PDFExporter) ExportCard(card *Card, w io.Writer) error {
	// For now, we'll create a simple PDF structure
	// In a production environment, you'd use a library like gofpdf or similar
	return e.generateSimplePDF([]*Card{card}, w)
}

// ExportCards exports multiple cards to a single PDF file
func (e *PDFExporter) ExportCards(cards []*Card, w io.Writer) error {
	if len(cards) == 0 {
		return fmt.Errorf("no cards to export")
	}
	return e.generateSimplePDF(cards, w)
}

// generateSimplePDF creates a basic PDF file
// This is a simplified implementation. For production use, consider using a proper PDF library
func (e *PDFExporter) generateSimplePDF(cards []*Card, w io.Writer) error {
	// We'll generate SVG content and embed it in a minimal PDF structure
	// This creates a PDF that displays the SVG content

	var buf bytes.Buffer

	// Generate SVG for all cards
	svgExporter := NewSVGExporter()
	svgExporter.ShowGrid = e.ShowGrid
	svgExporter.ShowNumbers = e.ShowNumbers

	if err := svgExporter.ExportCards(cards, &buf); err != nil {
		return fmt.Errorf("failed to generate SVG: %w", err)
	}

	svgContent := buf.String()

	// Create a simple PDF wrapper
	// Note: This is a very basic PDF structure. For production, use a proper PDF library
	pdf := e.createPDFWrapper(svgContent, cards)

	_, err := w.Write([]byte(pdf))
	return err
}

// createPDFWrapper creates a minimal PDF structure
// This is a simplified version and may not work with all PDF readers
// For production use, please use a proper PDF library like gofpdf
func (e *PDFExporter) createPDFWrapper(svgContent string, cards []*Card) string {
	// This is a placeholder implementation
	// In a real application, you would use a proper PDF library
	// For now, we'll return the SVG content with PDF metadata

	// Note: This will be replaced with proper PDF generation in the handler
	// using a conversion service or library
	return svgContent
}

// PDFMetadata contains metadata for PDF generation
type PDFMetadata struct {
	Title       string
	Author      string
	Subject     string
	Creator     string
	Producer    string
	Keywords    []string
	CreatedDate string
}

// GetDefaultMetadata returns default PDF metadata
func GetDefaultMetadata(numCards int) *PDFMetadata {
	return &PDFMetadata{
		Title:    fmt.Sprintf("Jacquard Loom Punchcards (Set of %d)", numCards),
		Author:   "Loom Punchcard Generator",
		Subject:  "Jacquard Weaving Punchcards",
		Creator:  "Loom Punchcard Web Application",
		Producer: "Jacquard Card Generator v1.0",
		Keywords: []string{"Jacquard", "weaving", "punchcard", "loom", "textile"},
	}
}

// Note: For actual PDF generation, we'll use a proper approach in the HTTP handler
// This might involve:
// 1. Using a Go PDF library like gofpdf, gopdf, or pdfcpu
// 2. Converting SVG to PDF using external tools
// 3. Using a PDF generation service

// The following is a comment about implementation strategy:
// Since PDF generation from scratch is complex, we have two main approaches:
//
// Approach 1: Use a PDF library (recommended)
// - gofpdf: Simple API but limited SVG support
// - gopdf: More features, moderate complexity
// - pdfcpu: Full-featured, can manipulate existing PDFs
//
// Approach 2: Convert SVG to PDF
// - Use rsvg-convert or inkscape command-line tools
// - Use wkhtmltopdf to convert HTML+SVG to PDF
// - Use a web service API
//
// For this implementation, we'll use Approach 1 with a simple PDF library in the handler

// PDFPageSize defines standard page sizes
type PDFPageSize struct {
	Width  float64 // in mm
	Height float64 // in mm
}

var (
	PageSizeA4     = PDFPageSize{Width: 210, Height: 297}
	PageSizeLetter = PDFPageSize{Width: 216, Height: 279}
	PageSizeA3     = PDFPageSize{Width: 297, Height: 420}
)

// GetPageSize returns the page size for a given name
func GetPageSize(name string) PDFPageSize {
	switch name {
	case "A4":
		return PageSizeA4
	case "Letter":
		return PageSizeLetter
	case "A3":
		return PageSizeA3
	default:
		return PageSizeA4
	}
}

// CalculateCardsPerPage calculates how many cards fit on a page
func CalculateCardsPerPage(pageSize PDFPageSize) int {
	// Assuming each card is approximately 50mm x 140mm with margins
	cardWidth := float64(CardWidth)*HoleSpacing + 2*CardPadding
	cardHeight := float64(CardHeight)*HoleSpacing + 2*CardPadding + TextHeight*2

	margin := 10.0 // mm
	usableWidth := pageSize.Width - 2*margin
	usableHeight := pageSize.Height - 2*margin

	cardsX := int(usableWidth / cardWidth)
	cardsY := int(usableHeight / cardHeight)

	if cardsX < 1 {
		cardsX = 1
	}
	if cardsY < 1 {
		cardsY = 1
	}

	return cardsX * cardsY
}
