package punchcard

import (
	"fmt"
	"html/template"
	"io"
)

const (
	// SVG rendering constants (in millimeters, converted to pixels at 96 DPI)
	HoleRadius    = 2.0  // Radius of each hole in mm
	HoleSpacing   = 5.0  // Spacing between hole centers in mm
	CardPadding   = 10.0 // Padding around the card edge in mm
	TextHeight    = 8.0  // Height of text in mm
	MMToPixel     = 3.78 // Conversion factor: 96 DPI = 3.78 pixels per mm
)

// SVGExporter handles exporting punchcards to SVG format
type SVGExporter struct {
	ShowGrid      bool    // Whether to show a grid
	ShowNumbers   bool    // Whether to show card numbers
	HoleRadius    float64 // Radius of holes in mm
	HoleSpacing   float64 // Spacing between holes in mm
	Scale         float64 // Scale factor for the entire card
	Title         string  // Optional title to display on cards
	TotalCards    int     // Total number of cards in the series
}

// NewSVGExporter creates a new SVG exporter with default settings
func NewSVGExporter() *SVGExporter {
	return &SVGExporter{
		ShowGrid:    true,
		ShowNumbers: true,
		HoleRadius:  HoleRadius,
		HoleSpacing: HoleSpacing,
		Scale:       1.0,
		Title:       "",
		TotalCards:  0,
	}
}

// SetTitle sets the title and total card count for display on cards
func (e *SVGExporter) SetTitle(title string, totalCards int) {
	e.Title = title
	e.TotalCards = totalCards
}

// ExportCard exports a single card to SVG format
func (e *SVGExporter) ExportCard(card *Card, w io.Writer) error {
	if err := card.Validate(); err != nil {
		return fmt.Errorf("invalid card: %w", err)
	}

	// Calculate dimensions
	cardWidth := float64(card.Width)*e.HoleSpacing*e.Scale + 2*CardPadding
	cardHeight := float64(card.Height)*e.HoleSpacing*e.Scale + 2*CardPadding + TextHeight*2

	// Convert to pixels
	widthPx := cardWidth * MMToPixel
	heightPx := cardHeight * MMToPixel

	// Write SVG header
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%.2fmm" height="%.2fmm" viewBox="0 0 %.2f %.2f">`,
		cardWidth, cardHeight, widthPx, heightPx)
	fmt.Fprintf(w, "\n")

	// Add title and description
	fmt.Fprintf(w, `  <title>Jacquard Loom Punchcard #%d</title>`, card.Number)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, `  <desc>%s - For use in Jacquard weaving looms</desc>`, card.GetCardInfo())
	fmt.Fprintf(w, "\n\n")

	// Background
	fmt.Fprintf(w, `  <rect width="100%%" height="100%%" fill="white"/>`)
	fmt.Fprintf(w, "\n\n")

	// Card number at top (with optional title)
	if e.ShowNumbers {
		fmt.Fprintf(w, `  <text x="%.2f" y="%.2f" font-family="monospace" font-size="%.2f" text-anchor="middle" fill="black">`,
			widthPx/2, TextHeight*MMToPixel*0.8, TextHeight*MMToPixel*0.6)

		// Display title with card number in format "Title_name #1/156"
		if e.Title != "" && e.TotalCards > 0 {
			fmt.Fprintf(w, "%s #%d/%d", e.Title, card.Number, e.TotalCards)
		} else if e.TotalCards > 0 {
			fmt.Fprintf(w, "Card #%d/%d", card.Number, e.TotalCards)
		} else {
			fmt.Fprintf(w, "Card #%d", card.Number)
		}

		fmt.Fprintf(w, "</text>\n")
	}

	// Draw grid lines if enabled
	if e.ShowGrid {
		e.drawGrid(w, card, widthPx, heightPx)
	}

	// Draw holes
	startX := CardPadding * MMToPixel
	startY := (CardPadding + TextHeight) * MMToPixel

	for y := 0; y < card.Height; y++ {
		for x := 0; x < card.Width; x++ {
			cx := startX + float64(x)*e.HoleSpacing*e.Scale*MMToPixel
			cy := startY + float64(y)*e.HoleSpacing*e.Scale*MMToPixel

			if card.Matrix[y][x] == 1 {
				// Punched hole - filled circle
				fmt.Fprintf(w, `  <circle cx="%.2f" cy="%.2f" r="%.2f" fill="black"/>`,
					cx, cy, e.HoleRadius*e.Scale*MMToPixel)
				fmt.Fprintf(w, "\n")
			} else {
				// No hole - just a small guide mark
				fmt.Fprintf(w, `  <circle cx="%.2f" cy="%.2f" r="%.2f" fill="none" stroke="lightgray" stroke-width="0.5"/>`,
					cx, cy, e.HoleRadius*e.Scale*MMToPixel*0.3)
				fmt.Fprintf(w, "\n")
			}
		}
	}

	// Card info at bottom
	if e.ShowNumbers {
		infoY := heightPx - TextHeight*MMToPixel*0.3
		fmt.Fprintf(w, `  <text x="%.2f" y="%.2f" font-family="monospace" font-size="%.2f" text-anchor="middle" fill="gray">`,
			widthPx/2, infoY, TextHeight*MMToPixel*0.5)
		fmt.Fprintf(w, "%dx%d | %d holes | Card %d", card.Width, card.Height, card.CountHoles(), card.Number)
		fmt.Fprintf(w, "</text>\n")
	}

	// Close SVG
	fmt.Fprintf(w, "</svg>\n")

	return nil
}

