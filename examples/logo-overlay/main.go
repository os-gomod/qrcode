// Package main demonstrates logo overlay on QR codes.
// It creates a simple PNG logo programmatically, generates a QR code with
// LevelH error correction, applies the logo using WithLogo(), and saves the result.
// It also shows manual logo overlay using the logo package.
package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/logo"
	"github.com/os-gomod/qrcode/v2/payload"
)

func main() {
	fmt.Println("=== Logo Overlay Example ===")
	fmt.Println()

	ctx := context.Background()

	// Step 1: Create a simple PNG logo programmatically (100x100 blue square with border).
	fmt.Println("1. Creating sample logo image...")
	logoPath := "sample_logo.png"
	if err := createSampleLogo(logoPath, 100, color.RGBA{R: 30, G: 100, B: 220, A: 255}); err != nil {
		log.Fatalf("Failed to create sample logo: %v", err)
	}
	fmt.Printf("   Saved %s\n", logoPath)

	// Step 2: Generate QR code with logo overlay using WithLogo option.
	fmt.Println("2. Generating QR code with logo overlay (WithLogo option)...")

	client, err := qrcode.NewClient(
		qrcode.WithDefaultSize(400),
		qrcode.WithErrorCorrection(qrcode.LevelH),
		qrcode.WithQuietZone(4),
		qrcode.WithLogo(logoPath, 0.20),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	p := &payload.TextPayload{Text: "https://github.com/os-gomod/qrcode"}
	logoData, err := client.Render(ctx, p, qrcode.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to render QR with logo: %v", err)
	}
	if err := os.WriteFile("qr_with_logo.png", logoData, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("   Saved qr_with_logo.png (%d bytes)\n\n", len(logoData))

	// Step 3: Manual logo overlay using the logo package.
	fmt.Println("3. Manual logo overlay using logo package...")

	// Generate a plain QR code (without logo).
	plainClient, err := qrcode.NewClient(
		qrcode.WithDefaultSize(400),
		qrcode.WithErrorCorrection(qrcode.LevelH),
		qrcode.WithQuietZone(4),
		qrcode.WithForegroundColor("#000000"),
		qrcode.WithBackgroundColor("#FFFFFF"),
	)
	if err != nil {
		log.Fatalf("Failed to create plain client: %v", err)
	}
	defer plainClient.Close()

	// Generate the raw QR matrix.
	qr, err := plainClient.Generate(ctx, p)
	if err != nil {
		log.Fatalf("Failed to generate QR matrix: %v", err)
	}
	fmt.Printf("   QR version: %d, size: %d modules\n", qr.Version, qr.Size)

	// Render to PNG using the PNG renderer.
	renderer := qrcode.NewPNGRenderer()

	renderOpts := []qrcode.RenderOption{
		qrcode.WithRoundedModules(0.0),
	}

	plainBytes, err := renderer.Render(ctx, qr, renderOpts...)
	if err != nil {
		log.Fatalf("Failed to render plain QR: %v", err)
	}

	// Decode the plain PNG.
	plainImg, err := png.Decode(bytes.NewReader(plainBytes))
	if err != nil {
		log.Fatalf("Failed to decode plain PNG: %v", err)
	}

	// Load and resize the logo.
	logoProc := logo.New(logoPath, 0.20)
	logoImg, err := logoProc.Load()
	if err != nil {
		log.Fatalf("Failed to load logo: %v", err)
	}

	// Resize the logo to match QR module count.
	resizedLogo := logo.ResizeLogo(logoImg, qr.Size, 0.20)

	// Overlay the logo onto the QR code.
	finalImg := logo.OverlayLogo(plainImg, resizedLogo, 0)

	// Encode and save.
	finalBytes, err := logo.EncodePNG(finalImg)
	if err != nil {
		log.Fatalf("Failed to encode final PNG: %v", err)
	}
	if err := os.WriteFile("qr_manual_overlay.png", finalBytes, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("   Saved qr_manual_overlay.png (%d bytes)\n", len(finalBytes))

	// Step 4: Show logo utilities.
	fmt.Println()
	fmt.Println("4. Logo package utilities...")
	fmt.Printf("   Supported formats: %v\n", logo.SupportedFormats())
	fmt.Printf("   Is .png supported:  %v\n", logo.IsSupportedFormat(".png"))
	fmt.Printf("   Is .bmp supported:  %v\n", logo.IsSupportedFormat(".bmp"))

	if err := logo.Validate(logoPath); err != nil {
		log.Fatalf("Logo validation failed: %v", err)
	}
	fmt.Printf("   Logo validation:    OK\n")

	logoSize := logo.LogoSize(400, 0.20)
	fmt.Printf("   Logo pixel size:    %dpx (400px QR * 0.20 ratio)\n", logoSize)

	fmt.Println()
	fmt.Println("=== All done! ===")
}

// createSampleLogo creates a simple PNG logo file (blue square with a lighter border).
func createSampleLogo(path string, size int, bgColor color.Color) error {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with background color.
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	// Draw a white border.
	borderWidth := size / 10
	borderColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Top border.
	for y := 0; y < borderWidth; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, borderColor)
		}
	}
	// Bottom border.
	for y := size - borderWidth; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, borderColor)
		}
	}
	// Left border.
	for y := 0; y < size; y++ {
		for x := 0; x < borderWidth; x++ {
			img.Set(x, y, borderColor)
		}
	}
	// Right border.
	for y := 0; y < size; y++ {
		for x := size - borderWidth; x < size; x++ {
			img.Set(x, y, borderColor)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}
