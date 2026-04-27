// Package main demonstrates batch QR code generation using the batch.Processor.
// It generates 10 QR codes concurrently, saves them to an output directory,
// and displays timing statistics.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/batch"
	"github.com/os-gomod/qrcode/v2/payload"
)

func main() {
	fmt.Println("=== Batch QR Code Generation Example ===")
	fmt.Println()

	ctx := context.Background()

	// Create output directory.
	outputDir := "batch_output"
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create a client.
	client, err := qrcode.NewClient(
		qrcode.WithDefaultSize(256),
		qrcode.WithErrorCorrection(qrcode.LevelM),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Prepare 10 batch items with different payload types.
	items := []batch.Item{
		{ID: "01_text", Data: "Hello, World!"},
		{ID: "02_url", Payload: &payload.URLPayload{URL: "https://github.com/os-gomod/qrcode", Title: "QR Code Lib"}},
		{ID: "03_wifi", Payload: &payload.WiFiPayload{SSID: "MyNetwork", Password: "s3cret!", Encryption: "WPA"}},
		{ID: "04_sms", Payload: &payload.SMSPayload{Phone: "+1234567890", Message: "Hi from QR!"}},
		{ID: "05_email", Payload: &payload.EmailPayload{To: "user@example.com", Subject: "Hello", Body: "QR code email"}},
		{ID: "06_geo", Payload: &payload.GeoPayload{Latitude: 40.7128, Longitude: -74.0060}},
		{ID: "07_vcard", Payload: &payload.VCardPayload{FirstName: "Jane", LastName: "Doe", Phone: "+15551234567", Email: "jane@example.com"}},
		{ID: "08_text", Data: "Batch processing is efficient!"},
		{ID: "09_text", Data: "Concurrent QR generation"},
		{ID: "10_url", Payload: &payload.URLPayload{URL: "https://golang.org", Title: "Go Language"}},
	}

	fmt.Printf("Processing %d batch items...\n\n", len(items))

	// Create a batch processor with concurrency and output format.
	proc := batch.NewProcessor(
		client,
		batch.WithBatchConcurrency(4),
		batch.WithBatchFormat(qrcode.FormatPNG),
	)

	// Process with stats.
	results, stats, err := proc.ProcessWithStats(ctx, items)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Batch completed with errors: %v\n", err)
	}

	// Print per-item results.
	fmt.Println("Per-item results:")
	fmt.Println(strings.Repeat("-", 65))
	fmt.Printf("%-12s  %-10s  %-20s  %s\n", "ID", "Status", "Output Path", "Size")
	fmt.Println(strings.Repeat("-", 65))

	succeeded := 0
	failed := 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("%-12s  %-10s  %-20s  %s\n", r.ID, "FAIL", "-", r.Err)
			failed++
			continue
		}
		sizeStr := "-"
		if r.Data != nil {
			sizeStr = fmt.Sprintf("%d bytes", len(r.Data))
		}
		pathStr := "-"
		if r.Path != "" {
			pathStr = filepath.Base(r.Path)
		}
		fmt.Printf("%-12s  %-10s  %-20s  %s\n", r.ID, "OK", pathStr, sizeStr)
		succeeded++
	}

	// Print aggregate stats.
	fmt.Println()
	fmt.Println(strings.Repeat("-", 65))
	fmt.Println("Batch Statistics:")
	fmt.Printf("  Total:      %d\n", stats.Total)
	fmt.Printf("  Succeeded:  %d\n", succeeded)
	fmt.Printf("  Failed:     %d\n", failed)
	fmt.Printf("  Total time: %v\n", stats.TotalTime)
	fmt.Printf("  Avg time:   %v\n", stats.AvgTime)
	fmt.Printf("  Min time:   %v\n", stats.MinTime)
	fmt.Printf("  Max time:   %v\n", stats.MaxTime)

	// Also demonstrate SaveToDir.
	fmt.Println()
	fmt.Println("Saving batch to directory via SaveToDir...")
	dirResults, err := proc.SaveToDir(ctx, items, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SaveToDir completed with errors: %v\n", err)
	}
	for _, r := range dirResults {
		if r.Err == nil && r.Path != "" {
			fmt.Printf("  %s\n", r.Path)
		}
	}

	fmt.Println()
	fmt.Printf("=== Batch complete! %d files saved to %s ===\n", succeeded, outputDir)
}