// drawGrid draws a grid for alignment
func (e *SVGExporter) drawGrid(w io.Writer, card *Card, widthPx, heightPx float64) {
	startX := CardPadding * MMToPixel
	startY := (CardPadding + TextHeight) * MMToPixel
	endX := startX + float64(card.Width-1)*e.HoleSpacing*e.Scale*MMToPixel
	endY := startY + float64(card.Height-1)*e.HoleSpacing*e.Scale*MMToPixel

	fmt.Fprintf(w, `  <g id="grid" stroke="lightgray" stroke-width="0.5" opacity="0.3">`)
	fmt.Fprintf(w, "\n")

	// Vertical lines
	for x := 0; x < card.Width; x++ {
		cx := startX + float64(x)*e.HoleSpacing*e.Scale*MMToPixel
		fmt.Fprintf(w, `    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>`,
			cx, startY, cx, endY)
		fmt.Fprintf(w, "\n")
	}

	// Horizontal lines
	for y := 0; y < card.Height; y++ {
		cy := startY + float64(y)*e.HoleSpacing*e.Scale*MMToPixel
		fmt.Fprintf(w, `    <line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"/>`,
			startX, cy, endX, cy)
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "  </g>\n\n")
}

// ExportCards exports multiple cards to a single SVG file with all cards arranged vertically
func (e *SVGExporter) ExportCards(cards []*Card, w io.Writer) error {
	if len(cards) == 0 {
		return fmt.Errorf("no cards to export")
	}

	// Calculate dimensions for a single card
	cardWidth := float64(cards[0].Width)*e.HoleSpacing*e.Scale + 2*CardPadding
	cardHeight := float64(cards[0].Height)*e.HoleSpacing*e.Scale + 2*CardPadding + TextHeight*2

	// Total dimensions (stack cards vertically with spacing)
	totalWidth := cardWidth
	cardSpacing := 5.0 // mm between cards
	totalHeight := float64(len(cards))*(cardHeight+cardSpacing) - cardSpacing

	// Convert to pixels
	widthPx := totalWidth * MMToPixel
	heightPx := totalHeight * MMToPixel

	// Write SVG header
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%.2fmm" height="%.2fmm" viewBox="0 0 %.2f %.2f">`,
		totalWidth, totalHeight, widthPx, heightPx)
	fmt.Fprintf(w, "\n")

	// Add title and description
	fmt.Fprintf(w, `  <title>Jacquard Loom Punchcards (Set of %d)</title>`, len(cards))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, `  <desc>Complete set of %d punchcards for Jacquard weaving</desc>`, len(cards))
	fmt.Fprintf(w, "\n\n")

	// Background
	fmt.Fprintf(w, `  <rect width="100%%" height="100%%" fill="white"/>`)
	fmt.Fprintf(w, "\n\n")

	// Export each card in a group, stacked vertically
	for i, card := range cards {
		offsetY := float64(i) * (cardHeight + cardSpacing) * MMToPixel

		fmt.Fprintf(w, `  <g id="card-%d" transform="translate(0, %.2f)">`, card.Number, offsetY)
		fmt.Fprintf(w, "\n")

		// Render the card content directly (without SVG wrapper)
		e.renderCardContent(w, card, cardWidth*MMToPixel, cardHeight*MMToPixel)

		fmt.Fprintf(w, "  </g>\n\n")
	}

	// Close SVG
	fmt.Fprintf(w, "</svg>\n")

	return nil
}

