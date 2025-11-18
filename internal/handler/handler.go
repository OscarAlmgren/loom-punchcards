package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/oscaralmgren/loom-punchcards/internal/image"
	"github.com/oscaralmgren/loom-punchcards/internal/punchcard"
)

// Handler manages HTTP requests for the punchcard application
type Handler struct {
	templates *template.Template
}

// NewHandler creates a new HTTP handler
func NewHandler(templateDir string) (*Handler, error) {
	// Parse templates
	tmpl, err := template.ParseGlob(templateDir + "/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Handler{
		templates: tmpl,
	}, nil
}

// HomeHandler serves the main page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UploadRequest represents the upload request parameters
type UploadRequest struct {
	ColorMode int    // 2, 4, or 8
	Format    string // "svg" or "pdf"
}

// UploadHandler handles image upload and processing
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s (%d bytes)", header.Filename, header.Size)

	// Get color mode parameter
	colorModeStr := r.FormValue("colorMode")
	if colorModeStr == "" {
		colorModeStr = "2" // Default to 2-color
	}
	colorMode, err := strconv.Atoi(colorModeStr)
	if err != nil || image.ValidateColorMode(colorMode) != nil {
		http.Error(w, "Invalid color mode (must be 2, 4, or 8)", http.StatusBadRequest)
		return
	}

	// Get format parameter
	format := r.FormValue("format")
	if format == "" {
		format = "svg" // Default to SVG
	}
	if format != "svg" && format != "pdf" {
		http.Error(w, "Invalid format (must be 'svg' or 'pdf')", http.StatusBadRequest)
		return
	}

	// Process the image
	// Image width should be CardWidth * CardHeight (26 * 8 = 208)
	// Height is auto-calculated from aspect ratio
	processorWidth := punchcard.CardWidth * punchcard.CardHeight
	processor := image.NewProcessor(processorWidth, 0, image.ColorMode(colorMode))

	// Read the file into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Process the image to binary matrix
	matrix, err := processor.Process(bytes.NewReader(fileBytes))
	if err != nil {
		log.Printf("Error processing image: %v", err)
		http.Error(w, fmt.Sprintf("Failed to process image: %v", err), http.StatusBadRequest)
		return
	}

	// Safety check: ensure matrix is not empty
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		log.Printf("Error: processed image resulted in empty matrix")
		http.Error(w, "Failed to process image: resulted in empty matrix", http.StatusBadRequest)
		return
	}

	log.Printf("Processed image to %dx%d matrix", len(matrix[0]), len(matrix))

	// Generate punchcards
	generator := punchcard.NewGenerator()
	cards, err := generator.Generate(matrix)
	if err != nil {
		log.Printf("Error generating punchcards: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate punchcards: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Generated %d punchcards", len(cards))

	// Export based on format
	var output bytes.Buffer
	var contentType string
	var filename string

	if format == "svg" {
		exporter := punchcard.NewSVGExporter()
		err = exporter.ExportCards(cards, &output)
		contentType = "image/svg+xml"
		filename = "punchcards.svg"
	} else {
		// For PDF, we'll export as SVG and let the client handle conversion
		// Or we can use a simple PDF library
		exporter := punchcard.NewSVGExporter()
		err = exporter.ExportCards(cards, &output)
		contentType = "application/pdf"
		filename = "punchcards.pdf"
		// Note: In production, convert SVG to actual PDF here
	}

	if err != nil {
		log.Printf("Error exporting cards: %v", err)
		http.Error(w, "Failed to export punchcards", http.StatusInternalServerError)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(output.Len()))

	// Write output
	_, err = w.Write(output.Bytes())
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// PreviewHandler generates a preview of the punchcards
func (h *Handler) PreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get color mode parameter
	colorModeStr := r.FormValue("colorMode")
	if colorModeStr == "" {
		colorModeStr = "2"
	}
	colorMode, err := strconv.Atoi(colorModeStr)
	if err != nil || image.ValidateColorMode(colorMode) != nil {
		http.Error(w, "Invalid color mode", http.StatusBadRequest)
		return
	}

	// Process the image
	// Image width should be CardWidth * CardHeight (26 * 8 = 208)
	// Height is auto-calculated from aspect ratio
	processorWidth := punchcard.CardWidth * punchcard.CardHeight
	processor := image.NewProcessor(processorWidth, 0, image.ColorMode(colorMode))

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	matrix, err := processor.Process(bytes.NewReader(fileBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process image: %v", err), http.StatusBadRequest)
		return
	}

	// Safety check: ensure matrix is not empty
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		http.Error(w, "Failed to process image: resulted in empty matrix", http.StatusBadRequest)
		return
	}

	// Generate punchcards
	generator := punchcard.NewGenerator()
	cards, err := generator.Generate(matrix)
	if err != nil {
		http.Error(w, "Failed to generate punchcards", http.StatusInternalServerError)
		return
	}

	// Generate preview (first 3 cards only)
	previewCards := cards
	if len(previewCards) > 3 {
		previewCards = cards[:3]
	}

	// Export as SVG for preview
	var output bytes.Buffer
	exporter := punchcard.NewSVGExporter()
	err = exporter.ExportCards(previewCards, &output)
	if err != nil {
		http.Error(w, "Failed to generate preview", http.StatusInternalServerError)
		return
	}

	// Return SVG directly for inline display
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(output.Bytes())
}

// InfoHandler returns information about the generated punchcards
func (h *Handler) InfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get color mode
	colorModeStr := r.FormValue("colorMode")
	if colorModeStr == "" {
		colorModeStr = "2"
	}
	colorMode, err := strconv.Atoi(colorModeStr)
	if err != nil || image.ValidateColorMode(colorMode) != nil {
		colorMode = 2
	}

	// Process the image
	// Image width should be CardWidth * CardHeight (26 * 8 = 208)
	// Height is auto-calculated from aspect ratio
	processorWidth := punchcard.CardWidth * punchcard.CardHeight
	processor := image.NewProcessor(processorWidth, 0, image.ColorMode(colorMode))

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	matrix, err := processor.Process(bytes.NewReader(fileBytes))
	if err != nil {
		http.Error(w, "Failed to process image", http.StatusBadRequest)
		return
	}

	// Safety check: ensure matrix is not empty
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		http.Error(w, "Failed to process image: resulted in empty matrix", http.StatusBadRequest)
		return
	}

	// Generate punchcards
	generator := punchcard.NewGenerator()
	cards, err := generator.Generate(matrix)
	if err != nil {
		http.Error(w, "Failed to generate punchcards", http.StatusInternalServerError)
		return
	}

	// Generate metadata
	metadata := punchcard.GenerateMetadata(cards)

	// Create response
	response := map[string]interface{}{
		"filename":        header.Filename,
		"fileSize":        header.Size,
		"colorMode":       processor.DescribeColorMode(),
		"totalCards":      metadata.TotalCards,
		"cardDimensions":  fmt.Sprintf("%dx%d", metadata.CardWidth, metadata.CardHeight),
		"totalRows":       metadata.TotalRows,
		"averageDensity":  fmt.Sprintf("%.1f%%", metadata.AverageDensity),
		"holesPerCard":    metadata.HolesPerCard,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthHandler provides a health check endpoint
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "Jacquard Loom Punchcard Generator",
	})
}
