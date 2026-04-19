// Package qrcode provides a high-performance QR code generation library in pure Go
// with zero external dependencies. It supports encoding over 30 payload types including
// URLs, WiFi credentials, vCard contacts, calendar events, social media links, payments,
// and geographic locations. QR codes can be rendered to PNG, SVG, PDF, terminal Unicode
// blocks, or base64-encoded data URIs, with support for custom module shapes, gradient
// fills, transparency, and centered logo overlays.
//
// The library exposes two primary usage patterns. The first is a functional-options
// Generator interface for long-lived, concurrent use:
//
//	     gen, err := qrcode.New(
//	         qrcode.WithErrorCorrection(qrcode.LevelH),
//	         qrcode.WithDefaultSize(400),
//	         qrcode.WithQuietZone(4),
//	     )
//	     if err != nil {
//	         log.Fatal(err)
//	     }
//	     defer gen.Close(ctx)
//
//		payload := payload.URL("https://example.com")
//		qr, err := gen.Generate(ctx, payload)
//
// The second is a fluent Builder API that chains configuration and provides
// one-shot convenience methods:
//
//	png, err := qrcode.NewBuilder().
//	    ErrorCorrection(qrcode.LevelQ).
//	    ForegroundColor("#1a1a2e").
//	    BackgroundColor("#e0e0e0").
//	    Quick("https://example.com", 512)
//
// For fire-and-forget use cases, package-level Quick* functions create and
// discard a generator internally:
//
//	png, err := qrcode.Quick("Hello, world!")
//	svg, err := qrcode.QuickSVG("https://example.com")
//	err := qrcode.QuickFile("Hello, world!", "output.png")
//
// Batch generation is available through both the Generator.Batch method and the
// dedicated batch.Processor type, which supports concurrency control, file output,
// and statistics collection.
package qrcode

import (
	"context"
	"io"

	"github.com/os-gomod/qrcode/encoding"
	qrerrors "github.com/os-gomod/qrcode/errors"
	"github.com/os-gomod/qrcode/payload"
)

// ErrorCorrectionLevel represents the Reed–Solomon error correction level for a QR code.
// Higher levels provide greater damage tolerance but reduce the available data capacity.
// The four standard levels are L (7%), M (15%), Q (25%), and H (30%).
//
// LevelH is recommended when a logo overlay is used, as the additional redundancy
// helps scanners recover the original data despite the obstructed center modules.
type ErrorCorrectionLevel int

const (
	// LevelL recovers approximately 7% of codewords.
	// Suitable for clean environments where the QR code will not be damaged or obscured.
	LevelL ErrorCorrectionLevel = iota
	// LevelM recovers approximately 15% of codewords.
	// The default level and a good balance between capacity and resilience.
	LevelM
	// LevelQ recovers approximately 25% of codewords.
	// Recommended for QR codes that may be partially covered or printed on rough surfaces.
	LevelQ
	// LevelH recovers approximately 30% of codewords.
	// Recommended when using a logo overlay, as the highest redundancy compensates
	// for the obstructed center region.
	LevelH
)

// String returns the human-readable label for the error correction level.
// Returns "L", "M", "Q", or "H" for valid levels; defaults to "M".
func (l ErrorCorrectionLevel) String() string {
	switch l {
	case LevelL:
		return "L"
	case LevelM:
		return "M"
	case LevelQ:
		return "Q"
	case LevelH:
		return "H"
	default:
		return "M"
	}
}

// Format enumerates the supported output formats for rendering QR codes.
//
// Each format maps to a dedicated renderer in the renderer sub-package. The Format
// value is passed to Generator.GenerateToWriter to select the output encoding.
//
//	FormatPNG      // raster PNG image, best for web and print
//	FormatSVG      // vector SVG, best for scalable or styled output
//	FormatTerminal // Unicode block characters for CLI display
//	FormatPDF      // standalone PDF document
//	FormatBase64   // base64-encoded PNG data URI for HTML embedding
type Format int