// renderCardContent renders the content of a card (without SVG wrapper)
func (e *SVGExporter) renderCardContent(w io.Writer, card *Card, widthPx, heightPx float64) {
	// Card number at top (with optional title)
	if e.ShowNumbers {
		fmt.Fprintf(w, `    <text x="%.2f" y="%.2f" font-family="monospace" font-size="%.2f" text-anchor="middle" fill="black">`,
			widthPx/2, TextHeight*MMToPixel*0.8, TextHeight*MMToPixel*0.6)

		// Display title with card number in format "Title_name #1/156"
		if e.Title != "" && e.TotalCards > 0 {
			fmt.Fprintf(w, "%s #%d/%d", e.Title, card.Number, e.TotalCards)
		} else if e.TotalCards > 0 {
			fmt.Fprintf(w, "Card #%d/%d", card.Number, e.TotalCards)
		} else {
			fmt.Fprintf(w, "Card #%d", card.Number)
		}

		fmt.Fprintf(w, "</text>\n")
	}

	// Draw grid lines if enabled
	if e.ShowGrid {
		e.drawGrid(w, card, widthPx, heightPx)
	}

	// Draw holes
	startX := CardPadding * MMToPixel
	startY := (CardPadding + TextHeight) * MMToPixel

	for y := 0; y < card.Height; y++ {
		for x := 0; x < card.Width; x++ {
			cx := startX + float64(x)*e.HoleSpacing*e.Scale*MMToPixel
			cy := startY + float64(y)*e.HoleSpacing*e.Scale*MMToPixel

			if card.Matrix[y][x] == 1 {
				// Punched hole - filled circle
				fmt.Fprintf(w, `    <circle cx="%.2f" cy="%.2f" r="%.2f" fill="black"/>`,
					cx, cy, e.HoleRadius*e.Scale*MMToPixel)
				fmt.Fprintf(w, "\n")
			} else {
				// No hole - just a small guide mark
				fmt.Fprintf(w, `    <circle cx="%.2f" cy="%.2f" r="%.2f" fill="none" stroke="lightgray" stroke-width="0.5"/>`,
					cx, cy, e.HoleRadius*e.Scale*MMToPixel*0.3)
				fmt.Fprintf(w, "\n")
			}
		}
	}

	// Card info at bottom
	if e.ShowNumbers {
		infoY := heightPx - TextHeight*MMToPixel*0.3
		fmt.Fprintf(w, `    <text x="%.2f" y="%.2f" font-family="monospace" font-size="%.2f" text-anchor="middle" fill="gray">`,
			widthPx/2, infoY, TextHeight*MMToPixel*0.5)
		fmt.Fprintf(w, "%dx%d | %d holes | Card %d", card.Width, card.Height, card.CountHoles(), card.Number)
		fmt.Fprintf(w, "</text>\n")
	}
}

