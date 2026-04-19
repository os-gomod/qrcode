package renderer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/os-gomod/qrcode/encoding"
)

// SVGRenderer renders QR codes as SVG (Scalable Vector Graphics) documents.
//
// The renderer produces a standalone SVG file with an XML declaration,
// viewBox, and properly namespaced elements. Module shapes other than
// "square" (rounded, circle, diamond) are supported via the SVG helper
// functions in svg_helpers.go. Gradient fills are supported via the
// SVGGradientDef helper.
//
// SVG output is resolution-independent and ideal for use in web pages,
// print media, or any context where scaling without quality loss is needed.
//
// Example:
//
//	r := renderer.NewSVGRenderer()
//	err := r.Render(ctx, qr, os.Stdout,
//	    renderer.WithForegroundColor("#1A1A2E"),
//	    renderer.WithBackgroundColor("#F5F5F5"),
//	)
type SVGRenderer struct{}

// NewSVGRenderer creates a new SVGRenderer. The returned renderer is
// stateless and safe for concurrent use.
func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{}
}

// Type returns the format identifier "svg".
func (r *SVGRenderer) Type() string { return "svg" }

// ContentType returns the MIME type "image/svg+xml".
func (r *SVGRenderer) ContentType() string { return "image/svg+xml" }

// Render writes the QR code as an SVG document to w using the given render options.
//
// Each dark module is rendered as an SVG <rect> element positioned within
// a viewBox that includes the quiet zone. The canvas uses a fixed module
// size of 10 units. Foreground and background colors are validated as
// "#RRGGBB" hex strings before rendering.
func (r *SVGRenderer) Render(_ context.Context, qr *encoding.QRCode, w io.Writer, opts ...RenderOption) error {
	cfg := ApplyOptions(opts...)
	_, _, _, err := ParseHexColor(cfg.ForegroundColor)
	if err != nil {
		return fmt.Errorf("invalid foreground color: %w", err)
	}
	_, _, _, err = ParseHexColor(cfg.BackgroundColor)
	if err != nil {
		return fmt.Errorf("invalid background color: %w", err)
	}
	moduleSize := 10
	totalModules := qr.Size + 2*cfg.QuietZone
	canvasSize := totalModules * moduleSize
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
`,
		canvasSize, canvasSize, canvasSize, canvasSize)
	b.WriteString("\n")
	fmt.Fprintf(&b, `  <rect width="%d" height="%d" fill="%s"/>`,
		canvasSize, canvasSize, cfg.BackgroundColor)
	b.WriteString("\n")
	for row := 0; row < qr.Size; row++ {
		for col := 0; col < qr.Size; col++ {
			if qr.Modules[row][col] {
				x := (col + cfg.QuietZone) * moduleSize
				y := (row + cfg.QuietZone) * moduleSize
				fmt.Fprintf(&b, `  <rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
					x, y, moduleSize, moduleSize, cfg.ForegroundColor)
				b.WriteString("\n")
			}
		}
	}
	b.WriteString("</svg>\n")
	_, err = io.WriteString(w, b.String())
	return err
}
