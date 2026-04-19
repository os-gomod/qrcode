package renderer

import (
	"image"
	"image/color"
	"math"
)

// DrawModule renders a single QR code module onto the image at position (x, y)
// with the given pixel scale and style.
//
// For square modules (or nil style), the module cell is filled directly.
// For other shapes, DrawModule dispatches to the appropriate drawing function:
//   - "rounded" → DrawRoundedRect with radius = scale * roundness * 0.5
//   - "circle"  → DrawCircle inscribed in the module cell
//   - "diamond"  → DrawDiamond inscribed in the module cell
//
// If the style has GradientEnabled, the gradient color at the module's
// position is computed and used as the fill color. If Transparency < 1.0,
// the alpha channel is adjusted accordingly.
func DrawModule(img *image.RGBA, x, y, scale int, style *ModuleStyle, fgColor color.Color) {
	if style == nil || style.Shape == "square" {
		c := fgColor
		if style != nil && style.GradientEnabled {
			start, end, _ := ParseGradient(style.GradientStart, style.GradientEnd)
			c = GradientColor(x, y, img.Bounds().Dx(), img.Bounds().Dy(), style.GradientAngle, start, end)
		}
		if style != nil && style.Transparency < 1.0 {
			c = ApplyTransparency(c, style.Transparency)
		}
		for py := y; py < y+scale; py++ {
			for px := x; px < x+scale; px++ {
				img.Set(px, py, c)
			}
		}
		return
	}
	switch style.Shape {
	case "rounded":
		radius := int(float64(scale) * style.Roundness * 0.5)
		if radius < 0 {
			radius = 0
		}
		if radius > scale/2 {
			radius = scale / 2
		}
		c := fgColor
		if style.GradientEnabled {
			start, end, _ := ParseGradient(style.GradientStart, style.GradientEnd)
			c = GradientColor(x, y, img.Bounds().Dx(), img.Bounds().Dy(), style.GradientAngle, start, end)
		}
		if style.Transparency < 1.0 {
			c = ApplyTransparency(c, style.Transparency)
		}
		DrawRoundedRect(img, x, y, scale, scale, radius, c)
	case "circle":
		c := fgColor
		if style.GradientEnabled {
			start, end, _ := ParseGradient(style.GradientStart, style.GradientEnd)
			c = GradientColor(x, y, img.Bounds().Dx(), img.Bounds().Dy(), style.GradientAngle, start, end)
		}
		if style.Transparency < 1.0 {
			c = ApplyTransparency(c, style.Transparency)
		}
		DrawCircle(img, x+scale/2, y+scale/2, scale/2, c)
	case "diamond":
		c := fgColor
		if style.GradientEnabled {
			start, end, _ := ParseGradient(style.GradientStart, style.GradientEnd)
			c = GradientColor(x, y, img.Bounds().Dx(), img.Bounds().Dy(), style.GradientAngle, start, end)
		}
		if style.Transparency < 1.0 {
			c = ApplyTransparency(c, style.Transparency)
		}
		DrawDiamond(img, x+scale/2, y+scale/2, scale/2, scale/2, c)
	}
}

// DrawRoundedRect draws a filled rounded rectangle onto the image.
//
// The radius parameter controls the corner curvature. If radius exceeds
// half the width or half the height, it is clamped to the smaller of the
// two. Negative radii are clamped to zero. Pixels that fall outside the
// rounded corners (i.e., beyond the arc at each corner) are not painted.
//
// Parameters:
//   - img:    target RGBA image
//   - x, y:   top-left corner of the rectangle
//   - w, h:   width and height in pixels
//   - radius: corner radius in pixels
//   - c:      fill color
//
//nolint:gocyclo // geometric corner logic requires sequential conditional checks
//nolint:gocritic
func DrawRoundedRect(img *image.RGBA, x, y, w, h, radius int, c color.Color) {
	maxR := w / 2
	if h/2 < maxR {
		maxR = h / 2
	}
	if radius > maxR {
		radius = maxR
	}
	if radius < 0 {
		radius = 0
	}
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			inCorner := false
			corners := []struct {
				cx, cy int
			}{
				{x + radius, y + radius},
				{x + w - 1 - radius, y + radius},
				{x + radius, y + h - 1 - radius},
				{x + w - 1 - radius, y + h - 1 - radius},
			}
			for _, corner := range corners {
				inCornerRegion := false
				//nolint:gocritic // ifElseChain: corner region checks are spatial, not type-based
				if px < x+radius && py < y+radius && corner.cx == x+radius && corner.cy == y+radius {
					inCornerRegion = true
				} else if px >= x+w-radius && py < y+radius && corner.cx == x+w-1-radius && corner.cy == y+radius {
					inCornerRegion = true
				} else if px < x+radius && py >= y+h-radius && corner.cx == x+radius && corner.cy == y+h-1-radius {
					inCornerRegion = true
				} else if px >= x+w-radius && py >= y+h-radius && corner.cx == x+w-1-radius && corner.cy == y+h-1-radius {
					inCornerRegion = true
				}
				if inCornerRegion {
					dx := px - corner.cx
					dy := py - corner.cy
					distSq := dx*dx + dy*dy
					rSq := radius * radius
					if distSq > rSq {
						inCorner = true
					}
					break
				}
			}
			if !inCorner {
				img.Set(px, py, c)
			}
		}
	}
}

