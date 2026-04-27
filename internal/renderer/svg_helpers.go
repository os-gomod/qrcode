package renderer

import (
	"fmt"
	"math"
	"strings"
)

// SVGModuleElement returns an SVG element string for a single QR module.
// This is the single consolidated implementation — it replaces the former
// SVGModule (svg_helpers.go) and SVGModuleAttributes (advanced_svg.go).
func SVGModuleElement(col, row, scale int, style *ModuleStyle, fgColor string) string {
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
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill=%q%s/>`,
			x, y, scale, scale, fill, opacityAttr)
	}

	switch style.Shape {
	case "rounded":
		radius := float64(scale) * style.Roundness * 0.5
		return fmt.Sprintf(`<path d=%q fill=%q%s/>`,
			SVGRoundedRectPath(float64(x), float64(y), float64(scale), float64(scale), radius),
			fill, opacityAttr)
	case "circle":
		cx := x + scale/2
		cy := y + scale/2
		r := scale / 2
		return fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill=%q%s/>`,
			cx, cy, r, fill, opacityAttr)
	case "diamond":
		cx := x + scale/2
		cy := y + scale/2
		pts := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d",
			cx, y, x+scale, cy, cx, y+scale, x, cy)
		return fmt.Sprintf(`<polygon points=%q fill=%q%s/>`,
			pts, fill, opacityAttr)
	default:
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill=%q%s/>`,
			x, y, scale, scale, fill, opacityAttr)
	}
}

// SVGGradientDefinition produces an SVG <defs><linearGradient> block.
// This is the single consolidated implementation — it replaces the former
// SVGGradientDef (svg_helpers.go) and SVGGradientDefinition (advanced_svg.go).
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

// SVGRoundedRectPath produces an SVG path "d" attribute for a rounded rectangle.
// If radius is 0, the path has no arc commands.  Negative radius is clamped to 0.
func SVGRoundedRectPath(x, y, w, h, r float64) string {
	maxR := min(w/2, h/2)
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
		x+w-r, r, r, x+w, y+r,
		y+h-r, r, r, x+w-r, y+h,
		x+r, r, r, x, y+h-r,
		y+r, r, r, x+r, y,
	)
}
