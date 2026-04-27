// Package qrcode provides a high-performance QR code generation library in pure Go with zero external dependencies.
package qrcode

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/os-gomod/qrcode/v2/internal/renderer"
)

// ---------------------------------------------------------------------------
// ECLevel — error correction level (canonical type).
// ---------------------------------------------------------------------------

// ECLevel represents the QR code error correction level.
// Higher levels provide more error recovery capability but reduce data capacity.
type ECLevel int

const (
	// LevelL (Low) — ~7% of codewords can be restored.
	LevelL ECLevel = iota
	// LevelM (Medium) — ~15% of codewords can be restored.
	LevelM
	// LevelQ (Quartile) — ~25% of codewords can be restored.
	LevelQ
	// LevelH (High) — ~30% of codewords can be restored.
	LevelH
)

func (l ECLevel) String() string {
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

// ---------------------------------------------------------------------------
// Format — output format (single source of truth).
// ---------------------------------------------------------------------------

// Format represents the output format for a rendered QR code.
type Format int

const (
	// FormatPNG encodes the QR code as a PNG image.
	FormatPNG Format = iota
	// FormatSVG encodes the QR code as an SVG document.
	FormatSVG
	// FormatTerminal encodes the QR code as terminal/ANSI block characters.
	FormatTerminal
	// FormatPDF encodes the QR code as a PDF document.
	FormatPDF
	// FormatBase64 encodes the QR code as a Base64 data URL (PNG-based).
	FormatBase64
)

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

// Extension returns the common file extension for the format (with leading dot).
// Returns ".png" as the default for unknown formats.
func (f Format) Extension() string {
	switch f {
	case FormatPNG:
		return ".png"
	case FormatSVG:
		return ".svg"
	case FormatTerminal:
		return ".txt"
	case FormatPDF:
		return ".pdf"
	case FormatBase64:
		return ".b64"
	default:
		return ".png"
	}
}

// FormatFromPath returns the QR format inferred from a file path's extension.
func FormatFromPath(path string) Format {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return FormatPNG
	case ".svg":
		return FormatSVG
	case ".txt":
		return FormatTerminal
	case ".pdf":
		return FormatPDF
	case ".b64":
		return FormatBase64
	default:
		return FormatPNG
	}
}

// ---------------------------------------------------------------------------
// Renderer convenience re-exports
// ---------------------------------------------------------------------------

// GetRenderer returns the renderer for the given format.
// This re-export allows consumers to access renderers without importing internal packages.
func GetRenderer(f Format) (renderer.Renderer, error) {
	r, err := renderer.GetRenderer(renderer.Format(f))
	if err != nil {
		return nil, fmt.Errorf("get renderer: %w", err)
	}
	return r, nil
}

// NewPNGRenderer returns a new PNG renderer for direct use.
func NewPNGRenderer() *renderer.PNGRenderer {
	return renderer.NewPNGRenderer()
}

// NewSVGRenderer returns a new SVG renderer for direct use.
func NewSVGRenderer() *renderer.SVGRenderer {
	return renderer.NewSVGRenderer()
}

// NewTerminalRenderer returns a new terminal renderer for direct use.
func NewTerminalRenderer() *renderer.TerminalRenderer {
	return renderer.NewTerminalRenderer()
}

// NewPDFRenderer returns a new PDF renderer for direct use.
func NewPDFRenderer() *renderer.PDFRenderer {
	return renderer.NewPDFRenderer()
}

// NewBase64Renderer returns a new Base64 renderer for direct use.
func NewBase64Renderer() *renderer.Base64Renderer {
	return renderer.NewBase64Renderer()
}

// ModuleStyle is a type alias for the internal renderer's ModuleStyle,
// exposed for convenience when using GetRenderer or New*Renderer directly.
type ModuleStyle = renderer.ModuleStyle

// RenderOption is a type alias for the internal renderer's RenderOption.
type RenderOption = renderer.RenderOption

// WithModuleStyle is a convenience re-export.
func WithModuleStyle(style *renderer.ModuleStyle) RenderOption {
	return renderer.WithModuleStyle(style)
}

// WithRoundedModules is a convenience re-export.
func WithRoundedModules(roundness float64) RenderOption {
	return renderer.WithRoundedModules(roundness)
}

// WithCircleModules is a convenience re-export.
func WithCircleModules() RenderOption {
	return renderer.WithCircleModules()
}

// WithGradient is a convenience re-export.
func WithGradient(startColor, endColor string, angle float64) RenderOption {
	return renderer.WithGradient(startColor, endColor, angle)
}

// WithTransparency is a convenience re-export.
func WithTransparency(alpha float64) RenderOption {
	return renderer.WithTransparency(alpha)
}

// DefaultModuleStyle is a convenience re-export.
func DefaultModuleStyle() *renderer.ModuleStyle {
	return renderer.DefaultModuleStyle()
}