// DrawCircle draws a filled circle onto the image using the midpoint
// distance test. Pixels within radius distance of (cx, cy) are painted.
//
// Parameters:
//   - img:    target RGBA image
//   - cx, cy: center coordinates
//   - radius: circle radius in pixels
//   - c:      fill color
func DrawCircle(img *image.RGBA, cx, cy, radius int, c color.Color) {
	for py := cy - radius; py <= cy+radius; py++ {
		for px := cx - radius; px <= cx+radius; px++ {
			dx := px - cx
			dy := py - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(px, py, c)
			}
		}
	}
}

// DrawDiamond draws a filled diamond (rhombus) shape onto the image.
// The diamond is centered at (cx, cy) with horizontal half-width halfW
// and vertical half-height halfH. A pixel at (px, py) is inside the
// diamond when |dx|/halfW + |dy|/halfH ≤ 1.0.
//
// Parameters:
//   - img:           target RGBA image
//   - cx, cy:        center coordinates
//   - halfW, halfH:  horizontal and vertical half-sizes in pixels
//   - c:             fill color
func DrawDiamond(img *image.RGBA, cx, cy, halfW, halfH int, c color.Color) {
	if halfW <= 0 || halfH <= 0 {
		return
	}
	for py := cy - halfH; py <= cy+halfH; py++ {
		for px := cx - halfW; px <= cx+halfW; px++ {
			dx := absInt(px - cx)
			dy := absInt(py - cy)
			if float64(dx)/float64(halfW)+float64(dy)/float64(halfH) <= 1.0 {
				img.Set(px, py, c)
			}
		}
	}
}

// ParseGradient parses two "#RRGGBB" hex color strings into RGBA colors
// suitable for gradient interpolation. Both colors are returned with alpha
// set to 255 (fully opaque). Returns an error if either color string is
// malformed.
func ParseGradient(startHex, endHex string) (color.RGBA, color.RGBA, error) {
	r1, g1, b1, err := ParseHexColor(startHex)
	if err != nil {
		return color.RGBA{}, color.RGBA{}, err
	}
	r2, g2, b2, err := ParseHexColor(endHex)
	if err != nil {
		return color.RGBA{}, color.RGBA{}, err
	}
	return color.RGBA{R: r1, G: g1, B: b1, A: 255},
		color.RGBA{R: r2, G: g2, B: b2, A: 255},
		nil
}

// InterpolateColor linearly interpolates between two RGBA colors by factor t.
// The factor t is clamped to [0.0, 1.0]. At t=0 the result equals start;
// at t=1 the result equals end. Each channel (R, G, B, A) is interpolated
// independently.
func InterpolateColor(start, end color.RGBA, t float64) color.RGBA {
	if t < 0.0 {
		t = 0.0
	}
	if t > 1.0 {
		t = 1.0
	}
	r := uint8(float64(start.R) + t*(float64(end.R)-float64(start.R)))
	g := uint8(float64(start.G) + t*(float64(end.G)-float64(start.G)))
	b := uint8(float64(start.B) + t*(float64(end.B)-float64(start.B)))
	a := uint8(float64(start.A) + t*(float64(end.A)-float64(start.A)))
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// GradientColor computes the gradient color at pixel position (x, y) for
// a linear gradient with the given angle over an image of the specified
// dimensions.
//
// The gradient direction is determined by angle (in degrees, clockwise
// from the x-axis). The interpolation factor t is computed by projecting
// the point onto the gradient axis and normalising against the image diagonal.
// Returns start if width or height is zero.
func GradientColor(x, y, width, height int, angle float64, start, end color.RGBA) color.RGBA {
	if width == 0 || height == 0 {
		return start
	}
	rad := angle * math.Pi / 180.0
	diag := math.Sqrt(float64(width*width + height*height))
	dx := float64(x)*math.Cos(rad) + float64(y)*math.Sin(rad)
	t := (dx + diag/2) / diag
	return InterpolateColor(start, end, t)
}

// ApplyTransparency returns a copy of the color with the alpha channel set
// to the given value. The alpha parameter is clamped to [0.0, 1.0] and
// converted to an 8-bit value (0–255). The RGB channels are preserved.
func ApplyTransparency(c color.Color, alpha float64) color.RGBA {
	if alpha < 0.0 {
		alpha = 0.0
	}
	if alpha > 1.0 {
		alpha = 1.0
	}
	r, g, b, _ := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(alpha * 255),
	}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
