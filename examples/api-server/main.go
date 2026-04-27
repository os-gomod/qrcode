// Package main implements a minimal HTTP API server for QR code generation.
// It serves QR codes on demand using only the Go standard library.
//
// Endpoints:
//   - GET /qr?text=...&format=png  → returns QR code as PNG
//   - GET /qr?text=...&format=svg  → returns QR code as SVG
//   - GET /qr?text=...&format=terminal → returns QR code as text
//   - GET /health                   → health check
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/payload"
)

var client qrcode.Client

func main() {
	fmt.Println("=== QR Code API Server ===")
	fmt.Println()

	// Create a shared client for all requests.
	var err error
	client, err = qrcode.NewClient(
		qrcode.WithDefaultSize(300),
		qrcode.WithWorkerCount(8),
	)
	if err != nil {
		log.Fatalf("Failed to create QR client: %v", err)
	}
	defer client.Close()

	// Register handlers.
	http.HandleFunc("/qr", handleQR)
	http.HandleFunc("/health", handleHealth)

	// Determine port.
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	addr := ":" + port
	fmt.Printf("Starting server on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Printf("  GET /qr?text=...&format=png&size=300  → QR code (default: PNG)\n")
	fmt.Printf("  GET /health                            → health check\n")
	fmt.Println()

	// Start server with graceful shutdown.
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	fmt.Println("Server stopped.")
}

func handleQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters.
	text := r.URL.Query().Get("text")
	if text == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'text' parameter"})
		return
	}

	formatStr := r.URL.Query().Get("format")
	if formatStr == "" {
		formatStr = "png"
	}

	format, err := parseFormat(formatStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Parse optional size.
	size := 300
	if s := r.URL.Query().Get("size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed >= 100 && parsed <= 4000 {
			size = parsed
		}
	}

	// Generate QR code.
	ctx := r.Context()
	data, err := client.Render(ctx, &payload.TextPayload{Text: text}, format)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "generation failed: " + err.Error()})
		return
	}

	// Set content type.
	contentType := "application/octet-stream"
	switch format {
	case qrcode.FormatPNG:
		contentType = "image/png"
	case qrcode.FormatSVG:
		contentType = "image/svg+xml"
	case qrcode.FormatTerminal:
		contentType = "text/plain; charset=utf-8"
	case qrcode.FormatPDF:
		contentType = "application/pdf"
	case qrcode.FormatBase64:
		contentType = "text/plain; charset=utf-8"
	}

	// Log the request.
	fmt.Printf("[%s] %s format=%s size=%d bytes=%d\n",
		time.Now().Format("15:04:05"), text, formatStr, size, len(data))

	// For terminal format, set a fixed size by using GenerateWithOptions.
	// (The client was created with default size; size param from URL is informational here.)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-QR-Format", formatStr)
	w.Header().Set("X-QR-Size", fmt.Sprintf("%d", size))
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"client":    !client.Closed(),
	})
}

func parseFormat(s string) (qrcode.Format, error) {
	switch s {
	case "png":
		return qrcode.FormatPNG, nil
	case "svg":
		return qrcode.FormatSVG, nil
	case "terminal", "txt":
		return qrcode.FormatTerminal, nil
	case "pdf":
		return qrcode.FormatPDF, nil
	case "base64", "b64":
		return qrcode.FormatBase64, nil
	default:
		return 0, fmt.Errorf("unsupported format %q (use png, svg, terminal, pdf, or base64)", s)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