const (
	// FormatPNG renders the QR code as a PNG image.
	// Uses the standard Go image/png encoder. Supports module-style customization
	// including rounded, circle, and diamond shapes, gradient fills, and transparency.
	FormatPNG Format = iota
	// FormatSVG renders the QR code as a scalable vector graphic.
	// Produces a self-contained SVG document with no external dependencies.
	FormatSVG
	// FormatTerminal renders the QR code as Unicode block characters for terminal output.
	// Uses half-block characters (\u2580, \u2584, \u2588) to achieve near-square
	// module rendering. Supports 24-bit ANSI color codes when a foreground color is set.
	FormatTerminal
	// FormatPDF renders the QR code as a PDF document.
	// Generates a minimal PDF 1.4 file with the QR code centered on the page.
	FormatPDF
	// FormatBase64 renders the QR code as a base64-encoded PNG data URI.
	// Output is a "data:image/png;base64,..." string suitable for embedding in HTML
	// <img> tags or CSS backgrounds.
	FormatBase64
)

// String returns the file-extension style label for the format.
// Returns "png", "svg", "terminal", "pdf", or "base64" for known formats;
// returns "unknown" for undefined values.
func (f Format) String() string {
	switch f {
	case FormatPNG:
		return "png"
	case FormatSVG:
		return "svg"
	case FormatTerminal:
		return "terminal"
	case FormatPDF:
		return "pdf"
	case FormatBase64:
		return "base64"
	default:
		return "unknown"
	}
}

// Generator is the primary interface for creating QR codes. All generation and
// rendering operations are performed through this interface. A Generator is safe
// for concurrent use: its internal configuration is protected by a read-write lock,
// and duplicate generation requests are deduplicated via a singleflight group.
//
// The typical lifecycle is New, one or more Generate/GenerateToWriter calls, and
// Close. After Close, all subsequent method calls return an error with code
// ErrCodeClosed.
//
//	gen, err := qrcode.New(qrcode.WithErrorCorrection(qrcode.LevelH))
//	if err != nil { /* handle error */ }
//	defer gen.Close(ctx)
//
// qr, err := gen.Generate(ctx, payload.URL("https://example.com"))
// err = gen.GenerateToWriter(ctx, p, os.Stdout, qrcode.FormatPNG).
type Generator interface {
	// Generate produces a QR code from the given payload using default options.
	Generate(ctx context.Context, p payload.Payload) (*encoding.QRCode, error)
	// GenerateWithOptions produces a QR code with per-call option overrides.
	GenerateWithOptions(ctx context.Context, p payload.Payload, opts ...Option) (*encoding.QRCode, error)
	// GenerateToWriter renders a QR code in the specified format and writes it to w.
	GenerateToWriter(ctx context.Context, p payload.Payload, w io.Writer, format Format) error
	// Batch generates multiple QR codes concurrently, one per payload.
	Batch(ctx context.Context, payloads []payload.Payload, opts ...Option) ([]*encoding.QRCode, error)
	// Close releases resources held by the generator.
	Close(ctx context.Context) error
	// Closed reports whether the generator has been closed.
	Closed() bool
	// SetOptions updates the generator's default options after construction.
	SetOptions(opts ...Option) error
}

// New creates a new Generator configured with the given options. If no options
// are provided, sensible defaults are applied: error correction level M, automatic
// version sizing, 300px image size, 4-module quiet zone, and 4 concurrent workers.
//
// New validates the configuration and returns a wrapped validation error if any
// option produces an invalid combination (e.g., MinVersion > MaxVersion).
//
//	gen, err := qrcode.New(
//	    qrcode.WithErrorCorrection(qrcode.LevelH),
//	    qrcode.WithDefaultSize(400),
//	    qrcode.WithQuietZone(4),
//	    qrcode.WithLogo("logo.png", 0.25),
//	)
func New(opts ...Option) (Generator, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if err := cfg.Validate(); err != nil {
		return nil, qrerrors.Wrap(qrerrors.ErrCodeValidation, "invalid configuration", err)
	}
	return newGenerator(cfg)
}

// MustNew is like New but panics on error. Useful in package-level variable
// initialization where error handling is not practical.
//
//	var gen = qrcode.MustNew(qrcode.WithDefaultSize(512))
func MustNew(opts ...Option) Generator {
	g, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return g
}
