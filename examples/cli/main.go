// Package main implements a simple CLI tool for QR code generation.
// It uses only the standard library flag package (no external dependencies).
//
// Usage:
//
//	qrcode-cli <text> [options]
//	  -size int     (default 300)
//	  -format string (png|svg|pdf|terminal, default png)
//	  -output string (file path, default stdout for png/svg/pdf, terminal for terminal)
//	  -ec string    (L|M|Q|H, default M)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	qrcode "github.com/os-gomod/qrcode/v2"
	"github.com/os-gomod/qrcode/v2/payload"
)

func main() {
	// Define flags.
	size := flag.Int("size", 300, "QR code size in pixels (100-4000)")
	format := flag.String("format", "png", "Output format: png, svg, pdf, terminal")
	output := flag.String("output", "", "Output file path (default: stdout)")
	ec := flag.String("ec", "M", "Error correction level: L, M, Q, H")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: qrcode-cli <text> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Generate QR codes from the command line.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  qrcode-cli 'Hello World'\n")
		fmt.Fprintf(os.Stderr, "  qrcode-cli 'https://example.com' -size 500 -format svg\n")
		fmt.Fprintf(os.Stderr, "  qrcode-cli 'WiFi:MyNet' -ec H -output qr.png\n")
	}

	flag.Parse()

	// Get the text argument.
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: text argument is required")
		flag.Usage()
		os.Exit(1)
	}
	text := args[0]

	// Parse error correction level.
	ecLevel, err := parseECLevel(*ec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse format.
	qrFormat, err := parseFormat(*format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate size.
	if *size < 100 || *size > 4000 {
		fmt.Fprintf(os.Stderr, "Error: size must be between 100 and 4000, got %d\n", *size)
		os.Exit(1)
	}

	// Create client.
	client, err := qrcode.NewClient(
		qrcode.WithDefaultSize(*size),
		qrcode.WithErrorCorrection(ecLevel),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Generate QR code.
	p := &payload.TextPayload{Text: text}
	data, err := client.Render(ctx, p, qrFormat)
	if err != nil {
		log.Fatalf("Failed to generate QR code: %v", err)
	}

	// Output result.
	if *output != "" {
		// Write to file.
		if err := os.WriteFile(*output, data, 0o644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Fprintf(os.Stderr, "QR code saved to %s (%d bytes, %s, EC-%s, %dpx)\n",
			*output, len(data), *format, *ec, *size)
	} else {
		// Write to stdout.
		if qrFormat == qrcode.FormatTerminal {
			os.Stdout.Write(data)
		} else {
			os.Stdout.Write(data)
		}
	}
}

func parseECLevel(s string) (qrcode.ECLevel, error) {
	switch s {
	case "L", "l":
		return qrcode.LevelL, nil
	case "M", "m":
		return qrcode.LevelM, nil
	case "Q", "q":
		return qrcode.LevelQ, nil
	case "H", "h":
		return qrcode.LevelH, nil
	default:
		return 0, fmt.Errorf("invalid error correction level %q (use L, M, Q, or H)", s)
	}
}

func parseFormat(s string) (qrcode.Format, error) {
	switch s {
	case "png":
		return qrcode.FormatPNG, nil
	case "svg":
		return qrcode.FormatSVG, nil
	case "pdf":
		return qrcode.FormatPDF, nil
	case "terminal", "txt":
		return qrcode.FormatTerminal, nil
	default:
		return 0, fmt.Errorf("unsupported format %q (use png, svg, pdf, or terminal)", s)
	}
}
