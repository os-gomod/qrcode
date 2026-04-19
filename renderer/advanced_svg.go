package renderer

import (
	"fmt"
	"math"
	"strings"
)

// SVGModuleAttributes returns a complete SVG element string (including tag
// name and attributes) for a module at grid position (col, row) with pixel
// size scale. This function selects the appropriate SVG element based on
// the ModuleStyle:
//   - nil or "square" → <rect> with x, y, width, height
//   - "rounded"       → <path> using SVGRoundedRectPath
//   - "circle"        → <circle> with cx, cy, r
//   - "diamond"        → <polygon> with four vertices
//
// If GradientEnabled is set, the fill references "url(#qrgradient)".
// If Transparency < 1.0, an opacity attribute is added.
//
// This is used by the advanced SVG rendering path in the SVGRenderer.
//
//nolint:gocritic // sprintfQuotedString: fill values are SVG attribute contents, not Go string literals
func SVGModuleAttributes(col, row, scale int, style *ModuleStyle, fgColor string) string {
	x := col * scale
	y := row * scale
	fill := fgColor
	opacityAttr := ""
	if style != nil && style.Transparency < 1.0 {
		opacityAttr = fmt.Sprintf(` opacity="%.2f"`, style.Transparency)
	}
	if style != nil && style.GradientEnabled {
		fill = "url(#qrgradient)"
	}
	if style == nil || style.Shape == "square" {
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"%s/>`,
			x, y, scale, scale, fill, opacityAttr)
	}
	switch style.Shape {
	case "rounded":
		radius := float64(scale) * style.Roundness * 0.5
		return fmt.Sprintf(`<path d="%s" fill="%s"%s/>`,
			SVGRoundedRectPath(float64(x), float64(y), float64(scale), float64(scale), radius),
			fill, opacityAttr)
	case "circle":
		cx := x + scale/2
		cy := y + scale/2
		r := scale / 2
		return fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s"%s/>`,
			cx, cy, r, fill, opacityAttr)
	case "diamond":
		cx := x + scale/2
		cy := y + scale/2
		pts := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d",
			cx, y,
			x+scale, cy,
			cx, y+scale,
			x, cy,
		)
		return fmt.Sprintf(`<polygon points="%s" fill="%s"%s/>`,
			pts, fill, opacityAttr)
	default:
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"%s/>`,
			x, y, scale, scale, fill, opacityAttr)
	}
}

// SVGGradientDefinition returns an SVG <defs> block containing a
// <linearGradient> element with objectBoundingBox gradient units.
//
// The angle is adjusted by -90° so that 0° maps to a left-to-right gradient
// (matching CSS linear-gradient conventions). The x1/y1/x2/y2 coordinates
// are expressed as fractions in [0, 1].
//
// Parameters:
//   - startColor: gradient start color as a CSS color string
//   - endColor:   gradient end color as a CSS color string
//   - angle:      gradient angle in degrees (0° = left-to-right after adjustment)
//   - id:         unique identifier for the gradient element
func SVGGradientDefinition(startColor, endColor string, angle float64, id string) string {
	rad := (angle - 90) * math.Pi / 180.0
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	x1 := 0.5 - 0.5*cosA
	y1 := 0.5 - 0.5*sinA
	x2 := 0.5 + 0.5*cosA
	y2 := 0.5 + 0.5*sinA
	var b strings.Builder
	fmt.Fprintf(&b, `<defs><linearGradient id="%s" `, id)
	fmt.Fprintf(&b, `x1="%.4f" y1="%.4f" x2="%.4f" y2="%.4f" `, x1, y1, x2, y2)
	b.WriteString(`gradientUnits="objectBoundingBox">`)
	fmt.Fprintf(&b, `<stop offset="0%%" stop-color="%s"/>`, startColor)
	fmt.Fprintf(&b, `<stop offset="100%%" stop-color="%s"/>`, endColor)
	b.WriteString(`</linearGradient></defs>`)
	return b.String()
}

// SVGRoundedRectPath returns an SVG path "d" attribute string for a rounded
// rectangle using arc commands. The path traces the rectangle starting from
// the top-left edge (after the corner radius), moving clockwise with four
// arc segments at each corner.
//
// If the radius is zero, a simple path using only H (horizontal) and V
// (vertical) line-to commands is returned. The radius is clamped to
// [0, min(w, h)/2].
//
// Parameters:
//   - x, y: top-left corner coordinates
//   - w, h: width and height
//   - r:    corner radius
//
// Example output:
//
//	M10.0,10.0 H20.0 A2.0,2.0 0 0,1 22.0,12.0 V20.0 ... Z
func SVGRoundedRectPath(x, y, w, h, r float64) string {
	maxR := w / 2
	if h/2 < maxR {
		maxR = h / 2
	}
	if r > maxR {
		r = maxR
	}
	if r < 0 {
		r = 0
	}
	if r == 0 {
		return fmt.Sprintf("M%.1f,%.1f H%.1f V%.1f H%.1f Z", x, y, x+w, y+h, x)
	}
	return fmt.Sprintf(
		"M%.1f,%.1f H%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f V%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f H%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f V%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f Z",
		x+r, y,
		x+w-r,
		r, r,
		x+w, y+r,
		y+h-r,
		r, r,
		x+w-r, y+h,
		x+r,
		r, r,
		x, y+h-r,
		y+r,
		r, r,
		x+r, y,
	)
}
