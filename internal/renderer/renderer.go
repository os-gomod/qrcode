// Package renderer provides modular, pure-rendering QR code output formats.
//
// Every Renderer in this package performs zero file I/O — it receives a
// QR matrix and configuration, and returns rendered bytes.  File persistence
// is handled exclusively by the storage layer (internal/storage).
//
// Renderers are registered in a global map and retrieved via GetRenderer.
package renderer

import (
	"context"
	"fmt"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// Format identifies the output format for a QR code rendering.
// These constants mirror qrcode.Format iota values and are used internally
// for registry dispatch.  External code should use qrcode.Format.
type Format int

const (
	// FormatPNG encodes as a PNG image.
	FormatPNG Format = iota
	// FormatSVG encodes as an SVG document.
	FormatSVG
	// FormatTerminal encodes as terminal/ANSI block characters.
	FormatTerminal
	// FormatPDF encodes as a PDF document.
	FormatPDF
	// FormatBase64 encodes as a Base64 data URL (PNG-based).
	FormatBase64
)

// String returns the lowercase format name.
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

// ContentType returns the MIME type for the format.
func (f Format) ContentType() string {
	switch f {
	case FormatPNG:
		return "image/png"
	case FormatSVG:
		return "image/svg+xml"
	case FormatTerminal:
		return "text/plain"
	case FormatPDF:
		return "application/pdf"
	case FormatBase64:
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// Renderer is the interface that all QR code format renderers must satisfy.
// Implementations are pure — they perform zero file I/O, context
// propagation notwithstanding.  All renderers must be safe for concurrent use.
type Renderer interface {
	// Render encodes the QR code matrix into the specified format and returns
	// the raw bytes.  The opts parameter allows per-call configuration overrides
	// such as size, colors, and module style.
	Render(ctx context.Context, qr *encoding.QRCode, opts ...RenderOption) ([]byte, error)
}

// Compile-time checks: all renderers satisfy the Renderer interface.
var (
	_ Renderer = (*PNGRenderer)(nil)
	_ Renderer = (*SVGRenderer)(nil)
	_ Renderer = (*TerminalRenderer)(nil)
	_ Renderer = (*PDFRenderer)(nil)
	_ Renderer = (*Base64Renderer)(nil)
)

// registry holds the mapping of Format → Renderer.
// It is initialized once at package load and safe for concurrent reads.
var registry = map[Format]Renderer{
	FormatPNG:      NewPNGRenderer(),
	FormatSVG:      NewSVGRenderer(),
	FormatTerminal: NewTerminalRenderer(),
	FormatPDF:      NewPDFRenderer(),
	FormatBase64:   NewBase64Renderer(),
}

// GetRenderer returns the registered Renderer for the given Format.
// Returns an error if the format is not supported.
func GetRenderer(f Format) (Renderer, error) {
	r, ok := registry[f]
	if !ok {
		return nil, fmt.Errorf("renderer: unsupported format %d (%s)", f, f.String())
	}
	return r, nil
}

// RegisterRenderer adds or replaces a Renderer for the given Format.
// This is useful for custom or mock renderers in tests.
func RegisterRenderer(f Format, r Renderer) {
	registry[f] = r
}
