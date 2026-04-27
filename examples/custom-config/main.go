// Package main demonstrates custom configuration options for QR code generation.
// It shows how to use NewClient with options and the Builder API to create
// QR codes with custom size, error correction, colors, and quiet zones.
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
	fmt.Println("=== Custom Configuration Example ===")
	fmt.Println()

	ctx := context.Background()

	// --- 1. NewClient with options ---
	fmt.Println("1. Using NewClient() with custom options...")

	client, err := qrcode.NewClient(
		qrcode.WithDefaultSize(400),
		qrcode.WithErrorCorrection(qrcode.LevelH),
		qrcode.WithQuietZone(8),
		qrcode.WithForegroundColor("#1a1a2e"),
		qrcode.WithBackgroundColor("#e0e0e0"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	p := &payload.TextPayload{Text: "Custom config via NewClient"}
	pngData, err := client.Render(ctx, p, qrcode.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to render: %v", err)
	}
	if err := os.WriteFile("custom_newclient.png", pngData, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("   Size: 400px, EC: H, QuietZone: 8, FG: #1a1a2e, BG: #e0e0e0\n")
	fmt.Printf("   Saved custom_newclient.png (%d bytes)\n\n", len(pngData))

	// --- 2. Builder API ---
	fmt.Println("2. Using Builder API (fluent interface)...")

	builderClient, err := qrcode.NewBuilder().
		Size(500).
		ErrorCorrection(qrcode.LevelQ).
		Margin(6).
		ForegroundColor("#ff6600").
		BackgroundColor("#fffbe6").
		Build()
	if err != nil {
		log.Fatalf("Failed to build client: %v", err)
	}
	defer builderClient.Close()

	p2 := &payload.TextPayload{Text: "Built with Builder API!"}
	pngData2, err := builderClient.Render(ctx, p2, qrcode.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to render: %v", err)
	}
	if err := os.WriteFile("custom_builder.png", pngData2, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("   Size: 500px, EC: Q, Margin: 6, FG: #ff6600, BG: #fffbe6\n")
	fmt.Printf("   Saved custom_builder.png (%d bytes)\n\n", len(pngData2))

	// --- 3. Builder SVG output ---
	fmt.Println("3. Builder generating SVG...")

	svgData, err := builderClient.Render(ctx, p2, qrcode.FormatSVG)
	if err != nil {
		log.Fatalf("Failed to render SVG: %v", err)
	}
	if err := os.WriteFile("custom_builder.svg", svgData, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("   Saved custom_builder.svg (%d bytes)\n\n", len(svgData))

	// --- 4. Compare error correction levels ---
	fmt.Println("4. Comparing error correction levels...")
	levels := []struct {
		name  string
		level qrcode.ECLevel
	}{
		{"Low (L)", qrcode.LevelL},
		{"Medium (M)", qrcode.LevelM},
		{"Quartile (Q)", qrcode.LevelQ},
		{"High (H)", qrcode.LevelH},
	}

	for _, l := range levels {
		client, err := qrcode.NewClient(
			qrcode.WithDefaultSize(300),
			qrcode.WithErrorCorrection(l.level),
		)
		if err != nil {
			log.Fatalf("Failed to create client for %s: %v", l.name, err)
		}
		data, err := client.Render(ctx, &payload.TextPayload{Text: "EC Level " + l.name}, qrcode.FormatPNG)
		client.Close()
		if err != nil {
			log.Fatalf("Failed to render for %s: %v", l.name, err)
		}
		filename := fmt.Sprintf("ec_%s.png", l.level)
		if err := os.WriteFile(filename, data, 0o644); err != nil {
			log.Fatalf("Failed to write %s: %v", filename, err)
		}
		fmt.Printf("   %-12s → %s (%d bytes)\n", l.name, filename, len(data))
	}

	// --- 5. MustBuild (panic on error) ---
	fmt.Println()
	fmt.Println("5. Using MustBuild (panic variant)...")
	mustClient := qrcode.NewBuilder().
		Size(300).
		MustBuild()
	defer mustClient.Close()

	mustData, err := mustClient.Render(ctx, &payload.TextPayload{Text: "MustBuild works!"}, qrcode.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to render: %v", err)
	}
	if err := os.WriteFile("custom_mustbuild.png", mustData, 0o644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("   Saved custom_mustbuild.png (%d bytes)\n", len(mustData))

	fmt.Println()
	fmt.Println("=== All done! ===")
}