// SVGTemplate is an alternative template-based approach for generating SVG
const svgCardTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="{{.Width}}mm" height="{{.Height}}mm" viewBox="0 0 {{.WidthPx}} {{.HeightPx}}">
  <title>Jacquard Loom Punchcard #{{.CardNumber}}</title>
  <desc>{{.Description}}</desc>
  <rect width="100%" height="100%" fill="white"/>
  {{if .ShowNumber}}
  <text x="{{.CenterX}}" y="{{.TitleY}}" font-family="monospace" font-size="{{.TitleSize}}" text-anchor="middle" fill="black">Card #{{.CardNumber}}</text>
  {{end}}
  {{range .Holes}}
  <circle cx="{{.X}}" cy="{{.Y}}" r="{{.R}}" fill="{{.Fill}}" {{if .Stroke}}stroke="{{.Stroke}}" stroke-width="{{.StrokeWidth}}"{{end}}/>
  {{end}}
</svg>`

// SVGTemplateData holds data for template rendering
type SVGTemplateData struct {
	Width       float64
	Height      float64
	WidthPx     float64
	HeightPx    float64
	CardNumber  int
	Description string
	ShowNumber  bool
	CenterX     float64
	TitleY      float64
	TitleSize   float64
	Holes       []SVGHole
}

// SVGHole represents a single hole in the SVG
type SVGHole struct {
	X           float64
	Y           float64
	R           float64
	Fill        string
	Stroke      string
	StrokeWidth float64
}

// ExportCardTemplate exports a card using the template approach
func (e *SVGExporter) ExportCardTemplate(card *Card, w io.Writer) error {
	tmpl, err := template.New("svg").Parse(svgCardTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := e.prepareTemplateData(card)
	return tmpl.Execute(w, data)
}

// prepareTemplateData prepares data for template rendering
func (e *SVGExporter) prepareTemplateData(card *Card) *SVGTemplateData {
	cardWidth := float64(card.Width)*e.HoleSpacing*e.Scale + 2*CardPadding
	cardHeight := float64(card.Height)*e.HoleSpacing*e.Scale + 2*CardPadding + TextHeight*2

	widthPx := cardWidth * MMToPixel
	heightPx := cardHeight * MMToPixel

	data := &SVGTemplateData{
		Width:       cardWidth,
		Height:      cardHeight,
		WidthPx:     widthPx,
		HeightPx:    heightPx,
		CardNumber:  card.Number,
		Description: card.GetCardInfo(),
		ShowNumber:  e.ShowNumbers,
		CenterX:     widthPx / 2,
		TitleY:      TextHeight * MMToPixel * 0.8,
		TitleSize:   TextHeight * MMToPixel * 0.6,
		Holes:       make([]SVGHole, 0),
	}

	startX := CardPadding * MMToPixel
	startY := (CardPadding + TextHeight) * MMToPixel

	for y := 0; y < card.Height; y++ {
		for x := 0; x < card.Width; x++ {
			cx := startX + float64(x)*e.HoleSpacing*e.Scale*MMToPixel
			cy := startY + float64(y)*e.HoleSpacing*e.Scale*MMToPixel

			if card.Matrix[y][x] == 1 {
				data.Holes = append(data.Holes, SVGHole{
					X:    cx,
					Y:    cy,
					R:    e.HoleRadius * e.Scale * MMToPixel,
					Fill: "black",
				})
			} else {
				data.Holes = append(data.Holes, SVGHole{
					X:           cx,
					Y:           cy,
					R:           e.HoleRadius * e.Scale * MMToPixel * 0.3,
					Fill:        "none",
					Stroke:      "lightgray",
					StrokeWidth: 0.5,
				})
			}
		}
	}

	return data
}
