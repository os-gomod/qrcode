package renderer

import (
	"context"
	"fmt"
	"strings"

	"github.com/os-gomod/qrcode/v2/internal/encoding"
)

// SVGRenderer renders QR codes as SVG documents.
// It is safe for concurrent use.
type SVGRenderer struct{}

// NewSVGRenderer creates a new SVGRenderer.
func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{}
}

// Render encodes the QR matrix as an SVG document.
func (*SVGRenderer) Render(_ context.Context, qr *encoding.QRCode, opts ...RenderOption) ([]byte, error) {
	cfg := ApplyOptions(opts...)

	if _, _, _, err := ParseHexColor(cfg.ForegroundColor); err != nil {
		return nil, fmt.Errorf("invalid foreground color: %w", err)
	}
	if _, _, _, err := ParseHexColor(cfg.BackgroundColor); err != nil {
		return nil, fmt.Errorf("invalid background color: %w", err)
	}
	if cfg.ModuleStyle != nil {
		if err := cfg.ModuleStyle.Validate(); err != nil {
			return nil, err
		}
	}

	moduleSize := ScaleSize(qr.Size, cfg.QuietZone, cfg.Width)
	totalModules := qr.Size + 2*cfg.QuietZone
	canvasSize := totalModules * moduleSize

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`,
		canvasSize, canvasSize, canvasSize, canvasSize)
	b.WriteString("\n")
	fmt.Fprintf(&b, `  <rect width="%d" height="%d" fill="%s"/>`,
		canvasSize, canvasSize, cfg.BackgroundColor)
	b.WriteString("\n")

	// Prepend gradient definition when advanced rendering uses gradients.
	if cfg.ModuleStyle.UseAdvanced() && cfg.ModuleStyle.GradientEnabled {
		b.WriteString(SVGGradientDefinition(cfg.ModuleStyle.GradientStart, cfg.ModuleStyle.GradientEnd, cfg.ModuleStyle.GradientAngle, "qrgradient"))
		b.WriteString("\n")
	}

	for row := range qr.Size {
		for col := range qr.Size {
			if qr.Modules[row][col] {
				if cfg.ModuleStyle.UseAdvanced() {
					b.WriteString("  ")
					b.WriteString(SVGModuleElement(col+cfg.QuietZone, row+cfg.QuietZone, moduleSize, cfg.ModuleStyle, cfg.ForegroundColor))
					b.WriteString("\n")
				} else {
					x := (col + cfg.QuietZone) * moduleSize
					y := (row + cfg.QuietZone) * moduleSize
					fmt.Fprintf(&b, `  <rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
						x, y, moduleSize, moduleSize, cfg.ForegroundColor)
					b.WriteString("\n")
				}
			}
		}
	}
	b.WriteString("</svg>\n")

	return []byte(b.String()), nil
}
