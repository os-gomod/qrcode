package renderer

import (
	"image"
	"image/color"
	"math"
)

// ---------------------------------------------------------------------------
// Module drawing (used by PNG renderer for advanced module styles)
// ---------------------------------------------------------------------------

// DrawModule renders a single QR module onto the image at the given position.
// The style determines the shape (square, rounded, circle, diamond) and
// optional gradient/transparency effects.
func DrawModule(img *image.RGBA, x, y, scale int, style *ModuleStyle, fgColor color.Color) {
	imgW, imgH := img.Bounds().Dx(), img.Bounds().Dy()

	if style == nil || style.Shape == "square" {
		c := resolveColor(x, y, imgW, imgH, style, fgColor)
		for py := y; py < y+scale; py++ {
			for px := x; px < x+scale; px++ {
				img.Set(px, py, c)
			}
		}
		return
	}

	c := resolveColor(x, y, imgW, imgH, style, fgColor)
	switch style.Shape {
	case "rounded":
		radius := int(float64(scale) * style.Roundness * 0.5)
		clampRadius(&radius, scale)
		DrawRoundedRect(img, x, y, scale, scale, radius, c)
	case "circle":
		DrawCircle(img, x+scale/2, y+scale/2, scale/2, c)
	case "diamond":
		DrawDiamond(img, x+scale/2, y+scale/2, scale/2, scale/2, c)
	}
}

// resolveColor applies gradient and transparency to the foreground color.
func resolveColor(x, y, imgWidth, imgHeight int, style *ModuleStyle, fg color.Color) color.Color {
	c := fg
	if style != nil && style.GradientEnabled {
		start, end, _ := ParseGradient(style.GradientStart, style.GradientEnd)
		c = GradientColor(x, y, imgWidth, imgHeight, style.GradientAngle, start, end)
	}
	if style != nil && style.Transparency < 1.0 {
		c = ApplyTransparency(c, style.Transparency)
	}
	return c
}

// DrawRoundedRect draws a rounded rectangle on the image.
//
//nolint:gocyclo,cyclop // rounded rect requires per-pixel corner calculation
func DrawRoundedRect(img *image.RGBA, x, y, w, h, radius int, c color.Color) {
	maxR := min(w/2, h/2)
	if radius > maxR {
		radius = maxR
	}
	if radius < 0 {
		radius = 0
	}
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			inCorner := false
			corners := []struct{ cx, cy int }{
				{x + radius, y + radius},
				{x + w - 1 - radius, y + radius},
				{x + radius, y + h - 1 - radius},
				{x + w - 1 - radius, y + h - 1 - radius},
			}
			for _, corner := range corners {
				inCornerRegion := false
				switch {
				case px < x+radius && py < y+radius && corner.cx == x+radius && corner.cy == y+radius:
					inCornerRegion = true
				case px >= x+w-radius && py < y+radius && corner.cx == x+w-1-radius && corner.cy == y+radius:
					inCornerRegion = true
				case px < x+radius && py >= y+h-radius && corner.cx == x+radius && corner.cy == y+h-1-radius:
					inCornerRegion = true
				case px >= x+w-radius && py >= y+h-radius && corner.cx == x+w-1-radius && corner.cy == y+h-1-radius:
					inCornerRegion = true
				}
				if inCornerRegion {
					dx := px - corner.cx
					dy := py - corner.cy
					if dx*dx+dy*dy > radius*radius {
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

// DrawCircle draws a filled circle on the image.
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

// DrawDiamond draws a filled diamond on the image.
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

// ---------------------------------------------------------------------------
// Color interpolation and gradient helpers
// ---------------------------------------------------------------------------

//nolint:revive // two RGBA colors follow Go convention
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

// InterpolateColor performs linear interpolation between two RGBA colors.
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

// GradientColor computes the gradient color at a specific pixel position.
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

// ApplyTransparency applies an alpha value to a color.
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

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func clampRadius(r *int, scale int) {
	if *r < 0 {
		*r = 0
	}
	half := scale / 2
	if *r > half {
		*r = half
	}
}
