package renderer

import (
	"fmt"
	"math"
)

// SVGGradientDef returns an SVG <defs> block containing a <linearGradient>
// element with the specified colors and angle.
//
// The gradient direction is computed from the angle in degrees by projecting
// from the center of the coordinate system. The x1/y1/x2/y2 attributes are
// expressed as percentages (0–100%).
//
// Parameters:
//   - startColor: gradient start color as a CSS color string (e.g., "#FF0000")
//   - endColor:   gradient end color as a CSS color string
//   - angle:      gradient angle in degrees
//   - id:         unique identifier for the gradient (referenced via url(#id))
//
// Example output:
//
//	<defs><linearGradient id="qr-gradient" x1="0.0%" y1="100.0%" ...
//	  <stop offset="0%" stop-color="#FF0000"/>
//	  <stop offset="100%" stop-color="#0000FF"/>
//	</linearGradient></defs>
func SVGGradientDef(startColor, endColor string, angle float64, id string) string {
	angleRad := angle * math.Pi / 180
	x1 := 50 - 50*math.Cos(angleRad)
	y1 := 50 - 50*math.Sin(angleRad)
	x2 := 50 + 50*math.Cos(angleRad)
	y2 := 50 + 50*math.Sin(angleRad)
	return fmt.Sprintf(`<defs><linearGradient id="%s" x1="%.1f%%" y1="%.1f%%" x2="%.1f%%" y2="%.1f%%">`+
		`<stop offset="0%%" stop-color="%s"/>`+
		`<stop offset="100%%" stop-color="%s"/>`+
		`</linearGradient></defs>`, id, x1, y1, x2, y2, startColor, endColor)
}

// SVGRoundedRect returns an SVG <rect> element string with rounded corners.
// The corner radius is computed as roundness * scale * 0.5. The fill and
// opacity are passed as CSS-compatible string values.
//
// Parameters:
//   - col, row: module grid position
//   - scale:    pixel size of each module
//   - color:    CSS fill color (or url(#gradient-id) for gradients)
//   - fill:     CSS opacity value (e.g., "1" or "0.85")
//   - roundness: corner radius factor in [0.0, 1.0]
//
//nolint:gocritic // sprintfQuotedString: SVG attribute values are not Go string literals
func SVGRoundedRect(col, row, scale int, color, fill string, roundness float64) string {
	r := roundness * float64(scale) * 0.5
	return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%.1f" ry="%.1f" fill="%s" opacity="%s"/>`,
		col*scale, row*scale, scale, scale, r, r, color, fill)
}

// SVGCircle returns an SVG <circle> element string centered in the
// module cell at grid position (col, row). The radius is half the
// module scale.
//
// Parameters:
//   - col, row: module grid position
//   - scale:    pixel size of each module
//   - color:    CSS fill color
//   - fill:     CSS opacity value
//
//nolint:gocritic // sprintfQuotedString: SVG attribute values are not Go string literals
func SVGCircle(col, row, scale int, color, fill string) string {
	cx := col*scale + scale/2
	cy := row*scale + scale/2
	r := float64(scale) / 2
	return fmt.Sprintf(`<circle cx="%d" cy="%d" r="%.1f" fill="%s" opacity="%s"/>`,
		cx, cy, r, color, fill)
}

// SVGDiamond returns an SVG <polygon> element string representing a
// diamond (rotated square) inscribed in the module cell at grid position
// (col, row). The four vertices are at the top-center, right-center,
// bottom-center, and left-center of the module cell.
//
// Parameters:
//   - col, row: module grid position
//   - scale:    pixel size of each module
//   - color:    CSS fill color
//   - fill:     CSS opacity value
//
//nolint:gocritic // sprintfQuotedString: SVG attribute values are not Go string literals
func SVGDiamond(col, row, scale int, color, fill string) string {
	x := col * scale
	y := row * scale
	mx := x + scale/2
	my := y + scale/2
	points := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d", mx, y, x+scale, my, mx, y+scale, x, my)
	return fmt.Sprintf(`<polygon points="%s" fill="%s" opacity="%s"/>`, points, color, fill)
}

// SVGModule returns the SVG element string for a single QR code module
// at grid position (col, row). It selects the appropriate SVG primitive
// based on the ModuleStyle:
//   - nil or "square" → <rect> element
//   - "rounded"       → SVGRoundedRect
//   - "circle"        → SVGCircle
//   - "diamond"        → SVGDiamond
//
// If the style has GradientEnabled, the fill is set to "url(#qr-gradient)".
// If Transparency > 0, an opacity attribute is added.
//
//nolint:gocritic // sprintfQuotedString: SVG attribute values are not Go string literals
func SVGModule(col, row, scale int, style *ModuleStyle, color string) string {
	if style == nil || style.Shape == "square" {
		opacity := "1"
		if style != nil && style.Transparency > 0 {
			opacity = fmt.Sprintf("%.2f", clamp01(1-style.Transparency))
		}
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="%s"/>`,
			col*scale, row*scale, scale, scale, color, opacity)
	}
	opacity := "1"
	if style.Transparency > 0 {
		opacity = fmt.Sprintf("%.2f", clamp01(1-style.Transparency))
	}
	if style.GradientEnabled {
		color = "url(#qr-gradient)"
	}
	switch style.Shape {
	case "rounded":
		return SVGRoundedRect(col, row, scale, color, opacity, style.Roundness)
	case "circle":
		return SVGCircle(col, row, scale, color, opacity)
	case "diamond":
		return SVGDiamond(col, row, scale, color, opacity)
	default:
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="%s"/>`,
			col*scale, row*scale, scale, scale, color, opacity)
	}
}

// IsGradientStyle reports whether the module style uses a gradient fill.
// Returns false if the style is nil.
func IsGradientStyle(style *ModuleStyle) bool {
	return style != nil && style.GradientEnabled
}

// NeedsAdvancedSVG reports whether the module style requires non-trivial
// SVG rendering (any shape other than "square"). Returns false if the style
// is nil or uses the default square shape.
func NeedsAdvancedSVG(style *ModuleStyle) bool {
	return style != nil && style.Shape != "square"
}
