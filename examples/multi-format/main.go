// Package main generates the same QR code in all 5 supported formats:
// PNG, SVG, Terminal, PDF, and Base64. It saves each to a file,
// prints content types and file sizes, and demonstrates Base64 data URI usage.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/payload"
)

func main() {
	fmt.Println("=== Multi-Format QR Code Generation Example ===")
	fmt.Println()

	ctx := context.Background()

	// Create client with default settings.
	client, err := qrcode.NewClient(
		qrcode.WithDefaultSize(300),
		qrcode.WithErrorCorrection(qrcode.LevelM),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	text := "https://github.com/os-gomod/qrcode"
	p := &payload.TextPayload{Text: text}

	fmt.Printf("Encoding: %s\n", text)
	fmt.Println()

	// Define all 5 formats with their metadata.
	formats := []struct {
		format      qrcode.Format
		name        string
		filename    string
		contentType string
	}{
		{qrcode.FormatPNG, "PNG", "multi_qr.png", "image/png"},
		{qrcode.FormatSVG, "SVG", "multi_qr.svg", "image/svg+xml"},
		{qrcode.FormatTerminal, "Terminal", "multi_qr.txt", "text/plain; charset=utf-8"},
		{qrcode.FormatPDF, "PDF", "multi_qr.pdf", "application/pdf"},
		{qrcode.FormatBase64, "Base64", "multi_qr.b64", "text/plain; charset=utf-8"},
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("  %-12s %-25s %-30s %s\n", "Format", "File", "Content-Type", "Size")
	fmt.Println(strings.Repeat("-", 70))

	// Generate in each format and save.
	for _, f := range formats {
		data, err := client.Render(ctx, p, f.format)
		if err != nil {
			log.Fatalf("Failed to render %s: %v", f.name, err)
		}

		if err := os.WriteFile(f.filename, data, 0o644); err != nil {
			log.Fatalf("Failed to write %s: %v", f.filename, err)
		}

		fmt.Printf("  %-12s %-25s %-30s %d bytes\n",
			f.name, f.filename, f.contentType, len(data))
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Println()

	// --- Base64 data URI demonstration ---
	fmt.Println("Base64 Data URI usage:")
	fmt.Println()

	// Read back the base64 file.
	b64Data, err := os.ReadFile("multi_qr.b64")
	if err != nil {
		log.Fatalf("Failed to read base64 file: %v", err)
	}

	fmt.Println("The Base64 output can be used as a data URI in HTML:")
	fmt.Println()
	fmt.Println("  <img src=\"data:image/png;base64,<BASE64_DATA>\" width=\"300\" />")
	fmt.Println()
	fmt.Printf("  Base64 data length: %d bytes\n", len(b64Data))
	fmt.Printf("  First 80 chars:     %s...\n", truncate(string(b64Data), 80))

	fmt.Println()

	// --- Terminal QR code preview ---
	fmt.Println("Terminal QR code preview:")
	termData, err := os.ReadFile("multi_qr.txt")
	if err != nil {
		log.Fatalf("Failed to read terminal file: %v", err)
	}
	fmt.Println(string(termData))

	// --- Also demonstrate the Save() method with auto format detection ---
	fmt.Println("Using Save() method with format auto-detection from extension:")
	savePath := "multi_saved.svg"
	if err := client.Save(ctx, p, savePath); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	info, _ := os.Stat(savePath)
	fmt.Printf("  Saved %s (%d bytes) - format auto-detected from .svg extension\n", savePath, info.Size())

	fmt.Println()
	fmt.Println("=== All done! ===")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
