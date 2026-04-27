// Package main demonstrates concurrent QR code generation using both
// client.Batch() and batch.Processor with ProcessWithStats.
// It also shows context.WithTimeout for deadline control.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/batch"
	"github.com/os-gomod/qrcode/v2/payload"
)

func main() {
	fmt.Println("=== Concurrent QR Code Generation Example ===")
	fmt.Println()

	// --- 1. client.Batch() with 20 payloads ---
	fmt.Println("1. Using client.Batch() with 20 payloads...")

	client, err := qrcode.NewClient(
		qrcode.WithDefaultSize(256),
		qrcode.WithWorkerCount(8),
		qrcode.WithErrorCorrection(qrcode.LevelM),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Build 20 payloads.
	payloads := make([]payload.Payload, 20)
	for i := 0; i < 20; i++ {
		payloads[i] = &payload.TextPayload{
			Text: fmt.Sprintf("Concurrent QR #%02d", i+1),
		}
	}

	start := time.Now()
	qrCodes, err := client.Batch(ctx, payloads)
	batchDuration := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Batch completed with some errors: %v\n", err)
	}

	successCount := 0
	for i, qr := range qrCodes {
		if qr != nil {
			successCount++
			if i < 3 || i >= 18 {
				fmt.Printf("   Item %2d: version=%d, size=%d modules\n", i+1, qr.Version, qr.Size)
			} else if i == 3 {
				fmt.Printf("   ... (%d more items)\n", 20-6)
			}
		}
	}
	fmt.Printf("   Generated %d/20 QR codes in %v\n\n", successCount, batchDuration)

	// --- 2. batch.Processor with ProcessWithStats ---
	fmt.Println("2. Using batch.Processor with ProcessWithStats...")

	items := make([]batch.Item, 20)
	for i := 0; i < 20; i++ {
		items[i] = batch.Item{
			ID:   fmt.Sprintf("item_%02d", i+1),
			Data: fmt.Sprintf("Processor item #%02d", i+1),
		}
	}

	proc := batch.NewProcessor(
		client,
		batch.WithBatchConcurrency(6),
		batch.WithBatchFormat(qrcode.FormatPNG),
	)

	start2 := time.Now()
	results, stats, err := proc.ProcessWithStats(ctx, items)
	procDuration := time.Since(start2)

	if err != nil {
		fmt.Fprintf(os.Stderr, "   Batch errors: %v\n", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("   %-12s  %-8s  %s\n", "ID", "Status", "Data Size")
	fmt.Println(strings.Repeat("-", 60))
	for _, r := range results {
		status := "OK"
		dataSize := "-"
		if r.Err != nil {
			status = "FAIL"
		} else if r.Data != nil {
			dataSize = fmt.Sprintf("%d bytes", len(r.Data))
		}
		fmt.Printf("   %-12s  %-8s  %s\n", r.ID, status, dataSize)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("   Aggregate Statistics:")
	fmt.Printf("   Total:       %d\n", stats.Total)
	fmt.Printf("   Succeeded:   %d\n", stats.Succeeded)
	fmt.Printf("   Failed:      %d\n", stats.Failed)
	fmt.Printf("   Total time:  %v\n", stats.TotalTime)
	fmt.Printf("   Avg per item: %v\n", stats.AvgTime)
	fmt.Printf("   Min per item: %v\n", stats.MinTime)
	fmt.Printf("   Max per item: %v\n", stats.MaxTime)
	fmt.Printf("   Wall clock:  %v\n", procDuration)
	fmt.Println(strings.Repeat("-", 60))

	// --- 3. Context with timeout ---
	fmt.Println()
	fmt.Println("3. Demonstrating context.WithTimeout (5s deadline)...")

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	timeoutStart := time.Now()
	timeoutPayloads := make([]payload.Payload, 5)
	for i := 0; i < 5; i++ {
		timeoutPayloads[i] = &payload.TextPayload{
			Text: fmt.Sprintf("Timeout test #%d", i+1),
		}
	}

	timeoutResults, err := client.Batch(timeoutCtx, timeoutPayloads)
	timeoutDuration := time.Since(timeoutStart)

	if err != nil {
		fmt.Printf("   Completed with errors (may include timeout): %v\n", err)
	}

	timeoutOK := 0
	for _, qr := range timeoutResults {
		if qr != nil {
			timeoutOK++
		}
	}
	fmt.Printf("   %d/5 completed before deadline of 5s (took %v)\n", timeoutOK, timeoutDuration)

	fmt.Println()
	fmt.Println("=== All done! ===")
}
