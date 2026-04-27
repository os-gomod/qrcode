// Package main demonstrates basic QR code generation using the qrcode library.
// It creates a client, renders a text QR code in PNG, SVG, and terminal formats,
// saves each to a file, and prints the resulting file sizes.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/payload"
)

func main() {
	fmt.Println("=== Basic QR Code Generation Example ===")
	fmt.Println()

	// Create a new client with default settings.
	client, err := qrcode.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Define the data to encode.
	text := "https://github.com/os-gomod/qrcode"
	p := &payload.TextPayload{Text: text}

	fmt.Printf("Encoding text: %s\n\n", text)

	// --- Generate PNG ---
	fmt.Println("1. Generating PNG...")
	pngData, err := client.Render(ctx, p, qrcode.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to render PNG: %v", err)
	}
	pngPath := "output_basic.png"
	if err := os.WriteFile(pngPath, pngData, 0o644); err != nil {
		log.Fatalf("Failed to write PNG file: %v", err)
	}
	fmt.Printf("   Saved %s (%d bytes)\n", pngPath, len(pngData))

	// --- Generate SVG ---
	fmt.Println("2. Generating SVG...")
	svgData, err := client.Render(ctx, p, qrcode.FormatSVG)
	if err != nil {
		log.Fatalf("Failed to render SVG: %v", err)
	}
	svgPath := "output_basic.svg"
	if err := os.WriteFile(svgPath, svgData, 0o644); err != nil {
		log.Fatalf("Failed to write SVG file: %v", err)
	}
	fmt.Printf("   Saved %s (%d bytes)\n", svgPath, len(svgData))

	// --- Generate Terminal ---
	fmt.Println("3. Generating Terminal output...")
	termData, err := client.Render(ctx, p, qrcode.FormatTerminal)
	if err != nil {
		log.Fatalf("Failed to render terminal: %v", err)
	}
	termPath := "output_basic.txt"
	if err := os.WriteFile(termPath, termData, 0o644); err != nil {
		log.Fatalf("Failed to write terminal file: %v", err)
	}
	fmt.Printf("   Saved %s (%d bytes)\n", termPath, len(termData))

	// --- Print terminal QR code to stdout ---
	fmt.Println()
	fmt.Println("Terminal QR code:")
	fmt.Println(string(termData))

	// --- Use Save() convenience method ---
	fmt.Println("4. Using Save() convenience method...")
	savePath := "output_saved.png"
	if err := client.Save(ctx, p, savePath); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}
	fmt.Printf("   Saved %s (%d bytes)\n", savePath, info.Size())

	// --- Use Quick helpers ---
	fmt.Println("5. Using Quick() helper function...")
	quickData, err := qrcode.Quick("Quick Hello!", 256)
	if err != nil {
		log.Fatalf("Failed to quick generate: %v", err)
	}
	fmt.Printf("   Quick PNG generated (%d bytes)\n", len(quickData))

	fmt.Println()
	fmt.Println("=== All done! ===")
}
