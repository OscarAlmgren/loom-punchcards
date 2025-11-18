package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oscaralmgren/loom-punchcards/internal/handler"
)

const (
	defaultPort        = "8080"
	defaultTemplateDir = "web/templates"
	defaultStaticDir   = "web/static"
)

func main() {
	// Command-line flags
	port := flag.String("port", getEnv("PORT", defaultPort), "HTTP server port")
	templateDir := flag.String("templates", defaultTemplateDir, "Templates directory")
	staticDir := flag.String("static", defaultStaticDir, "Static files directory")
	flag.Parse()

	// Print banner
	printBanner()

	// Initialize handler
	h, err := handler.NewHandler(*templateDir)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	// Set up routes
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir(*staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	mux.HandleFunc("/", h.HomeHandler)
	mux.HandleFunc("/upload", h.UploadHandler)
	mux.HandleFunc("/preview", h.PreviewHandler)
	mux.HandleFunc("/info", h.InfoHandler)
	mux.HandleFunc("/upload-text", h.UploadTextHandler)
	mux.HandleFunc("/preview-text", h.PreviewTextHandler)
	mux.HandleFunc("/info-text", h.InfoTextHandler)
	mux.HandleFunc("/health", h.HealthHandler)

	// Start server
	addr := ":" + *port
	log.Printf("Starting Jacquard Loom Punchcard Generator on http://localhost%s", addr)
	log.Printf("Template directory: %s", *templateDir)
	log.Printf("Static directory: %s", *staticDir)
	log.Printf("Ready to generate punchcards! 🧵")

	if err := http.ListenAndServe(addr, logRequest(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// logRequest is a middleware that logs HTTP requests
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// printBanner prints an ASCII art banner
func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║   🧵  JACQUARD LOOM PUNCHCARD GENERATOR  🧵                      ║
║                                                                   ║
║   Convert images to historical weaving punchcards                ║
║   For use in Jacquard silk weaving looms                         ║
║                                                                   ║
║   Inspired by Joseph Marie Jacquard's 1801 invention             ║
║   that revolutionized textile manufacturing                      ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
`
	log.Print(banner)
}

// ensureDirectories creates necessary directories if they don't exist
func ensureDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// getAbsPath returns the absolute path, resolving relative paths from the executable location
func getAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Get executable directory
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)

	return filepath.Join(exPath, path), nil
}
